// +build integration
// +build !shcd

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
