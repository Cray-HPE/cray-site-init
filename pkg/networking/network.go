/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

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

package networking

import (
	"fmt"
	"github.com/Cray-HPE/cray-site-init/pkg/sls"
	"log"
	"net"
	"strings"

	slsCommon "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/pkg/errors"

	"github.com/Cray-HPE/cray-site-init/pkg/cli"
)


const (
	// DefaultMTLVlan is the default MTL Bootstrap Vlan - zero (0) represents untagged.
	DefaultMTLVlan = 1
	// DefaultHMNString is the Default HMN String (bond0.hmn0)
	DefaultHMNString = "10.254.0.0/17"
	// DefaultHMNVlan is the default HMN Bootstrap Vlan
	DefaultHMNVlan = 4
	// DefaultHMNMTNString is the default HMN Network for Mountain Cabinets with Grouped Configuration
	DefaultHMNMTNString = "10.104.0.0/17"
	// DefaultHMNRVRString is the default HMN Network for River Cabinets with Grouped Configuration
	DefaultHMNRVRString = "10.107.0.0/17"
	// DefaultNMNString is the Default NMN String (bond0.nmn0)
	DefaultNMNString = "10.252.0.0/17"
	// DefaultNMNVlan is the default NMN Bootstrap Vlan
	DefaultNMNVlan = 2
	// DefaultMacVlanVlan is the default MacVlan Bootstrap Vlan
	DefaultMacVlanVlan = 2
	// DefaultNMNMTNString is the default NMN Network for Mountain Cabinets with Grouped Configuration
	DefaultNMNMTNString = "10.100.0.0/17"
	// DefaultNMNRVRString is the default NMN Network for River Cabinets with Grouped Configuration
	DefaultNMNRVRString = "10.106.0.0/17"
	// DefaultNMNLBString is the default LoadBalancer CIDR for the NMN
	DefaultNMNLBString = "10.92.100.0/24"
	// DefaultHMNLBString is the default LoadBalancer CIDR for the HMN
	DefaultHMNLBString = "10.94.100.0/24"
	// DefaultMacVlanString is the default Macvlan cidr (shares vlan with NMN)
	DefaultMacVlanString = "10.252.124.0/23"
	// DefaultHSNString is the Default HSN String
	DefaultHSNString = "10.253.0.0/16"
	// DefaultCMNString is the Default CMN String (bond0.cmn0)
	DefaultCMNString = "10.103.6.0/24"
	// DefaultCMNVlan is the default CMN Bootstrap Vlan
	DefaultCMNVlan = 7
	// DefaultCANString is the Default CAN String (bond0.can0)
	DefaultCANString = "10.102.11.0/24"
	// DefaultCANVlan is the default CAN Bootstrap Vlan
	DefaultCANVlan = 6
	// DefaultCHNString is the Default CHN String
	DefaultCHNString = "10.104.7.0/24"
	// DefaultCHNVlan is the default CHN Bootstrap Vlan
	DefaultCHNVlan = 5
	// DefaultMTLString is the Default MTL String (bond0 interface)
	DefaultMTLString = "10.1.1.0/16"
)


/*
Handy Netmask Cheet Sheet
/30	4	2	255.255.255.252	1/64
/29	8	6	255.255.255.248	1/32
/28	16	14	255.255.255.240	1/16
/27	32	30	255.255.255.224	1/8
/26	64	62	255.255.255.192	1/4
/25	128	126	255.255.255.128	1/2
/24	256	254	255.255.255.0	1
/23	512	510	255.255.254.0	2
/22	1024	1022	255.255.252.0	4
/21	2048	2046	255.255.248.0	8
/20	4096	4094	255.255.240.0	16
/19	8192	8190	255.255.224.0	32
/18	16384	16382	255.255.192.0	64
/17	32768	32766	255.255.128.0	128
/16	65536	65534	255.255.0.0	256
*/

// DefaultCabinetMask is the default subnet mask for each cabinet
var DefaultCabinetMask = net.CIDRMask(
	22,
	32,
)

// DefaultNetworkingHardwareMask is the default subnet mask for a subnet that contains all networking hardware
var DefaultNetworkingHardwareMask = net.CIDRMask(
	24,
	32,
)

// DefaultLoadBalancerNMN is a thing we need
var DefaultLoadBalancerNMN = IPV4Network{
	FullName: "Node Management Network LoadBalancers",
	CIDR:     DefaultNMNLBString,
	Name:     "NMNLB",
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

// DefaultLoadBalancerHMN is a thing we need
var DefaultLoadBalancerHMN = IPV4Network{
	FullName: "Hardware Management Network LoadBalancers",
	CIDR:     DefaultHMNLBString,
	Name:     "HMNLB",
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

// DefaultBICAN is the default structure for templating the initial BICAN toggle - CMN
var DefaultBICAN = IPV4Network{
	FullName:           "SystemDefaultRoute points the network name of the default route",
	CIDR:               "0.0.0.0/0",
	Name:               "BICAN",
	VlanRange:          []int16{0},
	MTU:                9000,
	NetType:            "ethernet",
	Comment:            "",
	SystemDefaultRoute: "",
}

// DefaultHSN is the default structure for templating initial HSN configuration
var DefaultHSN = IPV4Network{
	FullName: "High Speed Network",
	CIDR:     DefaultHSNString,
	Name:     "HSN",
	VlanRange: []int16{
		613,
		868,
	},
	MTU:     9000,
	NetType: "slingshot10",
	Comment: "",
}

// DefaultCMN is the default structure for templating initial CMN configuration
var DefaultCMN = IPV4Network{
	FullName:     "Customer Management Network",
	CIDR:         DefaultCMNString,
	Name:         "CMN",
	VlanRange:    []int16{DefaultCMNVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultCAN is the default structure for templating initial CAN configuration
var DefaultCAN = IPV4Network{
	FullName:     "Customer Access Network",
	CIDR:         DefaultCANString,
	Name:         "CAN",
	VlanRange:    []int16{DefaultCANVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultCHN is the default structure for templating initial CHN configuration
var DefaultCHN = IPV4Network{
	FullName:     "Customer High-Speed Network",
	CIDR:         DefaultCHNString,
	Name:         "CHN",
	VlanRange:    []int16{DefaultCHNVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultHMN is the default structure for templating initial HMN configuration
var DefaultHMN = IPV4Network{
	FullName:     "Hardware Management Network",
	CIDR:         DefaultHMNString,
	Name:         "HMN",
	VlanRange:    []int16{DefaultHMNVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultNMN is the default structure for templating initial NMN configuration
var DefaultNMN = IPV4Network{
	FullName:     "Node Management Network",
	CIDR:         DefaultNMNString,
	Name:         "NMN",
	VlanRange:    []int16{DefaultNMNVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultMTL is the default structure for templating initial MTL configuration
var DefaultMTL = IPV4Network{
	FullName:     "Provisioning Network (untagged)",
	CIDR:         DefaultMTLString,
	Name:         "MTL",
	VlanRange:    []int16{DefaultMTLVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "This network is only valid for the NCNs",
	ParentDevice: "bond0",
}


// IPV4Network is a type for managing IPv4 Networks
type IPV4Network struct {
	FullName           string                `yaml:"full_name"`
	CIDR               string                `yaml:"cidr"`
	Subnets            []*IPV4Subnet         `yaml:"subnets"`
	Name               string                `yaml:"name"`
	VlanRange          []int16               `yaml:"vlan_range"`
	MTU                int16                 `yaml:"mtu"`
	NetType            slsCommon.NetworkType `yaml:"type"`
	Comment            string                `yaml:"comment"`
	PeerASN            int                   `yaml:"peer-asn"`
	MyASN              int                   `yaml:"my-asn"`
	SystemDefaultRoute string                `yaml:"system_default_route"`
	ParentDevice       string                `yaml:"parent-device"`
}

// IPV4Subnet is a type for managing IPv4 Subnets
type IPV4Subnet struct {
	FullName         string          `yaml:"full_name" form:"full_name" mapstructure:"full_name"`
	CIDR             net.IPNet       `yaml:"cidr"`
	IPReservations   []IPReservation `yaml:"ip_reservations"`
	Name             string          `yaml:"name" form:"name" mapstructure:"name"`
	NetName          string          `yaml:"net-name"`
	VlanID           int16           `yaml:"vlan_id" form:"vlan_id" mapstructure:"vlan_id"`
	Comment          string          `yaml:"comment"`
	Gateway          net.IP          `yaml:"gateway"`
	PITServer        net.IP          `yaml:"_"`
	DNSServer        net.IP          `yaml:"dns_server"`
	DHCPStart        net.IP          `yaml:"iprange-start"`
	DHCPEnd          net.IP          `yaml:"iprange-end"`
	ReservationStart net.IP          `yaml:"reservation-start"`
	ReservationEnd   net.IP          `yaml:"reservation-end"`
	MetalLBPoolName  string          `yaml:"metallb-pool-name"`
	ParentDevice     string          `yaml:"parent-device"`
	InterfaceName    string          `yaml:"interface-name"`
}

// IPReservation is a type for managing IP Reservations
type IPReservation struct {
	IPAddress net.IP   `yaml:"ip_address"`
	Name      string   `yaml:"name"`
	Comment   string   `yaml:"comment"`
	Aliases   []string `yaml:"aliases"`
}

// ApplySupernetHack applys a dirty hack.
func (iNet *IPV4Network) ApplySupernetHack() {
	// Replace the gateway and netmask on the to better support the 1.3 network switch configuration
	// *** This is a HACK ***
	_, superNet, err := net.ParseCIDR(iNet.CIDR)
	if err != nil {
		log.Fatal(
			"Couldn't parse the CIDR for ",
			iNet.Name,
		)
	}
	for _, subnetName := range []string{
		"bootstrap_dhcp",
		"network_hardware",
		"can_metallb_static_pool",
		"can_metallb_address_pool",
	} {
		tempSubnet, err := iNet.LookUpSubnet(subnetName)
		if err == nil {
			// Replace the standard netmask with the supernet netmask
			// Replace the standard gateway with the supernet gateway
			// ** HACK ** We're doing this here to bypass all sanity checks
			// This **WILL** cause an overlap of broadcast domains, but is required
			// for reducing switch configuration changes from 1.3 to 1.4
			tempSubnet.Gateway = Add(
				superNet.IP,
				1,
			)
			tempSubnet.CIDR.Mask = superNet.Mask
		}
	}
}

// GenSubnets subdivides a network into a set of subnets
func (iNet *IPV4Network) GenSubnets(
	cabinetDetails []sls.CabinetGroupDetail, mask net.IPMask, cabinetFilter sls.CabinetFilterFunc,
) error {
	// log.Printf("Generating Subnets for %s\ncabinetType: %v,\n", iNet.Name, cabinetType)
	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	mySubnets := iNet.AllocatedSubnets()
	myIPv4Subnets := iNet.Subnets
	var minVlan, maxVlan int16 = 4095, 0

	for _, cabinetDetail := range cabinetDetails {
		for j, i := range cabinetDetail.CabinetDetails {
			if cabinetFilter(
				cabinetDetail,
				i,
			) {
				newSubnet, err := Free(
					*myNet,
					mask,
					mySubnets,
				)
				mySubnets = append(
					mySubnets,
					newSubnet,
				)
				if err != nil {
					log.Fatalf(
						"Gensubnets couldn't add subnet because %v \n",
						err,
					)
				}
				var tmpVlanID int16
				if strings.HasPrefix(
					iNet.Name,
					"NMN",
				) {
					tmpVlanID = i.NMNVlanID
				}
				if strings.HasPrefix(
					iNet.Name,
					"HMN",
				) {
					tmpVlanID = i.HMNVlanID
				}
				if tmpVlanID == 0 {
					tmpVlanID = int16(j) + iNet.VlanRange[0]
				}
				tempSubnet := IPV4Subnet{
					CIDR: newSubnet,
					Name: fmt.Sprintf(
						"cabinet_%d",
						i.ID,
					),
					Gateway: Add(
						newSubnet.IP,
						1,
					),
					VlanID: tmpVlanID,
				}
				tempSubnet.UpdateDHCPRange(false)
				myIPv4Subnets = append(
					myIPv4Subnets,
					&tempSubnet,
				)
				if tmpVlanID < minVlan {
					minVlan = tmpVlanID
				}
				if tmpVlanID > maxVlan {
					maxVlan = tmpVlanID
				}
			}
		}
	}
	iNet.VlanRange[0] = minVlan
	iNet.VlanRange[1] = maxVlan
	iNet.Subnets = myIPv4Subnets
	return nil
}

// AllocatedSubnets returns a list of the allocated subnets
func (iNet IPV4Network) AllocatedSubnets() []net.IPNet {
	var myNets []net.IPNet
	for _, v := range iNet.Subnets {
		myNets = append(
			myNets,
			v.CIDR,
		)
	}
	return myNets
}

// AllocatedVlans returns a list of all allocated vlan ids
func (iNet IPV4Network) AllocatedVlans() []int16 {
	var myVlans []int16
	for _, v := range iNet.Subnets {
		if v.VlanID > 0 {
			myVlans = append(
				myVlans,
				v.VlanID,
			)
		}
	}
	return myVlans
}

// AddSubnetbyCIDR allocates a new subnet
func (iNet *IPV4Network) AddSubnetbyCIDR(
	desiredNet net.IPNet, name string, vlanID int16,
) (
	*IPV4Subnet, error,
) {
	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	if Contains(
		*myNet,
		desiredNet,
	) {
		iNet.Subnets = append(
			iNet.Subnets,
			&IPV4Subnet{
				CIDR: desiredNet,
				Name: name,
				Gateway: Add(
					desiredNet.IP,
					1,
				),
				VlanID: vlanID,
			},
		)
		return iNet.Subnets[len(iNet.Subnets)-1], nil
	}
	return &IPV4Subnet{}, fmt.Errorf(
		"subnet %v is not part of %v",
		desiredNet.String(),
		myNet.String(),
	)
}

// AddSubnet allocates a new subnet
func (iNet *IPV4Network) AddSubnet(
	mask net.IPMask, name string, vlanID int16,
) (
	*IPV4Subnet, error,
) {
	var tempSubnet IPV4Subnet
	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	newSubnet, err := Free(
		*myNet,
		mask,
		iNet.AllocatedSubnets(),
	)
	if err != nil {
		return &tempSubnet, err
	}
	iNet.Subnets = append(
		iNet.Subnets,
		&IPV4Subnet{
			CIDR:    newSubnet,
			Name:    name,
			NetName: iNet.Name,
			Gateway: Add(
				newSubnet.IP,
				1,
			),
			VlanID: vlanID,
		},
	)
	return iNet.Subnets[len(iNet.Subnets)-1], nil
}

// AddBiggestSubnet allocates the largest subnet possible within the requested network and mask
func (iNet *IPV4Network) AddBiggestSubnet(
	mask net.IPMask, name string, vlanID int16,
) (
	*IPV4Subnet, error,
) {
	// Try for the largest available and go smaller if needed
	maskSize, _ := mask.Size() // the second output of this function is 32 for ipv4 or 64 for ipv6
	for i := maskSize; i < 29; i++ {
		// log.Printf("Trying to find room for a /%d mask in %v \n", i, iNet.Name)
		newSubnet, err := iNet.AddSubnet(
			net.CIDRMask(
				i,
				32,
			),
			name,
			vlanID,
		)
		if err == nil {
			return newSubnet, nil
		}
	}
	return &IPV4Subnet{}, fmt.Errorf(
		"no room for %v subnet within %v (tried from /%d to /29)",
		name,
		iNet.Name,
		maskSize,
	)
}

// LookUpSubnet returns a subnet by name
func (iNet *IPV4Network) LookUpSubnet(name string) (
	*IPV4Subnet, error,
) {
	var found []*IPV4Subnet
	if len(iNet.Subnets) == 0 {
		return &IPV4Subnet{}, fmt.Errorf(
			"subnet not found \"%v\"",
			name,
		)
	}
	for _, v := range iNet.Subnets {
		if v.Name == name {
			found = append(
				found,
				v,
			)
		}
	}
	if len(found) == 1 {
		return found[0], nil
	}
	if len(found) > 1 {
		// log.Printf("Found %v subnets named %v in the %v network instead of just one \n", len(found), name, iNet.Name)
		return found[0], fmt.Errorf(
			"found %v subnets instead of just one",
			len(found),
		)
	}
	return &IPV4Subnet{}, fmt.Errorf(
		"subnet not found \"%v\"",
		name,
	)
}

// SubnetbyName Return a copy of the subnet by name or a blank subnet if it doesn't exists
func (iNet IPV4Network) SubnetbyName(name string) IPV4Subnet {
	for _, v := range iNet.Subnets {
		if strings.EqualFold(
			v.Name,
			name,
		) {
			return *v
		}
	}
	return IPV4Subnet{}
}

// ReserveEdgeSwitchIPs reserves (n) IP addresses for edge switches
func (iSubnet *IPV4Subnet) ReserveEdgeSwitchIPs(edges []string) {
	for i := 0; i < len(edges); i++ {
		name := fmt.Sprintf(
			"chn-switch-%01d",
			i+1,
		)
		iSubnet.AddReservation(
			name,
			edges[i],
		)
	}
}

// ReserveNetMgmtIPs reserves (n) IP addresses for management networking equipment
func (iSubnet *IPV4Subnet) ReserveNetMgmtIPs(
	spines []string, leafs []string, leafbmcs []string, cdus []string,
) {
	for i := 0; i < len(spines); i++ {
		name := fmt.Sprintf(
			"sw-spine-%03d",
			i+1,
		)
		iSubnet.AddReservation(
			name,
			spines[i],
		)
	}
	for i := 0; i < len(leafs); i++ {
		name := fmt.Sprintf(
			"sw-leaf-%03d",
			i+1,
		)
		iSubnet.AddReservation(
			name,
			leafs[i],
		)
	}
	for i := 0; i < len(leafbmcs); i++ {
		name := fmt.Sprintf(
			"sw-leaf-bmc-%03d",
			i+1,
		)
		iSubnet.AddReservation(
			name,
			leafbmcs[i],
		)
	}
	for i := 0; i < len(cdus); i++ {
		name := fmt.Sprintf(
			"sw-cdu-%03d",
			i+1,
		)
		iSubnet.AddReservation(
			name,
			cdus[i],
		)
	}
}

// ReservedIPs returns a list of IPs already reserved within the subnet
func (iSubnet *IPV4Subnet) ReservedIPs() []net.IP {
	var addresses []net.IP
	for _, v := range iSubnet.IPReservations {
		addresses = append(
			addresses,
			v.IPAddress,
		)
	}
	return addresses
}

// ReservationsByName presents the IPReservations in a map by name
func (iSubnet *IPV4Subnet) ReservationsByName() map[string]IPReservation {
	reservations := make(map[string]IPReservation)
	for _, v := range iSubnet.IPReservations {
		reservations[v.Name] = v
	}
	return reservations
}

// LookupReservation searches the subnet for an IPReservation that matches the name provided
func (iSubnet *IPV4Subnet) LookupReservation(resName string) IPReservation {
	for _, v := range iSubnet.IPReservations {
		if resName == v.Name {
			return v
		}
	}
	return IPReservation{}
}

// TotalIPAddresses returns the number of ip addresses in a subnet see UsableHostAddresses
func (iSubnet *IPV4Subnet) TotalIPAddresses() int {
	maskSize, _ := iSubnet.CIDR.Mask.Size()
	return 2 << uint(31-maskSize)
}

// UsableHostAddresses returns the number of usable ip addresses in a subnet
func (iSubnet *IPV4Subnet) UsableHostAddresses() int {
	maskSize, _ := iSubnet.CIDR.Mask.Size()
	if maskSize == 32 {
		return 1
	} else if maskSize == 31 {
		return 2
	}
	return iSubnet.TotalIPAddresses() - 2
}

// UpdateDHCPRange resets the DHCPStart to exclude all IPReservations
func (iSubnet *IPV4Subnet) UpdateDHCPRange(applySupernetHack bool) {

	myReservedIPs := iSubnet.ReservedIPs()
	if len(myReservedIPs) > iSubnet.UsableHostAddresses() {
		log.Fatalf(
			"Could not create %s subnet in %s. There are %d reservations and only %d usable ip addresses in the subnet %v.",
			iSubnet.FullName,
			iSubnet.NetName,
			len(myReservedIPs),
			iSubnet.UsableHostAddresses(),
			iSubnet.CIDR.String(),
		)
	}

	// Bump the DHCP Start IP past the gateway
	// At least ten IPs are needed, but more if required
	staticLimit := Add(
		iSubnet.CIDR.IP,
		10,
	)
	dynamicLimit := Add(
		iSubnet.CIDR.IP,
		len(iSubnet.IPReservations)+2,
	)
	if IPLessThan(
		dynamicLimit,
		staticLimit,
	) {
		if iSubnet.Name == "uai_macvlan" {
			iSubnet.ReservationStart = staticLimit
		} else {
			iSubnet.DHCPStart = staticLimit
		}
	} else {
		if iSubnet.Name == "uai_macvlan" {
			iSubnet.ReservationStart = dynamicLimit
		} else {
			iSubnet.DHCPStart = dynamicLimit
		}
	}

	if applySupernetHack {
		if iSubnet.Name == "uai_macvlan" {
			iSubnet.ReservationEnd = Add(
				iSubnet.DHCPStart,
				200,
			)
		} else {
			iSubnet.DHCPEnd = Add(
				iSubnet.DHCPStart,
				200,
			) // In this strange world, we can't rely on the broadcast number to be accurate
		}
	} else {
		if iSubnet.Name == "uai_macvlan" {
			iSubnet.ReservationEnd = Add(
				Broadcast(iSubnet.CIDR),
				-1,
			)
		} else {
			iSubnet.DHCPEnd = Add(
				Broadcast(iSubnet.CIDR),
				-1,
			)
		}
	}
}

// AddReservationWithPin adds a new IPv4 reservation to the subnet with the last octet pinned
func (iSubnet *IPV4Subnet) AddReservationWithPin(
	name, comment string, pin uint8,
) *IPReservation {
	// Grab the "floor" of the subnet and alter the last byte to match the pinned byte
	// modulo 4/16 bit ip addresses
	// Worth noting that I could not seem to do this by copying the IP from the struct into a new
	// net.IP struct and modifying only the last byte. I suspected complier error, but as every
	// good programmer knows, it's probably not a compiler error and the time to debug the compiler
	// is not *NOW*
	newIP := make(
		net.IP,
		4,
	)
	if len(iSubnet.CIDR.IP) == 4 {
		newIP[0] = iSubnet.CIDR.IP[0]
		newIP[1] = iSubnet.CIDR.IP[1]
		newIP[2] = iSubnet.CIDR.IP[2]
		newIP[3] = pin
	}
	if len(iSubnet.CIDR.IP) == 16 {
		newIP[0] = iSubnet.CIDR.IP[12]
		newIP[1] = iSubnet.CIDR.IP[13]
		newIP[2] = iSubnet.CIDR.IP[14]
		newIP[3] = pin
	}
	if comment != "" {
		iSubnet.IPReservations = append(
			iSubnet.IPReservations,
			IPReservation{
				IPAddress: newIP,
				Name:      name,
				Comment:   comment,
				Aliases: strings.Split(
					comment,
					",",
				),
			},
		)
	} else {
		iSubnet.IPReservations = append(
			iSubnet.IPReservations,
			IPReservation{
				IPAddress: newIP,
				Name:      name,
			},
		)
	}
	return &iSubnet.IPReservations[len(iSubnet.IPReservations)-1]
}

// AddReservationAlias adds an alias to a reservation if it doesn't already exist
func (iReserv *IPReservation) AddReservationAlias(alias string) {
	if !cli.StringInSlice(
		alias,
		iReserv.Aliases,
	) {
		iReserv.Aliases = append(
			iReserv.Aliases,
			alias,
		)
	}
}

// AddReservation adds a new IP reservation to the subnet
func (iSubnet *IPV4Subnet) AddReservation(name, comment string) *IPReservation {
	myReservedIPs := iSubnet.ReservedIPs()
	// Commenting out this section because the supernet configuration we're using will trigger this all the time and it shouldn't be an error
	// floor := iSubnet.CIDR.IP.Mask(iSubnet.CIDR.Mask)
	// if !floor.Equal(iSubnet.CIDR.IP) {
	// 	log.Printf("VERY BAD - In reservation. CIDR.IP = %v and floor is %v", iSubnet.CIDR.IP.String(), floor)
	// }
	// Start counting from the bottom knowing the gateway is on the bottom
	tempIP := Add(
		iSubnet.CIDR.IP,
		2,
	)
	for {
		for _, v := range myReservedIPs {
			if tempIP.Equal(v) {
				tempIP = Add(
					tempIP,
					1,
				)
			}
		}
		iSubnet.IPReservations = append(
			iSubnet.IPReservations,
			IPReservation{
				IPAddress: tempIP,
				Name:      name,
				Comment:   comment,
			},
		)
		return &iSubnet.IPReservations[len(iSubnet.IPReservations)-1]
	}
}

// AddReservationWithIP adds a reservation with a specific ip address
func (iSubnet *IPV4Subnet) AddReservationWithIP(name, addr, comment string) (
	*IPReservation, error,
) {
	if iSubnet.CIDR.Contains(net.ParseIP(addr)) {
		iSubnet.IPReservations = append(
			iSubnet.IPReservations,
			IPReservation{
				IPAddress: net.ParseIP(addr),
				Name:      name,
				Comment:   comment,
			},
		)
		return &iSubnet.IPReservations[len(iSubnet.IPReservations)-1], nil
	}
	retError := errors.Errorf(
		"Cannot add \"%v\" to %v subnet as %v. %v is not part of %v.",
		name,
		iSubnet.Name,
		addr,
		addr,
		iSubnet.CIDR.String(),
	)

	if len(iSubnet.IPReservations) == 0 {
		return nil, retError
	}

	return &iSubnet.IPReservations[len(iSubnet.IPReservations)-1], retError
}

// GenInterfaceName generates the network interface name for a subnet.
func (iSubnet *IPV4Subnet) GenInterfaceName() error {
	if len(iSubnet.NetName) > 15 {
		return fmt.Errorf(
			"network name [%s] is greater than 15 bytes",
			iSubnet.NetName,
		)
	}
	if iSubnet.NetName == "" {
		return fmt.Errorf(
			"network name [%s] is empty/nil",
			iSubnet.NetName,
		)
	}

	// TODO - Vlans below should come out of sls Defaults, but have circular deps
	if iSubnet.VlanID == 0 || iSubnet.VlanID == DefaultMTLVlan {
		iSubnet.InterfaceName = fmt.Sprintf(
			"%s",
			iSubnet.ParentDevice,
		)
	} else {
		index := 0 // In the future, if we ever need to change this we can handle it with a loop and setting a range in IPV4Subnet.
		iSubnet.InterfaceName = fmt.Sprintf(
			"%s.%s%d",
			iSubnet.ParentDevice,
			strings.ToLower(iSubnet.NetName),
			index,
		)
	}
	return nil
}
