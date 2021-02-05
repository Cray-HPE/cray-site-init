package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var lastFailure error
var livecdProvisioning, livecdPreflight, ncnPreflight, validateCeph, validateK8s, validateNetwork, validateServices bool

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Runs unit tests",
	Long:  `Runs unit tests and validates a working livecd and NCN deployment.`,
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: Replace with GOSS tests
		if validateNetwork {
			runCommand("ip a show lan0")
			runCommand("ip a show bond0")
			runCommand("ip a show vlan002")
			runCommand("ip a show vlan004")
			runCommand("ip a show vlan007")
		}

		// TODO: Replace with GOSS tests
		if validateServices {
			runCommand("systemctl status dnsmasq")
			runCommand("systemctl status nexus")
			runCommand("systemctl status conman")
			runCommand("systemctl status basecamp")
			runCommand("podman container ls -a")
		}

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
	validateCmd.Flags().BoolVarP(&validateNetwork, "network", "N", false, "Run network tests")
	validateCmd.Flags().BoolVarP(&validateServices, "services", "S", false, "Run services tests")
	validateCmd.Flags().BoolVarP(&livecdProvisioning, "livecd-provisioning", "l", false, "Run LiveCD provisioning tests")
	validateCmd.Flags().BoolVarP(&livecdPreflight, "livecd-preflight", "l", false, "Run LiveCD pre-flight tests")
	validateCmd.Flags().BoolVarP(&ncnPreflight, "ncn-preflight", "n", false, "Run NCN pre-flight tests")
	validateCmd.Flags().BoolVarP(&validateCeph, "ceph", "c", false, "Validate that Ceph is working")
	validateCmd.Flags().BoolVarP(&validateK8s, "k8s", "k", false, "Validate that Kubernetes is working")
}
