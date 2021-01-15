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
)

var lastFailure error
var validateNetwork, validateServices, validateDNS, validateMtu, validateCeph, validateK8s, validateAll bool

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates the PIT liveCD during setup",
	Long:  `Validates certain requirements needed for effectively running the liveCD.`,
	Run: func(cmd *cobra.Command, args []string) {

		if validateServices || validateAll {
			log.Println("[csi] VALIDATING SERVICES")
			runCommand("systemctl status dnsmasq")
			runCommand("systemctl status nexus")
			runCommand("systemctl status basecamp")
			runCommand("podman container ls -a")
		}

		if validateNetwork || validateAll {
			log.Println("[csi] VALIDATING NETWORK")
			runCommand("ip a show lan0")
			runCommand("ip a show bond0")
			runCommand("ip a show vlan002")
			runCommand("ip a show vlan004")
			runCommand("ip a show vlan007")
		}

		if validateDNS || validateAll {
			log.Println("[csi] VALIDATING DNS")
			runCommand("grep -Eo ncn-.*-mgmt /var/lib/misc/dnsmasq.leases | sort")
		}

		if validateMtu || validateAll {
			log.Println("[csi] VALIDATING MTU")
			log.Printf("[csi] MANUAL ACTION: run the following snippet on a SPINE switch if reachable and verify MTU of the NCN ports is set to 9216\n\n\t# show interface status | include ^Mpo\n\n")
		}

		if validateCeph || validateAll {
			log.Println("[csi] VALIDATING CEPH")
			log.Printf("[csi] MANUAL ACTION: run the following snippet on a STORAGE node if booted and verify ceph quroum (should see all 3 storage in report)\n\n\t# ceph -s\n\n")
		}

		if validateK8s || validateAll {
			log.Println("[csi] VALIDATING K8S")
			log.Printf("[csi] MANUAL ACTION 1: run the following snippet on a STORAGE node if booted and verify if 3 classes are available\n\n\t# kubectl get storageclass\n\n")
			log.Printf("[csi] MANUAL ACTION 2: run the following snippet on a MANAGER node if booted to verify all nodes are in the cluister\n\n\t# kubectl get nodes\n\n")
			log.Printf("[csi] MANUAL ACTION 3: run the following snippet on a MANAGER node if booted to verify all nodes are in the cluister\n\n\t# kubectl get po -n kube-system\n\n")
		}

		// For now, if lastFailure is set then we failed.
		if lastFailure != nil {
			log.Fatal("Failed; see scrollback for test errors.")
		}
	},
}

func runCommand(shellCode string) {
	log.Println("[csi] Running...", shellCode)
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
	validateCmd.Flags().BoolVarP(&validateNetwork, "network", "n", viper.GetBool("validate_network"), "Validate the network when booted into the LiveCD (env: PIT_VALIDATE_NETWORK)")
	validateCmd.Flags().BoolVarP(&validateServices, "services", "s", viper.GetBool("validate_services"), "Validate services when booted into the LiveCD (env: PIT_VALIDATE_SERVICES)")
	validateCmd.Flags().BoolVarP(&validateDNS, "dns-dhcp", "d", viper.GetBool("validate_dns_dhcp"), "Validate the DNS leases (env: PIT_VALIDATE_DNS_DHCP)")
	validateCmd.Flags().BoolVarP(&validateMtu, "mtu", "m", viper.GetBool("validate_mtu"), "Validate the MTU of the spine ports (env: PIT_VALIDATE_MTU)")
	validateCmd.Flags().BoolVarP(&validateCeph, "ceph", "c", viper.GetBool("validate_ceph"), "Validate Ceph is working (env: PIT_VALIDATE_CEPH)")
	validateCmd.Flags().BoolVarP(&validateK8s, "k8s", "k", viper.GetBool("validate_k8s"), "Validate Kubernetes is working (env: PIT_VALIDATE_K8S)")
	validateCmd.Flags().BoolVarP(&validateAll, "all", "a", viper.GetBool("validate_all"), "Validate everything (env: PIT_VALIDATE_ALL)")
}
