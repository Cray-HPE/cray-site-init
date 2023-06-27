/*
 MIT License

 (C) Copyright 2023 Hewlett Packard Enterprise Development LP

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

package cmd

import (
	"encoding/json"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Repo is a list of repositories to add to the client node.
type Repo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Baseurl     string `json:"baseurl"`
	Enabled     int    `json:"enabled"`
	Autorefresh int    `json:"autorefresh"`
	Gpgcheck    int    `json:"gpgcheck"`
}

// Packages is a list of packages to install on the client node.
type Packages []string

// UserData is the cloud-init structure/representation of the new data.
type UserData struct {
	Repos    []Repo   `json:"repos"`
	Packages Packages `json:"packages"`
}

// Host is the cloud-init structure/representation of each of our host entries in the cloud-init data.
type Host struct {
	UserData UserData `json:"user-data"`
}

// CloudInitHost is the representation of all the new cloud-init data.
type CloudInitHost map[string]Host

// configFile is our input, our config to patch into cloud-init data.
var configFile string

// cloudInitFile is our file we're merging with.
var cloudInitFile string

var patchPackages = &cobra.Command{
	Use:   "packages",
	Short: "Patch cloud-init metadata with repositories and packages",
	Long: `
Patch cloud-init metadata (in place) with a list of repositories to add, and packages to install, during cloud-init from
CSM's cloud-init.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		userdata, err := loadPackagesConfig(configFile)
		if err != nil {
			log.Fatalf("Unable to load config data from CSM cloud-init config file, %v \n", err)
		}

		data, err := os.ReadFile(cloudInitFile)
		if err != nil {
			log.Fatalf("Unable to load cloud-init seed data, %v \n", err)
		}

		cloudInit := make(map[string]Host)
		var datas map[string]interface{}
		if err := json.Unmarshal(data, &datas); err != nil {
			log.Fatalf(string(data[:]))
		}
		for k := range datas {
			if k == "Global" {
				continue
			}
			var host Host
			host.UserData = userdata
			cloudInit[k] = host
		}

		update, err := json.Marshal(cloudInit)
		if err != nil {
			log.Fatalf("Unable to marshal zypper data into JSON, %v \n", err)
		}
		merged, err := jsonpatch.MergePatch(data, update)
		if err != nil {
			log.Fatalf("Could not create merge patch to update cloud-init seed data, %v \n", err)
		}

		currentTime := time.Now()
		ts := currentTime.Unix()

		cloudinitFileName := strings.TrimSuffix(cloudInitFile, filepath.Ext(cloudInitFile))
		backupFile := cloudinitFileName + "-" + fmt.Sprintf("%d", ts) + filepath.Ext(cloudInitFile)
		err = os.WriteFile(backupFile, data, 0640)
		if err != nil {
			log.Fatalf("Unable to create backup of cloud-init seed data, %v \n", err)
		}
		log.Println("Backup of cloud-init seed data at", backupFile)

		// Unmarshal merged cloud-init data, marshal it back with indent
		// then write it to the original cloud-init file (in place patch)
		var mergeUnmarshal map[string]interface{}
		json.Unmarshal(merged, &mergeUnmarshal)
		mergeMarshal, _ := json.MarshalIndent(mergeUnmarshal, "", "  ")
		err = os.WriteFile(cloudInitFile, mergeMarshal, 0640)
		if err != nil {
			log.Fatalf("Unable to patch cloud-init seed data in place, %v \n", err)
		}
		log.Println("Patched cloud-init seed data in place")
	},
}

func init() {
	patchCmd.AddCommand(patchPackages)
	patchPackages.DisableAutoGenTag = true
	patchPackages.Flags().StringVarP(&configFile, "config-file", "", "", "Path to cloud-init.yaml")
	patchPackages.Flags().StringVarP(&cloudInitFile, "cloud-init-seed-file", "", "", "Path to cloud-init metadata seed file")
	patchPackages.MarkFlagRequired("config-file")
	patchPackages.MarkFlagRequired("cloud-init-seed-file")
}

func loadPackagesConfig(filePath string) (UserData, error) {

	var data UserData
	config, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}

	if err := yaml.Unmarshal(config, &data); err != nil {
		return data, err
	}
	fmt.Println(data)

	return data, err
}
