/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

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
