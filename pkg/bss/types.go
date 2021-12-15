package bss

var IPAMNetworks = [...]string{"can", "cmn", "hmn", "mtl", "nmn"}

var KubernetesNCNRunCMD = [...]string{
	"/srv/cray/scripts/metal/net-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/metal/install.sh",
	"/srv/cray/scripts/common/kubernetes-cloudinit.sh",
	"/srv/cray/scripts/join-spire-on-storage.sh",
	"touch /etc/cloud/cloud-init.disabled",
}
var StorageNCNRunCMD = [...]string{
	"/srv/cray/scripts/metal/net-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/metal/install.sh",
	"/srv/cray/scripts/common/pre-load-images.sh",
	"touch /etc/cloud/cloud-init.disabled",
	"/srv/cray/scripts/common/ceph-enable-services.sh",
}

// ALL of these should live in the hms-bss repo once the effort to give the cloud-init data a formal structure is done.

type CloudInitIPAM map[string]IPAMNetwork

type IPAMNetwork struct {
	Gateway      string `json:"gateway"`
	CIDR         string `json:"ip"`
	ParentDevice string `json:"parent_device"`
	VlanID       int16  `json:"vlanid"`
}

type WriteFile struct {
	Content     string `json:"content"`
	Owner       string `json:"owner"`
	Path        string `json:"path"`
	Permissions string `json:"permissions"`
}
