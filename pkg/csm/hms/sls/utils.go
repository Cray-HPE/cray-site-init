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

package sls

import (
	"context"
	"encoding/json"
	"fmt"

	base "github.com/Cray-HPE/hms-base/v2"
	slsClient "github.com/Cray-HPE/hms-sls/v2/pkg/sls-client"
	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
)

// GetManagementNCNs - Returns all the management NCNs from SLS.
func GetManagementNCNs(client slsClient.SLSClient) (managementNCNs []slsCommon.GenericHardware, err error) {
	hardware, err := client.GetAllHardware(context.Background())
	if err != nil {
		return managementNCNs, fmt.Errorf(
			"failed to get management hardware because %v",
			err,
		)
	}
	for _, h := range hardware {
		extraProperties, err := UnmarshalComptypeNode(&h)
		if err != nil {
			continue
		}
		if extraProperties.Role == base.RoleManagement.String() {
			managementNCNs = append(
				managementNCNs,
				h,
			)
		}
	}
	fmt.Println(len(managementNCNs))
	return managementNCNs, err
}

// UnmarshalComptypeNode reads the hardware.ExtraPropertiesRaw string into a slsCommon.ComptypeNode struct.
func UnmarshalComptypeNode(hardware *slsCommon.GenericHardware) (extraProperties slsCommon.ComptypeNode, err error) {
	extraPropertiesRaw, err := json.Marshal(hardware.ExtraPropertiesRaw)
	if err != nil {
		return extraProperties, fmt.Errorf(
			"failed to marshal [%s] as ComptypeNode because %v",
			hardware.Xname,
			err,
		)
	}
	err = json.Unmarshal(
		extraPropertiesRaw,
		&extraProperties,
	)
	if err != nil {
		return extraProperties, fmt.Errorf(
			"failed to unmarshal hardware [%s] as ComptypeNode because %v",
			hardware.Xname,
			err,
		)
	}
	return extraProperties, nil
}

// UnmarshalNetworkExtraProperties reads the network.ExtraPropertiesRaw string into a struct.
func UnmarshalNetworkExtraProperties(network *slsCommon.Network) (extraProperties slsCommon.NetworkExtraProperties, err error) {
	extraPropertiesRaw, err := json.Marshal(network.ExtraPropertiesRaw)
	if err != nil {
		return extraProperties, fmt.Errorf(
			"failed to marshal extra properties from network [%s] because %v",
			network.Name,
			err,
		)
	}
	err = json.Unmarshal(
		extraPropertiesRaw,
		&extraProperties,
	)
	if err != nil {
		return extraProperties, fmt.Errorf(
			"failed to unmarshal extra properties from network [%s] because %v",
			network.Name,
			err,
		)
	}
	return extraProperties, nil
}
