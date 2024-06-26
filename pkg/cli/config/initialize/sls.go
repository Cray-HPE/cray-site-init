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

package initialize

/*
This package bridges the gap between the SLS view of the CRAY System and one that is useful
for administrators who are trying to install and upgrade a system. Where possible, we'd like
to reuse datastructures, but that's not practical, at least initially because of the very
ways the two tools use the data.

This is important so we can consume from the dumpstate endpoint of SLS and subsequently
generate a payload suitable for loadstate without forcing users to interact directly with
the SLS structure.
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"

	slsCommon "github.com/Cray-HPE/hms-sls/pkg/sls-common"

	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

// ParseSLSFile takes a path and returns an SLSState struct for parsing
func ParseSLSFile(path string) (
	slsCommon.SLSState, error,
) {
	var existingState slsCommon.SLSState
	jsonSLSFile, err := os.Open(path)
	if err != nil {
		return existingState, err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonSLSFile.Close()
	buf, _ := ioutil.ReadAll(jsonSLSFile)
	err = json.Unmarshal(
		buf,
		&existingState,
	)
	if err != nil {
		return existingState, err
	}
	return existingState, nil
}

// ParseSLSfromURL takes a url (likely the sls dumpstate url) and returns a useful struct
func ParseSLSfromURL(url string) (
	slsCommon.SLSState, error,
) {
	var existingState slsCommon.SLSState

	slsClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}
	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return existingState, err
	}
	req.Header.Set(
		"User-Agent",
		"shasta-1.4-installer",
	)
	res, err := slsClient.Do(req)
	if err != nil {
		return existingState, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return existingState, err
	}
	err = json.Unmarshal(
		body,
		&existingState,
	)
	return existingState, err
}

// ExtractSLSNetworks converts from the SLS version of
// Network Definitions to a list of IPV4Networks
func ExtractSLSNetworks(sls *slsCommon.SLSState) (
	[]networking.IPV4Network, error,
) {
	var ourNetworks []networking.IPV4Network
	for key, element := range sls.Networks {
		// log.Printf("Key:", key, "=>", "Element:", element, "(", reflect.TypeOf(element), ")")
		// log.Printf("Extra Properties:", element.ExtraPropertiesRaw)
		tempNetwork := new(networking.IPV4Network)
		tempNetwork.FullName = element.FullName
		tempNetwork.CIDR = strings.Join(
			element.IPRanges,
			",",
		)
		tempNetwork.Name = key
		// Pull the VLAN Range from the Extra Properties
		// tempNetwork.VlanRange
		ourNetworks = append(
			ourNetworks,
			*tempNetwork,
		)
	}
	return ourNetworks, nil
}

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
				// log.Printf("Adding %v to the list with Parent = %v", key, node.Parent)
				// log.Printf("Node = %v and Extra = %v", node, extra)
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

// ExtractSLSSwitches reads the SLSState object and finds any switches
func ExtractSLSSwitches(sls *slsCommon.SLSState) (
	[]networking.ManagementSwitch, error,
) {
	var switches []networking.ManagementSwitch
	for _, node := range sls.Hardware {
		if node.Type == slsCommon.MgmtSwitch {
			var extra slsCommon.ComptypeMgmtSwitch
			err := mapstructure.Decode(
				node.ExtraPropertiesRaw,
				&extra,
			)
			if err != nil {
				return switches, err
			}
			// TODO: Map the switch data to either an SLS object or an internal object as needed by SLS
			// Update the signature to return the lost of switches
		}
	}
	return switches, nil
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
	cabinets := []slsCommon.GenericHardware{}
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
