// +build integration
// +build !shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cray-HPE/cray-site-init/pkg/csi"
)

// ConfigInitTest runs 'csi config init' on a system passed to it using the cobra command object
func ConfigInitTest(system string) {

	var cwd string

	// pseudo-pushd: Move into the directory (this should have been created in get_test.go)
	confdir, _ := filepath.Abs(system)
	os.Chdir(filepath.Join(confdir))
	cwd, _ = os.Getwd()
	log.Printf("pushd  ===> %s", cwd)

	// Runs 'config init' without any arguments (this requires system_config.yaml to be present in the dir)
	// csi.ExecuteCommandC(rootCmd, []string{"config", "init"})
	conf := confdir + "/system_config.yaml"
	csi.ExecuteCommandC(rootCmd, []string{"--config", conf, "config", "init"})

	// pseudo-popd
	os.Chdir(filepath.Join(".."))
	cwd, _ = os.Getwd()
	log.Printf("popd  ===> %s", cwd)

}

func TestConfigInit_GeneratePayload(t *testing.T) {

	var systems []string

	// The sample data set of systems names to use
	// For now, these should align with the system names in get_test.go
	if os.Getenv("CSI_SHASTA_SYSTEMS") != "" {

		systems = strings.Split(os.Getenv("CSI_SHASTA_SYSTEMS"), " ")

	} else {

		t.Errorf("CSI_SHASTA_SYSTEMS needs to be set")

	}

	for _, s := range systems {
		// Run 'config init' on each system's set of seed files
		// These files should have been previously gathered from get_test.go
		ConfigInitTest(s)
	}
}

// TODO: test 'config init' using command line flags instead of system_config.yaml
// TODO: test--with some tolerance--that generated output is similiar to what we know is a good config
