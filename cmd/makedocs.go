/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// makedocsCmd represents the makedocs command
var makedocsCmd = &cobra.Command{
	Use:   "makedocs [directory]",
	Short: "Create a set of markdown files for the docs in the [directory] (docs/ is the default)",
	Run: func(cmd *cobra.Command, args []string) {
		var destinationDirectory string
		if len(args) < 1 {
			destinationDirectory = "docs/" // This is the default without passing an argument
		} else {
			destinationDirectory = args[0]
		}
		basepath, err := filepath.Abs(filepath.Clean(destinationDirectory))
		_, err = os.Stat(basepath)
		if err != nil {
			// Assert that the error is actually a PathError or bail
			_, ok := err.(*os.PathError)
			if ok != true {
				log.Fatalf("Error accessing %v :%v", basepath, err)
			}
		}
		err = os.Mkdir(basepath, 0777)
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
