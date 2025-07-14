/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

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

package config

import (
	"log"

	"github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/config/shcd"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/config/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewCommand represents the config command
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "config",
		Short:             "HPC configuration",
		Long:              `Creates Cray-HPE site configuration files`,
		DisableAutoGenTag: true,
		Args:              cobra.MinimumNArgs(1),
	}

	c.AddCommand(
		dumpCommand(),
		initialize.NewCommand(),
		shcd.NewCommand(),
		sls.NewCommand(),
		template.NewCommand(),
	)
	return c
}

func yamlStringSettings(v *viper.Viper) string {
	c := v.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf(
			"unable to marshal config to YAML: %v",
			err,
		)
	}
	return string(bs)
}
