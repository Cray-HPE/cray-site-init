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

import (
	"encoding/json"
)

// Allowable values for "PowerState" fields.
const (
	POWER_STATE_ON           = "On"
	POWER_STATE_OFF          = "Off"
	POWER_STATE_POWERING_ON  = "PoweringOn"
	POWER_STATE_POWERING_OFF = "PoweringOff"
)

// Resource link, often found in collection array or "Links" section.
// Example: {
//                "@odata.id": "/redfish/v1/Systems/System.Embedded.1"
//          }
type ResourceID struct {
	Oid string `json:"@odata.id"`
}

// Redfish "Health", "HealthRollUp" enum
type HealthRF string

// Redfish "State" Enum
type StateRF string

// Status struct, used in many places.
type StatusRF struct {
	Health       HealthRF `json:"Health"`
	HealthRollUp HealthRF `json:"HealthRollUp,omitempty"`
	State        StateRF  `json:"State,omitempty"`
}

// JSON decoded struct returned from Redfish for a particular set of
// ids.  Many collection resources are essentially identical in terms
// of their fields, so we define it generically here.
type GenericCollection struct {
	OContext      string       `json:"@odata.context"`
	OCount        int          `json:"@odata.count"` // Oldest schemas use
	Oid           string       `json:"@odata.id"`
	Otype         string       `json:"@odata.type"`
	Description   string       `json:"Description"`
	Members       []ResourceID `json:"Members"`
	Outlets       []ResourceID `json:"Outlets"`             // For HPE PDU Outlets
	MembersOCount int          `json:"Members@odata.count"` // Most schemas
	Name          string       `json:"Name"`
}

// JSON decoded collection struct returned from Redfish "Systems"
// Example: /redfish/v1/Systems
type SystemCollection GenericCollection

// JSON decoded collection struct returned from Redfish "Chassis"
// Example: /redfish/v1/Chassis
type ChassisCollection GenericCollection

// JSON decoded collection struct returned from Redfish "Manager"
// Example: /redfish/v1/Managers
type ManagerCollection GenericCollection

// JSON decoded Account collection struct linked from Redfish "AccountService"
// Example: /redfish/v1/AccountService/Accounts
type ManagerAccountCollection GenericCollection

// JSON decoded Role collection struct linked from Redfish "AccountService"
// Example: /redfish/v1/AccountService/Roles
type RoleCollection GenericCollection

// JSON decoded event subscription struct linked from Redfish "EventService"
// Example: /redfish/v1/EventService/Subscriptions
type EventDestinationCollection GenericCollection

// JSON decoded collection struct returned from Redfish "SessionService"
// There is also a pointer in the system root under "Links"
// Example: /redfish/v1/SessionService/Sessions
type SessionCollection GenericCollection

// JSON decoded task subscription struct linked from Redfish "TaskService"
// Example: /redfish/v1/TaskService/Subscriptions
type TaskCollection GenericCollection

// JSON decoded collection struct returned from Redfish "Processors"
// Example: /redfish/v1/Systems/<system_id>/Processors
type ProcessorCollection GenericCollection

// JSON decoded collection struct returned from Redfish "Memory"
// Example: /redfish/v1/Systems/<system_id>/Memory
type MemoryCollection GenericCollection

// JSON decoded collection struct returned from Redfish "SimpleStorage"
// Example: /redfish/v1/Systems/<system_id>/SimpleStorage
type SimpleStorageCollection GenericCollection

// JSON decoded collection struct returned from Redfish "Storage"
// Example: /redfish/v1/Systems/<system_id>/Storage
type Storage GenericCollection

// JSON decoded collection struct of Redfish type "EthernetInterfaceCollection"
// Examples: /redfish/v1/Systems/<system_id>/EthernetInterfaces
//           /redfish/v1/Managers/<manager_id>/EthernetInterfaces
type EthernetInterfaceCollection GenericCollection

// JSON decoded collection struct of Redfish type "SerialInterfaceCollection"
// Example: /redfish/v1/Managers/<manager_id>/SerialInterfaces
type SerialInterfaceCollection GenericCollection

// JSON decoded collection struct returned from Redfish "NetworkAdapter"
// Example: /redfish/v1/Chassis/<chassis_id>/NetworkAdapters
type NetworkAdapterCollection GenericCollection

// JSON decoded collection struct returned from Redfish "Controls"
// Example: /redfish/v1/Chassis/<chassis_id>/Controls
type ControlCollection GenericCollection

type ServiceRoot struct {
	OContext       string `json:"@odata.context"`
	Oid            string `json:"@odata.id"`
	Otype          string `json:"@odata.type"`
	Name           string `json:"Name"`
	Description    string `json:"Description"`
	RedfishVersion string `json:"RedfishVersion"`
	UUID           string `json:"UUID"`

	Systems        ResourceID `json:"Systems"`
	Chassis        ResourceID `json:"Chassis"`
	Managers       ResourceID `json:"Managers"`
	Tasks          ResourceID `json:"Tasks"` // Actually points to "TaskService"
	SessionService ResourceID `json:"SessionService"`
	AccountService ResourceID `json:"AccountService"`
	EventService   ResourceID `json:"EventService"`
	UpdateService  ResourceID `json:"UpdateService"`
	Registries     ResourceID `json:"Registries"`

	// TODO: Later stuff: StorageSystems, Fabrics, UpdateService, JsonSchemas

	// PDU stuff
	PowerEquipment    ResourceID `json:"PowerEquipment"`
	PowerDistribution ResourceID `json:"PowerDistribution"`

	Links ServiceRootLinks `json:"Links"`
}

// Redfish ServiceRoot - Links section
type ServiceRootLinks struct {
	Sessions ResourceID `json:"Sessions"`
}

// JSON decoded struct returned from an entry in the BMC "Managers" collection
//  Example: /redfish/v1/Managers/iDRAC.Embedded.1
type Manager struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	Actions *ManagerActions `json:"Actions,omitempty"`

	// Embedded structs.
	ManagerLocationInfoRF
	ManagerFRUInfoRF

	PowerState            string   `json:"PowerState"`
	ServiceEntryPointUUID string   `json:"ServiceEntryPointUUID"`
	UUID                  string   `json:"UUID"`
	Status                StatusRF `json:"Status"`

	// TODO: GraphicalConsole, SerialConsole, CommandShell

	EthernetInterfaces ResourceID `json:"EthernetInterfaces"`
	NetworkProtocol    ResourceID `json:"NetworkProtocol"`
	LogServices        ResourceID `json:"LogServices"`
	SerialInterfaces   ResourceID `json:"SerialInterfaces"`
	VirtualMedia       ResourceID `json:"VirtualMedia"`

	Links ManagerLinks `json:"Links"`
}

type ManagerLocationInfoRF struct {
	DateTime            string `json:"DateTime"`
	DateTimeLocalOffset string `json:"DateTimeLocalOffset"`
	Description         string `json:"Description"`
	FirmwareVersion     string `json:"FirmwareVersion"`
	Id                  string `json:"Id"`
	Name                string `json:"Name"`
}

type ManagerFRUInfoRF struct {
	ManagerType  string `json:"ManagerType"`
	Model        string `json:"Model"`
	Manufacturer string `json:"Manufacturer"`
	PartNumber   string `json:"PartNumber"`
	SerialNumber string `json:"SerialNumber"`
}

// RAParameter for the array of Reset actions
type RAParameter struct {
	Name            string   `json:"Name"`
	AllowableValues []string `json:"AllowableValues"`
	DataType        string   `json:"DataType"`
	Required        bool     `json:"Required"`
}

// ResetActionInfo contains a list of parameters for a Redfish ResetAction
type ResetActionInfo struct {
	ID           string        `json:"Id"`
	RAParameters []RAParameter `json:"Parameters"`
	Name         string        `json:"Name"`
}

// Action type Reset - May be found under Chassis, System or Manager Actions
type ActionReset struct {
	AllowableValues []string `json:"ResetType@Redfish.AllowableValues"`
	RFActionInfo    string   `json:"@Redfish.ActionInfo"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}

// FactoryReset - OEM (Cray) only so far
type ActionFactoryReset struct {
	AllowableValues []string `json:"FactoryResetType@Redfish.AllowableValues"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}

// Action "Name" - OEM (Cray) only so far
type ActionNamed struct {
	AllowableValues []string `json:"Name@Redfish.AllowableValues"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}

// Redfish Manager sub-struct - Actions
type ManagerActions struct {
	ManagerReset ActionReset        `json:"#Manager.Reset"`
	OEM          *ManagerActionsOEM `json:"Oem,omitempty"`
}

// Redfish Manager Actions - OEM sub-struct
type ManagerActionsOEM struct {
	ManagerFactoryReset *ActionFactoryReset `json:"#Manager.FactoryReset,omitempty"`
	CrayProcessSchedule *ActionNamed        `json:"#CrayProcess.Schedule,omitempty"`
}

type ManagerLinks struct {
	ManagerForChassis []ResourceID `json:"ManagerForChassis"`
	ManagerInChassis  ResourceID   `json:"ManagerInChassis"`
	ManagerForServers []ResourceID `json:"ManagerForServers"`
}

// JSON decoded struct returned from the BMC of type "Chassis"
//  Example: /redfish/v1/Chassis/System.Embedded.1
type Chassis struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	Actions *ChassisActions `json:"Actions,omitempty"`

	ChassisLocationInfoRF
	ChassisFRUInfoRF

	PowerState string   `json:"PowerState"`
	Status     StatusRF `json:"Status"`

	NetworkAdapters ResourceID `json:"NetworkAdapters"`
	Power           ResourceID `json:"Power"`
	Assembly        ResourceID `json:"Assembly"`
	Thermal         ResourceID `json:"Thermal"`
	Controls        ResourceID `json:"Controls"`

	Links ChassisLinks `json:"Links"`

	OEM *ChassisOEM `json:"Oem,omitempty"`
}

type ChassisOEM struct {
	Hpe *ChassisOemHpe `json:"Hpe,omitempty"`
}

// Redfish Actions for Chassis components
type ChassisActions struct {
	ChassisReset ActionReset        `json:"#Chassis.Reset"`
	OEM          *ChassisActionsOEM `json:"Oem,omitempty"`
}

// Redfish Chassis Actions - OEM sub-struct
type ChassisActionsOEM struct {
	ChassisEmergencyPower *ActionReset `json:"#Chassis.EmergencyPower,omitempty"`
}

// Redfish Chassis - Links section
type ChassisLinks struct {
	ComputerSystems []ResourceID `json:"ComputerSystems"` // Nodes in chassis
	Contains        []ResourceID `json:"Contains"`        // Sub-chassis ids
	ContainedBy     ResourceID   `json:"ContainedBy"`     // Parent chassis
	CooledBy        []ResourceID `json:"CooledBy"`
	Drives          []ResourceID `json:"Drives"`
	ManagedBy       []ResourceID `json:"ManagedBy"`
	PoweredBy       []ResourceID `json:"PoweredBy"`
	Storage         []ResourceID `json:"Storage"`
	Switches        []ResourceID `json:"Switches"`
}

// Redfish when following the Power URI in a chassis
type PowerInfo struct {
	OEM          *OEMPowerInfo   `json:"Oem,omitempty"`
	PowerControl []*PowerControl `json:"PowerControl"`
}

type OEMPowerInfo struct {
	HPE *OEMPowerInfoHPE `json:"Hpe,omitempty"`
}

type OEMPowerInfoHPE struct {
	Links struct {
		AccPowerService ResourceID `json:"AccPowerService"`
	} `json:"Links"`
}

// Redfish when following the HPE AccPowerService URI
type HPEAccPowerService struct {
	Links struct {
		PowerLimit ResourceID `json:"PowerLimit"`
	} `json:"Links"`
	PowerRegulationEnabled bool `json:"PowerRegulationEnabled"`
}

// Redfish when following the HPE PowerLimit URI
type HPEPowerLimit struct {
	Actions struct {
		ConfigurePowerLimit struct {
			Target string `json:"target"`
		} `json:"#HpeServerAccPowerLimit.ConfigurePowerLimit"`
	} `json:"Actions"`
	PowerLimitRanges []struct {
		MaximumPowerLimit int `json:"MaximumPowerLimit"`
		MinimumPowerLimit int `json:"MinimumPowerLimit"`
	} `json:"PowerLimitRanges"`
	Name string `json:"Name"`
}

type RFControl struct {
	ControlDelaySeconds int      `json:"ControlDelaySeconds"`
	ControlMode         string   `json:"ControlMode"`
	ControlType         string   `json:"ControlType"`
	Id                  string   `json:"Id"`
	Name                string   `json:"Name"`
	PhysicalContext     string   `json:"PhysicalContext"`
	SetPoint            int      `json:"SetPoint"`
	SetPointUnits       string   `json:"SetPointUnits"`
	SettingRangeMax     int      `json:"SettingRangeMax"`
	SettingRangeMin     int      `json:"SettingRangeMin"`
	Status              StatusRF `json:"Status"`
}

// Location-specific Redfish properties to be stored in hardware inventory
// These are only relevant to the currently installed location of the FRU
// TODO: How to version these (as HMS structures).
type ChassisLocationInfoRF struct {
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Hostname    string `json:"HostName"`
}

// Durable Redfish properties to be stored in hardware inventory as
// a specific FRU, which is then link with it's current location
// i.e. an x-name.  These properties should follow the hardware and
// allow it to be tracked even when it is removed from the system.
// TODO: How to version these (as HMS structures)
type ChassisFRUInfoRF struct {
	AssetTag     string `json:"AssetTag"`
	ChassisType  string `json:"ChassisType"`
	Model        string `json:"Model"`
	Manufacturer string `json:"Manufacturer"`
	PartNumber   string `json:"PartNumber"`
	SerialNumber string `json:"SerialNumber"`
	SKU          string `json:"SKU"`
}

// Redfish pass-through from Redfish Processor
//   Example: /redfish/v1/Systems/System.Embedded.1
// This is the set of Redfish fields for this object that HMS understands
// and/or finds useful.  Those assigned to either the *LocationInfo
// or *FRUInfo subfields consitiute the type specific fields in the
// HWInventory objects that are returned in response to queries.
// JSON decoded struct returned from the BMC of type "ComputerSystem"
type ComputerSystem struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	// Embedded structs - used for SystemHWInventory.
	SystemLocationInfoRF // Location/instance specifc properties
	SystemFRUInfoRF      // FRU-specific data that follows the physical HW.

	Actions *ComputerSystemActions `json:"Actions,omitempty"`

	Boot ComputerSystemBoot `json:"Boot"`

	PowerState   string   `json:"PowerState"`
	IndicatorLED string   `json:"IndicatorLED"`
	Status       StatusRF `json:"Status"`

	Bios               ResourceID `json:"Bios"`
	EthernetInterfaces ResourceID `json:"EthernetInterfaces"`
	LogServices        ResourceID `json:"LogServices"`
	Memory             ResourceID `json:"Memory"`
	Processors         ResourceID `json:"Processors"`
	SecureBoot         ResourceID `json:"SecureBoot"`
	SimpleStorage      ResourceID `json:"SimpleStorage"`
	Storage            ResourceID `json:"Storage"`

	Links ComputerSystemLinks `json:"Links"`
}

// Redfish ComputerSystem sub-struct - Actions
type ComputerSystemActions struct {
	ComputerSystemReset ActionReset `json:"#ComputerSystem.Reset"`
}

// Redfsh ComputerSystem sub-struct - Boot options/settings/parameters
type ComputerSystemBoot struct {
	BootSourceOverrideEnabled    string   `json:"BootSourceOverrideEnabled"`
	BootSourceOverrideTarget     string   `json:"BootSourceOverrideTarget"`
	AllowableValues              []string `json:"BootSourceOverrideTarget@Redfish.AllowableValues"`
	UefiTargetBootSourceOverride string   `json:"UefiTargetBootSourceOverride"`
}

// Redfish Links struct - All those defined for ComputerSystem objects
type ComputerSystemLinks struct {
	Chassis   []ResourceID `json:"Chassis"`
	ManagedBy []ResourceID `json:"ManagedBy"`
	PoweredBy []ResourceID `json:"PoweredBy"`
}

// Redfish ProcessorSummary struct - Sub-struct of ComputerSystem
type ComputerSystemProcessorSummary struct {
	Count json.Number `json:"Count"`
	Model string      `json:"Model"`
	//Status StatusRF    `json:"Status"`
}

// Redfish MemorySummary struct - Sub-struct of ComputerSystem
type ComputerSystemMemorySummary struct {
	TotalSystemMemoryGiB json.Number `json:"TotalSystemMemoryGiB"`
	//Status               rf.StatusRF    `json:"Status"`
}

// Location-specific Redfish properties to be stored in hardware inventory
// These are only relevant to the currently installed location of the FRU
// TODO: How to version these (as HMS structures).
type SystemLocationInfoRF struct {
	// Redfish pass-through from Redfish ComputerSystem
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Hostname    string `json:"HostName"`

	ProcessorSummary ComputerSystemProcessorSummary `json:"ProcessorSummary"`

	MemorySummary ComputerSystemMemorySummary `json:"MemorySummary"`
}

// Durable Redfish properties to be stored in hardware inventory as
// a specific FRU, which is then (typically) associated with a location
// i.e. an x-name in HMS terms and the ProcessorLocationInfo fields
// in Redfish terms on the controller.  These properties should
// follow the hardware and allow it to be tracked even when it is removed
// from the system.
// TODO: How to version these (as HMS structures).
type SystemFRUInfoRF struct {
	// Redfish pass-through from Redfish ComputerSystem
	AssetTag     string `json:"AssetTag"`
	BiosVersion  string `json:"BiosVersion"`
	Model        string `json:"Model"`
	Manufacturer string `json:"Manufacturer"`
	PartNumber   string `json:"PartNumber"`
	SerialNumber string `json:"SerialNumber"`
	SKU          string `json:"SKU"`
	SystemType   string `json:"SystemType"`
	UUID         string `json:"UUID"`
}

// JSON decoded struct returned from Redfish of type "EthernetInterface"
// Example:
//   /redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-3-1
type EthernetInterface struct {
	OContext               string              `json:"@odata.context"`
	Oid                    string              `json:"@odata.id"`
	Otype                  string              `json:"@odata.type"`
	AutoNeg                *bool               `json:"AutoNeg,omitempty"`
	Description            string              `json:"Description"`
	FQDN                   string              `json:"FQDN"`
	FullDuplex             *bool               `json:"FullDuplex,omitempty"`
	Hostname               string              `json:"HostName"`
	Id                     string              `json:"Id"`
	IPv4Addresses          []IPv4Address       `json:"IPv4Addresses"`
	IPv6Addresses          []IPv6Address       `json:"IPv6Addresses"`
	IPv6StaticAddresses    []IPv6StaticAddress `json:"IPv6StaticAddresses"`
	IPv6DefaultGateway     string              `json:"IPv6DefaultGateway"`
	InterfaceEnabled       *bool               `json:"InterfaceEnabled,omitempty"`
	MACAddress             string              `json:"MACAddress"`
	PermanentMACAddress    string              `json:"PermanentMACAddress"`
	MTUSize                json.Number         `json:"MTUSize"`
	MaxIPv6StaticAddresses json.Number         `json:"MaxIPv6StaticAddresses"`
	Name                   string              `json:"Name"`
	NameServers            []string            `json:"NameServers"`
	SpeedMbps              json.Number         `json:"SpeedMbps"`
	Status                 StatusRF            `json:"Status"`
	UefiDevicePath         string              `json:"UefiDevicePath"`
	VLAN                   VLAN                `json:"VLAN"`
}

// An IPv4 address, e.g. as found in EthernetInterface
type IPv4Address struct {
	Address       string `json:"Address"`
	AddressOrigin string `json:"AddressOrigin"`
	Gateway       string `json:"Gateway"`
	SubnetMask    string `json:"SubnetMask"`
}

// An IPv6 address, e.g. as found in EthernetInterface
type IPv6Address struct {
	Address       string      `json:"Address"`
	AddressOrigin string      `json:"AddressOrigin"`
	AddressState  string      `json:"AddressState"`
	PrefixLength  json.Number `json:"PrefixLength"`
}

// A static IPv4 address, e.g. as found in EthernetInterface
type IPv6StaticAddress struct {
	Address      string      `json:"Address"`
	PrefixLength json.Number `json:"PrefixLength"`
}

// A VLAN definition, e.g. as found in EthernetInterface.  This will
// not be used in the case where multiple VLANs exist.  An entry in
// the Links section will list these in the latter case.
type VLAN struct {
	VLANEnable *bool       `json:"VLANEnable,omitempty"`
	VLANid     json.Number `json:"VLANid"`
}

// Redfish pass-through from Redfish "Processor"
// This is the set of Redfish fields for this object that HMS understands
// and/or finds useful.  Those assigned to either the *LocationInfo
// or *FRUInfo subfields consitiute the type specific fields in the
// HWInventory objects that are returned in response to queries.
type Processor struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	// Embedded structs.
	ProcessorLocationInfoRF
	ProcessorFRUInfoRF

	Status StatusRF `json:"Status"`
}

type ProcessorIdRF struct {
	EffectiveFamily         string `json:"EffectiveFamily"`
	EffectiveModel          string `json:"EffectiveModel"`
	IdentificationRegisters string `json:"IdentificationRegisters"`
	MicrocodeInfo           string `json:"MicrocodeInfo"`
	Step                    string `json:"Step"`
	VendorID                string `json:"VendorID"`
}

// Location-specific Redfish properties to be stored in hardware inventory
// These are only relevant to the currently installed location of the FRU
// TODO: How to version these (as HMS structures).
type ProcessorLocationInfoRF struct {
	// Redfish pass-through from rf.Processor
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Socket      string `json:"Socket"`
}

// Durable Redfish properties to be stored in hardware inventory as
// a specific FRU, which is then link with it's current location
// i.e. an x-name.  These properties should follow the hardware and
// allow it to be tracked even when it is removed from the system.
// TODO: How to version these (as HMS structures)
type ProcessorFRUInfoRF struct {
	// Redfish pass-through from rf.Processor
	InstructionSet        string        `json:"InstructionSet"`
	Manufacturer          string        `json:"Manufacturer"`
	MaxSpeedMHz           json.Number   `json:"MaxSpeedMHz"`
	Model                 string        `json:"Model"`
	SerialNumber          string        `json:"SerialNumber"`
	PartNumber            string        `json:"PartNumber"`
	ProcessorArchitecture string        `json:"ProcessorArchitecture"`
	ProcessorId           ProcessorIdRF `json:"ProcessorId"`
	ProcessorType         string        `json:"ProcessorType"`
	TotalCores            json.Number   `json:"TotalCores"`
	TotalThreads          json.Number   `json:"TotalThreads"`
	Oem                   *ProcessorOEM `json:"Oem"`
}

type ProcessorOEM struct {
	GBTProcessorOemProperty *GBTProcessorOem `json:"GBTProcessorOemProperty,omitempty"`
}

type GBTProcessorOem struct {
	ProcessorSerialNumber string `json:"Processor Serial Number,omitempty"`
}

// Redfish pass-through from rf.Memory
// This is the set of Redfish fields for this object that HMS understands
// and/or finds useful.  Those assigned to either the *LocationInfo
// or *FRUInfo subfields consitiute the type specific fields in the
// HWInventory objects that are returned in response to queries.
type Memory struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	MemoryLocationInfoRF
	MemoryFRUInfoRF

	Status StatusRF `json:"Status"`
}

type MemoryLocationRF struct {
	Socket           json.Number `json:"Socket"`
	MemoryController json.Number `json:"MemoryController"`
	Channel          json.Number `json:"Channel"`
	Slot             json.Number `json:"Slot"`
}

// Location-specific Redfish properties to be stored in hardware inventory
// These are only relevant to the currently installed location of the FRU
// TODO: How to version these (as HMS structures)
type MemoryLocationInfoRF struct {
	// Redfish pass-through from rf.Memory
	Id             string           `json:"Id"`
	Name           string           `json:"Name"`
	Description    string           `json:"Description"`
	MemoryLocation MemoryLocationRF `json:"MemoryLocation"`
}

// Durable Redfish properties to be stored in hardware inventory as
// a specific FRU, which is then link with it's current location
// i.e. an x-name.  These properties should follow the hardware and
// allow it to be tracked even when it is removed from the system.
// TODO: How to version these (as HMS structures)
type MemoryFRUInfoRF struct {
	// Redfish pass-through from rf.Memory
	BaseModuleType    string      `json:"BaseModuleType,omitempty"`
	BusWidthBits      json.Number `json:"BusWidthBits,omitempty"`
	CapacityMiB       json.Number `json:"CapacityMiB"`
	DataWidthBits     json.Number `json:"DataWidthBits,omitempty"`
	ErrorCorrection   string      `json:"ErrorCorrection,omitempty"`
	Manufacturer      string      `json:"Manufacturer,omitempty"`
	MemoryType        string      `json:"MemoryType,omitempty"`
	MemoryDeviceType  string      `json:"MemoryDeviceType,omitempty"`
	OperatingSpeedMhz json.Number `json:"OperatingSpeedMhz"`
	PartNumber        string      `json:"PartNumber,omitempty"`
	RankCount         json.Number `json:"RankCount,omitempty"`
	SerialNumber      string      `json:"SerialNumber"`
}

// Redfish account service.  This is the top-level object linked via the
// service root.  It configures general account parameters and links
// to individual accounts and user roles.
type AccountService struct {
	OContext                        string      `json:"@odata.context"`
	Oid                             string      `json:"@odata.id"`
	Otype                           string      `json:"@odata.type"`
	Id                              string      `json:"Id"`
	Name                            string      `json:"Name"`
	Description                     string      `json:"Description"`
	Status                          StatusRF    `json:"Status"`
	ServiceEnabled                  *bool       `json:"ServiceEnabled,omitempty"`
	AuthFailureLoggingThreshold     json.Number `json:"AuthFailureLoggingThreshold"`
	MinPasswordLength               json.Number `json:"MinPasswordLength"`
	AccountLockoutThreshold         json.Number `json:"AccountLockoutThreshold"`
	AccountLockoutDuration          json.Number `json:"AccountLockoutDuration"`
	AccountLockoutCounterResetAfter json.Number `json:"AccountLockoutCounterResetAfter"`

	Accounts ResourceID `json:"Accounts"`
	Roles    ResourceID `json:"Roles"`
}

// An account on a Redfish endpoint.  Needs to link to a valid Role, e.g.
// administrator, read-only, etc.  Note password is always blank on GET,
// this is only used when creating or updating an account.
type ManagerAccount struct {
	OContext    string `json:"@odata.context"`
	Oid         string `json:"@odata.id"`
	Otype       string `json:"@odata.type"`
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Enabled     *bool  `json:"Enabled,omitempty"`
	UserName    string `json:"UserName"`
	Password    string `json:"Password"`
	RoleId      string `json:"RoleId"`
	Locked      *bool  `json:"Locked,omitempty"`

	Links ManagerAccountLinks `json:"Links"`
}

type ManagerAccountLinks struct {
	Role ResourceID `json:"Role"`
}

// A type of account on a Redfish endpoint.  Maybe be "built-in" or some
// custom type defined by the administrator.  Each ManagerAccount is linked
// to a role.
type Role struct {
	OContext           string   `json:"@odata.context"`
	Oid                string   `json:"@odata.id"`
	Otype              string   `json:"@odata.type"`
	Id                 string   `json:"Id"`
	Name               string   `json:"Name"`
	Description        string   `json:"Description"`
	IsPredefined       *bool    `json:"IsPredefined,omitempty"`
	AssignedPrivileges []string `json:"AssignedPrivileges"`
	OemPrivileges      []string `json:"OemPrivileges"`
}

// Redfish event service.  This is the top-level object linked via the
// service root.  It gives information of the types of subscriptions
// available and the URIs to create them.  It also links to the current
// set of subscriptions, each primarily defined by the type(s) of
// event/alert of interest and the URL to POST them to when they occur.
type EventService struct {
	OContext                        string      `json:"@odata.context"`
	Oid                             string      `json:"@odata.id"`
	Otype                           string      `json:"@odata.type"`
	Id                              string      `json:"Id"`
	Name                            string      `json:"Name"`
	Status                          StatusRF    `json:"Status"`
	ServiceEnabled                  *bool       `json:"ServiceEnabled,omitempty"`
	DeliveryRetryAttempts           json.Number `json:"DeliveryRetryAttempts"`
	DeliveryRetryIntervalInSeconds  json.Number `json:"DeliveryRetryIntervalInSeconds"`
	EventTypesForSubscription       []string    `json:"EventTypesForSubscription"`
	EventTypesForSubscriptionOCount json.Number `json:"EventTypesForSubscription@odata.count"`

	Subscriptions ResourceID `json:"Subscriptions"`

	Actions EventServiceActions `json:"Actions"`
}

// EventService - Actions defined by Redfish for this type
type EventServiceActions struct {
	SubmitTestEvent ActionTestEvent `json:"#EventService.SubmitTestEvent"`
}

// EventService - Submit Test Event action payload.
type ActionTestEvent struct {
	AllowableValues []string `json:"EventType@Redfish.AllowableValues"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}

// Redfish object representing a current event subscription being carried
// out by a Redfish manager/endpoint.
type EventDestination struct {
	OContext    string   `json:"@odata.context"`
	Oid         string   `json:"@odata.id"`
	Otype       string   `json:"@odata.type"`
	Id          string   `json:"Id"`
	Context     string   `json:"Context"`
	Destination string   `json:"Destination"`
	EventTypes  []string `json:"EventTypes"`
	Protocol    string   `json:"Protocol"`
}

// An individual event record.  Multiple EventRecords can be contained in the
// Event object that is actually POSTed to the subcriber.
//
// Note that "Message" is just a simple string and is optional. Instead,
// it may be necessary to use "MessageId" to look up a "Message" object in the
// corresponding Message Registry.  In any case, there may be greater detail
// in the Message entry as it is more than just a simple string, e.g. it
// will place the MessageArgs in the context of the particular message.
// The MessageId format is registry.version.message.
// See "Message" below for details.
type EventRecord struct {
	EventType         string     `json:"EventType"`
	EventId           string     `json:"EventId"`
	EventTimestamp    string     `json:"EventTimestamp"`
	Severity          string     `json:"Severity"`
	Message           string     `json:"Message"`
	MessageId         string     `json:"MessageId"`
	MessageArgs       []string   `json:"MessageArgs"`
	Context           string     `json:"Context"` // Older versions
	OriginOfCondition ResourceID `json:"OriginOfCondition"`
	// TODO: Actions - How structured?  OEM only?
}

// Redfish event.  Can contain multiple EventRecord structs in Events
// array.
type Event struct {
	OContext     string        `json:"@odata.context"`
	Oid          string        `json:"@odata.id"`
	Otype        string        `json:"@odata.type"`
	Id           string        `json:"Id"`
	Name         string        `json:"Name"`
	Context      string        `json:"Context"` // Later versions
	Description  string        `json:"Description"`
	Events       []EventRecord `json:"Events"`
	EventsOCount int           `json:"Events@odata.count"`
	// TODO: Actions - How structured?  OEM only?
}

// Redfish session service.  This is the top-level object linked via the
// service root.  It provides information about the session service and
// links to the set of active sessions.  Note that there should also
// be a direct link from the service root as it is the only part of
// the root that is available without authentication.
type SessionService struct {
	OContext       string      `json:"@odata.context"`
	Oid            string      `json:"@odata.id"`
	Otype          string      `json:"@odata.type"`
	Id             string      `json:"Id"`
	Name           string      `json:"Name"`
	Status         StatusRF    `json:"Status"`
	ServiceEnabled *bool       `json:"ServiceEnabled,omitempty"`
	SessionTimeout json.Number `json:"SessionTimeout"`

	Sessions ResourceID `json:"Sessions"`
}

// A session that has been created via the session service.
type Session struct {
	OContext    string `json:"@odata.context"`
	Oid         string `json:"@odata.id"`
	Otype       string `json:"@odata.type"`
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	UserName    string `json:"UserName"`
}

// Redfish task service.  This is the top-level object linked via the
// service root.  It provides information about the task service and
// links to the set of active tasks.
type TaskService struct {
	OContext                        string   `json:"@odata.context"`
	Oid                             string   `json:"@odata.id"`
	Otype                           string   `json:"@odata.type"`
	Id                              string   `json:"Id"`
	Name                            string   `json:"Name"`
	Status                          StatusRF `json:"Status"`
	ServiceEnabled                  *bool    `json:"ServiceEnabled,omitempty"`
	DateTime                        string   `json:"DateTime"`
	CompletedTaskOverWritePolicy    string   `json:"CompletedTaskOverWritePolicy"`
	LifeCycleEventOnTaskStateChange *bool    `json:"LifeCycleEventOnTaskStateChange"`

	Tasks ResourceID `json:"Tasks"`
}

// A Redfish task, via the Redfish TaskService
type Task struct {
	OContext   string `json:"@odata.context"`
	Oid        string `json:"@odata.id"`
	Otype      string `json:"@odata.type"`
	Id         string `json:"Id"`
	Name       string `json:"Name"`
	TaskState  string `json:"TaskState"`
	StartTime  string `json:"StartTime"`
	EndTime    string `json:"EndTime"`
	TaskStatus string `json:"TaskStatus"`

	Messages []Message `json:"Messages"`
}

// A Redfish "Message", given in tasks and message registries.  Not the same
// as a LogEntry or EventRecord.  Rather, EventRecords will contain a MessageId
// to look up a Message in the given registry of the given version for detailed
// information about what occurred, i.e. without actually packaging all of
// this info in each event.
//
// There is a "Registries" link at the service root that has a link to
// message registries that may be specific to the implementation.
//   e.g. /redfish/v1/Registries -> /redfish/v1/Registries/Messages/En
//
// "Standard" registries are available at:
//      http://redfish.dmtf.org/schemas/registries
// For example: http://redfish.dmtf.org/schemas/registries/v1/Base.1.1.0.json
//
// The MessageId format is registry.version.message.
//    e.g. Base.1.0.0.PropertyUnknown  (Standard Redfish Base schema)
//         iDRAC.1.4.0.AMP0300         (via /redfish/v1/Registries/Messages/En)
type Message struct {
	MessageId         string   `json:"MessageId"`
	Message           string   `json:"Message"`
	RelatedProperties []string `json:"RelatedProperties"`
	MessageArgs       []string `json:"MessageArgs"`
	Severity          string   `json:"Severity"`
	Resolution        string   `json:"Resolution"`
}

// Redfish update service.  This is the top-level object linked via the
// service root.  It gives information of the allowed actions and push
// URI configuration.  It also links to the available inventories.
type UpdateService struct {
	OContext          string      `json:"@odata.context"`
	Oetag             string      `json:"@odata.etag,omitempty"`
	Oid               string      `json:"@odata.id"`
	Otype             string      `json:"@odata.type"`
	Id                string      `json:"Id"`
	Name              string      `json:"Name"`
	Status            *StatusRF   `json:"Status,omitempty"`
	ServiceEnabled    *bool       `json:"ServiceEnabled,omitempty"`
	FirmwareInventory *ResourceID `json:"FirmwareInventory,omitempty"`
	SoftwareInventory *ResourceID `json:"SoftwareInventory,omitempty"`

	Actions UpdateServiceActions `json:"Actions"`

	HttpPushUri            string              `json:"HttpPushUri,omitempty"`
	HttpPushUriOptions     *HttpPushUriOptions `json:"HttpPushUriOptions,omitempty"`
	HttpPushUriOptionsBusy *bool               `json:"HttpPushUriOptionsBusy,omitempty"`
	HttpPushUriTargets     []string            `json:"HttpPushUriTargets,omitempty"`
	HttpPushUriTargetsBusy *bool               `json:"HttpPushUriTargetsBusy,omitempty"`
}

// UpdateService - Actions defined by Redfish for this type
type UpdateServiceActions struct {
	SimpleUpdate *ActionSimpleUpdate `json:"#UpdateService.SimpleUpdate,omitempty"`
}

// UpdateService - Simple update action payload.
type ActionSimpleUpdate struct {
	Target string `json:"target,omitempty"`
	Title  string `json:"title,omitempty"`
}

// UpdateService - HTTP push URI options payload.
type HttpPushUriOptions struct {
	HttpPushUriApplyTime *HttpPushUriApplyTime `json:"HttpPushUriApplyTime,omitempty"`
}

// UpdateService - HTTP push URI apply time payload.
type HttpPushUriApplyTime struct {
	ApplyTime                          string      `json:"ApplyTime,omitempty"`
	MaintenanceWindowDurationInSeconds json.Number `json:"MaintenanceWindowDurationInSeconds,omitempty"`
	MaintenanceWindowStartTime         string      `json:"MaintenanceWindowStartTime,omitempty"`
}

// RedfishErrorContents - Contains properties used to describe an error from a
// Redfish Service. Code - A string indicating a specific MessageId from the
// message registry. Message - A human-readable error message corresponding to
// the message in the message registry. ExtendedInfo - An array of message
// objects describing one or more error message(s).
// Schemas are available at:
//       https://redfish.dmtf.org/redfish/schema_index
// redfish-error: https://redfish.dmtf.org/schemas/redfish-error.v1_0_0.json
type RedfishErrorContents struct {
	Code         string    `json:"code"`
	Message      string    `json:"message"`
	ExtendedInfo []Message `json:"@Message.ExtendedInfo"`
}

// RedfishError - Contains an error payload from a Redfish service. Error -
// Contains properties used to describe an error from a Redfish Service.
// Schemas are available at:
//       https://redfish.dmtf.org/redfish/schema_index
// redfish-error: https://redfish.dmtf.org/schemas/redfish-error.v1_0_0.json
type RedfishError struct {
	Error RedfishErrorContents `json:"error"`
}
