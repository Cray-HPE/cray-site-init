/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
*/

package handoff

import (
	"crypto/tls"
	"github.com/Cray-HPE/cray-site-init/pkg/csm"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	hmsS3 "github.com/Cray-HPE/hms-s3"
	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/spf13/cobra"

	"github.com/Cray-HPE/cray-site-init/pkg/bss"
	"github.com/Cray-HPE/cray-site-init/pkg/sls"
)

const s3Prefix = "s3://boot-images"
const dsEndpoint = "http://10.92.100.81:8888/"

type paramTuple struct {
	key   string
	value string
}

var (
	token      string
	httpClient *http.Client

	bssBaseURL string
	hsmBaseURL string
	slsBaseURL string

	managementNCNs []slsCommon.GenericHardware

	paramsToUpdate []string
	paramsToDelete []string
	limitToXnames  []string
	userDataJSON   string

	desiredKubernetesVersion string
	desiredCephVersion       string

	bssClient *bss.UtilsClient
	slsClient *sls.UtilsClient

	verboseLogging bool

	s3Client *hmsS3.S3Client

	s3SecretName string
)

// NewCommand creates the handoff command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "handoff",
		DisableAutoGenTag: true,
		Short:             "Runs migration steps to transition from LiveCD",
		Long: "A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the " +
			"production version inside the Kubernetes cluster.",
	}

	verboseLogging = false
	verboseLogging, _ = strconv.ParseBool(os.Getenv("VERBOSE"))
	c.AddCommand(
		NewHandoffCloudInitCommand(),
		NewHandoffMetadataCommand(),
		NewHandoffBSSParamsCommand(),
	)
	return c
}

func setupEnvs() {
	token = os.Getenv("TOKEN")
	if token == "" {
		log.Println("TOKEN was not set. Attempting to read API token from Kubernetes directly ... ")
		var err error
		token, err = csm.GetToken(
			csm.DefaultNameSpace,
			csm.DefaultAdminTokenSecretName,
		)
		if err != nil {
			log.Panicf(
				"Neither the environment variable [TOKEN] or Kubernetes %s provided a useful token!",
				csm.DefaultAdminTokenSecretName,
			)
		}
	}

	bssBaseURL = os.Getenv("BSS_BASE_URL")
	if bssBaseURL == "" {
		bssBaseURL = csm.DefaultBSSBaseURL
	}

	hsmBaseURL = os.Getenv("HSM_BASE_URL")
	if hsmBaseURL == "" {
		hsmBaseURL = csm.DefaultSMDBaseURL
	}

	slsBaseURL = os.Getenv("SLS_BASE_URL")
	if slsBaseURL == "" {
		slsBaseURL = csm.DefaultSLSBaseURL
	}
}

func setupHTTPClient() {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	httpClient = &http.Client{Transport: transport}
}

// SetupClients - Preps clients for various services to abstract interactions with their APIs
// via packages in this project.
func SetupClients() {
	setupEnvs()
	setupHTTPClient()

	bssClient = bss.NewBSSClient(
		bssBaseURL,
		httpClient,
		token,
	)
	slsClient = sls.NewSLSClient(
		slsBaseURL,
		httpClient,
		token,
	)
}

// setupCommon - These are steps that every handoff function have in common.
func setupCommon() {
	var err error

	SetupClients()

	log.Println("Getting management NCNs from SLS...")
	managementNCNs, err = GetManagementNCNsFromSLS()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Done getting management NCNs from SLS.")
}

// GetManagementNCNsFromSLS gets a list of management NCNs from SLS.
func GetManagementNCNsFromSLS() (
	managementNCNs []slsCommon.GenericHardware, err error,
) {
	return slsClient.GetManagementNCNs()
}

func uploadEntryToBSS(
	bssEntry bssTypes.BootParams, method string,
) {
	uploadedBSSEntry, err := bssClient.UploadEntryToBSS(
		bssEntry,
		method,
	)
	if err != nil {
		log.Panicf(
			"Failed to upload entry to BSS: %s",
			err,
		)
	}

	if verboseLogging {
		log.Printf(
			"Successfully PUT BSS entry for %s:\n%s",
			bssEntry.Hosts[0],
			uploadedBSSEntry,
		)
	} else {
		log.Printf(
			"Successfully PUT BSS entry for %s",
			bssEntry.Hosts[0],
		)
	}
}

func getBSSBootparametersForXname(xname string) bssTypes.BootParams {
	bootParams, err := bssClient.GetBSSBootparametersForXname(xname)
	if err != nil {
		log.Panicf(
			"Failed to get BSS bootparameters for %s: %s",
			xname,
			err,
		)
	}

	return *bootParams
}

func setupHandoffCommon() (
	limitManagementNCNs []slsCommon.GenericHardware, setParams []paramTuple,
) {
	// Don't process the cli if we have a json input file
	if userDataJSON == "" {
		if len(paramsToUpdate) == 0 && len(paramsToDelete) == 0 {
			log.Fatalln("No parameters given to set or delete!")
		}

		// Build up a slice of tuples of all the values we want to set.
		for _, setParam := range paramsToUpdate {
			paramSplit := strings.Split(
				setParam,
				"=",
			)

			if len(paramSplit) != 2 {
				log.Panicf(
					"Set parameter had invalid format: %s",
					setParam,
				)
			}

			tuple := paramTuple{
				key:   paramSplit[0],
				value: paramSplit[1],
			}

			setParams = append(
				setParams,
				tuple,
			)
		}
	}

	// "Global" is a special NCN just used for cloud-init metadata in a global sense.
	managementNCNs = append(
		managementNCNs,
		slsCommon.GenericHardware{
			Xname: "Global",
		},
	)

	// Only process the NCNs specified.
	if len(limitToXnames) == 0 {
		limitManagementNCNs = managementNCNs
	} else {
		for _, xname := range limitToXnames {
			found := false

			for _, ncn := range managementNCNs {
				if ncn.Xname == xname {
					limitManagementNCNs = append(
						limitManagementNCNs,
						ncn,
					)
					found = true
					break
				}
			}

			if !found {
				log.Fatalf(
					"Limit to xname not found in management NCNs: %s",
					xname,
				)
			}
		}
	}

	return
}
