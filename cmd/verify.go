/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Hidden: true, // TODO: Remove this when ready.
	Use:    "verify",
	Short:  "",
	Long:   "",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("verify called \n")
		// TODO: load the directory of files and run any verifications we have available
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
