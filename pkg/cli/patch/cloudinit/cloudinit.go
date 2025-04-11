/*
 MIT License

 (C) Copyright 2025 Hewlett Packard Enterprise Development LP

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

package cloudinit

import (
	"encoding/json"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var cloudInitSeedFile string

// DeprecatedCACommand is an exported version of the caCommand() that shims functionality for anyone using the old `csi patch ca` command.
func DeprecatedCACommand() *cobra.Command {
	c := caCommand()
	c.Deprecated = "This command has moved to `csi patch init ca`"
	c.Hidden = true
	return c
}

// DeprecatedPackagesCommand is an exported version of the caCommand() that shims functionality for anyone using the old `csi patch packages` command.
func DeprecatedPackagesCommand() *cobra.Command {
	c := packageCommand()
	c.Deprecated = "This command has moved to `csi patch init packages`"
	c.Hidden = true
	return c
}

// NewCommand creates the init sub-command..
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "init",
		Short:             "Patch operations for initial system data.",
		DisableAutoGenTag: true,
		Long: `
Runs patch operations against a CSM deployment environment (e.g. the Pre-Install Toolkit).
`,
		Run: func(c *cobra.Command, args []string) {
			c.Usage()
		},
	}
	c.PersistentFlags().StringVarP(
		&cloudInitSeedFile,
		"cloud-init-seed-file",
		"",
		"",
		"Path to cloud-init metadata seed file",
	)
	c.MarkFlagRequired("cloud-init-seed-file")
	c.AddCommand(
		caCommand(),
		packageCommand(),
	)
	return c
}

// backupCloudInitData makes a backup of the cloudInitSeedFile.
func backupCloudInitData() (
	[]byte, error,
) {
	data, err := os.ReadFile(cloudInitSeedFile)
	if err != nil {
		log.Fatalf(
			"Unable to load cloud-init seed data, %v \n",
			err,
		)
	}
	currentTime := time.Now()
	ts := currentTime.Unix()

	cloudinitFileName := strings.TrimSuffix(
		cloudInitSeedFile,
		filepath.Ext(cloudInitSeedFile),
	)
	backupFile := cloudinitFileName + "-" + fmt.Sprintf(
		"%d",
		ts,
	) + filepath.Ext(cloudInitSeedFile)
	err = os.WriteFile(
		backupFile,
		data,
		0640,
	)
	return data, err
}

// writeCloudInit merges new data with an existing cloud-init seed file, saving the merged data to disk.
func writeCloudInit(
	data []byte, update []byte,
) error {
	// Unmarshal merged cloud-init data, marshal it back with indent
	// then write it to the original cloud-init file (in place patch)
	merged, err := jsonpatch.MergePatch(
		data,
		update,
	)
	if err != nil {
		log.Fatalf(
			"Could not create merge patch to update cloud-init seed data, %v \n",
			err,
		)
	}
	var mergeUnmarshal map[string]interface{}
	json.Unmarshal(
		merged,
		&mergeUnmarshal,
	)
	mergeMarshal, _ := json.MarshalIndent(
		mergeUnmarshal,
		"",
		"  ",
	)
	err = os.WriteFile(
		cloudInitSeedFile,
		mergeMarshal,
		0640,
	)
	return err
}
