/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"stash.us.cray.com/MTL/csi/pkg/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Run: func(cmd *cobra.Command, args []string) {
		v := viper.GetViper()
		clientVersion := version.Get()
		if v.GetBool("simple") {
			fmt.Printf("%v.%v\n", clientVersion.Major, clientVersion.Minor)
			os.Exit(0)
		}
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
			fmt.Printf("%-15s: %s\n", "Git Version", clientVersion.GitVersion)
			fmt.Printf("%-15s: %s\n", "Platform", clientVersion.Platform)
			fmt.Printf("%-15s: %v.%v.%v\n", "App. Version", clientVersion.Major, clientVersion.Minor, clientVersion.FixVr)
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
	versionCmd.Flags().StringP("output", "o", "pretty", "output format pretty,json")
	versionCmd.Flags().BoolP("simple", "s", false, "Simple version on a single line")
	versionCmd.Flags().BoolP("git", "g", false, "Simple commit sha of the source tree on a single line. \"-dirty\" added to the end if uncommitted changes present")
}
