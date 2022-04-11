//
//  MIT License
//
//  (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.

package pit

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"text/template"

	csiFiles "github.com/Cray-HPE/csm-common/go/internal/files"
	"github.com/Cray-HPE/csm-common/go/pkg/csi"
	"github.com/spf13/viper"
)

// CMNConfigTemplate manages the CAN portion of the DNSMasq configuration
var CMNConfigTemplate = []byte(`
# CMN:
server=/cmn/
address=/cmn/
dhcp-option=interface:bond0.cmn0,option:domain-search,cmn
interface-name=pit.cmn,bond0.cmn0
interface=bond0.cmn0
cname=packages.cmn,pit.cmn
cname=registry.cmn,pit.cmn
dhcp-option=interface:bond0.cmn0,option:router,{{.Gateway}}
dhcp-range=interface:bond0.cmn0,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// CANConfigTemplate manages the CAN portion of the DNSMasq configuration
var CANConfigTemplate = []byte(`
# CAN:
server=/can/
address=/can/
dhcp-option=interface:bond0.can0,option:domain-search,can
interface-name=pit.can,bond0.can0
interface=bond0.can0
cname=packages.can,pit.can
cname=registry.can,pit.can
dhcp-option=interface:bond0.can0,option:router,{{.Gateway}}
dhcp-range=interface:bond0.can0,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// HMNConfigTemplate manages the HMN portion of the DNSMasq configuration typically bond0.hmn0
var HMNConfigTemplate = []byte(`
# HMN:
server=/hmn/
address=/hmn/
domain=hmn,{{.CIDR.IP}},{{.DHCPEnd}},local
interface-name=pit.hmn,bond0.hmn0
dhcp-option=interace:bond0.hmn0,option:domain-search,hmn
interface=bond0.hmn0
cname=packages.hmn,pit.hmn
cname=registry.hmn,pit.hmn
# This needs to point to the liveCD IP for provisioning in bare-metal environments.
dhcp-option=interface:bond0.hmn0,option:dns-server,{{.PITServer}}
dhcp-option=interface:bond0.hmn0,option:ntp-server,{{.PITServer}}
dhcp-option=interface:bond0.hmn0,option:router,{{.Gateway}}
dhcp-range=interface:bond0.hmn0,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// MTLConfigTemplate manages the MTL portion of the DNSMasq configuration
var MTLConfigTemplate = []byte(`
# MTL:
server=/mtl/
address=/mtl/
domain=mtl,{{.CIDR.IP}},{{.DHCPEnd}},local
dhcp-option=interface:bond0,option:domain-search,mtl
interface=bond0
interface-name=pit.mtl,bond0
# This needs to point to the liveCD IP for provisioning in bare-metal environments.
dhcp-option=interface:bond0,option:dns-server,{{.PITServer}}
dhcp-option=interface:bond0,option:ntp-server,{{.PITServer}}
# This must point at the router for the network; the L3/IP for the VLAN.
dhcp-option=interface:bond0,option:router,{{.Gateway}}
dhcp-range=interface:bond0,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// NMNConfigTemplate manages the NMN portion of the DNSMasq configuration
var NMNConfigTemplate = []byte(`
# NMN:
server=/nmn/
address=/nmn/
interface-name=pit.nmn,bond0.nmn0
domain=nmn,{{.CIDR.IP}},{{.DHCPEnd}},local
dhcp-option=interface:bond0.nmn0,option:domain-search,nmn
interface=bond0.nmn0
cname=packages.nmn,pit.nmn
cname=registry.nmn,pit.nmn
# This needs to point to the liveCD IP for provisioning in bare-metal environments.
dhcp-option=interface:bond0.nmn0,option:dns-server,{{.PITServer}}
dhcp-option=interface:bond0.nmn0,option:ntp-server,{{.PITServer}}
dhcp-option=interface:bond0.nmn0,option:router,{{.Gateway}}
dhcp-range=interface:bond0.nmn0,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// StaticConfigTemplate manages the static portion of the DNSMasq configuration
// Systems with onboard NICs will have a MTL MAC.  Others will also use the NMN
var StaticConfigTemplate = []byte(`
# Static Configurations
{{range .NCNS}}
# DHCP Entries for {{.Hostname}}
dhcp-host=id:{{.Xname}},set:{{.Hostname}},{{.Bond0Mac0}},{{.Bond0Mac1}},{{.MtlIP}},{{.Hostname}},20m # MTL
dhcp-host=id:{{.Xname}},set:{{.Hostname}},{{.Bond0Mac0}},{{.Bond0Mac1}},{{.NmnIP}},{{.Hostname}},20m # Bond0 Mac0/Mac1
dhcp-host=id:{{.Xname}},set:{{.Hostname}},{{.Bond0Mac0}},{{.Bond0Mac1}},{{.HmnIP}},{{.Hostname}},20m # HMN
dhcp-host=id:{{.Xname}},set:{{.Hostname}},{{.Bond0Mac0}},{{.Bond0Mac1}},{{.CanIP}},{{.Hostname}},20m # CAN
dhcp-host={{.BmcMac}},{{.BmcIP}},{{.Hostname}}-mgmt,20m #HMN
# Host Record Entries for {{.Hostname}}
host-record={{.Hostname}},{{.Hostname}}.can,{{.CanIP}}
host-record={{.Hostname}},{{.Hostname}}.hmn,{{.HmnIP}}
host-record={{.Hostname}},{{.Hostname}}.nmn,{{.NmnIP}}
host-record={{.Hostname}},{{.Hostname}}.mtl,{{.MtlIP}}
host-record={{.Xname}},{{.Hostname}}.nmn,{{.NmnIP}}
host-record={{.Hostname}}-mgmt,{{.Hostname}}-mgmt.hmn,{{.BmcIP}}
# Override root-path with {{.Hostname}}'s xname
dhcp-option-force=tag:{{.Hostname}},17,{{.Xname}}
{{end}}
# Virtual IP Addresses for k8s and the rados gateway
host-record=kubeapi-vip,kubeapi-vip.nmn,{{.KUBEVIP}} # k8s-virtual-ip
host-record=rgw-vip,rgw-vip.nmn,{{.RGWVIP}} # rgw-virtual-ip
host-record={{.APIGWALIASES}},{{.APIGWIP}} # api gateway

cname=kubernetes-api.vshasta.io,ncn-m001
`)

// DNSMasqBootstrapNetwork holds information for configuring DNSMasq on the LiveCD
type DNSMasqBootstrapNetwork struct {
	Subnet    csi.IPV4Subnet
	Interface string
}

// WriteDNSMasqConfig writes the dnsmasq configuration files necssary for installation
func WriteDNSMasqConfig(path string, v *viper.Viper, bootstrap []csi.LogicalNCN, networks map[string]*csi.IPV4Network) {
	for i, tmpNcn := range bootstrap {
		for _, tmpNet := range tmpNcn.Networks {
			if tmpNet.NetworkName == "NMN" {
				tmpNcn.NmnIP = tmpNet.IPAddress
			}
			if tmpNet.NetworkName == "CAN" {
				tmpNcn.CanIP = tmpNet.IPAddress
			}
			if tmpNet.NetworkName == "MTL" {
				tmpNcn.MtlIP = tmpNet.IPAddress
			}
			if tmpNet.NetworkName == "HMN" {
				tmpNcn.HmnIP = tmpNet.IPAddress
			}
		}
		bootstrap[i] = tmpNcn
	}
	var kubevip, rgwvip string
	nmnSubnet, _ := networks["NMN"].LookUpSubnet("bootstrap_dhcp")
	for _, reservation := range nmnSubnet.IPReservations {
		if reservation.Name == "kubeapi-vip" {
			kubevip = reservation.IPAddress.String()
		}
		if reservation.Name == "rgw-vip" {
			rgwvip = reservation.IPAddress.String()
		}
	}

	var apigwAliases, apigwIP string
	nmnlbNet, _ := networks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	apigw := nmnlbNet.ReservationsByName()["istio-ingressgateway"]
	apigwAliases = strings.Join(apigw.Aliases, ",")
	apigwIP = apigw.IPAddress.String()

	data := struct {
		NCNS         []csi.LogicalNCN
		KUBEVIP      string
		RGWVIP       string
		APIGWALIASES string
		APIGWIP      string
	}{
		bootstrap,
		kubevip,
		rgwvip,
		apigwAliases,
		apigwIP,
	}

	// Shasta Networks:
	netCMN, _ := template.New("cmnconfig").Parse(string(CMNConfigTemplate))
	netHMN, _ := template.New("hmnconfig").Parse(string(HMNConfigTemplate))
	netNMN, _ := template.New("nmnconfig").Parse(string(NMNConfigTemplate))
	netMTL, _ := template.New("mtlconfig").Parse(string(MTLConfigTemplate))
	writeConfig("CMN", path, *netCMN, networks)
	writeConfig("HMN", path, *netHMN, networks)
	writeConfig("NMN", path, *netNMN, networks)
	writeConfig("MTL", path, *netMTL, networks)
	// Work some BICAN required magic
	if v.GetString("bican-user-network-name") == "CAN" || v.GetBool("retain-unused-user-network") {
		netCAN, _ := template.New("canconfig").Parse(string(CANConfigTemplate))
		writeConfig("CAN", path, *netCAN, networks)
	}

	// Expected NCNs (and other devices) reserved DHCP leases:
	netIPAM, _ := template.New("statics").Parse(string(StaticConfigTemplate))
	csiFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/statics.conf"), netIPAM, data)
}

func writeConfig(name, path string, tpl template.Template, networks map[string]*csi.IPV4Network) {
	// Pointer to the IPV4Network
	tempNet := networks[name]

	v := viper.GetViper()

	// Pointer to the subnet
	bootstrapSubnet, _ := tempNet.LookUpSubnet("bootstrap_dhcp")
	// Create a subnet copy (avoid modifying the base data with dnsmasq overrides)
	tempSubnet := *bootstrapSubnet

	// Look up the PIT IP for the network
	for _, reservation := range tempSubnet.IPReservations {
		if reservation.Name == v.GetString("install-ncn") {
			tempSubnet.PITServer = reservation.IPAddress
		}
	}
	if tempNet.Name == "CAN" {
		tempSubnet.Gateway = net.ParseIP(v.GetString("can-gateway"))
	}
	if tempNet.Name == "CMN" {
		tempSubnet.Gateway = net.ParseIP(v.GetString("cmn-gateway"))
	}

	nmnLBSubnet, _ := networks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	tempSubnet.DNSServer = nmnLBSubnet.LookupReservation("unbound").IPAddress
	csiFiles.WriteTemplate(filepath.Join(path, fmt.Sprintf("dnsmasq.d/%v.conf", name)), &tpl, tempSubnet)
}
