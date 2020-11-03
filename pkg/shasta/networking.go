/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	"stash.us.cray.com/MTL/sic/pkg/ipam"
)

// IPReservation is a type for managing IP Reservations
type IPReservation struct {
	IPAddress net.IP   `yaml:"ip_address"`
	Name      string   `yaml:"name"`
	Comment   string   `yaml:"comment"`
	Aliases   []string `yaml:"aliases"`
}

// IPV4Network is a type for managing IPv4 Networks
type IPV4Network struct {
	FullName  string                 `yaml:"full_name"`
	CIDR      string                 `yaml:"cidr"`
	Subnets   []IPV4Subnet           `yaml:"subnets"`
	Name      string                 `yaml:"name"`
	VlanRange []int16                `yaml:"vlan_range"`
	MTU       int16                  `yaml:"mtu"`
	NetType   sls_common.NetworkType `yaml:"type"`
	Comment   string                 `yaml:"comment"`
}

// IPV4Subnet is a type for managing IPv4 Subnets
type IPV4Subnet struct {
	FullName       string          `yaml:"full_name" form:"full_name" mapstructure:"full_name"`
	CIDR           net.IPNet       `yaml:"cidr"`
	IPReservations []IPReservation `yaml:"ip_reservations"`
	Name           string          `yaml:"name" form:"name" mapstructure:"name"`
	VlanID         int16           `yaml:"vlan_id" form:"vlan_id" mapstructure:"vlan_id"`
	Comment        string          `yaml:"comment"`
	Gateway        net.IP          `yaml:"gateway"`
	DHCPStart      net.IP          `yaml:"iprange-start"`
	DHCPEnd        net.IP          `yaml:"iprange-end"`
}

// ManagementSwitch is a type for managing Management switches
type ManagementSwitch struct {
	Xname               string `json:"xname" mapstructure:"xname" csv:"xname"` // Required for SLS
	Name                string `json:"name" mapstructure:"name" csv:"-"`       // Required for SLS to update DNS
	Brand               string `json:"brand" mapstructure:"brand" csv:"-"`
	Model               string `json:"model" mapstructure:"model" csv:"model"`
	Os                  string `json:"operating-system" mapstructure:"operating-system" csv:"-"`
	Firmware            string `json:"firmware" mapstructure:"firmware" csv:"-"`
	SwitchType          string `json:"type" mapstructure:"type" csv:"-"` //"CDU/Leaf/Spine"
	ManagementInterface net.IP `json:"ip" mapstructure:"ip" csv:"-"`     // SNMP/REST interface IP (not a distinct BMC)  // Required for SLS
}

// GenSubnets subdivides a network into a set of subnets
func (iNet *IPV4Network) GenSubnets(cabinetDetails []CabinetDetail, mask net.IPMask, mgmtIPReservations int, spines string, leafs string) error {
	spineXnames := strings.Split(spines, ",")
	leafXnames := strings.Split(spines, ",")

	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	mySubnets := iNet.AllocatedSubnets()
	myIPv4Subnets := iNet.Subnets

	for _, cabinetDetail := range cabinetDetails {

		for i := 0; i < int(cabinetDetail.Cabinets); i++ {
			newSubnet, err := ipam.Free(*myNet, mask, mySubnets)
			mySubnets = append(mySubnets, newSubnet)
			if err != nil {
				log.Printf("Couldn't add subnet because %v \n", err)
				return err
			}
			tempSubnet := IPV4Subnet{
				CIDR:    newSubnet,
				Name:    fmt.Sprintf("cabinet_%v", cabinetDetail.StartingCabinet+i),
				Gateway: ipam.Add(newSubnet.IP, 1),
				// Reserving the first vlan in the range for a non-cabinet aligned vlan if needed in the future.
				VlanID: iNet.VlanRange[1] + int16(i),
			}
			tempSubnet.ReserveNetMgmtIPs(mgmtIPReservations, spineXnames, leafXnames)
			myIPv4Subnets = append(myIPv4Subnets, tempSubnet)
		}
	}
	iNet.Subnets = myIPv4Subnets
	return nil
}

// AllocatedSubnets returns a list of the allocated subnets
func (iNet IPV4Network) AllocatedSubnets() []net.IPNet {
	var myNets []net.IPNet
	for _, v := range iNet.Subnets {
		myNets = append(myNets, v.CIDR)
	}
	return myNets
}

// AddSubnet allocates a new subnet
func (iNet *IPV4Network) AddSubnet(mask net.IPMask, name string, vlanID int16) (*IPV4Subnet, error) {
	var tempSubnet IPV4Subnet
	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	newSubnet, err := ipam.Free(*myNet, mask, iNet.AllocatedSubnets())
	if err != nil {
		log.Printf("Couldn't add subnet because %v \n", err)
		return &tempSubnet, err
	}
	iNet.Subnets = append(iNet.Subnets, IPV4Subnet{
		CIDR:    newSubnet,
		Name:    name,
		Gateway: ipam.Add(newSubnet.IP, 1),
		VlanID:  vlanID,
	})
	return &iNet.Subnets[len(iNet.Subnets)-1], nil
}

// LookUpSubnet returns a subnet by name
func (iNet *IPV4Network) LookUpSubnet(name string) (IPV4Subnet, error) {
	for _, v := range iNet.Subnets {
		if v.Name == name {
			return v, nil
		}
	}
	return IPV4Subnet{}, errors.New("Subnet not found")
}

// ReserveNetMgmtIPs reserves (n) IP addresses for management networking equipment
func (iSubnet *IPV4Subnet) ReserveNetMgmtIPs(n int, spines []string, leafs []string) {
	for i := 1; i <= n; i++ {
		// First allocate the spines and then the leafs
		if i < len(spines) {
			iSubnet.AddReservation(fmt.Sprintf("sw-spine-%03d", i), spines[i])
		} else if i < len(spines)+len(leafs) {
			iSubnet.AddReservation(fmt.Sprintf("sw-leaf-%03d", i-len(spines)), leafs[i-len(spines)])
		} else {
			iSubnet.AddReservation(fmt.Sprintf("mgmt_net_%03d", i), "")
		}
	}
}

// ReservedIPs returns a list of IPs already reserved within the subnet
func (iSubnet *IPV4Subnet) ReservedIPs() []net.IP {
	var addresses []net.IP
	for _, v := range iSubnet.IPReservations {
		addresses = append(addresses, v.IPAddress)
	}
	return addresses
}

// AddReservation adds a new IP reservation to the subnet
func (iSubnet *IPV4Subnet) AddReservation(name, comment string) *IPReservation {
	myReservedIPs := iSubnet.ReservedIPs()
	// Start counting from the bottom knowing the gateway is on the bottom
	tempIP := ipam.Add(iSubnet.CIDR.IP, 2)
	for {
		for _, v := range myReservedIPs {
			if tempIP.Equal(v) {
				// log.Printf("Found %v already in the reservations list. \n", v)
				tempIP = ipam.Add(tempIP, 1)
			}
		}
		iSubnet.IPReservations = append(iSubnet.IPReservations, IPReservation{
			IPAddress: tempIP,
			Name:      name,
			Comment:   comment,
		})
		return &iSubnet.IPReservations[len(iSubnet.IPReservations)-1]
	}

}
