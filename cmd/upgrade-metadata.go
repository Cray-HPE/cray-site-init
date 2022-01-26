package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/Cray-HPE/cray-site-init/pkg/bss"
	"github.com/Cray-HPE/cray-site-init/pkg/sls"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

// Constants

var OneToOneTwo_ParamsToDelete = []string{"bond", "bootdev", "hwprobe", "ip", "vlan"}
var OneToOneTwo_ParamsToSet = []paramTuple{{
	key:   "ip",
	value: "mgmt0:dhcp",
}}
var OneToOneTwo_RoutesFilesToWrite = []string{"cmn", "hmn", "nmn"}

const bondedInterfaceName = "bond0"

var (
	oneToOneTwo bool

	k8sKernel string
	k8sInitrd string

	storageKernel string
	storageInitrd string
)

func getIPAMForNCN(managementNCN sls_common.GenericHardware,
	networks sls_common.NetworkArray) (ipamNetworks bss.CloudInitIPAM) {
	ipamNetworks = make(bss.CloudInitIPAM)

	// For each of the required networks, go build an IPAMNetwork object and add that to the ipamNetworks
	// above.
	for _, ipamNetwork := range bss.IPAMNetworks {
		// Search SLS networks for this network.
		var targetSLSNetwork *sls_common.Network
		for _, slsNetwork := range networks {
			if strings.ToLower(slsNetwork.Name) == ipamNetwork {
				targetSLSNetwork = &slsNetwork
				break
			}
		}

		if targetSLSNetwork == nil {
			log.Fatalf("Failed to find required IPAM network %s in SLS networks!", ipamNetwork)
		}

		// Map this network to a usable structure.
		var networkExtraProperties sls.NetworkExtraProperties
		err := mapstructure.Decode(targetSLSNetwork.ExtraPropertiesRaw, &networkExtraProperties)
		if err != nil {
			log.Fatalf("Failed to decode raw network extra properties to correct structure: %s", err)
		}

		// Now that we have the target SLS network we just need to find the right reservation and pull the
		// details we need out of that.
		var targetSubnet *sls.IPV4Subnet
		var targetReservation *sls.IPReservation

		for _, subnet := range networkExtraProperties.Subnets {
			for _, reservation := range subnet.IPReservations {
				// Yeah, this is as strange as it looks...convention is to put the xname in the comment
				// field. ¯\_(ツ)_/¯
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
		//   "gateway": "10.103.0.129",
		//   "ip": "10.103.0.142/26",
		//   "parent_device": "bond0",
		//   "vlanid": 6
		//  }
		// Few things to note, the `ip` field is a bit of a misnomer as it also must include the mask bits.
		// Which is our first step, figure out just the mask bits from the subnet's CIDR. One might be
		// tempted to just use string splits and take the 1th element, but we get validation for free this
		// way.

		_, ipv4Net, err := net.ParseCIDR(targetSubnet.CIDR)
		if err != nil {
			log.Fatalf("Failed to parse SLS network CIDR (%s): %s", targetSubnet.CIDR, err)
		}

		maskBits, _ := ipv4Net.Mask.Size()

		// Now we can build an IPAM object.
		thisIPAMNetwork := bss.IPAMNetwork{
			Gateway:      targetSubnet.Gateway,
			CIDR:         fmt.Sprintf("%s/%d", targetReservation.IPAddress, maskBits),
			ParentDevice: bondedInterfaceName,
			VlanID:       targetSubnet.VlanID,
		}

		// ...and finally add it to the main returned object.
		ipamNetworks[ipamNetwork] = thisIPAMNetwork
	}

	return
}

func getWriteFiles(networks sls_common.NetworkArray, ipamNetworks bss.CloudInitIPAM) (writeFiles []bss.WriteFile) {
	// In the case of 1.0 -> 1.2 we need to add route files for a few of the networks.
	// The process is simple, get the CIDR and gateway for those networks and then format them as an ifroute file.
	// Here's an example:
	//  10.100.0.0/22 10.252.0.1 - bond0.nmn0
	//  10.106.0.0/22 10.252.0.1 - bond0.nmn0
	//  10.1.0.0/16 10.252.0.1 - bond0.nmn0
	//  10.92.100.0/24 10.252.0.1 - bond0.nmn0
	routeFiles := make(map[string][]string)

	for _, neededNetwork := range OneToOneTwo_RoutesFilesToWrite {
		ipamNetwork := ipamNetworks[neededNetwork]

		for _, network := range networks {
			// We have to check the prefix of the network because some networks like the CMN or HMN also have LB
			// networks associated with them. Debatable whether that actually make them a separate network
			// or not but the important point here is they have to be added to the correct file.
			if strings.HasPrefix(strings.ToLower(network.Name), neededNetwork) {
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
						ipv4Net.String(), gatewayIP.String(), bondedInterfaceName, neededNetwork)

					thisRouteFile = append(thisRouteFile, route)
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

func updateBSS_oneToOneTwo() {
	// Instead of hammering SLS some number of times for each NCN/network combination we just grab the entire
	// network block and will later pull out the pieces we need.
	networks, err := slsClient.GetNetworks()
	if err != nil {
		log.Fatalln(err)
	}

	// Grab the Global BSS Metadata
	globalBootParameters := getBSSBootparametersForXname("Global")
	var globalHostRecords bss.HostRecords
	err = mapstructure.Decode(globalBootParameters.CloudInit.MetaData["host_records"], &globalHostRecords)
	if err != nil {
		log.Fatalf("Failed to decode raw NCN extra properties to correct structure: %s", err)
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
			log.Fatalf("NCN has no aliases defined in SLS: %+v", managementNCN)
		}

		/*
		 * Specific to 1.2 we have several structures that need to be created or be changed. We want to make
		 * sure we do this in an idempotent way so that if necessary this logic can be run to ensure at least
		 * these settings are correct. That's a long way of saying build everything fresh and then update the
		 * main structure to have that fresh data.
		 */

		// IPAM
		ipamNetworks := getIPAMForNCN(managementNCN, networks)
		bootparameters.CloudInit.MetaData["ipam"] = ipamNetworks

		// Run-cmd
		switch ncnExtraProperties.SubRole {
		case "Storage":
			bootparameters.CloudInit.UserData["runcmd"] = bss.StorageNCNRunCMD
			if storageInitrd != "" {
				bootparameters.Initrd = storageInitrd
			}
			if storageKernel != "" {
				bootparameters.Kernel = storageKernel
			}
		case "Master", "Worker":
			bootparameters.CloudInit.UserData["runcmd"] = bss.KubernetesNCNRunCMD
			if k8sInitrd != "" {
				bootparameters.Initrd = k8sInitrd
			}
			if k8sKernel != "" {
				bootparameters.Kernel = k8sKernel
			}
		default:
			log.Fatalf("NCN has invalid SubRole: %+v", managementNCN)
		}

		// Params
		params := strings.Split(bootparameters.Params, " ")
		finalParams := updateParams(params, OneToOneTwo_ParamsToSet, OneToOneTwo_ParamsToDelete)
		bootparameters.Params = strings.Join(finalParams, " ")

		// Write files
		bootparameters.CloudInit.UserData["write_files"] = getWriteFiles(networks, ipamNetworks)

		uploadEntryToBSS(bootparameters, http.MethodPatch)

		// Update global host records
		for network, ipam := range ipamNetworks {
			expectedNCNAlias := fmt.Sprintf("%s.%s", ncnExtraProperties.Aliases[0], network)

			// Found a matching host record, time to update its IP address
			ip, _, err := net.ParseCIDR(ipam.CIDR)
			if err != nil {
				log.Fatalf("Failed to parse BSS IPAM Network CIDR (%s): %s", ipam.CIDR, err)
			}

			// First check to see if an there is already an existing host record for this interface
			updatedHostRecord := false
			for i := range globalHostRecords {
				if globalHostRecords[i].Aliases[0] != expectedNCNAlias {
					continue
				}

				globalHostRecords[i].IP = ip.String()
				updatedHostRecord = true
				break
			}

			// Next, add a new host record for interfaces that are new for CSM 1.2, if they haven't been added already.
			if !updatedHostRecord {
				hostRecord := bss.HostRecord{
					IP:      ip.String(),
					Aliases: []string{expectedNCNAlias},
				}
				globalHostRecords = append(globalHostRecords, hostRecord)
			}
		}
	}

	// Find istio-ingressgateway IP reverstation
	istioIngressGatewayReservation := sls.GetIPReservation(networks, "NMNLB", "nmn_metallb_address_pool", "istio-ingressgateway")
	if !globalHostRecords.AliasExists("packages.local") {
		globalHostRecords = append(globalHostRecords, bss.HostRecord{
			IP: istioIngressGatewayReservation.IPAddress,
			Aliases: []string{
				"packages.local",
				"registry.local",
			},
		})
	}

	// Update the Global BSS Metadata
	globalBootParameters.CloudInit.MetaData["host_records"] = globalHostRecords
	uploadEntryToBSS(globalBootParameters, http.MethodPatch)
}

// metadataCmd represents the upgrade command
var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Upgrades metadata",
	Long:  "Upgrades cloud-init metadata and pushes it to BSS",
	Run: func(cmd *cobra.Command, args []string) {
		if oneToOneTwo {
			setupCommon()

			updateBSS_oneToOneTwo()
		}
	},
}

func init() {
	upgradeCmd.AddCommand(metadataCmd)
	metadataCmd.DisableAutoGenTag = true

	metadataCmd.Flags().SortFlags = true
	metadataCmd.Flags().BoolVarP(&oneToOneTwo, "1-0-to-1-2", "", false,
		"Upgrade CSM 1.0 metadata to 1.2 metadata")

	metadataCmd.Flags().StringVar(&k8sKernel, "k8s-kernel", "",
		"Path to the kernel image to upload for K8s NCNs")
	metadataCmd.Flags().StringVar(&k8sInitrd, "k8s-initrd", "",
		"Path to the initrd image to upload for K8s NCNs")

	metadataCmd.Flags().StringVar(&storageKernel, "storage-kernel", "",
		"Path to the kernel image to upload for Storage NCNs")
	metadataCmd.Flags().StringVar(&storageInitrd, "storage-initrd", "",
		"Path to the initrd image to upload for Storage NCNs")
}
