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
	// DefaultNMNLBString is the default LoadBalancer CIDR4 for the NMN
	DefaultNMNLBString = "10.92.100.0/24"
	// DefaultHMNLBString is the default LoadBalancer CIDR4 for the HMN
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
var DefaultLoadBalancerNMN = IPNetwork{
	FullName: "Node Management Network LoadBalancers",
	CIDR4: DefaultNMNLBString,
	Name:     "NMNLB",
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

// DefaultLoadBalancerHMN is a thing we need
var DefaultLoadBalancerHMN = IPNetwork{
	FullName: "Hardware Management Network LoadBalancers",
	CIDR4: DefaultHMNLBString,
	Name:     "HMNLB",
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

// DefaultBICAN is the default structure for templating the initial BICAN toggle - CMN
var DefaultBICAN = IPNetwork{
	FullName:           "SystemDefaultRoute points the network name of the default route",
	CIDR4: "0.0.0.0/0",
	Name:               "BICAN",
	VlanRange:          []int16{0},
	MTU:                9000,
	NetType:            "ethernet",
	Comment:            "",
	SystemDefaultRoute: "",
}

// DefaultHSN is the default structure for templating initial HSN configuration
var DefaultHSN = IPNetwork{
	FullName: "High Speed Network",
	CIDR4: DefaultHSNString,
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
var DefaultCMN = IPNetwork{
	FullName:     "Customer Management Network",
	CIDR4: DefaultCMNString,
	Name:         "CMN",
	VlanRange:    []int16{DefaultCMNVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultCAN is the default structure for templating initial CAN configuration
var DefaultCAN = IPNetwork{
	FullName:     "Customer Access Network",
	CIDR4: DefaultCANString,
	Name:         "CAN",
	VlanRange:    []int16{DefaultCANVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultCHN is the default structure for templating initial CHN configuration
var DefaultCHN = IPNetwork{
	FullName:     "Customer High-Speed Network",
	CIDR4: DefaultCHNString,
	Name:         "CHN",
	VlanRange:    []int16{DefaultCHNVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultHMN is the default structure for templating initial HMN configuration
var DefaultHMN = IPNetwork{
	FullName:     "Hardware Management Network",
	CIDR4: DefaultHMNString,
	Name:         "HMN",
	VlanRange:    []int16{DefaultHMNVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultNMN is the default structure for templating initial NMN configuration
var DefaultNMN = IPNetwork{
	FullName:     "Node Management Network",
	CIDR4: DefaultNMNString,
	Name:         "NMN",
	VlanRange:    []int16{DefaultNMNVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "",
	ParentDevice: "bond0",
}

// DefaultMTL is the default structure for templating initial MTL configuration
var DefaultMTL = IPNetwork{
	FullName:     "Provisioning Network (untagged)",
	CIDR4: DefaultMTLString,
	Name:         "MTL",
	VlanRange:    []int16{DefaultMTLVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "This network is only valid for the NCNs",
	ParentDevice: "bond0",
}

// IPNetwork is a type for managing IPv4 Networks
type IPNetwork struct {
	FullName           string                `yaml:"full_name"`
	CIDR4   string      `yaml:"cidr4"`
	Subnets []*IPSubnet `yaml:"subnets"`
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

// IPSubnet is a type for managing IPv4 Subnets
type IPSubnet struct {
	FullName         string          `yaml:"full_name" form:"full_name" mapstructure:"full_name"`
	CIDR4    net.IPNet `yaml:"cidr4"`
	IPReservations   []IPReservation `yaml:"ip_reservations"`
	Name             string          `yaml:"name" form:"name" mapstructure:"name"`
	NetName          string          `yaml:"net-name"`
	VlanID           int16           `yaml:"vlan_id" form:"vlan_id" mapstructure:"vlan_id"`
	Comment          string          `yaml:"comment"`
	Gateway4 net.IP    `yaml:"gateway4"`
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
func (iNet *IPNetwork) ApplySupernetHack() {
	// Replace the gateway and netmask on the to better support the 1.3 network switch configuration
	// *** This is a HACK ***
	_, superNet, err := net.ParseCIDR(iNet.CIDR4)
	if err != nil {
		log.Fatal(
			"Couldn't parse the CIDR4 for ",
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
			tempSubnet.Gateway4 = Add(
				superNet.IP,
				1,
			)
			tempSubnet.CIDR4.Mask = superNet.Mask
		}
	}
}

// GenSubnets subdivides a network into a set of subnets
func (iNet *IPNetwork) GenSubnets(
	cabinetDetails []sls.CabinetGroupDetail, mask net.IPMask, cabinetFilter sls.CabinetFilterFunc,
) error {
	// log.Printf("Generating Subnets for %s\ncabinetType: %v,\n", iNet.Name, cabinetType)
	_, myNet, _ := net.ParseCIDR(iNet.CIDR4)
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
				tempSubnet := IPSubnet{
					CIDR4: newSubnet,
					Name: fmt.Sprintf(
						"cabinet_%d",
						i.ID,
					),
					Gateway4: Add(
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
func (iNet IPNetwork) AllocatedSubnets() []net.IPNet {
	var myNets []net.IPNet
	for _, v := range iNet.Subnets {
		myNets = append(
			myNets,
			v.CIDR4,
		)
	}
	return myNets
}

// AllocatedVlans returns a list of all allocated vlan ids
func (iNet IPNetwork) AllocatedVlans() []int16 {
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
func (iNet *IPNetwork) AddSubnetbyCIDR(
	desiredNet net.IPNet, name string, vlanID int16,
) (
	*IPSubnet, error,
) {
	_, myNet, _ := net.ParseCIDR(iNet.CIDR4)
	if Contains(
		*myNet,
		desiredNet,
	) {
		iNet.Subnets = append(
			iNet.Subnets,
			&IPSubnet{
				CIDR4: desiredNet,
				Name:  name,
				Gateway4: Add(
					desiredNet.IP,
					1,
				),
				VlanID: vlanID,
			},
		)
		return iNet.Subnets[len(iNet.Subnets)-1], nil
	}
	return &IPSubnet{}, fmt.Errorf(
		"subnet %v is not part of %v",
		desiredNet.String(),
		myNet.String(),
	)
}

// AddSubnet allocates a new subnet
func (iNet *IPNetwork) AddSubnet(
	mask net.IPMask, name string, vlanID int16,
) (
	*IPSubnet, error,
) {
	var tempSubnet IPSubnet
	_, myNet, _ := net.ParseCIDR(iNet.CIDR4)
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
		&IPSubnet{
			CIDR4: newSubnet,
			Name:    name,
			NetName: iNet.Name,
			Gateway4: Add(
				newSubnet.IP,
				1,
			),
			VlanID: vlanID,
		},
	)
	return iNet.Subnets[len(iNet.Subnets)-1], nil
}

// AddBiggestSubnet allocates the largest subnet possible within the requested network and mask
func (iNet *IPNetwork) AddBiggestSubnet(
	mask net.IPMask, name string, vlanID int16,
) (
	*IPSubnet, error,
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
	return &IPSubnet{}, fmt.Errorf(
		"no room for %v subnet within %v (tried from /%d to /29)",
		name,
		iNet.Name,
		maskSize,
	)
}

// LookUpSubnet returns a subnet by name
func (iNet *IPNetwork) LookUpSubnet(name string) (
	*IPSubnet, error,
) {
	var found []*IPSubnet
	if len(iNet.Subnets) == 0 {
		return &IPSubnet{}, fmt.Errorf(
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
	return &IPSubnet{}, fmt.Errorf(
		"subnet not found \"%v\"",
		name,
	)
}

// SubnetbyName Return a copy of the subnet by name or a blank subnet if it doesn't exists
func (iNet IPNetwork) SubnetbyName(name string) IPSubnet {
	for _, v := range iNet.Subnets {
		if strings.EqualFold(
			v.Name,
			name,
		) {
			return *v
		}
	}
	return IPSubnet{}
}

// ReserveEdgeSwitchIPs reserves (n) IP addresses for edge switches
func (iSubnet *IPSubnet) ReserveEdgeSwitchIPs(edges []string) {
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
func (iSubnet *IPSubnet) ReserveNetMgmtIPs(
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
func (iSubnet *IPSubnet) ReservedIPs() []net.IP {
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
func (iSubnet *IPSubnet) ReservationsByName() map[string]IPReservation {
	reservations := make(map[string]IPReservation)
	for _, v := range iSubnet.IPReservations {
		reservations[v.Name] = v
	}
	return reservations
}

// LookupReservation searches the subnet for an IPReservation that matches the name provided
func (iSubnet *IPSubnet) LookupReservation(resName string) IPReservation {
	for _, v := range iSubnet.IPReservations {
		if resName == v.Name {
			return v
		}
	}
	return IPReservation{}
}

// TotalIPAddresses returns the number of ip addresses in a subnet see UsableHostAddresses
func (iSubnet *IPSubnet) TotalIPAddresses() int {
	maskSize, _ := iSubnet.CIDR4.Mask.Size()
	return 2 << uint(31-maskSize)
}

// UsableHostAddresses returns the number of usable ip addresses in a subnet
func (iSubnet *IPSubnet) UsableHostAddresses() int {
	maskSize, _ := iSubnet.CIDR4.Mask.Size()
	if maskSize == 32 {
		return 1
	} else if maskSize == 31 {
		return 2
	}
	return iSubnet.TotalIPAddresses() - 2
}

// UpdateDHCPRange resets the DHCPStart to exclude all IPReservations
func (iSubnet *IPSubnet) UpdateDHCPRange(applySupernetHack bool) {

	myReservedIPs := iSubnet.ReservedIPs()
	if len(myReservedIPs) > iSubnet.UsableHostAddresses() {
		log.Fatalf(
			"Could not create %s subnet in %s. There are %d reservations and only %d usable ip addresses in the subnet %v.",
			iSubnet.FullName,
			iSubnet.NetName,
			len(myReservedIPs),
			iSubnet.UsableHostAddresses(),
			iSubnet.CIDR4.String(),
		)
	}

	// Bump the DHCP Start IP past the gateway
	// At least ten IPs are needed, but more if required
	staticLimit := Add(
		iSubnet.CIDR4.IP,
		10,
	)
	dynamicLimit := Add(
		iSubnet.CIDR4.IP,
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
				Broadcast(iSubnet.CIDR4),
				-1,
			)
		} else {
			iSubnet.DHCPEnd = Add(
				Broadcast(iSubnet.CIDR4),
				-1,
			)
		}
	}
}

// AddReservationWithPin adds a new IPv4 reservation to the subnet with the last octet pinned
func (iSubnet *IPSubnet) AddReservationWithPin(
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
	if len(iSubnet.CIDR4.IP) == 4 {
		newIP[0] = iSubnet.CIDR4.IP[0]
		newIP[1] = iSubnet.CIDR4.IP[1]
		newIP[2] = iSubnet.CIDR4.IP[2]
		newIP[3] = pin
	}
	if len(iSubnet.CIDR4.IP) == 16 {
		newIP[0] = iSubnet.CIDR4.IP[12]
		newIP[1] = iSubnet.CIDR4.IP[13]
		newIP[2] = iSubnet.CIDR4.IP[14]
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
func (iSubnet *IPSubnet) AddReservation(name, comment string) *IPReservation {
	myReservedIPs := iSubnet.ReservedIPs()
	// Commenting out this section because the supernet configuration we're using will trigger this all the time and it shouldn't be an error
	// floor := iSubnet.CIDR4.IP.Mask(iSubnet.CIDR4.Mask)
	// if !floor.Equal(iSubnet.CIDR4.IP) {
	// 	log.Printf("VERY BAD - In reservation. CIDR4.IP = %v and floor is %v", iSubnet.CIDR4.IP.String(), floor)
	// }
	// Start counting from the bottom knowing the gateway is on the bottom
	tempIP := Add(
		iSubnet.CIDR4.IP,
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
func (iSubnet *IPSubnet) AddReservationWithIP(name, addr, comment string) (
	*IPReservation, error,
) {
	if iSubnet.CIDR4.Contains(net.ParseIP(addr)) {
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
		iSubnet.CIDR4.String(),
	)

	if len(iSubnet.IPReservations) == 0 {
		return nil, retError
	}

	return &iSubnet.IPReservations[len(iSubnet.IPReservations)-1], retError
}

// GenInterfaceName generates the network interface name for a subnet.
func (iSubnet *IPSubnet) GenInterfaceName() error {
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
		index := 0 // In the future, if we ever need to change this we can handle it with a loop and setting a range in IPSubnet.
		iSubnet.InterfaceName = fmt.Sprintf(
			"%s.%s%d",
			iSubnet.ParentDevice,
			strings.ToLower(iSubnet.NetName),
			index,
		)
	}
	return nil
}
