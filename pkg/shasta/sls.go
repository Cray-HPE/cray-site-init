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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

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
