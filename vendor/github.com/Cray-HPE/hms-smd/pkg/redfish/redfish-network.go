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

// Redfish pass-through from Redfish "NetworkAdapter"
// This is the set of Redfish fields for this object that HMS understands
// and/or finds useful.  Those assigned to either the *LocationInfo
// or *FRUInfo subfields constitute the type specific fields in the
// HWInventory objects that are returned in response to queries.
type NetworkAdapter struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	NALocationInfoRF
	NAFRUInfoRF

	Actions                *NAActions     `json:"Actions,omitempty"`
	Controllers            []NAController `json:"Controllers,omitempty"`
	NetworkDeviceFunctions ResourceID     `json:"NetworkDeviceFunctions"`
	NetworkPorts           ResourceID     `json:"NetworkPorts"`
	Status                 *StatusRF      `json:"Status,omitempty"`
}

// Location-specific Redfish properties to be stored in hardware inventory
// These are only relevant to the currently installed location of the FRU
type NALocationInfoRF struct {
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

// Durable Redfish properties to be stored in hardware inventory as
// a specific FRU, which is then link with it's current location
// i.e. an x-name.  These properties should follow the hardware and
// allow it to be tracked even when it is removed from the system.
type NAFRUInfoRF struct {
	Manufacturer string `json:"Manufacturer"`
	Model        string `json:"Model"`
	PartNumber   string `json:"PartNumber"`
	SKU          string `json:"SKU,omitempty"`
	SerialNumber string `json:"SerialNumber"`
}

// Redfish NetworkAdapter sub-struct - Actions
type NAActions struct {
	NetworkAdapterResetToDefault NAActionResetToDefault `json:"#NetworkAdapter.ResetSettingsToDefault"`
}

// Redfish NetworkAdapter sub-struct - ResetToDefault
type NAActionResetToDefault struct {
	Target string `json:"target"`
	Title  string `json:"title,omitempty"`
}

// Redfish NetworkAdapter sub-struct - Controller
type NAController struct {
	ControllerCapabilities NAControllerCapabilities `json:"ControllerCapabilities"`
	FirmwarePackageVersion string                   `json:"FirmwarePackageVersion,omitempty"`
	Links                  NAControllerLinks        `json:"Links"`
}

// Redfish NetworkAdapter sub-struct - ControllerCapabilities
type NAControllerCapabilities struct {
	DataCenterBridging         NADataCenterBridging    `json:"DataCenterBridging"`
	NPIV                       NANPIV                  `json:"NPIV"`
	NetworkDeviceFunctionCount int                     `json:"NetworkDeviceFunctionCount,omitempty"`
	NetworkPortCount           int                     `json:"NetworkPortCount,omitempty"`
	VirtualizationOffload      NAVirtualizationOffload `json:"VirtualizationOffload"`
}

// Redfish NetworkAdapter sub-struct - ControllerLinks
type NAControllerLinks struct {
	NetworkDeviceFunctions     []ResourceID `json:"NetworkDeviceFunctions"`
	NetworkDeviceFunctionCount int          `json:"NetworkDeviceFunctions@odata.count,omitempty"`
	NetworkPorts               []ResourceID `json:"NetworkPorts"`
	NetworkPortCount           int          `json:"NetworkPorts@odata.count,omitempty"`
	PCIeDevices                []ResourceID `json:"PCIeDevices"`
	PCIeDeviceCount            int          `json:"PCIeDevices@odata.count,omitempty"`
}

// Redfish NetworkAdapter sub-struct - DataCenterBridging
type NADataCenterBridging struct {
	Capable bool `json:"Capable,omitempty"`
}

// Redfish NetworkAdapter sub-struct - NPIV
type NANPIV struct {
	MaxDeviceLogins int `json:"MaxDeviceLogins,omitempty"`
	MaxPortLogins   int `json:"MaxPortLogins,omitempty"`
}

// Redfish NetworkAdapter sub-struct - VirtualizationOffload
type NAVirtualizationOffload struct {
	SRIOV           NASRIOV           `json:"SRIOV"`
	VirtualFunction NAVirtualFunction `json:"VirtualFunction"`
}

// Redfish NetworkAdapter sub-struct - SRIOV
type NASRIOV struct {
	SRIOVVEPACapable bool `json:"SRIOVVEPACapable,omitempty"`
}

// Redfish NetworkAdapter sub-struct - VirtualFunction
type NAVirtualFunction struct {
	DeviceMaxCount         int `json:"DeviceMaxCount,omitempty"`
	MinAssignmentGroupSize int `json:"MinAssignmentGroupSize,omitempty"`
	NetworkPortMaxCount    int `json:"NetworkPortMaxCount,omitempty"`
}
