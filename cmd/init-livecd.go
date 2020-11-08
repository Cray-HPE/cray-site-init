/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/viper"

	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
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
			AvailabilityZone: "", // TODO: Use cabinet for AZ once that is ready
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
					MetaData: tempMetadata,
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
	csiFiles.WriteJSONConfig(path, basecampConfig)

	// https://stash.us.cray.com/projects/MTL/repos/docs-non-compute-nodes/browse/example-data.json
	/* Funky vars from the stopgap
	export site_nic=em1
	export bond_member0=p801p1
	export bond_member1=p801p2
	*/
}

// WriteConmanConfig provides conman configuration for the installer
func WriteConmanConfig(path string, ncns []*shasta.LogicalNCN, conf shasta.SystemConfig) {
	type conmanLine struct {
		Hostname string
		User     string
		IP       string
		Pass     string
	}
	v := viper.GetViper()
	ncnBMCUser := v.GetString("bootstrap-ncn-bmc-user")
	ncnBMCPass := v.GetString("bootstrap-ncn-bmc-pass")

	var conmanNCNs []conmanLine

	for _, k := range ncns {
		conmanNCNs = append(conmanNCNs, conmanLine{
			Hostname: k.Hostname,
			User:     ncnBMCUser,
			Pass:     ncnBMCPass,
			IP:       k.BMCIp,
		})
	}

	tpl6, _ := template.New("conmanconfig").Parse(string(shasta.ConmanConfigTemplate))
	csiFiles.WriteTemplate(path, tpl6, conmanNCNs)
}

// WriteMetalLBConfigMap creates the yaml configmap
func WriteMetalLBConfigMap(path string, conf shasta.SystemConfig, networks map[string]*shasta.IPV4Network) {

	// this lookup table should be redundant in the future
	// when we can better hint which pool an endpoint should pull from
	var metalLBLookupTable = map[string]string{
		"nmn_metallb_address_pool": "node-management",
		"hmn_metallb_address_pool": "hardware-management",
		"hsn_metallb_address_pool": "high-speed",
		"can_metallb_address_pool": "customer-access",
		"can_metallb_static_pool":  "customer-access-static",
	}

	tpl, err := template.New("mtllbconfigmap").Parse(string(shasta.MetalLBConfigMapTemplate))
	if err != nil {
		log.Printf("The template failed to render because: %v \n", err)
	}
	var configStruct shasta.MetalLBConfigMap
	configStruct.Networks = make(map[string]string)
	configStruct.ASN = "65533"
	configStruct.SpineSwitches = append(configStruct.SpineSwitches, "x3333")

	for _, network := range networks {
		for _, subnet := range network.Subnets {
			if strings.Contains(subnet.Name, "metallb") {
				if val, ok := metalLBLookupTable[subnet.Name]; ok {
					configStruct.Networks[val] = subnet.CIDR.String()
				}
			}
		}
	}
	csiFiles.WriteTemplate(filepath.Join(path, "metallb.yaml"), tpl, configStruct)
}

// WriteDNSMasqConfig writes the dnsmasq configuration files necssary for installation
func WriteDNSMasqConfig(path string, bootstrap []*shasta.LogicalNCN, networks map[string]*shasta.IPV4Network) {

	// DNSMasqNCN is the struct to manage NCNs within DNSMasq
	type DNSMasqNCN struct {
		Hostname string `form:"hostname"`
		NMNMac   string `form:"nmn-mac"`
		NMNIP    string `form:"nmn-ip"`
		MTLMac   string `form:"mtl-mac"`
		MTLIP    string `form:"mtl-ip"`
		BMCMac   string `form:"bmc-mac"`
		BMCIP    string `form:"bmc-ip"`
		CANIP    string `form:"can-ip"`
		HMNIP    string `form:"can-ip"`
	}

	tpl1, _ := template.New("statics").Parse(string(shasta.StaticConfigTemplate))
	tpl2, _ := template.New("canconfig").Parse(string(shasta.CANConfigTemplate))
	tpl3, _ := template.New("hmnconfig").Parse(string(shasta.HMNConfigTemplate))
	tpl4, _ := template.New("nmnconfig").Parse(string(shasta.NMNConfigTemplate))
	tpl5, _ := template.New("mtlconfig").Parse(string(shasta.MTLConfigTemplate))
	var ncns []DNSMasqNCN
	var nmnIP string
	for _, v := range bootstrap {
		for _, net := range v.Networks {
			if net.NetworkName == "NMN" {
				nmnIP = net.IPAddress
			}
		}
		// Get a new ip reservation for each one
		ncn := DNSMasqNCN{
			Hostname: v.Hostname,
			NMNMac:   v.NMNMac,
			NMNIP:    nmnIP,
			// MTLMac:   nil,
			BMCMac: v.BMCMac,
			BMCIP:  v.BMCIp,
		}
		ncns = append(ncns, ncn)
	}
	csiFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/statics.conf"), tpl1, ncns)

	// get a pointer to the MTL
	mtlNet := networks["MTL"]
	// get a pointer to the subnet
	mtlBootstrapSubnet, _ := mtlNet.LookUpSubnet("bootstrap_dhcp")
	csiFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/mtl.conf"), tpl5, mtlBootstrapSubnet)

	// Deal with the easy ones
	writeConfig("CAN", path, *tpl2, networks)
	writeConfig("HMN", path, *tpl3, networks)
	writeConfig("NMN", path, *tpl4, networks)
}

func writeConfig(name, path string, tpl template.Template, networks map[string]*shasta.IPV4Network) {
	// get a pointer to the IPV4Network
	tempNet := networks[name]
	// get a pointer to the subnet
	bootstrapSubnet, _ := tempNet.LookUpSubnet("bootstrap_dhcp")
	csiFiles.WriteTemplate(filepath.Join(path, fmt.Sprintf("dnsmasq.d/%v.conf", name)), &tpl, bootstrapSubnet)
}
