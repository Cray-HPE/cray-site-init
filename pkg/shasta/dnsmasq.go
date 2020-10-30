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
interface-name=spit.can,vlan007
interface=vlan007
cname=packages.can,spit.can
cname=registry.can,spit.can
dhcp-option=interface:vlan007,option:router,{{.Gateway}}
dhcp-range=interface:vlan007,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// HMNConfigTemplate manages the HMN portion of the DNSMasq configuration
var HMNConfigTemplate = []byte(`
# HMN:
server=/hmn/
address=/hmn/
domain=hmn,{{.DHCPStart}},{{.DHCPEnd}},local
interface-name=spit.hmn,vlan004
dhcp-option=interace:vlan004,option:domain-search,hmn
interface=vlan004
cname=packages.hmn,spit.hmn
cname=registry.hmn,spit.hmn
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
interface-name=spit.mtl,bond0
cname=packages.mtl,spit.mtl
cname=registry.mtl,spit.mtl
cname=packages.local,spit.mtl
cname=registry.local,spit.mtl
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
interface-name=spit.nmn,vlan002
domain=nmn,{{.DHCPStart}},{{.DHCPEnd}},local
dhcp-option=interface:vlan002,option:domain-search,nmn
interface=vlan002
cname=packages.nmn,spit.nmn
cname=registry.nmn,spit.nmn
dhcp-option=interface:vlan002,option:dns-server,{{.Gateway}}
dhcp-option=interface:vlan002,option:ntp-server,{{.Gateway}}
dhcp-option=interface:vlan002,option:router,{{.Gateway}}
dhcp-range=interface:vlan002,{{.DHCPStart}},{{.DHCPEnd}},10m
`)

// StaticConfigTemplate manages the static portion of the DNSMasq configuration
var StaticConfigTemplate = []byte(`
# Static Configurations
{{range .}}
# DHCP Entries for {{.Hostname}}
dhcp-host={{.NMNMac}},{{.NMNIP}},{{.Hostname}},infinite # NMN
dhcp-host={{.MTLMac}},{{.MTLIP}},{{.Hostname}},infinite # MTL
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
