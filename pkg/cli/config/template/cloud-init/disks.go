/*
 MIT License

 (C) Copyright 2024 Hewlett Packard Enterprise Development LP

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

package cloudinit

import "fmt"

// Provides configuration for lvm, filesystems and mounts for NCNS.
const (
	crays3cache = "CRAYS3CACHE"
	conrun      = "CONRUN"
	conlib      = "CONLIB"
	k8slet      = "K8SLET"
	cephetc     = "CEPHETC"
	cephvar     = "CEPHVAR"
	contain     = "CONTAIN"
	slingshot   = "SLINGSHOT"
	scfirmware = "SCFIRMWARE"
	volumeGroup = "metalvg0"
	raidArray   = "/dev/md/AUX"
)

// MasterBootCMD (cloud-init user-data)
var MasterBootCMD = [][]string{
	{
		"cloud-init-per",
		"once",
		"create_PV",
		"pvcreate",
		"-ff",
		"-y",
		"-M",
		"lvm2",
		raidArray,
	},
	{
		"cloud-init-per",
		"once",
		"create_VG",
		"vgcreate",
		volumeGroup,
		raidArray,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", crays3cache),
		"lvcreate",
		"-l",
		"25%PVS",
		"-n",
		crays3cache,
		"-y",
		volumeGroup,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", conrun),
		"lvcreate",
		"-l",
		"4%PVS",
		"-n",
		conrun,
		"-y",
		volumeGroup,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", conlib),
		"lvcreate",
		"-l",
		"36%PVS",
		"-n",
		conlib,
		"-y",
		volumeGroup,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", k8slet),
		"lvcreate",
		"-l",
		"10%PVS",
		"-n",
		k8slet,
		"-y",
		volumeGroup,
	},
}

// MasterFileSystems (cloud-init user-data)
var MasterFileSystems = []map[string]interface{}{
	{
		"label":      crays3cache,
		"filesystem": "ext4",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, crays3cache),
		"partition":  "auto",
		"overwrite":  true,
	},
	{
		"label":      conrun,
		"filesystem": "xfs",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, conrun),
		"partition":  "auto",
		"overwrite":  true,
	},
	{
		"label":      conlib,
		"filesystem": "xfs",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, conlib),
		"partition":  "auto",
		"overwrite":  true,
	},
	{
		"label":      k8slet,
		"filesystem": "xfs",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, k8slet),
		"partition":  "auto",
		"overwrite":  true,
	},
}

// MasterMounts (cloud-init user-data)
var MasterMounts = [][]string{
	{
		fmt.Sprintf("LABEL=%s", crays3cache),
		"/var/lib/s3fs_cache",
		"ext4",
		"defaults,nofail",
	},
	{
		fmt.Sprintf("LABEL=%s", conrun),
		"/run/containerd",
		"xfs",
		"defaults,nofail",
	},
	{
		fmt.Sprintf("LABEL=%s", conlib),
		"/var/lib/containerd",
		"xfs",
		"defaults,nofail",
	},
	{
		fmt.Sprintf("LABEL=%s", k8slet),
		"/var/lib/kubelet",
		"xfs",
		"defaults,nofail",
	},
}

// WorkerBootCMD (cloud-init user-data)
var WorkerBootCMD = [][]string{
	{
		"cloud-init-per",
		"once",
		"create_PV",
		"pvcreate",
		"-ff",
		"-y",
		"-M",
		"lvm2",
		raidArray,
	},
	{
		"cloud-init-per",
		"once",
		"create_VG",
		"vgcreate",
		volumeGroup,
		raidArray,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", crays3cache),
		"lvcreate",
		"-L",
		"200GB",
		"-n",
		crays3cache,
		"-y",
		volumeGroup,
	},
}

// WorkerFileSystems (cloud-init user-data)
var WorkerFileSystems = []map[string]interface{}{
	{
		"label":      crays3cache,
		"filesystem": "ext4",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, crays3cache),
		"partition":  "auto",
		"overwrite":  true,
	},
}

// WorkerMounts (cloud-init user-data)
var WorkerMounts = [][]string{
	{
		fmt.Sprintf("LABEL=%s", crays3cache),
		"/var/lib/s3fs_cache",
		"ext4",
		"defaults,nofail",
	},
}

// CephBootCMD (cloud-init user-data)
var CephBootCMD = [][]string{
	{
		"cloud-init-per",
		"once",
		"create_PV",
		"pvcreate",
		"-ff",
		"-y",
		"-M",
		"lvm2",
		raidArray,
	},
	{
		"cloud-init-per",
		"once",
		"create_VG",
		"vgcreate",
		volumeGroup,
		raidArray,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", cephetc),
		"lvcreate",
		"-L",
		"10GB",
		"-n",
		cephetc,
		"-y",
		volumeGroup,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", cephvar),
		"lvcreate",
		"-L",
		"60GB",
		"-n",
		cephvar,
		"-y",
		volumeGroup,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", contain),
		"lvcreate",
		"-L",
		"60GB",
		"-n",
		contain,
		"-y",
		volumeGroup,
	},
}

// CephFileSystems (cloud-init user-data)
var CephFileSystems = []map[string]interface{}{
	{
		"label":      cephetc,
		"filesystem": "ext4",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, cephetc),
		"partition":  "auto",
		"overwrite":  true,
	},
	{
		"label":      cephvar,
		"filesystem": "ext4",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, cephvar),
		"partition":  "auto",
		"overwrite":  true,
	},
	{
		"label":      contain,
		"filesystem": "xfs",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, contain),
		"partition":  "auto",
		"overwrite":  true,
	},
}

// CephMounts (cloud-init user-data)
var CephMounts = [][]string{
	{
		fmt.Sprintf("LABEL=%s", cephetc),
		"/etc/ceph",
		"auto",
		"defaults",
	},
	{
		fmt.Sprintf("LABEL=%s", cephvar),
		"/var/lib/ceph",
		"auto",
		"defaults",
	},
	{
		fmt.Sprintf("LABEL=%s", contain),
		"/var/lib/containers",
		"auto",
		"defaults",
	},
}

// FabricManagerBootCMD (cloud-init user-data)
var FabricManagerBootCMD = [][]string{
	{
		"cloud-init-per",
		"once",
		"create_PV",
		"pvcreate",
		"-ff",
		"-y",
		"-M",
		"lvm2",
		raidArray,
	},
	{
		"cloud-init-per",
		"once",
		"create_VG",
		"vgcreate",
		volumeGroup,
		raidArray,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", scfirmware),
		"lvcreate",
		"-L",
		"80GB",
		"-n",
		scfirmware,
		"-y",
		volumeGroup,
	},
	{
		"cloud-init-per",
		"once",
		fmt.Sprintf("create_LV_%s", slingshot),
		"lvcreate",
		"-L",
		"120GB",
		"-n",
		slingshot,
		"-y",
		volumeGroup,
	},
}

// FabricManagerFileSystems (cloud-init user-data)
var FabricManagerFileSystems = []map[string]interface{}{
	{
		"label":      scfirmware,
		"filesystem": "ext4",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, scfirmware),
		"partition":  "auto",
		"overwrite":  false,
	},
	{
		"label":      slingshot,
		"filesystem": "ext4",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, slingshot),
		"partition":  "auto",
		"overwrite":  false,
	},
}

// FabricManagerMounts (cloud-init user-data)
var FabricManagerMounts = [][]string{
	{
		fmt.Sprintf("LABEL=%s", scfirmware),
		"/opt/cray/FW/sc-firmware",
		"ext4",
		"defaults,nofail",
	},
	{
		fmt.Sprintf("LABEL=%s", slingshot),
		"/opt/slingshot",
		"ext4",
		"defaults,nofail",
	},
}
