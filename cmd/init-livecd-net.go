/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"net"
	"path/filepath"

	"github.com/spf13/viper"

	sicFiles "stash.us.cray.com/MTL/sic/internal/files"
	"stash.us.cray.com/MTL/sic/pkg/ipam"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// BuildLiveCDNetworks creates an array of IPv4 Networks based on the supplied system configuration
func BuildLiveCDNetworks(conf shasta.SystemConfig, v *viper.Viper) (map[string]*shasta.IPV4Network, error) {
	// our primitive ipam uses the number of cabinets to lay out a network for each one.
	// It is per-cabinet type which is pretty annoying, but here we are.

	cabinetDetails := buildCabinetDetails(v)

	var networkMap = make(map[string]*shasta.IPV4Network)

	// Start the NMN with out defaults
	tempNMN := shasta.DefaultNMN
	// Update the CIDR from flags/viper
	tempNMN.CIDR = v.GetString("nmn-cidr")
	// Add a /25 for the Load Balancers
	pool, err := tempNMN.AddSubnet(net.CIDRMask(25, 32), "nmn_metallb_address_pool", int16(v.GetInt("nmn-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	pool.AddReservation("api_gateway")
	// Add a /26 for bootstrap dhcp
	subnet, err := tempNMN.AddSubnet(net.CIDRMask(26, 32), "bootstrap_dhcp", int16(v.GetInt("nmn-bootstrap-vlan")))
	subnet.DHCPStart = ipam.Add(subnet.CIDR.IP, v.GetInt("management-net-ips"))
	subnet.DHCPEnd = ipam.Add(ipam.Broadcast(subnet.CIDR), -1)
	// Divide the network into an appropriate number of subnets
	tempNMN.GenSubnets(cabinetDetails, net.CIDRMask(22, 32), v.GetInt("management-net-ips"))
	networkMap["nmn"] = &tempNMN

	// Start the HMN with out defaults
	tempHMN := shasta.DefaultHMN
	// Update the CIDR from flags/viper
	tempHMN.CIDR = v.GetString("hmn-cidr")
	// Add a /25 for the Load Balancers
	pool, err = tempHMN.AddSubnet(net.CIDRMask(25, 32), "hmn_metallb_address_pool", int16(v.GetInt("hmn-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	pool.AddReservation("api_gateway")

	// Add a /26 for bootstrap dhcp
	subnet, err = tempHMN.AddSubnet(net.CIDRMask(26, 32), "bootstrap_dhcp", int16(v.GetInt("hmn-bootstrap-vlan")))
	subnet.DHCPStart = ipam.Add(subnet.CIDR.IP, v.GetInt("management-net-ips"))
	subnet.DHCPEnd = ipam.Add(ipam.Broadcast(subnet.CIDR), -1)
	// Divide the network into an appropriate number of subnets
	tempHMN.GenSubnets(cabinetDetails, net.CIDRMask(22, 32), v.GetInt("management-net-ips"))
	networkMap["hmn"] = &tempHMN

	// Start the HSN with out defaults
	tempHSN := shasta.DefaultHSN
	// Update the CIDR from flags/viper
	tempHSN.CIDR = v.GetString("hsn-cidr")
	// Add a /25 for the Load Balancers
	pool, err = tempHSN.AddSubnet(net.CIDRMask(25, 32), "hsn_metallb_address_pool", tempHSN.VlanRange[0])
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	pool.AddReservation("api_gateway")

	// Divide the network into an appropriate number of subnets
	tempHSN.GenSubnets(cabinetDetails, net.CIDRMask(22, 32), v.GetInt("management-net-ips"))
	networkMap["hsn"] = &tempHSN

	// Start the MTL with our defaults
	tempMTL := shasta.DefaultMTL
	// Update the CIDR from flags/viper
	tempMTL.CIDR = v.GetString("mtl-cidr")
	// No need to subdivide the mtl network by cabinets
	subnet, err = tempMTL.AddSubnet(net.CIDRMask(24, 32), "bootstrap_dhcp", 0)
	subnet.ReserveNetMgmtIPs(v.GetInt("management-net-ips"))
	subnet.DHCPStart = ipam.Add(subnet.CIDR.IP, v.GetInt("management-net-ips"))
	subnet.DHCPEnd = ipam.Add(ipam.Broadcast(subnet.CIDR), -1)
	networkMap["mtl"] = &tempMTL

	// Start the CAN with our defaults
	tempCan := shasta.DefaultCAN
	// Update the CIDR from flags/viper
	tempCan.CIDR = v.GetString("can-cidr") // This is probably a /24
	// Add a /25 for the Load Balancers on vlan0007
	_, err = tempCan.AddSubnet(net.CIDRMask(25, 32), "can_metallb_address_pool", int16(v.GetInt("can-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	// Add a /28 for the Static Pool on vlan0007
	_, err = tempCan.AddSubnet(net.CIDRMask(28, 32), "can_metallb_static_pool", int16(v.GetInt("can-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	// Add a /26 for bootstrap dhcp
	subnet, err = tempCan.AddSubnet(net.CIDRMask(26, 32), "bootstrap_dhcp", int16(v.GetInt("hmn-bootstrap-vlan")))
	subnet.ReserveNetMgmtIPs(v.GetInt("management-net-ips"))
	subnet.DHCPStart = ipam.Add(subnet.CIDR.IP, v.GetInt("management-net-ips"))
	subnet.DHCPEnd = ipam.Add(ipam.Broadcast(subnet.CIDR), -1)
	networkMap["can"] = &tempCan

	return networkMap, nil
}

// WriteNetworkFiles persistes our network configuration to disk in a directory of yaml files
func WriteNetworkFiles(basepath string, networks map[string]*shasta.IPV4Network) {
	for k, v := range networks {
		sicFiles.WriteYAMLConfig(filepath.Join(basepath, fmt.Sprintf("networks/%v.yaml", k)), v)
	}
}

func buildCabinetDetails(v *viper.Viper) []shasta.CabinetDetail {
	var cabinets []shasta.CabinetDetail
	// Add the River Cabinets
	cabinets = append(cabinets, shasta.CabinetDetail{
		Kind:            "river",
		Cabinets:        v.GetInt("river-cabinets"),
		StartingCabinet: v.GetInt("starting-rivier-cabinet"),
	})
	cabinets = append(cabinets, shasta.CabinetDetail{
		Kind:            "hill",
		Cabinets:        v.GetInt("hill-cabinets"),
		StartingCabinet: v.GetInt("starting-hill-cabinet"),
	})
	cabinets = append(cabinets, shasta.CabinetDetail{
		Kind:            "mountain",
		Cabinets:        v.GetInt("mountain-cabinets"),
		StartingCabinet: v.GetInt("starting-mountain-cabinet"),
	})
	return cabinets
}
