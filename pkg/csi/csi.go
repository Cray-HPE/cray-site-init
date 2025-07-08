/*
 MIT License

 (C) Copyright 2024-2025 Hewlett Packard Enterprise Development LP

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
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/Cray-HPE/cray-site-init/pkg/cli"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize"
	"github.com/Cray-HPE/cray-site-init/pkg/csm"
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
	DefaultConfigFilename      = "system_config"
	envPrefix                  = "csi"
	replaceHyphenWithCamelCase = false
	DefaultConfigSuffix        = "yaml"
)

var (
	cfgFile  string
	inputDir string
)

// NewCommand represents the base command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "csi",
		Short: "Cray Site Initializer (csi)",
		Long: `
Tooling for the initial configuration and deployment of a Cray-HPE
Exascale High-Performance Computer (HPC) or an HPCaaS (e.g. VShasta).
`,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			cli.Runtime = time.Now().UTC()
			cli.RuntimeTimestamp = cli.Runtime.Format(time.RFC3339Nano)
			cli.RuntimeTimestampShort = cli.Runtime.Format("20060102150405")
			return initializeConfig(c)
		},
	}
	c.PersistentFlags().StringVarP(
		&cfgFile,
		"config",
		"c",
		"",
		fmt.Sprintf(
			"Path to a CSI config file (default is $PWD/%s.%s).",
			DefaultConfigFilename,
			DefaultConfigSuffix,
		),
	)
	_ = c.MarkPersistentFlagFilename("config")

	c.PersistentFlags().StringVarP(
		&inputDir,
		"input-dir",
		"i",
		"",
		fmt.Sprintf(
			"A directory to read input files from (--config will take precedence, but only for %s.%s).",
			DefaultConfigFilename,
			DefaultConfigSuffix,
		),
	)
	_ = c.MarkPersistentFlagFilename("input-dir")

	c.PersistentFlags().StringVar(
		&csm.AdminTokenSecretName,
		"k8s-secret-name",
		csm.DefaultAdminTokenSecretName,
		`(for use against a completed CSM installation) The name of the Kubernetes secret to look for an OpenID credential in for CSM APIs (a.k.a. TOKEN=).`,
	)

	c.PersistentFlags().StringVar(
		&csm.AdminTokenSecretNamespace,
		"k8s-namespace",
		csm.DefaultAdminTokenSecretNamespace,
		"(for use against a completed CSM installation) The namespace that the --k8s-secret-name belongs to.",
	)

	c.PersistentFlags().StringVar(
		&csm.BaseAPIURL,
		"csm-api-url",
		csm.DefaultBaseAPIURL,
		"(for use against a completed CSM installation) The URL to a CSM API.",
	)

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
		v.SetConfigName(DefaultConfigFilename)
		v.SetConfigType(DefaultConfigSuffix)
		// Set as many paths as you like where viper should look for the
		// config file. We are only looking in the current working directory.
		v.AddConfigPath(".")
		cli.ConfigFilename = fmt.Sprintf(
			"%s.%s",
			DefaultConfigFilename,
			DefaultConfigSuffix,
		)
	} else {
		configFileAbs, _ := filepath.Abs(configFile)
		v.SetConfigFile(configFileAbs)
		configFileSuffix := filepath.Ext(configFileAbs)
		configFileSuffix = strings.TrimPrefix(
			configFileSuffix,
			".",
		)
		v.SetConfigType(filepath.Ext(configFileSuffix))
		cli.ConfigFilename = filepath.Base(configFileAbs)
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
			mergeFlags(
				cmd,
				v,
				f,
			)
		},
	)
}

func mergeFlags(cmd *cobra.Command, v *viper.Viper, f *pflag.Flag) {

	// Determine the naming convention of the flags when represented in the config file
	flagName := f.Name

	// If using camelCase in the config file, replace hyphens with a camelCased string.
	// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
	if replaceHyphenWithCamelCase {
		flagName = strings.ReplaceAll(
			f.Name,
			"-",
			"",
		)
	}

	// Cobra won't print the deprecation warning if the flag comes in through a config file. Here, we'll force print
	// the deprecation message for a flag (if it exists) when we're reading config files.
	if f.Deprecated != "" {
		// Only notify the user if the deprecated flag exists in a Viper config. Cobra will automatically notify them
		// for any usage on the CLi.
		if v.IsSet(flagName) {
			fmt.Printf(
				"Flag --%s has been deprecated, %s\n",
				flagName,
				f.Deprecated,
			)
		}

		// The deprecated message must have the new flag in it somewhere for this to work.
		var newFlag *pflag.Flag
		for _, word := range strings.Split(
			f.Deprecated,
			" ",
		) {
			// Trim any leading --.
			poleLess := strings.TrimLeft(
				word,
				"-",
			)
			periodLess := strings.TrimRight(
				poleLess,
				".",
			)
			newFlag = cmd.Flags().Lookup(periodLess)
			if newFlag != nil {
				break
			}
		}
		if v.Get(flagName) != nil && v.Get(newFlag.Name) == nil {
			fmt.Printf(
				"Updating flag %s\n",
				newFlag.Name,
			)
			// Rewrite the value to update any other aliases.
			v.Set(
				flagName,
				v.Get(flagName),
			)
			err := cmd.Flags().Set(
				newFlag.Name,
				v.GetString(flagName),
			)
			if err != nil {
				log.Fatalf(
					"Failed to set non-deprecated flag %s from %s because %v\n",
					newFlag.Name,
					flagName,
					err,
				)
			}
		}

		initialize.DeprecatedKeys = append(
			initialize.DeprecatedKeys,
			flagName,
		)
		return
	}
	// Apply the viper config value to the flag when the flag is not set and viper has a value
	if !f.Changed && v.IsSet(flagName) {
		val := v.Get(flagName)
		kind := reflect.TypeOf(val).Kind()
		switch kind {
		case reflect.Slice:

			// Cast Slices to real Slice types.
			values := reflect.ValueOf(val).Interface().([]interface{})
			var newSlice []string
			for _, v := range values {
				newSlice = append(
					newSlice,
					v.(string),
				)
			}
			err := cmd.Flags().Set(
				f.Name,
				strings.Join(
					newSlice,
					",",
				),
			)
			if err != nil {
				log.Fatalf(
					"Failed to parse flag %s because %v",
					f.Name,
					err,
				)
			}

		case reflect.String, reflect.Int, reflect.Bool, reflect.Float32, reflect.Float64:
			err := cmd.Flags().Set(
				f.Name,
				fmt.Sprintf(
					"%v",
					val,
				),
			)
			if err != nil {
				log.Fatalf(
					"Failed to parse flag %s because %v",
					f.Name,
					err,
				)
			}
		default:
			fmt.Printf(
				"No handling for type %s from %s\nAssuming it can be handled as a string.\n",
				kind,
				f.Name,
			)
			err := cmd.Flags().Set(
				f.Name,
				fmt.Sprintf(
					"%v",
					val,
				),
			)
			if err != nil {
				log.Fatalf(
					"Failed to parse flag %s because %v",
					f.Name,
					err,
				)
			}
		}
	}
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
