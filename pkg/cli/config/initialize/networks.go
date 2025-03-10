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

package initialize

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/viper"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/cli"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

// WriteCPTNetworkConfig writes the Network Configuration details for the installation node (PIT)
func WriteCPTNetworkConfig(
	path string, v *viper.Viper, ncn LogicalNCN, shastaNetworks map[string]*networking.IPV4Network,
) error {
	type Route struct {
		CIDR    net.IP
		Mask    net.IP
		Gateway net.IP
	}
	var bond0Net NCNNetwork
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
		Members []string
		Mask    string
		CIDR    string
	}{
		Members: strings.Split(
			v.GetString("install-ncn-bond-members"),
			",",
		),
		Mask: bond0Net.Mask,
		CIDR: bond0Net.CIDR,
	}
	csiFiles.WriteTemplate(
		filepath.Join(
			path,
			"ifcfg-bond0",
		),
		template.Must(template.New("bond0").Parse(string(BondConfigTemplate))),
		bond0Struct,
	)
	siteNetDef := strings.Split(
		v.GetString("site-ip"),
		"/",
	)
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
	}{
		"default",
		"-",
		v.GetString("site-gw"),
	}

	csiFiles.WriteTemplate(
		filepath.Join(
			path,
			"ifcfg-lan0",
		),
		template.Must(template.New("lan0").Parse(string(Lan0ConfigTemplate))),
		lan0struct,
	)
	lan0sysconfig := struct {
		SiteDNS string
	}{
		v.GetString("site-dns"),
	}
	csiFiles.WriteTemplate(
		filepath.Join(
			path,
			"config",
		),
		template.Must(template.New("netcofig").Parse(string(sysconfigNetworkConfigTemplate))),
		lan0sysconfig,
	)
	csiFiles.WriteTemplate(
		filepath.Join(
			path,
			"ifroute-lan0",
		),
		template.Must(template.New("vlan").Parse(string(VlanRouteTemplate))),
		[]interface{}{lan0RouteStruct},
	)
	for _, network := range ncn.Networks {
		if cli.StringInSlice(
			network.NetworkName,
			networking.ValidNetNames,
		) {
			if network.Vlan != networking.DefaultMTLVlan && network.NetworkName != "CHN" {
				csiFiles.WriteTemplate(
					filepath.Join(
						path,
						fmt.Sprintf(
							"ifcfg-%s",
							strings.ToLower(network.InterfaceName),
						),
					),
					template.Must(template.New("vlan").Parse(string(VlanConfigTemplate))),
					network,
				)
			}
			if network.NetworkName == "NMN" {
				csiFiles.WriteTemplate(
					filepath.Join(
						path,
						fmt.Sprintf(
							"ifroute-%s",
							strings.ToLower(network.InterfaceName),
						),
					),
					template.Must(template.New("vlan").Parse(string(VlanRouteTemplate))),
					[]Route{metalLBRoute},
				)
			}
		}
	}
	return nil
}

// VlanConfigTemplate is the text/template to bootstrap the install cd
var VlanConfigTemplate = []byte(`
{{- /* remove leading whitespace */ -}}
#
## This file was generated by cray-site-init.
#
NAME='{{.FullName}}'

# Set static IP (becomes "preferred" if dhcp is enabled)
BOOTPROTO='static'
IPADDR='{{.CIDR}}'
PREFIXLEN='{{.Mask}}'

# CHANGE AT OWN RISK:
ETHERDEVICE='{{.ParentInterfaceName}}'

# DO NOT CHANGE THESE:
VLAN_PROTOCOL='ieee802-1Q'
VLAN='yes'
VLAN_ID={{.Vlan}}
ONBOOT='yes'
STARTMODE='auto'
`)

// VlanRouteTemplate allows us to add static routes to the vlan(s) on the PIT node
var VlanRouteTemplate = []byte(`
{{- /* remove leading whitespace */ -}}
#
## This file was generated by cray-site-init.
#
{{ range . -}}
{{.CIDR}} {{.Gateway}} {{.Mask}} -
{{ end -}}
`)

// BondConfigTemplate is the text/template for setting up the bond on the install NCN
var BondConfigTemplate = []byte(`
{{- /* remove leading whitespace */ -}}
#
## This file was generated by cray-site-init.
#
NAME='Internal Interface'

# Set static IP (becomes "preferred" if dhcp is enabled)
BOOTPROTO='static'
IPADDR='{{.CIDR}}'
PREFIXLEN='{{.Mask}}'

# CHANGE AT OWN RISK:
BONDING_MODULE_OPTS='mode=802.3ad miimon=100 lacp_rate=fast xmit_hash_policy=layer2+3'

# DO NOT CHANGE THESE:
ONBOOT='yes'
STARTMODE='auto'
BONDING_MASTER='yes'

# BOND MEMBERS:
{{ range $i, $v := .Members -}}
BONDING_SLAVE{{$i}}='{{$v}}'
{{ end -}}
`)

// https://stash.us.cray.com/projects/MTL/repos/shasta-pre-install-toolkit/browse/suse/x86_64/shasta-pre-install-toolkit-sle15sp2/root

// Lan0ConfigTemplate is the text/template for handling the external site link
var Lan0ConfigTemplate = []byte(`
{{- /* remove leading whitespace */ -}}
#
## This file was generated by cray-site-init.
#
NAME='External Site-Link'

# Select the NIC(s) for direct, external access.
BRIDGE_PORTS='{{.Nic}}'

# Set static IP (becomes "preferred" if dhcp is enabled)
# NOTE: IPADDR's route will override DHCPs.
BOOTPROTO='static'
IPADDR='{{.IP}}'
PREFIXLEN='{{.IPPrefix}}'

# DO NOT CHANGE THESE:
ONBOOT='yes'
STARTMODE='auto'
BRIDGE='yes'
BRIDGE_STP='no'
`)

var sysconfigNetworkConfigTemplate = []byte(`
{{- /* remove leading whitespace */ -}}
#
## This file was generated by cray-site-init.
#
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
NETCONFIG_VERBOSE="no"
NETCONFIG_DNS_STATIC_SEARCHLIST="nmn mtl hmn"
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
# NETCONFIG_DNS_FORWARDER="dnsmasq" this will automatically add 
# 127.0.0.1 to the local /etc/resolv.conf when "netconfig update -f" is invoked
NETCONFIG_FORCE_REPLACE="no"
NETCONFIG_MODULES_ORDER="dns-resolver dns-bind dns-dnsmasq nis ntp-runtime"
NETCONFIG_DNS_POLICY="auto"
NETCONFIG_DNS_FORWARDER="dnsmasq"
NETCONFIG_DNS_FORWARDER_FALLBACK="yes"
`)
