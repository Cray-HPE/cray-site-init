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

	"stash.us.cray.com/MTL/csi/internal/files"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/ipam"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
)

// BuildLiveCDNetworks creates an array of IPv4 Networks based on the supplied system configuration
func BuildLiveCDNetworks(v *viper.Viper, internalCabinetDetails []shasta.CabinetDetail, switches []*shasta.ManagementSwitch) (map[string]*shasta.IPV4Network, error) {
	var networkMap = make(map[string]*shasta.IPV4Network)

	var internalNetConfigs = make(map[string]shasta.NetworkLayoutConfiguration)
	internalNetConfigs["HMN"] = shasta.DefaultHMNConfig
	internalNetConfigs["CAN"] = shasta.DefaultCANConfig
	internalNetConfigs["NMN"] = shasta.DefaultNMNConfig
	internalNetConfigs["HSN"] = shasta.DefaultHSNConfig
	internalNetConfigs["MTL"] = shasta.DefaultMTLConfig

	for name, layout := range internalNetConfigs {
		myLayout := layout

		// Update with computed fields
		myLayout.CabinetDetails = internalCabinetDetails
		myLayout.ManagementSwitches = switches

		// Update with flags
		myLayout.BaseVlan = int16(v.GetInt(fmt.Sprintf("%v-bootstrap-vlan", strings.ToLower(name))))
		myLayout.Template.CIDR = v.GetString(fmt.Sprintf("%v-cidr", strings.ToLower(name)))
		myLayout.AdditionalNetworkingSpace = v.GetInt("management-net-ips")

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
	// Add a /24 for the Load Balancers
	pool, _ := tempNMNLoadBalancer.AddSubnet(net.CIDRMask(24, 32), "nmn_metallb_address_pool", int16(v.GetInt("nmn-bootstrap-vlan")))
	pool.FullName = "NMN MetalLB"
	for nme, rsrv := range shasta.PinnedMetalLBReservations {
		pool.AddReservationWithPin(nme, strings.Join(rsrv.Aliases, ","), rsrv.IPByte)
	}
	networkMap["NMNLB"] = &tempNMNLoadBalancer

	//
	// Start the HMN Load Balancer with our Defaults
	//
	tempHMNLoadBalancer := shasta.DefaultLoadBalancerHMN
	pool, _ = tempHMNLoadBalancer.AddSubnet(net.CIDRMask(24, 32), "hmn_metallb_address_pool", int16(v.GetInt("hmn-bootstrap-vlan")))
	pool.FullName = "HMN MetalLB"
	for nme, rsrv := range shasta.PinnedMetalLBReservations {
		pool.AddReservationWithPin(nme, strings.Join(rsrv.Aliases, ","), rsrv.IPByte)
	}
	networkMap["HMNLB"] = &tempHMNLoadBalancer

	return networkMap, nil
}

// WriteNetworkFiles persistes our network configuration to disk in a directory of yaml files
func WriteNetworkFiles(basepath string, networks map[string]*shasta.IPV4Network) {
	for k, v := range networks {
		csiFiles.WriteYAMLConfig(filepath.Join(basepath, fmt.Sprintf("networks/%v.yaml", k)), v)
	}
}

func positionInCabinetList(kind string, cabs []shasta.CabinetDetail) (int, error) {
	for i, cab := range cabs {
		if cab.Kind == kind {
			return i, nil
		}
	}
	return 0, fmt.Errorf("%s not found", kind)
}

func buildCabinetDetails(v *viper.Viper) []shasta.CabinetDetail {
	var cabDetailFile shasta.CabinetDetailFile
	if v.IsSet("cabinets-yaml") {
		err := files.ReadYAMLConfig(v.GetString("cabinets-yaml"), &cabDetailFile)
		if err != nil {
			log.Fatalf("Unable to parse cabinets-yaml file: %s\nError: %v", v.GetString("cabinets-yaml"), err)
		}
	}
	for _, cabType := range []string{"river", "mountain", "hill"} {
		pos, err := positionInCabinetList(cabType, cabDetailFile.Cabinets)
		if err != nil {
			var tmpCabinet shasta.CabinetDetail
			tmpCabinet.Kind = cabType
			tmpCabinet.Cabinets = v.GetInt(fmt.Sprintf("%s-cabinets", cabType))
			tmpCabinet.StartingCabinet = v.GetInt(fmt.Sprintf("starting-%s-cabinet", cabType))
			tmpCabinet.PopulateIds()
			cabDetailFile.Cabinets = append(cabDetailFile.Cabinets, tmpCabinet)
		} else {
			cabDetailFile.Cabinets[pos].Cabinets = cabDetailFile.Cabinets[pos].Length()
			cabDetailFile.Cabinets[pos].PopulateIds()
		}
	}
	return cabDetailFile.Cabinets
}

// WriteCPTNetworkConfig writes the Network Configuration details for the installation node  (CPT)
func WriteCPTNetworkConfig(path string, v *viper.Viper, ncn shasta.LogicalNCN, shastaNetworks map[string]*shasta.IPV4Network) error {
	type Route struct {
		CIDR    net.IP
		Mask    net.IP
		Gateway net.IP
	}
	var bond0Net shasta.NCNNetwork
	for _, network := range ncn.Networks {
		if network.NetworkName == "MTL" {
			bond0Net = network
		}
	}
	_, metalNet, _ := net.ParseCIDR(shastaNetworks["NMNLB"].CIDR)
	nmnNetNet, _ := shastaNetworks["NMN"].LookUpSubnet("network_hardware")

	metalLBRoute := Route{
		CIDR:    metalNet.IP,
		Mask:    net.IP(metalNet.Mask),
		Gateway: nmnNetNet.Gateway,
	}
	bond0Struct := struct {
		Bond0 string
		Bond1 string
		Mask  string
		CIDR  string
	}{
		Bond0: strings.Split(v.GetString("install-ncn-bond-members"), ",")[0],
		Bond1: strings.Split(v.GetString("install-ncn-bond-members"), ",")[1],
		Mask:  bond0Net.Mask,
		CIDR:  bond0Net.CIDR,
	}
	csiFiles.WriteTemplate(filepath.Join(path, "ifcfg-bond0"), template.Must(template.New("bond0").Parse(string(Bond0ConfigTemplate))), bond0Struct)
	siteNetDef := strings.Split(v.GetString("site-ip"), "/")
	lan0struct := struct {
		Nic, IP, IPPrefix string
	}{
		v.GetString("site-nic"),
		v.GetString("site-ip"),
		siteNetDef[1],
	}
	lan0RouteStruct := struct {
		CIDR    string
		Mask    string
		Gateway string
	}{"default", "-", v.GetString("site-gw")}

	csiFiles.WriteTemplate(filepath.Join(path, "ifcfg-lan0"), template.Must(template.New("lan0").Parse(string(Lan0ConfigTemplate))), lan0struct)
	lan0sysconfig := struct {
		SiteDNS string
	}{
		v.GetString("site-dns"),
	}
	csiFiles.WriteTemplate(filepath.Join(path, "config"), template.Must(template.New("netcofig").Parse(string(sysconfigNetworkConfigTemplate))), lan0sysconfig)
	csiFiles.WriteTemplate(filepath.Join(path, "ifroute-lan0"), template.Must(template.New("vlan").Parse(string(VlanRouteTemplate))), []interface{}{lan0RouteStruct})
	for _, network := range ncn.Networks {
		if stringInSlice(network.NetworkName, []string{"HMN", "NMN", "MTL", "CAN"}) {
			if network.Vlan != 0 {
				csiFiles.WriteTemplate(filepath.Join(path, fmt.Sprintf("ifcfg-vlan%03d", network.Vlan)), template.Must(template.New("vlan").Parse(string(VlanConfigTemplate))), network)
			}
			if network.NetworkName == "NMN" {
				csiFiles.WriteTemplate(filepath.Join(path, fmt.Sprintf("ifroute-vlan%03d", network.Vlan)), template.Must(template.New("vlan").Parse(string(VlanRouteTemplate))), []Route{metalLBRoute})
			}
		}
	}
	return nil
}

func switchXnamesByType(switches []*shasta.ManagementSwitch, switchType shasta.ManagementSwitchType) []string {
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
			if tempNet.Name == "CAN" {
				subnet.Gateway = net.ParseIP(v.GetString("can-gateway"))
				subnet.AddReservation("can-switch-1", "")
				subnet.AddReservation("can-switch-2", "")

			} else {
				subnet.ReserveNetMgmtIPs(spineSwitches, leafSwitches, aggSwitches, cduSwitches, conf.AdditionalNetworkingSpace)
			}
			subnet.AddReservation("kubeapi-vip", "k8s-virtual-ip")
			subnet.AddReservation("rgw-vip", "rgw-virtual-ip")
		}
	}

	// Add the macvlan/uai subnet(s)
	if conf.IncludeUAISubnet {
		uaisubnet, err := tempNet.AddSubnet(net.CIDRMask(23, 32), "uai_macvlan", conf.BaseVlan)
		supernetIP, _, _ := net.ParseCIDR(tempNet.CIDR)
		uaisubnet.Gateway = ipam.Add(supernetIP, 1)
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

// ApplySupernetHack applys a dirty hack.
func ApplySupernetHack(tempNetPtr *shasta.IPV4Network) {
	// Replace the gateway and netmask on the to better support the 1.3 network switch configuration
	// *** This is a HACK ***
	supernetIP, superNet, err := net.ParseCIDR(tempNetPtr.CIDR)
	if err != nil {
		log.Fatal("Couldn't parse the CIDR for ", tempNetPtr.Name)
	}
	for _, subnetName := range []string{"bootstrap_dhcp", "uai_macvlan", "network_hardware",
		"can_metallb_static_pool", "can_metallb_address_pool"} {
		tempSubnet, err := tempNetPtr.LookUpSubnet(subnetName)
		if err == nil {
			// Replace the standard netmask with the supernet netmask
			// Replace the standard gateway with the supernet gateway
			// ** HACK ** We're doing this here to bypass all sanity checks
			// This **WILL** cause an overlap of broadcast domains, but is required
			// for reducing switch configuration changes from 1.3 to 1.4
			tempSubnet.Gateway = ipam.Add(supernetIP, 1)
			tempSubnet.CIDR.Mask = superNet.Mask
		}
	}
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

// VlanRouteTemplate allows us to add static routes to the vlan(s) on the CPT node
var VlanRouteTemplate = []byte(`
{{- range . -}}
{{.CIDR}} {{.Gateway}} {{.Mask}} -
{{ end -}}
`)

// Bond0ConfigTemplate is the text/template for setting up the bond on the install NCN
var Bond0ConfigTemplate = []byte(`
NAME='Internal Interface'# Select the NIC(s) for access to the CRAY.

# Select the NIC(s) for access.
BONDING_SLAVE0='{{.Bond0}}'
BONDING_SLAVE1='{{.Bond1}}'

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
BRIDGE_PORTS='{{.Nic}}'

# Set static IP (becomes "preferred" if dhcp is enabled)
# NOTE: IPADDR's route will override DHCPs.
BOOTPROTO='static'
IPADDR='{{.IP}}'    # i.e. 10.100.10.1/24
PREFIXLEN='{{.IPPrefix}}' # i.e. 24

# DO NOT CHANGE THESE:
ONBOOT='yes'
STARTMODE='auto'
BRIDGE='yes'
BRIDGE_STP='no'
`)

var sysconfigNetworkConfigTemplate = []byte(`
# Generated by CSI as part of the installation planning
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
NETCONFIG_DNS_STATIC_SEARCHLIST="nmn hmn"
NETCONFIG_DNS_STATIC_SERVERS="{{.SiteDNS}}"
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
