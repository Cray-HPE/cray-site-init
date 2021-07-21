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

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dumps an existing config to STDOUT",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			PrintConfig(viper.GetViper())
		} else {
			subv := viper.Sub(args[0])
			PrintConfig(subv)

		}
	},
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

func init() {
	dumpCmd.DisableAutoGenTag = true

}
