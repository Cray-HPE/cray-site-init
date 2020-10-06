/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

/*
This package bridges the gap between the SLS view of the Shasta System and one that is useful
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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
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

// ConvertSLSNetworks converts from the SLS version of
// Network Definitions to a list of IPV4Networks
func ConvertSLSNetworks(sls sls_common.SLSState) []IPV4Network {
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
	return ourNetworks
}

// ExtractNCNBMCInfo pulls the port information for the BMCs of all Management Nodes
func ExtractNCNBMCInfo(sls sls_common.SLSState) ([]BootstrapNCNMetadata, error) {
	var ncns []BootstrapNCNMetadata
	for key, node := range sls.Hardware {
		if node.Type == sls_common.Node {
			var extra sls_common.ComptypeNode
			err := mapstructure.Decode(node.ExtraPropertiesRaw, &extra)
			if err != nil {
				return ncns, err
			}
			if extra.Role == "Management" {
				//log.Printf("Adding ", key, " to the list with Parent = ", node.Parent)
				mgmtSwitch, port, err := portForXname(sls.Hardware, node.Parent)
				if err != nil { // Sometimes the port is not available.  We *should* be able to continue
					log.Printf("%v %v\n", err, port)
				}
				ncns = append(ncns, BootstrapNCNMetadata{
					Xname:   key,
					Role:    extra.Role,
					Subrole: extra.SubRole,
					BmcPort: mgmtSwitch + ":" + port,
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
	return "", "", errors.New("Couldn't find switch port for NCN: " + xname)
}

// ExtractSwitches reads the SLSState object and finds any switches
func ExtractSwitches(sls sls_common.SLSState) error {
	for _, node := range sls.Hardware {
		if node.Type == sls_common.MgmtSwitch {
			log.Printf("Found Switch: %v \n", node)
			var extra sls_common.ComptypeMgmtSwitch
			err := mapstructure.Decode(node.ExtraPropertiesRaw, &extra)
			if err != nil {
				return err
			}
			// TODO: Map the switch data to either an SLS object or an internal object as needed by SLS
			// Update the signature to return the lost of switches
		}
	}
	return nil
}

// func GenerateSLSState(hmn_connections_path string, sls_basic_info sls_common.SLSStateInput) {
// 	// TODO : or not TODO  The HMS team is still deciding if this belongs here.  I think it does.  10-06-2020 -ALT
// 	//
// }
