/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"github.com/spf13/cobra"
)

// rawCmd represents the raw command
var rawCmd = &cobra.Command{
	Use:   "raw",
	Short: "A collection of commands for generating raw yaml",
	Long: ``,
}

func init() {
	configCmd.AddCommand(rawCmd)
}
