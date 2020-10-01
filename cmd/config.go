/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config [directory]",
	Short: "Interact with a config in a named directory",
	Long:  `Interact with a config in a named directory`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a named config directory")
		}
		info, err := os.Stat(args[0])
		if err != nil {
			return err
		}
		if !info.Mode().IsDir() {
			return fmt.Errorf("%v is not a directory", args[0])
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("config called with", len(args), "arguments")
		fmt.Println("Echo: " + strings.Join(args, " "))
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
// Possibly no longer relevant
func LoadConfig() {
	// Read in the configfile
	viper.SetConfigName("system_config")
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

// MergeCustomerVar : Search reasonable places and read the customer_var as a config
func MergeCustomerVar() {
	mergeConfig("customer_var")
}

// MergeSLSInput : Search reasonable places and read the customer_var as a config
func MergeSLSInput() {
	mergeConfig("sls_input_file")
}

// MergeSiteNetwork : Search reasonable places and read the site_networking as a config
func MergeSiteNetwork() {
	mergeConfig("site_networking")
}

// MergeSystemNetwork : Search reasonable places and read the site_networking as a config
func MergeSystemNetwork() {
	mergeConfig("system_networking")
}

// MergeNCNMetadata : Search reasonable places and read the ncn_metadata.yaml
func MergeNCNMetadata() {
	// Read in the configfile
	bootstrapNodes := ReadCSV("ncn_metadata.csv")
	// Add it to the configuration
	viper.Set("ncn_metadata", bootstrapNodes)
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
