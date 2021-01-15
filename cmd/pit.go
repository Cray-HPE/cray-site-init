package cmd

/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
import (
	"log"

	"github.com/spf13/cobra"
)

// pitCmd represents the pit command
var pitCmd = &cobra.Command{
	Use:   "pit",
	Short: "Manipulate or Create a LiveCD (Pre-Install Toolkit)",
	Long: `
Interact with the Pre-Install Toolkit (LiveCD);
create, validate, or re-create a new or old USB stick with 
the liveCD tool. Fetches artifacts for deployment.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("pit called")
	},
}

func init() {
	rootCmd.AddCommand(pitCmd)
}
