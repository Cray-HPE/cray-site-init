/*
 MIT License

 (C) Copyright 2021-2025 Hewlett Packard Enterprise Development LP

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

package patch

import (
	"github.com/Cray-HPE/cray-site-init/pkg/cli/patch/cloudinit"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/patch/sls"
	"github.com/spf13/cobra"
)

// NewCommand is a command for applying changes to a system's running services.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "patch",
		Short:             "Apply patch operations",
		DisableAutoGenTag: true,
		Long: `
Runs patch operations against a running system or a fresh install environment.
`,
		Run: func(c *cobra.Command, args []string) {
			c.Usage()
		},
	}
	c.AddCommand(
		cloudinit.NewCommand(),
		sls.NewCommand(),

		// Legacy commands
		cloudinit.DeprecatedCACommand(),
		cloudinit.DeprecatedPackagesCommand(),
	)
	return c
}
