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

package pit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	"github.com/spf13/viper"
)

// MetaData is part of the cloud-init stucture and
// is only used for validating the required fields in the
// `CloudInit` struct below.
type MetaData struct {
	Hostname         string                 `yaml:"local-hostname" json:"local-hostname"`       // should be local hostname e.g. ncn-m003
	Xname            string                 `yaml:"xname" json:"xname"`                         // should be xname e.g. x3000c0s1b0n0
	InstanceID       string                 `yaml:"instance-id" json:"instance-id"`             // should be unique for the life of the image
	Region           string                 `yaml:"region" json:"region"`                       // unused currently
	AvailabilityZone string                 `yaml:"availability-zone" json:"availability-zone"` // unused currently
	ShastaRole       string                 `yaml:"shasta-role" json:"shasta-role"`             // map to HSM role
	IPAM             map[string]interface{} `yaml:"ipam" json:"ipam"`
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
	ConfPath string `json:"confpath"`
	Template string `json:"template"`
}

// NtpModule enables use of the cloud-init ntp module
type NtpModule struct {
	Enabled    bool      `json:"enabled"`
	NtpClient  string    `json:"ntp_client"`
	NTPPeers   []string  `json:"peers"`
	NTPAllow   []string  `json:"allow"`
	NTPServers []string  `json:"servers"`
	NTPPools   []string  `json:"pools,omitempty"`
	Config     NtpConfig `json:"config"`
}

// WriteFiles enables use of the cloud-init write_files module
type WriteFiles struct {
	Content     string `json:"content"`
	Owner       string `json:"owner"`
	Path        string `json:"path"`
	Permissions string `json:"permissions"`
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

	// Kubernetes Installation Globals
	KubernetesVIP          string `json:"kubernetes-virtual-ip"`
	KubernetesMaxPods      string `json:"kubernetes-max-pods-per-node"`
	KubernetesPodCIDR      string `json:"kubernetes-pods-cidr"`     // "10.32.0.0/12"
	KubernetesServicesCIDR string `json:"kubernetes-services-cidr"` // "10.16.0.0/12"
	KubernetesWeaveMTU     string `json:"kubernetes-weave-mtu"`     // 1376

	NumStorageNodes int `json:"num_storage_nodes"`
}

// Basecamp Defaults
// We should try to make these customizable by the user at some point
// k8sRunCMD has the list of scripts to run on NCN boot for
// all members of the kubernetes cluster
var k8sRunCMD = []string{
	"/srv/cray/scripts/metal/net-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/metal/install.sh",
	"/srv/cray/scripts/common/kubernetes-cloudinit.sh",
	"/srv/cray/scripts/join-spire-on-storage.sh",
	"touch /etc/cloud/cloud-init.disabled",
}

// cephRunCMD has the list of scripts to run on NCN boot for
// FIXME: MTL-1294 replace these with real usages of cloud-init when appropriate (some scripts may be necessary).
// the first Ceph member which is responsible for installing the others
var cephRunCMD = []string{
	"/srv/cray/scripts/metal/net-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/metal/install.sh",
	"/srv/cray/scripts/common/pre-load-images.sh",
	"/srv/cray/scripts/common/storage-ceph-cloudinit.sh",
	"touch /etc/cloud/cloud-init.disabled",
}

// cephWorkerRunCMD has the list of scripts to run on NCN boot for
// FIXME: MTL-1294 replace these with real usages of cloud-init when appropriate (some scripts may be necessary).
// the Ceph nodes that are not supposed to run the installation.
var cephWorkerRunCMD = []string{
	"/srv/cray/scripts/metal/net-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/metal/install.sh",
	"/srv/cray/scripts/common/pre-load-images.sh",
	"touch /etc/cloud/cloud-init.disabled",
}

// Make sure any "FIXME" added to this is updated in the MakeBasecampGlobals function below
var basecampGlobalString = `{
	"ceph-cephfs-image": "dtr.dev.cray.com/cray/cray-cephfs-provisioner:0.1.0-nautilus-1.3",
	"ceph-rbd-image": "dtr.dev.cray.com/cray/cray-rbd-provisioner:0.1.0-nautilus-1.3",
	"docker-image-registry": "dtr.dev.cray.com",
	"domain": "nmn mtl hmn",
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
	nmnLbNetwork, _ := shastaNetworks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	k8sres := nmnNetwork.ReservationsByName()["kubeapi-vip"]
	hostrecords = append(hostrecords, BasecampHostRecord{k8sres.IPAddress.String(), []string{k8sres.Name, fmt.Sprintf("%s.nmn", k8sres.Name)}})

	rgwres := nmnNetwork.ReservationsByName()["rgw-vip"]
	hostrecords = append(hostrecords, BasecampHostRecord{rgwres.IPAddress.String(), []string{rgwres.Name, fmt.Sprintf("%s.nmn", rgwres.Name)}})

	// using installNCN value as the host that pit.nmn will point to
	pitres := nmnNetwork.ReservationsByName()[installNCN]
	hostrecords = append(hostrecords, BasecampHostRecord{pitres.IPAddress.String(), []string{"pit", "pit.nmn"}})

	// adding packages.local and registry.local that point to api-gw to the host_records object
	apigwres := nmnLbNetwork.ReservationsByName()["istio-ingressgateway"]
	hostrecords = append(hostrecords, BasecampHostRecord{apigwres.IPAddress.String(), []string{"packages.local", "registry.local"}})

	// Add entries for the switches
	hmnNetNetwork, _ := shastaNetworks["HMN"].LookUpSubnet("network_hardware")
	for _, tmpReservation := range hmnNetNetwork.IPReservations {
		if strings.HasPrefix(tmpReservation.Name, "sw-") {
			hostrecords = append(hostrecords, BasecampHostRecord{tmpReservation.IPAddress.String(), []string{tmpReservation.Name}})
		}
	}
	return hostrecords
}

// unique de-dupes an array of string
func unique(arr []string) []string {
	occured := map[string]bool{}
	result := []string{}

	for e, s := range arr {
		if !occured[arr[e]] {
			occured[arr[e]] = true
			// only append if it's not an empty string
			// checks later in the code for ntp pools happen for a slice of len > 0
			// but that fails if the slice contains an empty string
			if s != "" {
				result = append(result, arr[e])
			}
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

	// First loop through and see if there's a viper flag
	// We register a few aliases because flags don't necessarily match data.json keys
	v.RegisterAlias("can-gw", "can-gateway")
	v.RegisterAlias("cmn-gw", "cmn-gateway")
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
	// get the nmnlb and hmnlb subnets
	nmnlbSubnet := shastaNetworks["NMNLB"].SubnetbyName("nmn_metallb_address_pool")
	hmnlbSubnet := shastaNetworks["HMNLB"].SubnetbyName("hmn_metallb_address_pool")
	// get the unbound network from subnets above
	unboundNMN := nmnlbSubnet.ReservationsByName()
	unboundHMN := hmnlbSubnet.ReservationsByName()
	// include the pit and unbound in the list of dns servers
	dnsServers := unboundNMN["unbound"].IPAddress.String() + " " + reservations[installNCN].IPAddress.String() + " " + unboundHMN["unbound"].IPAddress.String()
	// Add these to the dns-server key
	global["dns-server"] = dnsServers

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

	global["first-master-hostname"] = v.GetString("first-master-hostname")

	return global, nil
}

// Traverse the networks, assembling a list of NMN and HMN routes for Hill/Mountain Cabinets.
// Add a route from the MTL bootstrap network to the NMN network via bond0.nmn.
// Lastly, add the HMN/NMN k8s routes
// Format for ifroute-<interface> files
func getNCNStaticRoutes(v *viper.Viper, shastaNetworks map[string]*csi.IPV4Network) []WriteFiles {
	var nmnGateway string
	var hmnGateway string
	var ifrouteNMN bytes.Buffer
	var ifrouteHMN bytes.Buffer

	// Determine all the mountain/hill routes (one per cab, + HMN & NMN gateways)
	// N.B. The order of this range matters.  We get the nmn/hmn gateway values out of this first two.
	// They need to remain first here; otherwise the inner loop here needs to become two loops,
	// one to find & set the gateways, one to write the data out.
	for _, netName := range []string{"NMN", "HMN", "NMN_MTN", "HMN_MTN", "NMN_RVR", "HMN_RVR", "MTL"} {
		if shastaNetworks[netName] != nil {
			for _, subnet := range shastaNetworks[netName].Subnets {
				if subnet.Name == "network_hardware" {
					if netName == "NMN" {
						nmnGateway = subnet.Gateway.String()
					}
					if netName == "HMN" {
						hmnGateway = subnet.Gateway.String()
					}
					if netName == "MTL" {
						ifrouteNMN.WriteString(fmt.Sprintf("%s %s - bond0.nmn0\n", subnet.CIDR.String(), nmnGateway))
					}
				}
				if strings.HasPrefix(subnet.Name, "cabinet_") {
					if netName == "NMN" || netName == "NMN_MTN" || netName == "NMN_RVR" {
						ifrouteNMN.WriteString(fmt.Sprintf("%s %s - bond0.nmn0\n", subnet.CIDR.String(), nmnGateway))
					} else {
						ifrouteHMN.WriteString(fmt.Sprintf("%s %s - bond0.hmn0\n", subnet.CIDR.String(), hmnGateway))
					}
				}
			}
		}
	}

	// we should always have routes at this point
	if ifrouteNMN.Len() == 0 || ifrouteHMN.Len() == 0 {
		log.Panic("Error generating routes")
	}

	// add k8s routes
	ifrouteNMN.WriteString(fmt.Sprintf("%s %s - bond0.nmn0\n", shastaNetworks["NMNLB"].CIDR, nmnGateway))
	ifrouteHMN.WriteString(fmt.Sprintf("%s %s - bond0.hmn0\n", shastaNetworks["HMNLB"].CIDR, hmnGateway))

	writeFiles := []WriteFiles{
		{
			Content:     ifrouteNMN.String(),
			Owner:       "root:root",
			Path:        "/etc/sysconfig/network/ifroute-bond0.nmn0",
			Permissions: "0644",
		},
		{
			Content:     ifrouteHMN.String(),
			Owner:       "root:root",
			Path:        "/etc/sysconfig/network/ifroute-bond0.hmn0",
			Permissions: "0644",
		},
	}
	return writeFiles
}

// MakeBaseCampfromNCNs uses ncns and networks to create the basecamp config
func MakeBaseCampfromNCNs(v *viper.Viper, ncns []csi.LogicalNCN, shastaNetworks map[string]*csi.IPV4Network) (map[string]CloudInit, error) {
	basecampConfig := make(map[string]CloudInit)
	uaiMacvlanSubnet, err := shastaNetworks["NMN"].LookUpSubnet("uai_macvlan")
	if err != nil {
		log.Fatal("basecamp_gen: Couldn't find the macvlan subnet in the NMN")
	}
	uaiReservations := uaiMacvlanSubnet.ReservationsByName()
	writeFiles := getNCNStaticRoutes(v, shastaNetworks)

	for _, ncn := range ncns {
		mac0Interface := make(map[string]interface{})
		mac0Interface["ip"] = uaiReservations[ncn.Hostname].IPAddress
		mac0Interface["mask"] = uaiMacvlanSubnet.CIDR.String()
		mac0Interface["gateway"] = uaiMacvlanSubnet.Gateway
		tempAvailabilityZone, err := csi.CabinetForXname(ncn.Xname)
		if err != nil {
			log.Printf("Couldn't generate cabinet name for %v: %v \n", ncn.Xname, err)
		}
		ncnIPAM := make(map[string]interface{})
		for _, ncnNetwork := range ncn.Networks {

			// get interface configs
			for _, subnet := range shastaNetworks[ncnNetwork.NetworkName].Subnets {
				// Kea doesn't support multiple networks with vlan=0 so we need to special case CHN to not include in the ipam output
				if strings.ToLower(ncnNetwork.NetworkName) == "chn" {
					continue
				}

				// FIXME: Support multiple interfaces, nmn0-nmn1-nmn2 for each VLANID.
				ncnNICSubnet := make(map[string]interface{})
				ncnNICSubnet["gateway"] = subnet.Gateway
				ncnNICSubnet["ip"] = ncnNetwork.CIDR
				ncnNICSubnet["vlanid"] = ncnNetwork.Vlan
				if strings.ToLower(ncnNetwork.NetworkName) == "sun" {
					ncnNICSubnet["parent_device"] = "bond1"
				} else {
					ncnNICSubnet["parent_device"] = "bond0"
				}
				ncnIPAM[strings.ToLower(ncnNetwork.NetworkName)] = ncnNICSubnet
			}
		}
		tempMetadata := MetaData{
			Hostname:         ncn.Hostname,
			Xname:            ncn.Xname,
			InstanceID:       ncn.InstanceID,
			Region:           v.GetString("system-name"),
			AvailabilityZone: tempAvailabilityZone,
			ShastaRole:       "ncn-" + strings.ToLower(ncn.Subrole),
			IPAM:             ncnIPAM,
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
		for _, netNetwork := range ncn.Networks {
			// get this for ntp:
			nmnNets = append(nmnNets, netNetwork.CIDR)
		}

		// for use with the timezone cloud-init module
		userDataMap["timezone"] = v.GetString("ntp-timezone")

		// remove any duplicates
		pools := v.GetStringSlice("ntp-pools")

		ntpConfig := NtpConfig{
			ConfPath: "/etc/chrony.d/cray.conf",
			Template: "## template: jinja\n# csm-generated config for {{ local_hostname }}. Do not modify--changes can be overwritten\n{% for pool in pools | sort -%}\n{% if local_hostname == 'ncn-m001' and pool == 'ncn-m001' %}\n{% endif %}\n{% if local_hostname != 'ncn-m001' and pool != 'ncn-m001' %}\n{% else %}\npool {{ pool }} iburst\n{% endif %}\n{% endfor %}\n{% for server in servers | sort -%}\n{% if local_hostname == 'ncn-m001' and server == 'ncn-m001' %}\n# server {{ server }} will not be used as itself for a server\n{% else %}\nserver {{ server }} iburst trust\n{% endif %}\n{% if local_hostname != 'ncn-m001' and server != 'ncn-m001' %}\n# {{ local_hostname }}\n{% endif %}\n{% endfor %}\n{% for peer in peers | sort -%}\n{% if local_hostname == peer %}\n{% else %}\n{% if loop.index <= 9 %}\n{# Only add 9 peers to prevent too much NTP traffic #}\npeer {{ peer }} minpoll -2 maxpoll 9 iburst\n{% endif %}\n{% endif %}\n{% endfor %}\n{% for net in allow | sort -%}\nallow {{ net }}\n{% endfor %}\n{% if local_hostname == 'ncn-m001' %}\n# {{ local_hostname }} has a lower stratum than other NCNs since it is the primary server\nlocal stratum 8 orphan\n{% else %}\n# {{ local_hostname }} has a higher stratum so it selects ncn-m001 in the event of a tie\nlocal stratum 10 orphan\n{% endif %}\nlog measurements statistics tracking\nlogchange 1.0\nmakestep 0.1 3\n",
		}

		ntpModule := NtpModule{
			Enabled:    true,
			NtpClient:  "chrony",
			NTPPeers:   v.GetStringSlice("ntp-peers"),
			NTPAllow:   nmnNets,
			NTPServers: v.GetStringSlice("ntp-servers"),
			Config:     ntpConfig,
		}

		// Only configure pools if they are defined to avoid setting a default or an empty string in the chrony config
		if len(pools) > 0 {
			ntpModule.NTPPools = pools
		}

		userDataMap["ntp"] = ntpModule

		if writeFiles != nil {
			userDataMap["write_files"] = writeFiles
		}
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
