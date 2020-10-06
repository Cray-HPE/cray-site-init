/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	"stash.us.cray.com/MTL/sic/pkg/ipam"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init --from-sls-file=<file> --from-1.3-dir=<dir> --from-sls=<sls url> system-name",
	Short: "init generates the directory structure for a new system rooted in a directory matching the system-name argument",
	Long: `init generates a scaffolding the Shasta 1.4 configuration payload.  
	It can optionally load from the sls or files of an existing 1.3 system.
	Files used from the 1.3 system include:
	 - system_config.yml
	 - ncn_metadata.csv
	 - networks_derived.yml
	 - hmn_connections.json
	 - customer_var.yml`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		v := viper.GetViper()
		var conf shasta.SystemConfig
		// if the user has an old configuration dir, use that
		if v.GetString("from-1.3-dir") != "" {
			MergeNCNMetadata()
			MergeNetworksDerived()
			MergeSLSInput()
			MergeCustomerVar()
		}
		// We use the system-name for a directory.  Make sure it is set.
		if v.GetString("system-name") == "" {
			log.Fatalf("system-name is not set")
		}
		// Set up the path for our base directory using our systemname
		basepath, err := filepath.Abs(filepath.Clean(v.GetString("system-name")))
		if err != nil {
			log.Fatalln(err)
		}
		// Create our base directory
		if err := os.Mkdir(basepath, 0777); err != nil {
			log.Fatalln("Can't create directory", basepath, err)
		}

		// These Directories make up the overall structure for the Configuration Payload
		dirs := []string{
			filepath.Join(basepath, "networks"),
			filepath.Join(basepath, "manufacturing"),
			filepath.Join(basepath, "credentials"),
			filepath.Join(basepath, "certificates"),
		}
		// Add the Manifest directory if needed
		if v.GetString("manifest-release") != "" {
			dirs = append(dirs, filepath.Join(basepath, "loftsman-manifests"))
		}
		// Iterate through the directories and create them
		for _, dir := range dirs {
			if err := os.Mkdir(dir, 0777); err != nil {
				log.Fatalln("Can't create directory", dir, err)
			}
		}

		// Handle an SLSFile if one is provided
		var slsState sls_common.SLSState
		if v.GetString("sls-file-path") != "" {
			slsState = loadFromSLS("file://" + v.GetString("sls-file-path"))
		} else if v.GetString("sls-url") != "" {
			slsState = loadFromSLS(v.GetString("sls-url"))
		}

		networks := shasta.ConvertSLSNetworks(slsState)
		v.Set("networks.from_sls", networks)
		// shasta.ExtractSwitches(slsState)  // Uncomment this when it can actually produce the right switch configuration(s) here

		err = v.Unmarshal(&conf)
		if err != nil {
			log.Fatalf("unable to decode configuration into struct, %v \n", err)
		}

		conf.IPV4Resolvers = strings.Split(viper.GetString("ipv4-resolvers"), ",")
		conf.SiteServices.NtpPoolHostname = conf.NtpPoolHostname

		WriteSystemConfig(filepath.Join(basepath, "system_config.yaml"), conf)

		// our primitive ipam uses the number of cabinets to lay out a network for each one.
		var cabinets = uint(conf.MountainCabinets)

		// Merge configs with the NMN Defaults to create a yaml with our subnets in it
		shasta.DefaultNMN.CIDR = v.GetString("nmn-cidr")
		_, myNet, _ := net.ParseCIDR(shasta.DefaultNMN.CIDR)
		nmnSubnets, err := ipam.Split(*myNet, 32) // 32 allows for a /22 per cabinet
		for k, v := range nmnSubnets[0:cabinets] {
			shasta.DefaultNMN.Subnets = append(shasta.DefaultNMN.Subnets, shasta.IPV4Subnet{
				CIDR:    v,
				Name:    fmt.Sprintf("cabinet_%v_nmn", int(conf.StartingCabinet)+k),
				Gateway: ipam.Add(v.IP, 1),
				VlanID:  shasta.DefaultNMN.VlanRange[0] + int16(k),
			})
		}
		WriteNetworkConfig(filepath.Join(basepath, "networks/nmn.yaml"), shasta.DefaultNMN)

		// Merge configs with the HMN Defaults to create a yaml with our subnets in it
		shasta.DefaultHMN.CIDR = v.GetString("hmn-cidr")
		_, myNet, _ = net.ParseCIDR(shasta.DefaultHMN.CIDR)
		hmnSubnets, err := ipam.Split(*myNet, 32) // 32 allows for a /22 per cabinet
		for k, v := range hmnSubnets[0:cabinets] {
			shasta.DefaultHMN.Subnets = append(shasta.DefaultHMN.Subnets, shasta.IPV4Subnet{
				CIDR:    v,
				Name:    fmt.Sprintf("cabinet_%v_hmn", int(conf.StartingCabinet)+k),
				Gateway: ipam.Add(v.IP, 1),
				VlanID:  shasta.DefaultHMN.VlanRange[0] + int16(k),
			})
		}
		WriteNetworkConfig(filepath.Join(basepath, "networks/hmn.yaml"), shasta.DefaultHMN)

		// Merge configs with the HSN Defaults to create a yaml with our subnets in it
		shasta.DefaultHSN.CIDR = v.GetString("hsn-cidr")
		_, myNet, _ = net.ParseCIDR(shasta.DefaultHSN.CIDR)
		hsnSubnets, err := ipam.Split(*myNet, 32) // 32 allows for a /22 per cabinet
		for k, v := range hsnSubnets[0:cabinets] {
			shasta.DefaultHSN.Subnets = append(shasta.DefaultHSN.Subnets, shasta.IPV4Subnet{
				CIDR:    v,
				Name:    fmt.Sprintf("cabinet_%v_hsn", int(conf.StartingCabinet)+k),
				Gateway: ipam.Add(v.IP, 1),
				VlanID:  shasta.DefaultHMN.VlanRange[0] + int16(k),
			})
		}
		WriteNetworkConfig(filepath.Join(basepath, "networks/hsn.yaml"), shasta.DefaultHSN)

		shasta.DefaultMTL.CIDR = v.GetString("mtl-cidr")
		WriteNetworkConfig(filepath.Join(basepath, "networks/mtl.yaml"), shasta.DefaultMTL)
		shasta.DefaultCAN.CIDR = v.GetString("can-cidr")
		WriteNetworkConfig(filepath.Join(basepath, "networks/can.yaml"), shasta.DefaultCAN)
		WritePasswordCredential(filepath.Join(basepath, "credentials/root_password.json"), shasta.DefaultRootPW)
		WritePasswordCredential(filepath.Join(basepath, "credentials/bmc_password.json"), shasta.DefaultBMCPW)
		WritePasswordCredential(filepath.Join(basepath, "credentials/mgmt_switch_password.json"), shasta.DefaultNetPW)

		if v.GetString("manifest-release") != "" {
			initiailzeManifestDir(shasta.DefaultManifestURL, "release/shasta-1.4", filepath.Join(basepath, "loftsman-manifests"))
		}
	},
}

func init() {
	configCmd.AddCommand(initCmd)

	// Flags with defaults for initializing a configuration

	// Flags to deal with 1.3 configuration directories

	initCmd.Flags().String("from-1.3-dir", "", "Shasta 1.3 Configuration Directory")

	// System Configuration Flags based on previous system_config.yml and networks_derived.yml
	initCmd.Flags().String("system-name", "sn-2024", "Name of the System")
	initCmd.Flags().String("site-domain", "dev.cray.com", "Site Domain Name")
	initCmd.Flags().String("internal-domain", "unicos.shasta", "Internal Domain Name")
	initCmd.Flags().String("ntp-pool", "time.nist.gov", "Hostname for Upstream NTP Pool")
	initCmd.Flags().String("ipv4-resolvers", "8.8.8.8, 9.9.9.9", "List of IP Addresses for DNS")
	initCmd.Flags().String("v2-registry", "https://packages.local/", "URL for default v2 registry used for both helm and containers")
	initCmd.Flags().String("rpm-repository", "https://packages.local/repository/shasta-master", "URL for default rpm repository")

	// Default IPv4 Networks
	initCmd.Flags().String("nmn-cidr", shasta.DefaultNMNString, "Overall IPv4 CIDR for all Node Management subnets")
	initCmd.Flags().String("hmn-cidr", shasta.DefaultHMNString, "Overall IPv4 CIDR for all Hardware Management subnets")
	initCmd.Flags().String("can-cidr", shasta.DefaultCANString, "Overall IPv4 CIDR for all Customer Access subnets")
	initCmd.Flags().String("mtl-cidr", shasta.DefaultMTLString, "Overall IPv4 CIDR for all Provisioning subnets")
	initCmd.Flags().String("hsn-cidr", shasta.DefaultHSNString, "Overall IPv4 CIDR for all HSN subnets")

	// Hardware Details
	initCmd.Flags().Int16("mountain-cabinets", 5, "Number of Mountain Cabinets")
	initCmd.Flags().Int16("starting-cabinet", 1004, "Starting ID number for Mountain Cabinets")
	initCmd.Flags().Int16("starting-NID", 20000, "Starting NID for Compute Nodes")
	// Use these flags to prepare the basecamp metadata json
	initCmd.Flags().String("switch-xnames", "", "Comma separated list of xnames for management switches")

	// Dealing with an SLS file
	initCmd.Flags().String("from-sls-file", "", "SLS File Location")
	initCmd.Flags().String("from-sls", "", "Shasta 1.3 SLS dumpstate url")
	// Loftsman Manifest Shasta-CFG
	initCmd.Flags().String("manifest-release", "", "Loftsman Manifest Release Version (leave blank to prevent manifest generation)")
	initCmd.Flags().SortFlags = false
}

func loadFromSLS(source string) sls_common.SLSState {
	var slsState sls_common.SLSState
	var err error
	if strings.HasPrefix(source, "file://") {
		slsState, err = shasta.ParseSLSFile(strings.TrimPrefix(source, "file://"))
		if err != nil {
			log.Fatalln(err)
		}
	} else if strings.HasPrefix(source, "https://") {
		slsState, err = shasta.ParseSLSfromURL(strings.TrimPrefix(source, "https://"))
		if err != nil {
			log.Fatalln(err)
		}
	}
	// Testing that slsState is valid
	if unsafe.Sizeof(slsState) > 0 {
		// At this point, we should have a valid slsState
		ncns, err := shasta.ExtractNCNBMCInfo(slsState)
		if err != nil {
			log.Fatalln("Unable to extract the NCNs from", source, err)
		}
		log.Println("The NCNs are:", ncns) // TODO: Replace this with a call to write the NCN file(s) for basecamp and/or initialization
	}
	return slsState
}

func writeFile(path string, contents string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	size, err := w.WriteString(contents)
	if err != nil {
		return err
	}
	w.Flush()
	log.Printf("wrote %d bytes to %s\n", size, path)
	return nil
}

// WriteSystemConfig applies a SystemConfig Struct to the Yaml Template and writes the result to the path indicated
func WriteSystemConfig(path string, conf shasta.SystemConfig) error {
	bs, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	err = writeFile(path, string(bs))
	return err
}

// WriteNetworkConfig applies a IPV4Network Struct to the Yaml Template and writes the result to the path indicated
func WriteNetworkConfig(path string, network shasta.IPV4Network) error {
	bs, err := yaml.Marshal(network)
	if err != nil {
		return err
	}
	err = writeFile(path, string(bs))
	return err
}

// WritePasswordCredential applies a PasswordCredential to the Yaml Template and writes the result to the path indicated
func WritePasswordCredential(path string, pw shasta.PasswordCredential) error {
	creds, _ := json.Marshal(pw)
	return ioutil.WriteFile(path, creds, 0644)
}

// WriteDNSMasqConfig writes the dnsmasq configuration files necssary for installation
func WriteDNSMasqConfig(path string, conf shasta.SystemConfig) {
	log.Printf("NOT IMPLEMENTED")
	// include the statics.conf stuff too
}

// WriteNICConfigENV sets environment variables for nic bonding and configuration
func WriteNICConfigENV(path string, conf shasta.SystemConfig) {
	log.Printf("NOT IMPLEMENTED")
}

// WriteBaseCampData writes basecamp data.json for the installer
func WriteBaseCampData(path string, conf shasta.SystemConfig) {
	// https://stash.us.cray.com/projects/MTL/repos/docs-non-compute-nodes/browse/example-data.json
	/* Funky vars from the stopgap
	export site_nic=em1
	export bond_member0=p801p1
	export bond_member1=p801p2
	export mtl_cidr=10.1.1.1/16
	export mtl_dhcp_start=10.1.2.3
	export mtl_dhcp_end=10.1.2.254
	export nmn_cidr=10.252.0.4/17
	export nmn_dhcp_start=10.252.50.0
	export nmn_dhcp_end=10.252.99.252
	export hmn_cidr=10.254.0.4/17
	export hmn_dhcp_start=10.254.50.5
	export hmn_dhcp_end=10.254.99.252
	export site_cidr=172.30.52.220/20
	export site_gw=172.30.48.1
	export site_dns='172.30.84.40 172.31.84.40'
	export can_cidr=10.102.4.110/24
	export can_dhcp_start=10.102.4.5
	export can_dhcp_end=10.102.4.109
	export dhcp_ttl=2m
	*/
}

// WriteConmanConfig provides conman configuration for the installer
func WriteConmanConfig(path string, conf shasta.SystemConfig) {
	log.Printf("NOT IMPLEMENTED")
}

func initiailzeManifestDir(url, branch, destination string) {
	// First we need a temporary directory
	dir, err := ioutil.TempDir("", "loftsman-init")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.RemoveAll(dir)
	cloneCmd := exec.Command("git", "clone", url, dir)
	out, err := cloneCmd.Output()
	if err != nil {
		log.Fatalf("cloneCommand finished with error: %s (%v)\n", out, err)
	}
	checkoutCmd := exec.Command("git", "checkout", branch)
	checkoutCmd.Dir = dir
	out, err = checkoutCmd.Output()
	if err != nil {
		if err.Error() != "exit status 1" {
			log.Fatalf("checkoutCommand finished with error: %s (%v)\n", out, err)
		}
	}
	packageCmd := exec.Command("./package/package.sh", "1.4.0")
	packageCmd.Dir = dir
	out, err = packageCmd.Output()
	if err != nil {
		log.Fatalf("packageCommand finished with error: %s (%v)\n", out, err)
	}
	targz, err := filepath.Abs(filepath.Clean(dir + "/dist/shasta-cfg-1.4.0.tgz"))
	untarCmd := exec.Command("tar", "-zxvvf", targz)
	untarCmd.Dir = destination
	out, err = untarCmd.Output()
	if err != nil {
		log.Fatalf("untarCmd finished with error: %s (%v)\n", out, err)
	}
}
