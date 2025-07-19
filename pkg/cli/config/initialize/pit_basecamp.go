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

package initialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/netip"
	"regexp"
	"strings"

	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/cli"
	cloudInitTemplates "github.com/Cray-HPE/cray-site-init/pkg/cli/config/template/cloud-init"
	"github.com/Cray-HPE/cray-site-init/pkg/csm"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

// BasecampGlobalMetaData is the set of information needed for an install to reach
// the handoff point.
type BasecampGlobalMetaData struct {
	CiliumKubeProxyReplacement string      `json:"cilium-kube-proxy-replacement"`
	CiliumOperatorReplicas     string      `json:"cilium-operator-replicas"`
	DNSServer                  string      `json:"dns-server"`
	Domain                     string      `json:"domain"`
	FirstMasterHostname        string      `json:"first-master-hostname"`
	InternalDomain             string      `json:"internal-domain"`
	K8SApiAuditingEnabled      bool        `json:"k8s-api-auditing-enabled"`
	K8SPrimaryCni              string      `json:"k8s-primary-cni"`
	K8sVirtualIP               string      `json:"k8s-virtual-ip"`
	KubernetesMaxPodsPerNode   string      `json:"kubernetes-max-pods-per-node"`
	KubernetesPodsCidr         string      `json:"kubernetes-pods-cidr"`
	KubernetesServicesCidr     string      `json:"kubernetes-services-cidr"`
	KubernetesWeaveMtu         string      `json:"kubernetes-weave-mtu"`
	NcnMgmtNodeAuditingEnabled bool        `json:"ncn-mgmt-node-auditing-enabled"`
	RGWVirtualIP               string      `json:"rgw-virtual-ip"`
	SiteDomain                 string      `json:"site-domain"`
	StorageNodeCount           int         `json:"num_storage_nodes"`
	SystemName                 string      `json:"system-name"`
	WipeCephOsds               string      `json:"wipe-ceph-osds"`
	HostRecords                interface{} `json:"host_records"`
}

// Basecamp Defaults
// See disks.go for disk layout, filesystems, and mounts

// We should try to make these customizable by the user at some point
// k8sRunCMD has the list of scripts to run on NCN boot for
// all members of the kubernetes cluster
var k8sRunCMD = []string{
	"/srv/cray/scripts/metal/net-init.sh",
	"/srv/cray/scripts/common/update_ca_certs.py",
	"/srv/cray/scripts/metal/install.sh",
	"/srv/cray/scripts/common/kubernetes-cloudinit.sh",
	"/srv/cray/scripts/common/join-spire-on-storage.sh",
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

var chronyTemplate = `## template: jinja
# csm-generated config for {{ local_hostname }}. Do not modify--changes can be overwritten
{% for pool in pools | sort -%}
{% if local_hostname == 'ncn-m001' and pool == 'ncn-m001' %}
{% endif %}
{% if local_hostname != 'ncn-m001' and pool != 'ncn-m001' %}
{% else %}
pool {{ pool }} iburst
{% endif %}
{% endfor %}
{% for server in servers | sort -%}
{% if local_hostname == 'ncn-m001' and server == 'ncn-m001' %}
# server {{ server }} will not be used as itself for a server
{% else %}
server {{ server }} iburst trust
{% endif %}
{% if local_hostname != 'ncn-m001' and server != 'ncn-m001' %}
# {{ local_hostname }}
{% endif %}
{% endfor %}
{% for peer in peers | sort -%}
{% if local_hostname == peer %}
{% else %}
{% if loop.index <= 9 %}
{# Only add 9 peers to prevent too much NTP traffic #}
peer {{ peer }} minpoll -2 maxpoll 9 iburst
{% endif %}
{% endif %}
{% endfor %}
{% for net in allow | sort -%}
allow {{ net }}
{% endfor %}
{% if local_hostname == 'ncn-m001' %}
# {{ local_hostname }} has a lower stratum than other NCNs since it is the primary server
local stratum 8 orphan
{% else %}
# {{ local_hostname }} has a higher stratum so it selects ncn-m001 in the event of a tie
local stratum 10 orphan
{% endif %}
log measurements statistics tracking
logchange 1.0
makestep 0.1 3
`

// BasecampHostRecord is what we need for passing stuff to /etc/hosts
type BasecampHostRecord struct {
	IP      string   `json:"ip"`
	Aliases []string `json:"aliases"`
}

// MakeBasecampHostRecords uses the ncns to generate a list of host ips and their names for use in /etc/hosts
func MakeBasecampHostRecords(
	ncns []LogicalNCN, shastaNetworks map[string]*networking.IPNetwork, installNCN string,
) interface{} {
	var hostrecords []BasecampHostRecord
	hmnNetwork, _ := shastaNetworks["HMN"].LookUpSubnet("bootstrap_dhcp")
	for _, ncn := range ncns {
		for _, iface := range ncn.Networks {
			var aliases []string
			aliases = append(
				aliases,
				fmt.Sprintf(
					"%s.%s",
					ncn.Hostname,
					strings.ToLower(iface.NetworkName),
				),
			)
			if iface.NetworkName == "NMN" {
				aliases = append(
					aliases,
					ncn.Hostname,
				)
			}
			hostrecords = append(
				hostrecords,
				BasecampHostRecord{
					iface.IPv4Address.String(),
					aliases,
				},
			)
			if iface.NetworkName == "HMN" {
				for _, rsrv := range hmnNetwork.ReservationsByName() {
					if cli.StringInSlice(
						fmt.Sprintf(
							"%s-mgmt",
							ncn.Hostname,
						),
						rsrv.Aliases,
					) {
						var bmcAliases []string
						bmcAliases = append(
							bmcAliases,
							fmt.Sprintf(
								"%s-mgmt",
								ncn.Hostname,
							),
						)
						hostrecords = append(
							hostrecords,
							BasecampHostRecord{
								rsrv.IPAddress.String(),
								bmcAliases,
							},
						)
					}
				}
			}
		}
	}
	nmnNetwork, _ := shastaNetworks["NMN"].LookUpSubnet("bootstrap_dhcp")
	nmnLbNetwork, _ := shastaNetworks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	k8sres := nmnNetwork.ReservationsByName()["kubeapi-vip"]
	hostrecords = append(
		hostrecords,
		BasecampHostRecord{
			k8sres.IPAddress.String(),
			[]string{
				k8sres.Name,
				fmt.Sprintf(
					"%s.nmn",
					k8sres.Name,
				),
			},
		},
	)

	rgwres := nmnNetwork.ReservationsByName()["rgw-vip"]
	hostrecords = append(
		hostrecords,
		BasecampHostRecord{
			rgwres.IPAddress.String(),
			[]string{
				rgwres.Name,
				fmt.Sprintf(
					"%s.nmn",
					rgwres.Name,
				),
			},
		},
	)

	// using installNCN value as the host that pit.nmn will point to
	pitres := nmnNetwork.ReservationsByName()[installNCN]
	hostrecords = append(
		hostrecords,
		BasecampHostRecord{
			pitres.IPAddress.String(),
			[]string{
				"pit",
				"pit.nmn",
			},
		},
	)

	// adding packages.local and registry.local that point to api-gw to the host_records object
	apigwres := nmnLbNetwork.ReservationsByName()["istio-ingressgateway"]
	hostrecords = append(
		hostrecords,
		BasecampHostRecord{
			apigwres.IPAddress.String(),
			[]string{
				"packages.local",
				"registry.local",
			},
		},
	)

	// Add entries for the switches
	hmnNetNetwork, _ := shastaNetworks["HMN"].LookUpSubnet("network_hardware")
	for _, tmpReservation := range hmnNetNetwork.IPReservations {
		if strings.HasPrefix(
			tmpReservation.Name,
			"sw-",
		) {
			hostrecords = append(
				hostrecords,
				BasecampHostRecord{
					tmpReservation.IPAddress.String(),
					[]string{tmpReservation.Name},
				},
			)
		}
	}
	return hostrecords
}

// MakeBasecampGlobalMetaData uses the defaults above to create a suitable k/v pairing for the
// Globals in data.json for basecamp
func MakeBasecampGlobalMetaData(
	v *viper.Viper,
	logicalNcns []LogicalNCN,
	shastaNetworks map[string]*networking.IPNetwork,
	installNetwork string,
	installSubnet string,
	installNCN string,
) (
	BasecampGlobalMetaData, error,
) {
	// Create the map to return
	global := BasecampGlobalMetaData{}

	// First loop through and see if there's a viper flag
	// We register a few aliases because flags don't necessarily match data.json keys
	RegisterAlias(
		"can-gw",
		"can-gateway",
	)
	RegisterAlias(
		"cmn-gw",
		"cmn-gateway4",
	)
	allSettings, _ := json.Marshal(v.AllSettings())
	_ = json.Unmarshal(
		allSettings,
		&global,
	)

	tempSubnet := shastaNetworks[installNetwork].SubnetByName(installSubnet)
	if tempSubnet.Name == "" {
		log.Fatalf(
			"Couldn't find a '%v' subnet in the %v network for generating basecamp's data.json. Install is doomed.",
			installSubnet,
			installNetwork,
		)
	}
	reservations := tempSubnet.ReservationsByName()
	// get the nmnlb and hmnlb subnets
	nmnlbSubnet := shastaNetworks["NMNLB"].SubnetByName("nmn_metallb_address_pool")
	hmnlbSubnet := shastaNetworks["HMNLB"].SubnetByName("hmn_metallb_address_pool")
	// get the unbound network from subnets above
	unboundNMN := nmnlbSubnet.ReservationsByName()
	unboundHMN := hmnlbSubnet.ReservationsByName()
	// include the pit and unbound in the list of dns servers
	dnsServers := unboundNMN["unbound"].IPAddress.String() + " " + reservations[installNCN].IPAddress.String() + " " + unboundHMN["unbound"].IPAddress.String()
	// Add these to the dns-server key
	global.DNSServer = dnsServers
	//
	// "k8s-virtual-ip" is the nmn alias for k8s
	global.K8sVirtualIP = reservations["kubeapi-vip"].IPAddress.String()
	global.RGWVirtualIP = reservations["rgw-vip"].IPAddress.String()

	// "Set k8s-primary-cni" to Cilium if CSM 1.7 or later
	currentVersion, eval := csm.CompareMajorMinor("1.7")
	if eval != -1 {
		log.Printf(
			"Detected CSM %s, setting k8s-primary-cni to Cilium",
			currentVersion,
		)
		global.K8SPrimaryCni = "cilium"
	}

	global.HostRecords = MakeBasecampHostRecords(
		logicalNcns,
		shastaNetworks,
		installNCN,
	)
	// start storage count at zero
	var s = 0
	for _, ncn := range logicalNcns {
		if ncn.Subrole == "Storage" {
			// if a storage node is detected, increase the count by one
			s++
		}
	}
	global.StorageNodeCount = s

	return global, nil
}

// Traverse the networks, assembling a list of NMN and HMN routes for Hill/Mountain Cabinets.
// Add a route from the MTL bootstrap network to the NMN network via bond0.nmn.
// Lastly, add the HMN/NMN k8s routes
// Format for ifroute-<interface> files
func getNCNStaticRoutes(
	v *viper.Viper, shastaNetworks map[string]*networking.IPNetwork,
) []networking.WriteFiles {
	var nmnGateway string
	var hmnGateway string
	var ifrouteNMN bytes.Buffer
	var ifrouteHMN bytes.Buffer

	// Determine all the mountain/hill routes (one per cab, + HMN & NMN gateways)
	// N.B. The order of this range matters. We get the nmn/hmn gateway values out of this first two.
	// They need to remain first here; otherwise the inner loop here needs to become two loops,
	// one to find & set the gateways, one to write the Data out.
	for _, netName := range []string{
		"NMN",
		"HMN",
		"NMN_MTN",
		"HMN_MTN",
		"NMN_RVR",
		"HMN_RVR",
		"MTL",
	} {
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
						ifrouteNMN.WriteString(
							fmt.Sprintf(
								"%s %s - bond0.nmn0\n",
								subnet.CIDR,
								nmnGateway,
							),
						)
					}
				}
				if strings.HasPrefix(
					subnet.Name,
					"cabinet_",
				) {
					if netName == "NMN" || netName == "NMN_MTN" || netName == "NMN_RVR" {
						ifrouteNMN.WriteString(
							fmt.Sprintf(
								"%s %s - bond0.nmn0\n",
								subnet.CIDR,
								nmnGateway,
							),
						)
					} else {
						ifrouteHMN.WriteString(
							fmt.Sprintf(
								"%s %s - bond0.hmn0\n",
								subnet.CIDR,
								hmnGateway,
							),
						)
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
	ifrouteNMN.WriteString(
		fmt.Sprintf(
			"%s %s - bond0.nmn0\n",
			shastaNetworks["NMNLB"].CIDR4,
			nmnGateway,
		),
	)
	ifrouteHMN.WriteString(
		fmt.Sprintf(
			"%s %s - bond0.hmn0\n",
			shastaNetworks["HMNLB"].CIDR4,
			hmnGateway,
		),
	)

	writeFiles := []networking.WriteFiles{
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
func MakeBaseCampfromNCNs(
	v *viper.Viper, ncns []LogicalNCN, shastaNetworks map[string]*networking.IPNetwork,
) (
	map[string]networking.CloudInit, error,
) {
	basecampConfig := make(map[string]networking.CloudInit)
	uaiMacvlanSubnet, err := shastaNetworks["NMN"].LookUpSubnet("uai_macvlan")
	if err != nil {
		log.Fatal("basecamp_gen: Couldn't find the macvlan subnet in the NMN")
	}
	uaiReservations := uaiMacvlanSubnet.ReservationsByName()
	writeFiles := getNCNStaticRoutes(
		v,
		shastaNetworks,
	)
	_, oneSixCloudInit := csm.CompareMajorMinor("1.6")
	for _, ncn := range ncns {
		mac0Interface := networking.MAC0Interface{}
		mac0Interface.IP = uaiReservations[ncn.Hostname].IPAddress
		mac0Interface.Mask = uaiMacvlanSubnet.CIDR
		mac0Interface.Gateway = uaiMacvlanSubnet.Gateway
		tempAvailabilityZone, err := CabinetForXname(ncn.Xname)
		if err != nil {
			log.Printf(
				"Couldn't generate cabinet name for %v: %v \n",
				ncn.Xname,
				err,
			)
		}
		ncnIPAM := networking.IPAM{}
		for _, ncnNetwork := range ncn.Networks {

			// The CHN is configured as a subnet of the HSN which is not done by basecamp, but by SHS and Ansible.
			if strings.ToLower(ncnNetwork.NetworkName) == "chn" {
				continue
			}

			ncnNetworkConfig := networking.NetworkConfig{}
			ncnNetworkConfig.Gateway = ncnNetwork.Gateway4.String()
			ncnNetworkConfig.IP = ncnNetwork.CIDR4.String()
			if ncnNetwork.CIDR6.IsValid() && !ncnNetwork.CIDR6.Addr().IsUnspecified() {
				ncnNetworkConfig.IP6 = ncnNetwork.CIDR6.String()
			}
			if ncnNetwork.Gateway6.IsValid() && !ncnNetwork.Gateway6.IsUnspecified() {
				ncnNetworkConfig.Gateway6 = ncnNetwork.Gateway6.String()
			}

			// Fix cloud-init and remove this shenanigan
			if strings.ToLower(ncnNetwork.NetworkName) == "mtl" {
				ncnNetworkConfig.VLANID = networking.MinVLAN
			} else {
				ncnNetworkConfig.VLANID = ncnNetwork.Vlan
			}

			ncnNetworkConfig.ParentDevice = ncnNetwork.ParentInterfaceName

			ncnIPAM[strings.ToLower(ncnNetwork.NetworkName)] = ncnNetworkConfig
		}
		metaData := networking.MetaData{
			Hostname:         ncn.Hostname,
			Xname:            ncn.Xname,
			InstanceID:       ncn.InstanceID,
			Region:           v.GetString("system-name"),
			AvailabilityZone: tempAvailabilityZone,
			ShastaRole:       "ncn-" + strings.ToLower(ncn.Subrole),
			IPAM:             ncnIPAM,
		}

		userData := networking.UserData{}

		if writeFiles != nil {
			userData.WriteFiles = writeFiles
		}
		switch ncn.Subrole {
		case "Storage":
			if oneSixCloudInit != -1 {
				// Add disk configuration to cloud-init user-data if csm >= 1.6
				// prior to csm 1.6, the disk configuration was baked into the image
				userData.BootCMD = cloudInitTemplates.CephBootCMD
				userData.FSSetup = cloudInitTemplates.CephFileSystems
				userData.Mounts = cloudInitTemplates.CephMounts
			}
			if strings.HasSuffix(
				ncn.Hostname,
				"001",
			) {
				userData.RunCMD = cephRunCMD
			} else {
				userData.RunCMD = cephWorkerRunCMD
			}
		case "Master":
			userData.RunCMD = k8sRunCMD
			if oneSixCloudInit != -1 {
				// Add disk configuration to cloud-init user-data if csm >= 1.6
				// prior to csm 1.6, the disk configuration was baked into the image
				userData.BootCMD = cloudInitTemplates.MasterBootCMD
				userData.FSSetup = cloudInitTemplates.MasterFileSystems
				userData.Mounts = cloudInitTemplates.MasterMounts
			}
		case "Worker":
			userData.RunCMD = k8sRunCMD
			if oneSixCloudInit != -1 {
				// Add disk configuration to cloud-init user-data if csm >= 1.6
				// prior to csm 1.6, the disk configuration was baked into the image
				userData.BootCMD = cloudInitTemplates.WorkerBootCMD
				userData.FSSetup = cloudInitTemplates.WorkerFileSystems
				userData.Mounts = cloudInitTemplates.WorkerMounts
			}
		}

		userData.Hostname = ncn.Hostname
		userData.LocalHostname = ncn.Hostname
		userData.MAC0Interface = mac0Interface
		if ncn.Bond0Mac0 == "" && ncn.Bond0Mac1 == "" {
			basecampConfig[ncn.NmnMac] = networking.CloudInit{
				MetaData: &metaData,
				UserData: &userData,
			}
		}
		if ncn.Bond0Mac0 != "" {
			basecampConfig[ncn.Bond0Mac0] = networking.CloudInit{
				MetaData: &metaData,
				UserData: &userData,
			}
		}
		if ncn.Bond0Mac1 != "" {
			basecampConfig[ncn.Bond0Mac1] = networking.CloudInit{
				MetaData: &metaData,
				UserData: &userData,
			}
		}

		// ntp allowed networks should be a list of NMN and HMN CIDRS
		var nmnNets []string

		// Need to exclude the BICAN toggle network and the NMNLB/HMNLB networks.
		for _, netNetwork := range shastaNetworks {
			if (strings.Contains(
				netNetwork.Name,
				"HMN",
			) ||
				strings.Contains(
					netNetwork.Name,
					"NMN",
				)) &&
				!strings.Contains(
					netNetwork.Name,
					"LB",
				) {
				nmnNets = append(
					nmnNets,
					netNetwork.CIDR4,
				)
			}
		}

		// for use with the timezone cloud-init module
		userData.Timezone = v.GetString("ntp-timezone")

		// remove any duplicates
		pools := v.GetStringSlice("ntp-pools")

		ntpConfig := networking.NTPConfig{
			ConfPath: "/etc/chrony.d/cray.conf",
			Template: chronyTemplate,
		}

		// valid hostname/domain regex
		// FIXME: Incompatible with IPv6
		re, err := regexp.Compile(`^[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?(?:\.[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?)*\.?$`)
		if err != nil {
			log.Fatal(err)
		}

		// validate ntp domains
		for _, d := range v.GetStringSlice("ntp-peers") {
			_, err := netip.ParseAddr(d)
			if !re.Match([]byte(d)) && err != nil {
				log.Fatalf(
					"invalid ntp peer: %s",
					d,
				)
			}
		}
		for _, d := range v.GetStringSlice("ntp-servers") {
			_, err := netip.ParseAddr(d)
			if !re.Match([]byte(d)) && err != nil {
				log.Fatalf(
					"invalid ntp server: %s",
					d,
				)
			}
		}
		for _, d := range v.GetStringSlice("ntp-pools") {
			_, err := netip.ParseAddr(d)
			if !re.Match([]byte(d)) && err != nil {
				log.Fatalf(
					"invalid ntp pool: %s",
					d,
				)
			}
		}

		ntpModule := networking.NTPModule{
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
		userData.NTP = ntpModule

	}

	return basecampConfig, nil
}

// WriteBasecampData writes basecamp data.json for the installer
func WriteBasecampData(
	path string, ncns []LogicalNCN, shastaNetworks map[string]*networking.IPNetwork, globalMetaData interface{},
) {
	v := viper.GetViper()
	basecampConfig, err := MakeBaseCampfromNCNs(
		v,
		ncns,
		shastaNetworks,
	)
	if err != nil {
		log.Printf(
			"Error extracting NCNs: %v",
			err,
		)
	}
	// To write this the way we want to consume it, we need to convert it to a map of strings and interfaces
	globalMetaDataJSON, err := json.Marshal(globalMetaData)
	if err != nil {
		log.Fatalf(
			"Failed to marshal global data because %v",
			err,
		)
	}
	global := bssTypes.CloudInit{
		UserData: make(map[string]interface{}),
	}
	err = json.Unmarshal(
		globalMetaDataJSON,
		&global.MetaData,
	)
	if err != nil {
		log.Fatalf(
			"Failed to unmarshal global data into BSS object because %v",
			err,
		)
	}
	data := make(bssTypes.CloudDataType)
	data["Global"] = global
	for k, v := range basecampConfig {
		data[k] = v
	}

	err = files.WriteJSONConfig(
		path,
		data,
	)
	if err != nil {
		log.Printf(
			"Error writing data.json: %v",
			err,
		)
	}

}
