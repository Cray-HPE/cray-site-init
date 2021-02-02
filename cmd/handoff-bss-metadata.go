/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	base "stash.us.cray.com/HMS/hms-base"
	"stash.us.cray.com/HMS/hms-bss/pkg/bssTypes"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	"stash.us.cray.com/HMS/hms-smd/pkg/sm"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"strconv"
	"strings"
	"syscall"
)

const gatewayHostname = "api-gw-service-nmn.local"
const s3Prefix = "s3://ncn-images/"
const hwAddrPrefix = "Permanent HW addr: "
const macGatherCommand = "for interface in /sys/class/net/*; do echo -n " +
	"\"$interface,\" && cat \"$interface/address\"; done"
const vlanGatherCommand = "ip -j addr show "

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
	dataFile       string
	cloudInitData  map[string]bssTypes.CloudInit
	sshConfig      *ssh.ClientConfig
	token          string
	httpClient     *http.Client
	managementNCNs []sls_common.GenericHardware

	vlansToGather = []string{"vlan002", "vlan004", "vlan007"}
)

var handoffBSSMetadataCmd = &cobra.Command{
	Use:   "bss-metadata",
	Short: "runs migration steps to build BSS entries for all NCNs",
	Long:  "Using PIT configuration builds kernel command line arguments and cloud-init metadata for each NCN",
	Run: func(cmd *cobra.Command, args []string) {
		token = os.Getenv("TOKEN")
		if token == "" {
			log.Panicln("Environment variable TOKEN can NOT be blank!")
		}

		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		httpClient = &http.Client{Transport: transport}

		// Parse the data.json file.
		err := csiFiles.ReadJSONConfig(dataFile, &cloudInitData)
		if err != nil {
			log.Fatalln("Couldn't parse data file: ", err)
		}

		log.Println("Building BSS metadata for NCNs...")
		populateNCNMetadata()
		log.Println("Done building BSS metadata for NCNs.")

		log.Println("Transferring global cloud-init metadata to BSS...")
		populateGlobalMetadata()
		log.Println("Done transferring global cloud-init metadata to BSS.")
	},
}

func init() {
	handoffCmd.AddCommand(handoffBSSMetadataCmd)

	handoffBSSMetadataCmd.Flags().StringVar(&dataFile, "data-file",
		"", "data.json file with cloud-init configuration for each node and global")
	_ = handoffBSSMetadataCmd.MarkFlagRequired("data-file")
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

func getKernelCommandlineArgs(ncn sls_common.GenericHardware, cmdline string) string {
	var extraProperties sls_common.ComptypeNode
	_ = mapstructure.Decode(ncn.ExtraPropertiesRaw, &extraProperties)

	cmdlineParts := strings.Fields(cmdline)

	for i := range cmdlineParts {
		part := cmdlineParts[i]

		if strings.HasPrefix(part, "metal.server") {
			cmdlineParts[i] = fmt.Sprintf("metal.server=http://rgw-vip.nmn/ncn-images")
		} else if strings.HasPrefix(part, "ds=nocloud-net") {
			// The advertise address is fed into BSS as part of its deployment and automatically added to the cmdline
			// with the correct value, so just remove it here.
			cmdlineParts[1] = ""
			//cmdlineParts[i] = fmt.Sprintf("ds=nocloud-net;s=http://10.92.100.81:8888/")
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
		} else if strings.HasPrefix(part, "xname") {
			// BSS sets the xname, no need to have it here.
			cmdlineParts[i] = ""
		}
	}

	// Get rid of any empty values.
	var finalParts []string
	for _, part := range cmdlineParts {
		if part != "" {
			finalParts = append(finalParts, part)
		}
	}

	return strings.Join(finalParts, " ")
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

func getSSHClientForHostname(hostname string) (sshClient *ssh.Client) {
	var err error

	// Don't try to SSH to ourself.
	if hostname == "ncn-m001" {
		return
	}

	log.Printf("Connecting to %s...", hostname)

	sshClient, err = ssh.Dial("tcp", hostname+":22", sshConfig)
	if err != nil {
		log.Panic(err)
	}

	return
}

func runSSHCommandWithClient(sshClient *ssh.Client, command string) (output string) {
	if sshClient == nil {
		return ""
	}

	log.Printf("Creating session to %s...", sshClient.RemoteAddr())

	sshSession, err := sshClient.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer sshSession.Close()

	cmdline, err := sshSession.CombinedOutput(command)
	if err != nil {
		log.Panic(err)
	}

	return string(cmdline)
}

func getCloudInitMetadataForNCN(ncn sls_common.GenericHardware) (userData bssTypes.CloudDataType,
	metaData bssTypes.CloudDataType) {
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
	// In the macAddrString above is a block of a bunch of MACs so now we need to filter out only the ones we care
	// about.
	macAddrsSplit := strings.Split(macAddrStrings, "\n")
	for _, macAddrLine := range macAddrsSplit {
		// We have a range of entries that look like this:
		//   /sys/class/net/bond0,14:02:ec:d9:7c:40
		// So now we split once again to get the name and the MAC.
		macPieces := strings.Split(macAddrLine, ",")
		if len(macPieces) != 2 {
			// Ignore anything that doesn't for any reason have both pieces.
			continue
		}

		interfaceName := strings.TrimPrefix(macPieces[0], "/sys/class/net/")
		macAddrString := macPieces[1]

		// We want to ignore a fair number of interfaces.
		if interfaceName == "bonding_masters" ||
			interfaceName == "dummy" ||
			interfaceName == "lo" ||
			strings.Contains(interfaceName, "usb") ||
			strings.Contains(interfaceName, "veth") {
			continue
		}

		// No reason for this to not work, but, might as well really double check this MAC.
		hw, err := net.ParseMAC(macAddrString)
		if err != nil {
			continue
		}

		ncnMacs = append(ncnMacs, hw.String())
	}

	return
}

func populateHSMEthernetInterface(xname string, ipString string, vlan string) {
	// The input here will be a JSON blob in text form. So we will need to unmarshal and pick out the pieces we need.
	var ipStructArray ipJSONStructArray

	err := json.Unmarshal([]byte(ipString), &ipStructArray)
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

	generateAndSendInterfaceForNCN(xname, ip, mac, vlan)
}

func getBSSEntryForNCN(ncn sls_common.GenericHardware) (bssEntry bssTypes.BootParams) {
	hostname := getNCNHostname(ncn)

	var extraProperties sls_common.ComptypeNode
	_ = mapstructure.Decode(ncn.ExtraPropertiesRaw, &extraProperties)

	// Setup an SSH connection for those who need it.
	sshClient := getSSHClientForHostname(hostname)

	cmdline := runSSHCommandWithClient(sshClient, "cat /proc/cmdline")
	macAddrStrings := runSSHCommandWithClient(sshClient, macGatherCommand)

	// This is not even related to BSS but it makes the most sense to do here...we need to make sure HSM has correct
	// EthernetInterface entries for all the NCNs and since they don't DHCP Kea won't do it for us. So take advantage
	// of the fact we're already in here running commands and gather/populate those entries in HSM.
	for _, vlan := range vlansToGather {
		var vlanOutputString string

		// TODO: This is hacky and I don't like it.
		if hostname == "ncn-m001" {
			vlanOutputString = getPITVLanString(vlan)
		} else {
			vlanOutputString = runSSHCommandWithClient(sshClient, vlanGatherCommand + vlan)
		}
		populateHSMEthernetInterface(ncn.Xname, vlanOutputString, vlan)
	}

	userData, metaData := getCloudInitMetadataForNCN(ncn)

	// Close the SSH connection.
	if sshClient != nil {
		_ = sshClient.Close()
	}

	// K8s and CEPH nodes have different kernel/initrd's.
	var kernel string
	var initrd string

	if extraProperties.SubRole == "Storage" {
		kernel = cephKernelName
		initrd = cephInitrdName
	} else {
		kernel = k8sKernelName
		initrd = k8sInitrdName
	}

	// Now we can build the BSS structure.
	// TODO: Put in MAC address from bond.
	bssEntry = bssTypes.BootParams{
		Hosts:  []string{ncn.Xname},
		Macs:   getMACsFromString(macAddrStrings),
		Params: getKernelCommandlineArgs(ncn, cmdline),
		Kernel: s3Prefix + kernel,
		Initrd: s3Prefix + initrd,
		CloudInit: bssTypes.CloudInit{
			MetaData: metaData,
			UserData: userData,
		},
	}

	return
}

func buildPITArgs(base string) string {
	// Gather the MACs we need.
	var macs []string

	data, err := ioutil.ReadFile("/proc/net/bonding/bond0")
	if err != nil {
		log.Panicln(err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, hwAddrPrefix) {
			macs = append(macs, strings.TrimPrefix(line, hwAddrPrefix))
		}
	}

	// Hopefully future proofing against if we have 2 bonds.
	data, err = ioutil.ReadFile("/proc/net/bonding/bond1")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, hwAddrPrefix) {
				macs = append(macs, strings.TrimPrefix(line, hwAddrPrefix))
			}
		}
	}

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
		if strings.HasPrefix(part, "hostname") {
			cmdlineParts[i] = fmt.Sprintf("hostname=ncn-m001")
		} else if strings.HasPrefix(part, "ifname=mgmt0") && len(macs) >= 2 {
			cmdlineParts[i] = fmt.Sprintf("ifname=mgmt0:%s", macs[0])
		} else if strings.HasPrefix(part, "ifname=mgmt1") && len(macs) >= 2 {
			cmdlineParts[i] = fmt.Sprintf("ifname=mgmt1:%s", macs[1])
		} else if strings.HasPrefix(part, "ifname=mgmt2") {
			if len(macs) >= 4 {
				cmdlineParts[i] = fmt.Sprintf("ifname=mgmt2:%s", macs[2])
			} else {
				cmdlineParts[i] = ""
			}
		} else if strings.HasPrefix(part, "ip=mgmt3") && len(macs) < 4 {
			cmdlineParts[i] = ""
		} else if strings.HasPrefix(part, "ifname=mgmt3") {
			if len(macs) >= 4 {
				cmdlineParts[i] = fmt.Sprintf("ifname=mgmt3:%s", macs[3])
			} else {
				cmdlineParts[i] = ""
			}
		} else if strings.HasPrefix(part, "ip=mgmt2") && len(macs) < 4 {
			cmdlineParts[i] = ""
		} else if strings.HasPrefix(part, "bond") {
			if len(macs) == 2 {
				cmdlineParts[i] = strings.ReplaceAll(part, "mgmt2", "mgmt1")
			}
		}
	}

	// Just to get rid of the whitespace.
	var finalParts []string
	for _, part := range cmdlineParts {
		if part != "" {
			finalParts = append(finalParts, part)
		}
	}

	return strings.Join(finalParts, " ")
}

func getPITMACs() []string {
	out, err := exec.Command("bash", "-c", macGatherCommand).Output()
	if err != nil {
		log.Panic(err)
	}
	return getMACsFromString(string(out))
}

func getPITVLanString(vlan string) string {
	out, err := exec.Command("bash", "-c", vlanGatherCommand + vlan).Output()
	if err != nil {
		log.Panic(err)
	}
	return string(out)
}

func populateNCNMetadata() {
	var err error
	var ncnRootPassword string

	bssEntries := make(map[string]bssTypes.BootParams)

	managementNCNs, err = getManagementNCNsFromSLS()
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

	var hsmComponents base.ComponentArray

	// Loop over the management NCNs building their necessary configs.
	for _, ncn := range managementNCNs {
		hostname := getNCNHostname(ncn)

		// BSS entry.
		bssEntry := getBSSEntryForNCN(ncn)
		bssEntries[hostname] = bssEntry

		uploadEntryToBSS(bssEntry)

		// In the event that for whatever reason HSM inventory doesn't work the component won't get created and then
		// BSS won't be able to find the node.
		var extraProperties sls_common.ComptypeNode
		_ = mapstructure.Decode(ncn.ExtraPropertiesRaw, &extraProperties)

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

		hsmComponents.Components = append(hsmComponents.Components, &component)
	}

	// At this point we have all but m001 fully populated. m001 has a good cloud-init, but we need to update its
	// cmdline arguments. To do that we're going to "borrow" the arguments from its neighbor (m002) and update
	// anything specific to m001 using information we get from the system.
	pitArgs := buildPITArgs(bssEntries["ncn-m002"].Params)

	// Last thing for m001 is to add all the other MACs that might not be discoverable via Redfish.
	pitMACs := getPITMACs()

	// Finally update the structure whith these values and send the whole package off to BSS.
	pitEntry := bssEntries["ncn-m001"]
	pitEntry.Params = pitArgs
	pitEntry.Macs = pitMACs

	uploadEntryToBSS(pitEntry)

	uploadHSMComponents(hsmComponents)
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

	uploadEntryToBSS(bssEntry)
}

func uploadEntryToBSS(bssEntry bssTypes.BootParams) {
	url := fmt.Sprintf("https://%s/apis/bss/boot/v1/bootparameters", gatewayHostname)

	jsonBytes, err := json.Marshal(bssEntry)
	if err != nil {
		log.Panicln(err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Panicf("Failed to create new request: %s", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Panicf("Failed to put BSS entry: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		log.Panicf("Failed to put BSS entry: %s", string(bodyBytes))
	}

	jsonPrettyBytes, _ := json.MarshalIndent(bssEntry, "", "\t")

	log.Printf("Sucessfuly put BSS entry for %s:\n%s", bssEntry.Hosts[0], string(jsonPrettyBytes))
}

func uploadCompEthInterfaceToHSM(compInterface sm.CompEthInterface, url string,
	method string) (response *http.Response) {
	payloadBytes, marshalErr := json.Marshal(compInterface)
	if marshalErr != nil {
		log.Panicf("Failed to marshal HSM endpoint description: %s", marshalErr)
	}

	request, requestErr := http.NewRequest(method, url, bytes.NewBuffer(payloadBytes))
	if requestErr != nil {
		log.Panicf("Failed to construct request: %s", requestErr)
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")

	var doErr error
	response, doErr = httpClient.Do(request)
	if doErr != nil {
		log.Panicf("Failed to execute POST request: %s", doErr)
	}

	return
}

func generateAndSendInterfaceForNCN(xname string, ip string, macAddr string, vlan string) {
	url := fmt.Sprintf("https://%s/apis/smd/hsm/v1/Inventory/EthernetInterfaces", gatewayHostname)

	// Be sure to normalize all the MACs.
	macWithoutPunctuation := strings.ReplaceAll(macAddr, ":", "")

	componentEndpointInterfaces := sm.CompEthInterface{
		ID:      macWithoutPunctuation,
		Desc:    fmt.Sprintf("Bond0 - %s", vlan),
		MACAddr: macWithoutPunctuation,
		IPAddr:  ip,
		CompID:  xname,
		Type:    "Node",
	}

	response := uploadCompEthInterfaceToHSM(componentEndpointInterfaces, url, "POST")

	if response.StatusCode == http.StatusConflict {
		// If we're in conflict (almost certain to not be since the reason we're doing this is because these NCNs
		// don't get into this table any other way), then PATCH the entry.
		patchURL := fmt.Sprintf("%s/%s", url, macWithoutPunctuation)

		response := uploadCompEthInterfaceToHSM(componentEndpointInterfaces, patchURL, "PATCH")

		if response.StatusCode != http.StatusOK {
			log.Panicf("unexpected status code (%d): %s.", response.StatusCode, response.Status)
		}
	} else if response.StatusCode != http.StatusCreated {
		log.Panicf("unexpected status code (%d): %s", response.StatusCode, response.Status)
	}

	jsonPrettyBytes, _ := json.MarshalIndent(componentEndpointInterfaces, "", "\t")

	log.Printf("Sucessfuly put EthernetInterfaces entry for %s:\n%s", xname, string(jsonPrettyBytes))
}

func uploadHSMComponents(array base.ComponentArray) {
	url := fmt.Sprintf("https://%s/apis/smd/hsm/v1/State/Components", gatewayHostname)

	payloadBytes, marshalErr := json.MarshalIndent(array, "", "\t")
	if marshalErr != nil {
		log.Panicf("Failed to marshal component: %s", marshalErr)
	}

	request, requestErr := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if requestErr != nil {
		log.Panicf("Failed to construct request: %s", requestErr)
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")

	response, doErr := httpClient.Do(request)
	if doErr != nil {
		log.Panicf("Failed to execute POST request: %s", doErr)
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		log.Panicf("unexpected status code (%d): %s.", response.StatusCode, response.Status)
	}

	log.Printf("Sucessfuly put Components array:\n%s", string(payloadBytes))
}
