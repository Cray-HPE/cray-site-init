/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

/*
This package bridges the gap between the SLS view of the CRAY System and one that is useful
for administrators who are trying to install and upgrade a system.  Where possible, we'd like
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

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

// ParseSLSFile takes a path and returns an SLSState struct for parsing
func ParseSLSFile(path string) (sls_common.SLSState, error) {
	var existingState sls_common.SLSState
	jsonSLSFile, err := os.Open(path)
	if err != nil {
		return existingState, err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonSLSFile.Close()
	buf, _ := ioutil.ReadAll(jsonSLSFile)
	err = json.Unmarshal(buf, &existingState)
	if err != nil {
		return existingState, err
	}
	return existingState, nil
}

// ParseSLSfromURL takes a url (likely the sls dumpstate url) and returns a useful struct
func ParseSLSfromURL(url string) (sls_common.SLSState, error) {
	var existingState sls_common.SLSState

	slsClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return existingState, err
	}
	req.Header.Set("User-Agent", "shasta-1.4-installer")
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
	err = json.Unmarshal(body, &existingState)
	return existingState, err
}

// ExtractSLSNetworks converts from the SLS version of
// Network Definitions to a list of IPV4Networks
func ExtractSLSNetworks(sls *sls_common.SLSState) ([]IPV4Network, error) {
	var ourNetworks []IPV4Network
	for key, element := range sls.Networks {
		// log.Printf("Key:", key, "=>", "Element:", element, "(", reflect.TypeOf(element), ")")
		// log.Printf("Extra Properties:", element.ExtraPropertiesRaw)
		tempNetwork := new(IPV4Network)
		tempNetwork.FullName = element.FullName
		tempNetwork.CIDR = strings.Join(element.IPRanges, ",")
		tempNetwork.Name = key
		// Pull the VLAN Range from the Extra Properties
		// tempNetwork.VlanRange
		ourNetworks = append(ourNetworks, *tempNetwork)
	}
	return ourNetworks, nil
}

// ExtractUANs pulls the information needed to assign CAN addresses to the UAN xnames
func ExtractUANs(sls *sls_common.SLSState) ([]LogicalUAN, error) {
	var uans []LogicalUAN
	uanIndex := int(1)
	for key, node := range sls.Hardware {
		if node.Type == sls_common.Node {
			var extra sls_common.ComptypeNode
			err := mapstructure.Decode(node.ExtraPropertiesRaw, &extra)
			if err != nil {
				return uans, err
			}
			if extra.Role == "Application" && extra.SubRole == "UAN" {
				if extra.Aliases == nil {
					log.Fatal("ERROR: UANs must have at least one alias defined in the application-node-config-yaml file")
				}
				uans = append(uans, LogicalUAN{
					Xname:    key,
					Role:     extra.Role,
					Subrole:  extra.SubRole,
					Hostname: extra.Aliases[0],
					Aliases:  extra.Aliases,
				})
				uanIndex++
			}
		}
	}
	return uans, nil
}

// ExtractSLSNCNs pulls the port information for the BMCs of all Management Nodes
func ExtractSLSNCNs(sls *sls_common.SLSState) ([]LogicalNCN, error) {
	var ncns []LogicalNCN
	for key, node := range sls.Hardware {
		if node.Type == sls_common.Node {
			var extra sls_common.ComptypeNode
			err := mapstructure.Decode(node.ExtraPropertiesRaw, &extra)
			if err != nil {
				return ncns, err
			}
			if extra.Role == "Management" {
				// log.Printf("Adding %v to the list with Parent = %v", key, node.Parent)
				// log.Printf("Node = %v and Extra = %v", node, extra)
				mgmtSwitch, port, err := portForXname(sls.Hardware, node.Parent)
				if err != nil { // Sometimes the port is not available.  We *should* be able to continue
					log.Printf("%v %v\n", err, port)
				}
				ncns = append(ncns, LogicalNCN{
					Xname:    key,
					Role:     extra.Role,
					Subrole:  extra.SubRole,
					Hostname: extra.Aliases[0],
					Aliases:  extra.Aliases,
					BmcPort:  mgmtSwitch + ":" + port,
				})
			}
		}
	}
	return ncns, nil
}

// Return a tuple of strings that match switch and switchport for the BMC
func portForXname(hardware map[string]sls_common.GenericHardware, xname string) (string, string, error) {
	for _, node := range hardware {
		if node.Type == "comptype_mgmt_switch_connector" {
			var extra sls_common.ComptypeMgmtSwitchConnector
			err := mapstructure.Decode(node.ExtraPropertiesRaw, &extra)
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
func ExtractSLSSwitches(sls *sls_common.SLSState) ([]ManagementSwitch, error) {
	var switches []ManagementSwitch
	for _, node := range sls.Hardware {
		if node.Type == sls_common.MgmtSwitch {
			var extra sls_common.ComptypeMgmtSwitch
			err := mapstructure.Decode(node.ExtraPropertiesRaw, &extra)
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
func CabinetForXname(xname string) (string, error) {
	r := regexp.MustCompile("(x[0-9]+)") // the leading x is not part of the cabinet identifier
	matches := r.FindStringSubmatch(xname)
	if len(matches) != 2 {
		err := fmt.Errorf("failed to find cabinet for %v", xname)
		return "", err
	}
	return matches[0], nil
}

// GetSLSCabinets will get all of the cabinets from SLS of the specified class
func GetSLSCabinets(state sls_common.SLSState, class sls_common.CabinetType) []sls_common.GenericHardware {
	cabinets := []sls_common.GenericHardware{}
	for _, hardware := range state.Hardware {
		if hardware.Type == sls_common.Cabinet && hardware.Class == class {
			cabinets = append(cabinets, hardware)
		}
	}

	return cabinets
}
