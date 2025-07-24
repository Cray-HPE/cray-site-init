/*
 MIT License

 (C) Copyright 2025 Hewlett Packard Enterprise Development LP

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

package csm

import (
	"fmt"
	"log"
	"net/netip"
	"os"
	"path"
	"strings"

	"github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/cli"
	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/bss"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	slsClient "github.com/Cray-HPE/hms-sls/v2/pkg/sls-client"
	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	defaultSubnetsToPatch = []string{"network_hardware", "bootstrap_dhcp"}
	networksToPatch       = []string{"CHN", "CMN"}
)

var (
	backupDirectory string
	commit          bool
	force           bool
	forceAlert      = false
	toPatch         NetworkParams
	remove          bool
	subnetsToPatch  = defaultSubnetsToPatch
	bssSession      bss.UtilsClient
	slsSession      slsClient.SLSClient
)

type NetworkParam struct {
	CIDRStr    *string
	CIDR       netip.Prefix
	GatewayStr *string
	Gateway    netip.Addr
}

type NetworkParams map[string]NetworkParam

// NewCommand is a command for applying changes to an existing cloud-init file.
func ipv6Command() *cobra.Command {
	c := &cobra.Command{
		Use:               "ipv6",
		Short:             "Retroactively adds IPv6 data to CSM.",
		DisableAutoGenTag: true,
		Long: fmt.Sprintf(
			`
Patches a Cray System Management (CSM) deployment, modifying the System Layout Service (SLS) and the
Boot Script Service (BSS) for IPv6 enablement.

This command runs as a dry-run unless its --commit flag is present. As a dry-run, no changes are committed
to BSS or SLS, and all discovered changes are written to the local filesystem.

Backups of BSS and SLS will be created for inspection or rollback (using a backup to rollback to requires using the
CSM API, or its direct tools such as CrayCLI).

Only certain SLS subnets defined within the Customer Management Network and (if CSM has one configured) the
Customer High-Speed Network are targeted, by default the targeted subnets are:
- %s

NOTE: Any IPv6 enablement already present in SLS and BSS, such as reserved IP addresses, will not be overwritten unless
the force flag is given.`,
			strings.Join(
				defaultSubnetsToPatch,
				"\n- ",
			),
		),
		Run: func(c *cobra.Command, args []string) {
			v := viper.GetViper()
			err := v.BindPFlags(c.Flags())
			if err != nil {
				log.Fatalln(err)
			}
			for networkName, network := range toPatch {
				prefix, err := netip.ParsePrefix(*network.CIDRStr)
				if err != nil {
					network.CIDR = netip.Prefix{}
				} else {
					network.CIDR = prefix
				}

				gateway, err := netip.ParseAddr(*network.GatewayStr)
				if err != nil {
					if prefix.IsValid() && !prefix.Addr().IsUnspecified() {
						gateway = networking.FindGatewayIP(prefix)
						log.Printf(
							"%s (%s) gateway was not given, auto-resolved to [%s]",
							networkName,
							prefix,
							gateway,
						)
					}
				}
				network.Gateway = gateway
				toPatch[networkName] = network
			}
			if backupDirectory == "" {
				wd, err := os.Getwd()
				wd = path.Join(
					wd,
					cli.RuntimeTimestampShort,
				)
				if err != nil {
					log.Printf(
						"Failed to get current working directory, backups will not be written: %v",
						err,
					)
				}
				backupDirectory = wd
			}

			err = getBSSClient(&bssSession)
			if err != nil {
				log.Fatalf(
					"Failed to communicate with BSS because %v\n",
					err,
				)
			}

			err = getSLSClient(&slsSession)
			if err != nil {
				log.Fatalf(
					"Failed to communicate with SLS because %v\n",
					err,
				)
			}
			newBootParameters, newSLSNetworks, err := patchIPv6()
			if err != nil {
				log.Fatalf(
					"patch subcommand failed because %v\n",
					err,
				)
			}

			if commit {
				err = putSLSNetworks(
					slsSession,
					newSLSNetworks,
				)
				if err != nil {
					log.Fatalf(
						"failed to upload new SLS network data because %v\n",
						err,
					)
				}
				err = putBSSBootparemeters(
					bssSession,
					newBootParameters,
				)
				if err != nil {
					log.Fatalf(
						"failed to upload new BSS bootparameters because %v\n",
						err,
					)
				}
				fmt.Printf(
					"Changes have been made to BSS and SLS.\nBoth the backups and copies of the applied changes were written to %s\n",
					backupDirectory,
				)
			} else {
				fmt.Printf(
					"This is a dry-run, and no changes were made to BSS or SLS.\nProposed changes and backups of the current state were written to %s\n",
					backupDirectory,
				)
				fmt.Println("To commit these changes, use the --commit flag")
			}
			if forceAlert {
				fmt.Printf("WARNING: One or more BSS or SLS entries were skipped because IPv6 values already existed, and the --force flag was not set.\n")
			}

		},
	}

	c.Flags().StringVarP(
		&backupDirectory,
		"backup-dir",
		"b",
		"",
		"The directory to write backup files to (defaults to a timestamped directory within the current working directory).",
	)

	c.Flags().BoolVarP(
		&commit,
		"commit",
		"w",
		false,
		"Write all proposed changes into CSM; commit changes to BSS and SLS.",
	)

	c.Flags().BoolVarP(
		&force,
		"force",
		"f",
		false,
		"Force updating IPv6 reservations, subnet CIDRs, and more.",
	)

	c.Flags().BoolVarP(
		&remove,
		"remove",
		"r",
		false,
		"Remove all IPv6 enablement data from BSS and SLS.",
	)

	c.MarkFlagsMutuallyExclusive(
		"remove",
		"force",
	)

	c.Flags().StringSliceVar(
		&subnetsToPatch,
		"subnets",
		defaultSubnetsToPatch,
		"Comma-separated list of SLS subnets to patch.",
	)
	toPatch = make(NetworkParams)
	for _, network := range networksToPatch {
		cidr := c.Flags().String(
			fmt.Sprintf(
				"%s-cidr6",
				strings.ToLower(network),
			),
			"",
			fmt.Sprintf(
				"Overall IPv6 CIDR for all %s subnets.",
				network,
			),
		)
		gateway := c.Flags().String(
			fmt.Sprintf(
				"%s-gateway6",
				strings.ToLower(network),
			),
			"",
			fmt.Sprintf(
				"IPv6 Gateway for NCNs on the %s.",
				network,
			),
		)
		toPatch[network] = NetworkParam{
			CIDRStr:    cidr,
			GatewayStr: gateway,
		}
	}

	return c
}

func patchIPv6() (bootParamsCollection []bssTypes.BootParams, newSLSNetworks []slsCommon.Network, err error) {
	slsDump, err := getSLSData()
	if err != nil {
		return bootParamsCollection, newSLSNetworks, fmt.Errorf(
			"failed to get sls data because %w",
			err,
		)
	}

	err = writeToFile(
		"sls-dumpstate",
		slsDump,
	)
	if err != nil {
		return bootParamsCollection, newSLSNetworks, fmt.Errorf(
			"failed to backup SLS changes to disk because %v",
			err,
		)
	}

	slsNetworks := slsDump.Networks
	if remove {
		err = removeIPv6SLSNetworks(
			&slsNetworks,
			&newSLSNetworks,
		)
	} else {
		err = addIPv6SLSNetworks(
			&slsNetworks,
			&newSLSNetworks,
		)
	}
	if err != nil {
		return bootParamsCollection, newSLSNetworks, fmt.Errorf(
			"failed to update SLS networks because %v",
			err,
		)
	}
	err = writeToFile(
		"sls-patched",
		&newSLSNetworks,
	)
	if err != nil {
		return bootParamsCollection, newSLSNetworks, fmt.Errorf(
			"failed to write proposed SLS changes to disk because %v",
			err,
		)
	}
	bootParamsCollection, err = addIPv6BSSBootParameters(&newSLSNetworks)
	if err != nil {
		return bootParamsCollection, newSLSNetworks, fmt.Errorf(
			"failed to create new BSS bootparameters from IPv6 SLS data because %v",
			err,
		)
	}
	return bootParamsCollection, newSLSNetworks, err
}

func writeToFile(name string, data interface{}) (err error) {
	destinationPath := path.Join(
		backupDirectory,
		fmt.Sprintf(
			"%s-%s.json",
			name,
			cli.RuntimeTimestampShort,
		),
	)
	_, err = os.Stat(destinationPath)
	if err != nil {
		err := os.MkdirAll(
			backupDirectory,
			0755,
		)
		if err != nil {
			log.Fatalf(
				"failed to create destination for backups beause %v\n",
				err,
			)
		}
	}
	err = files.WriteJSONConfig(
		destinationPath,
		data,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to write because %v",
			err,
		)
	}
	return err
}
