//
//  MIT License
//
//  (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	"github.com/Cray-HPE/cray-site-init/pkg/ipam"
	"github.com/Cray-HPE/cray-site-init/pkg/pit"
	"github.com/Cray-HPE/cray-site-init/pkg/version"
	shcd_parser "github.com/Cray-HPE/hms-shcd-parser/pkg/shcd-parser"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generates a Shasta configuration payload",
	Long: `init generates a scaffolding the Shasta configuration payload.  It is based on several input files:
	1. The hmn_connections.json which describes the cabling for the BMCs on the NCNs
	2. The ncn_metadata.csv file documents the MAC addresses of the NCNs to be used in this installation
	   NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
	3. The switch_metadata.csv file which documents the Xname, Brand, Type, and Model of each switch.  Types are CDU, LeafBMC, Leaf, and Spine
	   Switch Xname,Type,Brand,Model

	** NB **
	For systems that use non-sequential cabinet id numbers, an additional mapping file is necessary and must be indicated
	with the --cabinets-yaml flag.
	** NB **

	** NB **
	For additional control of the application node identification durring the SLS Input File generation, an additional config file is necessary
	and must be indicated with the --application-node-config-yaml flag.

	Allows control of the following in the SLS Input File:
	1. System specific prefix for Applications node
	2. Specify HSM Subroles for system specifc application nodes
	3. Specify Application node Aliases
	** NB **

	In addition, there are many flags to impact the layout of the system.  The defaults are generally fine except for the networking flags.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		// Initialize the global viper
		v := viper.GetViper()

		v.BindPFlags(cmd.Flags())

		flagErrors := validateFlags()
		if len(flagErrors) > 0 {
			cmd.Usage()
			for _, e := range flagErrors {
				log.Println(e)
			}
			log.Fatal("One or more flags are invalid")
		}

		if len(strings.Split(v.GetString("site-ip"), "/")) != 2 {
			cmd.Usage()
			log.Fatalf("FATAL ERROR: Unable to parse %s as --site-ip.  Must be in the format \"192.168.0.1/24\"", v.GetString("site-ip"))

		}

		// Read and validate our three input files
		hmnRows, logicalNcns, switches, applicationNodeConfig, cabinetDetailList := collectInput(v)

		var riverCabinetCount, mountainCabinetCount, hillCabinetCount int
		for _, cab := range cabinetDetailList {

			log.Printf("\t%v: %d\n", cab.Kind, len(cab.CabinetIDs()))
			switch cab.Kind {
			case "river":
				riverCabinetCount = len(cab.CabinetIDs())
			case "mountain":
				mountainCabinetCount = len(cab.CabinetIDs())
			case "hill":
				hillCabinetCount = len(cab.CabinetIDs())
			}
		}

		// Prepare the network layout configs for generating the networks
		var internalNetConfigs = make(map[string]csi.NetworkLayoutConfiguration)
		internalNetConfigs["BICAN"] = csi.GenDefaultBICANConfig(v.GetString("bican-user-network-name"))
		internalNetConfigs["CMN"] = csi.GenDefaultCMNConfig(len(logicalNcns), len(switches))
		internalNetConfigs["HMN"] = csi.GenDefaultHMNConfig()
		internalNetConfigs["HSN"] = csi.GenDefaultHSNConfig()
		internalNetConfigs["MTL"] = csi.GenDefaultMTLConfig()
		internalNetConfigs["NMN"] = csi.GenDefaultNMNConfig()
		if v.GetString("bican-user-network-name") == "CAN" || v.GetBool("retain-unused-user-network") {
			internalNetConfigs["CAN"] = csi.GenDefaultCANConfig()
		}
		if v.GetString("bican-user-network-name") == "CHN" || v.GetBool("retain-unused-user-network") {
			internalNetConfigs["CHN"] = csi.GenDefaultCHNConfig()
		}

		if internalNetConfigs["HMN"].GroupNetworksByCabinetType {
			if mountainCabinetCount > 0 || hillCabinetCount > 0 {
				tmpHmnMtn := csi.GenDefaultHMNConfig()
				tmpHmnMtn.Template.Name = "HMN_MTN"
				tmpHmnMtn.Template.FullName = "Mountain Compute Hardware Management Network"
				tmpHmnMtn.Template.VlanRange = []int16{3000, 3999}
				tmpHmnMtn.Template.CIDR = csi.DefaultHMNMTNString
				tmpHmnMtn.SubdivideByCabinet = true
				tmpHmnMtn.IncludeBootstrapDHCP = false
				tmpHmnMtn.SuperNetHack = false
				tmpHmnMtn.IncludeNetworkingHardwareSubnet = false
				internalNetConfigs["HMN_MTN"] = tmpHmnMtn
			}
			if riverCabinetCount > 0 {
				tmpHmnRvr := csi.GenDefaultHMNConfig()
				tmpHmnRvr.Template.Name = "HMN_RVR"
				tmpHmnRvr.Template.FullName = "River Compute Hardware Management Network"
				tmpHmnRvr.Template.VlanRange = []int16{1513, 1769}
				tmpHmnRvr.Template.CIDR = csi.DefaultHMNRVRString
				tmpHmnRvr.SubdivideByCabinet = true
				tmpHmnRvr.IncludeBootstrapDHCP = false
				tmpHmnRvr.SuperNetHack = false
				tmpHmnRvr.IncludeNetworkingHardwareSubnet = false
				internalNetConfigs["HMN_RVR"] = tmpHmnRvr
			}

		}

		if internalNetConfigs["NMN"].GroupNetworksByCabinetType {
			if mountainCabinetCount > 0 || hillCabinetCount > 0 {
				tmpNmnMtn := csi.GenDefaultNMNConfig()
				tmpNmnMtn.Template.Name = "NMN_MTN"
				tmpNmnMtn.Template.FullName = "Mountain Compute Node Management Network"
				tmpNmnMtn.Template.VlanRange = []int16{2000, 2999}
				tmpNmnMtn.Template.CIDR = csi.DefaultNMNMTNString
				tmpNmnMtn.SubdivideByCabinet = true
				tmpNmnMtn.IncludeBootstrapDHCP = false
				tmpNmnMtn.SuperNetHack = false
				tmpNmnMtn.IncludeNetworkingHardwareSubnet = false
				tmpNmnMtn.IncludeUAISubnet = false
				internalNetConfigs["NMN_MTN"] = tmpNmnMtn
			}
			if riverCabinetCount > 0 {
				tmpNmnRvr := csi.GenDefaultNMNConfig()
				tmpNmnRvr.Template.Name = "NMN_RVR"
				tmpNmnRvr.Template.FullName = "River Compute Node Management Network"
				tmpNmnRvr.Template.VlanRange = []int16{1770, 1999}
				tmpNmnRvr.Template.CIDR = csi.DefaultNMNRVRString
				tmpNmnRvr.SubdivideByCabinet = true
				tmpNmnRvr.IncludeBootstrapDHCP = false
				tmpNmnRvr.SuperNetHack = false
				tmpNmnRvr.IncludeNetworkingHardwareSubnet = false
				tmpNmnRvr.IncludeUAISubnet = false
				internalNetConfigs["NMN_RVR"] = tmpNmnRvr
			}

		}

		for name, layout := range internalNetConfigs {
			myLayout := layout

			// Update with flags
			normalizedName := strings.ReplaceAll(strings.ToLower(name), "_", "-")

			// Use CLI values if available, otherwise defaults
			if v.IsSet(fmt.Sprintf("%v-bootstrap-vlan", normalizedName)) {
				myLayout.BaseVlan = int16(v.GetInt(fmt.Sprintf("%v-bootstrap-vlan", normalizedName)))
				myLayout.Template.VlanRange[0] = int16(v.GetInt(fmt.Sprintf("%v-bootstrap-vlan", normalizedName)))
			} else {
				myLayout.BaseVlan = layout.Template.VlanRange[0]
			}

			if v.IsSet(fmt.Sprintf("%v-cidr", normalizedName)) {
				myLayout.Template.CIDR = v.GetString(fmt.Sprintf("%v-cidr", normalizedName))
			}

			myLayout.AdditionalNetworkingSpace = v.GetInt("management-net-ips")
			internalNetConfigs[name] = myLayout
		}

		// Build a set of networks we can use
		shastaNetworks, err := csi.BuildCSMNetworks(internalNetConfigs, cabinetDetailList, switches)
		if err != nil {
			log.Panic(err)
		}

		// Use our new networks and our list of logicalNCNs to distribute ips
		AllocateIps(logicalNcns, shastaNetworks) // This function has no return because it is working with lists of pointers.

		// Now we can finally generate the slsState
		slsState := prepareAndGenerateSLS(cabinetDetailList, shastaNetworks, hmnRows, switches, applicationNodeConfig, v.GetInt("starting-mountain-nid"))
		// SLS can tell us which NCNs match with which Xnames, we need to update the IP Reservations
		slsNcns, err := csi.ExtractSLSNCNs(&slsState)
		if err != nil {
			log.Panic(err) // This should never happen.  I can't really imagine how it would.
		}

		// Merge the SLS NCN list with the NCN list we got at the beginning
		err = mergeNCNs(logicalNcns, slsNcns)
		if err != nil {
			log.Fatalln(err)
		}

		// Pull UANs from the completed slsState to assign CAN addresses
		slsUans, err := csi.ExtractUANs(&slsState)
		if err != nil {
			log.Panic(err) // This should never happen.  I can't really imagine how it would.
		}

		// Only add UANs if there actually is a CAN network
		if v.GetString("bican-user-network-name") == "CAN" || v.GetBool("retain-unused-user-network") {
			canSubnet, _ := shastaNetworks["CAN"].LookUpSubnet("bootstrap_dhcp")
			for _, uan := range slsUans {
				canSubnet.AddReservation(uan.Hostname, uan.Xname)
			}
		}
		// Only add UANs if there actually is a CHN network
		if v.GetString("bican-user-network-name") == "CHN" || v.GetBool("retain-unused-user-network") {
			chnSubnet, _ := shastaNetworks["CHN"].LookUpSubnet("bootstrap_dhcp")
			for _, uan := range slsUans {
				chnSubnet.AddReservation(uan.Hostname, uan.Xname)
			}
		}

		// Cycle through the main networks and update the reservations, masks and dhcp ranges as necessary
		for _, netName := range csi.ValidNetNames {
			if shastaNetworks[netName] != nil {
				// Grab the supernet details for use in HACK substitution
				tempSubnet, err := shastaNetworks[netName].LookUpSubnet("bootstrap_dhcp")
				if err == nil {
					// Loop the reservations and update the NCN reservations with hostnames
					// we likely didn't have when we registered the reservation
					updateReservations(tempSubnet, logicalNcns)
					if netName == "CAN" || netName == "CMN" || netName == "CHN" {
						netNameLower := strings.ToLower(netName)

						// Do not use supernet hack for the CAN/CMN/CHN
						tempSubnet.UpdateDHCPRange(false)

						myNetName := fmt.Sprintf("%s-cidr", netNameLower)
						myNetCIDR := v.GetString(myNetName)
						if myNetCIDR == "" {
							continue
						}
						_, myNet, _ := net.ParseCIDR(myNetCIDR)

						// If neither static nor dynamic pool is defined we can use the last available IP in the subnet
						poolStartIP := ipam.Broadcast(*myNet)

						// Do not overlap the static or dynamic pools
						myStaticPoolName := fmt.Sprintf("%s-static-pool", netNameLower)
						myDynPoolName := fmt.Sprintf("%s-dynamic-pool", netNameLower)

						myStaticPoolCIDR := v.GetString(myStaticPoolName)
						myDynPoolCIDR := v.GetString(myDynPoolName)

						if len(myStaticPoolCIDR) > 0 && len(myDynPoolCIDR) > 0 {
							// Both pools are defined so find the start of whichever pool comes first
							_, myStaticPool, _ := net.ParseCIDR(myStaticPoolCIDR)
							_, myDynamicPool, _ := net.ParseCIDR(myDynPoolCIDR)
							if ipam.IPLessThan(myStaticPool.IP, myDynamicPool.IP) {
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
						// range.  Here we account for it being at the end of the range.
						// Leaving this check in place for CMN because it is harmless to do so.
						if tempSubnet.Gateway.String() == ipam.Add(poolStartIP, -1).String() {
							// The gw *is* at the end, so shorten the range to accommodate
							tempSubnet.DHCPEnd = ipam.Add(poolStartIP, -2)
						} else {
							// The gw is not at the end
							tempSubnet.DHCPEnd = ipam.Add(poolStartIP, -1)
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
					updateReservations(tempSubnet, logicalNcns)
					tempSubnet.UpdateDHCPRange(false)
				}

			}
		}

		// Update the SLSState with the updated network information
		_, slsState.Networks = prepareNetworkSLS(shastaNetworks)

		// Switch from a list of pointers to a list of things before we write it out
		var ncns []csi.LogicalNCN
		for _, ncn := range logicalNcns {
			ncns = append(ncns, *ncn)
		}
		globals, err := pit.MakeBasecampGlobals(v, ncns, shastaNetworks, "NMN", "bootstrap_dhcp", v.GetString("install-ncn"))
		if err != nil {
			log.Fatalln("unable to generate basecamp globals: ", err)
		}
		writeOutput(v, shastaNetworks, slsState, ncns, switches, globals)

		// Gather SLS information for summary
		slsMountainCabinets := csi.GetSLSCabinets(slsState, sls_common.ClassMountain)
		slsHillCabinets := csi.GetSLSCabinets(slsState, sls_common.ClassHill)
		slsRiverCabinets := csi.GetSLSCabinets(slsState, sls_common.ClassRiver)

		if v.IsSet("cabinets-yaml") && (v.GetString("cabinets-yaml") != "") && (v.IsSet("mountain-cabinets") ||
			v.IsSet("starting-mountain-cabinet") ||
			v.IsSet("river-cabinets") ||
			v.IsSet("starting-river-cabinet") ||
			v.IsSet("hill-cabinets") ||
			v.IsSet("starting-hill-cabinet")) {
			fmt.Printf("\nWARNING: cabinet flags are not honored when a cabinets-yaml file is provided\n")
		}

		// Print Summary
		fmt.Printf("\n===== %v Installation Summary =====\n\n", v.GetString("system-name"))
		fmt.Printf("Installation Node: %v\n", v.GetString("install-ncn"))
		fmt.Printf("Customer Management: %v GW: %v\n", v.GetString("cmn-cidr"), v.GetString("cmn-gateway"))
		fmt.Printf("Customer Access: %v GW: %v\n", v.GetString("can-cidr"), v.GetString("can-gateway"))
		fmt.Printf("\tUpstream DNS: %v\n", v.GetString("ipv4-resolvers"))
		fmt.Printf("\tMetalLB Peers: %v\n", v.GetStringSlice("bgp-peer-types"))
		fmt.Println("Networking")
		fmt.Printf("\tBICAN user network toggle set to %v\n", v.GetString("bican-user-network-name"))
		if v.GetBool("supernet") {
			fmt.Printf("\tSupernet enabled!  Using the supernet gateway for some management subnets\n")
		}
		for _, tempNet := range shastaNetworks {
			fmt.Printf("\t* %v %v with %d subnets \n", tempNet.FullName, tempNet.CIDR, len(tempNet.Subnets))
		}
		fmt.Printf("System Information\n")
		fmt.Printf("\tNCNs: %v\n", len(ncns))
		fmt.Printf("\tMountain Compute Cabinets: %v\n", len(slsMountainCabinets))
		fmt.Printf("\tHill Compute Cabinets: %v\n", len(slsHillCabinets))
		fmt.Printf("\tRiver Compute Cabinets: %v\n", len(slsRiverCabinets))
		fmt.Printf("CSI Version Information\n\t%s\n\t%s\n\n", version.Get().GitCommit, version.Get())
	},
}

func init() {
	initCmd.DisableAutoGenTag = true

	// Flags with defaults for initializing a configuration

	// System Configuration Flags based on previous system_config.yml and networks_derived.yml
	initCmd.Flags().String("system-name", "sn-2024", "Name of the System")
	initCmd.Flags().String("csm-version", "1.0", "Version of CSM being installed")

	initCmd.Flags().String("site-domain", "", "Site Domain Name")
	initCmd.Flags().String("first-master-hostname", "ncn-m002", "Hostname of the first master node")
	// initCmd.Flags().String("internal-domain", "unicos.shasta", "Internal Domain Name")
	initCmd.Flags().String("ntp-pool", "", "Hostname for Upstream NTP Pool")
	initCmd.Flags().MarkDeprecated("ntp-pool", "please use --ntp-pools (plural) instead")
	initCmd.Flags().StringSlice("ntp-pools", []string{""}, "Comma-seperated list of upstream NTP pool(s)")
	initCmd.Flags().StringSlice("ntp-servers", []string{"ncn-m001"}, "Comma-seperated list of upstream NTP server(s) ncn-m001 should always be in this list")
	initCmd.Flags().StringSlice("ntp-peers", []string{"ncn-m001", "ncn-m002", "ncn-m003", "ncn-w001", "ncn-w002", "ncn-w003", "ncn-s001", "ncn-s002", "ncn-s003"}, "Comma-seperated list of NCNs that will peer together")
	initCmd.Flags().String("ntp-timezone", "UTC", "Timezone to be used on the NCNs and across the system")
	initCmd.Flags().String("ipv4-resolvers", "8.8.8.8, 9.9.9.9", "List of IP Addresses for DNS")
	initCmd.Flags().String("v2-registry", "https://registry.nmn/", "URL for default v2 registry used for both helm and containers")
	initCmd.Flags().String("rpm-repository", "https://packages.nmn/repository/shasta-master", "URL for default rpm repository")
	initCmd.Flags().String("cmn-gateway", "", "Gateway for NCNs on the CMN (Administrative/Management)")
	initCmd.Flags().String("can-gateway", "", "Gateway for NCNs on the CAN (User)")
	initCmd.Flags().String("chn-gateway", "", "Gateway for NCNs on the CHN (User)")
	initCmd.Flags().String("ceph-cephfs-image", "dtr.dev.cray.com/cray/cray-cephfs-provisioner:0.1.0-nautilus-1.3", "The container image for the cephfs provisioner")
	initCmd.Flags().String("ceph-rbd-image", "dtr.dev.cray.com/cray/cray-rbd-provisioner:0.1.0-nautilus-1.3", "The container image for the ceph rbd provisioner")
	initCmd.Flags().String("docker-image-registry", "dtr.dev.cray.com", "Upstream docker registry for use during the install")

	// Site Networking and Preinstall Toolkit Information
	initCmd.Flags().String("install-ncn", "ncn-m001", "Hostname of the node to be used for installation")
	initCmd.Flags().String("install-ncn-bond-members", "p1p1,p1p2", "List of devices to use to form a bond on the install ncn")
	initCmd.Flags().String("site-ip", "", "Site Network Information in the form ipaddress/prefix like 192.168.1.1/24")
	initCmd.Flags().String("site-gw", "", "Site Network IPv4 Gateway")
	initCmd.Flags().String("site-dns", "", "Site Network DNS Server which can be different from the upstream ipv4-resolvers if necessary")
	initCmd.Flags().String("site-nic", "em1", "Network Interface on install-ncn that will be connected to the site network")

	// BICAN Network Toggle
	initCmd.Flags().String("bican-user-network-name", "", "Name of the network over which non-admin users access the system [CAN, CHN, HSN]")
	initCmd.Flags().Bool("retain-unused-user-network", false, "Use the supernet mask and gateway for NCNs and Switches")

	// Default IPv4 Networks
	initCmd.Flags().String("nmn-cidr", csi.DefaultNMNString, "Overall IPv4 CIDR for all Node Management subnets")
	initCmd.Flags().String("nmn-static-pool", "", "Overall IPv4 CIDR for static Node Management load balancer addresses")
	initCmd.Flags().String("nmn-dynamic-pool", csi.DefaultNMNLBString, "Overall IPv4 CIDR for dynamic Node Management load balancer addresses")
	initCmd.Flags().String("nmn-mtn-cidr", csi.DefaultNMNMTNString, "IPv4 CIDR for grouped Mountain Node Management subnets")
	initCmd.Flags().String("nmn-rvr-cidr", csi.DefaultNMNRVRString, "IPv4 CIDR for grouped River Node Management subnets")

	initCmd.Flags().String("hmn-cidr", csi.DefaultHMNString, "Overall IPv4 CIDR for all Hardware Management subnets")
	initCmd.Flags().String("hmn-static-pool", "", "Overall IPv4 CIDR for static Hardware Management load balancer addresses")
	initCmd.Flags().String("hmn-dynamic-pool", csi.DefaultHMNLBString, "Overall IPv4 CIDR for dynamic Hardware Management load balancer addresses")
	initCmd.Flags().String("hmn-mtn-cidr", csi.DefaultHMNMTNString, "IPv4 CIDR for grouped Mountain Hardware Management subnets")
	initCmd.Flags().String("hmn-rvr-cidr", csi.DefaultHMNRVRString, "IPv4 CIDR for grouped River Hardware Management subnets")

	initCmd.Flags().String("cmn-cidr", "", "Overall IPv4 CIDR for all Customer Management subnets")
	initCmd.Flags().String("cmn-static-pool", "", "Overall IPv4 CIDR for static Customer Management load balancer addresses")
	initCmd.Flags().String("cmn-dynamic-pool", "", "Overall IPv4 CIDR for dynamic Customer Management load balancer addresses")
	initCmd.Flags().String("cmn-external-dns", "", "IP Address in the cmn-static-pool for the external dns service \"site-to-system lookups\"")
	initCmd.Flags().String("can-cidr", "", "Overall IPv4 CIDR for all Customer Access subnets")
	initCmd.Flags().String("can-static-pool", "", "Overall IPv4 CIDR for static Customer Access load balancer addresses")
	initCmd.Flags().String("can-dynamic-pool", "", "Overall IPv4 CIDR for dynamic Customer Access load balancer addresses")
	initCmd.Flags().String("chn-cidr", "", "Overall IPv4 CIDR for all Customer High-Speed subnets")
	initCmd.Flags().String("chn-static-pool", "", "Overall IPv4 CIDR for static Customer High-Speed load balancer addresses")
	initCmd.Flags().String("chn-dynamic-pool", "", "Overall IPv4 CIDR for dynamic Customer High-Speed load balancer addresses")

	initCmd.Flags().String("mtl-cidr", csi.DefaultMTLString, "Overall IPv4 CIDR for all Provisioning subnets")
	initCmd.Flags().String("hsn-cidr", csi.DefaultHSNString, "Overall IPv4 CIDR for all HSN subnets")
	initCmd.Flags().String("hsn-static-pool", "", "Overall IPv4 CIDR for static High Speed load balancer addresses")
	initCmd.Flags().String("hsn-dynamic-pool", "", "Overall IPv4 CIDR for dynamic High Speed load balancer addresses")

	initCmd.Flags().Bool("supernet", true, "Use the supernet mask and gateway for NCNs and Switches")

	// Bootstrap VLANS
	initCmd.Flags().Int("nmn-bootstrap-vlan", csi.DefaultNMNVlan, "Bootstrap VLAN for the NMN")
	initCmd.Flags().Int("hmn-bootstrap-vlan", csi.DefaultHMNVlan, "Bootstrap VLAN for the HMN")
	initCmd.Flags().Int("cmn-bootstrap-vlan", csi.DefaultCMNVlan, "Bootstrap VLAN for the CMN")
	initCmd.Flags().Int("can-bootstrap-vlan", csi.DefaultCANVlan, "Bootstrap VLAN for the CAN")

	// Hardware Details
	initCmd.Flags().Int("mountain-cabinets", 4, "Number of Mountain Cabinets") // 4 mountain cabinets per CDU
	initCmd.Flags().Int("starting-mountain-cabinet", 1000, "Starting ID number for Mountain Cabinets")

	initCmd.Flags().Int("river-cabinets", 1, "Number of River Cabinets")
	initCmd.Flags().Int("starting-river-cabinet", 3000, "Starting ID number for River Cabinets")

	initCmd.Flags().Int("hill-cabinets", 0, "Number of Hill Cabinets")
	initCmd.Flags().Int("starting-hill-cabinet", 9000, "Starting ID number for Hill Cabinets")

	initCmd.Flags().Int("starting-river-NID", 1, "Starting NID for Compute Nodes")
	initCmd.Flags().Int("starting-mountain-NID", 1000, "Starting NID for Compute Nodes")

	// Use these flags to prepare the basecamp metadata json
	initCmd.Flags().String("bgp-asn", "65533", "The autonomous system number for BGP router")
	initCmd.Flags().String("bgp-cmn-asn", "65532", "The autonomous system number for CMN BGP clients")
	initCmd.Flags().String("bgp-nmn-asn", "65531", "The autonomous system number for NMN BGP clients")
	initCmd.Flags().String("bgp-chn-asn", "65530", "The autonomous system number for CHN BGP clients")
	initCmd.Flags().String("bgp-peers", "spine", "Which set of switches to use as metallb peers, spine (default) or leaf")
	initCmd.Flags().MarkDeprecated("bgp-peers", "please use --bgp-peer-types instead")
	initCmd.Flags().StringSlice("bgp-peer-types", []string{"spine"}, "Comma separated list of which set of switches to use as metallb peers: spine (default), leaf and/or edge")
	initCmd.Flags().Int("management-net-ips", 0, "Additional number of ip addresses to reserve in each vlan for network equipment")
	initCmd.Flags().Bool("k8s-api-auditing-enabled", false, "Enable the kubernetes auditing API")
	initCmd.Flags().Bool("ncn-mgmt-node-auditing-enabled", false, "Enable management node auditing")

	// Use these flags to set the default ncn bmc credentials for bootstrap
	initCmd.Flags().String("bootstrap-ncn-bmc-user", "", "Username for connecting to the BMC on the initial NCNs")

	initCmd.Flags().String("bootstrap-ncn-bmc-pass", "", "Password for connecting to the BMC on the initial NCNs")

	// Dealing with SLS precursors
	initCmd.Flags().String("hmn-connections", "hmn_connections.json", "HMN Connections JSON Location (For generating an SLS File)")
	initCmd.Flags().String("ncn-metadata", "ncn_metadata.csv", "CSV for mapping the mac addresses of the NCNs to their xnames")
	initCmd.Flags().String("switch-metadata", "switch_metadata.csv", "CSV for mapping the switch xname, brand, type, and model")
	initCmd.Flags().String("cabinets-yaml", "", "YAML file listing the ids for all cabinets by type")
	initCmd.Flags().String("application-node-config-yaml", "", "YAML to control Application node identification durring the SLS Input File generation")

	// Loftsman Manifest Shasta-CFG
	initCmd.Flags().String("manifest-release", "", "Loftsman Manifest Release Version (leave blank to prevent manifest generation)")
	initCmd.Flags().SortFlags = false

	// DNS zone transfer settings
	initCmd.Flags().String("primary-server-name", "primary", "Desired name for the primary DNS server")
	initCmd.Flags().String("secondary-servers", "", "Comma seperated list of FQDN/IP for all DNS servers to notify when zone changes are made")
	initCmd.Flags().String("notify-zones", "", "Comma seperated list of the zones to be allowed transfer")
	initCmd.AddCommand(emptyCmd)
}

func initiailzeManifestDir(url, branch, destination string) {
	// First we need a temporary directory
	dir, err := ioutil.TempDir("", "loftsman-init")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.RemoveAll(dir)
	cloneCmd := exec.Command("git", "clone", url, dir)
	out, err := cloneCmd.Output()
	if err != nil {
		log.Fatalf("cloneCommand finished with error: %s (%v)\n", out, err)
	}
	checkoutCmd := exec.Command("git", "checkout", branch)
	checkoutCmd.Dir = dir
	out, err = checkoutCmd.Output()
	if err != nil {
		if err.Error() != "exit status 1" {
			log.Fatalf("checkoutCommand finished with error: %s (%v)\n", out, err)
		}
	}
	packageCmd := exec.Command("./package/package.sh", "1.4.0")
	packageCmd.Dir = dir
	out, err = packageCmd.Output()
	if err != nil {
		log.Fatalf("packageCommand finished with error: %s (%v)\n", out, err)
	}
	targz, _ := filepath.Abs(filepath.Clean(dir + "/dist/shasta-cfg-1.4.0.tgz"))
	untarCmd := exec.Command("tar", "-zxvvf", targz)
	untarCmd.Dir = destination
	out, err = untarCmd.Output()
	if err != nil {
		log.Fatalf("untarCmd finished with error: %s (%v)\n", out, err)
	}
}

func setupDirectories(systemName string, v *viper.Viper) (string, error) {
	// Set up the path for our base directory using our systemname
	basepath, err := filepath.Abs(filepath.Clean(systemName))
	if err != nil {
		return basepath, err
	}
	// Create our base directory
	if err = os.Mkdir(basepath, 0777); err != nil {
		return basepath, err
	}

	// These Directories make up the overall structure for the Configuration Payload
	// TODO: Refactor this out of the function and into defaults or some other config
	dirs := []string{
		filepath.Join(basepath, "manufacturing"),
		filepath.Join(basepath, "dnsmasq.d"),
		filepath.Join(basepath, "pit-files"),
		filepath.Join(basepath, "basecamp"),
	}
	// Add the Manifest directory if needed
	if v.GetString("manifest-release") != "" {
		dirs = append(dirs, filepath.Join(basepath, "loftsman-manifests"))
	}
	// Iterate through the directories and create them
	for _, dir := range dirs {
		if err := os.Mkdir(dir, 0777); err != nil {
			// log.Fatalln("Can't create directory", dir, err)
			return basepath, err
		}
	}
	return basepath, nil
}

func mergeNCNs(logicalNcns []*csi.LogicalNCN, slsNCNs []csi.LogicalNCN) error {
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
				// ncn.InstanceID = csi.GenerateInstanceID()

				found = true
				break
			}
		}

		// All NCNs from ncn-metadata need to appear in the generated SLS state
		if !found {
			return fmt.Errorf("failed to find NCN from ncn-metadata in generated SLS State: %s", ncn.Xname)
		}
	}

	return nil
}

func prepareNetworkSLS(shastaNetworks map[string]*csi.IPV4Network) ([]csi.IPV4Network, map[string]sls_common.Network) {
	// Fix up the network names & create the SLS friendly version of the shasta networks
	var networks []csi.IPV4Network
	for name, network := range shastaNetworks {
		if network.Name == "" {
			network.Name = name
		}
		networks = append(networks, *network)
	}
	return networks, convertIPV4NetworksToSLS(&networks)
}

func prepareAndGenerateSLS(cd []csi.CabinetGroupDetail, shastaNetworks map[string]*csi.IPV4Network, hmnRows []shcd_parser.HMNRow, inputSwitches []*csi.ManagementSwitch, applicationNodeConfig csi.SLSGeneratorApplicationNodeConfig, startingNid int) sls_common.SLSState {
	// Management Switch Information is included in the IP Reservations for each subnet
	switchNet, err := shastaNetworks["HMN"].LookUpSubnet("network_hardware")
	if err != nil {
		log.Fatalln("Couldn't find subnet for management switches in the HMN:", err)
	}
	reservedSwitches, _ := extractSwitchesfromReservations(switchNet)
	slsSwitches := make(map[string]sls_common.GenericHardware)
	for _, mySwitch := range reservedSwitches {
		xname := mySwitch.Xname

		// Extract Switch brand from data stored in switch_metdata.csv
		for _, inputSwitch := range inputSwitches {
			if inputSwitch.Xname == xname {
				mySwitch.Brand = inputSwitch.Brand
				break
			}
		}
		if mySwitch.Brand == "" {
			log.Fatalln("Couldn't determine switch brand for:", xname)
		}

		// Create SLS version of the switch
		slsSwitches[mySwitch.Xname], err = convertManagementSwitchToSLS(&mySwitch)
		if err != nil {
			log.Fatalln("Couldn't get SLS management switch representation:", err)
		}
	}

	// Iterate through the cabinets of each kind and build structures that work for SLS Generation
	slsCabinetMap := genCabinetMap(cd, shastaNetworks)

	// Convert shastaNetwork information to SLS Style Networking
	_, slsNetworks := prepareNetworkSLS(shastaNetworks)

	inputState := csi.SLSGeneratorInputState{
		ApplicationNodeConfig: applicationNodeConfig,
		ManagementSwitches:    slsSwitches,
		RiverCabinets:         slsCabinetMap["river"],
		HillCabinets:          slsCabinetMap["hill"],
		MountainCabinets:      slsCabinetMap["mountain"],
		MountainStartingNid:   startingNid,
		Networks:              slsNetworks,
	}

	slsState := csi.GenerateSLSState(inputState, hmnRows)
	return slsState
}

func updateReservations(tempSubnet *csi.IPV4Subnet, logicalNcns []*csi.LogicalNCN) {
	// Loop the reservations and update the NCN reservations with hostnames
	// we likely didn't have when we registered the resevation
	for index, reservation := range tempSubnet.IPReservations {
		for _, ncn := range logicalNcns {
			if reservation.Comment == ncn.Xname {
				reservation.Name = ncn.Hostname
				reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("%v-%v", ncn.Hostname, strings.ToLower(tempSubnet.NetName)))
				reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("time-%v", strings.ToLower(tempSubnet.NetName)))
				reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("time-%v.local", strings.ToLower(tempSubnet.NetName)))
				if strings.ToLower(ncn.Subrole) == "storage" && strings.ToLower(tempSubnet.NetName) == "hmn" {
					reservation.Aliases = append(reservation.Aliases, "rgw-vip.hmn")
				}
				if strings.ToLower(tempSubnet.NetName) == "nmn" {
					// The xname of a NCN will point to its NMN IP address
					reservation.Aliases = append(reservation.Aliases, ncn.Xname)
				}
				tempSubnet.IPReservations[index] = reservation
			}
			if reservation.Comment == fmt.Sprintf("%v-mgmt", ncn.Xname) {
				reservation.Comment = reservation.Name
				reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("%v-mgmt", ncn.Hostname))
				tempSubnet.IPReservations[index] = reservation
			}
		}
		if tempSubnet.NetName == "NMN" {
			reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("%v.local", reservation.Name))
			tempSubnet.IPReservations[index] = reservation
		}
	}
}

func writeOutput(v *viper.Viper, shastaNetworks map[string]*csi.IPV4Network, slsState sls_common.SLSState, logicalNCNs []csi.LogicalNCN, switches []*csi.ManagementSwitch, globals interface{}) {
	basepath, _ := setupDirectories(v.GetString("system-name"), v)
	err := csiFiles.WriteJSONConfig(filepath.Join(basepath, "sls_input_file.json"), &slsState)
	if err != nil {
		log.Fatalln("Failed to encode SLS state:", err)
	}
	v.SetConfigType("yaml")
	v.Set("VersionInfo", version.Get())
	v.WriteConfigAs(filepath.Join(basepath, "system_config.yaml"))

	csiFiles.WriteYAMLConfig(filepath.Join(basepath, "customizations.yaml"), pit.GenCustomizationsYaml(logicalNCNs, shastaNetworks, switches))

	for _, ncn := range logicalNCNs {
		// log.Println("Checking to see if we need PIT files for ", ncn.Hostname)
		if strings.HasPrefix(ncn.Hostname, v.GetString("install-ncn")) {
			log.Println("Generating Installer Node (PIT) interface configurations for:", ncn.Hostname)
			pit.WriteCPTNetworkConfig(filepath.Join(basepath, "pit-files"), v, ncn, shastaNetworks)
		}
	}
	pit.WriteDNSMasqConfig(basepath, v, logicalNCNs, shastaNetworks)
	pit.WriteConmanConfig(filepath.Join(basepath, "conman.conf"), logicalNCNs)
	pit.WriteMetalLBConfigMap(basepath, v, shastaNetworks, switches)
	pit.WriteBasecampData(filepath.Join(basepath, "basecamp/data.json"), logicalNCNs, shastaNetworks, globals)

	if v.GetString("manifest-release") != "" {
		initiailzeManifestDir(csi.DefaultManifestURL, "release/shasta-1.4", filepath.Join(basepath, "loftsman-manifests"))
	}
}

func validateFlags() []string {
	var errors []string
	v := viper.GetViper()
	expectedCSMVersion := "1.2"

	var requiredFlags = []string{
		"system-name",
		"csm-version",
		"can-gateway",
		"cmn-gateway",
		"cmn-cidr",
		"site-ip",
		"site-gw",
		"cmn-external-dns",
		"site-dns",
		"site-nic",
		"bootstrap-ncn-bmc-user",
		"bootstrap-ncn-bmc-pass",
		"bican-user-network-name",
	}

	for _, flagName := range requiredFlags {
		if !v.IsSet(flagName) || (v.GetString(flagName) == "") {
			errors = append(errors, fmt.Sprintf("%v is required and not set through flag or config file (%s)", flagName, v.ConfigFileUsed()))
		}
	}

	if v.GetString("csm-version") != expectedCSMVersion {
		errors = append(errors, fmt.Sprintf("CSI inputs must be for csm-version %s", expectedCSMVersion))
	}

	var ipv4Flags = []string{
		"site-dns",
		"cmn-gateway",
		"can-gateway",
		"site-gw",
	}
	for _, flagName := range ipv4Flags {
		if v.IsSet(flagName) {
			if net.ParseIP(v.GetString(flagName)) == nil {
				errors = append(errors, fmt.Sprintf("%v should be an ip address and is not set correctly through flag or config file (.%s)", flagName, v.ConfigFileUsed()))
			}
		}
	}

	var cidrFlags = []string{
		"cmn-cidr",
		"cmn-static-pool",
		"cmn-dynamic-pool",
		"can-cidr",
		"can-static-pool",
		"can-dynamic-pool",
		"chn-cidr",
		"chn-static-pool",
		"chn-dynamic-pool",
		"nmn-cidr",
		"hmn-cidr",
		"site-ip",
	}

	for _, flagName := range cidrFlags {
		if v.IsSet(flagName) && (v.GetString(flagName) != "") {
			_, _, err := net.ParseCIDR(v.GetString(flagName))
			if err != nil {
				errors = append(errors, fmt.Sprintf("%v should be a CIDR in the form 192.168.0.1/24 and is not set correctly through flag or config file (.%s)", flagName, v.ConfigFileUsed()))
			}
		}
	}

	validBican := false
	bicanFlag := "bican-user-network-name"
	for _, value := range [3]string{"CAN", "CHN", "HSN"} {
		if v.GetString(bicanFlag) == value {
			validBican = true
			break
		}
	}
	if !validBican {
		errors = append(errors, fmt.Sprintf("%v must be set to CAN, CHN or HSN. (HSN requires NAT device)", bicanFlag))
	}

	return errors
}

// AllocateIps distributes IP reservations for each of the NCNs within the networks
func AllocateIps(ncns []*csi.LogicalNCN, networks map[string]*csi.IPV4Network) {
	//log.Printf("I'm here in AllocateIps with %d ncns to work with and %d networks", len(ncns), len(networks))
	lookup := func(name string, subnetName string, networks map[string]*csi.IPV4Network) (*csi.IPV4Subnet, error) {
		tempNetwork := networks[name]
		subnet, err := tempNetwork.LookUpSubnet(subnetName)
		if err != nil {
			// log.Printf("couldn't find a %v subnet in the %v network \n", subnetName, name)
			return subnet, fmt.Errorf("couldn't find a %v subnet in the %v network", subnetName, name)
		}
		// log.Printf("found a %v subnet in the %v network", subnetName, name)
		return subnet, nil
	}

	// Build a map of networks based on their names
	subnets := make(map[string]*csi.IPV4Subnet)
	for name := range networks {
		bootstrapNet, err := lookup(name, "bootstrap_dhcp", networks)
		if err == nil {
			subnets[name] = bootstrapNet
		}
	}

	// Loop through the NCNs and then run through the networks to add reservations and assign ip addresses
	for _, ncn := range ncns {
		ncn.InstanceID = csi.GenerateInstanceID()
		for netName, subnet := range subnets {
			// reserve the bmc ip
			if netName == "HMN" {
				// The bmc xname is the ncn xname without the final two characters
				// NCN Xname = x3000c0s9b0n0  BMC Xname = x3000c0s9b0
				ncn.BmcIP = subnet.AddReservation(fmt.Sprintf("%v", strings.TrimSuffix(ncn.Xname, "n0")), fmt.Sprintf("%v-mgmt", ncn.Xname)).IPAddress.String()
			}
			// Hostname is not available a the point AllocateIPs should be called.
			reservation := subnet.AddReservation(ncn.Xname, ncn.Xname)
			//log.Printf("Adding %v %v reservation for %v(%v) at %v \n", netName, subnet.Name, ncn.Xname, ncn.Xname, reservation.IPAddress.String())
			prefixLen := strings.Split(subnet.CIDR.String(), "/")[1]
			tempNetwork := csi.NCNNetwork{
				NetworkName: netName,
				IPAddress:   reservation.IPAddress.String(),
				Vlan:        int(subnet.VlanID),
				FullName:    subnet.FullName,
				CIDR:        strings.Join([]string{reservation.IPAddress.String(), prefixLen}, "/"),
				Mask:        prefixLen,
			}
			ncn.Networks = append(ncn.Networks, tempNetwork)

		}
	}
}
