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
	Short: "Control PIT",
	Long:  `Control PIT`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("pit called")
	},
}

func init() {
	rootCmd.AddCommand(pitCmd)
}
