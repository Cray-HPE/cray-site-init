package cmd

import (
	"fmt"
	"github.com/Cray-HPE/cray-site-init/pkg/kubernetes"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	ncns []string

	hostnames []string
)

var automateNCNPreflight = &cobra.Command{
	Use:   "preflight",
	Short: "tools used to automate preflight checks",
	Long: "A series of subcommands that automates preflight checks around shutdown/reboot/rebuilt NCN lifecycle" +
		" activities.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		kubernetesClient, err = kubernetes.NewKubernetesClient(kubeconfig)
		if err != nil {
			// A bit extreme but hey, we're dead without it.
			log.Fatal(err)
		}

		// Enable logging.
		kubernetesClient.Logger.Logger = log.Default()

		if action == "verify-loss-acceptable" {
			verifyLossAcceptable()
		} else if action == "standardize-hostname" {
			setupEnvs()
			setupHTTPClient()

			standardizeHostnames, err := standardizeHostnames(hostnames)
			if err != nil {
				log.Fatalf("Failed to standardize hostname(s): %s", err)
			}

			fmt.Print(strings.Join(standardizeHostnames, ","))
		} else {
			log.Fatalf("Invalid action: %s!\n", action)
		}
	},
}

func init() {
	automateNCNCommand.AddCommand(automateNCNPreflight)

	automateNCNPreflight.Flags().StringVar(&kubeconfig, "kubeconfig", "",
		"Absolute path to the kubeconfig file")

	automateNCNPreflight.Flags().StringVar(&action, "action", "",
		"The etcd action to perform (verify-loss-acceptable/standardize-hostname)")
	_ = automateNCNPreflight.MarkFlagRequired("action")

	automateNCNPreflight.Flags().StringSliceVar(&ncns, "ncns", []string{},
		"The NCNs to assume go away (at least temporarily) for the action")
	automateNCNPreflight.Flags().StringSliceVar(&hostnames, "hostnames", []string{},
		"Hostname(s) to standardize so it will work with Ansible")
}

func standardizeHostname(hostname string) (string, error) {
	// This is an absolutely terrible check and should eventually be upgraded to use the full list of valid xnames but,
	// for now we don't begin any NCN conical names with `x`.
	if strings.HasPrefix(hostname, "x") {
		// Now we have to translate this to it's `ncnX-YYY` equivalent. SLS to the rescue!
		managementNCNs, err := getManagementNCNsFromSLS()
		if err != nil {
			return "", fmt.Errorf("failed to query SLS for management NCNs: %w", err)
		}

		for _, ncn := range managementNCNs {
			if ncn.Xname == hostname {
				var extraProperties sls_common.ComptypeNode
				err := mapstructure.Decode(ncn.ExtraPropertiesRaw, &extraProperties)
				if err != nil {
					return "", fmt.Errorf("failed to decode extra properties from SLS: %w", err)
				}

				for _, alias := range extraProperties.Aliases {
					if strings.HasPrefix(alias, "ncn") {
						return alias, nil
					}
				}

				return "", fmt.Errorf("node structure did not have NCN alias: %s", extraProperties.Aliases)
			}
		}

		return "", fmt.Errorf("failed to find given xname in SLS: %s", hostname)
	} else if strings.HasPrefix(hostname, "ncn") {
		// Good as we are.
		return hostname, nil
	}

	return "", fmt.Errorf("hostname format not recognized: %s", hostname)
}

func standardizeHostnames(hostnames []string) (translatedHostnames []string, err error) {
	for _, hostname := range hostnames {
		translatedHostname, err := standardizeHostname(hostname)
		if err != nil {
			return nil, err
		}

		translatedHostnames = append(translatedHostnames, translatedHostname)
	}

	return
}

func verifyLossAcceptable() {
	nodeMap, err := kubernetesClient.GetNodes()
	if err != nil {
		log.Fatalf("Failed to get node map: %s", err)
	}

	// Now remove all the suggested NCNs from the map...
	for _, targetNCN := range ncns {
		delete(nodeMap, targetNCN)
	}

	//...and see if we're left with enough of each type.
	numMasters := 0
	numWorkers := 0

	for _, node := range nodeMap {
		if kubernetes.IsMaster(node) {
			numMasters++
		} else {
			numWorkers++
		}
	}

	// Now check our conditions.
	if numMasters < kubernetes.MinMasters {
		log.Fatalf("Insufficant number of remaining masters (%d - need %d)!", numMasters, kubernetes.MinMasters)
	}
	if numWorkers < kubernetes.MinWorkers {
		log.Fatalf("Insufficant number of remaining workers (%d - need %d)!", numWorkers, kubernetes.MinWorkers)
	}

	log.Printf("Loss of %s is acceptable.", ncns)
}
