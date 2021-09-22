// MIT License
//
// (C) Copyright [2018-2021] Hewlett Packard Enterprise Development LP
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

package sm

import (
	"encoding/json"
	base "github.com/Cray-HPE/hms-base"
	rf "github.com/Cray-HPE/hms-smd/pkg/redfish"
)

// This is a generic link to a resource owned by state manager, i.e. a
// GET on this URI will return the expected resource.
type ResourceURI struct {
	URI string `json:"URI"`
}

///////////////////////////////////////////////////////////////////////////
//
// HMS RedfishEndpoints
//
///////////////////////////////////////////////////////////////////////////

// Representation of a RedfishEndpoint, a network endpoint running a
// Redfish entry point
type RedfishEndpoint struct {
	// Embedded struct
	rf.RedfishEPDescription

	ComponentEndpoints []*ComponentEndpoint `json:"ComponentEndpoints,omitempty"`
	ServiceEndpoints   []*ServiceEndpoint   `json:"ServiceEndpoints,omitempty"`
}

// RedfishEndpointPatch is just rf.RedfishEPDescription but everything is a pointer.
type RedfishEndpointPatch struct {
	ID             *string `json:"ID"`
	Type           *string `json:"Type"`
	Name           *string `json:"Name"`
	Hostname       *string `json:"Hostname"`
	Domain         *string `json:"Domain"`
	FQDN           *string `json:"FQDN"`
	Enabled        *bool   `json:"Enabled"`
	UUID           *string `json:"UUID"`
	User           *string `json:"User"`
	Password       *string `json:"Password"`
	UseSSDP        *bool   `json:"UseSSDP"`
	MACRequired    *bool   `json:"MACRequired"`
	MACAddr        *string `json:"MACAddr"`
	IPAddr         *string `json:"IPAddress"`
	RediscOnUpdate *bool   `json:"RediscoverOnUpdate"`
	TemplateID     *string `json:"TemplateID"`
}

// A collection of 0-n RedfishEndpoints.  It could just be an ordinary
// array but we want to save the option to have indentifying info, etc.
// packaged with it, e.g. the query parameters or options that produced it,
// especially if there are fewer fields than normal being included.
type RedfishEndpointArray struct {
	RedfishEndpoints []*RedfishEndpoint `json:"RedfishEndpoints"`
}

// This wraps basic RedfishEndpointDescription data with the structure
// used for query responses.
func NewRedfishEndpoint(epd *rf.RedfishEPDescription) *RedfishEndpoint {
	ep := new(RedfishEndpoint)
	ep.RedfishEPDescription = *epd

	return ep
}

// From a given RedfishEndpointArray, return a ResourceURI array using the
// endpoint IDs appended to the base path uriBase (which should NOT end in
// '/').
func (eps *RedfishEndpointArray) GetResourceURIArray(uriBase string) []*ResourceURI {
	uris := make([]*ResourceURI, 0, 1)
	for _, ep := range eps.RedfishEndpoints {
		uri := new(ResourceURI)
		uri.URI = uriBase + "/" + ep.ID
		uris = append(uris, uri)
	}
	return uris
}

///////////////////////////////////////////////////////////////////////////
//
// HMS ComponentEndpoints
//
///////////////////////////////////////////////////////////////////////////

type ComponentEndpoint struct {
	// Embedded struct
	rf.ComponentDescription

	Enabled        bool   `json:"Enabled"`
	RfEndpointFQDN string `json:"RedfishEndpointFQDN"`
	URL            string `json:"RedfishURL"`

	// This is used as a descriminator to determine the type of *Info
	// struct that will be included below.
	ComponentEndpointType string `json:"ComponentEndpointType"`

	// These are all stored in the same JSON blob, only one of these
	// one of these will be set, based on the value of RedfishType,
	// and will match the ComponentEndpointType
	RedfishChassisInfo *rf.ComponentChassisInfo `json:"RedfishChassisInfo,omitempty"`
	RedfishSystemInfo  *rf.ComponentSystemInfo  `json:"RedfishSystemInfo,omitempty"`
	RedfishManagerInfo *rf.ComponentManagerInfo `json:"RedfishManagerInfo,omitempty"`
	RedfishPDUInfo     *rf.ComponentPDUInfo     `json:"RedfishPDUInfo,omitempty"`
	RedfishOutletInfo  *rf.ComponentOutletInfo  `json:"RedfishOutletInfo,omitempty"`
}

// Valid values for ComponentEndpointType discriminator field above.
const (
	CompEPTypeChassis = "ComponentEndpointChassis"
	CompEPTypeSystem  = "ComponentEndpointComputerSystem"
	CompEPTypeManager = "ComponentEndpointManager"
	CompEPTypePDU     = "ComponentEndpointPowerDistribution"
	CompEPTypeOutlet  = "ComponentEndpointOutlet"
)

// A collection of 0-n ComponentEndpoints.  It could just be an ordinary
// array but we want to save the option to have indentifying info, etc.
// packaged with it, e.g. the query parameters or options that produced it,
// especially if there are fewer fields than normal being included.
type ComponentEndpointArray struct {
	ComponentEndpoints []*ComponentEndpoint `json:"ComponentEndpoints"`
}

////////////////////////////////////////////////////////////////////////////
//
// Encode and decode component info to and from JSON blobs for schemaless
// storage in data store.
//
////////////////////////////////////////////////////////////////////////////

// This routine takes raw ComponentEndpoint type-specific extended info
// captured as free-form JSON (e.g. from a schema-free database field) and
// unmarshals it into the correct struct for the type with the proper
// RF type-specific name.
//
// NOTEs: The location info should be that produced by EncodeComponentInfo.
//        MODIFIES caller.
func (cep *ComponentEndpoint) DecodeComponentInfo(infoJSON []byte) error {
	var err error

	switch cep.RedfishType {
	// HWInv based on Redfish "Chassis" Type.  Identical structs (for now).
	case rf.ChassisType:
		chassisInfo := new(rf.ComponentChassisInfo)
		err = json.Unmarshal(infoJSON, chassisInfo)
		cep.RedfishChassisInfo = chassisInfo
		cep.ComponentEndpointType = CompEPTypeChassis
	case rf.ComputerSystemType:
		systemInfo := new(rf.ComponentSystemInfo)
		err = json.Unmarshal(infoJSON, systemInfo)
		cep.RedfishSystemInfo = systemInfo
		cep.ComponentEndpointType = CompEPTypeSystem
	case rf.ManagerType:
		managerInfo := new(rf.ComponentManagerInfo)
		err = json.Unmarshal(infoJSON, managerInfo)
		cep.RedfishManagerInfo = managerInfo
		cep.ComponentEndpointType = CompEPTypeManager
	case rf.PDUType:
		pduInfo := new(rf.ComponentPDUInfo)
		err = json.Unmarshal(infoJSON, pduInfo)
		cep.RedfishPDUInfo = pduInfo
		cep.ComponentEndpointType = CompEPTypePDU
	case rf.OutletType:
		outInfo := new(rf.ComponentOutletInfo)
		err = json.Unmarshal(infoJSON, outInfo)
		cep.RedfishOutletInfo = outInfo
		cep.ComponentEndpointType = CompEPTypeOutlet
	default:
		err = base.ErrHMSTypeUnsupported
	}
	return err
}

// Takes ComponentEndpoint type-specific extended info and converts it into
// raw JSON, e.g. to store as a generic JSON blob.
func (cep *ComponentEndpoint) EncodeComponentInfo() ([]byte, error) {
	var err error
	var infoJSON []byte

	switch cep.RedfishType {
	// Based on Redfish "Chassis" Type.
	case rf.ChassisType:
		infoJSON, err = json.Marshal(cep.RedfishChassisInfo)
	case rf.ComputerSystemType:
		infoJSON, err = json.Marshal(cep.RedfishSystemInfo)
	case rf.ManagerType:
		infoJSON, err = json.Marshal(cep.RedfishManagerInfo)
	case rf.PDUType:
		infoJSON, err = json.Marshal(cep.RedfishPDUInfo)
	case rf.OutletType:
		infoJSON, err = json.Marshal(cep.RedfishOutletInfo)
	default:
		// Not supported for this type.
		err = base.ErrHMSTypeUnsupported
	}
	return infoJSON, err
}

type ServiceEndpoint struct {
	// Embedded struct
	rf.ServiceDescription

	// These are read-only, derived from associated RfEndpointId in
	// rf.ServiceDescription
	RfEndpointFQDN string `json:"RedfishEndpointFQDN"`
	URL            string `json:"RedfishURL"`

	// These are all stored in the same JSON blob, only one of these
	// one of these will be set, based on the value of RedfishType
	ServiceInfo json.RawMessage `json:"ServiceInfo,omitempty"`
}

// A collection of 0-n ComponentEndpoints.  It could just be an ordinary
// array but we want to save the option to have indentifying info, etc.
// packaged with it, e.g. the query parameters or options that produced it,
// especially if there are fewer fields than normal being included.
type ServiceEndpointArray struct {
	ServiceEndpoints []*ServiceEndpoint `json:"ServiceEndpoints"`
}
