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
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func init() {
	configCmd.AddCommand(rawCmd)
}
