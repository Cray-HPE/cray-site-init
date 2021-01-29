/*
Copyright 2020 Hewlett Packard Enterprise Development LP
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
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:dns-server,{{.DNSServer}}
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:ntp-server,{{.Gateway}}
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:router,{{.SupernetRouter}}
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
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:dns-server,{{.DNSServer}}
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
dhcp-host=id:{{.Xname}},set:{{.Hostname}},{{.NMNMac}},{{.NMNIP}},{{.Hostname}},20m # NMN
dhcp-host=id:{{.Xname}},set:{{.Hostname}},{{.NMNMac}},{{.MTLIP}},{{.Hostname}},20m # MTL
dhcp-host=id:{{.Xname}},set:{{.Hostname}},{{.NMNMac}},{{.HMNIP}},{{.Hostname}},20m # HMN
dhcp-host=id:{{.Xname}},set:{{.Hostname}},{{.NMNMac}},{{.CANIP}},{{.Hostname}},20m # CAN
dhcp-host={{.BMCMac}},{{.BMCIP}},{{.Hostname}}-mgmt,20m #HMN
# Host Record Entries for {{.Hostname}}
host-record={{.Hostname}},{{.Hostname}}.can,{{.CANIP}}
host-record={{.Hostname}},{{.Hostname}}.hmn,{{.HMNIP}}
host-record={{.Hostname}},{{.Hostname}}.nmn,{{.NMNIP}}
host-record={{.Hostname}},{{.Hostname}}.mtl,{{.MTLIP}}
host-record={{.Xname}},{{.Hostname}}.nmn,{{.NMNIP}}
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
	// DNSMasqNCN is the struct to manage NCNs within DNSMasq
	type DNSMasqNCN struct {
		Xname    string `form:"xname"`
		Hostname string `form:"hostname"`
		NMNMac   string `form:"nmn-mac"`
		NMNIP    string `form:"nmn-ip"`
		MTLMac   string `form:"mtl-mac"`
		MTLIP    string `form:"mtl-ip"`
		BMCMac   string `form:"bmc-mac"`
		BMCIP    string `form:"bmc-ip"`
		CANIP    string `form:"can-ip"`
		HMNIP    string `form:"can-ip"`
	}
	var ncns []DNSMasqNCN
	var mainMtlIP string
	for _, tmpNcn := range bootstrap {
		var canIP, nmnIP, hmnIP, mtlIP string
		for _, tmpNet := range tmpNcn.Networks {
			if tmpNet.NetworkName == "NMN" {
				nmnIP = tmpNet.IPAddress
			}
			if tmpNet.NetworkName == "CAN" {
				canIP = tmpNet.IPAddress
			}
			if tmpNet.NetworkName == "MTL" {
				if v.GetString("install-ncn") == tmpNcn.Hostname {
					mainMtlIP = tmpNet.IPAddress
				}
				mtlIP = tmpNet.IPAddress

			}
			if tmpNet.NetworkName == "HMN" {
				hmnIP = tmpNet.IPAddress
			}
		}
		// log.Println("Ready to build NCN list with:", v)
		ncn := DNSMasqNCN{
			Xname:    tmpNcn.Xname,
			Hostname: tmpNcn.Hostname,
			NMNMac:   tmpNcn.NmnMac,
			NMNIP:    nmnIP,
			// MTLMac:   ,
			MTLIP:  mtlIP,
			BMCMac: tmpNcn.BmcMac,
			BMCIP:  tmpNcn.BmcIP,
			CANIP:  canIP,
			HMNIP:  hmnIP,
		}
		ncns = append(ncns, ncn)
	}

	tpl1, _ := template.New("statics").Parse(string(StaticConfigTemplate))
	tpl2, _ := template.New("canconfig").Parse(string(CANConfigTemplate))
	tpl3, _ := template.New("hmnconfig").Parse(string(HMNConfigTemplate))
	tpl4, _ := template.New("nmnconfig").Parse(string(NMNConfigTemplate))
	tpl5, _ := template.New("mtlconfig").Parse(string(MTLConfigTemplate))

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
		NCNS         []DNSMasqNCN
		KUBEVIP      string
		RGWVIP       string
		APIGWALIASES string
		APIGWIP      string
	}{
		ncns,
		kubevip,
		rgwvip,
		apigwAliases,
		apigwIP,
	}
	// log.Println("Ready to write data with NCNs:", ncns)
	csiFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/statics.conf"), tpl1, data)

	// Save the install NCN Metal ip for use as dns/gateways during install

	// get a pointer to the MTL
	mtlNet := networks["MTL"]
	// get a pointer to the subnet
	mtlBootstrapSubnet, _ := mtlNet.LookUpSubnet("bootstrap_dhcp")
	tmpGateway := mtlBootstrapSubnet.Gateway
	mtlBootstrapSubnet.Gateway = net.ParseIP(mainMtlIP)
	mtlBootstrapSubnet.SupernetRouter = genPinnedIP(mtlBootstrapSubnet.CIDR.IP, uint8(1))
	nmnLBSubnet, _ := networks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	mtlBootstrapSubnet.DNSServer = nmnLBSubnet.LookupReservation("unbound").IPAddress
	csiFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/MTL.conf"), tpl5, mtlBootstrapSubnet)
	mtlBootstrapSubnet.Gateway = tmpGateway

	// Deal with the easy ones
	writeConfig("CAN", path, *tpl2, networks)
	writeConfig("HMN", path, *tpl3, networks)
	writeConfig("NMN", path, *tpl4, networks)
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
