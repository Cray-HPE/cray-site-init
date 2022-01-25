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
var livecdProvisioning, livecdPreflight, ncnPreflight, validateCeph, validateK8s, validateNetwork, validateServices, validatePg bool

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
			runCommand("ip a show bond0.nmn0")
			runCommand("ip a show bond0.hmn0")
			runCommand("ip a show bond0.cmn0")
			runCommand("ip a show bond0.can0")
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
		lastFailure = err
		log.Fatalln(err)
	}
}

func init() {
	pitCmd.AddCommand(validateCmd)
	validateCmd.DisableAutoGenTag = true
	viper.SetEnvPrefix("pit")
	viper.AutomaticEnv()
	validateCmd.Flags().BoolVarP(&validateNetwork, "network", "N", false, "Run network tests")
	validateCmd.Flags().BoolVarP(&validateServices, "services", "S", false, "Run services tests")
	validateCmd.Flags().BoolVarP(&livecdProvisioning, "livecd-provisioning", "p", false, "Run LiveCD provisioning tests")
	validateCmd.Flags().BoolVarP(&livecdPreflight, "livecd-preflight", "l", false, "Run LiveCD pre-flight tests")
	validateCmd.Flags().BoolVarP(&ncnPreflight, "ncn-preflight", "n", false, "Run NCN pre-flight tests")
	validateCmd.Flags().BoolVarP(&validateCeph, "ceph", "c", false, "Validate that Ceph is working")
	validateCmd.Flags().BoolVarP(&validateK8s, "k8s", "k", false, "Validate that Kubernetes is working")
	validateCmd.Flags().BoolVarP(&validatePg, "postgres", "p", false, "Validate that Postgres clusters are healthy")
}
