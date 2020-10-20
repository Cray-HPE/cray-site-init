/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"errors"
	"fmt"
	"log"
	"net"

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
	Xname               string     `form:"xname" mapstructure:"xname"` // Required for SLS
	Name                string     `form:"name" mapstructure:"name"`   // Required for SLS to update DNS
	Brand               string     `form:"brand" mapstructure:"brand"`
	Model               string     `form:"model" mapstructure:"model"`
	Os                  string     `form:"operating-system" mapstructure:"operating-system"`
	Firmware            string     `form:"firmware" mapstructure:"firmware"`
	SwitchType          string     //"CDU/Leaf/Spine"
	ManagementInterface net.IPAddr // SNMP/REST interface IP (not a distinct BMC)  // Required for SLS
}

func (iNet *IPV4Network) GenSubnets(cabinets uint, startingCabinet int, mask net.IPMask, mgmtIPReservations int) error {
	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	mySubnets := iNet.AllocatedSubnets()
	myIPv4Subnets := iNet.Subnets
	log.Printf("In the GenSubnets function. Cabinets = %v \n", int(cabinets))

	for i := 1; i < int(cabinets); i++ {
		newSubnet, err := ipam.Free(*myNet, mask, mySubnets)
		log.Printf("The %v subnet is %v \n", i, newSubnet)
		mySubnets = append(mySubnets, newSubnet)
		if err != nil {
			log.Printf("Couldn't add subnet because %v \n", err)
			return err
		}
		tempSubnet := IPV4Subnet{
			CIDR:    newSubnet,
			Name:    fmt.Sprintf("cabinet_%v", startingCabinet+i),
			Gateway: ipam.Add(newSubnet.IP, 1),
			// Reserving the first vlan in the range for a non-cabinet aligned vlan if needed in the future.
			VlanID: iNet.VlanRange[1] + int16(i),
		}
		tempSubnet.ReserveNetMgmtIPs(mgmtIPReservations)
		myIPv4Subnets = append(myIPv4Subnets, tempSubnet)
	}
	iNet.Subnets = myIPv4Subnets
	return nil
}

func (iNet IPV4Network) AllocatedSubnets() []net.IPNet {
	var myNets []net.IPNet
	for _, v := range iNet.Subnets {
		myNets = append(myNets, v.CIDR)
	}
	return myNets
}

func (iNet *IPV4Network) AddSubnet(mask net.IPMask, name string, vlanID int16) (*IPV4Subnet, error) {
	var tempSubnet IPV4Subnet
	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	newSubnet, err := ipam.Free(*myNet, mask, iNet.AllocatedSubnets())
	if err != nil {
		log.Printf("Couldn't add subnet because %v \n", err)
		return &tempSubnet, err
	}
	log.Printf("We've got %v subnets before append. \n", len(iNet.Subnets))
	iNet.Subnets = append(iNet.Subnets, IPV4Subnet{
		CIDR:    newSubnet,
		Name:    name,
		Gateway: ipam.Add(newSubnet.IP, 1),
		VlanID:  vlanID,
	})
	log.Printf("We've got %v subnets after append. \n", len(iNet.Subnets))
	return &iNet.Subnets[len(iNet.Subnets)-1], nil
}

func (iNet *IPV4Network) LookUpSubnet(name string) (*IPV4Subnet, error) {
	for _, v := range iNet.Subnets {
		if v.Name == name {
			return &v, nil
		}
	}
	return &IPV4Subnet{}, errors.New("Subnet not found")
}

func (iSubnet *IPV4Subnet) ReserveNetMgmtIPs(n int) {
	for i := 1; i < n+1; i++ {
		iSubnet.AddReservation(fmt.Sprintf("mgmt_net_%03d", i))
	}
}

func (iSubnet *IPV4Subnet) ReservedIPs() []net.IP {
	var addresses []net.IP
	for _, v := range iSubnet.IPReservations {
		addresses = append(addresses, v.IPAddress)
	}
	return addresses
}

func (iSubnet *IPV4Subnet) AddReservation(name string) *IPReservation {
	myReservedIPs := iSubnet.ReservedIPs()
	// Start counting from the bottom knowing the gateway is on the bottom
	tempIP := ipam.Add(iSubnet.CIDR.IP, 2)
	for {
		for _, v := range myReservedIPs {
			if tempIP.Equal(v) {
				log.Printf("Found %v already in the reservations list. \n", v)
				tempIP = ipam.Add(tempIP, 1)
			}
		}
		iSubnet.IPReservations = append(iSubnet.IPReservations, IPReservation{
			IPAddress: tempIP,
			Name:      name,
		})
		return &iSubnet.IPReservations[len(iSubnet.IPReservations)-1]
	}

}
