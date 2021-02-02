/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		LoadConfig()
		if len(args) < 1 {
			PrintConfig(viper.GetViper())
		} else {
			subv := viper.Sub(args[0])
			PrintConfig(subv)

		}
	},
}

func init() {
	configCmd.AddCommand(dumpCmd)
}
