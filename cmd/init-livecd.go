/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fatih/structs"
	"github.com/spf13/viper"

	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	sicFiles "stash.us.cray.com/MTL/sic/internal/files"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// WriteNICConfigENV sets environment variables for nic bonding and configuration
func WriteNICConfigENV(path string, conf shasta.SystemConfig) {
	log.Printf("NOT IMPLEMENTED")
}

func makeBaseCampfromSLS(conf shasta.SystemConfig, sls *sls_common.SLSState, ncnMeta []*shasta.BootstrapNCNMetadata) (map[string]shasta.CloudInit, error) {
	basecampConfig := make(map[string]shasta.CloudInit)
	globalViper := viper.GetViper()

	var k8sRunCMD = []string{
		"/srv/cray/scripts/metal/set-dns-config.sh",
		"/srv/cray/scripts/metal/set-ntp-config.sh",
		"/srv/cray/scripts/common/kubernetes-cloudinit.sh",
	}

	var cephRunCMD = []string{
		"/srv/cray/scripts/metal/set-dns-config.sh",
		"/srv/cray/scripts/metal/set-ntp-config.sh",
		"/srv/cray/scripts/common/storage-ceph-cloudinit.sh",
	}

	ncns, err := shasta.ExtractSLSNCNs(sls)
	if err != nil {
		return basecampConfig, err
	}
	log.Printf("Processing %d ncns from csv\n", len(ncnMeta))
	log.Printf("Processing %d ncns from sls\n", len(ncns))
	for _, v := range ncns {
		log.Printf("The aliases for %v are %v \n", v.BmcMac, v.Hostnames)

		tempMetadata := shasta.MetaData{
			Hostname:         v.Hostnames[0],
			InstanceID:       shasta.GenerateInstanceID(),
			Region:           globalViper.GetString("system-name"),
			AvailabilityZone: "", // Using cabinet for AZ
			ShastaRole:       "ncn-" + strings.ToLower(v.Subrole),
		}
		for _, value := range ncnMeta {
			if value.Xname == v.Xname {
				// log.Printf("Found %v in both lists. \n", value.Xname)
				userDataMap := make(map[string]interface{})
				if v.Subrole == "Storage" {
					// TODO: the first ceph node needs to run ceph init.  Not the others
					userDataMap["runcmd"] = cephRunCMD
				} else {
					userDataMap["runcmd"] = k8sRunCMD
				}
				userDataMap["hostname"] = v.Hostnames[0]
				userDataMap["local_hostname"] = v.Hostnames[0]
				basecampConfig[value.NmnMac] = shasta.CloudInit{
					MetaData: structs.Map(tempMetadata),
					UserData: userDataMap,
				}
			}
		}
	}
	return basecampConfig, nil
}

// WriteBaseCampData writes basecamp data.json for the installer
func WriteBaseCampData(path string, conf shasta.SystemConfig, sls *sls_common.SLSState, ncnMeta []*shasta.BootstrapNCNMetadata) {
	basecampConfig, err := makeBaseCampfromSLS(conf, sls, ncnMeta)
	if err != nil {
		log.Printf("Error extracting NCNs: %v", err)
	}
	sicFiles.WriteJSONConfig(path, basecampConfig)

	// https://stash.us.cray.com/projects/MTL/repos/docs-non-compute-nodes/browse/example-data.json
	/* Funky vars from the stopgap
	export site_nic=em1
	export bond_member0=p801p1
	export bond_member1=p801p2
	export mtl_cidr=10.1.1.1/16
	export mtl_dhcp_start=10.1.2.3
	export mtl_dhcp_end=10.1.2.254
	export nmn_cidr=10.252.0.4/17
	export nmn_dhcp_start=10.252.50.0
	export nmn_dhcp_end=10.252.99.252
	export hmn_cidr=10.254.0.4/17
	export hmn_dhcp_start=10.254.50.5
	export hmn_dhcp_end=10.254.99.252
	export site_cidr=172.30.52.220/20
	export site_gw=172.30.48.1
	export site_dns='172.30.84.40 172.31.84.40'
	export can_cidr=10.102.4.110/24
	export can_dhcp_start=10.102.4.5
	export can_dhcp_end=10.102.4.109
	export dhcp_ttl=2m
	*/
}

// WriteConmanConfig provides conman configuration for the installer
func WriteConmanConfig(path string, conf shasta.SystemConfig) {
	log.Printf("NOT IMPLEMENTED")
}

// WriteDNSMasqConfig writes the dnsmasq configuration files necssary for installation
func WriteDNSMasqConfig(path string, bootstrap []*shasta.BootstrapNCNMetadata, networks map[string]shasta.IPV4Network) {
	tpl1, _ := template.New("statics").Parse(string(shasta.StaticConfigTemplate))
	tpl2, _ := template.New("canconfig").Parse(string(shasta.CANConfigTemplate))
	tpl3, _ := template.New("hmnconfig").Parse(string(shasta.HMNConfigTemplate))
	tpl4, _ := template.New("nmnconfig").Parse(string(shasta.NMNConfigTemplate))
	tpl5, _ := template.New("mtlconfig").Parse(string(shasta.MTLConfigTemplate))
	log.Printf("Loaded Templates \n")

	var ncns []shasta.DNSMasqNCN
	log.Printf("Parsing bootstrap for dnsmasq: %v \n", bootstrap)
	for _, v := range bootstrap {
		// Get a new ip reservation for each one
		ncn := shasta.DNSMasqNCN{
			Hostname: v.GetHostname(),
			NMNMac:   v.NmnMac,
			// NMNIP:    nil,
			// MTLMac:   nil,
			BMCMac: v.BmcMac,
			// BMCIP:    nil,
		}
		log.Printf("Adding ncn for dnsmasq: %v", ncn)
		ncns = append(ncns, ncn)
	}
	sicFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/statics.conf"), tpl1, ncns)
	// get a pointer to the CAN
	canNet := networks["can"]
	// get a pointer to the subnet
	canBootstrapSubnet, _ := canNet.LookUpSubnet("bootstrap_dhcp")
	log.Printf("Calling WriteTemplate with canBootstrapSubnet = %v", &canBootstrapSubnet)
	sicFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/can.conf"), tpl2, canBootstrapSubnet)
	// get a pointer to the HMN
	hmnNet := networks["hmn"]
	// get a pointer to the subnet
	hmnBootstrapSubnet, _ := hmnNet.LookUpSubnet("bootstrap_dhcp")
	sicFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/hmn.conf"), tpl3, hmnBootstrapSubnet)
	// get a pointer to the NMN
	nmnNet := networks["nmn"]
	// get a pointer to the subnet
	nmnBootstrapSubnet, _ := nmnNet.LookUpSubnet("bootstrap_dhcp")
	sicFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/nmn.conf"), tpl4, nmnBootstrapSubnet)
	// get a pointer to the MTL
	mtlNet := networks["mtl"]
	// get a pointer to the subnet
	mtlBootstrapSubnet, _ := mtlNet.LookUpSubnet("mtl_subnet")
	sicFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/mtl.conf"), tpl5, mtlBootstrapSubnet)

}
