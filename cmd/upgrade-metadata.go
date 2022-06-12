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
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/Cray-HPE/cray-site-init/pkg/bss"
	"github.com/Cray-HPE/cray-site-init/pkg/sls"
	base "github.com/Cray-HPE/hms-base"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

/*
UpgradeParamsToDelete Linux Kernel Parameters to-remove for upgrades to 1.2.
	1. bond

		Remove any bond settings; a more simple bond config is passed-in with UpgradeParamsToSet.

	2. bootdev

		This is not necessary to have anymore, it has no effect.

	3. hwprobe

		Previously thought to ensure the bond comes up, this actually has no effect.

	4. ip

		We do need one of these to exist, so we set it in UpgradeParamsToSet. This ensures we purge all undesirables and start fresh.

	5. rd.peerdns

		Cease and desist any unofficial DNS from "peers" and use what the DHCP Authority has given us.

	6. vlan

		VLANs are now entirely driven by CSI and cloud-init, thus they no longer exist as hardcodes in kernel parameters. These must be removed to prevent
		conflicts with any dynamic VLAN changes. These also provide incorrect interface names, relative to CSM 1.2 the VLAN interface names have
		changed.

*/
var UpgradeParamsToDelete = []string{
	"bond",
	"bootdev",
	"hwprobe",
	"ip",
	"rd.peerdns",
	"rd.net.dhcp.retry",
	"vlan",
}

/*
UpgradeParamsToSet Linux Kernel Parameters to-set for upgrades to 1.2.

	1. ip=mgmt0:dhcp

		DHCP must be setup. The bond will be formed during cloud-init

	2. rd.peerdns=0

		Set to prevent inconsistent DNS resolution (e.g. always use the actual Domain Name Server and not each other).

	3. rd.net.dhcp.retry=5

		Set to ensure we wait out STP blocks, at least giving them more chances than 3 (from 1.0).

*/
var UpgradeParamsToSet = []paramTuple{{
	key:   "ip",
	value: "mgmt0:dhcp",
}, {
	key:   "rd.peerdns",
	value: "0",
}, {
	key:   "rd.net.dhcp.retry",
	value: "5",
}}

/*
UpgradeParamsToAdd Linux Kernel Parameters to-add for upgrades to 1.2.
This is for multiple parameters that need to be added with the same key.
e.g. ifname=mgmt0:01:02:03:04:05:06


        1. ifnames
		Ifnames have changed on nodes that have two NICs.

				1.0	-->	1.2
		-----------------------------------
		OCP port1	mgmt0		mgmt0
		OCP port2	mgmt1		sun0
		PCIe port1	mgmt2		mgmt1
		PCIe port2	mgmt3		sun1

*/
var UpgradeParamsToAdd = []paramTuple{}

func getIPAMForNCN(managementNCN sls_common.GenericHardware,
	networks sls_common.NetworkArray, extraSLSNetworks ...string) (ipamNetworks bss.CloudInitIPAM) {
	ipamNetworks = make(bss.CloudInitIPAM)

	// For each of the required networks, go build an IPAMNetwork object and add that to the ipamNetworks
	// above.
	for _, ipamNetwork := range append(bss.IPAMNetworks[:], extraSLSNetworks...) {
		// Search SLS networks for this network.
		var targetSLSNetwork *sls_common.Network
		for _, slsNetwork := range networks {
			if strings.ToLower(slsNetwork.Name) == ipamNetwork {
				targetSLSNetwork = &slsNetwork
				break
			}
		}

		if targetSLSNetwork == nil {
			log.Fatalf("Failed to find required IPAM network [%s] in SLS networks!", ipamNetwork)
		}

		// Map this network to a usable structure.
		var networkExtraProperties sls.NetworkExtraProperties
		err := mapstructure.Decode(targetSLSNetwork.ExtraPropertiesRaw, &networkExtraProperties)
		if err != nil {
			log.Fatalf("Failed to decode raw network extra properties to correct structure: %s", err)
		}

		// The target SLS network is determined, now we need the right reservation.
		var targetSubnet *sls.IPV4Subnet
		var targetReservation *sls.IPReservation

		_, targetNet, err := net.ParseCIDR(networkExtraProperties.CIDR)
		if err != nil {
			log.Fatalf("Failed to parse SLS network CIDR (%s): %s", networkExtraProperties.CIDR, err)
		}

		for _, subnet := range networkExtraProperties.Subnets {
			for _, reservation := range subnet.IPReservations {
				if reservation.Comment == managementNCN.Xname {
					targetSubnet = &subnet
					targetReservation = &reservation

					break
				}
			}

			if targetSubnet != nil {
				break
			}
		}

		if targetSubnet == nil || targetReservation == nil {
			log.Fatalf("Failed to find subnet/reservation for this managment NCN xname (%s)!",
				managementNCN.Xname)
		}

		// Finally, we have all the pieces, wrangle the data! Speaking of, here's an example of what this
		// should look like after we're done:
		//  "can": {
		//   "gateway": "10.10.0.1",
		//   "ip": "10.10.0.1/26",
		//   "parent_device": "bond0",
		//   "vlanid": 999
		//  }
		// Few things to note, the `ip` field is a bit of a misnomer as it also must include the mask bits.
		// Which is our first step, figure out just the mask bits from the subnet's CIDR. One might be
		// tempted to just use string splits and take the 1st element, but we get validation for free this
		// way.

		_, ipv4Net, err := net.ParseCIDR(targetSubnet.CIDR)
		if err != nil {
			log.Fatalf("Failed to parse SLS network CIDR (%s): %s", targetSubnet.CIDR, err)
		}

		var maskBits int
		if ipamNetwork == "cmn" || ipamNetwork == "can" {
			maskBits, _ = targetNet.Mask.Size()
		} else {
			maskBits, _ = ipv4Net.Mask.Size()
		}

		// Now we can build an IPAM object.
		thisIPAMNetwork := bss.IPAMNetwork{
			Gateway:      targetSubnet.Gateway,
			CIDR:         fmt.Sprintf("%s/%d", targetReservation.IPAddress, maskBits),
			ParentDevice: "bond0", // FIXME: Remove bond0 hardcode.
			VlanID:       targetSubnet.VlanID,
		}
		ipamNetworks[ipamNetwork] = thisIPAMNetwork
	}

	return
}

func getWriteFiles(networks sls_common.NetworkArray, ipamNetworks bss.CloudInitIPAM) (writeFiles []bss.WriteFile) {
	// In the case of 1.0 -> 1.2 we need to add route files for a few of the networks.
	// The process is simple, get the CIDR and gateway for those networks and then format them as an ifroute file.
	// Here's an example:
	routeFiles := make(map[string][]string)

	for _, neededNetwork := range []string{"cmn", "hmn", "nmn"} {
		ipamNetwork := ipamNetworks[neededNetwork]

		for _, network := range networks {
			// We have to check the prefix of the network because some networks like the NMN or HMN also have LB
			// networks associated with them. Debatable whether that actually make them a separate network
			// or not but the important point here is they have to be added to the correct file.
			// We also need to add a route for the MTL network to the NMN gateway
			if strings.HasPrefix(strings.ToLower(network.Name), neededNetwork) ||
				(neededNetwork == "nmn" && strings.ToLower(network.Name) == "mtl") {
				// Map this network to a usable structure.
				var networkExtraProperties sls.NetworkExtraProperties
				err := mapstructure.Decode(network.ExtraPropertiesRaw, &networkExtraProperties)
				if err != nil {
					log.Fatalf("Failed to decode raw network extra properties to correct structure: %s", err)
				}

				thisRouteFile := routeFiles[neededNetwork]

				// Now we know we need to add this network, go through all the subnets and build up the route file.
				for _, subnet := range networkExtraProperties.Subnets {
					_, ipv4Net, err := net.ParseCIDR(subnet.CIDR)
					if err != nil {
						log.Fatalf("Failed to parse SLS network CIDR (%s): %s", subnet.CIDR, err)
					}

					// If the gateway fits in the CIDR then we don't need it, the OS will give us that for free.
					gatewayIP := net.ParseIP(ipamNetwork.Gateway)
					if ipv4Net.Contains(gatewayIP) {
						continue
					}

					route := fmt.Sprintf("%s %s - %s.%s0",
						ipv4Net.String(), gatewayIP.String(), "bond0", neededNetwork)

					// Don't add the route if we already have it
					found := false
					for _, a := range thisRouteFile {
						if a == route {
							found = true
							break
						}
					}

					if !found {
						thisRouteFile = append(thisRouteFile, route)
					}
				}

				routeFiles[neededNetwork] = thisRouteFile
			}
		}
	}

	// We now have all the write files, let's make objects for them.
	for networkName, routeFile := range routeFiles {
		writeFile := bss.WriteFile{
			Content:     strings.Join(routeFile, "\n"),
			Owner:       "root:root",
			Path:        fmt.Sprintf("/etc/sysconfig/network/ifroute-bond0.%s0", networkName),
			Permissions: "0644",
		}
		writeFiles = append(writeFiles, writeFile)
	}

	return
}

// buildBSSHostRecords will build a BSS HostRecords
func buildBSSHostRecords(networkEPs map[string]*sls.NetworkExtraProperties, networkName, subnetName, reservationName string, aliases []string) bss.HostRecord {
	subnet, err := networkEPs[networkName].LookupSubnet(subnetName)
	if err != nil {
		log.Fatalf("Unable to find %s in the %s network", subnetName, networkName)
	}
	ipReservation, found := subnet.ReservationsByName()[reservationName]
	if !found {
		log.Fatalf("Failed to find IP reservation for %s in the %s %s subnet", reservationName, networkName, subnetName)
	}

	return bss.HostRecord{
		IP:      ipReservation.IPAddress,
		Aliases: aliases,
	}
}

// getBSSGlobalHostRecords is the BSS analog of the pit.MakeBasecampHostRecords that works with SLS data
func getBSSGlobalHostRecords(managementNCNs []sls_common.GenericHardware, networks sls_common.NetworkArray) bss.HostRecords {

	// Collase all of the Network ExtraProperties into single map for lookups
	networkEPs := map[string]*sls.NetworkExtraProperties{}
	for _, network := range networks {
		// Map this network to a usable structure.
		var networkExtraProperties sls.NetworkExtraProperties
		err := mapstructure.Decode(network.ExtraPropertiesRaw, &networkExtraProperties)
		if err != nil {
			log.Fatalf("Failed to decode raw network extra properties to correct structure: %s", err)
		}

		networkEPs[network.Name] = &networkExtraProperties
	}

	var globalHostRecords bss.HostRecords

	// Add the NCN Interfaces.
	for _, managementNCN := range managementNCNs {
		var ncnExtraProperties sls_common.ComptypeNode
		err := mapstructure.Decode(managementNCN.ExtraPropertiesRaw, &ncnExtraProperties)
		if err != nil {
			log.Fatalf("Failed to decode raw NCN extra properties to correct structure: %s", err)
		}

		if len(ncnExtraProperties.Aliases) == 0 {
			log.Fatalf("NCN has no aliases defined in SLS: %+v", managementNCN)
		}

		ncnAlias := ncnExtraProperties.Aliases[0]

		// Add the NCN interface host records.
		var ipamNetworks bss.CloudInitIPAM
		extraNets := []string{}

		if _, ok := networkEPs["CHN"]; ok {
			extraNets = append(extraNets, "chn")
		}
		if _, ok := networkEPs["CAN"]; ok {
			extraNets = append(extraNets, "can")
		}

		if len(extraNets) == 0 {
			log.Fatalf("SLS must have either CAN or CHN defined")
		}
		ipamNetworks = getIPAMForNCN(managementNCN, networks, extraNets...)

		for network, ipam := range ipamNetworks {
			// Get the IP of the NCN for this network.
			ip, _, err := net.ParseCIDR(ipam.CIDR)
			if err != nil {
				log.Fatalf("Failed to parse BSS IPAM Network CIDR (%s): %s", ipam.CIDR, err)
			}

			hostRecord := bss.HostRecord{
				IP:      ip.String(),
				Aliases: []string{fmt.Sprintf("%s.%s", ncnAlias, network)},
			}

			// The NMN network gets the privilege of also containing the bare NCN Alias without network domain.
			if strings.ToLower(network) == "nmn" {
				hostRecord.Aliases = append(hostRecord.Aliases, ncnAlias)
			}
			globalHostRecords = append(globalHostRecords, hostRecord)
		}

		// Next add the NCN BMC host record
		bmcXname := base.GetHMSCompParent(managementNCN.Xname)
		globalHostRecords = append(globalHostRecords,
			buildBSSHostRecords(networkEPs, "HMN", "bootstrap_dhcp", bmcXname, []string{fmt.Sprintf("%s-mgmt", ncnAlias)}),
		)
	}

	// Add kubeapi-vip
	globalHostRecords = append(globalHostRecords,
		buildBSSHostRecords(networkEPs, "NMN", "bootstrap_dhcp", "kubeapi-vip", []string{"kubeapi-vip", "kubeapi-vip.nmn"}),
	)

	// Add rgw-vip
	globalHostRecords = append(globalHostRecords,
		buildBSSHostRecords(networkEPs, "NMN", "bootstrap_dhcp", "rgw-vip", []string{"rgw-vip", "rgw-vip.nmn"}),
	)

	// Using the original InstallNCN as the host for pit.nmn
	// HACK, I'm assuming ncn-m001
	globalHostRecords = append(globalHostRecords,
		buildBSSHostRecords(networkEPs, "NMN", "bootstrap_dhcp", "ncn-m001", []string{"pit", "pit.nmn"}),
	)

	// Add in packages.local and registry.local pointing toward the API Gateway
	globalHostRecords = append(globalHostRecords,
		buildBSSHostRecords(networkEPs, "NMNLB", "nmn_metallb_address_pool", "istio-ingressgateway", []string{"packages.local", "registry.local"}),
	)

	// Add entries for switches
	hmnNetSubnet, err := networkEPs["HMN"].LookupSubnet("network_hardware")
	if err != nil {
		log.Fatal("Unable to find network_hardware in the HMN network")
	}

	for _, ipReservation := range hmnNetSubnet.IPReservations {
		if strings.HasPrefix(ipReservation.Name, "sw-") {
			globalHostRecords = append(globalHostRecords, bss.HostRecord{
				IP:      ipReservation.IPAddress,
				Aliases: []string{ipReservation.Name},
			})
		}
	}

	return globalHostRecords
}

var (
	oneToOneTwo bool

	k8sVersion     string
	storageVersion string
)

// updateBSS pushes the changes to BSS.
func updateBSS() (err error) {
	// Instead of hammering SLS some number of times for each NCN/network combination we just grab the entire
	// network block and will later pull out the pieces we need.
	networks, err := slsClient.GetNetworks()
	if err != nil {
		log.Fatalln(err)
	}

	// Now we can loop through all the NCNs and update their metadata in BSS.
	for _, managementNCN := range managementNCNs {
		bootparameters := getBSSBootparametersForXname(managementNCN.Xname)

		var ncnExtraProperties sls_common.ComptypeNode
		err = mapstructure.Decode(managementNCN.ExtraPropertiesRaw, &ncnExtraProperties)
		if err != nil {
			log.Fatalf("Failed to decode raw NCN extra properties to correct structure: %s", err)
		}

		if len(ncnExtraProperties.Aliases) == 0 {
			err = fmt.Errorf("NCN has no aliases defined in SLS: %+v", managementNCN)
			log.Fatalf("NCN has no aliases defined in SLS: %+v", managementNCN)
		}

		/*
		 * Specific to 1.2 we have several structures that need to be created or be changed. We want to make
		 * sure we do this in an idempotent way so that if necessary this logic can be run to ensure at least
		 * these settings are correct. That's a long way of saying build everything fresh and then update the
		 * main structure to have that fresh data.
		 */

		extraNets := []string{}
		var foundCAN = false
		var foundCHN = false

		for _, net := range networks {
			if strings.ToLower(net.Name) == "can" {
				extraNets = append(extraNets, "can")
				foundCAN = true
			}
			if strings.ToLower(net.Name) == "chn" {
				foundCHN = true
			}
		}
		if !foundCAN && !foundCHN {
			err = fmt.Errorf("No CAN or CHN network defined in SLS")
			return
		}

		// IPAM
		ipamNetworks := getIPAMForNCN(managementNCN, networks, extraNets...)
		bootparameters.CloudInit.MetaData["ipam"] = ipamNetworks

		// Params
		params := strings.Split(bootparameters.Params, " ")

		var ifnameMaps = make(map[string]string)
		ifnameMaps["mgmt0"] = "mgmt0"
		ifnameMaps["mgmt1"] = "sun0"
		ifnameMaps["mgmt2"] = "mgmt1"
		ifnameMaps["mgmt3"] = "sun1"

		// Run-cmd
		switch ncnExtraProperties.SubRole {
		case "Storage":
			bootparameters.CloudInit.UserData["runcmd"] = bss.StorageNCNRunCMD
			if storageVersion != "" {
				bootparameters.Initrd = fmt.Sprintf("s3://ncn-images/ceph/%s/initrd", storageVersion)
				bootparameters.Kernel = fmt.Sprintf("s3://ncn-images/ceph/%s/kernel", storageVersion)
				setMetalServerParam := paramTuple{
					key:   "metal.server",
					value: fmt.Sprintf("http://rgw-vip.nmn/ncn-images/ceph/%s", storageVersion),
				}
				UpgradeParamsToSet = append(UpgradeParamsToSet, setMetalServerParam)
			}
			if strings.Contains(bootparameters.Params, "mgmt3") {
				// This is a storage node with 2 NICs.  Change the ifnames.
				UpgradeParamsToDelete = append(UpgradeParamsToDelete, "ifname")

				for _, param := range params {
					paramSplit := strings.Split(param, "=")
					for oldifname, newifname := range ifnameMaps {
						if strings.Contains(param, fmt.Sprintf("ifname=%s", oldifname)) {
							newVal := strings.Replace(paramSplit[1], oldifname, newifname, 1)
							addParam := paramTuple{
								key:   paramSplit[0],
								value: newVal,
							}
							UpgradeParamsToAdd = append(UpgradeParamsToAdd, addParam)
						}
					}
				}
			}
		case "Master", "Worker":
			bootparameters.CloudInit.UserData["runcmd"] = bss.KubernetesNCNRunCMD
			if k8sVersion != "" {
				bootparameters.Initrd = fmt.Sprintf("s3://ncn-images/k8s/%s/initrd", k8sVersion)
				bootparameters.Kernel = fmt.Sprintf("s3://ncn-images/k8s/%s/kernel", k8sVersion)
				setMetalServerParam := paramTuple{
					key:   "metal.server",
					value: fmt.Sprintf("http://rgw-vip.nmn/ncn-images/k8s/%s", k8sVersion),
				}
				UpgradeParamsToSet = append(UpgradeParamsToSet, setMetalServerParam)
			}
			if strings.Contains(bootparameters.Params, "mgmt3") {
				// This is a master node with 2 NICs.  Change the ifnames.
				UpgradeParamsToDelete = append(UpgradeParamsToDelete, "ifname")

				for _, param := range params {
					paramSplit := strings.Split(param, "=")
					for oldifname, newifname := range ifnameMaps {
						if strings.Contains(param, fmt.Sprintf("ifname=%s", oldifname)) {
							newVal := strings.Replace(paramSplit[1], oldifname, newifname, 1)
							addParam := paramTuple{
								key:   paramSplit[0],
								value: newVal,
							}
							UpgradeParamsToAdd = append(UpgradeParamsToAdd, addParam)
						}
					}
				}
			}
		default:
			err = fmt.Errorf("NCN has invalid SubRole: %+v", managementNCN)
			log.Fatalf("NCN has invalid SubRole: %+v", managementNCN)
		}

		finalParams := updateParams(params, UpgradeParamsToSet, UpgradeParamsToAdd, UpgradeParamsToDelete)
		bootparameters.Params = strings.Join(finalParams, " ")

		// Reset UpgradeParamsToAdd and UpgradeParamsToDelete
		UpgradeParamsToAdd = []paramTuple{}
		for k, v := range UpgradeParamsToDelete {
			if v == "ifname" {
				UpgradeParamsToDelete = append(UpgradeParamsToDelete[:k], UpgradeParamsToDelete[k+1:]...)
				break
			}
		}

		// Write files
		bootparameters.CloudInit.UserData["write_files"] = getWriteFiles(networks, ipamNetworks)

		uploadEntryToBSS(bootparameters, http.MethodPatch)
	}

	// Update the Global BSS Metadata
	globalBootParameters := getBSSBootparametersForXname("Global")
	globalBootParameters.CloudInit.MetaData["host_records"] = getBSSGlobalHostRecords(managementNCNs, networks)

	// Remove can-gw, and can-if
	delete(globalBootParameters.CloudInit.MetaData, "can-gw")
	delete(globalBootParameters.CloudInit.MetaData, "can-if")

	uploadEntryToBSS(globalBootParameters, http.MethodPatch)

	return
}

// metadataCmd represents the upgrade command.
var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Upgrades metadata",
	Long:  "Upgrades cloud-init metadata and pushes it to BSS",
	Run: func(cmd *cobra.Command, args []string) {
		if oneToOneTwo {
			setupCommon()

			updateBSS()
		}
	},
}

func init() {
	upgradeCmd.AddCommand(metadataCmd)
	metadataCmd.DisableAutoGenTag = true

	metadataCmd.Flags().SortFlags = true
	metadataCmd.Flags().BoolVarP(&oneToOneTwo, "1-0-to-1-2", "", false,
		"Upgrade CSM 1.0 metadata to 1.2 metadata")

	metadataCmd.Flags().StringVar(&k8sVersion, "k8s-version", "",
		"K8s nodes image version")
	metadataCmd.Flags().StringVar(&storageVersion, "storage-version", "",
		"Storage nodes image version")
}
