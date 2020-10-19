/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"
	"strings"

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
