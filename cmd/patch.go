package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

import (
	"github.com/spf13/cobra"
)

var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Apply patch operations",
	Long: `
Runs patch operations against the CRAY.
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)
	patchCmd.DisableAutoGenTag = true
}
