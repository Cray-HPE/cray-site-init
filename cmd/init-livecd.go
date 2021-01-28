/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/viper"
	"stash.us.cray.com/MTL/csi/pkg/cpt"
	"stash.us.cray.com/MTL/csi/pkg/csi"

	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
)

// WriteBasecampData writes basecamp data.json for the installer
func WriteBasecampData(path string, ncns []csi.LogicalNCN, shastaNetworks map[string]*csi.IPV4Network, globals interface{}) {
	v := viper.GetViper()
	basecampConfig, err := cpt.MakeBaseCampfromNCNs(v, ncns, shastaNetworks)
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

	csiFiles.WriteJSONConfig(path, data)
}

// WriteConmanConfig provides conman configuration for the installer
func WriteConmanConfig(path string, ncns []csi.LogicalNCN) {
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

	tpl6, _ := template.New("conmanconfig").Parse(string(cpt.ConmanConfigTemplate))
	csiFiles.WriteTemplate(path, tpl6, conmanNCNs)
}

// WriteMetalLBConfigMap creates the yaml configmap
func WriteMetalLBConfigMap(path string, v *viper.Viper, networks map[string]*csi.IPV4Network, switches []*csi.ManagementSwitch) {

	// this lookup table should be redundant in the future
	// when we can better hint which pool an endpoint should pull from
	var metalLBLookupTable = map[string]string{
		"nmn_metallb_address_pool": "node-management",
		"hmn_metallb_address_pool": "hardware-management",
		"hsn_metallb_address_pool": "high-speed",
		// "can_metallb_address_pool": "customer-access",
		// "can_metallb_static_pool":  "customer-access-static",
	}

	tpl, err := template.New("mtllbconfigmap").Parse(string(cpt.MetalLBConfigMapTemplate))
	if err != nil {
		log.Printf("The template failed to render because: %v \n", err)
	}
	var configStruct cpt.MetalLBConfigMap
	configStruct.Networks = make(map[string]string)
	configStruct.ASN = v.GetString("bgp-asn")

	var spineSwitchXnames []string
	for _, mgmtswitch := range switches {
		if mgmtswitch.SwitchType == "Spine" {
			spineSwitchXnames = append(spineSwitchXnames, mgmtswitch.Xname)
		}
	}

	for name, network := range networks {
		for _, subnet := range network.Subnets {
			// This is a v1.4 HACK related to the supernet.
			if name == "NMN" && subnet.Name == "network_hardware" {
				for _, reservation := range subnet.IPReservations {
					for _, switchXname := range spineSwitchXnames {
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
			configStruct.Networks["customer-access-static"] = v.GetString("can-static-pool")
			configStruct.Networks["customer-access"] = v.GetString("can-dynamic-pool")
		}
	}
	csiFiles.WriteTemplate(filepath.Join(path, "metallb.yaml"), tpl, configStruct)
}

// WriteDNSMasqConfig writes the dnsmasq configuration files necssary for installation
func WriteDNSMasqConfig(path string, v *viper.Viper, bootstrap []csi.LogicalNCN, networks map[string]*csi.IPV4Network) {
	// DNSMasqNCN is the struct to manage NCNs within DNSMasq
	type DNSMasqNCN struct {
		Xname    string `form:"xname"`
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
	var ncns []DNSMasqNCN
	var mainMtlIP string
	for _, tmpNcn := range bootstrap {
		var canIP, nmnIP, hmnIP, mtlIP string
		for _, tmpNet := range tmpNcn.Networks {
			if tmpNet.NetworkName == "NMN" {
				nmnIP = tmpNet.IPAddress
			}
			if tmpNet.NetworkName == "CAN" {
				canIP = tmpNet.IPAddress
			}
			if tmpNet.NetworkName == "MTL" {
				if v.GetString("install-ncn") == tmpNcn.Hostname {
					mainMtlIP = tmpNet.IPAddress
				}
				mtlIP = tmpNet.IPAddress

			}
			if tmpNet.NetworkName == "HMN" {
				hmnIP = tmpNet.IPAddress
			}
		}
		// log.Println("Ready to build NCN list with:", v)
		ncn := DNSMasqNCN{
			Xname:    tmpNcn.Xname,
			Hostname: tmpNcn.Hostname,
			NMNMac:   tmpNcn.NmnMac,
			NMNIP:    nmnIP,
			// MTLMac:   ,
			MTLIP:  mtlIP,
			BMCMac: tmpNcn.BmcMac,
			BMCIP:  tmpNcn.BmcIP,
			CANIP:  canIP,
			HMNIP:  hmnIP,
		}
		ncns = append(ncns, ncn)
	}

	tpl1, _ := template.New("statics").Parse(string(cpt.StaticConfigTemplate))
	tpl2, _ := template.New("canconfig").Parse(string(cpt.CANConfigTemplate))
	tpl3, _ := template.New("hmnconfig").Parse(string(cpt.HMNConfigTemplate))
	tpl4, _ := template.New("nmnconfig").Parse(string(cpt.NMNConfigTemplate))
	tpl5, _ := template.New("mtlconfig").Parse(string(cpt.MTLConfigTemplate))

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

	var apigwAliases, apigwIP string
	nmnlbNet, _ := networks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	apigw := nmnlbNet.ReservationsByName()["istio-ingressgateway"]
	apigwAliases = strings.Join(apigw.Aliases, ",")
	apigwIP = apigw.IPAddress.String()

	data := struct {
		NCNS         []DNSMasqNCN
		KUBEVIP      string
		RGWVIP       string
		APIGWALIASES string
		APIGWIP      string
	}{
		ncns,
		kubevip,
		rgwvip,
		apigwAliases,
		apigwIP,
	}
	// log.Println("Ready to write data with NCNs:", ncns)
	csiFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/statics.conf"), tpl1, data)

	// Save the install NCN Metal ip for use as dns/gateways during install

	// get a pointer to the MTL
	mtlNet := networks["MTL"]
	// get a pointer to the subnet
	mtlBootstrapSubnet, _ := mtlNet.LookUpSubnet("bootstrap_dhcp")
	tmpGateway := mtlBootstrapSubnet.Gateway
	mtlBootstrapSubnet.Gateway = net.ParseIP(mainMtlIP)
	mtlBootstrapSubnet.SupernetRouter = genPinnedIP(mtlBootstrapSubnet.CIDR.IP, uint8(1))
	nmnLBSubnet, _ := networks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	mtlBootstrapSubnet.DNSServer = nmnLBSubnet.LookupReservation("unbound").IPAddress
	csiFiles.WriteTemplate(filepath.Join(path, "dnsmasq.d/MTL.conf"), tpl5, mtlBootstrapSubnet)
	mtlBootstrapSubnet.Gateway = tmpGateway

	// Deal with the easy ones
	writeConfig("CAN", path, *tpl2, networks)
	writeConfig("HMN", path, *tpl3, networks)
	writeConfig("NMN", path, *tpl4, networks)
}

func writeConfig(name, path string, tpl template.Template, networks map[string]*csi.IPV4Network) {
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
	if tempNet.Name == "CAN" {
		bootstrapSubnet.Gateway = net.ParseIP(v.GetString("can-gateway"))
	}
	// Normalize the CIDR before using it
	_, superNet, _ := net.ParseCIDR(bootstrapSubnet.CIDR.String())
	bootstrapSubnet.SupernetRouter = genPinnedIP(superNet.IP, uint8(1))
	nmnLBSubnet, _ := networks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	bootstrapSubnet.DNSServer = nmnLBSubnet.LookupReservation("unbound").IPAddress
	csiFiles.WriteTemplate(filepath.Join(path, fmt.Sprintf("dnsmasq.d/%v.conf", name)), &tpl, bootstrapSubnet)
}

func genPinnedIP(ip net.IP, pin uint8) net.IP {
	newIP := make(net.IP, 4)
	if len(ip) == 4 {
		newIP[0] = ip[0]
		newIP[1] = ip[1]
		newIP[2] = ip[2]
		newIP[3] = pin
	}
	if len(ip) == 16 {
		newIP[0] = ip[12]
		newIP[1] = ip[13]
		newIP[2] = ip[14]
		newIP[3] = pin
	}
	return newIP
}
