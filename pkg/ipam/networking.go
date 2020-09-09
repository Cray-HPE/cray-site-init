/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package ipam

import (
	"fmt"
	"io/ioutil"
	"net"

	"gopkg.in/yaml.v2"
)

// IPReservation is a type for managing IP Reservations
type IPReservation struct {
	IPAddress net.IPAddr
	name      string
	comment   string
}

// ShastaIPV4Subnet is a type for managing IPv4 Subnets
type shastaIPV4SubnetFields struct {
	FullName string `yaml:"full_name"`
	CIDR     string `yaml:"cidr"` // Convert to net.IPNet on use
	// IPReservations []IPReservation `yaml:"ip_reservations"`
	Name      string  `yaml:"name"`
	Comment   string  `yaml:"comment"`
	Gateway   string  `yaml:"gateway"` // Convert to net.IPAddr on use
	DHCPRange IPRange `yaml:"iprange"`
}

// ShastaIPV4Subnet is a type for managing IPv4 Subnets
type ShastaIPV4Subnet struct {
	FullName       string          `yaml:"full_name"`
	CIDR           net.IPNet       `yaml:"cidr"` // Convert to net.IPNet on use
	IPReservations []IPReservation `yaml:"ip_reservations"`
	Name           string          `yaml:"name"`
	Comment        string          `yaml:"comment"`
	Gateway        string          `yaml:"gateway"` // Convert to net.IPAddr on use
	DHCPRange      IPRange         `yaml:"iprange"`
}

// shastaIPV4NetworkFields is a type for managing the high level IPv4  networks in Shasta
type shastaIPV4NetworkFields struct {
	FullName string `yaml:"full_name"`
	CIDR     string `yaml:"cidr"` // Convert to net.IPNet on use
	// IPReservations []IPReservation    `yaml:"ip_reservations"`
	// Subnets        []ShastaIPV4Subnet `yaml:"subnets"`
	Name    string `yaml:"name"`
	Vlan    int16  `yaml:"vlan_id"`
	MTU     int16  `yaml:"mtu"`
	NetType string `yaml:"type"` // ethernet or ???
	Comment string `yaml:"comment"`
}

// ShastaIPV4Network is a type for managing IPv4 Networks
type ShastaIPV4Network struct {
	FullName       string             `yaml:"full_name"`
	CIDR           net.IPNet          `yaml:"cidr"`
	IPReservations []IPReservation    `yaml:"ip_reservations"`
	Subnets        []ShastaIPV4Subnet `yaml:"subnets"`
	Name           string             `yaml:"name"`
	Vlan           int16              `yaml:"vlan_id"`
	MTU            int16              `yaml:"mtu"`
	NetType        string             `yaml:"type"` // ethernet or ???
	Comment        string             `yaml:"comment"`
}

// NewShastaIPV4Network creates a validated IPV4Network object
func NewShastaIPV4Network(fields shastaIPV4NetworkFields) ShastaIPV4Network {
	_, net, _ := net.ParseCIDR(fields.CIDR)
	n := ShastaIPV4Network{
		FullName: fields.FullName,
		CIDR:     *net,
		Name:     fields.Name,
		Vlan:     fields.Vlan,
		MTU:      fields.MTU,
		NetType:  fields.NetType,
		Comment:  fields.Comment,
	}
	return n
}

// AddSubnet attempts to safely add a subnet to an existing Network definition
func (network ShastaIPV4Network) AddSubnet(subnet ShastaIPV4Subnet) error {
	// test to see if the overall network can contain the CIDR
	if Contains(network.CIDR, subnet.CIDR) {
		// Check to see if the subnet is already contained in any of the existing subnets
		// This not only tests for duplicates, but also for potential overlap problems
		for _, s := range network.Subnets {
			if Contains(s.CIDR, subnet.CIDR) {
				return fmt.Errorf("ipam: Network (%s) can contain Subnet (%s), but already has (%s) which overlaps ", network.CIDR, subnet.CIDR, s.CIDR)
			}
			network.Subnets = append(network.Subnets, subnet)
			return nil
		}
	}
	return fmt.Errorf("ipam: Network (%s) cannot contain Subnet (%s)", network.CIDR, subnet.CIDR)
}

// LoadNetworkFromFile attempts to load a network definition from a yaml file
func LoadNetworkFromFile(filename string) (ShastaIPV4Network, error) {
	var myNetworkFields shastaIPV4NetworkFields
	var myNetwork ShastaIPV4Network
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return myNetwork, err
	}
	err = yaml.Unmarshal(yamlFile, &myNetworkFields)
	myNetwork = NewShastaIPV4Network(myNetworkFields)
	return myNetwork, err
}
