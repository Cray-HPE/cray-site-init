/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"github.com/spf13/cobra"
)

// handoffCmd represents the handoff command
var handoffCmd = &cobra.Command{
	Use:   "handoff",
	Short: "runs migration steps to transition from LiveCD",
	Long: "A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the " +
		"production version inside the Kubernetes cluster.",
}

func init() {
	rootCmd.AddCommand(handoffCmd)
}
