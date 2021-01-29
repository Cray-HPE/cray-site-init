/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package pit

import (
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/viper"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/csi"
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
    peers:{{range .SpineSwitches}}
    - peer-address: {{ . }}
      peer-asn: {{ $.ASN }}
      my-asn: {{ $.ASN }}
      {{- end}}
    address-pools:{{range $name, $subnet := .Networks}}
    - name: {{$name}}
      protocol: bgp
      addresses:
      - {{ $subnet }}
    {{- end}}
`)

// MetalLBConfigMap holds information needed by the MetalLBConfigMapTemplate
type MetalLBConfigMap struct {
	ASN           string
	SpineSwitches []string
	Networks      map[string]string
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

	tpl, err := template.New("mtllbconfigmap").Parse(string(MetalLBConfigMapTemplate))
	if err != nil {
		log.Printf("The template failed to render because: %v \n", err)
	}
	var configStruct MetalLBConfigMap
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
