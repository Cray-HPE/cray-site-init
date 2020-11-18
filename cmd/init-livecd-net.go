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
	"text/template"

	"github.com/spf13/viper"

	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
)

// BuildLiveCDNetworks creates an array of IPv4 Networks based on the supplied system configuration
func BuildLiveCDNetworks(conf shasta.SystemConfig, v *viper.Viper) (map[string]*shasta.IPV4Network, error) {
	// our primitive ipam uses the number of cabinets to lay out a network for each one.
	// It is per-cabinet type which is pretty annoying, but here we are.

	cabinetDetails := buildCabinetDetails(v)

	var networkMap = make(map[string]*shasta.IPV4Network)

	//
	// Start the NMN with out defaults
	//
	tempNMN := shasta.DefaultNMN
	// Update the CIDR from flags/viper
	tempNMN.CIDR = v.GetString("nmn-cidr")
	// Add a /24 for Network Hardware
	hardware, err := tempNMN.AddSubnet(net.CIDRMask(24, 32), "nmn_network_hardware", int16(v.GetInt("nmn-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	hardware.FullName = "NMN Management Networking Infrastructure"
	hardware.ReserveNetMgmtIPs(strings.Split(v.GetString("spine-switch-xnames"), ","), strings.Split(v.GetString("leaf-switch-xnames"), ","), v.GetInt("management-net-ips"))
	// Add a /26 for bootstrap dhcp
	subnet, err := tempNMN.AddSubnet(net.CIDRMask(26, 32), "bootstrap_dhcp", int16(v.GetInt("nmn-bootstrap-vlan")))
	subnet.FullName = "NMN NCNs"
	subnet.AddReservation("kubeapi-vip", "k8s-virtual-ip")
	subnet.AddReservation("rgw-vip", "rgw-virtual-ip")
	// Add the macvlan network for uais
	uaisubnet, err := tempNMN.AddSubnet(net.CIDRMask(23, 32), "uai_macvlan", int16(v.GetInt("nmn-bootstrap-vlan")))
	uaisubnet.FullName = "NMN UAIs"
	uaisubnet.AddReservation("uai_macvlan_bridge", "")
	uaisubnet.AddReservation("slurmctld_service", "")
	uaisubnet.AddReservation("slurmdbd_service", "")
	uaisubnet.AddReservation("pbs_service", "")
	uaisubnet.AddReservation("pbs_comm_service", "")
	// Divide the network into an appropriate number of subnets
	tempNMN.GenSubnets(cabinetDetails, net.CIDRMask(22, 32), v.GetInt("management-net-ips"), v.GetString("spine-switch-xnames"), v.GetString("leaf-switch-xnames"))
	networkMap["NMN"] = &tempNMN

	//
	// Start the HMN with our defaults
	//
	tempHMN := shasta.DefaultHMN
	// Update the CIDR from flags/viper
	tempHMN.CIDR = v.GetString("hmn-cidr")
	// Add a /24 for Network Hardware
	hardware, err = tempHMN.AddSubnet(net.CIDRMask(24, 32), "hmn_network_hardware", int16(v.GetInt("hmn-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	hardware.FullName = "HMN Management Networking Infrastructure"
	hardware.ReserveNetMgmtIPs(strings.Split(v.GetString("spine-switch-xnames"), ","), strings.Split(v.GetString("leaf-switch-xnames"), ","), v.GetInt("management-net-ips"))
	// Add a /26 for bootstrap dhcp
	subnet, err = tempHMN.AddSubnet(net.CIDRMask(26, 32), "bootstrap_dhcp", int16(v.GetInt("hmn-bootstrap-vlan")))
	subnet.FullName = "HMN NCNs"

	// TODO - removing this causes a "Couldn't find switch port for NCN error", but I don't want this
	subnet.ReserveNetMgmtIPs(strings.Split(v.GetString("spine-switch-xnames"), ","), strings.Split(v.GetString("leaf-switch-xnames"), ","), v.GetInt("management-net-ips"))
	// Divide the network into an appropriate number of subnets
	tempHMN.GenSubnets(cabinetDetails, net.CIDRMask(22, 32), v.GetInt("management-net-ips"), v.GetString("spine-switch-xnames"), v.GetString("leaf-switch-xnames"))
	networkMap["HMN"] = &tempHMN

	//
	// Start the HSN with our defaults
	//

	tempHSN := shasta.DefaultHSN
	// Update the CIDR from flags/viper
	tempHSN.CIDR = v.GetString("hsn-cidr")
	// Divide the network into an appropriate number of subnets
	tempHSN.GenSubnets(cabinetDetails, net.CIDRMask(22, 32), v.GetInt("management-net-ips"), v.GetString("spine-switch-xnames"), v.GetString("leaf-switch-xnames"))
	networkMap["HSN"] = &tempHSN

	//
	// Start the MTL with our defaults
	//
	tempMTL := shasta.DefaultMTL
	// Update the CIDR from flags/viper
	tempMTL.CIDR = v.GetString("mtl-cidr")
	// Add a /24 for Network Hardware
	hardware, err = tempMTL.AddSubnet(net.CIDRMask(24, 32), "mtl_network_hardware", int16(v.GetInt("mtl-bootstrap-vlan")))
	if err != nil {
		log.Printf("Couldn't add subnet: %v", err)
	}
	hardware.FullName = "MTL Management Networking Infrastructure"
	hardware.ReserveNetMgmtIPs(strings.Split(v.GetString("spine-switch-xnames"), ","), strings.Split(v.GetString("leaf-switch-xnames"), ","), v.GetInt("management-net-ips"))
	// No need to subdivide the mtl network by cabinets
	subnet, err = tempMTL.AddSubnet(net.CIDRMask(24, 32), "bootstrap_dhcp", 0)
	// TODO: Probably only really required for masters.
	subnet.FullName = "MTL NCNs"
	networkMap["MTL"] = &tempMTL

	//
	// Start the NMN Load Balancer with our Defaults
	//
	tempNMNLoadBalancer := shasta.DefaultLoadBalancerNMN
	// Add a /25 for the Load Balancers
	pool, err := tempNMNLoadBalancer.AddSubnet(net.CIDRMask(25, 32), "nmn_metallb_address_pool", int16(v.GetInt("nmn-bootstrap-vlan")))
	pool.FullName = "NMN MetalLB"
	pool.AddReservation("api_gateway", "")
	networkMap["NMNLB"] = &tempNMNLoadBalancer

	//
	// Start the HMN Load Balancer with our Defaults
	//
	tempHMNLoadBalancer := shasta.DefaultLoadBalancerHMN
	pool, err = tempHMNLoadBalancer.AddSubnet(net.CIDRMask(25, 32), "hmn_metallb_address_pool", int16(v.GetInt("hmn-bootstrap-vlan")))
	pool.FullName = "HMN MetalLB"
	pool.AddReservation("api_gateway", "")
	networkMap["HMNLB"] = &tempHMNLoadBalancer

	//
	// Start the CAN with our defaults
	//
	tempCAN := shasta.DefaultCAN
	// Update the CIDR from flags/viper
	tempCAN.CIDR = v.GetString("can-cidr") // This is probably a /24
	// Add a /28 for the Static Pool on vlan0007
	_, canStaticPool, err := net.ParseCIDR(v.GetString("can-static-pool"))
	if err != nil {
		log.Printf("Invalid can-static-pool.  Cowardly refusing to create it.")
	} else {
		static, err := tempCAN.AddSubnetbyCIDR(*canStaticPool, "can_metallb_static_pool", int16(v.GetInt("can-bootstrap-vlan")))
		if err != nil {
			log.Printf("Couldn't add subnet: %v", err)
		}
		static.FullName = "CAN Static Pool MetalLB"
	}
	_, canDynamicPool, err := net.ParseCIDR(v.GetString("can-dynamic-pool"))
	if err != nil {
		log.Printf("Invalid can-dynamic-pool.  Cowardly refusing to create it.")
	} else {
		pool, err = tempCAN.AddSubnetbyCIDR(*canDynamicPool, "can_metallb_address_pool", int16(v.GetInt("can-bootstrap-vlan")))
		if err != nil {
			log.Printf("Couldn't add subnet: %v", err)
		}
		pool.FullName = "CAN Dynamic MetalLB"
	}
	// Add a /26 for bootstrap dhcp
	subnet, err = tempCAN.AddSubnet(net.CIDRMask(26, 32), "bootstrap_dhcp", int16(v.GetInt("hmn-bootstrap-vlan")))
	subnet.FullName = "CAN NCNs"
	// TODO: Is this really necessary?  At best this is far too many.
	subnet.ReserveNetMgmtIPs(strings.Split(v.GetString("spine-switch-xnames"), ","), strings.Split(v.GetString("leaf-switch-xnames"), ","), v.GetInt("management-net-ips"))
	subnet.AddReservation("kubeapi-vip", "k8s-virtual-ip")
	subnet.AddReservation("rgw-vip", "rgw-virtual-ip")
	networkMap["CAN"] = &tempCAN

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

// WriteCPTNetworkConfig writes the Network Configuration details for the installation node  (CPT)
func WriteCPTNetworkConfig(path string, ncn shasta.LogicalNCN, shastaNetworks map[string]*shasta.IPV4Network) error {
	// log.Println("Interface Networks:", ncn.Networks)
	// log.Println("Networks are:", shastaNetworks)
	csiFiles.WriteTemplate(filepath.Join(path, "ifcfg-bond0"), template.Must(template.New("bond0").Parse(string(Bond0ConfigTemplate))), ncn)
	csiFiles.WriteTemplate(filepath.Join(path, "ifcfg-lan0"), template.Must(template.New("lan0").Parse(string(Lan0ConfigTemplate))), ncn)
	for _, network := range ncn.Networks {
		csiFiles.WriteTemplate(filepath.Join(path, fmt.Sprintf("ifcfg-vlan%03d", network.Vlan)), template.Must(template.New("vlan").Parse(string(VlanConfigTemplate))), network)
	}
	return nil
}

// VlanConfigTemplate is the text/template to bootstrap the install cd
var VlanConfigTemplate = []byte(`
NAME='{{.FullName}}'

# Set static IP (becomes "preferred" if dhcp is enabled)
BOOTPROTO='static'
IPADDR='{{.CIDR}}'    # i.e. '192.168.80.1/20'
PREFIXLEN='{{.Mask}}' # i.e. '20'

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
BONDING_SLAVE0='{{.Bond0Mac0}}'
BONDING_SLAVE1='{{.Bond0Mac1}}'

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
