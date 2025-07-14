/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

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
	"net"
	"strings"
)

// BootstrapSwitchMetadata is a type that matches the switch_metadata.csv file as
// Switch Xname, Type
// The type can be CDU, Spine, Leaf, or LeafBMC
type BootstrapSwitchMetadata struct {
	Xname string `json:"xname" csv:"Switch Xname"`
	Type  string `json:"type" csv:"Type"`
}

// DefaultApplicationNodePrefixes is the list of default Application node prefixes, for source column in the hmn_connections.json
var DefaultApplicationNodePrefixes = []string{
	"uan",
	"gn",
	"ln",
}

// DefaultApplicationNodeSubroles is the default prefix<->subrole mapping for application node subroles, these can be overridden via ApplicationNodeConfig
var DefaultApplicationNodeSubroles = map[string]string{
	"uan": "UAN",
	"ln":  "UAN", // Nodes with the ln prefix are also UAN nodes
	"gn":  "Gateway4",
}

// SubrolePlaceHolder is the placeholder used to indicate that a prefix has no subrole mapping in ApplicationNodeConfig.
const SubrolePlaceHolder = "~fixme~"

// ValidNetNames is the list of strings that enumerate valid main network names
var ValidNetNames = []string{
	"BICAN",
	"CAN",
	"CHN",
	"CMN",
	"HMN",
	"HMN_MTN",
	"HMN_RVR",
	"MTL",
	"NMN",
	"NMN_MTN",
	"NMN_RVR",
}

// DefaultUAISubnetReservations is the map of dns names and aliases
var DefaultUAISubnetReservations = map[string][]string{
	"uai_nmn_blackhole": {"uai-nmn-blackhole"},
	"slurmctld_service": {
		"slurmctld-service",
		"slurmctld-service-nmn",
	},
	"slurmdbd_service": {
		"slurmdbd-service",
		"slurmdbd-service-nmn",
	},
	"pbs_service": {
		"pbs-service",
		"pbs-service-nmn",
	},
	"pbs_comm_service": {
		"pbs-comm-service",
		"pbs-comm-service-nmn",
	},
}

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
	DefaultMTLString = "10.1.0.0/16"
	// MinVLAN is the minimum VLAN bound for VLANs.
	MinVLAN = 0
	// FirstVLAN is the first VLAN (often "native" VLAN).
	FirstVLAN = 1
	// MaxVLAN is the maximum VLAN we plan to use.
	MaxVLAN = 4096
	// MaxUsableVLAN is the maximum, usable VLAN ID.
	MaxUsableVLAN = 4095
	// DefaultIPv4Block is a recommended default block size (<=256 hosts).
	DefaultIPv4Block = 24
	// DefaultIPv6Block is a recommended default block size (<=256 hosts).
	DefaultIPv6Block = 120
	// SmallestIPv4Block is the smallest IPv4 subnet that we'll allocate (<10 hosts).
	SmallestIPv4Block = 29
	// SmallestIPv6Block is the smallest IPv6 subnet that we'll allocate (<10 hosts).
	SmallestIPv6Block = 125
)

// DefaultCabinetMask is the default subnet mask for each Cabinet
var DefaultCabinetMask = net.CIDRMask(
	22,
	IPv4Size,
)

// DefaultNetworkingHardwareMask is the default subnet mask for a subnet that contains all networking hardware
var DefaultNetworkingHardwareMask = net.CIDRMask(
	24,
	IPv4Size,
)

// DefaultLoadBalancerNMN is a thing we need
var DefaultLoadBalancerNMN = IPNetwork{
	FullName: "Node Management Network LoadBalancers",
	CIDR4:    DefaultNMNLBString,
	Name:     "NMNLB",
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

// DefaultLoadBalancerHMN is a thing we need
var DefaultLoadBalancerHMN = IPNetwork{
	FullName: "Hardware Management Network LoadBalancers",
	CIDR4:    DefaultHMNLBString,
	Name:     "HMNLB",
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

// DefaultBICAN is the default structure for templating the initial BICAN toggle - CMN
var DefaultBICAN = IPNetwork{
	FullName:           "SystemDefaultRoute points the network name of the default route",
	CIDR4:              "0.0.0.0/0",
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
	CIDR4:    DefaultHSNString,
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
	CIDR4:        DefaultCMNString,
	CIDR6:        "",
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
	CIDR4:        DefaultCANString,
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
	CIDR4:        DefaultCHNString,
	CIDR6:        "",
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
	CIDR4:        DefaultHMNString,
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
	CIDR4:        DefaultNMNString,
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
	CIDR4:        DefaultMTLString,
	Name:         "MTL",
	VlanRange:    []int16{DefaultMTLVlan},
	MTU:          9000,
	NetType:      "ethernet",
	Comment:      "This network is only valid for the NCNs",
	ParentDevice: "bond0",
}

// PinnedMetalLBReservations is the map of dns names and aliases with the
// required final octet of th ip address
// *** This structure is only necessary to pin ip addresses as we shift from 1.3 to 1.4 ***
// *** *** *** To anyone editing this code in the future, PLEASE DON'T MAKE IT BETTER *** *** ***
// *** *** *** This code is written to be thrown away with a fully dynamic ip addressing scheme *** *** ***
var PinnedMetalLBReservations = map[string]PinnedReservation{
	"istio-ingressgateway": {
		71,
		strings.Split(
			"api-gw-service api-gw-service-nmn.local packages registry spire.local api_gw_service registry.local packages packages.local spire",
			" ",
		),
	},
	"istio-ingressgateway-local": {
		81,
		[]string{"api-gw-service.local"},
	},
	"rsyslog-aggregator": {
		72,
		[]string{"rsyslog-agg-service"},
	},
	"cray-tftp": {
		60,
		[]string{"tftp-service"},
	},
	"unbound": {
		225,
		[]string{"unbound"},
	},
	"docker-registry": {
		73,
		[]string{"docker_registry_service"},
	},
}
