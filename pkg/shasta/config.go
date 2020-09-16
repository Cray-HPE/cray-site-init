/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import "net"

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
	ShastaRole       string `form:"shasta-role" json:"shasta-role"`             //map to HSM role
}

// CloudInit is the main cloud-init struct. Leave the meta-data, user-data, and phone home
// info as generic interfaces as the user defines how much info exists in it.
type CloudInit struct {
	MetaData  map[string]interface{} `form:"meta-data" json:"meta-data"`
	UserData  map[string]interface{} `form:"user-data" json:"user-data"`
	PhoneHome PhoneHome              `form:"phone-home" json:"phone-home" binding:"omitempty"`
}
