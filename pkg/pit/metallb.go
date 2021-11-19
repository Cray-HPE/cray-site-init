/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package pit

import (
	"log"
	"path/filepath"
	"strings"
	"text/template"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	"github.com/spf13/viper"
)

// MetalLBConfigMapTemplate manages the ConfigMap for MetalLB
var MetalLBConfigMapTemplate = []byte(`
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    peers:{{range .PeerSwitches}}
    - peer-address: {{ .IPAddress }}
      peer-asn: {{ .PeerASN }}
      my-asn: {{ .MyASN }}
      {{- end}}
    address-pools:{{range .Networks}}
    - name: {{ .Name}}
      protocol: {{ .Protocol }}
      addresses: {{range $subnet := .Addresses}}
      - {{ $subnet }}
      {{- end}}
    {{- end}}
`)

// PeerDetail holds information about each of the BGP routers that we peer with in MetalLB
type PeerDetail struct {
	IPAddress string `yaml:"peer-address" valid:"_,required"`
	PeerASN   int    `yaml:"peer-asn" valid:"_,required"`
	MyASN     int    `yaml:"my-asn" valid:"_,required"`
}

// AddressPoolDetail holds information about each of the MetalLB address pools
type AddressPoolDetail struct {
	Name      string   `yaml:"name" valid:"_,required"`
	Protocol  string   `yaml:"protocol" valid:"_,required"`
	Addresses []string `yaml:"addresses" valid:"required"`
}

// MetalLBConfigMap holds information needed by the MetalLBConfigMapTemplate
type MetalLBConfigMap struct {
	AggSwitches   []PeerDetail
	PeerSwitches  []PeerDetail
	SpineSwitches []PeerDetail
	Networks      []AddressPoolDetail
}

// GetMetalLBConfig gathers the information for the metallb config map
func GetMetalLBConfig(v *viper.Viper, networks map[string]*csi.IPV4Network, switches []*csi.ManagementSwitch) MetalLBConfigMap {

	var metalLBLookupTable = map[string]string{
		"nmn_metallb_address_pool": "node-management",
		"hmn_metallb_address_pool": "hardware-management",
		"hsn_metallb_address_pool": "high-speed",
	}

	var configStruct MetalLBConfigMap

	var spineSwitchXnames, aggSwitchXnames []string
	var bgpPeers = v.GetString("bgp-peers")

	// Split out switches into spine and aggregation lists
	for _, mgmtswitch := range switches {
		if mgmtswitch.SwitchType == "Spine" {
			spineSwitchXnames = append(spineSwitchXnames, mgmtswitch.Xname)
		}
		if mgmtswitch.SwitchType == "Aggregation" {
			aggSwitchXnames = append(aggSwitchXnames, mgmtswitch.Xname)
		}
	}

	for name, network := range networks {
		for _, subnet := range network.Subnets {
			// This is a v1.4 HACK related to the supernet.
			if name == "NMN" && subnet.Name == "network_hardware" {
				var tmpPeer PeerDetail
				for _, reservation := range subnet.IPReservations {
					tmpPeer = PeerDetail{}
					tmpPeer.PeerASN = v.GetInt("bgp-asn")
					tmpPeer.MyASN = v.GetInt("bgp-asn")
					tmpPeer.IPAddress = reservation.IPAddress.String()
					for _, switchXname := range spineSwitchXnames {
						if reservation.Comment == switchXname {
							configStruct.SpineSwitches = append(configStruct.SpineSwitches, tmpPeer)
						}
					}
					for _, switchXname := range aggSwitchXnames {
						if reservation.Comment == switchXname {
							configStruct.AggSwitches = append(configStruct.AggSwitches, tmpPeer)
						}
					}
				}
			}
			if name == "CMN" && subnet.Name == "bootstrap_dhcp" {
				var tmpPeer PeerDetail
				for _, reservation := range subnet.IPReservations {
					if strings.Contains(reservation.Name, "cmn-switch") {
						tmpPeer = PeerDetail{}
						tmpPeer.PeerASN = v.GetInt("bgp-asn")
						tmpPeer.MyASN = v.GetInt("bgp-cmn-asn")
						tmpPeer.IPAddress = reservation.IPAddress.String()
						if bgpPeers == "spine" {
							configStruct.SpineSwitches = append(configStruct.SpineSwitches, tmpPeer)
						} else if bgpPeers == "aggregation" {
							configStruct.AggSwitches = append(configStruct.AggSwitches, tmpPeer)
						} else {
							log.Fatalf("bgp-peers: unrecognized option: %s\n", bgpPeers)
						}
					}
				}
			}
			if strings.Contains(subnet.Name, "metallb") {
				if val, ok := metalLBLookupTable[subnet.Name]; ok {
					var tmpAddPool AddressPoolDetail
					tmpAddPool = AddressPoolDetail{}
					tmpAddPool.Name = val
					tmpAddPool.Protocol = "bgp"
					tmpAddPool.Addresses = append(tmpAddPool.Addresses, subnet.CIDR.String())
					configStruct.Networks = append(configStruct.Networks, tmpAddPool)
				}
			}
		}
	}

	configStruct.PeerSwitches = getMetalLBPeerSwitches(bgpPeers, configStruct)

	var tmpAddPool AddressPoolDetail

	// CMN - static
	tmpAddPool = AddressPoolDetail{}
	tmpAddPool.Name = "customer-management-static"
	tmpAddPool.Protocol = "bgp"
	tmpAddPool.Addresses = append(tmpAddPool.Addresses, v.GetString("cmn-static-pool"))
	configStruct.Networks = append(configStruct.Networks, tmpAddPool)

	// CMN - dynamic
	tmpAddPool = AddressPoolDetail{}
	tmpAddPool.Name = "customer-management"
	tmpAddPool.Protocol = "bgp"
	tmpAddPool.Addresses = append(tmpAddPool.Addresses, v.GetString("cmn-dynamic-pool"))
	configStruct.Networks = append(configStruct.Networks, tmpAddPool)

	// CAN - static
	if v.GetString("can-static-pool") != "" {
		tmpAddPool = AddressPoolDetail{}
		tmpAddPool.Name = "customer-access-static"
		tmpAddPool.Protocol = "bgp"
		tmpAddPool.Addresses = append(tmpAddPool.Addresses, v.GetString("can-static-pool"))
		configStruct.Networks = append(configStruct.Networks, tmpAddPool)
	}

	// CAN - dynamic
	if v.GetString("can-dynamic-pool") != "" {
		tmpAddPool = AddressPoolDetail{}
		tmpAddPool.Name = "customer-access"
		tmpAddPool.Protocol = "bgp"
		tmpAddPool.Addresses = append(tmpAddPool.Addresses, v.GetString("can-dynamic-pool"))
		configStruct.Networks = append(configStruct.Networks, tmpAddPool)
	}

	// CHN - static
	if v.GetString("chn-static-pool") != "" {
		tmpAddPool = AddressPoolDetail{}
		tmpAddPool.Name = "customer-high-speed-static"
		tmpAddPool.Protocol = "bgp"
		tmpAddPool.Addresses = append(tmpAddPool.Addresses, v.GetString("chn-static-pool"))
		configStruct.Networks = append(configStruct.Networks, tmpAddPool)
	}

	// CHN - dynamic
	if v.GetString("chn-dynamic-pool") != "" {
		tmpAddPool = AddressPoolDetail{}
		tmpAddPool.Name = "customer-high-speed"
		tmpAddPool.Protocol = "bgp"
		tmpAddPool.Addresses = append(tmpAddPool.Addresses, v.GetString("chn-dynamic-pool"))
		configStruct.Networks = append(configStruct.Networks, tmpAddPool)
	}

	return configStruct
}

// WriteMetalLBConfigMap creates the yaml configmap
func WriteMetalLBConfigMap(path string, v *viper.Viper, networks map[string]*csi.IPV4Network, switches []*csi.ManagementSwitch) {

	tpl, err := template.New("mtllbconfigmap").Parse(string(MetalLBConfigMapTemplate))
	if err != nil {
		log.Printf("The template failed to render because: %v \n", err)
	}
	var configStruct MetalLBConfigMap

	configStruct = GetMetalLBConfig(v, networks, switches)

	csiFiles.WriteTemplate(filepath.Join(path, "metallb.yaml"), tpl, configStruct)
}

// getMetalLBPeerSwitches returns a list of switches  that should be used as metallb peers
func getMetalLBPeerSwitches(bgpPeers string, configStruct MetalLBConfigMap) []PeerDetail {

	switchTypeMap := map[string][]PeerDetail{
		"spine":       configStruct.SpineSwitches,
		"aggregation": configStruct.AggSwitches,
	}

	if peerSwitches, ok := switchTypeMap[bgpPeers]; ok {
		if len(peerSwitches) == 0 {
			log.Fatalf("bgp-peers: %s specified but none defined in switch_metadata.csv\n", bgpPeers)
		}
		for _, switchDetail := range peerSwitches {
			configStruct.PeerSwitches = append(configStruct.PeerSwitches, switchDetail)
		}
	} else {
		log.Fatalf("bgp-peers: unrecognized option: %s\n", bgpPeers)
	}

	return configStruct.PeerSwitches
}
