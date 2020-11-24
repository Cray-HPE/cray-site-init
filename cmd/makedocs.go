/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// makedocsCmd represents the makedocs command
var makedocsCmd = &cobra.Command{
	Use:   "makedocs",
	Short: "Create a set of markdown files for the docs",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("makedocs called")
		destination := "docs/"
		basepath, err := filepath.Abs(filepath.Clean(destination))
		os.Mkdir(basepath, 777)
		err = doc.GenMarkdownTree(rootCmd, basepath)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(makedocsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// makedocsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// makedocsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
