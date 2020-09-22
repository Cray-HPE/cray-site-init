/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load [path]",
	Short: "load a valid Shasta 1.4 directory of configuration",
	Long: `Load a set of files that represent a Shasta 1.4 system.
	Often load is used with init which generates the files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			panic(errors.New("path needs to be provided"))
		}
		basepath, err := filepath.Abs(filepath.Clean(args[0]))
		if err != nil {
			panic(err)
		}
		// This is copied from newSystem.go.  Is it worth abstracting it?
		// Likely more code, but harder to get them out of sync.
		dirs := []string{
			filepath.Join(basepath, "internal_networks"),
			filepath.Join(basepath, "manufacturing"),
			filepath.Join(basepath, "credentials"),
			filepath.Join(basepath, "certificates"),
		}

		for _, dir := range dirs {

			err := filepath.Walk(dir,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					fmt.Println(path, info.Size())
					return nil
				})
			if err != nil {
				log.Println(err)
			}
		}
	},
}

func init() {
	configCmd.AddCommand(loadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
