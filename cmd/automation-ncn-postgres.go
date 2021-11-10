package cmd

import (
	"github.com/Cray-HPE/cray-site-init/pkg/postgres"
	"github.com/spf13/cobra"
	"log"
)

var (
	postgresClient *postgres.UtilsClient
)

var automateNCNPostgresCommand = &cobra.Command{
	Use:   "postgres",
	Short: "tools used to automate actions to Postgres clusters",
	Long:  "A series of subcommands that automates administrative tasks to Postgres clusters running in the mesh.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		postgresClient, err = postgres.NewPostgresClient(kubeconfig)
		if err != nil {
			// A bit extreme but hey, we're dead without it.
			log.Fatal(err)
		}

		// Enable logging.
		postgresClient.Logger.Logger = log.Default()

		if action == "failover-leaders" {
			err := postgresClient.FailoverPostgresLeaders(ncn)
			if err != nil {
				log.Fatalf("Failed to failover Postgres leaders on %s: %s", ncn, err)
			}

			log.Printf("All Postgres leaders on %s now failed over.", ncn)
		} else if action == "check-all-clusters" {
			err := postgresClient.HealthCheckAllClusters()
			if err != nil {
				log.Fatalf("Cluster health check failed: %s", err)
			}

			log.Printf("All Postgres clusters healthy.")
		} else {
			log.Fatalf("Invalid action: %s!\n", action)
		}
	},
}

func init() {
	automateNCNCommand.AddCommand(automateNCNPostgresCommand)

	automateNCNPostgresCommand.Flags().StringVar(&kubeconfig, "kubeconfig", "",
		"Absolute path to the kubeconfig file")

	automateNCNPostgresCommand.Flags().StringVar(&action, "action", "",
		"The action to perform (failover-leaders/check-all-clusters)")
	_ = automateNCNPostgresCommand.MarkFlagRequired("action")

	automateNCNPostgresCommand.Flags().StringVar(&ncn, "ncn", "",
		"The NCN to perform the action on")
	//_ = automateNCNETCDCommand.MarkFlagRequired("ncn")
}
