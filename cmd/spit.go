package cmd
/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"

	"github.com/spf13/cobra"
)

// spitCmd represents the spit command
var spitCmd = &cobra.Command{
	Use:   "spit",
	Short: "Control SPIT",
	Long: `Control SPIT`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("spit called")
	},
}

func init() {
	rootCmd.AddCommand(spitCmd)
}
