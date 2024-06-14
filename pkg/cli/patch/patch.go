/*
 MIT License

 (C) Copyright 2021-2024 Hewlett Packard Enterprise Development LP

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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/spf13/cobra"
)

var cloudInitSeedFile string

// NewCommand is a command for applying changes to an existing cloud-init file.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "patch",
		Short:             "Apply patch operations",
		DisableAutoGenTag: true,
		Long: `
Runs patch operations against the CRAY.
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
