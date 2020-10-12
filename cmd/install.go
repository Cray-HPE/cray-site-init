/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Cray System Management",
	Long:  `Perform a system installation from valid Configuration Payload using Matching Artifact Payload`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("install called")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
