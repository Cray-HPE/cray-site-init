/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

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

package ncn

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/Cray-HPE/cray-site-init/pkg/kubernetes"
)

func kubernetesCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "kubernetes",
		Short:             "tools used to automate actions to Kubernetes",
		Long:              "A series of subcommands that automates administrative tasks to Kubernetes.",
		DisableAutoGenTag: true,
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
					log.Fatalf(
						"Failed to cordon %s: %s",
						ncn,
						err,
					)
				}

				log.Printf(
					"%s cordoned.",
					ncn,
				)
			} else if action == "uncordon-ncn" {
				err := kubernetesClient.UnCordonNCN(ncn)
				if err != nil {
					log.Fatalf(
						"Failed to uncordon %s: %s",
						ncn,
						err,
					)
				}

				log.Printf(
					"%s uncordoned.",
					ncn,
				)
			} else if action == "drain-ncn" {
				err := kubernetesClient.DrainNCN(ncn)
				if err != nil {
					log.Fatalf(
						"Failed to drain %s: %s",
						ncn,
						err,
					)
				}

				log.Printf(
					"%s drained.",
					ncn,
				)
			} else if action == "delete-ncn" {
				err := kubernetesClient.DeleteNCN(ncn)
				if err != nil {
					log.Fatalf(
						"Failed to delete %s: %s",
						ncn,
						err,
					)
				}

				log.Printf(
					"%s deleted.",
					ncn,
				)
			} else if action == "is-member" {
				isMember, err := kubernetesClient.IsMember(ncn)
				if err != nil {
					log.Fatalf(
						"Failed to determine membership for %s: %s",
						ncn,
						err,
					)
				}

				log.Printf(
					"%s is Kubernetes member: %t",
					ncn,
					isMember,
				)

				if isMember {
					os.Exit(0)
				} else {
					os.Exit(1)
				}
			} else {
				log.Fatalf(
					"Invalid action: %s!\n",
					action,
				)
			}
		},
	}

	c.Flags().StringVar(
		&action,
		"action",
		"",
		"The etcd action to perform (cordon-ncn/uncordon-ncn/drain-ncn/delete-ncn/is-member)",
	)
	_ = c.MarkFlagRequired("action")

	c.Flags().StringVar(
		&ncn,
		"ncn",
		"",
		"The NCN to perform the action on",
	)
	_ = c.MarkFlagRequired("ncn")
	return c
}
