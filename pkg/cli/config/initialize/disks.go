package initialize

import "fmt"

const (
	crays3cache = "CRAYS3CACHE"
	conrun      = "CONRUN"
	conlib      = "CONLIB"
	k8slet      = "K8SLET"
	cephetc     = "CEPHETC"
	cephvar     = "CEPHVAR"
	contain     = "CONTAIN"
	volumeGroup = "metalvg0"
	raidArray   = "/dev/md/AUX"
)

// master bootcmd (cloud-init user-data)
var masterBootCMD = [][]string{
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

// master fs_setup (cloud-init user-data)
var masterFileSystems = []map[string]interface{}{
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

// master mounts (cloud-init user-data)
var masterMounts = [][]string{
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

// worker bootcmd (cloud-init user-data)
var workerBootCMD = [][]string{
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

// worker fs_setup (cloud-init user-data)
var workerFileSystems = []map[string]interface{}{
	{
		"label":      crays3cache,
		"filesystem": "ext4",
		"device":     fmt.Sprintf("/dev/disk/by-id/dm-name-%s-%s", volumeGroup, crays3cache),
		"partition":  "auto",
		"overwrite":  true,
	},
}

// worker mounts (cloud-init user-data)
var workerMounts = [][]string{
	{
		fmt.Sprintf("LABEL=%s", crays3cache),
		"/var/lib/s3fs_cache",
		"ext4",
		"defaults,nofail",
	},
}

// storage bootcmd (cloud-init user-data)
var cephBootCMD = [][]string{
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

// storage fs_setup (cloud-init user-data)
var cephFileSystems = []map[string]interface{}{
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

// storage mounts (cloud-init user-data)
var cephMounts = [][]string{
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
