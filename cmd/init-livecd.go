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

func makeBaseCampfromNCNs(v *viper.Viper, ncns []shasta.LogicalNCN) (map[string]shasta.CloudInit, error) {
	basecampConfig := make(map[string]shasta.CloudInit)
	for _, ncn := range ncns {
		tempAvailabilityZone, err := shasta.CabinetForXname(ncn.Xname)
		if err != nil {
			log.Printf("Couldn't generate cabinet name for %v: %v \n", ncn.Xname, err)
		}
		tempMetadata := shasta.MetaData{
			Hostname:         ncn.Hostname,
			InstanceID:       shasta.GenerateInstanceID(),
			Region:           v.GetString("system-name"),
			AvailabilityZone: tempAvailabilityZone,
			ShastaRole:       "ncn-" + strings.ToLower(ncn.Subrole),
		}
		userDataMap := make(map[string]interface{})
		if ncn.Subrole == "Storage" {
			if strings.HasSuffix(ncn.Hostname, "001") {
				userDataMap["runcmd"] = shasta.BasecampcephRunCMD
			} else {
				userDataMap["runcmd"] = shasta.BasecampcephWorkerRunCMD
			}
		} else {
			userDataMap["runcmd"] = shasta.Basecampk8sRunCMD
		}
		userDataMap["hostname"] = ncn.Hostname
		userDataMap["local_hostname"] = ncn.Hostname
		basecampConfig[ncn.NmnMac] = shasta.CloudInit{
			MetaData: tempMetadata,
			UserData: userDataMap,
		}
	}

	return basecampConfig, nil
}

func makeBaseCampfromSLS(sls *sls_common.SLSState, ncnMeta []shasta.LogicalNCN) (map[string]shasta.CloudInit, error) {
	basecampConfig := make(map[string]shasta.CloudInit)
	globalViper := viper.GetViper()

	var k8sRunCMD = []string{
		"/srv/cray/scripts/metal/set-dns-config.sh",
		"/srv/cray/scripts/metal/set-ntp-config.sh",
		"/srv/cray/scripts/metal/install-bootloader.sh",
		"/srv/cray/scripts/common/update_ca_certs.py",
		"/srv/cray/scripts/common/kubernetes-cloudinit.sh",
	}

	var cephRunCMD = []string{
		"/srv/cray/scripts/metal/set-dns-config.sh",
		"/srv/cray/scripts/metal/set-ntp-config.sh",
		"/srv/cray/scripts/metal/install-bootloader.sh",
		"/srv/cray/scripts/common/update_ca_certs.py",
		"/srv/cray/scripts/common/storage-ceph-cloudinit.sh",
	}

	var cephWorkerRunCMD = []string{
		"/srv/cray/scripts/metal/set-dns-config.sh",
		"/srv/cray/scripts/metal/set-ntp-config.sh",
		"/srv/cray/scripts/metal/install-bootloader.sh",
	}

	ncns, err := shasta.ExtractSLSNCNs(sls)
	if err != nil {
		return basecampConfig, err
	}
	for _, v := range ncns {
		tempAvailabilityZone, err := shasta.CabinetForXname(v.Xname)
		if err != nil {
			log.Printf("Couldn't generate cabinet name for %v: %v \n", v.Xname, err)
		}
		tempMetadata := shasta.MetaData{
			Hostname:         v.Hostname,
			InstanceID:       shasta.GenerateInstanceID(),
			Region:           globalViper.GetString("system-name"),
			AvailabilityZone: tempAvailabilityZone,
			ShastaRole:       "ncn-" + strings.ToLower(v.Subrole),
		}
		for _, value := range ncnMeta {
			if value.Xname == v.Xname {
				// log.Printf("Found %v in both lists. \n", value.Xname)
				userDataMap := make(map[string]interface{})
				if v.Subrole == "Storage" {
					if strings.HasSuffix(v.Hostname, "001") {
						userDataMap["runcmd"] = cephRunCMD
					} else {
						userDataMap["runcmd"] = cephWorkerRunCMD
					}
				} else {
					userDataMap["runcmd"] = k8sRunCMD
				}
				userDataMap["hostname"] = v.Hostname
				userDataMap["local_hostname"] = v.Hostname
				basecampConfig[value.NmnMac] = shasta.CloudInit{
					MetaData: tempMetadata,
					UserData: userDataMap,
				}
			}
		}
	}
	return basecampConfig, nil
}

// WriteBasecampData writes basecamp data.json for the installer
func WriteBasecampData(path string, ncns []shasta.LogicalNCN, globals interface{}) {
	v := viper.GetViper()
	basecampConfig, err := makeBaseCampfromNCNs(v, ncns)
	if err != nil {
		log.Printf("Error extracting NCNs: %v", err)
	}
	// To write this the way we want to consume it, we need to convert it to a map of strings and interfaces
	data := make(map[string]interface{})
	for k, v := range basecampConfig {
		data[k] = v
	}
	globalMetadata := make(map[string]interface{})
	globalMetadata["meta-data"] = globals.(map[string]string)
	data["Global"] = globalMetadata

	csiFiles.WriteJSONConfig(path, data)
	// https://stash.us.cray.com/projects/MTL/repos/docs-non-compute-nodes/browse/example-data.json
	/* Funky vars from the stopgap
	export site_nic=em1
	export bond_member0=p801p1
	export bond_member1=p801p2
	*/
}

// WriteConmanConfig provides conman configuration for the installer
func WriteConmanConfig(path string, ncns []shasta.LogicalNCN) {
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
			IP:       k.BmcIP,
		})
	}

	tpl6, _ := template.New("conmanconfig").Parse(string(shasta.ConmanConfigTemplate))
	csiFiles.WriteTemplate(path, tpl6, conmanNCNs)
}

// WriteMetalLBConfigMap creates the yaml configmap
func WriteMetalLBConfigMap(path string, v *viper.Viper, networks map[string]*shasta.IPV4Network) {

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
	configStruct.ASN = v.GetString("bgp-asn")

	for name, network := range networks {
		for _, subnet := range network.Subnets {
			if name == "NMN" && subnet.Name == "nmn_network_hardware" {
				for _, reservation := range subnet.IPReservations {
					for _, switchXname := range strings.Split(v.GetString("spine-switch-xnames"), ",") {
						if reservation.Comment == switchXname {
							configStruct.SpineSwitches = append(configStruct.SpineSwitches, reservation.IPAddress.String())
						}
					}
				}
			}
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
func WriteDNSMasqConfig(path string, bootstrap []shasta.LogicalNCN, networks map[string]*shasta.IPV4Network) {
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
	var canIP, nmnIP, hmnIP, mtlIP string
	for _, v := range bootstrap {
		for _, net := range v.Networks {
			if net.NetworkName == "NMN" {
				nmnIP = net.IPAddress
			}
			if net.NetworkName == "CAN" {
				canIP = net.IPAddress
			}
			if net.NetworkName == "MTL" {
				mtlIP = net.IPAddress
			}
			if net.NetworkName == "HMN" {
				hmnIP = net.IPAddress
			}
		}
		// log.Println("Ready to build NCN list with:", v)
		ncn := DNSMasqNCN{
			Hostname: v.Hostname,
			NMNMac:   v.NmnMac,
			NMNIP:    nmnIP,
			// MTLMac:   ,
			MTLIP:  mtlIP,
			BMCMac: v.BmcMac,
			BMCIP:  v.BmcIP,
			CANIP:  canIP,
			HMNIP:  hmnIP,
		}
		ncns = append(ncns, ncn)
	}
	var kubevip, rgwvip string
	nmnSubnet, _ := networks["NMN"].LookUpSubnet("bootstrap_dhcp")
	for _, reservation := range nmnSubnet.IPReservations {
		if reservation.Name == "kubeapi-vip" {
			kubevip = reservation.IPAddress.String()
		}
		if reservation.Name == "rgw-vip" {
			rgwvip = reservation.IPAddress.String()
		}
	}

	data := struct {
		NCNS    []DNSMasqNCN
		KUBEVIP string
		RGWVIP  string
	}{
		ncns,
		kubevip,
		rgwvip,
	}
	// log.Println("Ready to write data with NCNs:", ncns)
	csiFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/statics.conf"), tpl1, data)

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
	v := viper.GetViper()
	bootstrapSubnet, _ := tempNet.LookUpSubnet("bootstrap_dhcp")
	for _, reservation := range bootstrapSubnet.IPReservations {
		if reservation.Name == v.GetString("install-ncn") {
			bootstrapSubnet.Gateway = reservation.IPAddress
		}
	}
	csiFiles.WriteTemplate(filepath.Join(path, fmt.Sprintf("dnsmasq.d/%v.conf", name)), &tpl, bootstrapSubnet)
}
