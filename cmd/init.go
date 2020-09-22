/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"unsafe"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"

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

		// After processing all the flags in init,
		// if the user has an old configuration dir, use that
		if viper.GetString("oldconfigdir") != "" {
			LoadConfig()
			MergeNCNMetadata()
			MergeNetworksDerived()
			MergeSLSInput()
			MergeCustomerVar()
		}

		// Set up the path for our base directory using our systemname
		basepath, err := filepath.Abs(filepath.Clean(viper.GetString("SystemName")))
		if err != nil {
			panic(err)
		}
		// Create our base directory
		if err := os.Mkdir(basepath, 0777); err != nil {
			panic(err)
		}

		// These Directories make up the overall structure for the Configuration Payload
		dirs := []string{
			filepath.Join(basepath, "internal_networks"),
			filepath.Join(basepath, "manufacturing"),
			filepath.Join(basepath, "credentials"),
			filepath.Join(basepath, "certificates"),
		}
		// Add the Manifest directory if needed
		if viper.GetString("ManifestRelease") != "" {
			dirs = append(dirs, filepath.Join(basepath, "loftsman-manifests"))
		}
		// Iterate through the directories and create them
		for _, dir := range dirs {
			if err := os.Mkdir(dir, 0777); err != nil {
				panic(err)
			}
		}

		// Handle an SLSFile if one is provided
		var slsState sls_common.SLSState
		if viper.GetString("slsfilepath") != "" {
			slsState = loadFromSLS("file://" + viper.GetString("slsfilepath"))
		} else if viper.GetString("slsurl") != "" {
			slsState = loadFromSLS(viper.GetString("slsurl"))
		}

		networks := shasta.ConvertSLSNetworks(slsState)
		fmt.Println("The networks are: ", networks)

		var conf shasta.SystemConfig
		viper.Unmarshal(&conf)

		WriteSystemConfig(filepath.Join(basepath, "system_config.yaml"), conf)
		DefaultNMN.CIDR = viper.GetString("NMNCidr")
		WriteNetworkConfig(filepath.Join(basepath, "internal_networks/nmn.yaml"), DefaultNMN)
		DefaultHMN.CIDR = viper.GetString("HMNCidr")
		WriteNetworkConfig(filepath.Join(basepath, "internal_networks/hmn.yaml"), DefaultHMN)
		DefaultHSN.CIDR = viper.GetString("HSNCidr")
		WriteNetworkConfig(filepath.Join(basepath, "internal_networks/hsn.yaml"), DefaultHSN)
		DefaultMTL.CIDR = viper.GetString("MTLCidr")
		WriteNetworkConfig(filepath.Join(basepath, "internal_networks/mtl.yaml"), DefaultMTL)
		WritePasswordCredential(filepath.Join(basepath, "credentials/root_password.json"), DefaultRootPW)
		WritePasswordCredential(filepath.Join(basepath, "credentials/bmc_password.json"), DefaultBMCPW)
		WritePasswordCredential(filepath.Join(basepath, "credentials/mgmt_switch_password.json"), DefaultNetPW)

		if viper.GetString("ManifestRelease") != "" {
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
	viper.BindPFlag("oldconfigdir", initCmd.Flags().Lookup("from-1.3-dir"))

	// System Configuration Flags based on previous system_config.yml and networks_derived.yml
	initCmd.Flags().String("system_name", "sn-2024", "Name of the System")
	viper.BindPFlag("SystemName", initCmd.Flags().Lookup("system_name"))

	initCmd.Flags().String("site_domain", "cray.io", "Site Domain Name")
	viper.BindPFlag("SiteDomain", initCmd.Flags().Lookup("site_domain"))

	initCmd.Flags().String("internal_domain", "unicos.shasta", "Internal Domain Name")
	viper.BindPFlag("InternalDomain", initCmd.Flags().Lookup("internal_domain"))

	initCmd.Flags().String("ntp_pool", "time.nist.gov", "Hostname for Upstream NTP Pool")
	viper.BindPFlag("NtpPoolHostname", initCmd.Flags().Lookup("ntp_pool"))

	initCmd.Flags().StringArray("ipv4_resolvers", []string{"8.8.8.8", "9.9.9.9"}, "List of IP Addresses for DNS")
	viper.BindPFlag("IPV4Resolvers", initCmd.Flags().Lookup("ipv4_resolvers"))

	// Default IPv4 Networks
	initCmd.Flags().IPNet("nmn_cidr", ipam.DefaultNmnCIDR, "Overall IPv4 CIDR for all Node Management subnets")
	viper.BindPFlag("NMNCidr", initCmd.Flags().Lookup("nmn_cidr"))

	initCmd.Flags().IPNet("hmn_cidr", ipam.DefaultHmnCIDR, "Overall IPv4 CIDR for all Hardware Management subnets")
	viper.BindPFlag("HMNCidr", initCmd.Flags().Lookup("hmn_cidr"))

	initCmd.Flags().IPNet("can_cidr", ipam.DefaultCanCIDR, "Overall IPv4 CIDR for all Customer Access subnets")
	viper.BindPFlag("CANCidr", initCmd.Flags().Lookup("can_cidr"))

	initCmd.Flags().IPNet("mtl_cidr", ipam.DefaultMtlCIDR, "Overall IPv4 CIDR for all Provisioning subnets")
	viper.BindPFlag("MTLCidr", initCmd.Flags().Lookup("mtl_cidr"))

	initCmd.Flags().IPNet("hsn_cidr", ipam.DefaultHsnCIDR, "Overall IPv4 CIDR for all HSN subnets")
	viper.BindPFlag("HSNCidr", initCmd.Flags().Lookup("hsn_cidr"))

	// Hardware Details
	initCmd.Flags().Int16("mountain_cabinets", 5, "Number of Mountain Cabinets")
	viper.BindPFlag("MountainCabinets", initCmd.Flags().Lookup("mountain_cabinets"))

	// Dealing with an SLS file
	initCmd.Flags().String("from-sls-file", "", "SLS File Location")
	viper.BindPFlag("slsfilepath", initCmd.Flags().Lookup("from-sls-file"))

	initCmd.Flags().String("from-sls", "", "Shasta 1.3 SLS dumpstate url")
	viper.BindPFlag("slsurl", initCmd.Flags().Lookup("from-sls"))

	// Loftsman Manifest Shasta-CFG
	initCmd.Flags().String("manifest-release", "", "Loftsman Manifest Release Version (leave blank to prevent manifest generation")
	viper.BindPFlag("ManifestRelease", initCmd.Flags().Lookup("manifest-release"))

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
		fmt.Println(ncns)
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
	tmpl, err := template.New("config").Parse(string(DefaultSystemConfigYamlTemplate))
	if err != nil {
		panic(err)
	}
	buff := &bytes.Buffer{}
	err = tmpl.Execute(buff, conf)
	if err != nil {
		panic(err)
	}
	writeFile(path, buff.String())
}

// WriteNetworkConfig applies a IPV4Network Struct to the Yaml Template and writes the result to the path indicated
func WriteNetworkConfig(path string, network shasta.IPV4Network) {
	tmpl, err := template.New("config").Parse(string(DefaultNetworkConfigYamlTemplate))
	if err != nil {
		panic(err)
	}
	buff := &bytes.Buffer{}
	err = tmpl.Execute(buff, network)
	if err != nil {
		panic(err)
	}
	writeFile(path, buff.String())
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
	defer os.RemoveAll(dir)
	cloneCmd := exec.Command("git", "clone", url, dir)
	err = cloneCmd.Run()
	if err != nil {
		fmt.Printf("cloneCommand finished with error: %v", err)
	}
	checkoutCmd := exec.Command("git", "checkout", branch)
	checkoutCmd.Dir = dir
	out, err := checkoutCmd.Output()
	if err != nil {
		if err.Error() != "exit status 1" {
			fmt.Printf("checkoutCommand finished with error: %s (%v)\n", out, err)
			panic(err)
		}
	}
	initCmd := exec.Command("./meta/init.sh", destination)
	initCmd.Dir = dir
	out, err = initCmd.Output()
	if err != nil {
		fmt.Printf("initCommand ffinished with error: %s (%v)\n", out, err)
	}
}
