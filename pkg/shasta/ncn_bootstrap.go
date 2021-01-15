/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"fmt"
	"log"
	"net"
	"strings"

	base "stash.us.cray.com/HMS/hms-base"
)

// LogicalNCN is the main struct for NCNs
// It replaces the deprecated BootstrapNCNMetadata
// and still matches the ncn_metadata.csv file as
// NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
type LogicalNCN struct {
	Role             string         `yaml:"role" json:"role" csv:"NCN Role"`
	Subrole          string         `yaml:"subrole" json:"subrole" csv:"NCN Subrole"`
	BmcMac           string         `yaml:"bmc-mac" json:"bmc-mac" csv:"BMC MAC"`
	BmcPort          string         `yaml:"bmc-port" json:"bmc-port" csv:"BMC Switch Port"`
	NmnMac           string         `yaml:"nmn-mac" json:"nmn-mac" csv:"NMN MAC"`
	NmnPort          string         `yaml:"nmn-port" json:"nmn-port" csv:"NMN Switch Port"`
	Xname            string         `yaml:"xname" json:"xname" csv:"NCN xname"`
	Hostname         string         `yaml:"hostname" json:"hostname" csv:"-"`
	InstanceID       string         `yaml:"instance-id" json:"instance-id" csv:"-"` // should be unique for the life of the image
	Region           string         `yaml:"region" json:"region" csv:"-"`
	AvailabilityZone string         `yaml:"availability-zone" json:"availability-zone" csv:"-"`
	ShastaRole       string         `yaml:"shasta-role" json:"shasta-role" csv:"-"` // map to HSM Subrole
	Aliases          []string       `yaml:"aliases" json:"aliases" csv:"-"`
	Networks         []NCNNetwork   `yaml:"networks" json:"networks" csv:"-"`
	Interfaces       []NCNInterface `yaml:"interfaces" json:"interfaces" csv:"-"`
	BmcIP            string         `yaml:"bmc-ip" json:"bmc-ip" csv:"-"`
	Bond0Mac0        string         `yaml:"bond0-mac0" json:"bond0-mac0" csv:"-"`
	Bond0Mac1        string         `yaml:"bond0-mac1" json:"bond0-mac1" csv:"-"`
	Cabinet          string         `yaml:"cabinet" json:"cabinet" csv:"-"` // Use to establish availability zone
}

// Validate is a validator that checks for a minimum set of info
func (lncn *LogicalNCN) Validate() error {
	xname := lncn.Xname

	// First off verify that this is a valid xname
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid xname for NCN: %s", xname)
	}

	// Next, verify that the xname is type of Node
	if base.GetHMSType(xname) != base.Node {
		return fmt.Errorf("invalid type %s for NCN xname: %s", base.GetHMSTypeString(xname), xname)
	}

	if lncn.Role == "" {
		// TODO Verify the role against the listing of valid Roles
		return fmt.Errorf("empty role")
	}
	if lncn.Subrole == "" {
		// TODO Verify the role against the listing of valid SubRoles
		return fmt.Errorf("empty sub-role")
	}
	return nil
}

// Normalize the values of a LogicalNCN
func (lncn *LogicalNCN) Normalize() error {
	// Right now we only need to the normalize the xname for the switch. IE strip any leading 0s
	lncn.Xname = base.NormalizeHMSCompID(lncn.Xname)

	return nil
}

// GetIP takes in a netname and returns an IP address for that netname
func (lncn *LogicalNCN) GetIP(netName string) net.IP {
	for _, inet := range lncn.Networks {
		if inet.NetworkName == netName {
			return net.ParseIP(inet.IPAddress)
		}
	}
	return net.IP{}
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
	lookup := func(name string, subnetName string, networks map[string]*IPV4Network) *IPV4Subnet {
		tempNetwork := networks[name]
		subnet, err := tempNetwork.LookUpSubnet(subnetName)
		if err != nil {
			log.Printf("couldn't find a %v subnet in the %v network \n", subnetName, name)
		}
		// log.Printf("found a %v subnet in the %v network", subnetName, name)
		return subnet
	}

	// Build a map of networks based on their names
	netNames := []string{"CAN", "MTL", "NMN", "HMN"}
	subnets := make(map[string]*IPV4Subnet)
	for _, name := range netNames {
		subnets[name] = lookup(name, "bootstrap_dhcp", networks)
	}

	// Loop through the NCNs and then run through the networks to add reservations and assign ip addresses
	for _, ncn := range ncns {
		for netName, subnet := range subnets {
			// reserve the bmc ip
			if netName == "HMN" {
				// The bmc xname is the ncn xname without the final two characters
				// NCN Xname = x3000c0s9b0n0  BMC Xname = x3000c0s9b0
				ncn.BmcIP = subnet.AddReservation(fmt.Sprintf("%v", strings.TrimSuffix(ncn.Xname, "n0")), fmt.Sprintf("%v-mgmt", ncn.Xname)).IPAddress.String()
			}
			// Hostname is not available a the point AllocateIPs should be called.
			reservation := subnet.AddReservation(ncn.Xname, ncn.Xname)
			// log.Printf("Adding %v %v reservation for %v(%v) at %v \n", netName, subnet.Name, ncn.Xname, ncn.Xname, reservation.IPAddress.String())
			prefixLen := strings.Split(subnet.CIDR.String(), "/")[1]
			tempNetwork := NCNNetwork{
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
