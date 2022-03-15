//
//  MIT License
//
//  (C) Copyright 2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//

package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var livecdProvisioning, livecdPreflight, ncnPreflight, validateCeph, validateK8s, validateServices, validatePg bool

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Runs unit tests",
	Long:  `Runs unit tests and validates a working livecd and NCN deployment.`,
	Run: func(cmd *cobra.Command, args []string) {

		if livecdProvisioning {
			runCommand(filepath.Join("/opt/cray/tests/install/livecd/automated/", "livecd-provisioning-checks"))
		}

		if livecdPreflight {
			runCommand(filepath.Join("/opt/cray/tests/install/livecd/automated/", "livecd-preflight-checks"))
		}

		if ncnPreflight {
			runCommand(filepath.Join("/opt/cray/tests/install/livecd/automated/", "ncn-preflight-checks"))
		}

		if validateCeph {
			runCommand(filepath.Join("/opt/cray/tests/install/livecd/automated/", "ncn-storage-checks"))
		}

		if validateK8s {
			runCommand(filepath.Join("/opt/cray/tests/install/livecd/automated/", "ncn-kubernetes-checks"))
		}

		if validatePg {
			runCommand(filepath.Join("/opt/cray/tests/install/ncn/scripts/", "postgres_clusters_running.sh"))
			runCommand(filepath.Join("/opt/cray/tests/install/ncn/scripts/", "postgres_pods_running.sh") + " -p")
			runCommand(filepath.Join("/opt/cray/tests/install/ncn/scripts/", "postgres_clusters_leader.sh") + " -p")
			runCommand(filepath.Join("/opt/cray/tests/install/ncn/scripts/", "postgres_replication_lag.sh") + " -p -e")
		}
	},
}

func runCommand(shellCode string) {
	cmd := exec.Command("bash", "-c", shellCode)
	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		log.Fatalln(err)
	}
}

func init() {
	pitCmd.AddCommand(validateCmd)
	validateCmd.DisableAutoGenTag = true
	viper.SetEnvPrefix("pit")
	viper.AutomaticEnv()
	validateCmd.Flags().BoolVarP(&livecdProvisioning, "livecd-provisioning", "p", false, "Run LiveCD provisioning tests")
	validateCmd.Flags().BoolVarP(&livecdPreflight, "livecd-preflight", "l", false, "Run LiveCD pre-flight tests")
	validateCmd.Flags().BoolVarP(&ncnPreflight, "ncn-preflight", "n", false, "Run NCN pre-flight tests")
	validateCmd.Flags().BoolVarP(&validateCeph, "ceph", "c", false, "Validate that Ceph is working")
	validateCmd.Flags().BoolVarP(&validateK8s, "k8s", "k", false, "Validate that Kubernetes is working")
	validateCmd.Flags().BoolVar(&validatePg, "postgres", false, "Validate that Postgres clusters are healthy")
}
