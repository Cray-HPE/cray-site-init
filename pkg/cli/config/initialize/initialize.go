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

package initialize

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/mod/semver"

	shcdParser "github.com/Cray-HPE/hms-shcd-parser/pkg/shcd-parser"
	slsCommon "github.com/Cray-HPE/hms-sls/pkg/sls-common"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	slsInit "github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/csm"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
	"github.com/Cray-HPE/cray-site-init/pkg/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/version"
)

var appVersion string

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

			clientVersion := version.Get()
			appVersion = clientVersion.Version

			err := v.BindPFlags(c.Flags())
			if err != nil {
				log.Fatalln(err)
			}
			flagErrors := validateFlags()
			if len(flagErrors) > 0 {
				c.Usage()
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
				c.Usage()
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
					c.Usage()
					log.Fatalln(
						"FATAL ERROR: BGP ASNs must be within the private range 64512-65534, fix the value for:",
						network,
					)
				}
			}

			// Read and validate our three input files
			hmnRows, logicalNcns, switches, applicationNodeConfig, cabinetDetailList := collectInput(v)

			var riverCabinetCount, mountainCabinetCount, hillCabinetCount int
			for _, cab := range cabinetDetailList {
				switch class, _ := cab.Kind.Class(); class {
				case slsCommon.ClassRiver:
					riverCabinetCount += len(cab.CabinetIDs())
				case slsCommon.ClassMountain:
					mountainCabinetCount += len(cab.CabinetIDs())
				case slsCommon.ClassHill:
					hillCabinetCount += len(cab.CabinetIDs())
				}
			}

			// Prepare the network layout configs for generating the networks
			var internalNetConfigs = make(map[string]slsInit.NetworkLayoutConfiguration)
			internalNetConfigs["BICAN"] = slsInit.GenDefaultBICANConfig(v.GetString("bican-user-network-name"))
			internalNetConfigs["CMN"] = slsInit.GenDefaultCMNConfig(
				len(logicalNcns),
				len(switches),
			)
			internalNetConfigs["HMN"] = slsInit.GenDefaultHMNConfig()
			internalNetConfigs["HSN"] = slsInit.GenDefaultHSNConfig()
			internalNetConfigs["MTL"] = slsInit.GenDefaultMTLConfig()
			internalNetConfigs["NMN"] = slsInit.GenDefaultNMNConfig()
			if v.GetString("bican-user-network-name") == "CAN" || v.GetBool("retain-unused-user-network") {
				internalNetConfigs["CAN"] = slsInit.GenDefaultCANConfig()
			}
			if v.GetString("bican-user-network-name") == "CHN" || v.GetBool("retain-unused-user-network") {
				internalNetConfigs["CHN"] = slsInit.GenDefaultCHNConfig()
			}

			if internalNetConfigs["HMN"].GroupNetworksByCabinetType {
				if mountainCabinetCount > 0 || hillCabinetCount > 0 {
					tmpHmnMtn := slsInit.GenDefaultHMNConfig()
					tmpHmnMtn.Template.Name = "HMN_MTN"
					tmpHmnMtn.Template.FullName = "Mountain Compute Hardware Management Network"
					tmpHmnMtn.Template.VlanRange = []int16{
						3000,
						3999,
					}
					tmpHmnMtn.Template.CIDR = networking.DefaultHMNMTNString
					tmpHmnMtn.SubdivideByCabinet = true
					tmpHmnMtn.IncludeBootstrapDHCP = false
					tmpHmnMtn.SuperNetHack = false
					tmpHmnMtn.IncludeNetworkingHardwareSubnet = false
					internalNetConfigs["HMN_MTN"] = tmpHmnMtn
				}
				if riverCabinetCount > 0 {
					tmpHmnRvr := slsInit.GenDefaultHMNConfig()
					tmpHmnRvr.Template.Name = "HMN_RVR"
					tmpHmnRvr.Template.FullName = "River Compute Hardware Management Network"
					tmpHmnRvr.Template.VlanRange = []int16{
						1513,
						1769,
					}
					tmpHmnRvr.Template.CIDR = networking.DefaultHMNRVRString
					tmpHmnRvr.SubdivideByCabinet = true
					tmpHmnRvr.IncludeBootstrapDHCP = false
					tmpHmnRvr.SuperNetHack = false
					tmpHmnRvr.IncludeNetworkingHardwareSubnet = false
					internalNetConfigs["HMN_RVR"] = tmpHmnRvr
				}

			}

			if internalNetConfigs["NMN"].GroupNetworksByCabinetType {
				if mountainCabinetCount > 0 || hillCabinetCount > 0 {
					tmpNmnMtn := slsInit.GenDefaultNMNConfig()
					tmpNmnMtn.Template.Name = "NMN_MTN"
					tmpNmnMtn.Template.FullName = "Mountain Compute Node Management Network"
					tmpNmnMtn.Template.VlanRange = []int16{
						2000,
						2999,
					}
					tmpNmnMtn.Template.CIDR = networking.DefaultNMNMTNString
					tmpNmnMtn.SubdivideByCabinet = true
					tmpNmnMtn.IncludeBootstrapDHCP = false
					tmpNmnMtn.SuperNetHack = false
					tmpNmnMtn.IncludeNetworkingHardwareSubnet = false
					tmpNmnMtn.IncludeUAISubnet = false
					internalNetConfigs["NMN_MTN"] = tmpNmnMtn
				}
				if riverCabinetCount > 0 {
					tmpNmnRvr := slsInit.GenDefaultNMNConfig()
					tmpNmnRvr.Template.Name = "NMN_RVR"
					tmpNmnRvr.Template.FullName = "River Compute Node Management Network"
					tmpNmnRvr.Template.VlanRange = []int16{
						1770,
						1999,
					}
					tmpNmnRvr.Template.CIDR = networking.DefaultNMNRVRString
					tmpNmnRvr.SubdivideByCabinet = true
					tmpNmnRvr.IncludeBootstrapDHCP = false
					tmpNmnRvr.SuperNetHack = false
					tmpNmnRvr.IncludeNetworkingHardwareSubnet = false
					tmpNmnRvr.IncludeUAISubnet = false
					internalNetConfigs["NMN_RVR"] = tmpNmnRvr
				}

			}

			// Remember a loop over a map is random ordered in Go
			for name, layout := range internalNetConfigs {
				myLayout := layout

				// Update with flags
				normalizedName := strings.ReplaceAll(
					strings.ToLower(name),
					"_",
					"-",
				)

				// Use CLI/file input values if available, otherwise defaults
				baseVlanName := fmt.Sprintf("%v-bootstrap-vlan", normalizedName)
				if v.IsSet(baseVlanName) {
					baseVlan := int16(v.GetInt(baseVlanName))
					myLayout.BaseVlan = baseVlan
					myLayout.Template.VlanRange[0] = baseVlan
				} else {
					myLayout.BaseVlan = layout.Template.VlanRange[0]
				}

				// Check VLAN allocations for re-use and overlaps
				if len(layout.Template.VlanRange) == 2 {
					err := networking.AllocateVlanRange(
						myLayout.Template.VlanRange[0],
						myLayout.Template.VlanRange[1],
					)
					if err != nil {
						log.Fatalln(
							"Unable to allocate VLAN range for",
							myLayout.Template.Name,
							err,
						)
					} else {
						log.Println(
							"Allocating VLANs",
							myLayout.Template.Name,
							myLayout.Template.VlanRange[0],
							myLayout.Template.VlanRange[1],
						)
					}
				} else {
					err := networking.AllocateVlan(
						myLayout.Template.VlanRange[0],
					)
					if err != nil {
						log.Fatalln(
							"Unable to allocate single VLAN for",
							myLayout.Template.Name,
							myLayout.Template.VlanRange[0],
							err,
						)
					} else {
						log.Println(
							"Allocating VLAN ",
							myLayout.Template.Name,
							myLayout.Template.VlanRange[0],
						)
					}
				}

				allocated, err := networking.IsVlanAllocated(myLayout.BaseVlan)
				if !allocated {
					log.Fatalln(
						"VLAN for",
						layout.Template.Name,
						"has not been initialized by defaults or input values:",
						err,
					)
				}

				if v.IsSet(
					fmt.Sprintf(
						"%v-cidr",
						normalizedName,
					),
				) {
					myLayout.Template.CIDR = v.GetString(
						fmt.Sprintf(
							"%v-cidr",
							normalizedName,
						),
					)
				}

				myLayout.AdditionalNetworkingSpace = v.GetInt("management-net-ips")
				internalNetConfigs[name] = myLayout
			}

			// Build a set of networks we can use
			shastaNetworks, err := slsInit.BuildCSMNetworks(
				internalNetConfigs,
				cabinetDetailList,
				switches,
			)
			if err != nil {
				log.Panic(err)
			}

			// Use our new networks and our list of logicalNCNs to distribute ips
			AllocateIps(
				logicalNcns,
				shastaNetworks,
			) // This function has no return because it is working with lists of pointers.

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
				logicalNcns,
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
					canSubnet.AddReservation(
						uan.Hostname,
						uan.Xname,
					)
				}
			}
			// Only add UANs if there actually is a CHN network
			if v.GetString("bican-user-network-name") == "CHN" || v.GetBool("retain-unused-user-network") {
				chnSubnet, _ := shastaNetworks["CHN"].LookUpSubnet("bootstrap_dhcp")
				for _, uan := range slsUans {
					chnSubnet.AddReservation(
						uan.Hostname,
						uan.Xname,
					)
				}
			}

			// Cycle through the main networks and update the reservations, masks and dhcp ranges as necessary
			for _, netName := range networking.ValidNetNames {
				if shastaNetworks[netName] != nil {
					// Grab the supernet details for use in HACK substitution
					tempSubnet, err := shastaNetworks[netName].LookUpSubnet("bootstrap_dhcp")
					if err == nil {
						// Loop the reservations and update the NCN reservations with hostnames
						// we likely didn't have when we registered the reservation
						updateReservations(
							tempSubnet,
							logicalNcns,
						)
						if netName == "CAN" || netName == "CMN" || netName == "CHN" {
							netNameLower := strings.ToLower(netName)

							// Do not use supernet hack for the CAN/CMN/CHN
							tempSubnet.UpdateDHCPRange(false)

							myNetName := fmt.Sprintf(
								"%s-cidr",
								netNameLower,
							)
							myNetCIDR := v.GetString(myNetName)
							if myNetCIDR == "" {
								continue
							}
							_, myNet, _ := net.ParseCIDR(myNetCIDR)

							// If neither static nor dynamic pool is defined we can use the last available IP in the subnet
							poolStartIP := networking.Broadcast(*myNet)

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
								if networking.IPLessThan(
									myStaticPool.IP,
									myDynamicPool.IP,
								) {
									poolStartIP = myStaticPool.IP
								} else {
									poolStartIP = myDynamicPool.IP
								}
							} else if len(myStaticPoolCIDR) > 0 && len(myDynPoolCIDR) == 0 {
								// Only the static pool is defined so use the first IP of that pool
								_, myStaticPool, _ := net.ParseCIDR(myStaticPoolCIDR)
								poolStartIP = myStaticPool.IP
							} else if len(myStaticPoolCIDR) == 0 && len(myDynPoolCIDR) > 0 {
								// Only the dynamic pool is defined so use the first IP of that pool
								_, myDynamicPool, _ := net.ParseCIDR(myDynPoolCIDR)
								poolStartIP = myDynamicPool.IP
							}

							// Guidance has changed on whether the CAN gw should be at the start or end of the
							// range. Here we account for it being at the end of the range.
							// Leaving this check in place for CMN because it is harmless to do so.
							if tempSubnet.Gateway.String() == networking.Add(
								poolStartIP,
								-1,
							).String() {
								// The gw *is* at the end, so shorten the range to accommodate
								tempSubnet.DHCPEnd = networking.Add(
									poolStartIP,
									-2,
								)
							} else {
								// The gw is not at the end
								tempSubnet.DHCPEnd = networking.Add(
									poolStartIP,
									-1,
								)
							}
						} else {
							tempSubnet.UpdateDHCPRange(v.GetBool("supernet"))
						}
					}

					// We expect a bootstrap_dhcp in every net, but uai_macvlan is only in
					// the NMN range for today
					if netName == "NMN" {
						tempSubnet, err = shastaNetworks[netName].LookUpSubnet("uai_macvlan")
						if err != nil {
							log.Panic(err)
						}
						updateReservations(
							tempSubnet,
							logicalNcns,
						)
						tempSubnet.UpdateDHCPRange(false)
					}

				}
			}

			// Update the SLSState with the updated network information
			_, slsState.Networks = prepareNetworkSLS(shastaNetworks)

			// Switch from a list of pointers to a list of things before we write it out
			var ncns []LogicalNCN
			for _, ncn := range logicalNcns {
				ncns = append(
					ncns,
					*ncn,
				)
			}
			globals, err := MakeBasecampGlobals(
				v,
				ncns,
				shastaNetworks,
				"NMN",
				"bootstrap_dhcp",
				v.GetString("install-ncn"),
			)
			if err != nil {
				log.Fatalln(
					"unable to generate basecamp globals: ",
					err,
				)
			}
			writeOutput(
				v,
				shastaNetworks,
				slsState,
				ncns,
				switches,
				globals,
			)

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
				"\n===== %v Installation Summary =====\n\n",
				v.GetString("system-name"),
			)
			fmt.Printf(
				"Installation Node: %v\n",
				v.GetString("install-ncn"),
			)
			fmt.Printf(
				"Customer Management: %v GW: %v\n",
				v.GetString("cmn-cidr"),
				v.GetString("cmn-gateway"),
			)
			if v.GetString("bican-user-network-name") == "CHN" {
				fmt.Printf(
					"Customer Management: %v GW: %v\n",
					v.GetString("can-cidr"),
					v.GetString("can-gateway"),
				)
			} else {
				fmt.Printf(
					"Customer Management: %v GW: %v\n",
					v.GetString("can-cidr"),
					v.GetString("can-gateway"),
				)
			}
			fmt.Printf(
				"\tUpstream DNS: %v\n",
				v.GetString("site-dns"),
			)
			fmt.Printf(
				"\tMetalLB Peers: %v\n",
				v.GetStringSlice("bgp-peer-types"),
			)
			fmt.Println("Networking")
			fmt.Printf(
				"\tBICAN user network toggle set to %v\n",
				v.GetString("bican-user-network-name"),
			)
			if v.GetBool("supernet") {
				fmt.Printf("\tSupernet enabled!  Using the supernet gateway for some management subnets\n")
			}
			for _, tempNet := range shastaNetworks {
				fmt.Printf(
					"\t* %v %v with %d subnets \n",
					tempNet.FullName,
					tempNet.CIDR,
					len(tempNet.Subnets),
				)
			}
			fmt.Printf("System Information\n")
			fmt.Printf(
				"\tNCNs: %v\n",
				len(ncns),
			)
			fmt.Printf(
				"\tMountain Compute Cabinets: %v\n",
				len(slsMountainCabinets),
			)
			fmt.Printf(
				"\tHill Compute Cabinets: %v\n",
				len(slsHillCabinets),
			)
			fmt.Printf(
				"\tRiver Compute Cabinets: %v\n",
				len(slsRiverCabinets),
			)
			fmt.Printf(
				"CSI Version Information\n\t%s\n\t%s\n\n",
				version.Get().GitCommit,
				version.Get(),
			)
		},
	}

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
		"Site Network IPv4 Gateway",
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
		"chn-static-pool",
		"chn-dynamic-pool",
	)

	// Customer management network.
	c.Flags().String(
		"cmn-cidr",
		"",
		"Overall IPv4 CIDR for all Customer Management subnets",
	)
	c.Flags().String(
		"cmn-gateway",
		"",
		"Gateway for NCNs on the CMN (Administrative/Management)",
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
	err := c.MarkFlagRequired("bootstrap-ncn-bmc-pass")
	if err != nil {
		return nil
	}
	err = c.MarkFlagRequired("bootstrap-ncn-bmc-user")
	if err != nil {
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
	// Set up the path for our base directory using our systemname
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
				// log.Println("Updated to be :", ncn)
				// ncn.InstanceID = GenerateInstanceID()

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

func prepareNetworkSLS(shastaNetworks map[string]*networking.IPV4Network) (
	[]networking.IPV4Network, map[string]slsCommon.Network,
) {
	// Fix up the network names & create the SLS friendly version of the shasta networks
	var networks []networking.IPV4Network
	for name, network := range shastaNetworks {
		if network.Name == "" {
			network.Name = name
		}
		networks = append(
			networks,
			*network,
		)
	}
	return networks, slsInit.ConvertIPV4NetworksToSLS(&networks)
}

func prepareAndGenerateSLS(
	cd []sls.CabinetGroupDetail,
	shastaNetworks map[string]*networking.IPV4Network,
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

		// Extract Switch brand from data stored in switch_metadata.csv
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

func updateReservations(tempSubnet *networking.IPV4Subnet, logicalNcns []*LogicalNCN) {
	// Loop the reservations and update the NCN reservations with hostnames
	// we likely didn't have when we registered the reservation
	for index, reservation := range tempSubnet.IPReservations {
		for _, ncn := range logicalNcns {
			if reservation.Comment == ncn.Xname {
				reservation.Name = ncn.Hostname
				reservation.Aliases = append(
					reservation.Aliases,
					fmt.Sprintf(
						"%v-%v",
						ncn.Hostname,
						strings.ToLower(tempSubnet.NetName),
					),
				)
				reservation.Aliases = append(
					reservation.Aliases,
					fmt.Sprintf(
						"time-%v",
						strings.ToLower(tempSubnet.NetName),
					),
				)
				reservation.Aliases = append(
					reservation.Aliases,
					fmt.Sprintf(
						"time-%v.local",
						strings.ToLower(tempSubnet.NetName),
					),
				)
				if strings.ToLower(ncn.Subrole) == "storage" && strings.ToLower(tempSubnet.NetName) == "hmn" {
					reservation.Aliases = append(
						reservation.Aliases,
						"rgw-vip.hmn",
					)
				}
				if strings.ToLower(tempSubnet.NetName) == "nmn" {
					// The xname of a NCN will point to its NMN IP address
					reservation.Aliases = append(
						reservation.Aliases,
						ncn.Xname,
					)
				}
				tempSubnet.IPReservations[index] = reservation
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
				tempSubnet.IPReservations[index] = reservation
			}
		}
		if tempSubnet.NetName == "NMN" {
			reservation.Aliases = append(
				reservation.Aliases,
				fmt.Sprintf(
					"%v.local",
					reservation.Name,
				),
			)
			tempSubnet.IPReservations[index] = reservation
		}
	}
}

func writeOutput(
	v *viper.Viper,
	shastaNetworks map[string]*networking.IPV4Network,
	slsState slsCommon.SLSState,
	logicalNCNs []LogicalNCN,
	switches []*networking.ManagementSwitch,
	globals interface{},
) {
	basepath, _ := setupDirectories(
		v.GetString("system-name"),
		v,
	)
	err := csiFiles.WriteJSONConfig(
		filepath.Join(
			basepath,
			"sls_input_file.json",
		),
		&slsState,
	)
	if err != nil {
		log.Fatalln(
			"Failed to encode SLS state:",
			err,
		)
	}
	v.SetConfigName("system_config")
	v.SetConfigType("yaml")
	v.Set(
		"VersionInfo",
		version.Get(),
	)
	v.WriteConfigAs(
		filepath.Join(
			basepath,
			"system_config.yaml",
		),
	)

	csiFiles.WriteYAMLConfig(
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
			WriteCPTNetworkConfig(
				filepath.Join(
					basepath,
					"pit-files",
				),
				v,
				ncn,
				shastaNetworks,
			)
		}
	}
	WriteDNSMasqConfig(
		basepath,
		v,
		logicalNCNs,
		shastaNetworks,
	)
	WriteConmanConfig(
		filepath.Join(
			basepath,
			"conman.conf",
		),
		logicalNCNs,
	)
	WriteMetalLBCRD(
		basepath,
		v,
		shastaNetworks,
		switches,
	)
	WriteBasecampData(
		filepath.Join(
			basepath,
			"basecamp/data.json",
		),
		logicalNCNs,
		shastaNetworks,
		globals,
	)

}
func validateFlags() []string {
	var errors []string
	v := viper.GetViper()

	var cidrFlags = []string{
		"can-cidr",
		"can-dynamic-pool",
		"can-static-pool",
		"chn-cidr",
		"chn-dynamic-pool",
		"chn-static-pool",
		"cmn-cidr",
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

	// for _, flagName := range requiredFlags {
	// 	if !v.IsSet(flagName) || (v.GetString(flagName) == "") {
	// 		errors = append(errors, fmt.Sprintf("%v is required and not set through flag or config file (%s)", flagName, v.ConfigFileUsed()))
	// 	}
	// }

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
				if !v.IsSet("chn-gateway") || v.GetString("chn-gateway") == "" {
					errors = append(
						errors,
						fmt.Sprintln("chn-gateway is required because bican-user-network-name is set to CHN but chn-gateway was not set or was blank."),
					)
				} else {
					ipv4Flags = append(
						ipv4Flags,
						"chn-gateway",
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
						"%v should be a CIDR in the form 192.168.0.1/24 and is not set correctly through flag or config file (.%s)",
						flagName,
						v.ConfigFileUsed(),
					),
				)
			}
		}
	}

	if v.IsSet("cilium-operator-replicas") {
		validFlag := false
		var cor int
		cor = v.GetInt("cilium-operator-replicas")
		if cor > 0 {
			validFlag = true
		}
		if !validFlag {
			errors = append(
				errors,
				fmt.Sprintf("cilium-operator-replicas must be an integer > 0"),
			)
		}
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
				fmt.Sprintf("cilium-kube-proxy-replacement must be set to strict, partial, or disabled"),
			)
		}
	}

	if v.IsSet("k8s-primary-cni") {
		validFlag := false
		for _, value := range [3]string{
			"weave",
			"cilium",
		} {
			if v.GetString("k8s-primary-cni") == value {
				validFlag = true
				break
			}
		}
		if !validFlag {
			errors = append(
				errors,
				fmt.Sprintf("k8s-primary-cni must be set to weave or cilium"),
			)
		}
	}

	return errors
}

// AllocateIps distributes IP reservations for each of the NCNs within the networks
func AllocateIps(ncns []*LogicalNCN, networks map[string]*networking.IPV4Network) {
	// log.Printf("I'm here in AllocateIps with %d ncns to work with and %d networks", len(ncns), len(networks))
	lookup := func(name string, subnetName string, networks map[string]*networking.IPV4Network) (
		*networking.IPV4Subnet, error,
	) {
		tempNetwork := networks[name]
		subnet, err := tempNetwork.LookUpSubnet(subnetName)
		if err != nil {
			// log.Printf("couldn't find a %v subnet in the %v network \n", subnetName, name)
			return subnet, fmt.Errorf(
				"couldn't find a %v subnet in the %v network",
				subnetName,
				name,
			)
		}
		// log.Printf("found a %v subnet in the %v network", subnetName, name)
		return subnet, nil
	}

	// Build a map of networks based on their names
	subnets := make(map[string]*networking.IPV4Subnet)
	for name := range networks {
		bootstrapNet, err := lookup(
			name,
			"bootstrap_dhcp",
			networks,
		)
		if err == nil {
			subnets[name] = bootstrapNet
		}
	}

	// Loop through the NCNs and then run through the networks to add reservations and assign ip addresses
	for _, ncn := range ncns {
		ncn.InstanceID = GenerateInstanceID()
		for netName, subnet := range subnets {
			// reserve the bmc ip
			if netName == "HMN" {
				// The bmc xname is the ncn xname without the final two characters
				// NCN Xname = x3000c0s9b0n0  BMC Xname = x3000c0s9b0
				ncn.BmcIP = subnet.AddReservation(
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
				).IPAddress.String()
			}
			// Hostname is not available a the point AllocateIPs should be called.
			reservation := subnet.AddReservation(
				ncn.Xname,
				ncn.Xname,
			)
			// log.Printf("Adding %v %v reservation for %v(%v) at %v \n", netName, subnet.Name, ncn.Xname, ncn.Xname, reservation.IPAddress.String())
			prefixLen := strings.Split(
				subnet.CIDR.String(),
				"/",
			)[1]
			err := subnet.GenInterfaceName()
			if err != nil {
				// pass
			}
			tempNetwork := NCNNetwork{
				NetworkName: netName,
				IPAddress:   reservation.IPAddress.String(),
				Vlan:        int(subnet.VlanID),
				FullName:    subnet.FullName,
				CIDR: strings.Join(
					[]string{
						reservation.IPAddress.String(),
						prefixLen,
					},
					"/",
				),
				Mask:                prefixLen,
				InterfaceName:       subnet.InterfaceName,
				ParentInterfaceName: subnet.ParentDevice,
				Gateway:             subnet.Gateway,
			}
			ncn.Networks = append(
				ncn.Networks,
				tempNetwork,
			)

		}
	}
}
