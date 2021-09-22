// MIT License
//
// (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
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
package rf

// JSON decoded struct returned from Redfish for a particular set of
// ids.  Assembly resources deviate from GenericCollections by
// by using the Assemblies array instead of a Members array
// Example: /redfish/v1/Chassis/<system_id>/Assembly
type Assembly struct {
	OContext    string            `json:"@odata.context"`
	Oid         string            `json:"@odata.id"`
	Otype       string            `json:"@odata.type"`
	Description string            `json:"Description"`
	Name        string            `json:"Name"`
	Id          string            `json:"Id"`
	Assemblies  []*NodeAccelRiser `json:"Assemblies"`
}

// Redfish pass-through from Redfish "Assembly"
// This is the set of Redfish fields for this object that HMS understands
// and/or finds useful.  Those assigned to either the *LocationInfo
// or *FRUInfo subfields constitute the type specific fields in the
// HWInventory objects that are returned in response to queries.
type NodeAccelRiser struct {
	Oid string `json:"@odata.id"`

	// Embedded structs.
	NodeAccelRiserLocationInfoRF
	NodeAccelRiserFRUInfoRF

	Status StatusRF `json:"Status"`
}

// Location-specific Redfish properties to be stored in hardware inventory
// These are only relevant to the currently installed location of the FRU
// TODO: How to version these (as HMS structures).
type NodeAccelRiserLocationInfoRF struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

// Durable Redfish properties to be stored in hardware inventory as
// a specific FRU, which is then link with it's current location
// i.e. an x-name.  These properties should follow the hardware and
// allow it to be tracked even when it is removed from the system.
type NodeAccelRiserFRUInfoRF struct {
	//Manufacturer Info
	PhysicalContext        string             `json:"PhysicalContext"`
	Producer               string             `json:"Producer"`
	SerialNumber           string             `json:"SerialNumber"`
	PartNumber             string             `json:"PartNumber"`
	Model                  string             `json:"Model"`
	ProductionDate         string             `json:"ProductionDate"`
	Version                string             `json:"Version"`
	EngineeringChangeLevel string             `json:"EngineeringChangeLevel"`
	OEM                    *NodeAccelRiserOEM `json:"Oem,omitempty"`
}

type NodeAccelRiserOEM struct {
	PCBSerialNumber string `json:"PCBSerialNumber"`
}
