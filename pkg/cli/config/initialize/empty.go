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

package initialize

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/pkg/version"
)

// emptyCommand represents the empty sub-command.
func emptyCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "empty",
		Short:             "Write a empty config file.",
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
		Run: func(c *cobra.Command, args []string) {

			// Initialize the global viper
			v := viper.GetViper()
			err := v.BindPFlags(c.Parent().Flags())
			if err != nil {
				log.Fatalln(err)
			}

			path, err := os.Getwd()
			if err != nil {
				log.Println(err)
			}
			v.SetConfigType("yaml")
			v.Set(
				"VersionInfo",
				version.Get(),
			)
			err = v.SafeWriteConfigAs("system_config.yaml")
			if err != nil {
				log.Fatal(err)
			}
			log.Printf(
				"Empty config file written to: %s/system_config.yaml\n",
				path,
			)
		},
	}
	return c
}
