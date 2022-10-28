/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

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

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGet_ShastaSystemConfigs(t *testing.T) {

	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	// These are the five input files needed (two are optional) for running 'csi config init'
	configs := []string{"application_node_config.yaml",
		"cabinets.yaml",
		"hmn_connections.json",
		"ncn_metadata.csv",
		"switch_metadata.csv",
		"system_config.yaml"}

	var systems []string
	var configURL, url string

	// The sample data set of systems names to use
	// For now, this should align with the system names in init_test.go
	if os.Getenv("CSI_SHASTA_SYSTEMS") != "" {

		systems = strings.Split(os.Getenv("CSI_SHASTA_SYSTEMS"), " ")

	} else {

		t.Errorf("CSI_SHASTA_SYSTEMS needs to be set")

	}

	for _, s := range systems {

		// Make a directory to hold each set of configs
		os.MkdirAll(filepath.Join(s), 0755)

		for _, f := range configs {

			configURL = os.Getenv("CSI_SHASTA_CONFIG_URL")

			if configURL != "" {

				url = configURL + s + "/" + f

			} else {

				t.Errorf("CSI_SHASTA_CONFIG_URL needs to be set")

			}

			// Download the file
			output := GetArtifact(filepath.Join(s), url)

			if output != nil {
				t.Errorf("Expected no error, but got %s", output)
			}
		}
	}
}
