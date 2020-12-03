/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/spf13/viper"
)

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

// Basecamp Defaults
// We should try to make these customizable by the user at some point

// Basecampk8sRunCMD has the list of scripts to run on NCN boot for
// all members of the kubernets cluster
var Basecampk8sRunCMD = []string{
	"/srv/cray/scripts/metal/set-dns-config.sh",
	"/srv/cray/scripts/metal/set-ntp-config.sh",
	"/srv/cray/scripts/metal/install-bootloader.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/common/kubernetes-cloudinit.sh",
}

// BasecampcephRunCMD has the list of scripts to run on NCN boot for
// the first Ceph member which is responsible for installing the others
var BasecampcephRunCMD = []string{
	"/srv/cray/scripts/metal/set-dns-config.sh",
	"/srv/cray/scripts/metal/set-ntp-config.sh",
	"/srv/cray/scripts/metal/install-bootloader.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/common/storage-ceph-cloudinit.sh",
}

// BasecampcephWorkerRunCMD has the list of scripts to run on NCN boot for
// the Ceph nodes that are not supposed to run the installation.
var BasecampcephWorkerRunCMD = []string{
	"/srv/cray/scripts/metal/set-dns-config.sh",
	"/srv/cray/scripts/metal/set-ntp-config.sh",
	"/srv/cray/scripts/metal/install-bootloader.sh",
}

// Make sure any "FIXME" added to this is updated in the MakeBasecampGlobals function below
var basecampGlobalString = `{
	"can-gw": "~FIXME~ e.g. 10.102.9.20",
	"can-if": "vlan007",
	"ceph-cephfs-image": "dtr.dev.cray.com/cray/cray-cephfs-provisioner:0.1.0-nautilus-1.3",
	"ceph-rbd-image": "dtr.dev.cray.com/cray/cray-rbd-provisioner:0.1.0-nautilus-1.3",
	"chart-repo": "http://helmrepo.dev.cray.com:8080",
	"dns-server": "~FIXME~ e.g. 10.252.1.1",
	"docker-image-registry": "dtr.dev.cray.com",
	"domain": "nmn hmn",
	"first-master-hostname": "~FIXME~ e.g. ncn-m002",
	"k8s-virtual-ip": "~FIXME~ e.g. 10.252.120.2",
	"kubernetes-max-pods-per-node": "200",
	"kubernetes-pods-cidr": "10.32.0.0/12",
	"kubernetes-services-cidr": "10.16.0.0/12",
	"kubernetes-weave-mtu": "1460",
	"ntp_local_nets": "~FIXME~ e.g. 10.252.0.0/17 10.254.0.0/17",
	"ntp_peers": "~FIXME~ e.g. ncn-w001 ncn-w002 ncn-w003 ncn-s001 ncn-s002 ncn-s003 ncn-m001 ncn-m002 ncn-m003",
	"num_storage_nodes": "3",
	"rgw-virtual-ip": "~FIXME~ e.g. 10.252.2.100",
	"upstream_ntp_server": "~FIXME~",
	"wipe-ceph-osds": "yes",
	"system-name": "~FIXME~",
	"site-domain": "~FIXME~",
	"internal-domain": "~FIXME~"
	}`

// MakeBasecampGlobals uses the defaults above to create a suitable k/v pairing for the
// Globals in data.json for basecamp
func MakeBasecampGlobals(v *viper.Viper, shastaNetworks map[string]*IPV4Network, installNetwork string, installSubnet string, installNCN string) (map[string]string, error) {
	// Create the map to return
	global := make(map[string]string)
	// Cheat and pull in the string as json
	err := json.Unmarshal([]byte(basecampGlobalString), &global)
	if err != nil {
		return global, err
	}

	// Update the FIXME values with our configs

	// First loop through and see if there's a viper flag
	// We register a few aliases because flags don't necessarily match data.json keys
	v.RegisterAlias("upstream_ntp_server", "ntp-pool")
	v.RegisterAlias("can-gw", "can-gateway")
	for key := range global {
		if v.IsSet(key) {
			global[key] = v.GetString(key)
		}
	}
	// Our install takes place on the nmn.  We'll need that subnet for several values
	tempSubnet := shastaNetworks[installNetwork].SubnetbyName(installSubnet)
	if tempSubnet.Name == "" {
		log.Fatalf("Couldn't find a '%v' subnet in the %v network for generating basecamp's data.json.  Install is doomed.", installSubnet, installNetwork)
	}
	reservations := tempSubnet.ReservationsByName()
	var ncns []string
	for k := range reservations {
		if strings.HasPrefix(k, "ncn-") {
			ncns = append(ncns, k)
		}
	}
	// Now update with the ones that are part of the config.
	// dns-server should be the internal interface ip for the node running the installer
	global["dns-server"] = reservations[installNCN].IPAddress.String()
	// ntp_local_nets should be a list of NMN and HMN CIDRS
	global["ntp_local_nets"] = strings.Join([]string{shastaNetworks["NMN"].CIDR, shastaNetworks["HMN"].CIDR}, " ")
	// first-master-hostname is used to ??? TODO:
	global["first-master-hostname"] = "ncn-m002"
	// "k8s-virtual-ip" is the nmn alias for k8s
	global["k8s-virtual-ip"] = reservations["kubeapi-vip"].IPAddress.String()
	global["rgw-virtual-ip"] = reservations["rgw-vip"].IPAddress.String()
	global["ntp_peers"] = strings.Join(ncns, " ")

	return global, nil
}
