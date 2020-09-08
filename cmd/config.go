/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("config called")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// LoadConfig : Search reasonable places and read the installer configuration file
func LoadConfig() {
	// Read in the configfile
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}

	viper.SetEnvPrefix("csm")
	viper.AutomaticEnv()
	viper.WatchConfig()

}

//mergeConfig : Merge a configuration file from the local directory.  It will try all known extensions added to the configName
func mergeConfig(configName string) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")

	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}
}

// MergeNetworksDerived : Search reasonable places and read the networks_derived as a config
func MergeNetworksDerived() {
	mergeConfig("networks_derived")
}

// MergeCustomerNetwork : Search reasonable places and read the site_networking as a config
func MergeCustomerNetwork() {
	mergeConfig("site_networking")
}

// MergeNCNMetadata : Search reasonable places and read the ncn_metadata.yaml
func MergeNCNMetadata() {
	// Read in the configfile
	bootstrapNodes := ReadCSV("ncn_metadata.csv")
	// Add it to the configuration
	viper.Set("ncn_metadata", bootstrapNodes)
}

// InitializeConfiguration : Set defaults and load from config files.
func InitializeConfiguration() {

	// Authentication and Security
	viper.SetDefault("system.auth.CERTIFICATE_AUTHORITY_CERT", "ca-cert.pem")
	viper.SetDefault("system.auth.CERTIFICATE_AUTHORITY_KEY", "ca-cert.key")

	// Default Credentials
	viper.SetDefault("system.credentials.NCN_BMC_USERNAME", "")
	viper.SetDefault("system.credentials.NCN_BMC_PASSWORD", "")
	viper.SetDefault("system.credentials.CN_BMC_USERNAME", "")
	viper.SetDefault("system.credentials.CN_BMC_PASSWORD", "")
	viper.SetDefault("system.credentials.REDFISH_USERNAME", "")
	viper.SetDefault("system.credentials.REDFISH_PASSWORD", "")
	viper.SetDefault("system.credentials.LINUX_ROOT_PASSWORD", "initial0")

}

// PrintConfig : Dump all configuration information as a yaml file on stdout
func PrintConfig(v *viper.Viper) {
	log.Print(" == Viper configdump == \n" + yamlStringSettings(v))
}

func yamlStringSettings(v *viper.Viper) string {
	c := v.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	return string(bs)
}

// WriteConfigFile : Capture viper config and writes to config.yaml
func WriteConfigFile() {
	log.Println("Writing configuration to config.yaml")
	viper.WriteConfigAs("config.yaml")
}
