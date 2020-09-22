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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
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
		fmt.Println("Key:", key, "=>", "Element:", element, "(", reflect.TypeOf(element), ")")
		fmt.Println("Extra Properties:", element.ExtraPropertiesRaw)
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
func ExtractNCNBMCInfo(sls sls_common.SLSState) []BootstrapNCNMetadata {
	var ncns []BootstrapNCNMetadata
	for key, node := range sls.Hardware {
		if node.Type == sls_common.Node {
			var extra sls_common.ComptypeNode
			err := mapstructure.Decode(node.ExtraPropertiesRaw, &extra)
			if err != nil {
				fmt.Printf("got data of type %T but wanted sls_common.ComptypeNode. \n", extra)
				fmt.Println(extra)
			} else {
				if extra.Role == "Management" {
					//fmt.Println("Adding ", key, " to the list with Parent = ", node.Parent)
					fmt.Println("Xname is", node.Xname, "Parent is", node.Parent)
					parent, err := nodeForXname(sls.Hardware, node.Parent)
					if err != nil {
						fmt.Println(err, parent)
					}
					ncns = append(ncns, BootstrapNCNMetadata{
						Xname:   key,
						Role:    extra.Role,
						Subrole: extra.SubRole,
					})
				}
			}
		}
	}
	return ncns
}

// Return a GeneicHardware struct for a particular xname
func nodeForXname(hardware map[string]sls_common.GenericHardware, xname string) (sls_common.GenericHardware, error) {
	var thing sls_common.GenericHardware
	for key, node := range hardware {
		err := mapstructure.Decode(node, &thing)
		if err != nil {
			return thing, err
		}
		// fmt.Println(thing.Xname, "!=", xname)
		if key == xname {
			fmt.Println("Found ", thing)
			return thing, nil
		}
	}
	fmt.Println("Couldn't find", xname)
	return sls_common.GenericHardware{}, errors.New(" not found" + xname)
}
