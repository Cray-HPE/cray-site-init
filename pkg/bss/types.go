//
//  MIT License
//
//  (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.

package bss

// IPAMNetworks - The networks that need IPAM.
var IPAMNetworks = [...]string{"cmn", "hmn", "mtl", "nmn"}

// KubernetesNCNRunCMD - The run-cmd for Kubernetes nodes.
var KubernetesNCNRunCMD = [...]string{
	"/srv/cray/scripts/metal/net-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/metal/install.sh",
	"/srv/cray/scripts/common/kubernetes-cloudinit.sh",
	"/srv/cray/scripts/join-spire-on-storage.sh",
	"touch /etc/cloud/cloud-init.disabled",
}

// StorageNCNRunCMD - The run-cmd for storage nodes.
var StorageNCNRunCMD = [...]string{
	"/srv/cray/scripts/metal/net-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/metal/install.sh",
	"/srv/cray/scripts/common/pre-load-images.sh",
	"touch /etc/cloud/cloud-init.disabled",
	"/srv/cray/scripts/common/ceph-enable-services.sh",
}

// ALL of these should live in the hms-bss repo once the effort to give the cloud-init data a formal structure is done.

// CloudInitIPAM - Typdef for IPAM map.
type CloudInitIPAM map[string]IPAMNetwork

// IPAMNetwork - Structure that describes the IPAM portion of cloud-init.
type IPAMNetwork struct {
	Gateway      string `json:"gateway" mapstructure:"gateway"`
	CIDR         string `json:"ip" mapstructure:"ip"`
	ParentDevice string `json:"parent_device" mapstructure:"parent_device"`
	VlanID       int16  `json:"vlanid" mapstructure:"vlanid"`
}

// WriteFile - Structure that describes the write-files portion of cloud-init.
type WriteFile struct {
	Content     string `json:"content"`
	Owner       string `json:"owner"`
	Path        string `json:"path"`
	Permissions string `json:"permissions"`
}

// HostRecord - Structure that describes an element of host_records portion of cloud-init
type HostRecord struct {
	IP      string   `json:"ip"`
	Aliases []string `json:"aliases"`
}

// HostRecords - Structure that describes the host_records portion of cloud-init
type HostRecords []HostRecord
