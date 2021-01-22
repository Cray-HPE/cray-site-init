/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"stash.us.cray.com/HMS/hms-bss/pkg/bssTypes"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
	"strings"
	"syscall"
)

const gatewayHostname = "api-gw-service.nmn"
const s3Prefix = "s3://ncn-images/"

var (
	dataFile string
	cloudInitData map[string]shasta.CloudInit
	sshConfig *ssh.ClientConfig
)

// handoffCmd represents the handoff command
var handoffBSSMetadataCmd = &cobra.Command{
	Use:   "bss-metadata",
	Short: "runs migration steps to build BSS entries for all NCNs",
	Long:  "Using PIT configuration builds kernel command line arguments and cloud-init metadata for each NCN",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building BSS metadata for NCNs...")
		populateNCNMetadata()
		fmt.Println("Done building BSS metadata for NCNs.")

		fmt.Println("Transferring global cloud-init metadata to BSS...")
		// TODO: Build this.
		fmt.Println("Done transferring global cloud-init metadata to BSS.")
	},
}

func init() {
	handoffCmd.AddCommand(handoffBSSMetadataCmd)

	handoffBSSMetadataCmd.Flags().StringVar(&dataFile, "data-file",
		"", "data.json file with cloud-init configuration for each node and global")
	_ = handoffBSSMetadataCmd.MarkFlagRequired("data-file")
}

func getManagementNCNsFromSLS() (managementNCNs []sls_common.GenericHardware, err error) {
	token := os.Getenv("TOKEN")
	if token == "" {
		err = fmt.Errorf("environment variable TOKEN can NOT be blank")
		return
	}

	url := fmt.Sprintf("https://%s/apis/sls/v1/search/hardware?extra_properties.Role=Management",
		gatewayHostname)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("failed to create new request: %w", err)
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
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

func getKernelCommandlineArgs(ncn sls_common.GenericHardware, cmdline string) string {
	var extraProperties sls_common.ComptypeNode
	_ = mapstructure.Decode(ncn.ExtraPropertiesRaw, &extraProperties)

	cmdlineParts := strings.Fields(cmdline)

	for i, _ := range cmdlineParts {
		part := cmdlineParts[i]

		if strings.HasPrefix(part, "metal.server") {
			cmdlineParts[i] = fmt.Sprintf("metal.server=https://rgw.local/ncn-images")
		} else if strings.HasPrefix(part, "ds=nocloud-net") {
			cmdlineParts[i] = fmt.Sprintf("ds=nocloud-net;s=http://%s:8888/", gatewayHostname)
		} else if strings.HasPrefix(part, "rd.live.squashimg") {
			var squashFSName string

			// Storage NCNs get a different image than masters/workers.
			if extraProperties.SubRole == "Storage" {
				squashFSName = cephSquashFSName
			} else {
				squashFSName = k8sSquashFSName
			}

			cmdlineParts[i] = fmt.Sprintf("rd.live.squashimg=%s", squashFSName)
		} else if strings.HasPrefix(part, "hostname") {
			cmdlineParts[i] = fmt.Sprintf("hostname=%s", getNCNHostname(ncn))
		}
	}

	return strings.Join(cmdlineParts, " ")
}

func getNCNHostname(ncn sls_common.GenericHardware) (hostname string) {
	var extraProperties sls_common.ComptypeNode
	_ = mapstructure.Decode(ncn.ExtraPropertiesRaw, &extraProperties)

	if len(extraProperties.Aliases) > 0 {
		hostname = extraProperties.Aliases[0]
	} else {
		hostname = ncn.Xname
	}

	return
}

func getBSSEntryForNCN(ncn sls_common.GenericHardware) (bssEntry bssTypes.BootParams){
	hostname := getNCNHostname(ncn)

	if hostname == "ncn-m001" {
		return
	}

	log.Printf("SSHing to %s...", hostname)

	sshConnection, err := ssh.Dial("tcp", hostname+":22", sshConfig)
	if err != nil {
		log.Panic(err)
	}

	sshSession, err := sshConnection.NewSession()
	if err != nil {
		log.Panic(err)
	}

	cmdline, err := sshSession.CombinedOutput("cat /proc/cmdline")
	if err != nil {
		log.Panic(err)
	}

	_ = sshSession.Close()
	_ = sshConnection.Close()

	// This might seem strange given we have the MAC addresses we could most likely just O(1) lookup the record we
	// need, but, that depends on the MAC addresses being correct and matching what is booted _right now_. However,
	// the xname will never change, so if we find the entry that matches on an xname basis, we know we're good.
	var ncnCloudInitData shasta.CloudInit
	for _, data := range cloudInitData {
		if data.MetaData.Xname == ncn.Xname {
			ncnCloudInitData = data
			break
		}
	}

	// BSS expects generic structures.
	var metaData map[string]interface{}
	err = mapstructure.Decode(ncnCloudInitData.MetaData, &metaData)
	if err != nil {
		log.Panic(err)
	}

	// Now we can build the BSS structure.
	bssEntry = bssTypes.BootParams{
		Hosts:     []string{ncn.Xname},
		Params:    getKernelCommandlineArgs(ncn, string(cmdline)),
		Kernel:    s3Prefix + kernelName,
		Initrd:    s3Prefix + initrdName,
		CloudInit: bssTypes.CloudInit{
			MetaData:  metaData,
			UserData:  ncnCloudInitData.UserData,
			PhoneHome: bssTypes.PhoneHome{},
		},
	}

	return
}

func populateNCNMetadata() {
	var err error
	var ncnRootPassword string
	var bssEntries []bssTypes.BootParams

	// Parse the data.json file.
	err = csiFiles.ReadJSONConfig(dataFile, &cloudInitData)
	if err != nil {
		log.Fatalln("Couldn't parse data file: ", err)
	}

	managementNCNs, err := getManagementNCNsFromSLS()
	if err != nil {
		log.Panicln(err)
	}

	fmt.Print("Enter root password for NCNs: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		log.Panicln(err)
	} else {
		ncnRootPassword = string(bytePassword)
	}

	// Now we must build the kernel cmdline parameters for each NCN. The thing that's not so fun about this is those
	// are calculated as part of PXE booting, so there is no file we can reference as a source of truth. This means
	// we have no choice, we have to gather this from already booted NCNs and replace the values specific to each
	// node. As of the time of writing the only way to do this is to SSH to the thing and read the value directly.
	sshConfig = &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(ncnRootPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Now SSH to each of the NCNs that is *not* m001 and fetch their booted cmdline arguments.
	for _, ncn := range managementNCNs {
		bssEntry := getBSSEntryForNCN(ncn)
		bssEntries = append(bssEntries, bssEntry)
	}

	spew.Dump(bssEntries)
}
