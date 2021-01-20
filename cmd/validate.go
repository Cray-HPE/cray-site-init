package cmd

/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os/exec"
	"path/filepath"
)

var lastFailure error
var livecdPreflight, ncnPreflight, validateCeph, validateK8s bool

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Runs unit tests",
	Long:  `Runs unit tests and validates a working livecd and NCN deployment.`,
	Run: func(cmd *cobra.Command, args []string) {

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
	},
}

func runCommand(shellCode string) {
	cmd := exec.Command("bash", "-c", shellCode)
	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		lastFailure = err
		log.Println(err)
	}
}

func init() {
	pitCmd.AddCommand(validateCmd)
	viper.SetEnvPrefix("pit")
	viper.AutomaticEnv()
	validateCmd.Flags().BoolVarP(&livecdPreflight, "livecd-preflight", "l", false, "Run LiveCD pre-flight tests")
	validateCmd.Flags().BoolVarP(&ncnPreflight, "ncn-preflight", "n", false, "Run NCN pre-flight tests")
	validateCmd.Flags().BoolVarP(&validateCeph, "ceph", "c", false, "Validate that Ceph is working")
	validateCmd.Flags().BoolVarP(&validateK8s, "k8s", "k", false, "Validate that Kubernetes is working")
}
