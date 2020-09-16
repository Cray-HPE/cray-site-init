/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"stash.us.cray.com/MTL/sic/pkg/ipam"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// newSystemCmd represents the newSystem command
var newSystemCmd = &cobra.Command{
	Use:   "newSystem [path]",
	Short: "Create default configs for a new system (skeleton)",
	Long: `Create new configs in the provided directory.
	The new configs will have the correct structure and some sensible defaults, but no localized config yet.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			panic(errors.New("path needs to be provided"))
		}
		basepath, err := filepath.Abs(filepath.Clean(args[0]))
		if err != nil {
			panic(err)
		}
		dirs := []string{
			filepath.Join(basepath, "internal_networks"),
			filepath.Join(basepath, "manufacturing"),
			filepath.Join(basepath, "credentials"),
			filepath.Join(basepath, "certificates"),
		}

		if err := os.Mkdir(basepath, 0777); err != nil {
			panic(err)
		}
		for _, dir := range dirs {
			if err := os.Mkdir(dir, 0777); err != nil {
				panic(err)
			}
		}
		var conf shasta.SystemConfig
		viper.Unmarshal(&conf)

		WriteSystemConfig(filepath.Join(basepath, "system_config.yaml"), conf)
		WriteNetworkConfig(filepath.Join(basepath, "internal_networks/nmn.yaml"), DefaultNMN)
		WriteNetworkConfig(filepath.Join(basepath, "internal_networks/hmn.yaml"), DefaultHMN)
		WriteNetworkConfig(filepath.Join(basepath, "internal_networks/hsn.yaml"), DefaultHSN)
		WriteNetworkConfig(filepath.Join(basepath, "internal_networks/mtl.yaml"), DefaultMTL)

	},
}

func init() {
	initCmd.AddCommand(newSystemCmd)

	newSystemCmd.Flags().Int16("mountain_cabinets", 5, "Number of Mountain Cabinets")
	viper.BindPFlag("MountainCabinets", newSystemCmd.Flags().Lookup("mountain_cabinets"))

	newSystemCmd.Flags().String("system_name", "sn-2024", "Name of the System")
	viper.BindPFlag("SystemName", newSystemCmd.Flags().Lookup("system_name"))

	newSystemCmd.Flags().String("site_domain", "cray.io", "Site Domain Name")
	viper.BindPFlag("SiteDomain", newSystemCmd.Flags().Lookup("site_domain"))

	newSystemCmd.Flags().String("internal_domain", "unicos.shasta", "Internal Domain Name")
	viper.BindPFlag("InternalDomain", newSystemCmd.Flags().Lookup("internal_domain"))

	newSystemCmd.Flags().String("ntp_pool", "time.nist.gov", "Hostname for Upstream NTP Pool")
	viper.BindPFlag("NtpPoolHostname", newSystemCmd.Flags().Lookup("ntp_pool"))

	newSystemCmd.Flags().StringArray("ipv4_resolvers", []string{"8.8.8.8", "9.9.9.9"}, "List of IP Addresses for DNS")
	viper.BindPFlag("IPV4Resolvers", newSystemCmd.Flags().Lookup("ipv4_resolvers"))

	newSystemCmd.Flags().IPNet("nmn_cidr", ipam.NmnCIDR, "Overall CIDR for all Node Management subnets")
	newSystemCmd.Flags().IPNet("hmn_cidr", ipam.HmnCIDR, "Overall CIDR for all Hardware Management subnets")
	newSystemCmd.Flags().IPNet("can_cidr", ipam.CanCIDR, "Overall CIDR for all Customer Access subnets")
	newSystemCmd.Flags().IPNet("mtl_cidr", ipam.MtlCIDR, "Overall CIDR for all Provisioning subnets")
	newSystemCmd.Flags().IPNet("slingshot_cidr", ipam.HsnCIDR, "Overall CIDR for all Slingshot subnets")
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
