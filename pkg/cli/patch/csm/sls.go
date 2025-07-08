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
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"slices"
	"sort"

	"github.com/Cray-HPE/cray-site-init/pkg/csm"
	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
	slsClient "github.com/Cray-HPE/hms-sls/v2/pkg/sls-client"
	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/hashicorp/go-retryablehttp"
)

func getSLSClient(client *slsClient.SLSClient) (err error) {
	token, err := csm.GetToken()
	if err != nil {
		err = fmt.Errorf(
			"could not communicate with CSM, failed to fetch the API token because %v",
			err,
		)
		return err
	}
	*client = *slsClient.NewSLSClient(
		sls.GetSLSBaseURL(),
		retryablehttp.NewClient().StandardClient(),
		"",
	).WithAPIToken(token)
	return err
}

func getSLSData() (slsDump slsCommon.SLSState, err error) {
	slsDump, err = slsSession.GetDumpState(context.Background())
	if err != nil {
		err = fmt.Errorf(
			"failed to dump SLS because %v",
			err,
		)
	}

	return slsDump, err
}

func addIPv6SLSNetworks(slsNetworks *map[string]slsCommon.Network, updatedSLSNetworks *[]slsCommon.Network) (err error) {
	for _, network := range *slsNetworks {
		if slices.Contains(
			networksToPatch,
			network.Name,
		) {
			cidr := toPatch[network.Name].CIDR
			gateway := toPatch[network.Name].Gateway
			if !cidr.IsValid() && !gateway.IsValid() {
				continue
			}
			err := addIPv6ToNetwork(
				&network,
				cidr,
				gateway,
			)
			if err != nil {
				return fmt.Errorf(
					"failed to add IPv6 to network [%s] because %v",
					network.Name,
					err,
				)
			}
			*updatedSLSNetworks = append(
				*updatedSLSNetworks,
				network,
			)
			continue
		}
	}
	return err
}

func removeIPv6SLSNetworks(slsNetworks *map[string]slsCommon.Network, updatedSLSNetworks *[]slsCommon.Network) (err error) {
	for _, network := range *slsNetworks {
		if slices.Contains(
			networksToPatch,
			network.Name,
		) {
			err := removeIPv6FromNetwork(&network)
			if err != nil {
				return fmt.Errorf(
					"failed to remove IPv6 from network [%s] because %v",
					network.Name,
					err,
				)
			}
			*updatedSLSNetworks = append(
				*updatedSLSNetworks,
				network,
			)
		}
	}
	return err
}

func addIPv6ToNetwork(network *slsCommon.Network, prefix netip.Prefix, gateway netip.Addr) (err error) {
	extraProperties, err := sls.UnmarshalNetworkExtraProperties(network)
	if err != nil {
		return fmt.Errorf(
			"failed to add IPv6 to network [%s] because %v",
			network.Name,
			err,
		)
	}
	if !force && extraProperties.CIDR6 != "" {
		log.Printf(
			"Network [%s] CIDR6 was already defined as [%s].\n",
			network.Name,
			extraProperties.CIDR6,
		)
		forceAlert = true
	} else {
		extraProperties.CIDR6 = prefix.String()
		fmt.Printf(
			"%-40s ...\n%-40s : %s\n%-40s : %s\n",
			fmt.Sprintf(
				"Adding IPv6 to %s network",
				network.Name,
			),
			"* CIDR6",
			prefix.String(),
			"* Gateway6",
			gateway.String(),
		)
	}
	ipNetwork := networking.IPNetwork{
		Name:     network.Name,
		FullName: network.FullName,
		CIDR6:    extraProperties.CIDR6,
	}
	err = allocateSLSSubnets(
		&ipNetwork,
		&extraProperties,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to add IPv6 to network [%s] subnets because %v",
			network.Name,
			err,
		)
	}
	err = allocateIPReservations(
		&extraProperties,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to add IPv6 reservations to network [%s] because %v",
			network.Name,
			err,
		)
	}

	// Apply the supernet hack to our limited subnets.
	networking.SupernetSubnets = subnetsToPatch
	ipNetwork.ApplySupernetHack()
	for _, subnet := range ipNetwork.Subnets {
		_, index, _ := extraProperties.LookupSubnet(subnet.Name)
		extraProperties.Subnets[index].CIDR6 = subnet.CIDR6
		extraProperties.Subnets[index].Gateway6 = subnet.Gateway6
	}
	network.ExtraPropertiesRaw = extraProperties
	return err
}

func removeIPv6FromNetwork(network *slsCommon.Network) (err error) {
	extraProperties, err := sls.UnmarshalNetworkExtraProperties(network)
	if err != nil {
		return fmt.Errorf(
			"failed to add IPv6 to network [%s] because %v",
			network.Name,
			err,
		)
	}
	fmt.Printf(
		"Removing [%s] from %s network.\n",
		extraProperties.CIDR6,
		network.Name,
	)
	extraProperties.CIDR6 = ""
	err = deallocateSLSSubnets(
		&extraProperties,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to remove IPv6 addresses from [%s] subnets because %v",
			network.Name,
			err,
		)
	}
	err = deallocateIPReservations(
		&extraProperties,
	)
	network.ExtraPropertiesRaw = extraProperties
	return err
}

func allocateSLSSubnets(network *networking.IPNetwork, slsProperties *slsCommon.NetworkExtraProperties) (err error) {
	for _, subnetName := range subnetsToPatch {
		subnet, i, err := slsProperties.LookupSubnet(subnetName)
		if err != nil {
			continue
		}
		numReservations := len(subnet.ReservedIPs())
		_, subnetPrefix6, _ := network.SubnetWithin(
			uint64(numReservations),
		)
		newSubnet, err := network.CreateSubnetByMask(
			nil,
			net.CIDRMask(
				subnetPrefix6.Bits(),
				networking.IPv6Size,
			),
			subnetName,
			subnet.VlanID,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to create IPv6 subnet because %v",
				err,
			)
		}
		if !force && subnet.CIDR6 != "" {
			forceAlert = true
		} else {
			subnet.CIDR6 = newSubnet.CIDR6
		}
		if !force && subnet.Gateway6 != nil {
			forceAlert = true
		} else {
			subnet.Gateway6 = newSubnet.Gateway6
		}
		slsProperties.Subnets[i] = subnet
	}
	return err
}

func deallocateSLSSubnets(slsProperties *slsCommon.NetworkExtraProperties) (err error) {
	for _, subnetName := range subnetsToPatch {
		subnet, i, err := slsProperties.LookupSubnet(subnetName)
		if err != nil {
			continue
		}
		subnet.CIDR6 = ""
		subnet.Gateway6 = nil
		slsProperties.Subnets[i] = subnet
	}
	return err
}

func allocateIPReservations(slsProperties *slsCommon.NetworkExtraProperties) (err error) {
	for _, subnetName := range subnetsToPatch {
		subnet, i, err := slsProperties.LookupSubnet(subnetName)
		if err != nil {
			// Silently continue to avoid printing redundant "not found" messages.
			continue
		}
		totalReservations := len(subnet.IPReservations)
		var skippedReservations uint

		// FIXME: replace this sort with something less ambiguous
		IPReservations := subnet.IPReservations
		sort.Slice(
			IPReservations,
			func(i, j int) bool {
				iAddr, err := netip.ParseAddr(IPReservations[i].IPAddress.String())
				if err != nil {
					log.Fatalf(
						"Failed to parse IPv4 address of %s because %v\n",
						IPReservations[i].Name,
						err,
					)
				}
				jAddr, err := netip.ParseAddr(IPReservations[j].IPAddress.String())
				if err != nil {
					log.Fatalf(
						"Failed to parse IPv4 address of %s because %v\n",
						IPReservations[j].Name,
						err,
					)
				}
				return iAddr.Less(jAddr)
			},
		)
		for j, IPReservation := range subnet.IPReservations {
			if !force && IPReservation.IPAddress6 != nil {
				skippedReservations++
				forceAlert = true
				continue
			}
			// Free the reserved IPAddress and IPAddress6, allowing us to re-use it.
			slsProperties.Subnets[i].IPReservations[j].IPAddress = nil
			slsProperties.Subnets[i].IPReservations[j].IPAddress6 = nil
			updatedReservation, err := networking.UpdateReservation(
				&subnet,
				IPReservation,
				true,
			)
			if err != nil {
				log.Fatalf(
					"Failed to update reservation for [%s] in subnet [%s] because %v",
					IPReservation.Name,
					subnetName,
					err,
				)
			}
			slsProperties.Subnets[i].IPReservations[j] = updatedReservation
		}
		fmt.Printf(
			"%-40s ...\n",
			fmt.Sprintf(
				"* %s subnet",
				subnetName,
			),
		)
		fmt.Printf(
			"%-40s : %-3d\n",
			"** new reservations",
			totalReservations,
		)
		if skippedReservations > 0 {
			fmt.Printf(
				"%-40s : %-3d\n",
				"** prior reservations",
				skippedReservations,
			)
		}
	}
	return err
}

func deallocateIPReservations(slsProperties *slsCommon.NetworkExtraProperties) (err error) {
	for _, subnetName := range subnetsToPatch {
		subnet, i, err := slsProperties.LookupSubnet(subnetName)
		if err != nil {
			// Silently continue to avoid printing redundant "not found" messages.
			continue
		}
		for j, ipReservation := range subnet.IPReservations {
			ipReservation.IPAddress6 = nil
			slsProperties.Subnets[i].IPReservations[j] = ipReservation
		}
	}
	return err
}

func putSLSNetworks(client slsClient.SLSClient, slsNetworks []slsCommon.Network) (err error) {
	for _, network := range slsNetworks {
		err = client.PutNetwork(
			context.Background(),
			network,
		)
		if err != nil {
			err = fmt.Errorf(
				"failed to update %s network in SLS because %v",
				network.Name,
				err,
			)
			break
		}
	}
	return err
}
