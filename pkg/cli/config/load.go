/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

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
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"

	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

// loadCommand represents the load sub-command
func loadCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "load <path>",
		Short: "Load an existing Shasta configuration",
		Long: `Load a set of files that represent a Shasta system.
	Often, 'load' is used with 'init', which generates the files.`,
		DisableAutoGenTag: true,
		Run: func(c *cobra.Command, args []string) {
			if len(args) < 1 {
				log.Fatalln(errors.New("path needs to be provided"))
			}
			basepath, err := filepath.Abs(filepath.Clean(args[0]))
			if err != nil {
				log.Fatalln(err)
			}
			sysconfig, _ := loadSystemConfig(
				filepath.Join(
					basepath,
					"system_config.yaml",
				),
			)
			networks, _ := extractNetworks(
				filepath.Join(
					basepath,
					"networks",
				),
			)

			log.Println(
				"Loaded ",
				sysconfig.SystemName,
				sysconfig.SiteDomain,
			)
			for _, v := range networks {
				log.Println(
					v.Name,
					":",
					v.CIDR,
					len(v.Subnets),
					"Subnets",
				)
			}

		},
	}
	return c
}

func loadSystemConfig(path string) (
	sysconf SystemConfig, err error,
) {
	err = csiFiles.ReadYAMLConfig(
		path,
		&sysconf,
	)
	return
}

func loadNetwork(path string) (
	network networking.IPV4Network, err error,
) {
	err = csiFiles.ReadJSONConfig(
		path,
		&network,
	)
	return
}

func extractNetworks(basepath string) (
	[]networking.IPV4Network, error,
) {
	// TODO: Handle incoming error?
	var networks []networking.IPV4Network
	err := filepath.Walk(
		basepath,
		func(
			path string, info os.FileInfo, err error,
		) error {
			if info.Mode().IsRegular() {
				log.Println(
					"Processing",
					path,
					"as ipam.IPV4Network -",
					info.Size(),
				)
				network, err := loadNetwork(path)
				if err != nil {
					log.Printf(
						"Failed loading network from %v: %v\n",
						path,
						err,
					)
				}
				networks = append(
					networks,
					network,
				)
			}
			return nil
		},
	)
	return networks, err
}
