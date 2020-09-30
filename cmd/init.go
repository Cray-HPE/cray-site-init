/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
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
		// viperWiper(v)
		var conf shasta.SystemConfig

		// LoadConfig()
		// fmt.Println("Initial Config Loaded")
		// After processing all the flags in init,
		// if the user has an old configuration dir, use that
		if v.GetString("from-1.3-dir") != "" {
			// LoadConfig()
			MergeNCNMetadata()
			MergeNetworksDerived()
			MergeSLSInput()
			MergeCustomerVar()
		}

		if v.GetString("system-name") == "" {
			fmt.Println("system-name is not set")
			os.Exit(1)
		}
		// Set up the path for our base directory using our systemname
		basepath, err := filepath.Abs(filepath.Clean(v.GetString("system-name")))
		if err != nil {
			panic(err)
		}
		// Create our base directory
		if err := os.Mkdir(basepath, 0777); err != nil {
			fmt.Println("Can't create directory", basepath)
			panic(err)
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
				fmt.Println("Can't create directory", dir)
				panic(err)
			}
		}

		fmt.Println("Directory Stuctures Initialized")

		// Handle an SLSFile if one is provided
		var slsState sls_common.SLSState
		if v.GetString("sls-file-path") != "" {
			slsState = loadFromSLS("file://" + v.GetString("sls-file-path"))
		} else if v.GetString("sls-url") != "" {
			slsState = loadFromSLS(v.GetString("sls-url"))
		}
		fmt.Println("SLS File Loaded")

		networks := shasta.ConvertSLSNetworks(slsState)
		// fmt.Println("The networks are: ", networks)
		v.Set("networks.from_sls", networks)
		fmt.Println("Networks Loaded from SLS")
		// shasta.ExtractSwitches(slsState)

		err = v.Unmarshal(&conf)
		if err != nil {
			fmt.Printf("unable to decode into struct, %v \n", err)
		}

		// PrintConfig(v)

		conf.IPV4Resolvers = strings.Split(viper.GetString("ipv4-resolvers"), ",")
		conf.SiteServices.NtpPoolHostname = conf.NtpPoolHostname

		WriteSystemConfig(filepath.Join(basepath, "system_config.yaml"), conf)

		// our primitive ipam uses the number of cabinets to lay out a network for each one.
		var cabinets = uint(conf.MountainCabinets)

		// Merge configs with the NMN Defaults to create a yaml with our subnets in it
		DefaultNMN.CIDR = v.GetString("nmn-cidr")
		_, myNet, _ := net.ParseCIDR(DefaultNMN.CIDR)
		nmnSubnets, err := ipam.Split(*myNet, 32) // 32 allows for a /22 per cabinet
		for k, v := range nmnSubnets[0:cabinets] {
			DefaultNMN.Subnets = append(DefaultNMN.Subnets, shasta.IPV4Subnet{
				CIDR:    v,
				Name:    fmt.Sprintf("cabinet_%v_nmn", int(conf.StartingCabinet)+k),
				Gateway: ipam.Add(v.IP, 1),
				VlanID:  DefaultNMN.VlanRange[0] + int16(k),
			})
		}
		WriteNetworkConfig(filepath.Join(basepath, "networks/nmn.yaml"), DefaultNMN)

		// Merge configs with the HMN Defaults to create a yaml with our subnets in it
		DefaultHMN.CIDR = v.GetString("hmn-cidr")
		_, myNet, _ = net.ParseCIDR(DefaultHMN.CIDR)
		hmnSubnets, err := ipam.Split(*myNet, 32) // 32 allows for a /22 per cabinet
		for k, v := range hmnSubnets[0:cabinets] {
			DefaultHMN.Subnets = append(DefaultHMN.Subnets, shasta.IPV4Subnet{
				CIDR:    v,
				Name:    fmt.Sprintf("cabinet_%v_hmn", int(conf.StartingCabinet)+k),
				Gateway: ipam.Add(v.IP, 1),
				VlanID:  DefaultHMN.VlanRange[0] + int16(k),
			})
		}
		WriteNetworkConfig(filepath.Join(basepath, "networks/hmn.yaml"), DefaultHMN)

		// Merge configs with the HSN Defaults to create a yaml with our subnets in it
		DefaultHSN.CIDR = v.GetString("hsn-cidr")
		_, myNet, _ = net.ParseCIDR(DefaultHSN.CIDR)
		hsnSubnets, err := ipam.Split(*myNet, 32) // 32 allows for a /22 per cabinet
		for k, v := range hsnSubnets[0:cabinets] {
			DefaultHSN.Subnets = append(DefaultHSN.Subnets, shasta.IPV4Subnet{
				CIDR:    v,
				Name:    fmt.Sprintf("cabinet_%v_hsn", int(conf.StartingCabinet)+k),
				Gateway: ipam.Add(v.IP, 1),
				VlanID:  DefaultHMN.VlanRange[0] + int16(k),
			})
		}
		WriteNetworkConfig(filepath.Join(basepath, "networks/hsn.yaml"), DefaultHSN)

		DefaultMTL.CIDR = v.GetString("mtl-cidr")
		WriteNetworkConfig(filepath.Join(basepath, "networks/mtl.yaml"), DefaultMTL)
		DefaultCAN.CIDR = v.GetString("can-cidr")
		WriteNetworkConfig(filepath.Join(basepath, "networks/can.yaml"), DefaultCAN)
		WritePasswordCredential(filepath.Join(basepath, "credentials/root_password.json"), DefaultRootPW)
		WritePasswordCredential(filepath.Join(basepath, "credentials/bmc_password.json"), DefaultBMCPW)
		WritePasswordCredential(filepath.Join(basepath, "credentials/mgmt_switch_password.json"), DefaultNetPW)

		if v.GetString("manifest-release") != "" {
			initiailzeManifestDir("release/shasta-1.4", filepath.Join(basepath, "loftsman-manifests"))
		}
		// InitializeConfiguration()
		// WriteConfigFile()
	},
}

func init() {
	configCmd.AddCommand(initCmd)

	// Flags with defaults for initializing a configuration

	// Flags to deal with 1.3 configuration directories

	initCmd.Flags().String("from-1.3-dir", "", "Shasta 1.3 Configuration Directory")

	// System Configuration Flags based on previous system_config.yml and networks_derived.yml
	initCmd.Flags().String("system-name", "sn-2024", "Name of the System")
	initCmd.Flags().String("site-domain", "cray.io", "Site Domain Name")
	initCmd.Flags().String("internal-domain", "unicos.shasta", "Internal Domain Name")
	initCmd.Flags().String("ntp-pool", "time.nist.gov", "Hostname for Upstream NTP Pool")
	initCmd.Flags().String("ipv4-resolvers", "8.8.8.8, 9.9.9.9", "List of IP Addresses for DNS")
	initCmd.Flags().String("v2-registry", "https://packages.local/", "URL for default v2 registry (helm and containers)")
	initCmd.Flags().String("rpm-repository", "https://packages.local/repository/shasta-master", "URL for default rpm repository")

	// Default IPv4 Networks
	initCmd.Flags().String("nmn-cidr", ipam.DefaultNMN, "Overall IPv4 CIDR for all Node Management subnets")
	initCmd.Flags().String("hmn-cidr", ipam.DefaultHMN, "Overall IPv4 CIDR for all Hardware Management subnets")
	initCmd.Flags().String("can-cidr", ipam.DefaultCAN, "Overall IPv4 CIDR for all Customer Access subnets")
	initCmd.Flags().String("mtl-cidr", ipam.DefaultMTL, "Overall IPv4 CIDR for all Provisioning subnets")
	initCmd.Flags().String("hsn-cidr", ipam.DefaultHSN, "Overall IPv4 CIDR for all HSN subnets")

	// Hardware Details
	initCmd.Flags().Int16("mountain-cabinets", 5, "Number of Mountain Cabinets")
	initCmd.Flags().Int16("starting-cabinet", 1004, "Starting ID number for Mountain Cabinets")
	initCmd.Flags().Int16("starting-NID", 20000, "Starting NID for Compute Nodes")

	// Dealing with an SLS file
	initCmd.Flags().String("from-sls-file", "", "SLS File Location")
	initCmd.Flags().String("from-sls", "", "Shasta 1.3 SLS dumpstate url")
	// Loftsman Manifest Shasta-CFG
	initCmd.Flags().String("manifest-release", "", "Loftsman Manifest Release Version (leave blank to prevent manifest generation")
}

func loadFromSLS(source string) sls_common.SLSState {
	var slsState sls_common.SLSState
	var err error
	if strings.HasPrefix(source, "file://") {
		slsState, err = shasta.ParseSLSFile(strings.TrimPrefix(source, "file://"))
		if err != nil {
			panic(err)
		}
	} else if strings.HasPrefix(source, "https://") {
		slsState, err = shasta.ParseSLSfromURL(strings.TrimPrefix(source, "https://"))
		if err != nil {
			panic(err)
		}
	}
	// Testing that slsState is valid
	if unsafe.Sizeof(slsState) > 0 {
		// At this point, we should have a valid slsState
		// networks := shasta.ConvertSLSNetworks(slsState)
		// fmt.Println(networks)
		ncns := shasta.ExtractNCNBMCInfo(slsState)
		fmt.Println("The NCNs are:", ncns)
	}
	return slsState
}

func writeFile(path string, contents string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	size, err := w.WriteString(contents)
	if err != nil {
		panic(err)
	}
	w.Flush()
	jww.FEEDBACK.Printf("wrote %d bytes to %s\n", size, path)
}

// WriteSystemConfig applies a SystemConfig Struct to the Yaml Template and writes the result to the path indicated
func WriteSystemConfig(path string, conf shasta.SystemConfig) {
	// fmt.Println("Configuration in WriteSystemConfig is:", conf)
	// tmpl, err := template.New("config").Parse(string(DefaultSystemConfigYamlTemplate))
	// if err != nil {
	// panic(err)
	// }
	bs, err := yaml.Marshal(conf)
	// fmt.Print(string(bs))
	if err != nil {
		panic(err)
	}
	writeFile(path, string(bs))
}

// WriteNetworkConfig applies a IPV4Network Struct to the Yaml Template and writes the result to the path indicated
func WriteNetworkConfig(path string, network shasta.IPV4Network) {
	bs, err := yaml.Marshal(network)
	// fmt.Print(string(bs))
	if err != nil {
		panic(err)
	}
	writeFile(path, string(bs))
}

// WritePasswordCredential applies a PasswordCredential to the Yaml Template and writes the result to the path indicated
func WritePasswordCredential(path string, pw shasta.PasswordCredential) error {
	creds, _ := json.Marshal(pw)
	return ioutil.WriteFile(path, creds, 0644)
}

func initiailzeManifestDir(branch, destination string) {
	var url string = "ssh://git@stash.us.cray.com:7999/shasta-cfg/stable.git"
	// First we need a temporary directory
	dir, err := ioutil.TempDir("", "loftsman-init")
	if err != nil {
		panic(err)
	}
	fmt.Println("Adding a temp directory for the git checkout:", dir)
	defer os.RemoveAll(dir)
	cloneCmd := exec.Command("git", "clone", url, dir)
	out, err := cloneCmd.Output()
	if err != nil {
		fmt.Printf("cloneCommand finished with error: %s (%v)\n", out, err)
	}
	fmt.Printf("cloneCommand finished without error: %s \n", out)
	checkoutCmd := exec.Command("git", "checkout", branch)
	checkoutCmd.Dir = dir
	out, err = checkoutCmd.Output()
	if err != nil {
		if err.Error() != "exit status 1" {
			fmt.Printf("checkoutCommand finished with error: %s (%v)\n", out, err)
			panic(err)
		}
	}
	fmt.Printf("checkoutCommand finished without error: %s \n", out)
	packageCmd := exec.Command("./package/package.sh", "1.4.0")
	packageCmd.Dir = dir
	out, err = packageCmd.Output()
	if err != nil {
		fmt.Printf("packageCommand finished with error: %s (%v)\n", out, err)
	}
	fmt.Printf("package finished without error: %s \n", out)
	targz, err := filepath.Abs(filepath.Clean(dir + "/dist/shasta-cfg-1.4.0.tgz"))
	untarCmd := exec.Command("tar", "-zxvvf", targz)
	untarCmd.Dir = destination
	out, err = untarCmd.Output()
	if err != nil {
		fmt.Printf("untarCmd finished with error: %s (%v)\n", out, err)
	}
	fmt.Printf("untarCmd finished without error: %s \n", out)
}
