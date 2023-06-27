/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
*/

package cmd

import (
	"github.com/Cray-HPE/cray-site-init/cmd/patch"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultConfigFilename = "csi"
	envPrefix             = "csi"
)

var (
	cfgFile string
	cwd     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "csi",
	Short: "Cray Site Init. For new sites, re-installs, and upgrades.",
	Long: `
	CSI creates, validates, installs, and upgrades a CRAY supercomputer or HPCaaS platform.

	It supports initializing a set of configuration files from a variety of inputs including
	flags and Shasta 1.3 configuration files. It can also validate that a set of
	configuration details are accurate before attempting to use them for installation.

	Configs aside, this will prepare USB sticks for deploying on baremetal or for recovery and
	triage.`,
	Run: func(cmd *cobra.Command, args []string) {
		// When we bind flags to environment variables expect that the
		// environment variables are prefixed, e.g. a flag like --number
		// binds to an environment variable STRING_NUMBER. This helps
		// avoid conflicts.
		reg, err := regexp.Compile("[^A-Za-z0-9]+")
		if err != nil {
			log.Fatal(err)
		}

		// "csi config init --option" => csi_CONFIG_INIT_OPTION
		reggie := reg.ReplaceAllString(cmd.CommandPath(), "_")
		log.Println(reggie)
		viper.BindPFlags(cmd.Flags())
		cmd.Usage()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Set the config file from the logic and checks above
		cfgFile, _ = filepath.Abs(cfgFile)
	} else {
		// log.Println("Using default config location")
		cfgFile = "system_config.yaml"
		cfgFile, _ = filepath.Abs(cfgFile)
	}

	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

func init() {
	// We don't actually use this until a user runs 'config init', but it's more
	// standard to keep it here in the main file
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(patch.PatchCmd)
	// Add a global '--config' option, so someone can pass in their own
	// config file if desired, overriding the one in the current dir when
	// we initialize this cobra program
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}
