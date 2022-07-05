/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/pkg/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Run: func(cmd *cobra.Command, args []string) {
		v := viper.GetViper()
		v.BindPFlags(cmd.Flags())
		clientVersion := version.Get()
		if v.GetBool("git") {
			fmt.Println(clientVersion.GitCommit)
			os.Exit(0)
		}
		switch output := v.GetString("output"); output {
		case "pretty":
			fmt.Println("CRAY-Site-Init build signature...")
			fmt.Printf("%-15s: %s\n", "Build Commit", clientVersion.GitCommit)
			fmt.Printf("%-15s: %s\n", "Build Time", clientVersion.BuildDate)
			fmt.Printf("%-15s: %s\n", "Go Version", clientVersion.GoVersion)
			fmt.Printf("%-15s: %s\n", "Version", clientVersion.Version)
			fmt.Printf("%-15s: %s\n", "Platform", clientVersion.Platform)
		case "json":
			b, err := json.Marshal(clientVersion)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(b))
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.DisableAutoGenTag = true
	versionCmd.Flags().StringP("output", "o", "pretty", "output format pretty,json")
	versionCmd.Flags().BoolP("git", "g", false, "Simple commit sha of the source tree on a single line. \"-dirty\" added to the end if uncommitted changes present")
}
