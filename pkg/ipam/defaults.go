package ipam

import (
	"net"
)

var _, NmnCIDR, _ = net.ParseCIDR("10.242.0.0/17")
var _, HmnCIDR, _ = net.ParseCIDR("10.254.0.0/17")
var _, HsnCIDR, _ = net.ParseCIDR("10.253.0.0/16")
var _, CanCIDR, _ = net.ParseCIDR("192.168.20.0/24")
var _, MtlCIDR, _ = net.ParseCIDR("192.168.1.0/24")

var DefaultNodeManagementNetwork = ShastaIPV4Network{
	FullName: "Node Management Network",
	CIDR:     *NmnCIDR,
	Name:     "nmn",
	Vlan:     5,
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

var DefaultHardwareManagementNetwork = ShastaIPV4Network{
	FullName: "Hardware Management Network",
	CIDR:     *HmnCIDR,
	Name:     "hmn",
	Vlan:     6,
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

var DefaultHighSpeedNetwork = ShastaIPV4Network{
	FullName: "High Speed Network",
	CIDR:     *HsnCIDR,
	Name:     "hsn",
	Vlan:     7,
	MTU:      9000,
	NetType:  "slingshot10",
	Comment:  "",
}

var DefaultCanNetwork = ShastaIPV4Network{
	FullName: "Customer Access Network",
	CIDR:     *CanCIDR,
	Name:     "can",
	Vlan:     8,
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

var DefaultMTLNetwork = ShastaIPV4Network{
	FullName: "Untagged MTL Network",
	CIDR:     *MtlCIDR,
	Name:     "mtl",
	Vlan:     8,
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}
