/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

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

package initialize

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Cray-HPE/cray-site-init/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/mod/semver"

	shcdParser "github.com/Cray-HPE/hms-shcd-parser/pkg/shcd-parser"
	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"

	"github.com/Cray-HPE/cray-site-init/internal/files"
	slsInit "github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/csm"
	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
	"github.com/Cray-HPE/cray-site-init/pkg/version"
)

// NewCommand represents the init command
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "init",
		Short: "Generates a Shasta configuration payload",
		Long: `init generates a scaffolding the Shasta configuration payload. It is based on several input files:
	1. The hmn_connections.json which describes the cabling for the BMCs on the NCNs
	2. The ncn_metadata.csv file documents the MAC addresses of the NCNs to be used in this installation
	   NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
	3. The switch_metadata.csv file which documents the Xname, Brand, Type, and Model of each switch. Types are CDU, LeafBMC, Leaf, and Spine
	   Switch Xname,Type,Brand,Model

	** NB **
	For systems that use non-sequential cabinet id numbers, an additional mapping file is necessary and must be indicated
	with the --cabinets-yaml flag.
	** NB **

	** NB **
	For additional control of the application node identification during the SLS Input File generation, an additional config file is necessary
	and must be indicated with the --application-node-config-yaml flag.

	Allows control of the following in the SLS Input File:
	1. System specific prefix for Applications node
	2. Specify HSM Subroles for system-specific application nodes
	3. Specify Application node Aliases
	** NB **

	In addition, there are many flags to impact the layout of the system. The defaults are generally fine except for the networking flags.
	`,
		DisableAutoGenTag: true,
		Run: func(c *cobra.Command, args []string) {
			// Initialize the global viper
			v := viper.GetViper()

			err := v.BindPFlags(c.Flags())
			if err != nil {
				log.Fatalln(err)
			}
			flagErrors := validateFlags()
			if len(flagErrors) > 0 {
				_ = c.Usage()
				for _, e := range flagErrors {
					log.Println(e)
				}
				log.Fatal("One or more flags are invalid")
			}

			if len(
				strings.Split(
					v.GetString("site-ip"),
					"/",
				),
			) != 2 {
				_ = c.Usage()
				log.Fatalf(
					"FATAL ERROR: Unable to parse %s as --site-ip. Must be in the format \"192.168.0.1/24\"",
					v.GetString("site-ip"),
				)

			}

			// Validate that the BGP ASN is within the private range
			var bgp = map[string]int{
				"bgp-asn":     v.GetInt("bgp-asn"),
				"bgp-chn-asn": v.GetInt("bgp-chn-asn"),
				"bgp-cmn-asn": v.GetInt("bgp-cmn-asn"),
				"bgp-nmn-asn": v.GetInt("bgp-nmn-asn"),
			}

			for network, asn := range bgp {
				if asn > 65534 || asn < 64512 {
					_ = c.Usage()
					log.Fatalln(
						"FATAL ERROR: BGP ASNs must be within the private range 64512-65534, fix the value for:",
						network,
					)
				}
			}

			// Read and validate our three input files
			hmnRows, logicalNCNs, switches, applicationNodeConfig, cabinetDetailList, errs := collectInput(v)
			if errs != nil {
				log.Fatalf(
					"FATAL ERROR: Failed to collect one or more input files: %v",
					errs,
				)
			}

			defaultNetConfigs := GenerateDefaultNetworkConfigs(
				switches,
				logicalNCNs,
				cabinetDetailList,
			)
			internalNetConfigs, err := GenerateNetworkConfigs(defaultNetConfigs)
			if err != nil {
				log.Fatal(err)
			}
			// Build a set of networks we can use
			shastaNetworks, err := slsInit.BuildCSMNetworks(
				internalNetConfigs,
				cabinetDetailList,
				switches,
			)
			if err != nil {
				log.Fatal(err)
			}

			// Use our new networks and our list of logicalNCNs to distribute ips
			AllocateIPs(
				logicalNCNs,
				shastaNetworks,
			)

			// Now we can finally generate the slsState
			slsState := prepareAndGenerateSLS(
				cabinetDetailList,
				shastaNetworks,
				hmnRows,
				switches,
				applicationNodeConfig,
				v.GetInt("starting-mountain-nid"),
			)
			// SLS can tell us which NCNs match with which Xnames, we need to update the IP Reservations
			slsNcns, err := ExtractSLSNCNs(&slsState)
			if err != nil {
				log.Panic(err) // This should never happen. I can't really imagine how it would.
			}

			// Merge the SLS NCN list with the NCN list we got at the beginning
			err = mergeNCNs(
				logicalNCNs,
				slsNcns,
			)
			if err != nil {
				log.Fatalln(err)
			}

			// Pull UANs from the completed slsState to assign CAN addresses
			slsUans, err := ExtractUANs(&slsState)
			if err != nil {
				log.Panic(err) // This should never happen. I can't really imagine how it would.
			}

			// Only add UANs if there actually is a CAN network
			if v.GetString("bican-user-network-name") == "CAN" || v.GetBool("retain-unused-user-network") {
				canSubnet, _ := shastaNetworks["CAN"].LookUpSubnet("bootstrap_dhcp")
				for _, uan := range slsUans {
					_, err := networking.AddReservation(
						canSubnet,
						uan.Hostname,
						uan.Xname,
					)
					if err != nil {
						return
					}
				}
			}
			// Only add UANs if there actually is a CHN network
			if v.GetString("bican-user-network-name") == "CHN" || v.GetBool("retain-unused-user-network") {
				chnSubnet, _ := shastaNetworks["CHN"].LookUpSubnet("bootstrap_dhcp")
				for _, uan := range slsUans {
					_, err := networking.AddReservation(
						chnSubnet,
						uan.Hostname,
						uan.Xname,
					)
					if err != nil {
						return
					}
				}
			}

			// Cycle through the main networks and update the reservations, masks and dhcp ranges as necessary
			for _, network := range networking.ValidNetNames {
				if shastaNetworks[network] != nil {
					// Grab the supernet details for use in HACK substitution
					if err == nil {
						// Loop the reservations and update the NCN reservations with hostnames
						// we likely didn't have when we registered the reservation
						subnet, err := updateReservations(
							shastaNetworks[network],
							"bootstrap_dhcp",
							logicalNCNs,
						)
						if err != nil {
							continue
						}
						if network == "CAN" || network == "CMN" || network == "CHN" {
							netNameLower := strings.ToLower(network)

							// Do not use supernet hack for the CAN/CMN/CHN, these should reflect the abstracted subnets.
							err := networking.UpdateDHCPRange(
								subnet,
								false,
							)
							if err != nil {
								log.Fatalf(
									"Error updating DHCP range: %v\n",
									err,
								)
							}

							cidr4Key := fmt.Sprintf(
								"%s-cidr4",
								netNameLower,
							)
							cidrKey := fmt.Sprintf(
								"%s-cidr",
								netNameLower,
							)

							// Handle IPv4 CIDRs, networks with IPv6 will use a different key for their cidr4.
							var cidr4 string
							if v.IsSet(cidr4Key) {
								cidr4 = v.GetString(cidr4Key)
							} else if v.IsSet(cidrKey) {
								cidr4 = v.GetString(cidrKey)
							}

							if cidr4 == "" {
								continue
							}
							myPrefix, err := netip.ParsePrefix(cidr4)
							if err != nil {
								log.Fatalf(
									"Unable to parse CIDR '%s': %v",
									cidr4,
									err,
								)
							}

							// If neither static nor dynamic pool is defined we can use the last available IP in the subnet
							poolStartIP, err := networking.Broadcast(myPrefix)
							if err != nil {
								log.Fatal(err)
							}

							// Do not overlap the static or dynamic pools
							myStaticPoolName := fmt.Sprintf(
								"%s-static-pool",
								netNameLower,
							)
							myDynPoolName := fmt.Sprintf(
								"%s-dynamic-pool",
								netNameLower,
							)

							myStaticPoolCIDR := v.GetString(myStaticPoolName)
							myDynPoolCIDR := v.GetString(myDynPoolName)

							if len(myStaticPoolCIDR) > 0 && len(myDynPoolCIDR) > 0 {
								// Both pools are defined so find the start of whichever pool comes first
								_, myStaticPool, _ := net.ParseCIDR(myStaticPoolCIDR)
								_, myDynamicPool, _ := net.ParseCIDR(myDynPoolCIDR)
								myStaticPoolPrefix, parseErr := netip.ParsePrefix(myStaticPoolCIDR)
								myDynamicPoolPrefix, parseDynamicErr := netip.ParsePrefix(myDynPoolCIDR)
								if parseErr != nil || parseDynamicErr != nil {
									log.Fatalln(
										parseErr,
										parseDynamicErr,
									)
								}
								if myStaticPoolPrefix.Addr().Compare(myDynamicPoolPrefix.Addr()) == -1 {
									poolStartIP, err = netip.ParseAddr(myStaticPool.IP.String())
								} else {
									poolStartIP, err = netip.ParseAddr(myDynamicPool.IP.String())
								}
							} else if len(myStaticPoolCIDR) > 0 && len(myDynPoolCIDR) == 0 {
								// Only the static pool is defined so use the first IP of that pool
								_, myStaticPool, _ := net.ParseCIDR(myStaticPoolCIDR)
								poolStartIP, err = netip.ParseAddr(myStaticPool.IP.String())
							} else if len(myStaticPoolCIDR) == 0 && len(myDynPoolCIDR) > 0 {
								// Only the dynamic pool is defined so use the first IP of that pool
								_, myDynamicPool, _ := net.ParseCIDR(myDynPoolCIDR)
								poolStartIP, err = netip.ParseAddr(myDynamicPool.IP.String())
							}
							if err != nil {
								log.Fatalf(
									"Failed to parse a static or dynamic pool because %v\n",
									err,
								)
							}

							// Guidance has changed on whether the CAN gw should be at the start or end of the
							// range. Here we account for it being at the end of the range.
							// Leaving this check in place for CMN because it is harmless to do so.
							subnetGateway, err := netip.ParseAddr(subnet.Gateway.String())
							if err != nil {
								log.Fatalf(
									"Failed to parse subnet gateway because %v\n",
									err,
								)
							}
							if subnetGateway == poolStartIP.Prev() {
								// The gw *is* at the end, so shorten the range to accommodate
								subnet.DHCPEnd = poolStartIP.Prev().Prev().AsSlice()
							} else {
								// The gw is not at the end
								subnet.DHCPEnd = poolStartIP.Prev().AsSlice()
							}
						} else {
							err := networking.UpdateDHCPRange(
								subnet,
								v.GetBool("supernet"),
							)
							if err != nil {
								log.Fatalf(
									"Error updating DHCP range: %v\n",
									err,
								)
							}
						}
					}

					// We expect a bootstrap_dhcp in every net, but uai_macvlan is only in
					// the NMN range for today
					if strings.ToUpper(network) == "NMN" {
						subnet, err := updateReservations(
							shastaNetworks[network],
							"uai_macvlan",
							logicalNCNs,
						)
						if err != nil {
							continue
						}
						err = networking.UpdateDHCPRange(
							subnet,
							false,
						)
						if err != nil {
							log.Fatalf(
								"Error updating DHCP range: %v\n",
								subnet.Name,
							)
						}
					}

				}
			}

			// Update the SLSState with the updated network information
			_, slsState.Networks = prepareNetworkSLS(shastaNetworks)

			// Switch from a list of pointers to a list of things before we write it out
			var ncns []LogicalNCN
			for _, ncn := range logicalNCNs {
				ncns = append(
					ncns,
					*ncn,
				)
			}
			globalMetaData, err := MakeBasecampGlobalMetaData(
				v,
				ncns,
				shastaNetworks,
				"NMN",
				"bootstrap_dhcp",
				v.GetString("install-ncn"),
			)
			if err != nil {
				log.Fatalln(
					"unable to generate basecamp globalMetaData: ",
					err,
				)
			}
			err = writeOutput(
				v,
				shastaNetworks,
				slsState,
				ncns,
				switches,
				globalMetaData,
			)
			if err != nil {
				log.Fatalf(
					"unable to write one or more output files: %v",
					err,
				)
			}

			// Gather SLS information for summary
			slsMountainCabinets := GetSLSCabinets(
				slsState,
				slsCommon.ClassMountain,
			)
			slsHillCabinets := GetSLSCabinets(
				slsState,
				slsCommon.ClassHill,
			)
			slsRiverCabinets := GetSLSCabinets(
				slsState,
				slsCommon.ClassRiver,
			)

			if v.IsSet("cabinets-yaml") && (v.GetString("cabinets-yaml") != "") && (v.IsSet("mountain-cabinets") ||
				v.IsSet("starting-mountain-cabinet") ||
				v.IsSet("river-cabinets") ||
				v.IsSet("starting-river-cabinet") ||
				v.IsSet("hill-cabinets") ||
				v.IsSet("starting-hill-cabinet")) {
				fmt.Printf("\nWARNING: cabinet flags are not honored when a cabinets-yaml file is provided\n")
			}

			// Print Summary
			fmt.Printf(
				"\n===== %s Installation Summary =====\n\n",
				v.GetString("system-name"),
			)
			fmt.Printf(
				"%-20s: %s\n",
				"Installation Node",
				v.GetString("install-ncn"),
			)

			fmt.Printf(
				"%-20s: %s\n",
				"Upstream DNS",
				v.GetString("site-dns"),
			)
			fmt.Printf(
				"%-20s: %v\n",
				"MetalLB Peers",
				v.GetStringSlice("bgp-peer-types"),
			)
			fmt.Printf(
				"\n----- %s Network Summary -----\n\n",
				v.GetString("system-name"),
			)
			fmt.Printf(
				"%-20s: %s\n",
				"BICAN user network",
				v.GetString("bican-user-network-name"),
			)
			fmt.Printf(
				"%-20s: ",
				"Supernet",
			)
			if v.GetBool("supernet") {
				fmt.Printf("Enabled (Network gateway used for all SLS subnets)\n")
			} else {
				fmt.Printf("Disabled (SLS subnets use their own gateway)\n")
			}

			fmt.Printf("\nDefined networks:\n\n")
			sortedNetworks := make(
				networking.IPNetworks,
				0,
				len(shastaNetworks),
			)
			for shastaNetwork := range shastaNetworks {
				sortedNetworks = append(
					sortedNetworks,
					shastaNetworks[shastaNetwork],
				)
			}
			sort.Sort(sortedNetworks)
			for _, network := range sortedNetworks {
				prefix4, err := netip.ParsePrefix(network.CIDR4)
				if err != nil || prefix4.Addr().IsUnspecified() {
					continue
				}
				fmt.Printf(
					"    * %-65s (subnets: %3d) %s\n",
					network.FullName,
					len(network.AllocatedIPv4Subnets()),
					network.CIDR4,
				)
				if network.CIDR6 != "" {
					prefix6, err := netip.ParsePrefix(network.CIDR4)
					if err != nil || prefix6.Addr().IsUnspecified() {
						continue
					}
					fmt.Printf(
						"    * %-65s (subnets: %3d) %s\n",
						fmt.Sprintf(
							"%s (IPv6)",
							network.FullName,
						),
						len(network.AllocatedIPv6Subnets()),
						network.CIDR6,
					)
				}
			}
			fmt.Printf(
				"\n----- %s System Summary -----\n\n",
				v.GetString("system-name"),
			)
			fmt.Printf(
				"%-30s: %-3d\n",
				"NCNs",
				len(ncns),
			)
			fmt.Printf(
				"%-30s: %-3d\n",
				"UANs",
				len(slsUans),
			)
			fmt.Printf(
				"%-30s: %-3d\n",
				"Switches",
				len(switches),
			)
			fmt.Printf(
				"%-30s: %-3d\n",
				"Mountain Compute Cabinets",
				len(slsMountainCabinets),
			)
			fmt.Printf(
				"%-30s: %-3d\n",
				"Hill Compute Cabinets",
				len(slsHillCabinets),
			)
			fmt.Printf(
				"%-30s: %-3d\n",
				"River Compute Cabinets",
				len(slsRiverCabinets),
			)
			fmt.Printf(
				"%-30s: %s\n",
				"CSI Version Information",
				version.Get(),
			)
			fmt.Printf(
				"\n%s\n********** CONFIG INITIALIZED **********\n%s\n",
				strings.Repeat(
					"*",
					40,
				),
				strings.Repeat(
					"*",
					40,
				),
			)
			fmt.Printf(
				"\nNEW %s file!\n\nReplace any offline copies\nof %s with:\n\n%s\n\n",
				cli.ConfigFilename,
				cli.ConfigFilename,
				path.Join(
					"./",
					v.GetString("system-name"),
					cli.ConfigFilename,
				),
			)
			fmt.Println(
				strings.Repeat(
					"*",
					40,
				),
			)
			fmt.Println(
				strings.Repeat(
					"*",
					40,
				),
			)
		},
	}
	var flagErr error

	// System Configuration Flags based on previous system_config.yml and networks_derived.yml
	c.Flags().String(
		csm.APIKeyName,
		"",
		"Version of CSM being installed (e.g. <major>.<minor> such as \"1.5\" or \"v1.5\").",
	)
	_ = c.MarkFlagRequired(csm.APIKeyName)

	// Installer specific.
	c.Flags().String(
		"system-name",
		"sn-2024",
		"Name of the System",
	)
	c.Flags().String(
		"site-domain",
		"",
		"Site Domain Name",
	)
	c.Flags().String(
		"first-master-hostname",
		"ncn-m002",
		"Hostname of the first master node",
	)
	c.Flags().String(
		"install-ncn",
		"ncn-m001",
		"Hostname of the node to be used for installation",
	)
	c.Flags().String(
		"install-ncn-bond-members",
		"p1p1,p1p2",
		"List of devices to use to form a bond on the install ncn",
	)
	_ = c.MarkFlagRequired("system-name")

	// NTP
	c.Flags().StringSlice(
		"ntp-pools",
		[]string{""},
		"Comma-separated list of upstream NTP pool(s)",
	)
	c.Flags().StringSlice(
		"ntp-servers",
		[]string{"ncn-m001"},
		"Comma-separated list of upstream NTP server(s); ncn-m001 should always be in this list",
	)
	c.Flags().StringSlice(
		"ntp-peers",
		[]string{
			"ncn-m001",
			"ncn-m002",
			"ncn-m003",
			"ncn-w001",
			"ncn-w002",
			"ncn-w003",
			"ncn-s001",
			"ncn-s002",
			"ncn-s003",
		},
		"Comma-separated list of NCNs that will peer together",
	)
	c.Flags().String(
		"ntp-timezone",
		"UTC",
		"Timezone to be used on the NCNs and across the system",
	)

	// Site networking for the installer environment.
	c.Flags().String(
		"site-dns",
		"",
		"Site Network DNS Server",
	)
	c.Flags().String(
		"site-gw",
		"",
		"Site Network IPv4 Gateway4",
	)
	c.Flags().String(
		"site-ip",
		"",
		"Site Network Information in the form ipaddress/prefix like 192.168.1.1/24",
	)
	c.Flags().String(
		"site-nic",
		"em1",
		"Network Interface on install-ncn that will be connected to the site network",
	)
	_ = c.MarkFlagRequired("site-ip")
	c.MarkFlagsRequiredTogether(
		"site-ip",
		"site-gw",
		"site-dns",
		"site-nic",
	)

	// BICAN Network Toggle
	c.Flags().String(
		"bican-user-network-name",
		"",
		"Name of the network over which non-admin users access the system [CAN, CHN, HSN]",
	)
	c.Flags().Bool(
		"retain-unused-user-network",
		false,
		"Use the supernet mask and gateway for NCNs and Switches",
	)

	// Node management network.
	c.Flags().String(
		"nmn-cidr",
		networking.DefaultNMNString,
		"Overall IPv4 CIDR for all Node Management subnets",
	)
	c.Flags().String(
		"nmn-static-pool",
		"",
		"Overall IPv4 CIDR for static Node Management load balancer addresses",
	)
	c.Flags().String(
		"nmn-dynamic-pool",
		networking.DefaultNMNLBString,
		"Overall IPv4 CIDR for dynamic Node Management load balancer addresses",
	)
	c.Flags().String(
		"nmn-mtn-cidr",
		networking.DefaultNMNMTNString,
		"IPv4 CIDR for grouped Mountain Node Management subnets",
	)
	c.Flags().String(
		"nmn-rvr-cidr",
		networking.DefaultNMNRVRString,
		"IPv4 CIDR for grouped River Node Management subnets",
	)
	_ = c.MarkFlagRequired("nmn-cidr")

	// Hardware management network.
	c.Flags().String(
		"hmn-cidr",
		networking.DefaultHMNString,
		"Overall IPv4 CIDR for all Hardware Management subnets",
	)
	c.Flags().String(
		"hmn-static-pool",
		"",
		"Overall IPv4 CIDR for static Hardware Management load balancer addresses",
	)
	c.Flags().String(
		"hmn-dynamic-pool",
		networking.DefaultHMNLBString,
		"Overall IPv4 CIDR for dynamic Hardware Management load balancer addresses",
	)
	c.Flags().String(
		"hmn-mtn-cidr",
		networking.DefaultHMNMTNString,
		"IPv4 CIDR for grouped Mountain Hardware Management subnets",
	)
	c.Flags().String(
		"hmn-rvr-cidr",
		networking.DefaultHMNRVRString,
		"IPv4 CIDR for grouped River Hardware Management subnets",
	)
	_ = c.MarkFlagRequired("hmn-cidr")

	// Customer access network.
	c.Flags().String(
		"can-cidr",
		"",
		"Overall IPv4 CIDR for all Customer Access subnets",
	)
	c.Flags().String(
		"can-gateway",
		"",
		"Gateway for NCNs on the CAN (User)",
	)
	c.Flags().String(
		"can-static-pool",
		"",
		"Overall IPv4 CIDR for static Customer Access load balancer addresses",
	)
	c.Flags().String(
		"can-dynamic-pool",
		"",
		"Overall IPv4 CIDR for dynamic Customer Access load balancer addresses",
	)
	c.MarkFlagsRequiredTogether(
		"can-cidr",
		"can-gateway",
		"can-static-pool",
		"can-dynamic-pool",
	)

	// Customer high-speed network.
	c.Flags().String(
		"chn-cidr",
		"",
		"Overall IPv4 CIDR for all Customer High-Speed subnets",
	)
	c.Flags().String(
		"chn-gateway",
		"",
		"Gateway for NCNs on the CHN (User)",
	)
	flagErr = c.Flags().MarkDeprecated(
		"chn-cidr",
		"Please use --chn-cidr4 instead",
	)
	if flagErr != nil {
		log.Fatalf(
			"Failed to mark flag as deprecated because %v",
			flagErr,
		)
		return nil
	}

	flagErr = c.Flags().MarkDeprecated(
		"chn-gateway",
		"Please use --chn-gateway4 instead",
	)
	if flagErr != nil {
		log.Fatalf(
			"Failed to mark flag as deprecated because %v",
			flagErr,
		)
		return nil
	}
	c.Flags().String(
		"chn-cidr4",
		"",
		"Overall IPv4 CIDR for all Customer High-Speed subnets",
	)
	c.Flags().String(
		"chn-gateway4",
		"",
		"IPv4 Gateway for NCNs on the CHN (User)",
	)
	c.Flags().String(
		"chn-cidr6",
		"",
		"Overall IPv6 CIDR for all Customer High-Speed subnets",
	)
	c.Flags().String(
		"chn-gateway6",
		"",
		"IPv6 Gateway for NCNs on the CHN (User)",
	)
	c.Flags().String(
		"chn-static-pool",
		"",
		"Overall IPv4 CIDR for static Customer High-Speed load balancer addresses",
	)
	c.Flags().String(
		"chn-dynamic-pool",
		"",
		"Overall IPv4 CIDR for dynamic Customer High-Speed load balancer addresses",
	)

	c.MarkFlagsRequiredTogether(
		"chn-cidr",
		"chn-gateway",
	)
	c.MarkFlagsRequiredTogether(
		"chn-cidr4",
		"chn-gateway4",
	)
	c.MarkFlagsRequiredTogether(
		"chn-cidr6",
		"chn-gateway6",
	)
	RegisterAlias(
		"chn-cidr",
		"chn-cidr4",
	)
	RegisterAlias(
		"chn-gateway",
		"chn-gateway4",
	)
	c.MarkFlagsMutuallyExclusive(
		"chn-gateway",
		"chn-gateway4",
	)
	c.MarkFlagsMutuallyExclusive(
		"chn-cidr",
		"chn-cidr4",
	)
	RegisterAlias(
		"chn-cidr",
		"chn-cidr4",
	)
	RegisterAlias(
		"chn-gateway",
		"chn-gateway4",
	)
	c.MarkFlagsRequiredTogether(
		"chn-static-pool",
		"chn-dynamic-pool",
	)
	// Customer management network.
	c.Flags().String(
		"cmn-cidr",
		"",
		"Overall IPv4 CIDR for all Customer Management subnets",
	)
	flagErr = c.Flags().MarkDeprecated(
		"cmn-cidr",
		"Please use --cmn-cidr4 instead",
	)
	if flagErr != nil {
		log.Fatalf(
			"Failed to mark flag as deprecated because %v",
			flagErr,
		)
		return nil
	}
	c.Flags().String(
		"cmn-gateway",
		"",
		"Gateway for NCNs on the CMN (Administrative/Management)",
	)
	flagErr = c.Flags().MarkDeprecated(
		"cmn-gateway",
		"Please use --cmn-gateway4 instead",
	)
	if flagErr != nil {
		log.Fatalf(
			"Failed to mark flag as deprecated because %v",
			flagErr,
		)
		return nil
	}
	c.Flags().String(
		"cmn-cidr4",
		"",
		"Overall IPv4 CIDR for all Customer Management subnets",
	)
	c.Flags().String(
		"cmn-gateway4",
		"",
		"IPv4 Gateway for NCNs on the CMN (Administrative/Management)",
	)
	c.Flags().String(
		"cmn-cidr6",
		"",
		"Overall IPv6 CIDR for all Customer Management subnets",
	)
	c.Flags().String(
		"cmn-gateway6",
		"",
		"IPv6 Gateway for NCNs on the CMN (Administrative/Management)",
	)
	c.Flags().String(
		"cmn-static-pool",
		"",
		"Overall IPv4 CIDR for static Customer Management load balancer addresses",
	)
	c.Flags().String(
		"cmn-dynamic-pool",
		"",
		"Overall IPv4 CIDR for dynamic Customer Management load balancer addresses",
	)
	c.Flags().String(
		"cmn-external-dns",
		"",
		"IP Address in the cmn-static-pool for the external dns service \"site-to-system lookups\"",
	)
	c.MarkFlagsRequiredTogether(
		"cmn-cidr",
		"cmn-gateway",
	)
	c.MarkFlagsRequiredTogether(
		"cmn-cidr4",
		"cmn-gateway4",
	)
	c.MarkFlagsRequiredTogether(
		"cmn-cidr6",
		"cmn-gateway6",
	)

	RegisterAlias(
		"cmn-cidr",
		"cmn-cidr4",
	)
	RegisterAlias(
		"cmn-gateway",
		"cmn-gateway4",
	)
	c.MarkFlagsMutuallyExclusive(
		"cmn-cidr",
		"cmn-cidr4",
	)
	c.MarkFlagsMutuallyExclusive(
		"cmn-gateway",
		"cmn-gateway4",
	)

	c.MarkFlagsRequiredTogether(
		"cmn-static-pool",
		"cmn-dynamic-pool",
		"cmn-external-dns",
	)

	// Metal network.
	c.Flags().String(
		"mtl-cidr",
		networking.DefaultMTLString,
		"Overall IPv4 CIDR for all Provisioning subnets",
	)

	// High-speed network.
	c.Flags().String(
		"hsn-cidr",
		networking.DefaultHSNString,
		"Overall IPv4 CIDR for all HSN subnets",
	)

	// Misc network.
	c.Flags().Bool(
		"supernet",
		true,
		"Use the supernet mask and gateway for NCNs and Switches",
	)
	c.Flags().Int(
		"management-net-ips",
		0,
		"Additional number of IP addresses to reserve in each vlan for network equipment",
	)

	// Bootstrap VLANS
	c.Flags().Int(
		"can-bootstrap-vlan",
		networking.DefaultCANVlan,
		"Bootstrap VLAN for the CAN",
	)
	c.Flags().Int(
		"cmn-bootstrap-vlan",
		networking.DefaultCMNVlan,
		"Bootstrap VLAN for the CMN",
	)
	c.Flags().Int(
		"hmn-bootstrap-vlan",
		networking.DefaultHMNVlan,
		"Bootstrap VLAN for the HMN",
	)
	c.Flags().Int(
		"nmn-bootstrap-vlan",
		networking.DefaultNMNVlan,
		"Bootstrap VLAN for the NMN",
	)

	// Hardware Details
	c.Flags().Int(
		"mountain-cabinets",
		4,
		"Number of Mountain Cabinets",
	) // 4 mountain cabinets per CDU
	c.Flags().Int(
		"starting-mountain-cabinet",
		1000,
		"Starting ID number for Mountain Cabinets",
	)

	c.Flags().Int(
		"river-cabinets",
		1,
		"Number of River Cabinets",
	)
	c.Flags().Int(
		"starting-river-cabinet",
		3000,
		"Starting ID number for River Cabinets",
	)

	c.Flags().Int(
		"hill-cabinets",
		0,
		"Number of Hill Cabinets",
	)
	c.Flags().Int(
		"starting-hill-cabinet",
		9000,
		"Starting ID number for Hill Cabinets",
	)

	c.Flags().Int(
		"starting-river-NID",
		1,
		"Starting NID for Compute Nodes",
	)
	c.Flags().Int(
		"starting-mountain-NID",
		1000,
		"Starting NID for Compute Nodes",
	)

	// BGP
	c.Flags().String(
		"bgp-asn",
		"65533",
		"The autonomous system number for BGP router",
	)
	c.Flags().String(
		"bgp-cmn-asn",
		"65532",
		"The autonomous system number for CMN BGP clients",
	)
	c.Flags().String(
		"bgp-nmn-asn",
		"65531",
		"The autonomous system number for NMN BGP clients",
	)
	c.Flags().String(
		"bgp-chn-asn",
		"65530",
		"The autonomous system number for CHN BGP clients",
	)
	c.Flags().String(
		"bgp-peers",
		"spine",
		"Which set of switches to use as metallb peers, spine (default) or leaf",
	)
	c.Flags().MarkDeprecated(
		"bgp-peers",
		"Use --bgp-peer-types.",
	)
	if flagErr != nil {
		log.Fatalf(
			"Failed to mark flag as deprecated because %v",
			flagErr,
		)
		return nil
	}
	c.Flags().StringSlice(
		"bgp-peer-types",
		[]string{"spine"},
		"Comma-separated list of which set of switches to use as metallb peers: spine (default), leaf and/or edge",
	)

	// Kubernetes.
	c.Flags().Bool(
		"k8s-api-auditing-enabled",
		false,
		"Enable the kubernetes auditing API",
	)
	c.Flags().Bool(
		"ncn-mgmt-node-auditing-enabled",
		false,
		"Enable management node auditing",
	)

	// Hardware controllers.
	c.Flags().String(
		"bootstrap-ncn-bmc-pass",
		"",
		"Password for connecting to the BMC on the initial NCNs",
	)
	c.Flags().String(
		"bootstrap-ncn-bmc-user",
		"",
		"Username for connecting to the BMC on the initial NCNs",
	)
	flagErr = c.MarkFlagRequired("bootstrap-ncn-bmc-pass")
	if flagErr != nil {
		log.Fatalf(
			"Failed to mark flag as required because %v",
			flagErr,
		)
		return nil
	}
	flagErr = c.MarkFlagRequired("bootstrap-ncn-bmc-user")
	if flagErr != nil {
		log.Fatalf(
			"Failed to mark flag as required because %v",
			flagErr,
		)
		return nil
	}
	c.MarkFlagsRequiredTogether(
		"bootstrap-ncn-bmc-pass",
		"bootstrap-ncn-bmc-user",
	)

	// Seed files.
	c.Flags().String(
		"application-node-config-yaml",
		"",
		"YAML to control Application node identification during the SLS Input File generation",
	)
	c.Flags().String(
		"cabinets-yaml",
		"",
		"YAML file listing the ids for all cabinets by type",
	)
	c.Flags().String(
		"hmn-connections",
		"hmn_connections.json",
		"HMN Connections JSON Location (For generating an SLS File)",
	)
	c.Flags().String(
		"ncn-metadata",
		"ncn_metadata.csv",
		"CSV for mapping the mac addresses of the NCNs to their xnames",
	)
	c.Flags().String(
		"switch-metadata",
		"switch_metadata.csv",
		"CSV for mapping the switch xname, brand, type, and model",
	)

	// DNS zone transfer settings
	c.Flags().String(
		"primary-server-name",
		"primary",
		"Desired name for the primary DNS server",
	)
	c.Flags().String(
		"secondary-servers",
		"",
		"Comma-separated list of FQDN/IP for all DNS servers to notify when zone changes are made",
	)
	c.Flags().String(
		"notify-zones",
		"",
		"Comma-separated list of the zones to be allowed transfer",
	)
	c.AddCommand(emptyCommand())

	return c
}

func setupDirectories(systemName string, v *viper.Viper) (string, error) {
	// Set up the path for our base directory using our system name.
	basepath, err := filepath.Abs(filepath.Clean(systemName))
	if err != nil {
		return basepath, err
	}
	// Create our base directory
	if err = os.Mkdir(
		basepath,
		0777,
	); err != nil {
		return basepath, err
	}

	// These Directories make up the overall structure for the Configuration Payload
	// TODO: Refactor this out of the function and into defaults or some other config
	dirs := []string{
		filepath.Join(
			basepath,
			"manufacturing",
		),
		filepath.Join(
			basepath,
			"dnsmasq.d",
		),
		filepath.Join(
			basepath,
			"pit-files",
		),
		filepath.Join(
			basepath,
			"basecamp",
		),
	}

	// Iterate through the directories and create them
	for _, dir := range dirs {
		if err := os.Mkdir(
			dir,
			0777,
		); err != nil {
			// log.Fatalln("Can't create directory", dir, err)
			return basepath, err
		}
	}
	return basepath, nil
}

func mergeNCNs(logicalNcns []*LogicalNCN, slsNCNs []LogicalNCN) error {
	// Merge the SLS NCN list with the NCN list from ncn-metadata
	for _, ncn := range logicalNcns {
		found := false
		for _, slsNCN := range slsNCNs {
			if ncn.Xname == slsNCN.Xname {
				// log.Printf("Found match for %v: %v \n", ncn.Xname, tempNCN)
				ncn.Hostname = slsNCN.Hostname
				ncn.Aliases = slsNCN.Aliases
				ncn.BmcPort = slsNCN.BmcPort

				found = true
				break
			}
		}

		// All NCNs from ncn-metadata need to appear in the generated SLS state
		if !found {
			return fmt.Errorf(
				"failed to find NCN from ncn-metadata in generated SLS State: %s",
				ncn.Xname,
			)
		}
	}

	return nil
}

func prepareNetworkSLS(shastaNetworks map[string]*networking.IPNetwork) (
	[]networking.IPNetwork, map[string]slsCommon.Network,
) {
	// Fix up the network names & create the SLS friendly version of the shasta networks
	var networks []networking.IPNetwork
	for name, network := range shastaNetworks {
		if network.Name == "" {
			network.Name = name
		}
		networks = append(
			networks,
			*network,
		)
	}
	return networks, slsInit.ConvertIPNetworksToSLS(&networks)
}

func prepareAndGenerateSLS(
	cd []sls.CabinetGroupDetail,
	shastaNetworks map[string]*networking.IPNetwork,
	hmnRows []shcdParser.HMNRow,
	inputSwitches []*networking.ManagementSwitch,
	applicationNodeConfig slsInit.GeneratorApplicationNodeConfig,
	startingNid int,
) slsCommon.SLSState {
	// Management Switch Information is included in the IP Reservations for each subnet
	switchNet, err := shastaNetworks["HMN"].LookUpSubnet("network_hardware")
	if err != nil {
		log.Fatalln(
			"Couldn't find subnet for management switches in the HMN:",
			err,
		)
	}
	reservedSwitches, _ := slsInit.ExtractSwitchesfromReservations(switchNet)
	slsSwitches := make(map[string]slsCommon.GenericHardware)
	for _, mySwitch := range reservedSwitches {
		xname := mySwitch.Xname

		// Extract Switch brand from Data stored in switch_metadata.csv
		for _, inputSwitch := range inputSwitches {
			if inputSwitch.Xname == xname {
				mySwitch.Brand = inputSwitch.Brand
				break
			}
		}
		if mySwitch.Brand == "" {
			log.Fatalln(
				"Couldn't determine switch brand for:",
				xname,
			)
		}

		// Create SLS version of the switch
		slsSwitches[mySwitch.Xname], err = slsInit.ConvertManagementSwitchToSLS(&mySwitch)
		if err != nil {
			log.Fatalln(
				"Couldn't get SLS management switch representation:",
				err,
			)
		}
	}

	// Iterate through the cabinets of each kind and build structures that work for SLS Generation
	slsCabinetMap := slsInit.GenCabinetMap(
		cd,
		shastaNetworks,
	)

	// Convert shastaNetwork information to SLS Style Networking
	_, slsNetworks := prepareNetworkSLS(shastaNetworks)

	log.Printf("SLS Cabinet Map\n")
	for class, cabinets := range slsCabinetMap {
		log.Printf(
			"\t Class %v",
			class,
		)
		for xname, cabinet := range cabinets {
			if cabinet.Model == "" {
				log.Printf(
					"\t\t%v\n",
					xname,
				)
			} else {
				log.Printf(
					"\t\t%v - Model %v\n",
					xname,
					cabinet.Model,
				)
			}
		}
	}

	inputState := slsInit.GeneratorInputState{
		ApplicationNodeConfig: applicationNodeConfig,
		ManagementSwitches:    slsSwitches,
		RiverCabinets:         slsCabinetMap[slsCommon.ClassRiver],
		HillCabinets:          slsCabinetMap[slsCommon.ClassHill],
		MountainCabinets:      slsCabinetMap[slsCommon.ClassMountain],
		MountainStartingNid:   startingNid,
		Networks:              slsNetworks,
	}

	slsState := slsInit.GenerateSLSState(
		inputState,
		hmnRows,
	)
	return slsState
}

func updateReservations(network *networking.IPNetwork, subnetName string, logicalNcns []*LogicalNCN) (subnet *slsCommon.IPSubnet, err error) {
	subnet, err = network.LookUpSubnet(subnetName)
	if err != nil {
		return subnet, err
	}
	// Loop the reservations and update the NCN reservations with hostnames
	// we likely didn't have when we registered the reservation
	for index, reservation := range subnet.IPReservations {
		for _, ncn := range logicalNcns {
			if reservation.Comment == ncn.Xname {
				reservation.Name = ncn.Hostname
				reservation.Aliases = append(
					reservation.Aliases,
					fmt.Sprintf(
						"%v-%v",
						ncn.Hostname,
						strings.ToLower(network.Name),
					),
				)
				reservation.Aliases = append(
					reservation.Aliases,
					fmt.Sprintf(
						"time-%v",
						strings.ToLower(network.Name),
					),
				)
				reservation.Aliases = append(
					reservation.Aliases,
					fmt.Sprintf(
						"time-%v.local",
						strings.ToLower(network.Name),
					),
				)
				if strings.ToLower(ncn.Subrole) == "storage" && strings.ToLower(network.Name) == "hmn" {
					reservation.Aliases = append(
						reservation.Aliases,
						"rgw-vip.hmn",
					)
				}
				if strings.ToLower(network.Name) == "nmn" {
					// The xname of a NCN will point to its NMN IP address
					reservation.Aliases = append(
						reservation.Aliases,
						ncn.Xname,
					)
				}
				subnet.IPReservations[index] = reservation
			}
			if reservation.Comment == fmt.Sprintf(
				"%v-mgmt",
				ncn.Xname,
			) {
				reservation.Comment = reservation.Name
				reservation.Aliases = append(
					reservation.Aliases,
					fmt.Sprintf(
						"%v-mgmt",
						ncn.Hostname,
					),
				)
				subnet.IPReservations[index] = reservation
			}
		}
		if strings.ToLower(network.Name) == "nmn" {
			reservation.Aliases = append(
				reservation.Aliases,
				fmt.Sprintf(
					"%v.local",
					reservation.Name,
				),
			)
			subnet.IPReservations[index] = reservation
		}
	}
	return subnet, err
}

func writeOutput(
	v *viper.Viper,
	shastaNetworks map[string]*networking.IPNetwork,
	slsState slsCommon.SLSState,
	logicalNCNs []LogicalNCN,
	switches []*networking.ManagementSwitch,
	globalMetaData interface{},
) (err error) {
	basepath, _ := setupDirectories(
		v.GetString("system-name"),
		v,
	)

	err = files.WriteJSONConfig(
		filepath.Join(
			basepath,
			slsInit.OutputFile,
		),
		&slsState,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to encode SLS state because %v",
			err,
		)
	}
	v.Set(
		"VersionInfo",
		version.Get(),
	)

	err = WriteConfigAs(
		filepath.Join(
			basepath,
			cli.ConfigFilename,
		),
	)
	if err != nil {
		return fmt.Errorf(
			"failed to create %s because %v",
			cli.ConfigFilename,
			err,
		)
	}

	err = files.WriteYAMLConfig(
		filepath.Join(
			basepath,
			"customizations.yaml",
		),
		GenCustomizationsYaml(
			logicalNCNs,
			shastaNetworks,
			switches,
		),
	)
	if err != nil {
		return fmt.Errorf(
			"failed to create customizations YAML because %v",
			err,
		)
	}
	var pit LogicalNCN
	for _, ncn := range logicalNCNs {
		// log.Println("Checking to see if we need PIT files for ", ncn.Hostname)
		if strings.HasPrefix(
			ncn.Hostname,
			v.GetString("install-ncn"),
		) {
			log.Println(
				"Generating Installer Node (PIT) interface configurations for:",
				ncn.Hostname,
			)
			pit = ncn
			break
		}
	}
	err = WriteCPTNetworkConfig(
		filepath.Join(
			basepath,
			"pit-files",
		),
		v,
		pit,
		shastaNetworks,
	)
	if err != nil {
		return
	}
	err = WriteDNSMasqConfig(
		basepath,
		v,
		logicalNCNs,
		shastaNetworks,
	)
	if err != nil {
		return err
	}
	err = WriteConmanConfig(
		filepath.Join(
			basepath,
			"conman.conf",
		),
		logicalNCNs,
	)
	if err != nil {
		return err
	}
	// The MetalLB ConfigMap is no longer used in CSM 1.7
	_, eval := csm.CompareMajorMinor("1.7")
	if eval == -1 {
		err := WriteMetalLBConfigMap(
			basepath,
			v,
			shastaNetworks,
			switches,
		)
		if err != nil {
			return err
		}
	}
	WriteBasecampData(
		filepath.Join(
			basepath,
			"basecamp/data.json",
		),
		logicalNCNs,
		shastaNetworks,
		globalMetaData,
	)
	return err
}

func validateFlags() []string {
	var errors []string
	v := viper.GetViper()

	var cidrFlags = []string{
		"can-cidr",
		"can-dynamic-pool",
		"can-static-pool",
		"chn-cidr4",
		"chn-cidr6",
		"chn-dynamic-pool",
		"chn-static-pool",
		"cmn-cidr4",
		"cmn-cidr6",
		"cmn-dynamic-pool",
		"cmn-static-pool",
		"hmn-cidr",
		"nmn-cidr",
		"site-ip",
	}

	var ipv4Flags = []string{
		"site-dns",
		"site-gw",
	}

	// Validate the CSM version, exiting early if there is a potential mismatch. Proper guidance for fixing parameters can't be given without aligning intentions.
	detectedVersion, versionEnvError := csm.DetectedVersion()
	if versionEnvError != nil {
		// non-fatal; csi config init can be called without the csm.APIEnvName set, but the user should be notified that the env is CSM release-less.
		log.Print(versionEnvError)
	} else {
		// The currentVersion needs to be canonical for semver to compare it.
		// Exit and complain if the version of the inputs does not match the environment. It is not clear whether the inputs or the environment are wrong.
		currentVersion, eval := csm.CompareMajorMinor(detectedVersion)
		if eval != 0 {
			log.Fatalf(
				"Detected a potential mismatch of parameters and CSM environment!\n\n%15s=%-20s (a.k.a. %s)\n%15s=%-20s (a.k.a. %s)\n\nERROR: [%s != %s]\n\n- Both values must have matching <major>.<minor> versions\n- %s must be greater than this CSI's minimum allowed CSM version of: [%s]\n\nSince these values did not match one another it is possible that other inputs are also wrong.\nIf inputs from a prior release are being used, please double-check and/or refresh all other inputs and try again.\n",
				csm.APIKeyName,
				currentVersion,
				semver.MajorMinor(currentVersion),
				csm.APIEnvName,
				detectedVersion,
				semver.MajorMinor(detectedVersion),
				semver.MajorMinor(detectedVersion),
				semver.MajorMinor(currentVersion),
				csm.APIKeyName,
				csm.MinimumVersion,
			)
		}
	}
	currentVersion, versionError := csm.IsCompatible()
	if versionError != nil {
		log.Fatal(versionError)
	}
	log.Printf(
		"[%s] was set to [%s]; All inputs are targeted for CSM %s",
		csm.APIKeyName,
		currentVersion,
		currentVersion,
	)

	// BiCAN Validation.
	validBican := false
	bicanFlag := "bican-user-network-name"
	bicanNetworkName := v.GetString(bicanFlag)
	for _, value := range [3]string{
		"CAN",
		"CHN",
		"HSN",
	} {
		if bicanNetworkName == value {
			validBican = true
			break
		}
	}
	if !validBican {
		errors = append(
			errors,
			fmt.Sprintf(
				"%v must be set to CAN, CHN or HSN. (HSN requires NAT device)",
				bicanFlag,
			),
		)
	} else {
		if v.IsSet("bican-user-network-name") {
			if bicanNetworkName == "CAN" {
				if !v.IsSet("can-gateway") || v.GetString("can-gateway") == "" {
					errors = append(
						errors,
						fmt.Sprintln("can-gateway is required because bican-user-network-name is set to CAN but can-gateway was not set or was blank."),
					)
				} else {
					ipv4Flags = append(
						ipv4Flags,
						"can-gateway",
					)
				}
			} else if bicanNetworkName == "CHN" {
				if !v.IsSet("chn-gateway4") || v.GetString("chn-gateway4") == "" {
					errors = append(
						errors,
						fmt.Sprintln("chn-gateway4 is required because bican-user-network-name is set to CHN but chn-gateway4 was not set or was blank."),
					)
				} else {
					ipv4Flags = append(
						ipv4Flags,
						"chn-gateway4",
					)
				}
			}
		}
	}

	for _, flagName := range ipv4Flags {
		if v.IsSet(flagName) {
			if net.ParseIP(v.GetString(flagName)) == nil {
				errors = append(
					errors,
					fmt.Sprintf(
						"%v should be an ip address and is not set correctly through flag or config file (.%s)",
						flagName,
						v.ConfigFileUsed(),
					),
				)
			}
		}
	}

	for _, flagName := range cidrFlags {
		if v.IsSet(flagName) && (v.GetString(flagName) != "") {
			_, _, err := net.ParseCIDR(v.GetString(flagName))
			if err != nil {
				errors = append(
					errors,
					fmt.Sprintf(
						"%v should be a CIDR in the form 192.168.0.1/24 or a CIDR6 2001:db8::/32 and is not set correctly through flag or config file (.%s)",
						flagName,
						v.ConfigFileUsed(),
					),
				)
			}
		}
	}

	if v.IsSet("cilium-operator-replicas") {
		validFlag := false
		cor := v.GetInt("cilium-operator-replicas")
		if cor > 0 {
			validFlag = true
		}
		if !validFlag {
			errors = append(
				errors,
				"cilium-operator-replicas must be an integer > 0",
			)
		}
	} else {
		v.Set(
			"cilium-operator-replicas",
			"1",
		)
		NoWriteKeys = append(
			NoWriteKeys,
			"cilium-operator-replicas",
		)
	}

	if v.IsSet("cilium-kube-proxy-replacement") {
		validFlag := false
		for _, value := range [3]string{
			"strict",
			"partial",
			"disabled",
		} {
			if v.GetString("cilium-kube-proxy-replacement") == value {
				validFlag = true
				break
			}
		}
		if !validFlag {
			errors = append(
				errors,
				"cilium-kube-proxy-replacement must be set to strict, partial, or disabled",
			)
		}
	} else {
		v.Set(
			"cilium-kube-proxy-replacement",
			"disabled",
		)
		NoWriteKeys = append(
			NoWriteKeys,
			"cilium-kube-proxy-replacement",
		)
	}

	if v.IsSet("k8s-primary-cni") {
		validFlag := false
		_, eval := csm.CompareMajorMinor("1.7")
		var allowedValues []string
		if eval != -1 {
			allowedValues = []string{"cilium"}
		} else {
			allowedValues = []string{"cilium", "weave"}
		}
		if slices.Contains(
			allowedValues,
			v.GetString("k8s-primary-cni"),
		) {
			validFlag = true
		}
		if !validFlag {
			errors = append(
				errors,
				fmt.Sprintf(
					"k8s-primary-cni must be set to one of [%s]",
					strings.Join(
						allowedValues,
						",",
					),
				),
			)
		}
	} else {
		currentVersion, eval := csm.CompareMajorMinor("1.7")
		if eval != -1 {
			log.Printf(
				"Detected CSM %s, setting k8s-primary-cni to Cilium",
				currentVersion,
			)
			v.Set(
				"k8s-primary-cni",
				"cilium",
			)
		} else {
			v.Set(
				"k8s-primary-cni",
				"weave",
			)
			NoWriteKeys = append(
				NoWriteKeys,
				"k8s-primary-cni",
			)
		}
	}

	if v.IsSet("kubernetes-max-pods-per-node") {
		validFlag := false
		if v.GetInt("kubernetes-max-pods-per-node") > 0 {
			validFlag = true
		}
		if !validFlag {
			errors = append(
				errors,
				"kubernetes-max-pods-per-node must be set an integer (>0)",
			)
		}
	} else {
		v.Set(
			"kubernetes-max-pods-per-node",
			"200",
		)
		NoWriteKeys = append(
			NoWriteKeys,
			"kubernetes-max-pods-per-node",
		)
	}

	if v.IsSet("kubernetes-pods-cidr") {
		validFlag := false
		_, err := netip.ParsePrefix(v.GetString("kubernetes-pods-cidr"))
		if err == nil {
			validFlag = true
		}
		if !validFlag {
			errors = append(
				errors,
				"kubernetes-pods-cidr was set to an invalid IP",
			)
		}
	} else {
		v.Set(
			"kubernetes-pods-cidr",
			"10.32.0.0/12",
		)
		NoWriteKeys = append(
			NoWriteKeys,
			"kubernetes-pods-cidr",
		)
	}

	if v.IsSet("kubernetes-services-cidr") {
		validFlag := false
		_, err := netip.ParsePrefix(v.GetString("kubernetes-services-cidr"))
		if err == nil {
			validFlag = true
		}
		if !validFlag {
			errors = append(
				errors,
				"kubernetes-services-cidr was set to an invalid IP",
			)
		}
	} else {
		v.Set(
			"kubernetes-services-cidr",
			"10.16.0.0/12",
		)
		NoWriteKeys = append(
			NoWriteKeys,
			"kubernetes-services-cidr",
		)
	}

	if v.IsSet("kubernetes-weave-mtu") {
		validFlag := false
		if v.GetInt("kubernetes-weave-mtu") > 0 {
			validFlag = true
		}
		if !validFlag {
			errors = append(
				errors,
				"kubernetes-weave-mtu must be set an integer (>0)",
			)
		}
	} else {
		v.Set(
			"kubernetes-weave-mtu",
			"1376",
		)
		NoWriteKeys = append(
			NoWriteKeys,
			"kubernetes-weave-mtu",
		)
	}

	if v.IsSet("wipe-ceph-osds") {
		validFlag := false
		for _, value := range []string{
			"yes",
			"no",
		} {
			if v.GetString("wipe-ceph-osds") == value {
				validFlag = true
				break
			}
		}
		if !validFlag {
			errors = append(
				errors,
				"wipe-ceph-osds must be set to yes or no",
			)
		}
	} else {
		v.Set(
			"wipe-ceph-osds",
			"yes",
		)
		NoWriteKeys = append(
			NoWriteKeys,
			"wipe-ceph-osds",
		)
	}

	if v.IsSet("domain") {
		validFlag := false
		if v.GetString("domain") != "" {
			validFlag = true
		}
		if !validFlag {
			errors = append(
				errors,
				"domain must be set a string (delimited by spaces) of search domains.",
			)
		}
	} else {
		v.Set(
			"domain",
			"nmn mtl hmn",
		)
		NoWriteKeys = append(
			NoWriteKeys,
			"domain",
		)
	}

	return errors
}

// AllocateIPs distributes IP reservations for each of the NCNs within the networks
func AllocateIPs(ncns []*LogicalNCN, networks map[string]*networking.IPNetwork) {
	lookup := func(name string, subnetName string, networks map[string]*networking.IPNetwork) (
		subnet *slsCommon.IPSubnet, err error,
	) {
		tempNetwork := networks[name]
		subnet, err = tempNetwork.LookUpSubnet(subnetName)
		if err != nil {
			return subnet, fmt.Errorf(
				"couldn't find a %v subnet in the %v network",
				subnetName,
				name,
			)
		}
		return subnet, err
	}

	// Build a map of networks based on their names
	subnets := make(map[string]*slsCommon.IPSubnet)
	for network := range networks {
		bootstrapNet, err := lookup(
			network,
			"bootstrap_dhcp",
			networks,
		)
		if err == nil {
			subnets[network] = bootstrapNet
		}
	}

	// Loop through the NCNs and then run through the networks to add reservations and assign ip addresses
	for _, ncn := range ncns {
		ncn.InstanceID = GenerateInstanceID()
		for netName, subnet := range subnets {
			// reserve the bmc ip
			if strings.ToLower(netName) == "hmn" {
				// The bmc xname is the ncn xname without the final two characters
				// NCN Xname = x3000c0s9b0n0  BMC Xname = x3000c0s9b0
				reservation, err := networking.AddReservation(
					subnet,
					fmt.Sprintf(
						"%v",
						strings.TrimSuffix(
							ncn.Xname,
							"n0",
						),
					),
					fmt.Sprintf(
						"%v-mgmt",
						ncn.Xname,
					),
				)
				if err != nil {
					log.Fatalf(
						"failed to allocate hmn reservation %v because %v",
						ncn.Xname,
						err,
					)
				}
				ncn.BmcIP = reservation.IPAddress.String()
			}
			reservation, err := networking.AddReservation(
				subnet,
				ncn.Xname,
				ncn.Xname,
			)
			if err != nil {
				log.Fatalf(
					"Failed to allocate reservation for %s: %v",
					ncn.Xname,
					err,
				)
			}
			reservationAddr, _ := netip.ParseAddr(reservation.IPAddress.String())
			prefix, _ := netip.ParsePrefix(subnet.CIDR)

			interfaceName, err := networks[netName].GenInterfaceName(subnet)
			if err != nil {
				log.Fatalf(
					"Couldn't generate interface name for %v on subnet %v\n",
					ncn.Xname,
					subnet.CIDR,
				)
			}
			addr4, err := netip.ParseAddr(reservation.IPAddress.String())
			if err != nil {
				log.Fatalf(
					"Failed to parse network IP address %v because %v",
					reservation.IPAddress.String(),
					err,
				)
			}
			gw4, err := netip.ParseAddr(subnet.Gateway.String())
			if err != nil {
				log.Fatalf(
					"Failed to parse gateway IP address %v because %v",
					subnet.Gateway.String(),
					err,
				)
			}
			tempNetwork := NCNNetwork{
				NetworkName: networks[netName].Name,
				IPv4Address: addr4,
				Vlan:        int(subnet.VlanID),
				FullName:    subnet.FullName,
				CIDR4: netip.PrefixFrom(
					reservationAddr,
					prefix.Bits(),
				),
				InterfaceName:       interfaceName,
				ParentInterfaceName: networks[netName].ParentDevice,
				Gateway4:            gw4,
			}

			if reservation.IPAddress6 != nil {
				addr6, err := netip.ParseAddr(reservation.IPAddress6.String())
				if err != nil {
					log.Fatalf(
						"Host %s had an unparseable address for IPv6: %v",
						ncn.Hostname,
						reservation.IPAddress6.String(),
					)
				}

				prefix6, err := netip.ParsePrefix(subnet.CIDR6)
				if err != nil {
					log.Fatalf(
						"Unparseable subnet IPv6 CIDR for %s: %v",
						subnet.Name,
						subnet.CIDR6,
					)
				}

				gw6, err := netip.ParseAddr(subnet.Gateway6.String())
				if err != nil {
					log.Fatalf(
						"Unparseable subnet Gateway IPv6 CIDR for %s: %v",
						subnet.Name,
						subnet.Gateway6,
					)
				}
				tempNetwork.IPv6Address = addr6
				tempNetwork.CIDR6 = netip.PrefixFrom(
					addr6,
					prefix6.Bits(),
				)
				tempNetwork.Gateway6 = gw6

			}

			ncn.Networks = append(
				ncn.Networks,
				tempNetwork,
			)

		}
	}
}
