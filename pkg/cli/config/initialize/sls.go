/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

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

package initialize

/*
This package bridges the gap between the SLS view of the CRAY System and one that is useful
for administrators who are trying to install and upgrade a system. Where possible, we'd like
to reuse datastructures, but that's not practical, at least initially because of the very
ways the two tools use the Data.

This is important so we can consume from the dumpstate endpoint of SLS and subsequently
generate a payload suitable for loadstate without forcing users to interact directly with
the SLS structure.
*/

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	slsInit "github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
)

// ExtractUANs pulls the information needed to assign CAN addresses to the UAN xnames
func ExtractUANs(sls *slsCommon.SLSState) (
	[]LogicalUAN, error,
) {
	var uans []LogicalUAN
	uanIndex := int(1)
	for key, node := range sls.Hardware {
		if node.Type == slsCommon.Node {
			var extra slsCommon.ComptypeNode
			err := mapstructure.Decode(
				node.ExtraPropertiesRaw,
				&extra,
			)
			if err != nil {
				return uans, err
			}
			if extra.Role == "Application" && extra.SubRole == "UAN" {
				if extra.Aliases == nil {
					log.Fatal("ERROR: UANs must have at least one alias defined in the application-node-config-yaml file")
				}
				uans = append(
					uans,
					LogicalUAN{
						Xname:    key,
						Role:     extra.Role,
						Subrole:  extra.SubRole,
						Hostname: extra.Aliases[0],
						Aliases:  extra.Aliases,
					},
				)
				uanIndex++
			}
		}
	}
	return uans, nil
}

// ExtractSLSNCNs pulls the port information for the BMCs of all Management Nodes
func ExtractSLSNCNs(sls *slsCommon.SLSState) (
	[]LogicalNCN, error,
) {
	var ncns []LogicalNCN
	for key, node := range sls.Hardware {
		if node.Type == slsCommon.Node {
			var extra slsCommon.ComptypeNode
			err := mapstructure.Decode(
				node.ExtraPropertiesRaw,
				&extra,
			)
			if err != nil {
				return ncns, err
			}
			if extra.Role == "Management" {
				mgmtSwitch, port, err := portForXname(
					sls.Hardware,
					node.Parent,
				)
				if err != nil { // Sometimes the port is not available. We *should* be able to continue
					log.Printf(
						"%v %v\n",
						err,
						port,
					)
				}
				ncns = append(
					ncns,
					LogicalNCN{
						Xname:    key,
						Role:     extra.Role,
						Subrole:  extra.SubRole,
						Hostname: extra.Aliases[0],
						Aliases:  extra.Aliases,
						BmcPort:  mgmtSwitch + ":" + port,
					},
				)
			}
		}
	}
	return ncns, nil
}

// Return a tuple of strings that match switch and switchport for the BMC
func portForXname(
	hardware map[string]slsCommon.GenericHardware, xname string,
) (
	string, string, error,
) {
	for _, node := range hardware {
		if node.Type == "comptype_mgmt_switch_connector" {
			var extra slsCommon.ComptypeMgmtSwitchConnector
			err := mapstructure.Decode(
				node.ExtraPropertiesRaw,
				&extra,
			)
			if err != nil {
				return "", "", err
			}
			for _, nodeNIC := range extra.NodeNics {
				if xname == nodeNIC {
					networkSwitch := node.Parent
					networkPort := extra.VendorName
					return networkSwitch, networkPort, nil

				}
			}
		}
	}
	// log.Printf("Couldn't find", xname)
	return "", "", errors.New("WARNING (Not Fatal): Couldn't find switch port for NCN: " + xname)
}

// CabinetForXname extracts the cabinet identifier from an xname
func CabinetForXname(xname string) (
	string, error,
) {
	r := regexp.MustCompile("(x[0-9]+)") // the leading x is not part of the cabinet identifier
	matches := r.FindStringSubmatch(xname)
	if len(matches) != 2 {
		err := fmt.Errorf(
			"failed to find cabinet for %v",
			xname,
		)
		return "", err
	}
	return matches[0], nil
}

// GetSLSCabinets will get all of the cabinets from SLS of the specified class
func GetSLSCabinets(
	state slsCommon.SLSState, class slsCommon.CabinetType,
) []slsCommon.GenericHardware {
	var cabinets []slsCommon.GenericHardware
	for _, hardware := range state.Hardware {
		if hardware.Type == slsCommon.Cabinet && hardware.Class == class {
			cabinets = append(
				cabinets,
				hardware,
			)
		}
	}

	return cabinets
}

// GenerateDefaultNetworkConfigs generates a map containing every network that could be used on the system, initialized
// with default values.
func GenerateDefaultNetworkConfigs(
	switches []*networking.ManagementSwitch,
	logicalNCNs []*LogicalNCN,
	cabinetDetailList []sls.CabinetGroupDetail,
) (defaultNetConfigs map[string]slsInit.NetworkLayoutConfiguration) {
	v := viper.GetViper()
	var riverCabinetCount, mountainCabinetCount, hillCabinetCount int
	for _, cab := range cabinetDetailList {
		switch class, _ := cab.Kind.Class(); class {
		case slsCommon.ClassRiver:
			riverCabinetCount += len(cab.CabinetIDs())
		case slsCommon.ClassMountain:
			mountainCabinetCount += len(cab.CabinetIDs())
		case slsCommon.ClassHill:
			hillCabinetCount += len(cab.CabinetIDs())
		}
	}

	// Check if any FabricManager nodes exist in the system
	hasFabricManagerNodes := false
	for _, ncn := range logicalNCNs {
		if ncn.Subrole == "FabricManager" {
			hasFabricManagerNodes = true
			break
		}
	}

	defaultNetConfigs = make(map[string]slsInit.NetworkLayoutConfiguration)
	// Prepare the network layout configs for generating the networks
	defaultNetConfigs["BICAN"] = slsInit.GenDefaultBICANConfig(v.GetString("bican-user-network-name"))
	defaultNetConfigs["CMN"] = slsInit.GenDefaultCMNConfig(
		len(logicalNCNs),
		len(switches),
	)
	hmnConfig := slsInit.GenDefaultHMNConfig()
	hmnConfig.HasFabricManagerNodes = hasFabricManagerNodes
	defaultNetConfigs["HMN"] = hmnConfig
	defaultNetConfigs["HSN"] = slsInit.GenDefaultHSNConfig()
	defaultNetConfigs["MTL"] = slsInit.GenDefaultMTLConfig()
	nmnConfig := slsInit.GenDefaultNMNConfig()
	nmnConfig.HasFabricManagerNodes = hasFabricManagerNodes
	defaultNetConfigs["NMN"] = nmnConfig
	if v.GetString("bican-user-network-name") == "CAN" || v.GetBool("retain-unused-user-network") {
		defaultNetConfigs["CAN"] = slsInit.GenDefaultCANConfig()
	}
	if v.GetString("bican-user-network-name") == "CHN" || v.GetBool("retain-unused-user-network") {
		defaultNetConfigs["CHN"] = slsInit.GenDefaultCHNConfig()
	}

	if defaultNetConfigs["HMN"].GroupNetworksByCabinetType {
		if mountainCabinetCount > 0 || hillCabinetCount > 0 {
			tmpHmnMtn := slsInit.GenDefaultHMNConfig()
			tmpHmnMtn.Template.Name = "HMN_MTN"
			tmpHmnMtn.Template.FullName = "Mountain Compute Hardware Management Network"
			tmpHmnMtn.Template.VlanRange = []int16{
				3000,
				3999,
			}
			tmpHmnMtn.Template.CIDR4 = networking.DefaultHMNMTNString
			tmpHmnMtn.SubdivideByCabinet = true
			tmpHmnMtn.IncludeBootstrapDHCP = false
			tmpHmnMtn.SuperNetHack = false
			tmpHmnMtn.IncludeNetworkingHardwareSubnet = false
			defaultNetConfigs["HMN_MTN"] = tmpHmnMtn
		}
		if riverCabinetCount > 0 {
			tmpHmnRvr := slsInit.GenDefaultHMNConfig()
			tmpHmnRvr.Template.Name = "HMN_RVR"
			tmpHmnRvr.Template.FullName = "River Compute Hardware Management Network"
			tmpHmnRvr.Template.VlanRange = []int16{
				1513,
				1769,
			}
			tmpHmnRvr.Template.CIDR4 = networking.DefaultHMNRVRString
			tmpHmnRvr.SubdivideByCabinet = true
			tmpHmnRvr.IncludeBootstrapDHCP = false
			tmpHmnRvr.SuperNetHack = false
			tmpHmnRvr.IncludeNetworkingHardwareSubnet = false
			defaultNetConfigs["HMN_RVR"] = tmpHmnRvr
		}

	}

	if defaultNetConfigs["NMN"].GroupNetworksByCabinetType {
		if mountainCabinetCount > 0 || hillCabinetCount > 0 {
			tmpNmnMtn := slsInit.GenDefaultNMNConfig()
			tmpNmnMtn.Template.Name = "NMN_MTN"
			tmpNmnMtn.Template.FullName = "Mountain Compute Node Management Network"
			tmpNmnMtn.Template.VlanRange = []int16{
				2000,
				2999,
			}
			tmpNmnMtn.Template.CIDR4 = networking.DefaultNMNMTNString
			tmpNmnMtn.SubdivideByCabinet = true
			tmpNmnMtn.IncludeBootstrapDHCP = false
			tmpNmnMtn.SuperNetHack = false
			tmpNmnMtn.IncludeNetworkingHardwareSubnet = false
			tmpNmnMtn.IncludeUAISubnet = false
			defaultNetConfigs["NMN_MTN"] = tmpNmnMtn
		}
		if riverCabinetCount > 0 {
			tmpNmnRvr := slsInit.GenDefaultNMNConfig()
			tmpNmnRvr.Template.Name = "NMN_RVR"
			tmpNmnRvr.Template.FullName = "River Compute Node Management Network"
			tmpNmnRvr.Template.VlanRange = []int16{
				1770,
				1999,
			}
			tmpNmnRvr.Template.CIDR4 = networking.DefaultNMNRVRString
			tmpNmnRvr.SubdivideByCabinet = true
			tmpNmnRvr.IncludeBootstrapDHCP = false
			tmpNmnRvr.SuperNetHack = false
			tmpNmnRvr.IncludeNetworkingHardwareSubnet = false
			tmpNmnRvr.IncludeUAISubnet = false
			defaultNetConfigs["NMN_RVR"] = tmpNmnRvr
		}

	}
	return defaultNetConfigs
}

// GenerateNetworkConfigs creates a network configuration map of all networks for the system.
func GenerateNetworkConfigs(netconfig map[string]slsInit.NetworkLayoutConfiguration) (internalNetConfigs map[string]slsInit.NetworkLayoutConfiguration, err error) {
	v := viper.GetViper()
	internalNetConfigs = make(map[string]slsInit.NetworkLayoutConfiguration)
	for name, layout := range netconfig {
		myLayout := layout

		// Update with flags
		normalizedName := strings.ReplaceAll(
			strings.ToLower(name),
			"_",
			"-",
		)
		cidr4Key := fmt.Sprintf(
			"%s-cidr4",
			normalizedName,
		)
		cidrKey := fmt.Sprintf(
			"%s-cidr",
			normalizedName,
		)
		cidr6Key := fmt.Sprintf(
			"%s-cidr6",
			normalizedName,
		)

		// Handle IPv4 CIDRs, networks with IPv6 will use a different key for their cidr4.
		if v.IsSet(cidr4Key) {
			myLayout.Template.CIDR4 = v.GetString(cidr4Key)
		} else if v.IsSet(cidrKey) {
			myLayout.Template.CIDR4 = v.GetString(cidrKey)
		}
		if v.IsSet(cidr6Key) {
			myLayout.Template.CIDR6 = v.GetString(cidr6Key)
		}

		// Use CLI/file input values if available, otherwise defaults
		baseVlanName := fmt.Sprintf(
			"%v-bootstrap-vlan",
			normalizedName,
		)
		if v.IsSet(baseVlanName) {
			baseVlan := int16(v.GetInt(baseVlanName))
			myLayout.BaseVlan = baseVlan
			myLayout.Template.VlanRange[0] = baseVlan
		} else {
			myLayout.BaseVlan = layout.Template.VlanRange[0]
		}

		// Check VLAN allocations for re-use and overlaps
		if len(layout.Template.VlanRange) == 2 {
			err := networking.AllocateVlanRange(
				myLayout.Template.VlanRange[0],
				myLayout.Template.VlanRange[1],
			)
			if err != nil {
				log.Fatalln(
					"Unable to allocate VLAN range for",
					myLayout.Template.Name,
					err,
				)
			} else {
				log.Println(
					"Allocating VLANs",
					myLayout.Template.Name,
					myLayout.Template.VlanRange[0],
					myLayout.Template.VlanRange[1],
				)
			}
		} else {
			err := networking.AllocateVLAN(
				uint16(myLayout.Template.VlanRange[0]),
			)
			if err != nil {
				log.Fatalln(
					"Unable to allocate single VLAN for",
					myLayout.Template.Name,
					myLayout.Template.VlanRange[0],
					err,
				)
			} else {
				log.Println(
					"Allocating VLAN ",
					myLayout.Template.Name,
					myLayout.Template.VlanRange[0],
				)
			}
		}

		allocated, err := networking.IsVLANAllocated(uint16(myLayout.BaseVlan))
		if !allocated {
			log.Fatalln(
				"VLAN for",
				layout.Template.Name,
				"has not been initialized by defaults or input values:",
				err,
			)
		}
		myLayout.AdditionalNetworkingSpace = v.GetInt("management-net-ips")
		internalNetConfigs[name] = myLayout
	}
	return internalNetConfigs, err
}
