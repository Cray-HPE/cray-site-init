package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/
import (
	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrades components of a CSM installation",
	Long:  "Upgrades components of a CSM installation",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.DisableAutoGenTag = true
}
