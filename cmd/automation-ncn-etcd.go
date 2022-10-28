/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
*/

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
