/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

// CANConfigTemplate manages the CAN portion of the DNSMasq configuration
var CANConfigTemplate = []byte(`
# CAN:
server=/can/
address=/can/
dhcp-option=interface:vlan007,option:domain-search,can
interface-name=pit.can,vlan007
interface=vlan007
cname=packages.can,pit.can
cname=registry.can,pit.can
dhcp-option=interface:vlan007,option:router,{{.Gateway}}
dhcp-range=interface:vlan007,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// HMNConfigTemplate manages the HMN portion of the DNSMasq configuration
var HMNConfigTemplate = []byte(`
# HMN:
server=/hmn/
address=/hmn/
domain=hmn,{{.DHCPStart}},{{.DHCPEnd}},local
interface-name=pit.hmn,vlan004
dhcp-option=interace:vlan004,option:domain-search,hmn
interface=vlan004
cname=packages.hmn,pit.hmn
cname=registry.hmn,pit.hmn
dhcp-option=interface:vlan004,option:dns-server,{{.Gateway}}
dhcp-option=interface:vlan004,option:ntp-server,{{.Gateway}}
dhcp-option=interface:vlan004,option:router,{{.Gateway}}
dhcp-range=interface:vlan004,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// MTLConfigTemplate manages the MTL portion of the DNSMasq configuration
var MTLConfigTemplate = []byte(`
# MTL:
server=/mtl/
address=/mtl/
domain=mtl,{{.DHCPStart}},{{.DHCPEnd}},local
dhcp-option=interface:bond0,option:domain-search,mtl
interface=bond0
interface-name=pit.mtl,bond0
cname=packages.mtl,pit.mtl
cname=registry.mtl,pit.mtl
cname=packages.local,pit.mtl
cname=registry.local,pit.mtl
dhcp-option=interface:bond0,option:dns-server,{{.Gateway}}
dhcp-option=interface:bond0,option:ntp-server,{{.Gateway}}
dhcp-option=interface:bond0,option:router,{{.Gateway}}
dhcp-range=interface:bond0,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// NMNConfigTemplate manages the NMN portion of the DNSMasq configuration
var NMNConfigTemplate = []byte(`
# NMN:
server=/nmn/
address=/nmn/
interface-name=pit.nmn,vlan002
domain=nmn,{{.DHCPStart}},{{.DHCPEnd}},local
dhcp-option=interface:vlan002,option:domain-search,nmn
interface=vlan002
cname=packages.nmn,pit.nmn
cname=registry.nmn,pit.nmn
dhcp-option=interface:vlan002,option:dns-server,{{.Gateway}}
dhcp-option=interface:vlan002,option:ntp-server,{{.Gateway}}
dhcp-option=interface:vlan002,option:router,{{.Gateway}}
dhcp-range=interface:vlan002,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// StaticConfigTemplate manages the static portion of the DNSMasq configuration
// Systems with onboard NICs will have a MTL MAC.  Others will also use the NMN
var StaticConfigTemplate = []byte(`
# Static Configurations
{{range .}}
# DHCP Entries for {{.Hostname}}
dhcp-host={{.NMNMac}},{{.NMNIP}},{{.Hostname}},infinite # NMN
dhcp-host={{.MTLMac}},{{.MTLIP}},{{.Hostname}},infinite # MTL
dhcp-host={{.NMNMac}},{{.HMNIP}},{{.Hostname}},infinite # HMN
dhcp-host={{.NMNMac}},{{.CANIP}},{{.Hostname}},infinite # CAN
dhcp-host={{.BMCMac}},{{.BMCIP}},{{.Hostname}}-mgmt,infinite #HMN
# Host Record Entries for {{.Hostname}}
host-record={{.Hostname}},{{.Hostname}}.can,{{.CANIP}}
host-record={{.Hostname}},{{.Hostname}}.hmn,{{.HMNIP}}
host-record={{.Hostname}},{{.Hostname}}.nmn
host-record={{.Hostname}},{{.Hostname}}.mtl
{{end -}}
cname=kubernetes-api.vshasta.io,ncn-m001
`)

// DNSMasqBootstrapNetwork holds information for configuring DNSMasq on the LiveCD
type DNSMasqBootstrapNetwork struct {
	Subnet    IPV4Subnet
	Interface string
}
