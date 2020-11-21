/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"crypto/rand"
	"fmt"
	"net"
)

// BootstrapSwitchMetadata is a type that matches the switch_metadata.csv file as
// Switch Xname, Type
// The type can be CDU, Spine, Aggregation, or Leaf
type BootstrapSwitchMetadata struct {
	Xname string `json:"xname" csv:"Switch Xname"`
	Type  string `json:"type" csv:"Type"`
}

// NewBootstrapNCNMetadata is a type that matches the updated ncn_metadata.csv file as
// Xname,Role,Subrole,BMC MAC,Bootstrap MAC,Bond0 MAC0,Bond0 Mac1
// It is probable that on many machines bootstrap mac will be the same as one of the bond macs
// Do not be alarmed.
type NewBootstrapNCNMetadata struct {
	Xname        string `json:"xname" csv:"Xname"`
	Role         string `json:"role" csv:"Role"`
	Subrole      string `json:"subrole" csv:"Subrole"`
	BmcMac       string `json:"bmc-mac" csv:"BMC MAC"`
	BootstrapMac string `json:"bootstrap-mac" csv:"Bootstrap MAC"`
	Bond0Mac0    string `json:"bond0-mac0" csv:"Bond0 MAC0"`
	Bond0Mac1    string `json:"bond0-mac1" csv:"Bond0 MAC1"`
}

// SystemConfig stores the overall set of system configuration parameters
type SystemConfig struct {
	SystemName      string       `form:"system-name" mapstructure:"system-name"`
	SiteDomain      string       `form:"site-domain" mapstructure:"site-domain"`
	InternalDomain  string       `form:"internal-domain" mapstructure:"internal-domain"`
	Cabinets        int16        `form:"cabinets" mapstructure:"cabinets"`
	StartingCabinet int16        `form:"starting-cabinet" mapstructure:"starting-cabinet"`
	StartingNID     int          `form:"starting-NID" mapstructure:"starting-NID"`
	NtpPoolHostname string       `form:"ntp-pool" mapstructure:"ntp-pool"`
	NtpHosts        []string     `form:"ntp-hosts" mapstructure:"ntp-hosts"`
	IPV4Resolvers   []string     `form:"ipv4-resolvers" mapstructure:"ipv4-resolvers"`
	V2Registry      string       `form:"v2-registry" mapstructure:"v2-registry"`
	RpmRegistry     string       `form:"rpm-repository" mapstructure:"rpm-repository"`
	NMNCidr         string       `form:"nmn-cidr" mapstructure:"nmn-cidr"`
	HMNCidr         string       `form:"hmn-cidr" mapstructure:"hmn-cidr"`
	CANCidr         string       `form:"can-cidr" mapstructure:"can-cidr"`
	MTLCidr         string       `form:"mtl-cidr" mapstructure:"mtl-cidr"`
	HSNCidr         string       `form:"hsn-cidr" mapstructure:"hsn-cidr"`
	SiteServices    SiteServices `form:"site-services" mapstructure:"site-services"`
}

// CabinetDetail stores information that can only come from Manufacturing
type CabinetDetail struct {
	Kind            string `mapstructure:"cabinet-type"`
	Cabinets        int    `mapstructure:"cabinets"`
	StartingCabinet int    `mapstructure:"starting-cabinet"`
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
func (node LogicalNCN) GetHostname() string {
	if node.Hostname == "" {
		return node.Xname
	}
	return node.Hostname
}
