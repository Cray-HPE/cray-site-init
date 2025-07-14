// MIT License
//
// (C) Copyright [2020-2022] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

// This file is contains struct defines for CompEthInterfaces
package sm

// This package defines structures for component ethernet interfaces

import (
	"strings"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

//
// Format checking for database keys and query parameters.
//

var ErrCompEthInterfaceBadMAC = base.NewHMSError("sm", "Invalid CompEthInterface MAC Address")
var ErrCompEthInterfaceBadCompID = base.NewHMSError("sm", "Invalid CompEthInterface component ID")
var ErrCompEthInterfaceBadIPAddress = base.NewHMSError("sm", "Invalid CompEthInterface IP Address")

///////////////////////////////////////////////////////////////////////////
//
// CompEthInterface
//
///////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////
// V1 API
///////////////////////////////////////////////////////////////////////////

// A component ethernet interface is an IP address <-> MAC address relation.
// This structure is used on the v1 CompEthInterface APIs
type CompEthInterface struct {
	ID         string `json:"ID"`
	Desc       string `json:"Description"`
	MACAddr    string `json:"MACAddress"`
	IPAddr     string `json:"IPAddress"`
	LastUpdate string `json:"LastUpdate"`
	CompID     string `json:"ComponentID"`
	Type       string `json:"Type"`
}

// Allocate and initialize new CompEthInterface struct, validating it.
func NewCompEthInterface(desc, macAddr, ipAddr, compID string) (*CompEthInterface, error) {
	if macAddr == "" {
		return nil, ErrCompEthInterfaceBadMAC
	}
	cei := new(CompEthInterface)
	cei.Desc = desc
	cei.MACAddr = strings.ToLower(macAddr)
	cei.ID = strings.ReplaceAll(cei.MACAddr, ":", "")
	if cei.ID == "" {
		return nil, ErrCompEthInterfaceBadMAC
	}
	cei.IPAddr = ipAddr
	if compID != "" {
		cei.CompID = xnametypes.VerifyNormalizeCompID(compID)
		if cei.CompID == "" {
			return nil, ErrCompEthInterfaceBadCompID
		}
		cei.Type = xnametypes.GetHMSTypeString(cei.CompID)
	}
	return cei, nil
}

// Patchable fields if included in payload.
type CompEthInterfacePatch struct {
	Desc   *string `json:"Description"`
	IPAddr *string `json:"IPAddress"`
	CompID *string `json:"ComponentID"`
}

///////////////////////////////////////////////////////////////////////////
// V2 API
///////////////////////////////////////////////////////////////////////////

// A component ethernet interface is an IP addresses <-> MAC address relation.
// This structure is used on the v2 CompEthInterface APIs
type CompEthInterfaceV2 struct {
	ID         string `json:"ID"`
	Desc       string `json:"Description"`
	MACAddr    string `json:"MACAddress"`
	LastUpdate string `json:"LastUpdate"`
	CompID     string `json:"ComponentID"`
	Type       string `json:"Type"`

	IPAddrs []IPAddressMapping `json:"IPAddresses"`
}

func (cei *CompEthInterfaceV2) ToV1() *CompEthInterface {
	ceiV1 := new(CompEthInterface)

	ceiV1.ID = cei.ID
	ceiV1.Desc = cei.Desc
	ceiV1.MACAddr = cei.MACAddr
	ceiV1.LastUpdate = cei.LastUpdate
	ceiV1.CompID = cei.CompID
	ceiV1.Type = cei.Type

	// Provide backwards compatible-ness use the first element (if present) to represent the
	// IPAddr field
	if len(cei.IPAddrs) > 0 {
		ceiV1.IPAddr = cei.IPAddrs[0].IPAddr
	}

	return ceiV1
}

// Allocate and initialize new CompEthInterfaceV2 struct, validating it.
func NewCompEthInterfaceV2(desc, macAddr, compID string, ipAddrs []IPAddressMapping) (*CompEthInterfaceV2, error) {
	if macAddr == "" {
		return nil, ErrCompEthInterfaceBadMAC
	}
	cei := new(CompEthInterfaceV2)
	cei.Desc = desc
	cei.MACAddr = strings.ToLower(macAddr)
	cei.ID = strings.ReplaceAll(cei.MACAddr, ":", "")
	if cei.ID == "" {
		return nil, ErrCompEthInterfaceBadMAC
	}
	// Initialize empty slices
	if ipAddrs == nil {
		ipAddrs = []IPAddressMapping{}
	}
	cei.IPAddrs = ipAddrs
	for _, ipm := range cei.IPAddrs {
		if err := ipm.Verify(); err != nil {
			return nil, err
		}
	}
	if compID != "" {
		cei.CompID = xnametypes.VerifyNormalizeCompID(compID)
		if cei.CompID == "" {
			return nil, ErrCompEthInterfaceBadCompID
		}
		cei.Type = xnametypes.GetHMSTypeString(cei.CompID)
	}
	return cei, nil
}

// Patchable fields if included in payload.
type CompEthInterfaceV2Patch struct {
	Desc    *string             `json:"Description"`
	CompID  *string             `json:"ComponentID"`
	IPAddrs *[]IPAddressMapping `json:"IPAddresses"`
}

// IPAddressMapping represents an IP Address to network mapping. The network field is optional
type IPAddressMapping struct {
	IPAddr  string `json:"IPAddress"`
	Network string `json:"Network,omitempty"`
}

// Allocate and initialize new IPAddressMapping struct, validating it.
func NewIPAddressMapping(ipAddr, network string) (*IPAddressMapping, error) {
	ipm := new(IPAddressMapping)
	ipm.IPAddr = ipAddr
	ipm.Network = network

	return ipm, ipm.Verify()
}

// Validate the contents of the IP Address mapping
func (ipm *IPAddressMapping) Verify() error {
	// Can't have an empty IP Address
	if ipm.IPAddr == "" {
		return ErrCompEthInterfaceBadIPAddress
	}

	return nil
}

// Patchable fields if included in payload.
type IPAddressMappingPatch struct {
	Network *string `json:"Network"`
}
