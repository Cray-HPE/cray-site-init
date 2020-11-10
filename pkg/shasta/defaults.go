/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"net"
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

const (
	// DefaultHMNString is the Default HMN String (vlan004)
	DefaultHMNString = "10.254.1.1/16"
	// DefaultHMNVlan is the default HMN Bootstrap Vlan
	DefaultHMNVlan = 4
	// DefaultNMNString is the Default NMN String (vlan002)
	DefaultNMNString = "10.252.1.0/16"
	// DefaultNMNVlan is the default NMN Bootstrap Vlan
	DefaultNMNVlan = 2
	// DefaultMacVlanString is the default Macvlan cidr (shares vlan with NMN)
	DefaultMacVlanString = "10.252.124.0/23"
	// DefaultHSNString is the Default HSN String
	DefaultHSNString = "10.250.0.0/16"
	// DefaultCANString is the Default CAN String (vlan007)
	DefaultCANString = "10.102.9.0/24"
	// DefaultCANVlan is the default CAN Bootstrap Vlan
	DefaultCANVlan = 7
	// DefaultMTLString is the Default MTL String (bond0 interface)
	DefaultMTLString = "10.1.1.0/16"
)

var slash24 = net.IPv4Mask(255, 255, 255, 0)
var slash23 = net.IPv4Mask(255, 255, 254, 0)
var slash17 = net.IPv4Mask(255, 255, 128, 0)
var slash16 = net.IPv4Mask(255, 255, 0, 0)

// IPNetfromCIDRString converts from a string to an net.IPNet struct
func IPNetfromCIDRString(mynet string) *net.IPNet {
	_, ipnet, _ := net.ParseCIDR(mynet)
	return ipnet
}

// DefaultHMN is the default structure for templating initial HMN configuration
var DefaultHMN = IPV4Network{
	FullName:  "Hardware Management Network",
	CIDR:      DefaultHMNString,
	Name:      "HMN",
	VlanRange: []int16{11, 40},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "",
}

// DefaultNMN is the default structure for templating initial NMN configuration
var DefaultNMN = IPV4Network{
	FullName:  "Node Management Network",
	CIDR:      DefaultNMNString,
	Name:      "NMN",
	VlanRange: []int16{41, 70},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "",
}

// DefaultHSN is the default structure for templating initial HSN configuration
var DefaultHSN = IPV4Network{
	FullName:  "High Speed Network",
	CIDR:      DefaultHSNString,
	Name:      "HSN",
	VlanRange: []int16{71, 90},
	MTU:       9000,
	NetType:   "slingshot10",
	Comment:   "",
}

// DefaultCAN is the default structure for templating initial CAN configuration
var DefaultCAN = IPV4Network{
	FullName:  "Customer Access Network",
	CIDR:      DefaultCANString,
	Name:      "CAN",
	VlanRange: []int16{91, 120},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "",
}

// DefaultMTL is the default structure for templating initial MTL configuration
var DefaultMTL = IPV4Network{
	FullName:  "Provisioning Network (untagged)",
	CIDR:      DefaultMTLString,
	Name:      "MTL",
	VlanRange: []int16{121, 150},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "This network is only valid for the NCNs",
}

// DefaultRootPW is the default root password
var DefaultRootPW = PasswordCredential{
	Username: "root",
	Password: "changem3",
}

// DefaultBMCPW is the default root password
var DefaultBMCPW = PasswordCredential{
	Username: "root",
	Password: "changem3",
}

// DefaultNetPW is the default root password
var DefaultNetPW = PasswordCredential{
	Username: "root",
	Password: "changem3",
}

// DefaultManifestURL is the git URL for downloading the loftsman manifests for packaging
var DefaultManifestURL string = "ssh://git@stash.us.cray.com:7999/shasta-cfg/stable.git"
