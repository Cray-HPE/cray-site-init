/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Interact with a Shasta config",
	Long:  `Interact with a Shasta config`,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	configCmd.DisableAutoGenTag = true
	configCmd.AddCommand(dumpCmd)
	configCmd.AddCommand(genSLSCmd)
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(loadCmd)
	configCmd.AddCommand(shcdCmd)
}

// PrintConfig : Dump all configuration information as a yaml file on stdout
func PrintConfig(v *viper.Viper) {
	log.Println(" == Viper configdump == \n" + yamlStringSettings(v))
}

func yamlStringSettings(v *viper.Viper) string {
	c := v.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	return string(bs)
}
