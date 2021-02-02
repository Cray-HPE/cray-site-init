/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// loftsmanCmd represents the loftsman command
var loftsmanCmd = &cobra.Command{
	Hidden:     true,                    // TODO: Remove this when ready.
	Deprecated: "use loftsmen directly", // FIXME: This may never be used pending loftsmen use in 1.4.
	Use:        "loftsman",
	Short:      "",
	Long:       "",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("loftsman called")
	},
}

func init() {
	rootCmd.AddCommand(loftsmanCmd)
}
