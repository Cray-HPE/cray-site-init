/*
 * MIT License
 *
 * (C) Copyright 2025 Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package networking

import "net"

// MetaData is part of the cloud-init structure and
// is only used for validating the required fields in the
// `CloudInit` struct below.
type MetaData struct {
	Hostname         string `yaml:"local-hostname" json:"local-hostname"`       // should be local hostname e.g. ncn-m003
	Xname            string `yaml:"xname" json:"xname"`                         // should be xname e.g. x3000c0s1b0n0
	InstanceID       string `yaml:"instance-id" json:"instance-id"`             // should be unique for the life of the image
	Region           string `yaml:"region" json:"region"`                       // unused currently
	AvailabilityZone string `yaml:"availability-zone" json:"availability-zone"` // unused currently
	ShastaRole       string `yaml:"shasta-role" json:"shasta-role"`             // map to HSM role
	IPAM             IPAM   `yaml:"ipam" json:"ipam"`                           // network configs for a node
}

// CloudInit is the main cloud-init struct. Leave the meta-data, user-data, and phone home
// info as generic interfaces as the user defines how much info exists in it.
type CloudInit struct {
	MetaData *MetaData `yaml:"meta-data" json:"meta-data"`
	UserData *UserData `yaml:"user-data" json:"user-data"`
}

type UserData struct {
	BootCMD       [][]string               `yaml:"bootcmd" json:"bootcmd"`
	FSSetup       []map[string]interface{} `yaml:"fs_setup" json:"fs_setup"`
	Hostname      string                   `yaml:"hostname" json:"hostname"`
	LocalHostname string                   `yaml:"local_hostname" json:"local_hostname"`
	MAC0Interface MAC0Interface            `yaml:"mac0" json:"mac0"`
	Mounts        [][]string               `yaml:"mounts" json:"mounts"`
	RunCMD        []string                 `yaml:"runcmd" json:"runcmd"`
	NTP           NTPModule                `yaml:"ntp" json:"ntp"`
	Timezone      string                   `yaml:"timezone" json:"timezone"`
	WriteFiles    []WriteFiles             `yaml:"write_files" json:"write_files"`
}

// MAC0Interface represents a macVLAN interface
type MAC0Interface struct {
	IP      net.IP `yaml:"ip" json:"ip"`
	Mask    string `yaml:"mask" json:"mask"`
	Gateway net.IP `yaml:"gateway" json:"gateway"`
}

// NTPConfig is the options for the cloud-init ntp module.
// this is mainly the template that gets deployed to the NCNs
type NTPConfig struct {
	ConfPath string `json:"confpath"`
	Template string `json:"template"`
}

type NetworkConfig struct {
	Gateway      string `json:"gateway"`
	IP           string `json:"ip"`
	ParentDevice string `json:"parent_device"`
	VLANID       int    `json:"vlanid"`
	IP6          string `json:"ip6,omitempty"`
	Gateway6     string `json:"gateway6,omitempty"`
}

type IPAM map[string]NetworkConfig

// NTPModule enables use of the cloud-init ntp module
type NTPModule struct {
	Enabled    bool      `json:"enabled"`
	NtpClient  string    `json:"ntp_client"`
	NTPPeers   []string  `json:"peers"`
	NTPAllow   []string  `json:"allow"`
	NTPServers []string  `json:"servers"`
	NTPPools   []string  `json:"pools,omitempty"`
	Config     NTPConfig `json:"config"`
}

// WriteFiles enables use of the cloud-init write_files module
type WriteFiles struct {
	Content     string `json:"content"`
	Owner       string `json:"owner"`
	Path        string `json:"path"`
	Permissions string `json:"permissions"`
}
