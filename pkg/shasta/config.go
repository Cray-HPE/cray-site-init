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
	Xname   string `form:"xname" csv:"NCN xname"`
	Role    string `form:"role" csv:"NCN Role"`
	Subrole string `form:"subrole" csv:"NCN Subrole"`
	BmcMac  string `form:"bmc-mac" csv:"BMC MAC"`
	BmcPort string `form:"bmc-port" csv:"BMC Switch Port"`
	NmnPac  string `form:"nmn-port" csv:"NMN MAC"`
	NmnPort string `form:"nmn-port" csv:"NMN Switch Port"`
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

// PhoneHome should not exist in data.json before installation has started
type PhoneHome struct {
	PublicKeyDSA     string `form:"pub_key_dsa" json:"pub_key_dsa" binding:"omitempty"`
	PublicKeyRSA     string `form:"pub_key_rsa" json:"pub_key_rsa" binding:"omitempty"`
	PublicKeyECDSA   string `form:"pub_key_ecdsa" json:"pub_key_ecdsa" binding:"omitempty"`
	PublicKeyED25519 string `form:"pub_key_ed25519" json:"pub_key_ed25519,omitempty"`
	InstanceID       string `form:"instance_id" json:"instance_id" binding:"omitempty"`
	Hostname         string `form:"hostname" json:"hostname" binding:"omitempty"`
	FQDN             string `form:"fdqn" json:"fdqn" binding:"omitempty"`
}

// MetaData is part of the cloud-init stucture and
// is only used for validating the required fields in the
// `CloudInit` struct below.
type MetaData struct {
	Hostname         string `form:"local-hostname" json:"local-hostname"`       // should be xname
	InstanceID       string `form:"instance-id" json:"instance-id"`             // should be unique for the life of the image
	Region           string `form:"region" json:"region"`                       // unused currently
	AvailabilityZone string `form:"availability-zone" json:"availability-zone"` // unused currently
	ShastaRole       string `form:"shasta-role" json:"shasta-role"`             // map to HSM role
}

// CloudInit is the main cloud-init struct. Leave the meta-data, user-data, and phone home
// info as generic interfaces as the user defines how much info exists in it.
type CloudInit struct {
	MetaData  map[string]interface{} `form:"meta-data" json:"meta-data"`
	UserData  map[string]interface{} `form:"user-data" json:"user-data"`
	PhoneHome PhoneHome              `form:"phone-home" json:"phone-home" binding:"omitempty"`
}

// GenerateInstanceID creates an instance-id fit for use in the instance metadata
func GenerateInstanceID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("i-%X", b)
}
