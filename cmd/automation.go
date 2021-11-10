package cmd

import (
	"github.com/spf13/cobra"
)

var automateCommand = &cobra.Command{
	Use:   "automate",
	Short: "tools used to automate system lifecycle events",
	Long:  "A series of subcommands that automates the day-to-day administration of common lifecycle events.",
}

func init() {
	rootCmd.AddCommand(automateCommand)
	automateCommand.DisableAutoGenTag = true
}
