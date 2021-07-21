/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
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

	// Add a global '--config' option, so someone can pass in their own
	// config file if desired, overriding the one in the current dir when
	// we initialize this cobra program
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}
