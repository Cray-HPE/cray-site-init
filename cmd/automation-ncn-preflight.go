package cmd

import (
	"github.com/Cray-HPE/cray-site-init/pkg/kubernetes"
	"github.com/spf13/cobra"
	"log"
)

var (
	ncns []string
)

var automateNCNPreflight = &cobra.Command{
	Use:   "preflight",
	Short: "tools used to automate preflight checks",
	Long:  "A series of subcommands that automates preflight checks around shutdown/reboot/rebuilt NCN lifecycle" +
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
		"The etcd action to perform (verify-loss-acceptable)")
	_ = automateNCNPreflight.MarkFlagRequired("action")

	automateNCNPreflight.Flags().StringSliceVar(&ncns, "ncns", []string{},
		"The NCNs to assume go away (at least temporarily) for the action")
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
	if numMasters < kubernetes.MIN_MASTERS {
		log.Fatalf("Insufficant number of remaining masters (%d - need %d)!", numMasters, kubernetes.MIN_MASTERS)
	}
	if numWorkers < kubernetes.MIN_WORKERS {
		log.Fatalf("Insufficant number of remaining workers (%d - need %d)!", numWorkers, kubernetes.MIN_WORKERS)
	}

	log.Printf("Loss of %s is acceptable.", ncns)
}
