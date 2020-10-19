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
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// BuildLiveCDNetworks creates an array of IPv4 Networks based on the supplied system configuration
func BuildLiveCDNetworks(conf shasta.SystemConfig, v *viper.Viper) (map[string]shasta.IPV4Network, error) {
	// our primitive ipam uses the number of cabinets to lay out a network for each one.
	var networkMap = make(map[string]shasta.IPV4Network)

	// Start the NMN with out defaults
	tempNMN := shasta.DefaultNMN
	// Update the CIDR from flags/viper
	tempNMN.CIDR = v.GetString("nmn-cidr")
	// Add a /25 for the Load Balancers
	pool, err := tempNMN.AddSubnet(net.CIDRMask(25, 32), "nmn_metallb_address_pool", tempNMN.VlanRange[0])
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	pool.AddReservation("api_gateway")
	// Add a /26 for the Static Pool
	_, err = tempNMN.AddSubnet(net.CIDRMask(25, 32), "nmn_metallb_static_pool", tempNMN.VlanRange[0])
	// Divide the network into an appropriate number of subnets
	tempNMN.GenSubnets(uint(conf.MountainCabinets), int(conf.StartingCabinet), net.CIDRMask(22, 32))
	networkMap["nmn"] = tempNMN

	// Start the HMN with out defaults
	tempHMN := shasta.DefaultHMN
	// Update the CIDR from flags/viper
	tempHMN.CIDR = v.GetString("hmn-cidr")
	// Add a /25 for the Load Balancers
	_, err = tempHMN.AddSubnet(net.CIDRMask(25, 32), "hmn_metallb_address_pool", tempHMN.VlanRange[0])
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	// Add a /26 for the Static Pool
	_, err = tempHMN.AddSubnet(net.CIDRMask(25, 32), "hmn_metallb_static_pool", tempHMN.VlanRange[0])
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	// Divide the network into an appropriate number of subnets
	tempHMN.GenSubnets(uint(conf.MountainCabinets), int(conf.StartingCabinet), net.CIDRMask(22, 32))
	networkMap["hmn"] = tempHMN

	// Start the HSN with out defaults
	tempHSN := shasta.DefaultHSN
	// Update the CIDR from flags/viper
	tempHSN.CIDR = v.GetString("hsn-cidr")
	// Add a /25 for the Load Balancers on vlan0007
	_, err = tempHSN.AddSubnet(net.CIDRMask(25, 32), "hsn_metallb_address_pool", tempHSN.VlanRange[0])
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	// Add a /26 for the Static Pool on vlan0007
	_, err = tempHSN.AddSubnet(net.CIDRMask(25, 32), "hsn_metallb_static_pool", tempHSN.VlanRange[0])
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	// Divide the network into an appropriate number of subnets
	tempHSN.GenSubnets(uint(conf.MountainCabinets), int(conf.StartingCabinet), net.CIDRMask(22, 32))
	networkMap["hsn"] = tempHSN

	// Start the MTL with our defaults
	tempMTL := shasta.DefaultMTL
	// Update the CIDR from flags/viper
	tempMTL.CIDR = v.GetString("mtl-cidr")
	// No need to subdivide the mtl network by cabinets
	_, err = tempMTL.AddSubnet(net.CIDRMask(24, 32), "mtl_subnet", tempMTL.VlanRange[0])
	networkMap["mtl"] = tempMTL

	// Start the CAN with our defaults
	tempCan := shasta.DefaultCAN
	// Update the CIDR from flags/viper
	tempCan.CIDR = v.GetString("can-cidr") // This is probably a /24
	// Add a /25 for the Load Balancers on vlan0007
	_, err = tempCan.AddSubnet(net.CIDRMask(25, 32), "can_metallb_address_pool", int16(7))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	// Add a /26 for the Static Pool on vlan0007
	_, err = tempCan.AddSubnet(net.CIDRMask(25, 32), "can_metallb_static_pool", int16(7))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	networkMap["can"] = tempCan

	return networkMap, nil
}

// WriteNetworkFiles persistes our network configuration to disk in a directory of yaml files
func WriteNetworkFiles(basepath string, networks map[string]shasta.IPV4Network) {
	for k, v := range networks {
		sicFiles.WriteYamlConfig(filepath.Join(basepath, fmt.Sprintf("networks/%v.yaml", k)), v)
	}
}
