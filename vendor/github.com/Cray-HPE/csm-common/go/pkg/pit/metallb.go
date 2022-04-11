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
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	csiFiles "github.com/Cray-HPE/csm-common/go/internal/files"
	"github.com/Cray-HPE/csm-common/go/pkg/csi"
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
	LeafSwitches  []PeerDetail
	PeerSwitches  []PeerDetail
	SpineSwitches []PeerDetail
	EdgeSwitches  []PeerDetail
	Networks      []AddressPoolDetail
}

// GetMetalLBConfig gathers the information for the metallb config map
func GetMetalLBConfig(v *viper.Viper, networks map[string]*csi.IPV4Network, switches []*csi.ManagementSwitch) MetalLBConfigMap {

	var configStruct MetalLBConfigMap

	var bgpPeers = v.GetStringSlice("bgp-peer-types")

	spineSwitchNameRegexp := regexp.MustCompile(`sw-spine-\d{3}`)
	leafSwitchNameRegexp := regexp.MustCompile(`sw-leaf-\d{3}`)
	edgeSwitchNameRegexp := regexp.MustCompile(`chn-switch-\d`)

	for name, network := range networks {
		for _, subnet := range network.Subnets {
			// This is a v1.4 HACK related to the supernet.
			if (name == "NMN" || name == "CMN") && subnet.Name == "network_hardware" {
				var tmpPeer PeerDetail
				for _, reservation := range subnet.IPReservations {
					tmpPeer = PeerDetail{}
					tmpPeer.PeerASN = network.PeerASN
					tmpPeer.MyASN = network.MyASN
					tmpPeer.IPAddress = reservation.IPAddress.String()
					if spineSwitchNameRegexp.FindString(reservation.Name) != "" {
						configStruct.SpineSwitches = append(configStruct.SpineSwitches, tmpPeer)
					}
					if leafSwitchNameRegexp.FindString(reservation.Name) != "" {
						configStruct.LeafSwitches = append(configStruct.LeafSwitches, tmpPeer)
					}
				}
			} else if name == "CHN" && subnet.Name == "bootstrap_dhcp" {
				var tmpPeer PeerDetail
				for _, reservation := range subnet.IPReservations {
					tmpPeer = PeerDetail{}
					tmpPeer.PeerASN = network.PeerASN
					tmpPeer.MyASN = network.MyASN
					tmpPeer.IPAddress = reservation.IPAddress.String()
					if edgeSwitchNameRegexp.FindString(reservation.Name) != "" {
						configStruct.EdgeSwitches = append(configStruct.EdgeSwitches, tmpPeer)
					}
				}
			}
			if strings.Contains(subnet.Name, "metallb") {
				tmpAddPool := AddressPoolDetail{}
				tmpAddPool.Name = subnet.MetalLBPoolName
				tmpAddPool.Protocol = "bgp"
				tmpAddPool.Addresses = append(tmpAddPool.Addresses, subnet.CIDR.String())
				configStruct.Networks = append(configStruct.Networks, tmpAddPool)
			}
		}
	}

	configStruct.PeerSwitches = getMetalLBPeerSwitches(bgpPeers, configStruct)

	return configStruct
}

// WriteMetalLBConfigMap creates the yaml configmap
func WriteMetalLBConfigMap(path string, v *viper.Viper, networks map[string]*csi.IPV4Network, switches []*csi.ManagementSwitch) {

	tpl, err := template.New("mtllbconfigmap").Parse(string(MetalLBConfigMapTemplate))
	if err != nil {
		log.Printf("The template failed to render because: %v \n", err)
	}

	configStruct := GetMetalLBConfig(v, networks, switches)

	csiFiles.WriteTemplate(filepath.Join(path, "metallb.yaml"), tpl, configStruct)
}

// getMetalLBPeerSwitches returns a list of switches  that should be used as metallb peers
func getMetalLBPeerSwitches(bgpPeers []string, configStruct MetalLBConfigMap) []PeerDetail {

	switchTypeMap := map[string][]PeerDetail{
		"spine": configStruct.SpineSwitches,
		"leaf":  configStruct.LeafSwitches,
		"edge":  configStruct.EdgeSwitches,
	}

	for _, peerType := range bgpPeers {
		if peerSwitches, ok := switchTypeMap[peerType]; ok {
			if len(peerSwitches) == 0 {
				log.Fatalf("bgp-peer-types: %s specified but none defined in switch_metadata.csv\n", peerType)
			}
			configStruct.PeerSwitches = append(configStruct.PeerSwitches, peerSwitches...)
		} else {
			log.Fatalf("bgp-peer-types: unrecognized option: %s\n", peerType)
		}
	}

	return configStruct.PeerSwitches
}
