//
//  MIT License
//
//  (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.

package cmd

import (
	"context"
	"crypto/tls"
	"github.com/Cray-HPE/cray-site-init/pkg/bss"
	"github.com/Cray-HPE/cray-site-init/pkg/sls"
	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	hms_s3 "github.com/Cray-HPE/hms-s3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/spf13/cobra"
)

const s3Prefix = "s3://ncn-images"

type paramTuple struct {
	key   string
	value string
}

var (
	managementNCNs []sls_common.GenericHardware

	paramsToUpdate []string
	paramsToDelete []string
	limitToXnames  []string
	userDataJSON   string

	desiredKubernetesVersion string
	desiredCEPHVersion       string

	bssClient *bss.UtilsClient
	slsClient *sls.UtilsClient

	verboseLogging bool

	s3Client *hms_s3.S3Client

	s3SecretName string
)

// handoffCmd represents the handoff command
var handoffCmd = &cobra.Command{
	Use:   "handoff",
	Short: "Runs migration steps to transition from LiveCD",
	Long: "A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the " +
		"production version inside the Kubernetes cluster.",
}

func init() {
	rootCmd.AddCommand(handoffCmd)
	handoffCmd.DisableAutoGenTag = true

	verboseLogging = false
	verboseLogging, _ = strconv.ParseBool(os.Getenv("VERBOSE"))
}

// setupClients - Preps clients for various services to abstract interactions with their APIs
// via packages in this project.
func setupClients() {
	setupEnvs()
	setupHTTPClient()

	bssClient = bss.NewBSSClient(bssBaseURL, httpClient, token)
	slsClient = sls.NewSLSClient(slsBaseURL, httpClient, token)
}

// setupCommon - These are steps that every handoff function have in common.
func setupCommon() {
	var err error

	setupClients()

	log.Println("Getting management NCNs from SLS...")
	managementNCNs, err = getManagementNCNsFromSLS()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Done getting management NCNs from SLS.")
}

func getManagementNCNsFromSLS() (managementNCNs []sls_common.GenericHardware, err error) {
	return slsClient.GetManagementNCNs()
}

func uploadEntryToBSS(bssEntry bssTypes.BootParams, method string) {
	uploadedBSSEntry, err := bssClient.UploadEntryToBSS(bssEntry, method)
	if err != nil {
		log.Panicf("Failed to upload entry to BSS: %s", err)
	}

	if verboseLogging {
		log.Printf("Successfully PUT BSS entry for %s:\n%s", bssEntry.Hosts[0], uploadedBSSEntry)
	} else {
		log.Printf("Successfully PUT BSS entry for %s", bssEntry.Hosts[0])
	}
}

func getBSSBootparametersForXname(xname string) bssTypes.BootParams {
	bootParams, err := bssClient.GetBSSBootparametersForXname(xname)
	if err != nil {
		log.Panicf("Failed to get BSS bootparameters for %s: %s", xname, err)
	}

	return *bootParams
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
				log.Panicf("Set parameter had invalid format: %s", setParam)
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

func setupS3(s3BucketName string) {
	// Built kubeconfig.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Panic(err)
	}

	// Create the clientset.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err)
	}

	// Get the secret from Kubernetes.
	s3Secret, err := clientset.CoreV1().Secrets("services").Get(context.TODO(), s3SecretName, v1.GetOptions{})
	if err != nil {
		log.Panic(err)
	}

	// Normally the HMS S3 library uses environment variables but since the vast majority are just arguments to this
	// program manually create the object for connection info.
	s3Connection := hms_s3.ConnectionInfo{
		AccessKey: string(s3Secret.Data["access_key"]),
		SecretKey: string(s3Secret.Data["secret_key"]),
		Endpoint:  string(s3Secret.Data["s3_endpoint"]),
		Bucket:    s3BucketName,
		Region:    "default",
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			// If the image is sufficiently large it's possible for the connection to go stale.
			KeepAlive: 10 * time.Second,
		}).DialContext,
	}
	httpClient := &http.Client{Transport: tr}

	s3Client, err = hms_s3.NewS3Client(s3Connection, httpClient)
	if err != nil {
		log.Panic(err)
	}
}
