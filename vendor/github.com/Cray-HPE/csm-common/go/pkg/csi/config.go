/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

import (
	"net"
)

// BootstrapSwitchMetadata is a type that matches the switch_metadata.csv file as
// Switch Xname, Type
// The type can be CDU, Spine, Leaf, or LeafBMC
type BootstrapSwitchMetadata struct {
	Xname string `json:"xname" csv:"Switch Xname"`
	Type  string `json:"type" csv:"Type"`
}

// SystemConfig stores the overall set of system configuration parameters
type SystemConfig struct {
	SystemName      string `form:"system-name" mapstructure:"system-name"`
	SiteDomain      string `form:"site-domain" mapstructure:"site-domain"`
	Install         InstallConfig
	Cabinets        int16    `form:"cabinets" mapstructure:"cabinets"`
	StartingCabinet int16    `form:"starting-cabinet" mapstructure:"starting-cabinet"`
	StartingNID     int      `form:"starting-NID" mapstructure:"starting-NID"`
	NtpPools        []string `form:"ntp-pools" mapstructure:"ntp-pools"`
	NtpServers      []string `form:"ntp-servers" mapstructure:"ntp-servers"`
	NtpPeers        []string `form:"ntp-peers" mapstructure:"ntp-peers"`
	NtpAllow        []string `form:"ntp-allow" mapstructure:"ntp-allow"`
	NtpTimezone     string   `form:"ntp-timezone" mapstructure:"ntp-timezone"`
	IPV4Resolvers   []string `form:"ipv4-resolvers" mapstructure:"ipv4-resolvers"`
	V2Registry      string   `form:"v2-registry" mapstructure:"v2-registry"`
	RpmRegistry     string   `form:"rpm-repository" mapstructure:"rpm-repository"`
	NMNCidr         string   `form:"nmn-cidr" mapstructure:"nmn-cidr"`
	HMNCidr         string   `form:"hmn-cidr" mapstructure:"hmn-cidr"`
	CMNCidr         string   `form:"cmn-cidr" mapstructure:"cmn-cidr"`
	CANCidr         string   `form:"can-cidr" mapstructure:"can-cidr"`
	MTLCidr         string   `form:"mtl-cidr" mapstructure:"mtl-cidr"`
	HSNCidr         string   `form:"hsn-cidr" mapstructure:"hsn-cidr"`
}

// InstallConfig stores information about the site for the installer to use
type InstallConfig struct {
	NCN                 string `desc:"Hostname of the node to be used for installation"`
	NCNBondMembers      string `desc:"Comma separated list of Linux device names to set up the bond on the installation node"`
	SiteIP              net.IP `desc:"IP address for the site connection of the installer node"  valid:"ipv4 notnull"`
	SitePrefix          string `desc:"Subnet Prefix for the site connection"`
	SiteDNS             net.IP `desc:"IP address for the site dns server" valid:"ipv4"`
	SiteGW              net.IP `desc:"Gateway IP address for the site connection of the installer node" valid:"ipv4"`
	SiteNIC             string `desc:"Linux Interface Identifier for the NIC connected to the site network" flag:",required" valid:"stringlength(2|20)"`
	CephCephfsImage     string `desc:"The container image for the cephfs provisioner" valid:"url"`
	CephRBDImage        string `desc:"The container image for the ceph rbd provisioner" valid:"url"`
	ChartRepo           string `desc:"Upstream chart repo for use during the install" valid:"url"`
	DockerImageRegistry string `desc:"Upstream docker registry for use during the install" valid:"url"`
}
