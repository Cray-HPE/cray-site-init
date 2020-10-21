/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"crypto/rand"
	"fmt"
	"net"
)

// BootstrapNCNMetadata is a type that matches the ncn_metadata.csv file as
// NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
type BootstrapNCNMetadata struct {
	Xname     string   `form:"xname" csv:"NCN xname"`
	Role      string   `form:"role" csv:"NCN Role"`
	Subrole   string   `form:"subrole" csv:"NCN Subrole"`
	BmcMac    string   `form:"bmc-mac" csv:"BMC MAC"`
	BmcPort   string   `form:"bmc-port" csv:"BMC Switch Port"`
	NmnMac    string   `form:"nmn-port" csv:"NMN MAC"`
	NmnPort   string   `form:"nmn-port" csv:"NMN Switch Port"`
	Hostnames []string `form:"hostnames" csv:"-"`
}

// SystemConfig stores the overall set of system configuration parameters
type SystemConfig struct {
	SystemName       string       `form:"system-name" mapstructure:"system-name"`
	SiteDomain       string       `form:"site-domain" mapstructure:"site-domain"`
	InternalDomain   string       `form:"internal-domain" mapstructure:"internal-domain"`
	MountainCabinets int16        `form:"mountain-cabinets" mapstructure:"mountain-cabinets"`
	StartingCabinet  int16        `form:"starting-cabinet" mapstructure:"starting-cabinet"`
	StartingNID      int          `form:"starting-NID" mapstructure:"starting-NID"`
	NtpPoolHostname  string       `form:"ntp-pool" mapstructure:"ntp-pool"`
	NtpHosts         []string     `form:"ntp-hosts" mapstructure:"ntp-hosts"`
	IPV4Resolvers    []string     `form:"ipv4-resolvers" mapstructure:"ipv4-resolvers"`
	V2Registry       string       `form:"v2-registry" mapstructure:"v2-registry"`
	RpmRegistry      string       `form:"rpm-repository" mapstructure:"rpm-repository"`
	NMNCidr          string       `form:"nmn-cidr" mapstructure:"nmn-cidr"`
	HMNCidr          string       `form:"hmn-cidr" mapstructure:"hmn-cidr"`
	CANCidr          string       `form:"can-cidr" mapstructure:"can-cidr"`
	MTLCidr          string       `form:"mtl-cidr" mapstructure:"mtl-cidr"`
	HSNCidr          string       `form:"hsn-cidr" mapstructure:"hsn-cidr"`
	SiteServices     SiteServices `form:"site-services" mapstructure:"site-services"`
}

// HardwareDetail stores information that can only come from Manufacturing
type HardwareDetail struct {
	MountainCabinets int16 `form:"mountain-cabinets" mapstructure:"mountain-cabinets"`
	StartingCabinet  int16 `form:"starting-cabinet" mapstructure:"starting-cabinet"`
}

// BGPPeering stores information about MetalLB Peering
type BGPPeering struct {
	// the two ends of the turtle
}

// PointToPoint is a structure for storing the Basics of Network Management
type PointToPoint struct {
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

// GenerateInstanceID creates an instance-id fit for use in the instance metadata
func GenerateInstanceID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("i-%X", b)
}

// GetHostname returns an explicit hostname if possible, otherwise the Xname, otherwise an empty string
func (node BootstrapNCNMetadata) GetHostname() string {
	if len(node.Hostnames) > 0 {
		return node.Hostnames[1]
	}
	if node.Xname != "" {
		return node.Xname
	}
	return ""
}
