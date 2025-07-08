/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

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

package csi

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cray-HPE/cray-site-init/pkg/cli"
)

// ConfigInitTest runs 'csi config init' on a system passed to it using the cobra command object
func ConfigInitTest(system string) {

	var cwd string

	// pseudo-pushd: Move into the directory (this should have been created in get_test.go)
	confdir, _ := filepath.Abs(system)
	os.Chdir(filepath.Join(confdir))
	cwd, _ = os.Getwd()
	log.Printf(
		"pushd  ===> %s",
		cwd,
	)

	// Runs 'config init' without any arguments (this requires system_config.yaml to be present in the dir)
	// csi.ExecuteCommandC(rootCmd, []string{"config", "init"})
	conf := fmt.Sprintf(
		"%s/%s",
		confdir,
		cli.ConfigFilename,
	)
	ExecuteCommandC(
		NewCommand(),
		[]string{
			"--config",
			conf,
			"config",
			"init",
			"--cmn-gateway",
			"10.99.0.1",
		},
	)

	// pseudo-popd
	os.Chdir(filepath.Join(".."))
	cwd, _ = os.Getwd()
	log.Printf(
		"popd  ===> %s",
		cwd,
	)

}

func TestConfigInit_GeneratePayload(t *testing.T) {

	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	var systems []string

	// The sample data set of systems names to use
	// For now, these should align with the system names in get_test.go
	if os.Getenv("CSI_SHASTA_SYSTEMS") != "" {

		systems = strings.Split(
			os.Getenv("CSI_SHASTA_SYSTEMS"),
			" ",
		)

	} else {
		if os.Getenv("CI") == "" {
			t.Skip("Skipping test due to missing CI environment. Set CSI_SHASTA_SYSTEMS to override.")
		}
	}

	for _, s := range systems {
		// Run 'config init' on each system's set of seed files
		// These files should have been previously gathered from get_test.go
		ConfigInitTest(s)
	}
}

// TODO: test 'config init' using command line flags instead of system_config.yaml
// TODO: test--with some tolerance--that generated output is similar to what we know is a good config
