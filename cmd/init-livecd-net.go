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
func BuildLiveCDNetworks(v *viper.Viper, switches []*shasta.ManagementSwitch) (map[string]*shasta.IPV4Network, error) {
	var networkMap = make(map[string]*shasta.IPV4Network)

	var internalNetConfigs = make(map[string]shasta.NetworkLayoutConfiguration)
	internalNetConfigs["HMN"] = shasta.DefaultHMNConfig
	internalNetConfigs["CAN"] = shasta.DefaultCANConfig
	internalNetConfigs["NMN"] = shasta.DefaultNMNConfig
	internalNetConfigs["HSN"] = shasta.DefaultHSNConfig
	internalNetConfigs["MTL"] = shasta.DefaultMTLConfig

	internalCabinetDetails := buildCabinetDetails(v)

	for name, layout := range internalNetConfigs {
		myLayout := layout

		// Update with computed fields
		myLayout.CabinetDetails = internalCabinetDetails
		myLayout.ManagementSwitches = switches

		// Update with flags
		myLayout.BaseVlan = int16(v.GetInt(fmt.Sprintf("%v-bootstrap-vlan", strings.ToLower(name))))
		myLayout.Template.CIDR = v.GetString(fmt.Sprintf("%v-cidr", strings.ToLower(name)))
		myLayout.AdditionalNetworkingSpace = v.GetInt("mangement-net-ips")

		netPtr, err := createNetFromLayoutConfig(myLayout, v)
		if err != nil {
			log.Fatalf("Couldn't add %v Network because %v", name, err)
		}
		networkMap[name] = netPtr
	}

	//
	// Start the NMN Load Balancer with our Defaults
	//
	tempNMNLoadBalancer := shasta.DefaultLoadBalancerNMN
	// Add a /25 for the Load Balancers
	pool, _ := tempNMNLoadBalancer.AddSubnet(net.CIDRMask(24, 32), "nmn_metallb_address_pool", int16(v.GetInt("nmn-bootstrap-vlan")))
	pool.FullName = "NMN MetalLB"
	pool.AddReservation("istio-ingressgateway", "api-gw-service packages registry")
	pool.AddReservation("rsyslog-aggregator", "rsyslog-agg-service")
	pool.AddReservation("rsyslog-aggregator-udp", "rsyslog-agg-service-udp")
	pool.AddReservation("cray-tftp", "tftp-service")
	networkMap["NMNLB"] = &tempNMNLoadBalancer

	//
	// Start the HMN Load Balancer with our Defaults
	//
	tempHMNLoadBalancer := shasta.DefaultLoadBalancerHMN
	pool, _ = tempHMNLoadBalancer.AddSubnet(net.CIDRMask(24, 32), "hmn_metallb_address_pool", int16(v.GetInt("hmn-bootstrap-vlan")))
	pool.FullName = "HMN MetalLB"
	pool.AddReservation("istio-ingressgateway-hmn", "api-gw-service packages registry")
	pool.AddReservation("rsyslog-aggregator-hmn", "rsyslog-agg-service")
	pool.AddReservation("rsyslog-aggregator-hmn-udp", "rsyslog-agg-service-udp")
	pool.AddReservation("cray-tftp-hmn", "tftp-service")
	networkMap["HMNLB"] = &tempHMNLoadBalancer

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
	log.Println("Interface Networks:", ncn.Networks)
	// log.Println("Networks are:", shastaNetworks)
	var bond0Net shasta.NCNNetwork
	for _, network := range ncn.Networks {
		if network.NetworkName == "MTL" {
			bond0Net = network
		}
	}
	csiFiles.WriteTemplate(filepath.Join(path, "ifcfg-bond0"), template.Must(template.New("bond0").Parse(string(Bond0ConfigTemplate))), bond0Net)
	csiFiles.WriteTemplate(filepath.Join(path, "ifcfg-lan0"), template.Must(template.New("lan0").Parse(string(Lan0ConfigTemplate))), ncn)
	for _, network := range ncn.Networks {
		if network.Vlan != 0 {
			csiFiles.WriteTemplate(filepath.Join(path, fmt.Sprintf("ifcfg-vlan%03d", network.Vlan)), template.Must(template.New("vlan").Parse(string(VlanConfigTemplate))), network)
		}
	}
	return nil
}

func switchXnamesByType(switches []*shasta.ManagementSwitch, switchType string) []string {
	var xnames []string
	for _, mswitch := range switches {
		if mswitch.SwitchType == switchType {
			xnames = append(xnames, mswitch.Xname)
		}
	}
	return xnames
}

func createNetFromLayoutConfig(conf shasta.NetworkLayoutConfiguration, v *viper.Viper) (*shasta.IPV4Network, error) {
	// start with the defaults
	tempNet := conf.Template
	// figure out what switches we have
	leafSwitches := switchXnamesByType(conf.ManagementSwitches, "Leaf")
	spineSwitches := switchXnamesByType(conf.ManagementSwitches, "Spine")
	aggSwitches := switchXnamesByType(conf.ManagementSwitches, "Aggregation")
	cduSwitches := switchXnamesByType(conf.ManagementSwitches, "CDU")

	// Do all the special stuff for the CAN
	if tempNet.Name == "CAN" {
		_, canStaticPool, err := net.ParseCIDR(v.GetString("can-static-pool"))
		if err != nil {
			log.Printf("IP Addressing Failure\nInvalid can-static-pool.  Cowardly refusing to create it.")
		} else {
			static, err := tempNet.AddSubnetbyCIDR(*canStaticPool, "can_metallb_static_pool", int16(v.GetInt("can-bootstrap-vlan")))
			if err != nil {
				log.Fatalf("IP Addressing Failure\nCouldn't add MetalLB Static pool of %v to net %v: %v", v.GetString("can-static-pool"), tempNet.CIDR, err)
			}
			static.FullName = "CAN Static Pool MetalLB"
		}
		_, canDynamicPool, err := net.ParseCIDR(v.GetString("can-dynamic-pool"))
		if err != nil {
			log.Printf("IP Addressing Failure\nInvalid can-dynamic-pool.  Cowardly refusing to create it.")
		} else {
			pool, err := tempNet.AddSubnetbyCIDR(*canDynamicPool, "can_metallb_address_pool", int16(v.GetInt("can-bootstrap-vlan")))
			if err != nil {
				log.Fatalf("IP Addressing Failure\nCouldn't add MetalLB Dynamic pool of %v to net %v: %v", v.GetString("can-dynamic-pool"), tempNet.CIDR, err)
			}
			pool.FullName = "CAN Dynamic MetalLB"
		}
	}

	// Process the dedicated Networking Hardware Subnet
	if conf.IncludeNetworkingHardwareSubnet {
		// create the subnet
		hardwareSubnet, err := tempNet.AddSubnet(conf.NetworkingHardwareNetmask, "network_hardware", conf.BaseVlan)
		if err != nil {
			return &tempNet, fmt.Errorf("unable to add network hardware subnet to %v because %v", conf.Template.Name, err)
		}
		// populate it with base information
		hardwareSubnet.FullName = fmt.Sprintf("%v Management Network Infrastructure", tempNet.Name)
		hardwareSubnet.ReserveNetMgmtIPs(spineSwitches, leafSwitches, aggSwitches, cduSwitches, conf.AdditionalNetworkingSpace)
	}

	// Set up the Boostrap DHCP subnet(s)
	if conf.IncludeBootstrapDHCP {
		var subnet *shasta.IPV4Subnet
		subnet, err := tempNet.AddBiggestSubnet(conf.DesiredBootstrapDHCPMask, "bootstrap_dhcp", conf.BaseVlan)
		if err != nil {
			return &tempNet, fmt.Errorf("unable to add bootstrap_dhcp subnet to %v because %v", conf.Template.Name, err)
		}
		subnet.FullName = fmt.Sprintf("%v Bootstrap DHCP Subnet", tempNet.Name)
		if tempNet.Name == "NMN" || tempNet.Name == "CAN" || tempNet.Name == "HMN" {
			subnet.ReserveNetMgmtIPs(spineSwitches, leafSwitches, []string{}, []string{}, 0)
			subnet.AddReservation("kubeapi-vip", "k8s-virtual-ip")
			subnet.AddReservation("rgw-vip", "rgw-virtual-ip")
			if tempNet.Name == "CAN" {
				subnet.Gateway = net.ParseIP(v.GetString("can-gateway"))
			}
		}
	}

	// Add the macvlan/uai subnet(s)
	if conf.IncludeUAISubnet {
		uaisubnet, err := tempNet.AddSubnet(net.CIDRMask(23, 32), "uai_macvlan", conf.BaseVlan)
		if err != nil {
			log.Fatalf("Couln't add the uai subnet to the %v Network: %v", tempNet.Name, err)
		}
		uaisubnet.FullName = "NMN UAIs"
		for reservationName, reservationAlias := range shasta.DefaultUAISubnetReservations {
			uaisubnet.AddReservation(reservationName, reservationAlias)
		}
	}

	// Build out the per-cabinet subnets
	if conf.SubdivideByCabinet {
		tempNet.GenSubnets(conf.CabinetDetails, conf.CabinetCIDR)
	}
	return &tempNet, nil
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
NAME='Internal Interface'# Select the NIC(s) for access to the CRAY.

# Select the NIC(s) for access.
BONDING_SLAVE0='p1p1'
BONDING_SLAVE1='p10p1' # Set static IP (becomes "preferred" if dhcp is enabled)

# Set static IP (becomes "preferred" if dhcp is enabled)
BOOTPROTO='static'
IPADDR='{{.CIDR}}'    # i.e. '192.168.64.1/20'
PREFIXLEN='{{.Mask}}' # i.e. '20'

# CHANGE AT OWN RISK:
BONDING_MODULE_OPTS='mode=802.3ad miimon=100 lacp_rate=fast xmit_hash_policy=layer2+3'# DO NOT CHANGE THESE:

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
