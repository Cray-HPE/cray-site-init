/*
 MIT License

 (C) Copyright 2024 Hewlett Packard Enterprise Development LP

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

package csi

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/pkg/cli/automation"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/config"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/handoff"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/patch"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/pit"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/upload"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/version"
)

const (
	defaultConfigFilename      = "system_config"
	envPrefix                  = "csi"
	replaceHyphenWithCamelCase = false
)

var cfgFile string

// NewCommand represents the base command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "csi",
		Short: "Cray Site Init. For new sites, re-installs, and upgrades.",
		Long: `
	CSI creates, validates, installs, and upgrades a CRAY supercomputer or HPCaaS platform.

	It supports initializing a set of configuration files from a variety of inputs including
	flags and Shasta 1.3 configuration files. It can also validate that a set of
	configuration details are accurate before attempting to use them for installation.

	Configs aside, this will prepare USB sticks for deploying on baremetal or for recovery and
	triage.`,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			return initializeConfig(c)
		},
	}
	c.PersistentFlags().StringVarP(
		&cfgFile,
		"config",
		"c",
		"",
		"CSI config file",
	)
	_ = c.MarkPersistentFlagFilename("config")

	c.AddCommand(
		automation.NewCommand(),
		config.NewCommand(),
		handoff.NewCommand(),
		patch.NewCommand(),
		pit.NewCommand(),
		DocsCommand(),
		sls.NewCommand(),
		version.NewCommand(),
	)
	return c
}

func initializeConfig(c *cobra.Command) error {
	v := viper.New()

	configFile := c.Flag("config").Value.String()
	if configFile == "" {
		// Set the base name of the config file, without the file extension.
		v.SetConfigName(defaultConfigFilename)

		// Set as many paths as you like where viper should look for the
		// config file. We are only looking in the current working directory.
		v.AddConfigPath(".")
	} else {
		configFileAbs, _ := filepath.Abs(configFile)
		v.SetConfigFile(configFileAbs)
	}

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file unless one was given.
		if configFile != "" {
			return err
		} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	v.SetEnvKeyReplacer(
		strings.NewReplacer(
			"-",
			"_",
		),
	)

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(
		c,
		v,
	)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(
		func(f *pflag.Flag) {

			// Determine the naming convention of the flags when represented in the config file
			configName := f.Name
			// If using camelCase in the config file, replace hyphens with a camelCased string.
			// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
			if replaceHyphenWithCamelCase {
				configName = strings.ReplaceAll(
					f.Name,
					"-",
					"",
				)
			}

			// Apply the viper config value to the flag when the flag is not set and viper has a value
			if !f.Changed && v.IsSet(configName) {

				val := v.Get(configName)
				kind := reflect.TypeOf(val).Kind()

				if kind == reflect.Slice {

					// Cast Slices to real Slice types.
					values := reflect.ValueOf(val).Interface().([]interface{})
					var newSlice []string
					for _, v := range values {
						newSlice = append(
							newSlice,
							v.(string),
						)
					}
					cmd.Flags().Set(
						f.Name,
						strings.Join(
							newSlice,
							",",
						),
					)

				} else if kind == reflect.String || kind == reflect.Int || kind == reflect.Bool || kind == reflect.Float32 || kind == reflect.Float64 {
					cmd.Flags().Set(
						f.Name,
						fmt.Sprintf(
							"%v",
							val,
						),
					)
				} else {
					fmt.Printf(
						"No handling for type %s from %s\nAssuming it can be handled as a string.\n",
						kind,
						f.Name,
					)
					cmd.Flags().Set(
						f.Name,
						fmt.Sprintf(
							"%v",
							val,
						),
					)
				}
			}
		},
	)
}

// stringInSlice is shorthand
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// ExecuteCommandC runs a cobra command
func ExecuteCommandC(root *cobra.Command, args []string) (
	c *cobra.Command, output string, err error,
) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}
