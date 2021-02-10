/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

import (
	"net"
	"strings"
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
	DefaultHMNString = "10.254.0.0/17"
	// DefaultHMNVlan is the default HMN Bootstrap Vlan
	DefaultHMNVlan = 4
	// DefaultHMNMTNString is the default HMN Network for Mountain Cabinets with Grouped Configuration
	DefaultHMNMTNString = "10.100.0.0/17"
	// DefaultHMNRVRString is the default HMN Network for River Cabinets with Grouped Configuration
	DefaultHMNRVRString = "10.107.0.0/17"
	// DefaultNMNString is the Default NMN String (vlan002)
	DefaultNMNString = "10.252.0.0/17"
	// DefaultNMNVlan is the default NMN Bootstrap Vlan
	DefaultNMNVlan = 2
	// DefaultNMNMTNString is the default NMN Network for Mountain Cabinets with Grouped Configuration
	DefaultNMNMTNString = "10.104.0.0/17"
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
	// DefaultCANString is the Default CAN String (vlan007)
	DefaultCANString = "10.102.11.0/24"
	// DefaultCANPoolString is the default pool for CAN addresses
	DefaultCANPoolString = "10.102.11.128/25"
	// DefaultCANStaticString is the default pool for Static CAN addresses
	DefaultCANStaticString = "10.102.11.112/28"
	// DefaultCANVlan is the default CAN Bootstrap Vlan
	DefaultCANVlan = 7
	// DefaultMTLString is the Default MTL String (bond0 interface)
	DefaultMTLString = "10.1.1.0/16"
)

// ValidNetNames is the list of strings that enumerate valid main network names
var ValidNetNames = []string{"HMN", "NMN", "CAN", "MTL", "HMN_RVR", "HMN_MTN", "NMN_RVR", "NMN_MTN"}

// ValidCabinetTypes is the list of strings that enumerate valid cabinet types
var ValidCabinetTypes = []string{"mountain", "river", "hill"}

// InstallerDefaults holds all of our defaults
var InstallerDefaults = SystemConfig{
	SystemName:      "sn-2024",
	SiteDomain:      "dev.cray.com",
	NtpPoolHostname: "time.nist.gov",
	RpmRegistry:     "https://packages.nmn/repository/shasta-master",
	V2Registry:      "https://registry.nmn/",
	Install: InstallConfig{
		NCN:                 "ncn-m001",
		NCNBondMembers:      "p1p1,p1p2",
		SiteNIC:             "em1",
		SitePrefix:          "/20",
		CephCephfsImage:     "dtr.dev.cray.com/cray/cray-cephfs-provisioner:0.1.0-nautilus-1.3",
		CephRBDImage:        "dtr.dev.cray.com/cray/cray-rbd-provisioner:0.1.0-nautilus-1.3",
		ChartRepo:           "http://helmrepo.dev.cray.com:8080",
		DockerImageRegistry: "dtr.dev.cray.com",
	},
}

// IPNetfromCIDRString converts from a string to an net.IPNet struct
func IPNetfromCIDRString(mynet string) *net.IPNet {
	_, ipnet, _ := net.ParseCIDR(mynet)
	return ipnet
}

// DefaultCabinetMask is the default subnet mask for each cabinet
var DefaultCabinetMask = net.CIDRMask(22, 32)

// DefaultNetworkingHardwareMask is the default subnet mask for a subnet that contains all networking hardware
var DefaultNetworkingHardwareMask = net.CIDRMask(24, 32)

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

// GenDefaultHMN returns the default structure for templating initial HMN configuration
func GenDefaultHMN() IPV4Network {
	return IPV4Network{
		FullName:  "Hardware Management Network",
		CIDR:      DefaultHMNString,
		Name:      "HMN",
		VlanRange: []int16{100, 356},
		MTU:       9000,
		NetType:   "ethernet",
		Comment:   "",
	}
}

// GenDefaultNMN returns the default structure for templating initial NMN configuration
func GenDefaultNMN() IPV4Network {
	return IPV4Network{
		FullName:  "Node Management Network",
		CIDR:      DefaultNMNString,
		Name:      "NMN",
		VlanRange: []int16{357, 612},
		MTU:       9000,
		NetType:   "ethernet",
		Comment:   "",
	}
}

// DefaultHSN is the default structure for templating initial HSN configuration
var DefaultHSN = IPV4Network{
	FullName:  "High Speed Network",
	CIDR:      DefaultHSNString,
	Name:      "HSN",
	VlanRange: []int16{613, 868},
	MTU:       9000,
	NetType:   "slingshot10",
	Comment:   "",
}

// DefaultCAN is the default structure for templating initial CAN configuration
var DefaultCAN = IPV4Network{
	FullName:  "Customer Access Network",
	CIDR:      DefaultCANString,
	Name:      "CAN",
	VlanRange: []int16{11, 35},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "",
}

// DefaultMTL is the default structure for templating initial MTL configuration
var DefaultMTL = IPV4Network{
	FullName:  "Provisioning Network (untagged)",
	CIDR:      DefaultMTLString,
	Name:      "MTL",
	VlanRange: []int16{36, 40},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "This network is only valid for the NCNs",
}

// GenDefaultHMNConfig is the set of defaults for mapping the HMN
func GenDefaultHMNConfig() NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        GenDefaultHMN(),
		SubdivideByCabinet:              false,
		GroupNetworksByCabinetType:      true,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: true,
		SuperNetHack:                    true,
		IncludeUAISubnet:                false,
		CabinetCIDR:                     DefaultCabinetMask,
		NetworkingHardwareNetmask:       DefaultNetworkingHardwareMask,
		DesiredBootstrapDHCPMask:        net.CIDRMask(24, 32),
	}
}

// GenDefaultNMNConfig returns the set of defaults for mapping the NMN
func GenDefaultNMNConfig() NetworkLayoutConfiguration {
	return NetworkLayoutConfiguration{
		Template:                        GenDefaultNMN(),
		SubdivideByCabinet:              false,
		GroupNetworksByCabinetType:      true,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: true,
		SuperNetHack:                    true,
		IncludeUAISubnet:                true,
		CabinetCIDR:                     DefaultCabinetMask,
		NetworkingHardwareNetmask:       DefaultNetworkingHardwareMask,
		DesiredBootstrapDHCPMask:        net.CIDRMask(24, 32),
	}
}

// GenDefaultHSNConfig returns the set of defaults for mapping the HSN
func GenDefaultHSNConfig() NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        DefaultHSN,
		SubdivideByCabinet:              false,
		IncludeBootstrapDHCP:            false,
		IncludeNetworkingHardwareSubnet: false,
		IncludeUAISubnet:                false,
	}
}

// GenDefaultCANConfig returns the set of defaults for mapping the CAN
func GenDefaultCANConfig() NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        DefaultCAN,
		SubdivideByCabinet:              false,
		SuperNetHack:                    false,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: false,
		IncludeUAISubnet:                false,
		DesiredBootstrapDHCPMask:        net.CIDRMask(24, 32),
	}
}

// GenDefaultMTLConfig returns the set of defaults for mapping the MTL
func GenDefaultMTLConfig() NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        DefaultMTL,
		SubdivideByCabinet:              false,
		SuperNetHack:                    true,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: true,
		IncludeUAISubnet:                false,
		NetworkingHardwareNetmask:       DefaultNetworkingHardwareMask,
		DesiredBootstrapDHCPMask:        net.CIDRMask(24, 32),
	}
}

// DefaultManifestURL is the git URL for downloading the loftsman manifests for packaging
var DefaultManifestURL string = "ssh://git@stash.us.cray.com:7999/shasta-cfg/stable.git"

// DefaultUAISubnetReservations is the map of dns names and aliases
var DefaultUAISubnetReservations = map[string][]string{
	"uai_macvlan_bridge": {"uai-macvlan-bridge"},
	"slurmctld_service":  {"slurmctld-service", "slurmctld-service-nmn"},
	"slurmdbd_service":   {"slurmdbd-service", "slurmdbd-service-nmn"},
	"pbs_service":        {"pbs-service", "pbs-service-nmn"},
	"pbs_comm_service":   {"pbs-comm-service", "pbs-comm-service-nmn"},
}

// PinnedReservation is a simple struct to work with our abomination of a PinnedMetalLBReservations
type PinnedReservation struct {
	IPByte  uint8
	Aliases []string
}

// PinnedMetalLBReservations is the map of dns names and aliases with the
// required final octet of th ip address
// *** This structure is only necessary to pin ip addresses as we shift from 1.3 to 1.4 ***
// *** *** *** To anyone editing this code in the future, PLEASE DON'T MAKE IT BETTER *** *** ***
// *** *** *** This code is written to be thrown away with a fully dynamic ip addressing scheme *** *** ***
//
var PinnedMetalLBReservations = map[string]PinnedReservation{
	"istio-ingressgateway":       {71, strings.Split("api-gw-service api-gw-service-nmn.local packages registry spire.local api_gw_service registry.local packages packages.local spire", " ")},
	"istio-ingressgateway-local": {81, []string{"api-gw-service.local"}},
	"rsyslog-aggregator":         {72, []string{"rsyslog-agg-service"}},
	"cray-tftp":                  {60, []string{"tftp-service"}},
	"unbound":                    {74, []string{"unbound"}},
	"docker-registry":            {73, []string{"docker_registry_service"}},
	"slingshot-kafka":            {75, []string{"slingshot_kafka_extern_service"}},
}