/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config [directory]",
	Short: "Interact with a config in a named directory",
	Long:  `Interact with a config in a named directory`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a named config directory")
		}
		info, err := os.Stat(args[0])
		if err != nil {
			return fmt.Errorf("could not read %v. %v", args[0], err)
		}
		if !info.Mode().IsDir() {
			return fmt.Errorf("%v is not a directory", args[0])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.DisableAutoGenTag = true
}

// LoadConfig : Search reasonable places and read the installer configuration file
// Possibly no longer relevant
func LoadConfig() {
	// Read in the configfile
	viper.SetConfigName("system_config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalln(fmt.Errorf("fatal error config file: %s", err))
		}
	}
	viper.SetEnvPrefix("CSM")
	viper.AutomaticEnv()
	viper.WatchConfig()
}

// PrintConfig : Dump all configuration information as a yaml file on stdout
func PrintConfig(v *viper.Viper) {
	log.Println(" == Viper configdump == \n" + yamlStringSettings(v))
}

func yamlStringSettings(v *viper.Viper) string {
	c := v.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	return string(bs)
}
