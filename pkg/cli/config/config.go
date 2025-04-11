/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
*/

package config

import (
	"log"
	"net"

	"github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/config/shcd"
	"github.com/Cray-HPE/cray-site-init/pkg/cli/config/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

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

// NewCommand represents the config command
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "config",
		Short:             "Interact with a Shasta config",
		Long:              `Interact with a Shasta config`,
		DisableAutoGenTag: true,
		Args:              cobra.MinimumNArgs(1), // PreRun: func(c *cobra.Command, args []string) {
	}

	c.AddCommand(
		dumpCommand(),
		initialize.NewCommand(),
		loadCommand(),
		shcd.NewCommand(),
		sls.NewCommand(),
		template.NewCommand(),
	)
	return c
}

func yamlStringSettings(v *viper.Viper) string {
	c := v.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf(
			"unable to marshal config to YAML: %v",
			err,
		)
	}
	return string(bs)
}
