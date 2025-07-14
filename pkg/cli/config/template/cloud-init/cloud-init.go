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

package cloudinit

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/spf13/cobra"
)

// NewCommand represents the 'cloud-init' sub-command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "cloud-init",
		Short:             "Process cloud-init templates",
		Long:              `Templates requested sections of cloud-init meta-data and/or user-date to file.`,
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
	}
	c.AddCommand(
		DisksCommand(),
	)
	return c
}

// DisksCommand represents the 'template disks' sub-command.
func DisksCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "disks",
		Short:             "Process cloud-init disk templates",
		Long:              `Process cloud-iniit meta-data for disks. Includes bootcmd, fs_setup, and mounts cloud-init user-data`,
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			WriteDiskTemplates()
		},
	}
	return c
}

// WriteDiskTemplates Write cloud-init user-data for disks (bootcmd, fs_setup, mounts) to files
func WriteDiskTemplates() {
	userDataMaps := map[string]map[string]map[string]interface{}{
		"ncn-master/cloud-init/user-data.json": {
			"user-data": {
				"bootcmd":  MasterBootCMD,
				"fs_setup": MasterFileSystems,
				"mounts":   MasterMounts,
			},
		},
		"ncn-worker/cloud-init/user-data.json": {
			"user-data": {
				"bootcmd":  WorkerBootCMD,
				"fs_setup": WorkerFileSystems,
				"mounts":   WorkerMounts,
			},
		},
		"ncn-storage/cloud-init/user-data.json": {
			"user-data": {
				"bootcmd":  CephBootCMD,
				"fs_setup": CephFileSystems,
				"mounts":   CephMounts,
			},
		},
	}

	for path, data := range userDataMaps {
		if cloudInitDir, err := filepath.Abs(filepath.Dir(path)); err == nil {
			if _, err = os.Stat(cloudInitDir); os.IsNotExist(err) {
				if err = os.MkdirAll(
					cloudInitDir,
					0755,
				); err != nil {
					log.Printf(
						"Error creating %s: %v",
						cloudInitDir,
						err,
					)
				}
			} else if err != nil {
				log.Printf(
					"%s",
					err,
				)
			}
		} else {
			log.Printf(
				"Error getting absolute path for %s: %v",
				path,
				err,
			)
		}

		if err := files.WriteJSONConfig(
			path,
			data,
		); err != nil {
			log.Printf(
				"Error writing %s: %v",
				path,
				err,
			)
		}
	}
}
