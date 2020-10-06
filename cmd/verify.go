/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("verify called \n")
		// TODO: load the directory of files and run any verifications we have available
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
