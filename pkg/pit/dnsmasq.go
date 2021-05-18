/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package pit

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/viper"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/csi"
)

// CANConfigTemplate manages the CAN portion of the DNSMasq configuration
var CANConfigTemplate = []byte(`
# CAN:
server=/can/
address=/can/
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:domain-search,can
interface-name=pit.can,{{.VlanID | printf "vlan%03d"}}
interface={{.VlanID | printf "vlan%03d"}}
cname=packages.can,pit.can
cname=registry.can,pit.can
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:router,{{.Gateway}}
dhcp-range=interface:{{.VlanID | printf "vlan%03d"}},{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// HMNConfigTemplate manages the HMN portion of the DNSMasq configuration typically vlan004
var HMNConfigTemplate = []byte(`
# HMN:
server=/hmn/
address=/hmn/
domain=hmn,{{.CIDR.IP}},{{.DHCPEnd}},local
interface-name=pit.hmn,{{.VlanID | printf "vlan%03d"}}
dhcp-option=interace:{{.VlanID | printf "vlan%03d"}},option:domain-search,hmn
interface={{.VlanID | printf "vlan%03d"}}
cname=packages.hmn,pit.hmn
cname=registry.hmn,pit.hmn
# This needs to point to the liveCD IP for provisioning in bare-metal environments.
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:dns-server,{{.Gateway}}
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:ntp-server,{{.Gateway}}
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:router,{{.Gateway}}
dhcp-range=interface:{{.VlanID | printf "vlan%03d"}},{{.DHCPStart}},{{.DHCPEnd}},10m
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
dhcp-option=interface:bond0,option:dns-server,{{.Gateway}}
dhcp-option=interface:bond0,option:ntp-server,{{.Gateway}}
# This must point at the router for the network; the L3/IP for the VLAN.
dhcp-option=interface:bond0,option:router,{{.SupernetRouter}}
dhcp-range=interface:bond0,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// NMNConfigTemplate manages the NMN portion of the DNSMasq configuration
var NMNConfigTemplate = []byte(`
# NMN:
server=/nmn/
address=/nmn/
interface-name=pit.nmn,{{.VlanID | printf "vlan%03d"}}
domain=nmn,{{.CIDR.IP}},{{.DHCPEnd}},local
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:domain-search,nmn
interface={{.VlanID | printf "vlan%03d"}}
cname=packages.nmn,pit.nmn
cname=registry.nmn,pit.nmn
# This needs to point to the liveCD IP for provisioning in bare-metal environments.
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:dns-server,{{.Gateway}}
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:ntp-server,{{.Gateway}}
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:router,{{.SupernetRouter}}
dhcp-range=interface:{{.VlanID | printf "vlan%03d"}},{{.DHCPStart}},{{.DHCPEnd}},10m
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
	netCAN, _ := template.New("canconfig").Parse(string(CANConfigTemplate))
	netHMN, _ := template.New("hmnconfig").Parse(string(HMNConfigTemplate))
	netNMN, _ := template.New("nmnconfig").Parse(string(NMNConfigTemplate))
	netMTL, _ := template.New("mtlconfig").Parse(string(MTLConfigTemplate))
	writeConfig("CAN", path, *netCAN, networks)
	writeConfig("HMN", path, *netHMN, networks)
	writeConfig("NMN", path, *netNMN, networks)
	writeConfig("MTL", path, *netMTL, networks)

	// Expected NCNs (and other devices) reserved DHCP leases:
	netIPAM, _ := template.New("statics").Parse(string(StaticConfigTemplate))
	csiFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/statics.conf"), netIPAM, data)
}

func writeConfig(name, path string, tpl template.Template, networks map[string]*csi.IPV4Network) {
	// get a pointer to the IPV4Network
	tempNet := networks[name]
	// get a pointer to the subnet
	v := viper.GetViper()
	bootstrapSubnet, _ := tempNet.LookUpSubnet("bootstrap_dhcp")
	for _, reservation := range bootstrapSubnet.IPReservations {
		if reservation.Name == v.GetString("install-ncn") {
			bootstrapSubnet.Gateway = reservation.IPAddress
		}
	}
	if tempNet.Name == "CAN" {
		bootstrapSubnet.Gateway = net.ParseIP(v.GetString("can-gateway"))
	}
	// Normalize the CIDR before using it
	_, superNet, _ := net.ParseCIDR(bootstrapSubnet.CIDR.String())
	bootstrapSubnet.SupernetRouter = genPinnedIP(superNet.IP, uint8(1))
	nmnLBSubnet, _ := networks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	bootstrapSubnet.DNSServer = nmnLBSubnet.LookupReservation("unbound").IPAddress
	csiFiles.WriteTemplate(filepath.Join(path, fmt.Sprintf("dnsmasq.d/%v.conf", name)), &tpl, bootstrapSubnet)
}

func genPinnedIP(ip net.IP, pin uint8) net.IP {
	newIP := make(net.IP, 4)
	if len(ip) == 4 {
		newIP[0] = ip[0]
		newIP[1] = ip[1]
		newIP[2] = ip[2]
		newIP[3] = pin
	}
	if len(ip) == 16 {
		newIP[0] = ip[12]
		newIP[1] = ip[13]
		newIP[2] = ip[14]
		newIP[3] = pin
	}
	return newIP
}
