/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// DefaultSystemConfigYamlTemplate is the go template for the Default System Config
var DefaultSystemConfigYamlTemplate = []byte(`
---
system_settings:
  system_name: {{.SystemName}}
  domain_name: {{.SiteDomain}}
  internal_domain: {{.InternalDomain}}
  site_services:
    {{if .IPV4Resolvers}}
	ipv4_resolvers:
	{{range .IPV4Resolvers}}
	- {{.}}
	{{end}}
	{{end -}}
	{{if .NtpPoolHostname}}
	ntp_pool_hostname: {{.NtpPoolHostname}}
	{{else if .NtpHosts}}
	upstream_ntp:
	{{range .NtpHosts}}
	  - {{.}}
	{{end}}
	{{end}}
    ldap:
      ad_groups:
      - name: admin_grp
        origin: ALL
      - name: dev_users
        origin: ALL
      bind_dn: CN=username,CN=Users,DC=yourdomain,DC=com
      bind_password: initial0
      domain: Cray_DC
      search_base: dc=datacenter,dc=cray,dc=com
      servers:
      - ldaps://ldap_server:port/dn?attributes?scope?filter?extensions
      user_attribute_mappers_to_remove:
	  - email


manufacturing_details:
  shcd_revision: D
  mountain_starting_nid: 20000
  river_layout:
    - cabinet_id: x3001
      compute_nodes: []
      worker: [14,16,18]
      master: [8,10,12]
      storage: [2,4,6]
  mountain_layout:
    number_of_cabinets: {{.MountainCabinets}}
    starting_cabinet: 1004
	last_address_gateway: True
	`)

// DefaultSwitchConfigYaml is the default yaml for Switch Configuration
// Manually from the River Rack Diagram
var DefaultSwitchConfigYaml = []byte(`
  # Do sw-100g01, and sw-100g02 need to be represented here?
  management_switches:
    -  {'ID': 'x3001c0w38R', 'NAME': 'sw-spine01'}  # sw-40g01
    -  {'ID': 'x3001c0w34', 'NAME': 'sw-leaf01', 'Brand': 'Dell', 'OS': 'OS10', 'Model': 'S3048-ON', "DHCP Relay": True}
    -  {'ID': 'x3001c0w38L', 'NAME': 'sw-spine02'}  # sw-40g02
    -  {'ID': 'x3001c0w35', 'NAME': 'sw-leaf02', 'Brand': 'Dell', 'OS': 'OS10', 'Model': 'S3048-ON', "DHCP Relay": True}
    -  {'ID': 'd0w1', 'NAME': 'sw-cdu01', 'Brand': 'Dell', 'OS': 'OS10', 'Model': 'S4148T', "DHCP Relay": True}
    -  {'ID': 'd0w2', 'NAME': 'sw-cdu02', 'Brand': 'Dell', 'OS': 'OS10', 'Model': 'S4148T', "DHCP Relay": True}

  columbia_switches:
    -  {'ID': 'x3001c0r39b0', 'NAME': 'sw-hsn01'}
    -  {'ID': 'x3001c0r40b0', 'NAME': 'sw-hsn02'}
    -  {'ID': 'x3001c0r41b0', 'NAME': 'sw-hsn03'}
    -  {'ID': 'x3001c0r42b0', 'NAME': 'sw-hsn04'}
`)

// DefaultNetworkConfigYamlTemplate is the go template for default network configuration yaml
var DefaultNetworkConfigYamlTemplate = []byte(`
full_name: {{.FullName}}
cidr: {{.CIDR}}
name: {{.Name}}
vlan_range: {{.VlanRange}}
mtu: {{.MTU}}
type: {{.NetType}}
comment: {{.Comment}}
{{if .Subnets}}
subnets:
{{range .Subnets}}
{{if .Name}}
  - name: {{.Name}}
{{else}}
- name: default
{{end}}
    cidr: {{.CIDR}}
	{{if .Gateway}}
    gateway: {{.Gateway}}
	{{end}}
{{end}}
{{end}}
`)

// DefaultHMN is the default structure for templating initial HMN configuration
var DefaultHMN = shasta.IPV4Network{
	FullName:  "Hardware Management Network",
	CIDR:      "10.254.0.0/17",
	Name:      "hmn",
	VlanRange: []int16{60, 100},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "",
}

// DefaultNMN is the default structure for templating initial NMN configuration
var DefaultNMN = shasta.IPV4Network{
	FullName:  "Node Management Network",
	CIDR:      "10.242.0.0/17",
	Name:      "nmn",
	VlanRange: []int16{60, 100},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "",
}

// DefaultHSN is the default structure for templating initial HSN configuration
var DefaultHSN = shasta.IPV4Network{
	FullName:  "High Speed Network",
	CIDR:      "10.253.0.0/16",
	Name:      "hsn",
	VlanRange: []int16{60, 100},
	MTU:       9000,
	NetType:   "slingshot10",
	Comment:   "",
}

// DefaultCAN is the default structure for templating initial CAN configuration
var DefaultCAN = shasta.IPV4Network{
	FullName:  "High Speed Network",
	CIDR:      "192.168.20.0/24",
	Name:      "can",
	VlanRange: []int16{60, 100},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "",
}

// DefaultMTL is the default structure for templating initial MTL configuration
var DefaultMTL = shasta.IPV4Network{
	FullName:  "Provisioning Network (untagged)",
	CIDR:      "192.168.1.0/24",
	Name:      "mtl",
	VlanRange: []int16{60, 100},
	MTU:       9000,
	NetType:   "ethernet",
	Comment:   "This network is only valid for the NCNs",
}

// DefaultRootPW is the default root password
var DefaultRootPW = shasta.PasswordCredential{
	Username: "root",
	Password: "changem3",
}

// DefaultBMCPW is the default root password
var DefaultBMCPW = shasta.PasswordCredential{
	Username: "root",
	Password: "changem3",
}

// DefaultNetPW is the default root password
var DefaultNetPW = shasta.PasswordCredential{
	Username: "root",
	Password: "changem3",
}
