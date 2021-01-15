/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

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
dhcp-option=interface:{{.VlanID | printf "vlan%03d"}},option:dns-server,{{.Gateway}}
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
dhcp-option=interface:bond0,option:dns-server,{{.Gateway}}
dhcp-option=interface:bond0,option:ntp-server,{{.Gateway}}
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
dhcp-host={{.NMNMac}},{{.NMNIP}},{{.Hostname}},20m # NMN
dhcp-host={{.NMNMac}},{{.MTLIP}},{{.Hostname}},20m # MTL
dhcp-host={{.NMNMac}},{{.HMNIP}},{{.Hostname}},20m # HMN
dhcp-host={{.NMNMac}},{{.CANIP}},{{.Hostname}},20m # CAN
dhcp-host={{.BMCMac}},{{.BMCIP}},{{.Hostname}}-mgmt,20m #HMN
# Host Record Entries for {{.Hostname}}
host-record={{.Hostname}},{{.Hostname}}.can,{{.CANIP}}
host-record={{.Hostname}},{{.Hostname}}.hmn,{{.HMNIP}}
host-record={{.Hostname}},{{.Hostname}}.nmn,{{.NMNIP}}
host-record={{.Hostname}},{{.Hostname}}.mtl,{{.MTLIP}}
{{end}}
# Virtual IP Addresses for k8s and the rados gateway
host-record=kubeapi-vip,kubeapi-vip.nmn,{{.KUBEVIP}} # k8s-virtual-ip
host-record=rgw-vip,rgw-vip.nmn,{{.RGWVIP}} # rgw-virtual-ip
host-record={{.APIGWALIASES}},{{.APIGWIP}} # api gateway

cname=kubernetes-api.vshasta.io,ncn-m001
`)

// DNSMasqBootstrapNetwork holds information for configuring DNSMasq on the LiveCD
type DNSMasqBootstrapNetwork struct {
	Subnet    IPV4Subnet
	Interface string
}
