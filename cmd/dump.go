/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
