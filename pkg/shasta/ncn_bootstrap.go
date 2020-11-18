/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"fmt"
	"log"
	"strings"
)

// LogicalNCN is the main struct for NCNs
type LogicalNCN struct {
	Xname            string         `yaml:"xname" json:"xname"`
	Hostname         string         `yaml:"hostname" json:"hostname"`
	InstanceID       string         `yaml:"instance-id" json:"instance-id"` // should be unique for the life of the image
	Region           string         `yaml:"region" json:"region"`
	AvailabilityZone string         `yaml:"availability-zone" json:"availability-zone"`
	ShastaRole       string         `yaml:"shasta-role" json:"shasta-role"` // map to HSM Subrole
	Aliases          []string       `yaml:"aliases" json:"aliases"`
	Networks         []NCNNetwork   `yaml:"networks" json:"networks"`
	Interfaces       []NCNInterface `yaml:"interfaces" json:"interfaces"`
	BMCMac           string         `yaml:"bmc-mac" json:"bmc-mac"`
	BMCIp            string         `yaml:"bmc-ip" json:"bmc-ip"`
	NMNMac           string         `yaml:"nmn-mac" json:"nmn-mac"`
	Bond0Mac0        string         `yaml:"bond0-mac0" json:"bond0-mac0"`
	Bond0Mac1        string         `yaml:"bond0-mac1" json:"bond0-mac1"`
	Cabinet          string         `yaml:"cabinet" json:"cabinet"` // Use to establish availability zone
}

// NCNNetwork holds information about networks
type NCNNetwork struct {
	NetworkName   string `json:"network-name"`
	FullName      string `json:"full-name"`
	IPAddress     string `json:"ip-address"`
	InterfaceName string `json:"net-device"`
	InterfaceMac  string `json:"mac-address"`
	Vlan          int    `json:"vlan"`
	CIDR          string `json:"cidr"`
	Mask          string `json:"mask"`
}

// NCNInterface holds information for all MAC addresses in all NCNs. CSV definitions are the lshw fields
type NCNInterface struct {
	InterfaceType string `json:"product" csv:"product"`
	PCIAddress    string `json:"pci-address" csv:"bus info"`
	DeviceName    string `json:"device-name" csv:"logical name"`
	MacAddress    string `json:"mac-address" csv:"serial"`
	IPAddress     string `json:"ip-address" csv:"_"`
	Usage         string `json:"usage" csv:"-"`
}

// AllocateIps distributes IP reservations for each of the NCNs within the networks
func AllocateIps(ncns []*LogicalNCN, networks map[string]*IPV4Network) {
	lookup := func(name string, networks map[string]*IPV4Network) *IPV4Subnet {
		tempNetwork := networks[name]
		subnet, err := tempNetwork.LookUpSubnet("bootstrap_dhcp")
		if err != nil {
			log.Printf("couldn't find a bootstrap_dhcp subnet in the %v network", name)
		}
		// log.Printf("found a bootstrap_dhcp subnet in the %v network", name)
		return subnet
	}

	// Build a map of networks based on their names
	netNames := [4]string{"CAN", "MTL", "NMN", "HMN"}
	subnets := make(map[string]*IPV4Subnet)
	for _, name := range netNames {
		subnets[name] = lookup(name, networks)
	}

	// Loop through the NCNs and then run through the networks to add reservations and assign ip addresses
	for _, ncn := range ncns {
		for netName, subnet := range subnets {
			// reserve the bmc ip
			if netName == "HMN" {
				ncn.BMCIp = subnet.AddReservation(fmt.Sprintf("bmc-%v", ncn.Hostname), fmt.Sprintf("bmc-%v", ncn.Xname)).IPAddress.String()
			}
			reservation := subnet.AddReservation(ncn.Hostname, ncn.Xname)
			// log.Printf("Adding %v %v reservation for %v(%v) at %v \n", netName, subnet.Name, reservation.Name, reservation.Comment, reservation.IPAddress.String())
			prefixLen := strings.Split(subnet.CIDR.String(), "/")[1]
			ncn.Networks = append(ncn.Networks, NCNNetwork{
				NetworkName: netName,
				IPAddress:   reservation.IPAddress.String(),
				Vlan:        int(subnet.VlanID),
				FullName:    subnet.FullName,
				CIDR:        strings.Join([]string{reservation.IPAddress.String(), prefixLen}, "/"),
				Mask:        prefixLen,
			})

		}
	}
}
