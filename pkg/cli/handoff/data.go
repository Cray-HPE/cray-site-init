/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/viper"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	slsCommon "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/Cray-HPE/hms-smd/pkg/sm"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
)

const hwAddrPrefix = "Permanent HW addr: "
const macGatherCommand = "for interface in /sys/class/net/*; do echo -n " +
	"\"$interface,\" && cat \"$interface/address\"; done"
const bondMACGatherCommand = "grep \"Permanent HW addr: \" /proc/net/bonding/bond0"
const vlanGatherCommand = "ip -j addr show "

const kernelName = "kernel"
const initrdName = "initrd"
const rootfsName = "rootfs"

// This structure definition here exists to allow the parsing of the JSON structure that comes from the `ip` command.
type ipJSONStruct struct {
	Link     string `json:"link"`
	IFName   string `json:"ifname"`
	Address  string `json:"address"`
	AddrInfo []struct {
		Family string `json:"family"`
		Local  string `json:"local"`
		Label  string `json:"label"`
	} `json:"addr_info"`
}
type ipJSONStructArray []ipJSONStruct

var (
	dataFile                                string
	kubernetesImsImageID, storageImsImageID string
	kubernetesUUID, storageUUID             string
	cloudInitData                           map[string]bssTypes.CloudInit
	sshPassword                             string
	// Track whether the password callback might have errored and
	// need a retry with a prompt
	sshPasswordRetry = false
	sshPasswordHost  = ""
	sshConfig        *ssh.ClientConfig
	vlansToGather    = []string{"bond0.nmn0"}
)

// NewHandoffMetadataCommand creates the bss-metadata subcommand.
func NewHandoffMetadataCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "bss-metadata",
		DisableAutoGenTag: true,
		Short:             "runs migration steps to build BSS entries for all NCNs",
		Long:              "Using PIT configuration builds kernel command line arguments and cloud-init metadata for each NCN",
		Run: func(c *cobra.Command, args []string) {
			v := viper.GetViper()

			v.BindPFlags(c.Flags())

			setupCommon()

			// Validate a given string is a valid UUID
			versionRegex, err := regexp.Compile(`^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}$`)
			if err != nil {
				fmt.Println(err)
			}

			if kubernetesUUID == "" || storageUUID == "" {
				log.Fatalln("ERROR: Missing --kubernetes-ims-image-id or --storage-ims-image-id")
			}

			// Validate the input is a valid UUID
			kubernetesImageIDMatch := versionRegex.FindStringSubmatch(kubernetesUUID)
			if kubernetesImageIDMatch == nil {
				log.Fatalf(
					"ERROR: Could not determine Kubernetes image ID from [%s]",
					kubernetesUUID,
				)
			} else {
				kubernetesImsImageID = kubernetesImageIDMatch[0]
			}

			// Validate the input is a valid UUID
			cephImsImageIDMatch := versionRegex.FindStringSubmatch(storageUUID)
			if cephImsImageIDMatch == nil {
				log.Fatalf(
					"ERROR: Could not determine storage image ID from [%s]",
					storageUUID,
				)
			} else {
				storageImsImageID = cephImsImageIDMatch[0]
			}

			// Parse the data.json file.
			err = csiFiles.ReadJSONConfig(
				dataFile,
				&cloudInitData,
			)
			if err != nil {
				log.Fatalln(
					"Couldn't parse data file: ",
					err,
				)
			}

			log.Println("Building BSS metadata for NCNs...")
			populateNCNMetadata()
			log.Println("Done building BSS metadata for NCNs.")

			log.Println("Transferring global cloud-init metadata to BSS...")
			populateGlobalMetadata()
			log.Println("Done transferring global cloud-init metadata to BSS.")
		},
	}

	c.Flags().StringVar(
		&dataFile,
		"data-file",
		"",
		"data.json file with cloud-init configuration for each node and global",
	)
	_ = c.MarkFlagRequired("data-file")

	c.Flags().StringVar(
		&kubernetesUUID,
		"kubernetes-ims-image-id",
		"",
		"The Kubernetes IMS_IMAGE_ID UUID value",
	)

	c.Flags().StringVar(
		&storageUUID,
		"storage-ims-image-id",
		"",
		"The storage-ceph IMS_IMAGE_ID UUID value",
	)

	c.MarkFlagsRequiredTogether(
		"kubernetes-ims-image-id",
		"storage-ims-image-id",
	)
	return c
}

func getKernelCommandlineArgs(
	ncn slsCommon.GenericHardware, cmdline string,
) string {
	var extraProperties slsCommon.ComptypeNode
	_ = mapstructure.Decode(
		ncn.ExtraPropertiesRaw,
		&extraProperties,
	)

	cmdlineParts := strings.Fields(cmdline)

	for i := range cmdlineParts {
		part := cmdlineParts[i]

		if strings.HasPrefix(
			part,
			"metal.server",
		) {
			var imsImageID string

			// Storage NCNs get different assets than masters/workers.
			if extraProperties.SubRole == "Storage" {
				imsImageID = storageImsImageID
			} else {
				imsImageID = kubernetesImsImageID
			}

			if imsImageID == "" {
				log.Fatalf("ERROR: Could not determine version, version was empty.")
			}

			cmdlineParts[i] = fmt.Sprintf(
				"metal.server=%s/%s/%s",
				s3Prefix,
				imsImageID,
				rootfsName,
			)
		} else if strings.HasPrefix(
			part,
			"ds=nocloud-net",
		) {
			cmdlineParts[i] = fmt.Sprintf(
				"ds=nocloud-net;s=%s",
				dsEndpoint,
			)
		} else if strings.HasPrefix(
			part,
			"hostname",
		) {
			cmdlineParts[i] = fmt.Sprintf(
				"hostname=%s",
				getNCNHostname(ncn),
			)
		} else if strings.HasPrefix(
			part,
			"xname",
		) {
			// BSS sets the xname, no need to have it here.
			cmdlineParts[i] = ""
		} else if strings.HasPrefix(
			part,
			"kernel",
		) {
			// We don't need to specific the kernel pram because BSS will *always* add it.
			cmdlineParts[i] = ""
		} else if strings.HasPrefix(
			part,
			"ifname=hsn",
		) {
			// Do not assign hsn* interface names.
			cmdlineParts[i] = ""
		} else if strings.HasPrefix(
			part,
			"ip=hsn",
		) {
			// Do not set parameters for hsn* interfaces.
			cmdlineParts[i] = ""
		}
	}

	// Get rid of any empty values.
	var finalParts []string
	for _, part := range cmdlineParts {
		if part != "" {
			finalParts = append(
				finalParts,
				part,
			)
		}
	}

	return strings.Join(
		finalParts,
		" ",
	)
}

func getNCNHostname(ncn slsCommon.GenericHardware) (hostname string) {
	var extraProperties slsCommon.ComptypeNode
	_ = mapstructure.Decode(
		ncn.ExtraPropertiesRaw,
		&extraProperties,
	)

	if len(extraProperties.Aliases) > 0 {
		hostname = extraProperties.Aliases[0]
	} else {
		hostname = ncn.Xname
	}

	return
}

func getSSHClientForHostname(hostname string) (sshClient *ssh.Client) {
	var err error

	// Don't try to SSH to ourself.
	if hostname == "ncn-m001" {
		return
	}

	log.Printf(
		"Connecting to %s...",
		hostname,
	)

	// Remember the hostname we are trying to connect to so we can
	// use it if we prompt for a password / password retry.
	sshPasswordHost = hostname
	sshClient, err = ssh.Dial(
		"tcp",
		hostname+":22",
		sshConfig,
	)
	if err != nil {
		log.Printf(
			"Unable to connect to %s. Was the supplied password correct?",
			hostname,
		)
		log.Fatal(
			"Error detail: ",
			err,
		)
	}

	return
}

func runSSHCommandWithClient(
	sshClient *ssh.Client, command string,
) (
	output string, err error,
) {
	if sshClient == nil {
		return "", nil
	}

	log.Printf(
		"Creating session to %s...",
		sshClient.RemoteAddr(),
	)

	// When making a new SSH session, the first time into the
	// password authenticator, if it gets called let it use any
	// existing cached password. If that fails, the retry logic in
	// the password authenticator will kick in and start prompting
	// again.
	sshPasswordRetry = false
	sshSession, err := sshClient.NewSession()
	if err != nil {
		log.Fatal(
			"Failed to create session: ",
			err,
		)
	}
	defer sshSession.Close()

	cmdline, err := sshSession.CombinedOutput(command)
	if err != nil {
		return "", err
	}

	return string(cmdline), nil
}

func getCloudInitMetadataForNCN(ncn slsCommon.GenericHardware) (
	userData bssTypes.CloudDataType, metaData bssTypes.CloudDataType,
) {
	// This might seem strange given we have the MAC addresses we could most likely just O(1) lookup the record we
	// need, but, that depends on the MAC addresses being correct and matching what is booted _right now_. However,
	// the xname will never change, so if we find the entry that matches on an xname basis, we know we're good.
	for _, data := range cloudInitData {
		xname := data.MetaData["xname"]
		if xname == ncn.Xname {
			userData = data.UserData
			metaData = data.MetaData

			return
		}
	}

	return
}

func getMACsFromString(macAddrStrings string) (ncnMacs []string) {
	// macAddrStrings is a block of MACs so now we need to filter out only the ones we care about.
	macAddrsSplit := strings.Split(
		macAddrStrings,
		"\n",
	)
	for _, macAddrLine := range macAddrsSplit {
		// We have a range of entries that look like this:
		//   /sys/class/net/bond0,14:02:ec:d9:7c:40
		// So now we split once again to get the name and the MAC.
		macPieces := strings.Split(
			macAddrLine,
			",",
		)
		if len(macPieces) != 2 {
			// Ignore anything that doesn't for any reason have both pieces.
			continue
		}

		interfaceName := strings.TrimPrefix(
			macPieces[0],
			"/sys/class/net/",
		)
		macAddrString := macPieces[1]

		// We want to ignore a fair number of interfaces.
		if interfaceName == "bonding_masters" ||
			interfaceName == "dummy" ||
			interfaceName == "lo" ||
			strings.Contains(
				interfaceName,
				"usb",
			) ||
			strings.Contains(
				interfaceName,
				"veth",
			) ||
			strings.Contains(
				interfaceName,
				"weave",
			) ||
			strings.Contains(
				interfaceName,
				"cilium",
			) ||
			strings.Contains(
				interfaceName,
				"kube",
			) ||
			strings.Contains(
				interfaceName,
				"lxc",
			) {
			continue
		}

		// No reason for this to not work, but, might as well really double check this MAC.
		hw, err := net.ParseMAC(macAddrString)
		if err != nil {
			continue
		}

		ncnMacs = append(
			ncnMacs,
			hw.String(),
		)
	}

	return
}

func getBSSEntryForNCN(ncn slsCommon.GenericHardware) (bssEntry bssTypes.BootParams) {
	hostname := getNCNHostname(ncn)

	var extraProperties slsCommon.ComptypeNode
	_ = mapstructure.Decode(
		ncn.ExtraPropertiesRaw,
		&extraProperties,
	)

	// Setup an SSH connection for those who need it.
	sshClient := getSSHClientForHostname(hostname)

	cmdline, err := runSSHCommandWithClient(
		sshClient,
		"cat /proc/cmdline",
	)
	if err != nil {
		log.Panic(err)
	}

	macAddrStrings, err := runSSHCommandWithClient(
		sshClient,
		macGatherCommand,
	)
	if err != nil {
		log.Panic(err)
	}
	macs := getMACsFromString(macAddrStrings)

	// If there is a bonding device called 'bond0' this will return a string with
	// the MAC addresses on that bonding device. Otherwise it will return an
	// error, which we can ignore because we have a MAC address already from
	// /proc/cmdline above.
	bondMACStrings, err := runSSHCommandWithClient(
		sshClient,
		bondMACGatherCommand,
	)
	if err == nil {
		macs = append(
			macs,
			getBondMACsFromString(bondMACStrings)...,
		)
	} else {
		log.Printf(
			"Warning: gathering MAC addresses from bond0 failed (this could be normal) - %s",
			err,
		)
	}

	// This is not even related to BSS but it makes the most sense to do here...we need to make sure HSM has correct
	// EthernetInterface entries for all the NCNs and since they don't DHCP Kea won't do it for us. So take advantage
	// of the fact we're already in here running commands and gather/populate those entries in HSM.
	for _, vlan := range vlansToGather {
		var vlanOutputString string

		// TODO: This is hacky and I don't like it.
		if hostname == "ncn-m001" {
			vlanOutputString = getPITVLanString(vlan)
		} else {
			vlanOutputString, err = runSSHCommandWithClient(
				sshClient,
				vlanGatherCommand+vlan,
			)
			if err != nil {
				log.Panic(err)
			}
		}

		populateHSMEthernetInterface(
			ncn.Xname,
			vlanOutputString,
			vlan,
		)
	}

	userData, metaData := getCloudInitMetadataForNCN(ncn)

	// Close the SSH connection.
	if sshClient != nil {
		_ = sshClient.Close()
	}

	var imsImageID string

	// Storage NCNs get different assets than masters/workers.
	if extraProperties.SubRole == "Storage" {
		imsImageID = storageImsImageID
	} else {
		imsImageID = kubernetesImsImageID
	}

	if imsImageID == "" {
		log.Fatalf("ERROR: Could not determine IMS_IMAGE_ID")
	}

	// Now we can build the BSS structure.
	bssEntry = bssTypes.BootParams{
		Hosts: []string{ncn.Xname},
		Macs:  macs,
		Params: getKernelCommandlineArgs(
			ncn,
			cmdline,
		),
		Kernel: fmt.Sprintf(
			"%s/%s/%s",
			s3Prefix,
			imsImageID,
			kernelName,
		),
		Initrd: fmt.Sprintf(
			"%s/%s/%s",
			s3Prefix,
			imsImageID,
			initrdName,
		),
		CloudInit: bssTypes.CloudInit{
			MetaData: metaData,
			UserData: userData,
		},
	}
	return
}

func buildPITArgs(base string) string {
	// Get the bond MACs we need.
	macs := getPITBondMACs()

	// Now just do a little find and replace.
	cmdlineParts := strings.Fields(base)

	for i := range cmdlineParts {
		part := cmdlineParts[i]

		// Looking at this might make your brain hurt a little and this can almost certainly be done better but the
		// idea is simple in nature: if we have 2 bonds then we have 4 interfaces; if we have 1 bond then we have 2
		// interfaces. We can guarantee the bond configuration that is in use right now is correct, but the naming of
		// the "mgmt" interfaces might be 0 and 1 or 0 and 2. So what we'll do is this:
		//   * If there is only one bond (i.e., 2 MACs), we're rewrite the config to be 0 and 1.
		//   * If there are 2 bonds (i.e., 4 MACs), then everything will just work out.
		if strings.HasPrefix(
			part,
			"hostname",
		) {
			cmdlineParts[i] = fmt.Sprintf("hostname=ncn-m001")
		} else if strings.HasPrefix(
			part,
			"ifname=mgmt0",
		) {
			if len(macs) >= 2 {
				cmdlineParts[i] = fmt.Sprintf(
					"ifname=mgmt0:%s",
					macs[0],
				)
			} else {
				cmdlineParts[i] = ""
			}
		} else if strings.HasPrefix(
			part,
			"ifname=mgmt1",
		) {
			if len(macs) >= 4 {
				cmdlineParts[i] = fmt.Sprintf(
					"ifname=mgmt1:%s",
					macs[2],
				)
			} else {
				cmdlineParts[i] = fmt.Sprintf(
					"ifname=mgmt1:%s",
					macs[1],
				)
			}
		} else if strings.HasPrefix(
			part,
			"ifname=sun0",
		) {
			if len(macs) >= 4 {
				cmdlineParts[i] = fmt.Sprintf(
					"ifname=sun0:%s",
					macs[1],
				)
			} else {
				cmdlineParts[i] = ""
			}
		} else if strings.HasPrefix(
			part,
			"ifname=sun1",
		) {
			if len(macs) >= 4 {
				cmdlineParts[i] = fmt.Sprintf(
					"ifname=sun1:%s",
					macs[3],
				)
			} else {
				cmdlineParts[i] = ""
			}
		}
	}

	// Just to get rid of the whitespace.
	var finalParts []string
	for _, part := range cmdlineParts {
		if part != "" {
			finalParts = append(
				finalParts,
				part,
			)
		}
	}

	return strings.Join(
		finalParts,
		" ",
	)
}

func getBondMACsFromString(bondMACs string) (macs []string) {
	for _, line := range strings.Split(
		bondMACs,
		"\n",
	) {
		if strings.HasPrefix(
			line,
			hwAddrPrefix,
		) {
			macs = append(
				macs,
				strings.TrimPrefix(
					line,
					hwAddrPrefix,
				),
			)
		}
	}

	return
}

func getPITBondMACs() (macs []string) {
	// We need to additionally add the *permanent* physical MACs for the bond members.
	data, err := ioutil.ReadFile("/proc/net/bonding/bond0")
	if err != nil {
		log.Printf(
			"Warning: gathering MAC addresses from bond0 failed (this could be normal) - %s",
			err,
		)
		macs = []string{}
	} else {
		macs = getBondMACsFromString(string(data))
	}

	// Hopefully future proofing against if we have 2 bonds.
	data, err = ioutil.ReadFile("/proc/net/bonding/bond1")
	if err == nil {
		macs = append(
			macs,
			getBondMACsFromString(string(data))...,
		)
	}

	return
}

func getAllPITMACs() (macs []string) {
	out, err := exec.Command(
		"bash",
		"-c",
		macGatherCommand,
	).Output()
	if err != nil {
		log.Panic(err)
	}
	macs = getMACsFromString(string(out))

	// Also include the bond MACs.
	macs = append(
		macs,
		getPITBondMACs()...,
	)

	return
}

func getPITVLanString(vlan string) string {
	out, err := exec.Command(
		"bash",
		"-c",
		vlanGatherCommand+vlan,
	).Output()
	if err != nil {
		log.Panic(err)
	}
	return string(out)
}

func publicKeysCallback() (
	signers []ssh.Signer, err error,
) {
	// Use public key auth for SSH if we can set it up...
	//
	// Find all of the public / private key pairs using the
	// ssh-keygen conventions. If someone wants to do something
	// more interesting, we might need to add an option to the
	// command.
	signers = []ssh.Signer{}
	publicKeys, err := filepath.Glob(os.Getenv("HOME") + "/.ssh/id_*.pub")
	if err != nil {
		// Error getting public keys. Warn the user then
		// return no error to let another auth method run.
		log.Printf(
			"Warning: failed to list keys, skipping public key auth - %s",
			err,
		)
		err = nil
		return
	}
	for _, publicKey := range publicKeys {
		// Quick and dirty convert public file name to private
		// using the ssh-keygen convention.
		privateKey := publicKey[:len(publicKey)-4]
		key, err := os.ReadFile(privateKey)
		if err != nil {
			// This should be a public key file, if it is
			// not there, skip it.
			log.Printf(
				"Warning: cannot open private key file '%s' (skipped) - %s",
				privateKey,
				err,
			)
			continue
		}
		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// Trouble parsing the private key, this is an
			// error so let the authentication process
			// fail on it so the user knows the key is
			// bad.
			log.Printf(
				"Warning: cannot parse private key file '%s' (skipped) - %s",
				privateKey,
				err,
			)
			continue
		}
		signers = append(
			signers,
			signer,
		)
	}
	// If this returns no error and signers has signers in it,
	// then they will be tried. If it returns no error and signers
	// is empty, we will move on to another authentication method.
	return
}

func passwordCallback() (
	ncnRootPassword string, err error,
) {
	// If this has already been called and we have not failed a
	// login attempt for this host with the cached password, just
	// return the cached password so that we don't have to prompt
	// for every NCN when we are using SSH.
	if sshPassword != "" && !sshPasswordRetry {
		// Make it possible to change the cached password if
		// it turns out to be incorrect for the current NCN by
		// requesting a password retry here. This will be
		// cleared for each new session by the caller.
		sshPasswordRetry = true
		ncnRootPassword = sshPassword
		err = nil
		return
	}
	// First (hopefully) successful time through, or after a
	// failed attempt using either a prompted for or cached
	// password, prompt for the password and then cache it for
	// future use. Note this will specify the host it is prompting
	// for in case the passwords differ across NCNs. Normally that
	// should not happen, but the retry logic permits it to be
	// recoverable if it ever does.
	fmt.Printf(
		"Enter root password for NCNs [%s]: ",
		sshPasswordHost,
	)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return
	}
	ncnRootPassword = string(bytePassword)
	sshPassword = ncnRootPassword
	// Make it possible to change the cached password if it turns
	// out to be incorrect for the current NCN by requesting a
	// password retry here. This will be cleared for each new
	// session by the caller.
	sshPasswordRetry = true
	return
}

func populateNCNMetadata() {
	bssEntries := make(map[string]bssTypes.BootParams)
	// Now we must build the kernel cmdline parameters for each NCN. The thing that's not so fun about this is those
	// are calculated as part of PXE booting, so there is no file we can reference as a source of truth. This means
	// we have no choice, we have to gather this from already booted NCNs and replace the values specific to each
	// node. As of the time of writing the only way to do this is to SSH to the thing and read the value directly.
	sshConfig = &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(publicKeysCallback),
			ssh.RetryableAuthMethod(
				ssh.PasswordCallback(passwordCallback),
				5,
			),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var hsmComponents base.ComponentArray

	// Loop over the management NCNs building their necessary configs.
	for _, ncn := range managementNCNs {
		hostname := getNCNHostname(ncn)

		// BSS entry.
		bssEntry := getBSSEntryForNCN(ncn)
		bssEntries[hostname] = bssEntry

		uploadEntryToBSS(
			bssEntry,
			http.MethodPut,
		)

		// In the event that for whatever reason HSM inventory doesn't work the component won't get created and then
		// BSS won't be able to find the node.
		var extraProperties slsCommon.ComptypeNode
		_ = mapstructure.Decode(
			ncn.ExtraPropertiesRaw,
			&extraProperties,
		)

		true := true
		component := base.Component{
			ID:      ncn.Xname,
			Type:    "Node",
			Flag:    "OK",
			State:   "Ready",
			Enabled: &true,
			Role:    extraProperties.Role,
			SubRole: extraProperties.SubRole,
			NID:     json.Number(strconv.Itoa(extraProperties.NID)),
			NetType: "Sling",
			Arch:    "X86",
			Class:   "River",
		}

		hsmComponents.Components = append(
			hsmComponents.Components,
			&component,
		)
	}

	// At this point we have all but m001 fully populated. m001 has a good cloud-init, but we need to update its
	// cmdline arguments. To do that we're going to "borrow" the arguments from its neighbor (m002) and update
	// anything specific to m001 using information we get from the system.
	pitArgs := buildPITArgs(bssEntries["ncn-m002"].Params)

	// Last thing for m001 is to add all the other MACs that might not be discoverable via Redfish.
	pitMACs := getAllPITMACs()

	// Finally update the structure with these values and send the whole package off to BSS.
	pitEntry := bssEntries["ncn-m001"]
	pitEntry.Params = pitArgs
	pitEntry.Macs = pitMACs
	bssEntries["ncn-m001"] = pitEntry

	uploadEntryToBSS(
		pitEntry,
		http.MethodPut,
	)

	// Create a component in HSM for each NCN. This should happen _eventually_ with discovery, but we might need
	// it sooner than that.
	uploadHSMComponents(hsmComponents)

	// To support PXE booting from any MAC from any node make sure that an EthernetInterfaces entry exists for them all.
	for _, entry := range bssEntries {
		for _, macAddr := range entry.Macs {
			compEthInterface := getCompEthInterfaceForMAC(macAddr)
			xname := entry.Hosts[0]

			if compEthInterface == nil {
				// MAC isn't in EthernetInterfaces, add it.
				generateAndSendInterfaceForNCN(
					xname,
					[]sm.IPAddressMapping{},
					macAddr,
					"CSI Handoff MAC",
				)
			} else {
				// So the MAC exists, the only other thing we care about is the ComponentID being correct.
				compEthInterface.CompID = xname

				// Be sure to normalize all the MACs.
				macWithoutPunctuation := strings.ReplaceAll(
					macAddr,
					":",
					"",
				)

				url := fmt.Sprintf(
					"%s/hsm/v2/Inventory/EthernetInterfaces/%s",
					hsmBaseURL,
					macWithoutPunctuation,
				)
				response := uploadCompEthInterfaceToHSM(
					*compEthInterface,
					url,
					"PATCH",
				)

				if response.StatusCode != http.StatusOK {
					log.Panicf(
						"Unexpected status code (%d): %s.",
						response.StatusCode,
						response.Status,
					)
				}
			}
		}
	}
}

func populateGlobalMetadata() {
	globalData := cloudInitData["Global"]

	bssEntry := bssTypes.BootParams{
		Hosts: []string{"Global"},
		CloudInit: bssTypes.CloudInit{
			MetaData: globalData.MetaData,
			UserData: globalData.UserData,
		},
	}

	uploadEntryToBSS(
		bssEntry,
		http.MethodPatch,
	)
}

func getCompEthInterfaceForMAC(macAddr string) *sm.CompEthInterfaceV2 {
	// Be sure to normalize all the MACs.
	macWithoutPunctuation := strings.ReplaceAll(
		macAddr,
		":",
		"",
	)

	url := fmt.Sprintf(
		"%s/hsm/v2/Inventory/EthernetInterfaces/%s",
		hsmBaseURL,
		macWithoutPunctuation,
	)

	request, requestErr := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if requestErr != nil {
		log.Panicf(
			"Failed to construct request: %s",
			requestErr,
		)
	}
	request.Header.Add(
		"Authorization",
		fmt.Sprintf(
			"Bearer %s",
			token,
		),
	)

	response, doErr := httpClient.Do(request)
	if doErr != nil {
		log.Panicf(
			"Failed to execute POST request: %s",
			doErr,
		)
	}

	if response.StatusCode == http.StatusNotFound {
		return nil
	}

	responseBytes, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		log.Panicf(
			"Failed to read response body: %s",
			readErr,
		)
	}

	var compInterface sm.CompEthInterfaceV2
	unmarshalErr := json.Unmarshal(
		responseBytes,
		&compInterface,
	)
	if unmarshalErr != nil {
		log.Panicf(
			"Failed to unmarshal response bytes: %s",
			unmarshalErr,
		)
	}

	return &compInterface
}

func uploadCompEthInterfaceToHSM(
	compInterface sm.CompEthInterfaceV2, url string, method string,
) (response *http.Response) {
	payloadBytes, marshalErr := json.Marshal(compInterface)
	if marshalErr != nil {
		log.Panicf(
			"Failed to marshal HSM endpoint description: %s",
			marshalErr,
		)
	}

	request, requestErr := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(payloadBytes),
	)
	if requestErr != nil {
		log.Panicf(
			"Failed to construct request: %s",
			requestErr,
		)
	}
	request.Header.Add(
		"Authorization",
		fmt.Sprintf(
			"Bearer %s",
			token,
		),
	)
	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	var doErr error
	response, doErr = httpClient.Do(request)
	if doErr != nil {
		log.Panicf(
			"Failed to execute POST request: %s",
			doErr,
		)
	}

	jsonPrettyBytes, _ := json.MarshalIndent(
		compInterface,
		"",
		"\t",
	)
	log.Printf(
		"Successfully %s EthernetInterfaces entry for %s:\n%s",
		method,
		compInterface.CompID,
		string(jsonPrettyBytes),
	)

	return
}

func generateAndSendInterfaceForNCN(
	xname string, ips []sm.IPAddressMapping, macAddr string, description string,
) {
	url := fmt.Sprintf(
		"%s/hsm/v2/Inventory/EthernetInterfaces",
		hsmBaseURL,
	)

	// Be sure to normalize all the MACs.
	macWithoutPunctuation := strings.ReplaceAll(
		macAddr,
		":",
		"",
	)

	componentEndpointInterfaces := sm.CompEthInterfaceV2{
		ID:      macWithoutPunctuation,
		Desc:    description,
		MACAddr: macWithoutPunctuation,
		IPAddrs: ips,
		CompID:  xname,
		Type:    "Node",
	}

	response := uploadCompEthInterfaceToHSM(
		componentEndpointInterfaces,
		url,
		"POST",
	)

	if response.StatusCode == http.StatusConflict {
		// If we're in conflict (almost certain to not be since the reason we're doing this is because these NCNs
		// don't get into this table any other way), then PATCH the entry.
		patchURL := fmt.Sprintf(
			"%s/%s",
			url,
			macWithoutPunctuation,
		)

		response := uploadCompEthInterfaceToHSM(
			componentEndpointInterfaces,
			patchURL,
			"PATCH",
		)

		if response.StatusCode != http.StatusOK {
			log.Panicf(
				"Unexpected status code (%d): %s.",
				response.StatusCode,
				response.Status,
			)
		}
	} else if response.StatusCode != http.StatusCreated {
		log.Panicf(
			"Unexpected status code (%d): %s",
			response.StatusCode,
			response.Status,
		)
	}
}

func populateHSMEthernetInterface(
	xname string, ipString string, vlan string,
) {
	// The input here will be a JSON blob in text form. So we will need to unmarshal and pick out the pieces we need.
	var ipStructArray ipJSONStructArray

	err := json.Unmarshal(
		[]byte(ipString),
		&ipStructArray,
	)
	if err != nil {
		log.Panic(err)
	}

	vlanInterface := ipStructArray[0]

	var ip string
	for _, addr := range vlanInterface.AddrInfo {
		if addr.Family == "inet" {
			ip = addr.Local
			break
		}
	}
	mac := vlanInterface.Address

	description := fmt.Sprintf(
		"Bond0 - %s",
		vlan,
	)

	ips := []sm.IPAddressMapping{
		{
			IPAddr: ip,
		},
	}

	generateAndSendInterfaceForNCN(
		xname,
		ips,
		mac,
		description,
	)
}

func uploadHSMComponents(array base.ComponentArray) {
	url := fmt.Sprintf(
		"%s/hsm/v2/State/Components",
		hsmBaseURL,
	)

	payloadBytes, marshalErr := json.MarshalIndent(
		array,
		"",
		"\t",
	)
	if marshalErr != nil {
		log.Panicf(
			"Failed to marshal component: %s",
			marshalErr,
		)
	}

	request, requestErr := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(payloadBytes),
	)
	if requestErr != nil {
		log.Panicf(
			"Failed to construct request: %s",
			requestErr,
		)
	}
	request.Header.Add(
		"Authorization",
		fmt.Sprintf(
			"Bearer %s",
			token,
		),
	)
	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response, doErr := httpClient.Do(request)
	if doErr != nil {
		log.Panicf(
			"Failed to execute POST request: %s",
			doErr,
		)
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		log.Panicf(
			"unexpected status code (%d): %s.",
			response.StatusCode,
			response.Status,
		)
	}

	log.Printf(
		"Successfully put Components array:\n%s",
		string(payloadBytes),
	)
}
