package ipam

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

var slash24 = net.IPv4Mask(255, 255, 255, 0)
var slash23 = net.IPv4Mask(255, 255, 254, 0)
var slash17 = net.IPv4Mask(255, 255, 128, 0)
var slash16 = net.IPv4Mask(255, 255, 0, 0)

var NmnCIDR = net.IPNet{
	IP:   net.ParseIP("10.242.0.0"),
	Mask: slash17,
}

var HmnCIDR = net.IPNet{
	IP:   net.ParseIP("10.254.0.0"),
	Mask: slash17,
}

var HsnCIDR = net.IPNet{
	IP:   net.ParseIP("10.253.0.0"),
	Mask: slash16,
}

var CanCIDR = net.IPNet{
	IP:   net.ParseIP("192.168.20.0/24"),
	Mask: slash24,
}

var MtlCIDR = net.IPNet{
	IP:   net.ParseIP("192.168.1.0/24"),
	Mask: slash24,
}

var DefaultNodeManagementNetwork = ShastaIPV4Network{
	FullName: "Node Management Network",
	CIDR:     &NmnCIDR,
	Name:     "nmn",
	Vlan:     5,
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

var DefaultHardwareManagementNetwork = ShastaIPV4Network{
	FullName: "Hardware Management Network",
	CIDR:     &HmnCIDR,
	Name:     "hmn",
	Vlan:     6,
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

var DefaultHighSpeedNetwork = ShastaIPV4Network{
	FullName: "High Speed Network",
	CIDR:     &HsnCIDR,
	Name:     "hsn",
	Vlan:     7,
	MTU:      9000,
	NetType:  "slingshot10",
	Comment:  "",
}

var DefaultCanNetwork = ShastaIPV4Network{
	FullName: "Customer Access Network",
	CIDR:     &CanCIDR,
	Name:     "can",
	Vlan:     8,
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}

var DefaultMTLNetwork = ShastaIPV4Network{
	FullName: "Untagged MTL Network",
	CIDR:     &MtlCIDR,
	Name:     "mtl",
	Vlan:     8,
	MTU:      9000,
	NetType:  "ethernet",
	Comment:  "",
}
