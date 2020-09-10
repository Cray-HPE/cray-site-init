/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import "net"

// BootstrapNCNMetadata is a type that matches the ncn_metadata.csv file as
// NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
type BootstrapNCNMetadata struct {
	Xname   string `csv:"NCN xname"`
	Role    string `csv:"NCN Role"`
	Subrole string `csv:"NCN Subrole"`
	BmcMac  string `csv:"BMC MAC"`
	BmcPort string `csv:"BMC Switch Port"`
	NmnPac  string `csv:"NMN MAC"`
	NmnPort string `csv:"NMN Switch Port"`
}

// SystemConfig stores the overall set of system configuration parameters
type SystemConfig struct {
	SystemName       string
	SiteDomain       string
	InternalDomain   string
	MountainCabinets int16
	NtpPoolHostname  string
	IPV4Resolvers    []string
}

// SiteServices stores identity information for system services
type SiteServices struct {
	IPV4Resolvers   []net.IPAddr
	LDAPConn        LDAPConnection
	NtpPoolHostname string
	NtpHosts        []net.IPAddr
}

// ADGroup maps names and origins
type ADGroup struct {
	Name   string
	Origin string
}

// LDAPConnection stores details related to LDAP Server Provisioning
type LDAPConnection struct {
	Servers                  []string
	ADGroups                 []ADGroup
	BindDn                   string
	BindPassword             string
	Domain                   string
	SearchBase               string
	AttributeMappersToRemove []string
}
