/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

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

package networking

import (
	"net"
	"strings"
)

// BootstrapSwitchMetadata is a type that matches the switch_metadata.csv file as
// Switch Xname, Type
// The type can be CDU, Spine, Leaf, or LeafBMC
type BootstrapSwitchMetadata struct {
	Xname string `json:"xname" csv:"Switch Xname"`
	Type  string `json:"type" csv:"Type"`
}

// DefaultApplicationNodePrefixes is the list of default Application node prefixes, for source column in the hmn_connections.json
var DefaultApplicationNodePrefixes = []string{
	"uan",
	"gn",
	"ln",
}

// DefaultApplicationNodeSubroles is the default prefix<->subrole mapping for application node subroles, these can be overridden via ApplicationNodeConfig
var DefaultApplicationNodeSubroles = map[string]string{
	"uan": "UAN",
	"ln":  "UAN", // Nodes with the ln prefix are also UAN nodes
	"gn":  "Gateway",
}

// SubrolePlaceHolder is the placeholder used to indicate that a prefix has no subrole mapping in ApplicationNodeConfig.
const SubrolePlaceHolder = "~fixme~"

// ValidNetNames is the list of strings that enumerate valid main network names
var ValidNetNames = []string{
	"BICAN",
	"CAN",
	"CHN",
	"CMN",
	"HMN",
	"HMN_MTN",
	"HMN_RVR",
	"MTL",
	"NMN",
	"NMN_MTN",
	"NMN_RVR",
}

// IPNetfromCIDRString converts from a string to an net.IPNet struct
func IPNetfromCIDRString(mynet string) *net.IPNet {
	_, ipnet, _ := net.ParseCIDR(mynet)
	return ipnet
}

// DefaultUAISubnetReservations is the map of dns names and aliases
var DefaultUAISubnetReservations = map[string][]string{
	"uai_nmn_blackhole": {"uai-nmn-blackhole"},
	"slurmctld_service": {
		"slurmctld-service",
		"slurmctld-service-nmn",
	},
	"slurmdbd_service": {
		"slurmdbd-service",
		"slurmdbd-service-nmn",
	},
	"pbs_service": {
		"pbs-service",
		"pbs-service-nmn",
	},
	"pbs_comm_service": {
		"pbs-comm-service",
		"pbs-comm-service-nmn",
	},
}

// PinnedReservation is a simple struct to work with our abomination of a PinnedMetalLBReservations
type PinnedReservation struct {
	IPByte  uint8
	Aliases []string
}

// PinnedMetalLBReservations is the map of dns names and aliases with the
// required final octet of th ip address
// *** This structure is only necessary to pin ip addresses as we shift from 1.3 to 1.4 ***
// *** *** *** To anyone editing this code in the future, PLEASE DON'T MAKE IT BETTER *** *** ***
// *** *** *** This code is written to be thrown away with a fully dynamic ip addressing scheme *** *** ***
var PinnedMetalLBReservations = map[string]PinnedReservation{
	"istio-ingressgateway": {
		71,
		strings.Split(
			"api-gw-service api-gw-service-nmn.local packages registry spire.local api_gw_service registry.local packages packages.local spire",
			" ",
		),
	},
	"istio-ingressgateway-local": {
		81,
		[]string{"api-gw-service.local"},
	},
	"rsyslog-aggregator": {
		72,
		[]string{"rsyslog-agg-service"},
	},
	"cray-tftp": {
		60,
		[]string{"tftp-service"},
	},
	"unbound": {
		225,
		[]string{"unbound"},
	},
	"docker-registry": {
		73,
		[]string{"docker_registry_service"},
	},
}
