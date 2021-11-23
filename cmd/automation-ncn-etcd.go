package cmd

import (
	"fmt"
	"github.com/Cray-HPE/cray-site-init/pkg/etcd"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	etcdClient *etcd.UtilsClient

	endpoints []string
)

var automateNCNETCDCommand = &cobra.Command{
	Use:   "etcd",
	Short: "tools used to automate actions to bare-metal etcd",
	Long:  "A series of subcommands that automates administrative tasks to the bare-metal etcd cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		etcdClient, err = etcd.NewETCDClient(endpoints, kubeconfig)
		if err != nil {
			// Bit extreme but hey, we're dead without it.
			log.Fatal(err)
		}
		defer etcdClient.CloseETCDClient()

		if action == "is-member" {
			if ncn == "" {
				log.Fatal("NCN must be given!")
			}

			isMember, err := etcdClient.IsMember(ncn)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("%s is etcd member: %t\n", ncn, isMember)

			if isMember {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		} else if action == "is-healthy" {
			err := etcdClient.ClusterIsHealthy()
			if err != nil {
				fmt.Printf("Cluster is NOT healthy: %s\n", err)
				os.Exit(1)
			} else {
				fmt.Printf("Cluster is healthy.\n")
				os.Exit(0)
			}
		} else if action == "add-member" {
			if ncn == "" {
				log.Fatal("NCN must be given!")
			}

			memberAdded, err := etcdClient.AddMember(ncn)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("%s added: %t\n", ncn, memberAdded)

			if memberAdded {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		} else if action == "remove-member" {
			if ncn == "" {
				log.Fatal("NCN must be given!")
			}

			memberRemoved, err := etcdClient.RemoveMember(ncn)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("%s removed: %t\n", ncn, memberRemoved)

			if memberRemoved {
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
	automateNCNCommand.AddCommand(automateNCNETCDCommand)
	automateCommand.DisableAutoGenTag = true

	automateNCNETCDCommand.Flags().StringVar(&kubeconfig, "kubeconfig", "",
		"Absolute path to the kubeconfig file")

	automateNCNETCDCommand.Flags().StringVar(&action, "action", "",
		"The etcd action to perform (is-member/is-healthy)")
	_ = automateNCNETCDCommand.MarkFlagRequired("action")

	automateNCNETCDCommand.Flags().StringVar(&ncn, "ncn", "",
		"The NCN to perform the action on")
	//_ = automateNCNETCDCommand.MarkFlagRequired("ncn")

	automateNCNETCDCommand.Flags().StringArrayVar(&endpoints, "endpoints",
		[]string{"ncn-m001.nmn:2379", "ncn-m002.nmn:2379", "ncn-m003.nmn:2379"},
		"List of endpoints to connect to")
}
