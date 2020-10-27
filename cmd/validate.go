package cmd
/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"
	"os/exec"
	// "log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strconv"
)

var validateNetwork, validateServices, validateDNS, validateMtu, validateCeph, validateK8s, validateAll string

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates the SPIT liveCD during setup",
	Long: `Validates certain requirements needed for effectively running the liveCD.`,
	Run: func(cmd *cobra.Command, args []string) {

		if s, err := strconv.ParseBool(validateServices); err == nil {
			fmt.Println("[sic] VALIDATING SERVICES: ", s)
			runCommand("systemctl status dnsmasq")
			runCommand("systemctl status nexus")
			runCommand("systemctl status basecamp")
			runCommand("podman container ls -a")
		}

		if n, err := strconv.ParseBool(validateNetwork); err == nil {
			fmt.Println("[sic] VALIDATING NETWORK: ", n)
			runCommand("ip a show lan0")
			runCommand("ip a show bond0")
			runCommand("ip a show vlan002")
			runCommand("ip a show vlan004")
			runCommand("ip a show vlan007")
		}

		if d, err := strconv.ParseBool(validateDNS); err == nil {
			fmt.Println("[sic] VALIDATING DNS: ", d)
			runCommand("grep -Eo ncn-.*-mgmt /var/lib/misc/dnsmasq.leases | sort")
		}

		if m, err := strconv.ParseBool(validateMtu); err == nil {
			fmt.Println("[sic] VALIDATING MTU: ", m)
			fmt.Println("[sic] MANUAL ACTION: verify the MTU of the spine ports connected to the NCNs is set to 9216")
		}

		if c, err := strconv.ParseBool(validateCeph); err == nil {
			fmt.Println("[sic] VALIDATING CEPH: ", c)
			fmt.Println("[sic] MANUAL ACTION: run 'ceph -s' on a storage node if booted")
		}

		if k, err := strconv.ParseBool(validateK8s); err == nil {
			fmt.Println("[sic] VALIDATING K8S: ", k)
			fmt.Println("[sic] MANUAL ACTION: run 'kubectl get storageclass' on a storage node if booted and verify if 3 classes are available")
			fmt.Println("[sic] MANUAL ACTION: run 'kubectl get nodes' on a manager node if booted to verify all nodes are in the cluister")
			fmt.Println("[sic] MANUAL ACTION: run 'kubectl get po -n kube-system' on a manager node if booted to verify all nodes are in the cluister")
		}
	},
}

func runCommand(shellCode string) {
	fmt.Println("[sic] Running...", shellCode)
	cmd := exec.Command("bash", "-c", shellCode)
	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		// Don't fail yet.  For now, we're just automating what humans currently do
		// This also gives an overview of the current state of things in one command
		// log.Fatal(err)
		fmt.Println(err)
	}
}


func init() {
	spitCmd.AddCommand(validateCmd)
	viper.SetEnvPrefix("spit") // will be uppercased automatically
	viper.AutomaticEnv()
	validateCmd.Flags().StringVarP(&validateNetwork, "network", "n", viper.GetString("validate_network"), "Validate the network when booted into the LiveCD (env: SPIT_VALIDATE_NETWORK)")
	validateCmd.Flags().StringVarP(&validateServices, "services", "s", viper.GetString("validate_services"), "Validate services when booted into the LiveCD (env: SPIT_VALIDATE_SERVICES)")
	validateCmd.Flags().StringVarP(&validateDNS, "dns", "d", viper.GetString("validate_dns"), "Validate the DNS leases (env: SPIT_VALIDATE_DNS)")
	validateCmd.Flags().StringVarP(&validateMtu, "mtu", "m", viper.GetString("validate_mtu"), "Validate the MTU of the spine ports (env: SPIT_VALIDATE_MTU)")
	validateCmd.Flags().StringVarP(&validateCeph, "ceph", "c", viper.GetString("validate_ceph"), "Validate Ceph is working (env: SPIT_VALIDATE_CEPH)")
	validateCmd.Flags().StringVarP(&validateK8s, "k8s", "k", viper.GetString("validate_k8s"), "Validate Kubernetes is working (env: SPIT_VALIDATE_K8S)")
	validateCmd.Flags().StringVarP(&validateAll, "all", "a", viper.GetString("validate_all"), "Validate everything (env: SPIT_VALIDATE_ALL)")
}
