// MIT License
//
// (C) Copyright [2021] Hewlett Packard Enterprise Development LP
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

type ChassisOemHpe struct {
	// Actions
	// BayNumber
	// Firmware
	Links HpeOemChassisLinks `json:"Links"`
	// SystemMaintenanceSwitches
}

type HpeOemChassisLinks struct {
	Devices ResourceID `json:"Devices"`
}

type HpeDeviceCollection GenericCollection

// Redfish pass-through from Redfish HPE OEM Device
// This is the set of Redfish fields for this object that HMS understands
// and/or finds useful.  Those assigned to either the *LocationInfo
// or *FRUInfo subfields consitiute the type specific fields in the
// HWInventory objects that are returned in response to queries.
type HpeDevice struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	HpeDeviceLocationInfoRF
	HpeDeviceFRUInfoRF

	Status StatusRF `json:"Status"`
}

// Location-specific Redfish properties to be stored in hardware inventory
// These are only relevant to the currently installed location of the FRU
type HpeDeviceLocationInfoRF struct {
	// Redfish pass-through from rf.Processor
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Location    string `json:"Location"`
}

// Durable Redfish properties to be stored in hardware inventory as
// a specific FRU, which is then link with it's current location
// i.e. an x-name.  These properties should follow the hardware and
// allow it to be tracked even when it is removed from the system.
type HpeDeviceFRUInfoRF struct {
	// Redfish pass-through from rf.HpeDevice
	Manufacturer      string `json:"Manufacturer"`
	Model             string `json:"Model"`
	SerialNumber      string `json:"SerialNumber"`
	PartNumber        string `json:"PartNumber"`
	DeviceType        string `json:"DeviceType"`
	ProductPartNumber string `json:"ProductPartNumber"`
	ProductVersion    string `json:"ProductVersion"`
}
