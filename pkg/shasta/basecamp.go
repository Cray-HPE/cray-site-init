/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

// MetaData is part of the cloud-init stucture and
// is only used for validating the required fields in the
// `CloudInit` struct below.
type MetaData struct {
	Hostname         string `yaml:"local-hostname" json:"local-hostname"`       // should be xname
	InstanceID       string `yaml:"instance-id" json:"instance-id"`             // should be unique for the life of the image
	Region           string `yaml:"region" json:"region"`                       // unused currently
	AvailabilityZone string `yaml:"availability-zone" json:"availability-zone"` // unused currently
	ShastaRole       string `yaml:"shasta-role" json:"shasta-role"`             // map to HSM role
}

// CloudInit is the main cloud-init struct. Leave the meta-data, user-data, and phone home
// info as generic interfaces as the user defines how much info exists in it.
type CloudInit struct {
	MetaData MetaData               `yaml:"meta-data" json:"meta-data"`
	UserData map[string]interface{} `yaml:"user-data" json:"user-data"`
}

// BaseCampGlobals is the set of information needed for an install to reach
// the handoff point.
type BaseCampGlobals struct {

	// CEPH Installation Globals
	CephFSImage          string `json:"ceph-cephfs-image"`
	CephRBDImage         string `json:"ceph-rbd-image"`
	CephStorageNodeCount string `json:"ceph-num-storage-nodes"` // "3"
	CephRGWVip           string `json:"rgw-virtual-ip"`
	CephWipe             bool   `json:"wipe-ceph-osds"`

	// Not sure what consumes these.
	// Are we sure we want to reference something outside the cluster for either of them?
	ChartRepo     string `json:"chart-repo"`            // "http://helmrepo.dev.cray.com:8080"
	ImageRegistry string `json:"docker-image-registry"` // dtr.dev.cray.com"

	// Commenting out several that I think we don't need
	// Domain string `json:domain`        // dnsmasq should provide this
	// DNSServer string `json:dns-server` // dnsmasq should provide this
	// CanGateway  string `json:can-gw`   // dnsmasq should provide this
	CanInterface string `json:"can-if"` // Do we need this?

	// Kubernetes Installation Globals
	KubernetesFirstMasterHostname string `json:"first-master-hostname"` // What's this for?
	KubernetesVIP                 string `json:"kubernetes-virtual-ip"`
	KubernetesMaxPods             string `json:"kubernetes-max-pods-per-node"`
	KubernetesPodCIDR             string `json:"kubernetes-pods-cidr"`     // "10.32.0.0/12"
	KubernetesServicesCIDR        string `json:"kubernetes-services-cidr"` // "10.16.0.0/12"
	KubernetesWeaveMTU            string `json:"kubernetes-weave-mtu"`     // 1460

	// NTP Setup Globals
	NTPPeers    string `json:"ntp-peers"`
	NTPAllow    string `json:"ntp_local_nets"`
	NTPUpstream string `json:"ntp-upstream-server"`
}
