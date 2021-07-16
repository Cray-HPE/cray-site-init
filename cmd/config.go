/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Interact with a Shasta config",
	Long:  `Interact with a Shasta config`,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	configCmd.AddCommand(dumpCmd)
	configCmd.AddCommand(genSLSCmd)
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(loadCmd)
}
