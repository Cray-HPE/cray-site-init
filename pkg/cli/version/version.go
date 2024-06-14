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

package version

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/pkg/version"
)

// NewCommand represents the version subcommand.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "version",
		DisableAutoGenTag: true,
		Short:             "version",
		Run: func(c *cobra.Command, args []string) {
			v := viper.GetViper()
			v.BindPFlags(c.Flags())
			clientVersion := version.Get()
			if v.GetBool("git") {
				fmt.Println(clientVersion.GitCommit)
				os.Exit(0)
			}
			switch output := v.GetString("output"); output {
			case "pretty":
				fmt.Println("CRAY-Site-Init build signature...")
				fmt.Printf(
					"%-15s: %s\n",
					"Build Commit",
					clientVersion.GitCommit,
				)
				fmt.Printf(
					"%-15s: %s\n",
					"Build Time",
					clientVersion.BuildDate,
				)
				fmt.Printf(
					"%-15s: %s\n",
					"Go Version",
					clientVersion.GoVersion,
				)
				fmt.Printf(
					"%-15s: %s\n",
					"Version",
					clientVersion.Version,
				)
				fmt.Printf(
					"%-15s: %s\n",
					"Platform",
					clientVersion.Platform,
				)
			case "json":
				b, err := json.Marshal(clientVersion)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(b))
			}
		},
	}
	c.Flags().StringP(
		"output",
		"o",
		"pretty",
		"output format pretty,json",
	)
	c.Flags().BoolP(
		"git",
		"g",
		false,
		"Simple commit sha of the source tree on a single line. \"-dirty\" added to the end if uncommitted changes present",
	)
	return c
}
