package cmd

import (
	"log"
	"os"

	"github.com/Cray-HPE/csm-common/go/pkg/kubernetes"
	"github.com/spf13/cobra"
)

var automateNCNKubernetesCommand = &cobra.Command{
	Use:   "kubernetes",
	Short: "tools used to automate actions to Kubernetes",
	Long:  "A series of subcommands that automates administrative tasks to Kubernetes.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		kubernetesClient, err = kubernetes.NewKubernetesClient(kubeconfig)
		if err != nil {
			// A bit extreme but hey, we're dead without it.
			log.Fatal(err)
		}

		// Enable logging.
		kubernetesClient.Logger.Logger = log.Default()

		if action == "cordon-ncn" {
			err := kubernetesClient.CordonNCN(ncn)
			if err != nil {
				log.Fatalf("Failed to cordon %s: %s", ncn, err)
			}

			log.Printf("%s cordoned.", ncn)
		} else if action == "uncordon-ncn" {
			err := kubernetesClient.UnCordonNCN(ncn)
			if err != nil {
				log.Fatalf("Failed to uncordon %s: %s", ncn, err)
			}

			log.Printf("%s uncordoned.", ncn)
		} else if action == "drain-ncn" {
			err := kubernetesClient.DrainNCN(ncn)
			if err != nil {
				log.Fatalf("Failed to drain %s: %s", ncn, err)
			}

			log.Printf("%s drained.", ncn)
		} else if action == "delete-ncn" {
			err := kubernetesClient.DeleteNCN(ncn)
			if err != nil {
				log.Fatalf("Failed to delete %s: %s", ncn, err)
			}

			log.Printf("%s deleted.", ncn)
		} else if action == "is-member" {
			isMember, err := kubernetesClient.IsMember(ncn)
			if err != nil {
				log.Fatalf("Failed to determine membership for %s: %s", ncn, err)
			}

			log.Printf("%s is Kubernetes member: %t", ncn, isMember)

			if isMember {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		} else {
			log.Fatalf("Invalid action: %s!\n", action)
		}
	},
}

func init() {
	automateNCNCommand.AddCommand(automateNCNKubernetesCommand)
	automateCommand.DisableAutoGenTag = true

	automateNCNKubernetesCommand.Flags().StringVar(&kubeconfig, "kubeconfig", "",
		"Absolute path to the kubeconfig file")

	automateNCNKubernetesCommand.Flags().StringVar(&action, "action", "",
		"The etcd action to perform (cordon-ncn/uncordon-ncn/drain-ncn/delete-ncn/is-member)")
	_ = automateNCNKubernetesCommand.MarkFlagRequired("action")

	automateNCNKubernetesCommand.Flags().StringVar(&ncn, "ncn", "",
		"The NCN to perform the action on")
	_ = automateNCNKubernetesCommand.MarkFlagRequired("ncn")
}
