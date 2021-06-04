/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"stash.us.cray.com/HMS/hms-bss/pkg/bssTypes"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
)

const s3Prefix = "s3://ncn-images"

type paramTuple struct {
	key   string
	value string
}

var (
	managementNCNs []sls_common.GenericHardware
	httpClient     *http.Client

	paramsToUpdate []string
	paramsToDelete []string
	limitToXnames  []string
	userDataJSON   string

	desiredKubernetesVersion string
	desiredCEPHVersion       string

	gatewayHostname string
	verboseLogging  bool
)

// handoffCmd represents the handoff command
var handoffCmd = &cobra.Command{
	Use:   "handoff",
	Short: "runs migration steps to transition from LiveCD",
	Long: "A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the " +
		"production version inside the Kubernetes cluster.",
}

func init() {
	rootCmd.AddCommand(handoffCmd)

	gatewayHostname = os.Getenv("GATEWAY_HOSTNAME")
	if gatewayHostname == "" {
		gatewayHostname = "api-gw-service-nmn.local"
	}

	verboseLogging = false
	verboseLogging, _ = strconv.ParseBool(os.Getenv("VERBOSE"))
}

func setupCommon() {
	var err error

	// These are steps that every handoff function have in common.
	token = os.Getenv("TOKEN")
	if token == "" {
		log.Panicln("Environment variable TOKEN can NOT be blank!")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	httpClient = &http.Client{Transport: transport}

	log.Println("Getting management NCNs from SLS...")
	managementNCNs, err = getManagementNCNsFromSLS()
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Done getting management NCNs from SLS.")
}

func getManagementNCNsFromSLS() (managementNCNs []sls_common.GenericHardware, err error) {
	url := fmt.Sprintf("https://%s/apis/sls/v1/search/hardware?extra_properties.Role=Management",
		gatewayHostname)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("failed to create new request: %w", err)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to do request: %w", err)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &managementNCNs)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal body: %w", err)
	}

	return
}

func uploadEntryToBSS(bssEntry bssTypes.BootParams, method string) {
	url := fmt.Sprintf("https://%s/apis/bss/boot/v1/bootparameters", gatewayHostname)

	jsonBytes, err := json.Marshal(bssEntry)
	if err != nil {
		log.Panicln(err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Panicf("Failed to create new request: %s", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Panicf("Failed to %s BSS entry: %s", method, err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		log.Panicf("Failed to %s BSS entry: %s", method, string(bodyBytes))
	}

	jsonPrettyBytes, _ := json.MarshalIndent(bssEntry, "", "  ")

	if verboseLogging {
		log.Printf("Sucessfuly %s BSS entry for %s:\n%s", method, bssEntry.Hosts[0], string(jsonPrettyBytes))
	} else {
		log.Printf("Sucessfuly %s BSS entry for %s", method, bssEntry.Hosts[0])
	}
}

func getBSSBootparametersForXname(xname string) bssTypes.BootParams {
	url := fmt.Sprintf("https://%s/apis/bss/boot/v1/bootparameters?name=%s", gatewayHostname, xname)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Panicf("Failed to create new request: %s", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Panicf("Failed to get BSS entry: %s", err)
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Panicf("Failed to put BSS entry: %s", string(bodyBytes))
	}

	// BSS gives back an array.
	var bssEntries []bssTypes.BootParams
	err = json.Unmarshal(bodyBytes, &bssEntries)
	if err != nil {
		log.Panicf("Failed to unmarshal BSS entries: %s", err)
	}

	// We should only ever get one entry for a given xname.
	if len(bssEntries) != 1 {
		log.Panicf("Unexpected number of BSS entries: %+v", bssEntries)
	}

	return bssEntries[0]
}

func setupHandoffCommon() (limitManagementNCNs []sls_common.GenericHardware, setParams []paramTuple) {
	// Don't process the cli if we have a json input file
	if userDataJSON == "" {
		if len(paramsToUpdate) == 0 && len(paramsToDelete) == 0 {
			log.Fatalln("No parameters given to set or delete!")
		}

		// Build up a slice of tuples of all the values we want to set.
		for _, setParam := range paramsToUpdate {
			paramSplit := strings.Split(setParam, "=")

			if len(paramSplit) != 2 {
				log.Panicf("Set paramater had invalid format: %s", setParam)
			}

			tuple := paramTuple{
				key:   paramSplit[0],
				value: paramSplit[1],
			}

			setParams = append(setParams, tuple)
		}
	}

	// "Global" is a special NCN just used for cloud-init metadata in a global sense.
	managementNCNs = append(managementNCNs, sls_common.GenericHardware{
		Xname: "Global",
	})

	// Only process the NCNs specified.
	if len(limitToXnames) == 0 {
		limitManagementNCNs = managementNCNs
	} else {
		for _, xname := range limitToXnames {
			found := false

			for _, ncn := range managementNCNs {
				if ncn.Xname == xname {
					limitManagementNCNs = append(limitManagementNCNs, ncn)
					found = true
					break
				}
			}

			if !found {
				log.Fatalf("Limit to xname not found in management NCNs: %s", xname)
			}
		}
	}

	return
}
