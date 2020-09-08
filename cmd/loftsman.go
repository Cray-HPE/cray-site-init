/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// loftsmanCmd represents the loftsman command
var loftsmanCmd = &cobra.Command{
	Use:   "loftsman",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("loftsman called")
	},
}

func init() {
	rootCmd.AddCommand(loftsmanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loftsmanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loftsmanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
