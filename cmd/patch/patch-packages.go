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

package patch

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
	"os"
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

// Zypper is a map for Zypper cloud-init data.
type Zypper struct {
	Repos []Repo `json:"repos"`
}

// ConfigFileData matches the current “cloud-init“ file in the CSM tarball.
type ConfigFileData struct {
	Repos    []Repo   `json:"repos"`
	Packages Packages `json:"packages"`
}

// UserData is the cloud-init structure/representation of the new data.
type UserData struct {
	Zypper   Zypper   `json:"zypper"`
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

		packageData, err := os.ReadFile(cloudInitSeedFile)
		if err != nil {
			log.Fatalf("Unable to load cloud-init seed data, %v \n", err)
		}

		cloudInit := mergePackagesData(userdata, packageData)
		update, err := json.Marshal(cloudInit)
		if err != nil {
			log.Fatalf("Unable to marshal zypper data into JSON, %v \n", err)
		}

		data, err := backupCloudInitData()
		if err != nil {
			log.Fatalf("Failed to write backup file, %v \n", err)
		}

		if err := writeCloudInit(data, update); err != nil {
			log.Fatalf("Unable to patch cloud-init seed data in place, %v \n", err)
		}
		log.Println("Patched cloud-init seed data in place")
	},
}

func init() {
	PatchCmd.AddCommand(patchPackages)
	patchPackages.DisableAutoGenTag = true
	patchPackages.Flags().StringVarP(&configFile, "config-file", "", "", "Path to cloud-init.yaml")
	patchPackages.MarkFlagRequired("config-file")
}

// loadPackagesConfig reads a configFile, returning its contents as unmarshalled YAML.
func loadPackagesConfig(filePath string) (UserData, error) {

	var configData ConfigFileData
	var data UserData
	config, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}

	if err := yaml.Unmarshal(config, &configData); err != nil {
		return data, err
	}

	data.Zypper = Zypper{
		Repos: configData.Repos,
	}
	data.Packages = configData.Packages

	return data, err
}

// mergePackagesData takes assembled userdata and merges it into each user-data entry for each host in a given cloud-init datasource.
func mergePackagesData(userdata UserData, data []byte) CloudInitHost {
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
	return cloudInit
}
