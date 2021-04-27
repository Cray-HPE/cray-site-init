/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package pit

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/csi"
)

// MetaData is part of the cloud-init stucture and
// is only used for validating the required fields in the
// `CloudInit` struct below.
type MetaData struct {
	Hostname         string `yaml:"local-hostname" json:"local-hostname"`       // should be local hostname e.g. ncn-m003
	Xname            string `yaml:"xname" json:"xname"`                         // should be xname e.g. x3000c0s1b0n0
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

// NtpConfig is the options for the cloud-init ntp module.
// this is mainly the template that gets deployed to the NCNs
type NtpConfig struct {
	ConfPath    string `json:"confpath"`
	Template    string `json:"template"`
}

// NtpModule enables use of the cloud-init ntp module
type NtpModule struct {
	Enabled    bool `json:"enabled"`
	NtpClient  string `json:"ntp_client"`
	NTPPeers   []string `json:"peers"`
	NTPAllow   []string `json:"allow"`
	NTPServers []string `json:"servers"`
	NTPPools   []string `json:"pools"`
	Config     NtpConfig `json:"config"`
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

	// Not sure what consumes this.
	// Are we sure we want to reference something outside the cluster for this?
	ImageRegistry string `json:"docker-image-registry"` // dtr.dev.cray.com"

	// Commenting out several that I think we don't need
	// Domain string `json:domain`        // dnsmasq should provide this
	DNSServer string `json:"dns-server"`
	// CanGateway  string `json:can-gw`   // dnsmasq should provide this
	CanInterface string `json:"can-if"` // Do we need this?

	// Kubernetes Installation Globals
	KubernetesFirstMasterHostname string `json:"first-master-hostname"` // What's this for?
	KubernetesVIP                 string `json:"kubernetes-virtual-ip"`
	KubernetesMaxPods             string `json:"kubernetes-max-pods-per-node"`
	KubernetesPodCIDR             string `json:"kubernetes-pods-cidr"`     // "10.32.0.0/12"
	KubernetesServicesCIDR        string `json:"kubernetes-services-cidr"` // "10.16.0.0/12"
	KubernetesWeaveMTU            string `json:"kubernetes-weave-mtu"`     // 1376

	NumStorageNodes int `json:"num_storage_nodes"`
}

// Basecamp Defaults
// We should try to make these customizable by the user at some point

// k8sRunCMD has the list of scripts to run on NCN boot for
// all members of the kubernets cluster
var k8sRunCMD = []string{
	"/srv/cray/scripts/metal/install-bootloader.sh",
	"/srv/cray/scripts/metal/set-host-records.sh",
	"/srv/cray/scripts/metal/set-dhcp-to-static.sh",
	"/srv/cray/scripts/metal/set-dns-config.sh",
	"/srv/cray/scripts/metal/set-ntp-config.sh",
  "/srv/cray/scripts/metal/enable-lldp.sh",
	"/srv/cray/scripts/metal/set-bmc-bbs.sh",
	"/srv/cray/scripts/metal/set-efi-bbs.sh",
	"/srv/cray/scripts/metal/disable-cloud-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/common/kubernetes-cloudinit.sh",
}

// cephRunCMD has the list of scripts to run on NCN boot for
// the first Ceph member which is responsible for installing the others
var cephRunCMD = []string{
	"/srv/cray/scripts/metal/install-bootloader.sh",
	"/srv/cray/scripts/metal/set-host-records.sh",
	"/srv/cray/scripts/metal/set-dhcp-to-static.sh",
	"/srv/cray/scripts/metal/set-dns-config.sh",
	"/srv/cray/scripts/metal/set-ntp-config.sh",
  "/srv/cray/scripts/metal/enable-lldp.sh",
	"/srv/cray/scripts/metal/set-bmc-bbs.sh",
	"/srv/cray/scripts/metal/set-efi-bbs.sh",
	"/srv/cray/scripts/metal/disable-cloud-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/common/storage-ceph-cloudinit.sh",
}

// cephWorkerRunCMD has the list of scripts to run on NCN boot for
// the Ceph nodes that are not supposed to run the installation.
var cephWorkerRunCMD = []string{
	"/srv/cray/scripts/metal/install-bootloader.sh",
	"/srv/cray/scripts/metal/set-host-records.sh",
	"/srv/cray/scripts/metal/set-dhcp-to-static.sh",
	"/srv/cray/scripts/metal/set-dns-config.sh",
	"/srv/cray/scripts/metal/set-ntp-config.sh",
  "/srv/cray/scripts/metal/enable-lldp.sh",
	"/srv/cray/scripts/metal/set-bmc-bbs.sh",
	"/srv/cray/scripts/metal/set-efi-bbs.sh",
	"/srv/cray/scripts/metal/disable-cloud-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
}

// Make sure any "FIXME" added to this is updated in the MakeBasecampGlobals function below
var basecampGlobalString = `{
	"can-gw": "~FIXME~ e.g. 10.102.9.20",
	"can-if": "vlan007",
	"ceph-cephfs-image": "dtr.dev.cray.com/cray/cray-cephfs-provisioner:0.1.0-nautilus-1.3",
	"ceph-rbd-image": "dtr.dev.cray.com/cray/cray-rbd-provisioner:0.1.0-nautilus-1.3",
	"docker-image-registry": "dtr.dev.cray.com",
	"domain": "nmn hmn",
	"first-master-hostname": "~FIXME~ e.g. ncn-m002",
	"k8s-virtual-ip": "~FIXME~ e.g. 10.252.120.2",
	"kubernetes-max-pods-per-node": "200",
	"kubernetes-pods-cidr": "10.32.0.0/12",
	"kubernetes-services-cidr": "10.16.0.0/12",
	"kubernetes-weave-mtu": "1376",
	"rgw-virtual-ip": "~FIXME~ e.g. 10.252.2.100",
	"wipe-ceph-osds": "yes",
	"system-name": "~FIXME~",
	"site-domain": "~FIXME~",
	"internal-domain": "~FIXME~",
	"k8s-api-auditing-enabled": "~FIXME~",
  "ncn-mgmt-node-auditing-enabled": "~FIXME~"
	}`

// BasecampHostRecord is what we need for passing stuff to /etc/hosts
type BasecampHostRecord struct {
	IP      string   `json:"ip"`
	Aliases []string `json:"aliases"`
}

// MakeBasecampHostRecords uses the ncns to generate a list of host ips and their names for use in /etc/hosts
func MakeBasecampHostRecords(ncns []csi.LogicalNCN, shastaNetworks map[string]*csi.IPV4Network, installNCN string) interface{} {
	var hostrecords []BasecampHostRecord
	hmnNetwork, _ := shastaNetworks["HMN"].LookUpSubnet("bootstrap_dhcp")
	for _, ncn := range ncns {
		for _, iface := range ncn.Networks {
			var aliases []string
			aliases = append(aliases, fmt.Sprintf("%s.%s", ncn.Hostname, strings.ToLower(iface.NetworkName)))
			if iface.NetworkName == "NMN" {
				aliases = append(aliases, ncn.Hostname)
			}
			hostrecords = append(hostrecords, BasecampHostRecord{iface.IPAddress, aliases})
			if iface.NetworkName == "HMN" {
				for _, rsrv := range hmnNetwork.ReservationsByName() {
					if stringInSlice(fmt.Sprintf("%s-mgmt", ncn.Hostname), rsrv.Aliases) {
						var bmcAliases []string
						bmcAliases = append(bmcAliases, fmt.Sprintf("%s-mgmt", ncn.Hostname))
						hostrecords = append(hostrecords, BasecampHostRecord{rsrv.IPAddress.String(), bmcAliases})
					}
				}
			}
		}
	}
	nmnNetwork, _ := shastaNetworks["NMN"].LookUpSubnet("bootstrap_dhcp")
	k8sres := nmnNetwork.ReservationsByName()["kubeapi-vip"]
	hostrecords = append(hostrecords, BasecampHostRecord{k8sres.IPAddress.String(), []string{k8sres.Name, fmt.Sprintf("%s.nmn", k8sres.Name)}})

	rgwres := nmnNetwork.ReservationsByName()["rgw-vip"]
	hostrecords = append(hostrecords, BasecampHostRecord{rgwres.IPAddress.String(), []string{rgwres.Name, fmt.Sprintf("%s.nmn", rgwres.Name)}})

	// using installNCN value as the host that pit.nmn will point to
	pitres := nmnNetwork.ReservationsByName()[installNCN]
	hostrecords = append(hostrecords, BasecampHostRecord{pitres.IPAddress.String(), []string{"pit", "pit.nmn"}})

	// Add entries for the switches
	nmnNetNetwork, _ := shastaNetworks["NMN"].LookUpSubnet("network_hardware")
	for _, tmpReservation := range nmnNetNetwork.IPReservations {
		if strings.HasPrefix(tmpReservation.Name, "sw-") {
			hostrecords = append(hostrecords, BasecampHostRecord{tmpReservation.IPAddress.String(), []string{tmpReservation.Name}})
		}
	}
	return hostrecords
}

// unique de-dupes an array of string
func unique(arr []string) []string {
	occured := map[string]bool{}
	result:=[]string{}

	for e:= range arr {
		if occured[arr[e]] != true {
		occured[arr[e]] = true
			result = append(result, arr[e])
		}
	}
	return result
}

// MakeBasecampGlobals uses the defaults above to create a suitable k/v pairing for the
// Globals in data.json for basecamp
func MakeBasecampGlobals(v *viper.Viper, logicalNcns []csi.LogicalNCN, shastaNetworks map[string]*csi.IPV4Network, installNetwork string, installSubnet string, installNCN string) (map[string]interface{}, error) {
	// Create the map to return
	global := make(map[string]interface{})
	// Cheat and pull in the string as json
	err := json.Unmarshal([]byte(basecampGlobalString), &global)
	if err != nil {
		return global, err
	}

	// Update the FIXME values with our configs

	// First loop through and see if there's a viper flag
	// We register a few aliases because flags don't necessarily match data.json keys
	v.RegisterAlias("can-gw", "can-gateway")
	for key := range global {
		if v.IsSet(key) {
			global[key] = v.GetString(key)
		}
	}
	// Handle the boolean flags too
	global["k8s-api-auditing-enabled"] = v.GetBool("k8s-api-auditing-enabled")
	global["ncn-mgmt-node-auditing-enabled"] = v.GetBool("ncn-mgmt-node-auditing-enabled")

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
	// get the nmnlb subnet
	nmnlbSubnet := shastaNetworks["NMNLB"].SubnetbyName("nmn_metallb_address_pool")
	// get the unbound network from subnets above
	unbound := nmnlbSubnet.ReservationsByName()
	// include the pit and unbound in the list of dns servers
	dnsServers := unbound["unbound"].IPAddress.String() + " " + reservations[installNCN].IPAddress.String()
	// Add these to the dns-server key
	global["dns-server"] = dnsServers

	// first-master-hostname is used to ??? TODO:
	global["first-master-hostname"] = "ncn-m002"
	// "k8s-virtual-ip" is the nmn alias for k8s
	global["k8s-virtual-ip"] = reservations["kubeapi-vip"].IPAddress.String()
	global["rgw-virtual-ip"] = reservations["rgw-vip"].IPAddress.String()

	global["host_records"] = MakeBasecampHostRecords(logicalNcns, shastaNetworks, installNCN)

	// start storage count at zero
	var s = 0
	for _, ncn := range logicalNcns {
		if ncn.Subrole == "Storage" {
			// if a storage node is detected, increase the count by one
			s++
		}
	}

	global["num_storage_nodes"] = s

	return global, nil
}

// MakeBaseCampfromNCNs uses ncns and networks to create the basecamp config
func MakeBaseCampfromNCNs(v *viper.Viper, ncns []csi.LogicalNCN, shastaNetworks map[string]*csi.IPV4Network) (map[string]CloudInit, error) {
	basecampConfig := make(map[string]CloudInit)
	uaiMacvlanSubnet, err := shastaNetworks["NMN"].LookUpSubnet("uai_macvlan")
	if err != nil {
		log.Fatal("basecamp_gen: Couldn't find the macvlan subnet in the NMN")
	}
	uaiReservations := uaiMacvlanSubnet.ReservationsByName()
	for _, ncn := range ncns {
		mac0Interface := make(map[string]interface{})
		mac0Interface["ip"] = uaiReservations[ncn.Hostname].IPAddress
		mac0Interface["mask"] = uaiMacvlanSubnet.CIDR.String()
		mac0Interface["gateway"] = uaiMacvlanSubnet.Gateway

		tempAvailabilityZone, err := csi.CabinetForXname(ncn.Xname)
		if err != nil {
			log.Printf("Couldn't generate cabinet name for %v: %v \n", ncn.Xname, err)
		}
		tempMetadata := MetaData{
			Hostname:         ncn.Hostname,
			Xname:            ncn.Xname,
			InstanceID:       ncn.InstanceID,
			Region:           v.GetString("system-name"),
			AvailabilityZone: tempAvailabilityZone,
			ShastaRole:       "ncn-" + strings.ToLower(ncn.Subrole),
		}
		userDataMap := make(map[string]interface{})
		if ncn.Subrole == "Storage" {
			if strings.HasSuffix(ncn.Hostname, "001") {
				userDataMap["runcmd"] = cephRunCMD
			} else {
				userDataMap["runcmd"] = cephWorkerRunCMD
			}
		} else {
			userDataMap["runcmd"] = k8sRunCMD
		}
		userDataMap["hostname"] = ncn.Hostname
		userDataMap["local_hostname"] = ncn.Hostname
		userDataMap["mac0"] = mac0Interface
		if ncn.Bond0Mac0 == "" && ncn.Bond0Mac1 == "" {
			basecampConfig[ncn.NmnMac] = CloudInit{
				MetaData: tempMetadata,
				UserData: userDataMap,
			}
		}
		if ncn.Bond0Mac0 != "" {
			basecampConfig[ncn.Bond0Mac0] = CloudInit{
				MetaData: tempMetadata,
				UserData: userDataMap,
			}
		}
		if ncn.Bond0Mac1 != "" {
			basecampConfig[ncn.Bond0Mac1] = CloudInit{
				MetaData: tempMetadata,
				UserData: userDataMap,
			}
		}

		// ntp allowed networks should be a list of NMN and HMN CIDRS
		var nmnNets []string
		for _, netNetwork := range shastaNetworks {
			nmnNets = append(nmnNets, netNetwork.CIDR)
		}

		// merge the deprecated ntp-pool flag to the new list of pools
		poolPool := append([]string{v.GetString("ntp-pool")}, v.GetStringSlice("ntp-pools")...)

		// remove any duplicates
		pools := unique(poolPool)

		ntpConfig := NtpConfig{
			ConfPath: "/etc/chrony.d/cray.conf",
			Template: "## template: jinja\n# csm-generated config.  Do not modify--changes can be overwritten\n{% for pool in pools | sort -%}\npool {{ pool }} iburst\n{% endfor %}\n{% for server in servers | sort -%}\n{% if local_hostname == 'ncn-m001' and server == 'ncn-m001' %}\n{% endif %}\n{% if local_hostname != 'ncn-m001' and server != 'ncn-m001' %}\n{% else %}\nserver {{ server }} iburst trust\n{% endif %}\n{% endfor %}\n{% for peer in peers | sort -%}\n{% if local_hostname == peer %}\n{% else %}\n{% if loop.index <= 9 %}\n{# Only add 9 peers to prevent too much NTP traffic #}\npeer {{ peer }} minpoll -2 maxpoll 9 iburst trust\n{% endif %}\n{% endif %}\n{% endfor %}\n{% for net in allow | sort -%}\nallow {{ net }}\n{% endfor %}\nlocal stratum 3 orphan\nlog measurements statistics tracking\nlogchange 1.0\nmakestep 0.1 3\n",
		}

		ntpModule := NtpModule{
				Enabled: true,
				NtpClient: "chrony",
				NTPPeers: v.GetStringSlice("ntp-peers"),
				NTPAllow: nmnNets,
				NTPServers: v.GetStringSlice("ntp-servers"),
				NTPPools: pools,
				Config: ntpConfig,
			}

		userDataMap["ntp"] = ntpModule
	}

	return basecampConfig, nil
}

// WriteBasecampData writes basecamp data.json for the installer
func WriteBasecampData(path string, ncns []csi.LogicalNCN, shastaNetworks map[string]*csi.IPV4Network, globals interface{}) {
	v := viper.GetViper()
	basecampConfig, err := MakeBaseCampfromNCNs(v, ncns, shastaNetworks)
	if err != nil {
		log.Printf("Error extracting NCNs: %v", err)
	}
	// To write this the way we want to consume it, we need to convert it to a map of strings and interfaces
	data := make(map[string]interface{})
	for k, v := range basecampConfig {
		data[k] = v
	}
	globalMetadata := make(map[string]interface{})
	globalMetadata["meta-data"] = globals.(map[string]interface{})
	data["Global"] = globalMetadata

	err = csiFiles.WriteJSONConfig(path, data)
	if err != nil {
		log.Printf("Error writing data.json: %v", err)
	}

}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
