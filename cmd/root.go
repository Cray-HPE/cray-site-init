/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spf13/viper"
)

var cfgFile string

const (
	// The name of our config file, without the file extension because viper supports many different config file languages.
	defaultConfigFilename = "sic"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --number is bound to STING_NUMBER.
	envPrefix = "SIC"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sic",
	Short: "Shasta Instance Configurator",
	Long: `SIC is a tool for creating and validating the configuration of a Shasta system.
	
	It supports initializing a set of configuration from a variety of inputs including 
	flags and/or Shasta 1.3 configuration files.  It can also validate that a set of 
	configuration details are accurate before attempting to use them for installation`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// fmt.Println("Inside PersistentPreRunE")
		// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
		// viperWiper(viper.GetViper())
		return initializeFlagswithViper(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		envVarSuffix := strings.ToUpper(f.Name)
		if strings.Contains(f.Name, "-") {
			envVarSuffix = strings.ReplaceAll(envVarSuffix, "-", "_")
		}
		v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		} else {
			v.Set(f.Name, f.Value)
		}
	})
}

// This function is useful for understanding what a particular viper contains
func viperWiper(v *viper.Viper) {
	fmt.Print("\n === Viper Wiper === \n\n")
	for _, name := range v.AllKeys() {
		fmt.Println("Key: ", name, " => Name:", v.GetString(name))
	}
	fmt.Print("\n === Viper Wiper Done === \n\n")
}

// This function maps all pflags to strings in viper
func initializeFlagswithViper(cmd *cobra.Command) error {
	v := viper.GetViper()

	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)

		}
		// Add the home directory to the config path
		v.AddConfigPath(home)

		// Add the local directory to the config path
		v.AddConfigPath(".")
		// Set the base name of the config file, without the file extension.
		v.SetConfigName(defaultConfigFilename)
	}

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}
