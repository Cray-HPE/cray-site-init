/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"stash.us.cray.com/MTL/csi/pkg/version"
)

var (
	sha1ver    string // sha1 revision used to build the program
	gitVersion string // git tag version
	buildTime  string // when the executable was built
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version called")
		clientVersion := version.Get()
		clientVersion.GitCommit = sha1ver
		clientVersion.BuildDate = buildTime
		clientVersion.GitVersion = gitVersion
		fmt.Println("commit is", clientVersion.GitCommit)
		fmt.Println("version is", clientVersion.GitVersion)
		fmt.Println("binary build signature:")
		fmt.Println("Go Version:", clientVersion.GoVersion)
		fmt.Println("Platform:", clientVersion.Platform)
		fmt.Println("Build Time:", clientVersion.BuildDate)
		fmt.Println("Build Version:", clientVersion.GitVersion)
		fmt.Println("Build Commit:", clientVersion.GitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
