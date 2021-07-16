/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cwd     string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
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
			cmd.Usage()
		},
	}
)

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
		log.Println("Using config override")

		cfgFile, _ := filepath.Abs(cfgFile)

		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {

			// Exit if the file does not exist
			log.Println(err)
			os.Exit(1)

		} else {

			// Use config file from the flag if it's set
			viper.SetConfigFile(cfgFile)

		}

	} else {

		// Find current directory
		// Search config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName("system_config")
		viper.SetConfigType("yaml")

	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
		// PrintConfig(viper.GetViper())
	} else {
		// PrintConfig(viper.GetViper())
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PWD/system_config.yaml)")
}
