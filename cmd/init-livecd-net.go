/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/ipam"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
)

// BuildLiveCDNetworks creates an array of IPv4 Networks based on the supplied system configuration
func BuildLiveCDNetworks(conf shasta.SystemConfig, v *viper.Viper) (map[string]*shasta.IPV4Network, error) {
	// our primitive ipam uses the number of cabinets to lay out a network for each one.
	// It is per-cabinet type which is pretty annoying, but here we are.

	cabinetDetails := buildCabinetDetails(v)

	var networkMap = make(map[string]*shasta.IPV4Network)

	// Start the NMN with out defaults
	tempNMN := shasta.DefaultNMN
	// Update the CIDR from flags/viper
	tempNMN.CIDR = v.GetString("nmn-cidr")
	// Add a /25 for the Load Balancers
	pool, err := tempNMN.AddSubnet(net.CIDRMask(25, 32), "nmn_metallb_address_pool", int16(v.GetInt("nmn-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	pool.AddReservation("api_gateway", "")
	// Add a /26 for bootstrap dhcp
	subnet, err := tempNMN.AddSubnet(net.CIDRMask(26, 32), "bootstrap_dhcp", int16(v.GetInt("nmn-bootstrap-vlan")))
	subnet.ReserveNetMgmtIPs(v.GetInt("management-net-ips"), strings.Split(v.GetString("spine-switch-xnames"), ","), strings.Split(v.GetString("leaf-switch-xnames"), ","))
	subnet.AddReservation("kubeapi-vip", "k8s-virtual-ip")
	subnet.AddReservation("rgw-vip", "rgw-virtual-ip")
	subnet.DHCPStart = ipam.Add(subnet.CIDR.IP, len(subnet.IPReservations))
	subnet.DHCPEnd = ipam.Add(ipam.Broadcast(subnet.CIDR), -1)
	// Divide the network into an appropriate number of subnets
	tempNMN.GenSubnets(cabinetDetails, net.CIDRMask(22, 32), v.GetInt("management-net-ips"), v.GetString("spine-switch-xnames"), v.GetString("leaf-switch-xnames"))
	networkMap["NMN"] = &tempNMN

	// Start the HMN with out defaults
	tempHMN := shasta.DefaultHMN
	// Update the CIDR from flags/viper
	tempHMN.CIDR = v.GetString("hmn-cidr")
	// Add a /25 for the Load Balancers
	pool, err = tempHMN.AddSubnet(net.CIDRMask(25, 32), "hmn_metallb_address_pool", int16(v.GetInt("hmn-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	pool.AddReservation("api_gateway", "")

	// Add a /26 for bootstrap dhcp
	subnet, err = tempHMN.AddSubnet(net.CIDRMask(26, 32), "bootstrap_dhcp", int16(v.GetInt("hmn-bootstrap-vlan")))
	subnet.ReserveNetMgmtIPs(v.GetInt("management-net-ips"), strings.Split(v.GetString("spine-switch-xnames"), ","), strings.Split(v.GetString("leaf-switch-xnames"), ","))
	subnet.DHCPStart = ipam.Add(subnet.CIDR.IP, len(subnet.IPReservations))
	subnet.DHCPEnd = ipam.Add(ipam.Broadcast(subnet.CIDR), -1)
	// Divide the network into an appropriate number of subnets
	tempHMN.GenSubnets(cabinetDetails, net.CIDRMask(22, 32), v.GetInt("management-net-ips"), v.GetString("spine-switch-xnames"), v.GetString("leaf-switch-xnames"))

	networkMap["HMN"] = &tempHMN

	// Start the HSN with out defaults
	tempHSN := shasta.DefaultHSN
	// Update the CIDR from flags/viper
	tempHSN.CIDR = v.GetString("hsn-cidr")
	// Add a /25 for the Load Balancers
	pool, err = tempHSN.AddSubnet(net.CIDRMask(25, 32), "hsn_metallb_address_pool", tempHSN.VlanRange[0])
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	pool.AddReservation("api_gateway", "")

	// Divide the network into an appropriate number of subnets
	tempHSN.GenSubnets(cabinetDetails, net.CIDRMask(22, 32), v.GetInt("management-net-ips"), v.GetString("spine-switch-xnames"), v.GetString("leaf-switch-xnames"))

	networkMap["HSN"] = &tempHSN

	// Start the MTL with our defaults
	tempMTL := shasta.DefaultMTL
	// Update the CIDR from flags/viper
	tempMTL.CIDR = v.GetString("mtl-cidr")
	// No need to subdivide the mtl network by cabinets
	subnet, err = tempMTL.AddSubnet(net.CIDRMask(24, 32), "bootstrap_dhcp", 0)
	subnet.ReserveNetMgmtIPs(v.GetInt("management-net-ips"), strings.Split(v.GetString("spine-switch-xnames"), ","), strings.Split(v.GetString("leaf-switch-xnames"), ","))
	subnet.DHCPStart = ipam.Add(subnet.CIDR.IP, len(subnet.IPReservations))
	subnet.DHCPEnd = ipam.Add(ipam.Broadcast(subnet.CIDR), -1)
	networkMap["MTL"] = &tempMTL

	// Start the CAN with our defaults
	tempCan := shasta.DefaultCAN
	// Update the CIDR from flags/viper
	tempCan.CIDR = v.GetString("can-cidr") // This is probably a /24
	// Add a /25 for the Load Balancers on vlan0007
	_, err = tempCan.AddSubnet(net.CIDRMask(25, 32), "can_metallb_address_pool", int16(v.GetInt("can-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	// Add a /28 for the Static Pool on vlan0007
	_, err = tempCan.AddSubnet(net.CIDRMask(28, 32), "can_metallb_static_pool", int16(v.GetInt("can-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	// Add a /26 for bootstrap dhcp
	subnet, err = tempCan.AddSubnet(net.CIDRMask(26, 32), "bootstrap_dhcp", int16(v.GetInt("hmn-bootstrap-vlan")))
	subnet.ReserveNetMgmtIPs(v.GetInt("management-net-ips"), strings.Split(v.GetString("spine-switch-xnames"), ","), strings.Split(v.GetString("leaf-switch-xnames"), ","))
	subnet.AddReservation("kubeapi-vip", "k8s-virtual-ip")
	subnet.AddReservation("rgw-vip", "rgw-virtual-ip")
	subnet.DHCPStart = ipam.Add(subnet.CIDR.IP, len(subnet.IPReservations))
	subnet.DHCPEnd = ipam.Add(ipam.Broadcast(subnet.CIDR), -1)
	networkMap["CAN"] = &tempCan

	return networkMap, nil
}

// WriteNetworkFiles persistes our network configuration to disk in a directory of yaml files
func WriteNetworkFiles(basepath string, networks map[string]*shasta.IPV4Network) {
	for k, v := range networks {
		csiFiles.WriteYAMLConfig(filepath.Join(basepath, fmt.Sprintf("networks/%v.yaml", k)), v)
	}
}

func buildCabinetDetails(v *viper.Viper) []shasta.CabinetDetail {
	var cabinets []shasta.CabinetDetail
	// Add the River Cabinets
	cabinets = append(cabinets, shasta.CabinetDetail{
		Kind:            "river",
		Cabinets:        v.GetInt("river-cabinets"),
		StartingCabinet: v.GetInt("starting-river-cabinet"),
	})
	cabinets = append(cabinets, shasta.CabinetDetail{
		Kind:            "hill",
		Cabinets:        v.GetInt("hill-cabinets"),
		StartingCabinet: v.GetInt("starting-hill-cabinet"),
	})
	cabinets = append(cabinets, shasta.CabinetDetail{
		Kind:            "mountain",
		Cabinets:        v.GetInt("mountain-cabinets"),
		StartingCabinet: v.GetInt("starting-mountain-cabinet"),
	})
	return cabinets
}

// VlanConfigTemplate is the text/template to bootstrap the install cd
var VlanConfigTemplate = []byte(`
NAME='{{.FullName}}'

# Set static IP (becomes "preferred" if dhcp is enabled)
BOOTPROTO='static'
IPADDR=''    # i.e. '192.168.80.1/20'
PREFIXLEN='' # i.e. '20'

# CHANGE AT OWN RISK:
ETHERDEVICE='bond0'

# DO NOT CHANGE THESE:
VLAN_PROTOCOL='ieee802-1Q'
ONBOOT='yes'
STARTMODE='auto'
`)

// Bond0ConfigTemplate is the text/template for setting up the bond on the install NCN
var Bond0ConfigTemplate = []byte(`
NAME='Internal Interface'

# Select the NIC(s) for access.
BONDING_SLAVE0='{{.Bond0Mac}}'
BONDING_SLAVE1='{{.Bond1Mac}}'

# Set static IP (becomes "preferred" if dhcp is enabled)
BOOTPROTO='static'
IPADDR=''    # i.e. '192.168.64.1/20'
PREFIXLEN='' # i.e. '20'

# CHANGE AT OWN RISK:
BONDING_MODULE_OPTS='mode=802.3ad miimon=100 lacp_rate=fast xmit_hash_policy=layer2+3'

# DO NOT CHANGE THESE:
ONBOOT='yes'
STARTMODE='manual'
BONDING_MASTER='yes'
`)

// https://stash.us.cray.com/projects/MTL/repos/shasta-pre-install-toolkit/browse/suse/x86_64/shasta-pre-install-toolkit-sle15sp2/root

// Lan0ConfigTemplate is the text/template for handling the external site link
var Lan0ConfigTemplate = []byte(`
NAME='External Site-Link'

# Select the NIC(s) for direct, external access.
BRIDGE_PORTS=''

# Set static IP (becomes "preferred" if dhcp is enabled)
# NOTE: IPADDR's route will override DHCPs.
BOOTPROTO='dhcp'
IPADDR=''    # i.e. 10.100.10.1/24
PREFIXLEN='' # i.e. 24

# DO NOT CHANGE THESE:
ONBOOT='yes'
STARTMODE='auto'
BRIDGE='yes'
BRIDGE_STP='no'
`)

var sysconfigNetworkConfigTemplate = []byte(`
AUTO6_WAIT_AT_BOOT=""
AUTO6_UPDATE=""
LINK_REQUIRED="auto"
WICKED_DEBUG=""
WICKED_LOG_LEVEL=""
CHECK_DUPLICATE_IP="yes"
SEND_GRATUITOUS_ARP="auto"
DEBUG="no"
WAIT_FOR_INTERFACES="30"
FIREWALL="yes"
NM_ONLINE_TIMEOUT="30"
NETCONFIG_MODULES_ORDER="dns-resolver dns-bind dns-dnsmasq nis ntp-runtime"
NETCONFIG_VERBOSE="no"
NETCONFIG_FORCE_REPLACE="no"
NETCONFIG_DNS_POLICY="auto"
NETCONFIG_DNS_FORWARDER="dnsmasq"
NETCONFIG_DNS_FORWARDER_FALLBACK="yes"
NETCONFIG_DNS_STATIC_SEARCHLIST="hmn"
NETCONFIG_DNS_STATIC_SERVERS=""
NETCONFIG_DNS_RANKING="auto"
NETCONFIG_DNS_RESOLVER_OPTIONS=""
NETCONFIG_DNS_RESOLVER_SORTLIST=""
NETCONFIG_NTP_POLICY="auto"
NETCONFIG_NTP_STATIC_SERVERS=""
NETCONFIG_NIS_POLICY="auto"
NETCONFIG_NIS_SETDOMAINNAME="yes"
NETCONFIG_NIS_STATIC_DOMAIN=""
NETCONFIG_NIS_STATIC_SERVERS=""
WIRELESS_REGULATORY_DOMAIN=''
`)
