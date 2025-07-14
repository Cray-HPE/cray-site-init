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

package initialize

import (
	"fmt"
	"net/netip"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/cli"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

type Route struct {
	Destination string
	Mask        string
	NextHop     string
}

// Routes is a list of Route structs.
type Routes []Route

type LACPBond struct {
	Members   []string
	PrefixLen int
	CIDR      string
}

type NetworkInterface struct {
	NIC       string
	IP        string
	PrefixLen int
	VLAN      int
}

type SysConfig struct {
	SiteDNS string
}

// WriteCPTNetworkConfig writes the Network Configuration details for the installation node (PIT)
func WriteCPTNetworkConfig(
	path string, v *viper.Viper, ncn LogicalNCN, shastaNetworks map[string]*networking.IPNetwork,
) (err error) {

	for _, network := range ncn.Networks {
		if network.NetworkName != "MTL" {
			continue
		}
		data := MakeTemplateData(
			LACPBond{
				Members: strings.Split(
					v.GetString("install-ncn-bond-members"),
					",",
				),
				CIDR:      network.CIDR4.String(),
				PrefixLen: network.CIDR4.Bits(),
			},
		)
		err = files.WriteTemplate(
			filepath.Join(
				path,
				"ifcfg-bond0",
			),
			template.Must(template.New("bond0").Parse(string(BondConfigTemplate))),
			data,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to write ifcfg-bond0 because %v",
				err,
			)
		}
		break
	}

	sitePrefix, err := netip.ParsePrefix(v.GetString("site-ip"))
	if err != nil {
		return fmt.Errorf(
			"invalid site-ip, cannot continue because %v",
			err,
		)
	}
	lan0ConfigData := MakeTemplateData(
		NetworkInterface{
			NIC:       v.GetString("site-nic"),
			IP:        sitePrefix.Addr().String(),
			PrefixLen: sitePrefix.Bits(),
		},
	)
	err = files.WriteTemplate(
		filepath.Join(
			path,
			"ifcfg-lan0",
		),
		template.Must(template.New("lan0").Parse(string(Lan0ConfigTemplate))),
		lan0ConfigData,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to write ifcfg-lan0 because %v",
			err,
		)
	}

	siteGw, err := netip.ParseAddr(v.GetString("site-gw"))
	if err != nil {
		return fmt.Errorf(
			"invalide site-gw, cannot continue because %v",
			err,
		)
	}
	lan0Route := Route{
		Destination: "default",
		Mask:        "-",
		NextHop:     siteGw.String(),
	}
	lan0RouteData := MakeTemplateData(Routes{lan0Route})
	err = files.WriteTemplate(
		filepath.Join(
			path,
			"ifroute-lan0",
		),
		template.Must(template.New("vlan").Parse(string(VlanRouteTemplate))),
		lan0RouteData,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to write ifroute-lan0 because %v",
			err,
		)
	}

	sysconfigData := MakeTemplateData(SysConfig{SiteDNS: v.GetString("site-dns")})
	err = files.WriteTemplate(
		filepath.Join(
			path,
			"config",
		),
		template.Must(template.New("netcofig").Parse(string(sysconfigNetworkConfigTemplate))),
		sysconfigData,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to write netcofig because %v",
			err,
		)
	}

	for _, network := range ncn.Networks {
		if cli.StringInSlice(
			network.NetworkName,
			networking.ValidNetNames,
		) {
			if network.ParentInterfaceName == "" {
				continue
			}
			ifcfgFilename := fmt.Sprintf(
				"ifcfg-%s",
				strings.ToLower(network.InterfaceName),
			)
			ifrouteFilename := fmt.Sprintf(
				"ifroute-%s",
				strings.ToLower(network.InterfaceName),
			)
			data := MakeTemplateData(network)
			if network.Vlan != networking.DefaultMTLVlan && network.NetworkName != "CHN" {
				err = files.WriteTemplate(
					filepath.Join(
						path,
						ifcfgFilename,
					),
					template.Must(template.New("vlan").Parse(string(VlanConfigTemplate))),
					data,
				)
				if err != nil {
					return fmt.Errorf(
						"failed to write %s because %v",
						ifcfgFilename,
						err,
					)
				}
			}
			if network.NetworkName == "NMN" {

				t, routes, err := genMetalLBTemplates(
					shastaNetworks,
				)
				if err != nil {
					return fmt.Errorf(
						"failed to generate metallb routes because %v",
						err,
					)
				}
				data := MakeTemplateData(routes)
				err = files.WriteTemplate(
					filepath.Join(
						path,
						ifrouteFilename,
					),
					t,
					data,
				)
				if err != nil {
					return fmt.Errorf(
						"failed to write %s because %v",
						ifrouteFilename,
						err,
					)
				}
			}
		}
	}
	return err
}

func genMetalLBTemplates(networks map[string]*networking.IPNetwork) (t *template.Template, routes Routes, err error) {
	metalLBPrefix, err := netip.ParsePrefix(networks["NMNLB"].CIDR4)
	if err != nil {
		return t, routes, fmt.Errorf(
			"failed to parse NMNLB network address because %v",
			err,
		)
	}
	nmnNetNet, _ := networks["NMN"].LookUpSubnet("network_hardware")

	mask, err := networking.PrefixLengthToSubnetMask(
		metalLBPrefix.Bits(),
		metalLBPrefix.Addr().BitLen(),
	)
	if err != nil {
		return t, routes, fmt.Errorf(
			"invalid metalLB network address [%s] because %v",
			nmnNetNet.Gateway,
			err,
		)
	}
	gateway, err := netip.ParseAddr(nmnNetNet.Gateway.String())
	if err != nil {
		return t, routes, fmt.Errorf(
			"invalid NMNLB gateway address [%s] because %v",
			nmnNetNet.Gateway,
			err,
		)
	}
	route := Route{
		Destination: metalLBPrefix.Addr().String(),
		Mask:        mask.String(),
		NextHop:     gateway.String(),
	}
	routes = append(
		routes,
		route,
	)
	t = template.Must(template.New("vlan").Parse(string(VlanRouteTemplate)))
	return t, routes, err
}

// VlanConfigTemplate is the text/template to bootstrap the install cd
var VlanConfigTemplate = []byte(`
{{- /* remove leading whitespace */ -}}
#
## This file was generated by cray-site-init.
## Version: {{ .Version }}
## Generated time: {{ .Timestamp }}
#
NAME='{{.Data.FullName}}'

# Set static IP (becomes "preferred" if dhcp is enabled)
BOOTPROTO='static'
IPADDR='{{.Data.CIDR4}}'
PREFIXLEN='{{.Data.CIDR4.Bits}}'

# CHANGE AT OWN RISK:
ETHERDEVICE='{{.Data.ParentInterfaceName}}'

# DO NOT CHANGE THESE:
VLAN_PROTOCOL='ieee802-1Q'
VLAN='yes'
VLAN_ID={{.Data.Vlan}}
ONBOOT='yes'
STARTMODE='auto'
`)

// VlanRouteTemplate allows us to add static routes to the vlan(s) on the PIT node
var VlanRouteTemplate = []byte(`
{{- /* remove leading whitespace */ -}}
#
## This file was generated by cray-site-init.
## Version: {{ .Version }}
## Generated time: {{ .Timestamp }}
#
{{ range .Data -}}
{{.Destination}} {{.NextHop}} {{.Mask}} -
{{ end -}}
`)

// BondConfigTemplate is the text/template for setting up the bond on the install NCN
var BondConfigTemplate = []byte(`
{{- /* remove leading whitespace */ -}}
#
## This file was generated by cray-site-init.
## Version: {{ .Version }}
## Generated time: {{ .Timestamp }}
#
NAME='Internal Interface'

# Set static IP (becomes "preferred" if dhcp is enabled)
BOOTPROTO='static'
IPADDR='{{.Data.CIDR}}'
PREFIXLEN='{{.Data.PrefixLen}}'

# CHANGE AT OWN RISK:
BONDING_MODULE_OPTS='mode=802.3ad miimon=100 lacp_rate=fast xmit_hash_policy=layer2+3'

# DO NOT CHANGE THESE:
ONBOOT='yes'
STARTMODE='auto'
BONDING_MASTER='yes'

# BOND MEMBERS:
{{ range $i, $v := .Data.Members -}}
BONDING_SLAVE{{$i}}='{{$v}}'
{{ end -}}
`)

// Lan0ConfigTemplate is the text/template for handling the external site link
var Lan0ConfigTemplate = []byte(`
{{- /* remove leading whitespace */ -}}
#
## This file was generated by cray-site-init.
## Version: {{ .Version }}
## Generated time: {{ .Timestamp }}
#
NAME='External Site-Link'

# Select the NIC(s) for direct, external access.
BRIDGE_PORTS='{{.Data.NIC}}'

# Set static IP (becomes "preferred" if dhcp is enabled)
# NOTE: IPADDR's route will override DHCPs.
BOOTPROTO='static'
IPADDR='{{.Data.IP}}'
PREFIXLEN='{{.Data.PrefixLen}}'

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
## Version: {{ .Version }}
## Generated time: {{ .Timestamp }}
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
NETCONFIG_DNS_STATIC_SERVERS="{{.Data.SiteDNS}}"
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
