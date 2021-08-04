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

/////////////////////////////////////////////////////////////////////////////

// Collections

// Collection of PowerDistribution references (e.g. RackPDU)
type PowerDistributionCollection GenericCollection

// Collection of Outlets, i.e. linked to a PDU
type OutletCollection GenericCollection

// Collection of Circuits, i.e. linked with a PDU
type CircuitCollection GenericCollection

// Collection of Sensors, i.e. linked to a parent component, e.g. PDU
type AlarmCollection GenericCollection

// Collection of Sensors, i.e. linked to a parent component, e.g. PDU
type SensorsCollection GenericCollection

/////////////////////////////////////////////////////////////////////////////

// Redfish PowerEquipment
//
// From DMTF: "This resource shall be used to represent the set of Power
// Equipment for a Redfish implementation."
//  Example: /redfish/v1/PowerEquipment
type PowerEquipment struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	Id          string `json:"Id"`
	Description string `json:"Description"`
	Name        string `json:"Name"`

	Status StatusRF `json:"Status"`

	Actions    *PowerEquipmentActions `json:"Actions,omitempty"`
	OemActions *json.RawMessage       `json:"OemActions,omitempty"`

	// These are all pointers to collections
	Alarms           ResourceID `json:"Alarms"`
	Generators       ResourceID `json:"Generators,omitempty"`
	FloorPDUs        ResourceID `json:"FloorPDUs"`
	PowerMeters      ResourceID `json:"PowerMeters,omitempty"`
	RackPDUs         ResourceID `json:"RackPDUs"`
	Rectifiers       ResourceID `json:"Rectifiers"`
	Sensors          ResourceID `json:"Sensors,omitempty"`
	Switchgear       ResourceID `json:"Switchgear,omitempty"`
	TransferSwitches ResourceID `json:"TransferSwitches"`
	UPSs             ResourceID `json:"UPSs"`
	VFDs             ResourceID `json:"VFDs,omitempty"`

	// Array of alarm objects
	TriggeredAlarms []Alarm `json:"TriggeredAlarms"`
	// Links to chassis/managers
	Links PowerEquipmentLinks `json:"Links"`
}

// PowerEquipment sub-struct - PowerEquipmentLinks
type PowerEquipmentLinks struct {
	Chassis         []ResourceID `json:"Chassis"`
	ChassisOCount   int          `json:"Chassis@odata.count"`
	ManagedBy       []ResourceID `json:"ManagedBy"`
	ManagedByOCount int          `json:"ManagedBy@odata.count"`
}

// Actions for PowerEquipment - OEM only for now
type PowerEquipmentActions struct {
	OEM *json.RawMessage `json:"Oem,omitempty"`
}

/////////////////////////////////////////////////////////////////////////////

// Redfish PowerDistribution
//
// From DMTF: "This resource shall be used to represent a Power Distribution
// component or unit for a Redfish implementation."
//  Example: /redfish/v1/PowerEquipment/RackPDUs/1
type PowerDistribution struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	// Embedded structs - see below
	PowerDistributionLocationInfo
	PowerDistributionFRUInfo

	Actions *PowerDistributionActions `json:"Actions,omitempty"`
	OEM     *json.RawMessage          `json:"Oem,omitempty"`

	// These are all links to collections that point to the given type
	Alarms       ResourceID `json:"Alarms"`             // Alarms
	Branches     ResourceID `json:"Branches"`           // Circuits
	Feeders      ResourceID `json:"Feeders,omitempty"`  // Circuit
	Mains        ResourceID `json:"Mains"`              // Circuits
	Outlets      ResourceID `json:"Outlets"`            // Outlets
	OutletGroups ResourceID `json:"OutletGroups"`       // OutletGroups
	Subfeeds     ResourceID `json:"Subfeeds,omitempty"` // Circuits
	Metrics      ResourceID `json:"Metrics"`            // PowerDistributionMetrics
	Sensors      ResourceID `json:"Sensors,omitempty"`  // Sensors

	// Current State (all enums)
	PowerState string   `json:"PowerState"`
	Status     StatusRF `json:"Status"`
}

// Redfish fields from the PowerDistributionFRUInfo schema that go into
// HWInventoryByLocation.  We capture them as an embedded struct within the
// full schema during inventory discovery.
type PowerDistributionLocationInfo struct {
	Id          string    `json:"Id"`
	Description string    `json:"Description"`
	Name        string    `json:"Name"`
	UUID        string    `json:"UUID"`
	Location    *Location `json:Location,omitempty"`
}

// Redfish fields from the PowerDistribution schema that go into
// HWInventoryByFRU.  We capture them as an embedded struct within the
// full schema during inventory discovery.
type PowerDistributionFRUInfo struct {
	AssetTag          string         `json:"AssetTag"`
	DateOfManufacture string         `json:"DateOfManufacture,omitempty"`
	EquipmentType     string         `json:"EquipmentType"`
	FirmwareVersion   string         `json:"FirmwareVersion"`
	HardwareRevision  string         `json:"HardwareRevision"`
	Manufacturer      string         `json:"Manufacturer"`
	Model             string         `json:"Model"`
	PartNumber        string         `json:"PartNumber"`
	SerialNumber      string         `json:"SerialNumber"`
	CircuitSummary    CircuitSummary `json:"CircuitSummary"`
}

// Redfish PowerDistribution Actions sub-struct
type PowerDistributionActions struct {
	OEM *json.RawMessage `json:"Oem,omitempty"`
}

// CircuitSummary sub-struct of PowerDistribution
// These are all-readonly
type CircuitSummary struct {
	ControlledOutlets json.Number `json:"ControlledOutlets,omitempty"`
	MonitoredBranches json.Number `json:"MonitoredBranches,omitempty"`
	MonitoredOutlets  json.Number `json:"MonitoredOutlets,omitempty"`
	MonitoredPhases   json.Number `json:"MonitoredPhases,omitempty"`
	TotalBranches     json.Number `json:"TotalBranches,omitempty"`
	TotalOutlets      json.Number `json:"TotalOutlets,omitempty"`
	TotalPhases       json.Number `json:"TotalPhases,omitempty"`
}

/////////////////////////////////////////////////////////////////////////////

// Redfish PDU Outlet
//
// This represents an individual outlet on a PDU
//  Example: /redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1
type Outlet struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	// HwInv instance properties - Embedded struct
	OutletLocationInfo

	// HwInv durable FRU properties - Embedded struct
	OutletFRUInfo

	// Actions
	Actions *OutletActions `json:"Actions,omitempty"`

	// Current State (all enums)
	BreakerState string   `json:"BreakerState"`
	PowerState   string   `json:"PowerState"`
	Status       StatusRF `json:"Status"`
	IndicatorLED string   `json:"IndicatorLED"`

	// Sensors
	CurrentSensor           *SensorExcerpt      `json:"CurrentSensor,omitempty"`
	EnergySensor            *SensorExcerpt      `json:"EnergySensor,omitempty"`
	FrequencySensor         *SensorExcerpt      `json:"FrequencySensor,omitempty"`
	PowerSensor             *SensorPowerExcerpt `json:"PowerSensor,omitempty"`
	PolyPhaseCurrentSensors *Currents           `json:"PolyPhaseCurrentSensors,omitempty"`
	PolyPhaseEnergySensors  *EnergyReadings     `json:"PolyPhaseEnergySensors,omitempty"`
	PolyPhasePowerSensors   *PowerReadings      `json:"PolyPhasePowerSensors,omitempty"`
	PolyPhaseVoltageSensors *Voltages           `json:"PolyPhaseVoltageSensors,omitempty"`
	TemperatureSensor       *SensorExcerpt      `json:"TemperatureSensor,omitempty"`
	VoltageSensor           *SensorExcerpt      `json:"VoltageSensor,omitempty"`

	// Configuration
	PowerCycleDelaySeconds   json.Number `json:"PowerCycleDelaySeconds,omitempty"`
	PowerOnDelaySeconds      json.Number `json:"PowerOnDelaySeconds,omitempty"`
	PowerRestoreDelaySeconds json.Number `json:"PowerRestoreDelaySeconds,omitempty"`
	PowerRestorePolicy       string      `json:"PowerRestorePolicy,omitempty"`

	Oem *json.RawMessage `json:"Oem,omitempty"`

	// Links - if useful to continue inventory discovery
	Links OutletLinks `json:"Links"`
}

// Outlets do not have individual FRUs, PDUs do, but their properties are
// potentially important.  This is location-dependent data for HwInventory
type OutletLocationInfo struct {
	Id          string `json:"Id"`
	Description string `json:"Description"`
	Name        string `json:"Name"`
}

// Outlets do not have individual FRUs, PDUs do, but their properties are
// potentially important. This is FRU-dependent data for HwInventory
// Note: omits configurable parameters.
type OutletFRUInfo struct {
	NominalVoltage   string         `json:"NominalVoltage,omitempty"`
	OutletType       string         `json:"OutletType"` // Enum
	EnergySensor     *SensorExcerpt `json:"EnergySensor,omitempty"`
	FrequencySensor  *SensorExcerpt `json:"FrequencySensor,omitempty"`
	PhaseWiringType  string         `json:"PhaseWiringType,omitempty"` // Enum
	PowerEnabled     *bool          `json:"PowerEnabled,omitempty"`    // Can be powered?
	RatedCurrentAmps json.Number    `json:"RatedCurrentAmps,omitempty"`
	VoltageType      string         `json:"VoltageType,omitempty"` // Enum
}

// Redfish Outlet Actions sub-struct
type OutletActions struct {
	PowerControl    *ActionPowerControl    `json:"#Outlet.PowerControl,omitempty"`
	ResetBreaker    *ActionResetBreaker    `json:"#Outlet.ResetBreaker,omitempty"`
	ResetStatistics *ActionResetStatistics `json:"#Outlet.ResetStatistics,omitempty"`
	OEM             *json.RawMessage       `json:"Oem,omitempty"`
}

// PowerControl - Outlet
type ActionPowerControl struct {
	AllowableValues []string `json:"PowerState@Redfish.AllowableValues,omitempty"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}

// Action type ResetBreaker - Outlet
type ActionResetBreaker struct {
	AllowableValues []string `json:"ResetBreaker@Redfish.AllowableValues,omitempty"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}

// Action type ResetStatistics - Outlet
// Is there AllowableValues here?  No idea what {type}@ prefix should be then.
type ActionResetStatistics struct {
	AllowableValues []string `json:"ResetStatistics@Redfish.AllowableValues,omitempty"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}

// Links for outlet
type OutletLinks struct {
	OEM           *json.RawMessage `json:"Oem,omitempty"`
	BranchCircuit ResourceID       `json:"BranchCircuit"`
}

// PowerReadings - Used in Outlet and Circuit
// Note: For all of these, we may only have one of the particular fields in
// a given resource.
type PowerReadings struct {
	Line1ToLine2   *SensorPowerExcerpt `json:"Line1ToLine2,omitempty"`
	Line1ToNeutral *SensorPowerExcerpt `json:"Line1ToNeutral,omitempty"`
	Line2ToLine3   *SensorPowerExcerpt `json:"Line2ToLine3,omitempty"`
	Line2ToNeutral *SensorPowerExcerpt `json:"Line2ToNeutral,omitempty"`
	Line3ToLine1   *SensorPowerExcerpt `json:"Line3ToLine1,omitempty"`
	Line3ToNeutral *SensorPowerExcerpt `json:"Line3ToNeutral,omitempty"`
}

// EnergyReadings - Used in Outlet and Circuit
type EnergyReadings struct {
	Line1ToLine2   *SensorExcerpt `json:"Line1ToLine2,omitempty"`
	Line1ToNeutral *SensorExcerpt `json:"Line1ToNeutral,omitempty"`
	Line2ToLine3   *SensorExcerpt `json:"Line2ToLine3,omitempty"`
	Line2ToNeutral *SensorExcerpt `json:"Line2ToNeutral,omitempty"`
	Line3ToLine1   *SensorExcerpt `json:"Line3ToLine1,omitempty"`
	Line3ToNeutral *SensorExcerpt `json:"Line3ToNeutral,omitempty"`
}

// Voltages  - Used in Outlet and Circuit
type Voltages struct {
	Line1ToLine2   *SensorExcerpt `json:"Line1ToLine2,omitempty"`
	Line1ToNeutral *SensorExcerpt `json:"Line1ToNeutral,omitempty"`
	Line2ToLine3   *SensorExcerpt `json:"Line2ToLine3,omitempty"`
	Line2ToNeutral *SensorExcerpt `json:"Line2ToNeutral,omitempty"`
	Line3ToLine1   *SensorExcerpt `json:"Line3ToLine1,omitempty"`
	Line3ToNeutral *SensorExcerpt `json:"Line3ToNeutral,omitempty"`
}

// Currents - Used in Outlet and Circuit
type Currents struct {
	Line1   *SensorExcerpt `json:"Line1,omitempty"`
	Line2   *SensorExcerpt `json:"Line2,omitempty"`
	Line3   *SensorExcerpt `json:"Line3,omitempty"`
	Neutral *SensorExcerpt `json:"Neutral,omitempty"`
}

/////////////////////////////////////////////////////////////////////////////

// Redfish Circuit
//
// This represents a Circuitwhich will have power components linked to it.
//  Example: /redfish/v1/PowerEquipment/RackPDUs/1/Branches/A
type Circuit struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	// Embedded structs
	CircuitsLocationInfo
	CircuitFRUInfo

	Actions *CircuitActions `json:"Actions,omitempty"`

	Id          string `json:"Id"`
	Description string `json:"Description"`
	Name        string `json:"Name"`
	CircuitType string `json:"CircuitType"`

	// Current State (all enums)
	BreakerState string   `json:"BreakerState"`
	PowerState   string   `json:"PowerState"`
	Status       StatusRF `json:"Status"`
	IndicatorLED string   `json:"IndicatorLED"`

	// Configuration
	CriticalCircuit          *bool       `json:"CriticalCircuit,omitempty"`
	PowerCycleDelaySeconds   json.Number `json:"PowerCycleDelaySeconds,omitempty"`
	PowerOnDelaySeconds      json.Number `json:"PowerOnDelaySeconds,omitempty"`
	PowerRestoreDelaySeconds json.Number `json:"PowerRestoreDelaySeconds,omitempty"`
	PowerRestorePolicy       string      `json:"PowerRestorePolicy,omitempty"`

	Outlets       []ResourceID     `json:"Outlets"`
	OutletsOCount int              `json:"Outlets@odata.count,omitempty"`
	Oem           *json.RawMessage `json:"Oem,omitempty"`

	// Links for outlet
	Links CircuitLinks `json:"Links"`

	// Sensors
	CurrentSensor           SensorExcerpt      `json:"CurrentSensor,omitempty"`
	EnergySensor            SensorExcerpt      `json:"EnergySensor,omitempty"`
	FrequencySensor         SensorExcerpt      `json:"FrequencySensor,omitempty"`
	PowerSensor             SensorPowerExcerpt `json:"PowerSensor,omitempty"`
	PolyPhaseCurrentSensors Currents           `json:"PolyPhaseCurrentSensors,omitempty"`
	PolyPhaseEnergySensors  EnergyReadings     `json:"PolyPhaseEnergySensors,omitempty"`
	PolyPhasePowerSensors   PowerReadings      `json:"PolyPhasePowerSensors,omitempty"`
	PolyPhaseVoltageSensors Voltages           `json:"PolyPhaseVoltageSensors,omitempty"`
	TemperatureSensor       SensorExcerpt      `json:"TemperatureSensor,omitempty"`
	VoltageSensor           SensorExcerpt      `json:"VoltageSensor,omitempty"`
}

// Redfish Circuit Actions sub-struct
type CircuitActions struct {
	PowerControl    *ActionPowerControl    `json:"#Circuit.PowerControl,omitempty"`
	ResetBreaker    *ActionResetBreaker    `json:"#Circuit.ResetBreaker,omitempty"`
	ResetStatistics *ActionResetStatistics `json:"#Circuit.ResetStatistics,omitempty"`
	OEM             *json.RawMessage       `json:"Oem,omitempty"`
}

// Circuits do not have individual FRUs, PDUs do, but their properties are
// potentially important.  This is location-dependent data for HwInventory
// Note: Circuit schema is fairly similar to Outlet, and links to Outlets.
type CircuitsLocationInfo struct {
	Id          string `json:"Id"`
	Description string `json:"Description"`
	Name        string `json:"Name"`
}

// Circuits do not have individual FRUs, PDUs do, but their properties are
// potentially important. This is FRU-dependent data for HwInventory
// Note: omits configurable parameters.
// Note2: Circuit schema is fairly similar to Outlet, and links to Outlets.
type CircuitFRUInfo struct {
	NominalVoltage   string      `json:"NominalVoltage,omitempty"`
	CircuitType      string      `json:"CircuitType"` // Enum
	PlugType         string      `json:"PlugType"`    // Enum - Matches OutletType
	PhaseWiringType  string      `json:"PhaseWiringType,omitempty"`
	PowerEnabled     *bool       `json:"PowerEnabled,omitempty"` // Can be powered?
	RatedCurrentAmps json.Number `json:"RatedCurrentAmps,omitempty"`
	VoltageType      string      `json:"VoltageType,omitempty"`
}

// Links for Circuit
type CircuitLinks struct {
	OEM           *json.RawMessage `json:"Oem,omitempty"`
	BranchCircuit ResourceID       `json:"BranchCircuit"`
}

/////////////////////////////////////////////////////////////////////////////

// Sensor
//
// From the DMTF: "This resource shall be used to represent resources that
// represent the sensor data."
//  Example: /redfish/v1/PowerEquipment/RackPDUs/1/Sensors/FrequencyA1
type Sensor struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	Id          string    `json:"Id"`
	Description string    `json:"Description"`
	Name        string    `json:"Name"`
	Location    *Location `json:Location,omitempty"`

	Actions                            SensorActions    `json:"Actions,omitempty"`
	Accuracy                           json.Number      `json:"Accuracy,omitempty"`
	AdjustedMaxAllowableOperatingValue json.Number      `json:"AdjustedMaxAllowableOperatingValue,omitempty"`
	AdjustedMinAllowableOperatingValue json.Number      `json:"AdjustedMinAllowableOperatingValue,omitempty"`
	ApparentVA                         json.Number      `json:"ApparentVA,omitempty"`
	ElectricalContext                  string           `json:"ElectricalContext"` // enum
	LoadPercent                        json.Number      `json:"LoadPercent,omitempty"`
	MaxAllowableOperatingValue         json.Number      `json:"MaxAllowableOperatingValue,omitempty"`
	MinAllowableOperatingValue         json.Number      `json:"MinAllowableOperatingValue,omitempty"`
	OEM                                *json.RawMessage `json:"Oem,omitempty"`
	PeakReading                        json.Number      `json:"PeakReading,omitempty"`
	PeakReadingTime                    string           `json:"PeakReadingtime,omitempty"`
	PowerFactor                        json.Number      `json:"PowerFactor,omitempty"`
	PhysicalContext                    string           `json:"PhysicalContext,omitempty"`    //enum
	PhysicalSubContext                 string           `json:"PhysicalSubContext,omitempty"` //enum
	Precision                          json.Number      `json:"Precision,omitempty"`
	ReactiveVAR                        json.Number      `json:"ReactiveVAR,omitempty"`
	Reading                            json.Number      `json:"Reading,omitempty"`
	ReadingRangeMax                    json.Number      `json:"ReadingRangeMax,omitempty"`
	ReadingRangeMin                    json.Number      `json:"ReadingRangeMin,omitempty"`
	ReadingType                        string           `json:"ReadingType,omitempty"` //enum
	ReadingUnits                       string           `json:"ReadingUnits,omitempty"`
	Status                             StatusRF         `json:"Status,omitempty"`
	SensingFrequency                   json.Number      `json:"SensingFrequency,omitempty"`
	SensorResetTime                    json.Number      `json:"SensorResetTime,omitempty"`
	Thresholds                         *Thresholds      `json:"Thresholds,omitempty"`
	VoltageType                        string           `json:"VoltageType,omitempty"` // enum
}

// Redfish Sensor Actions sub-struct
type SensorActions struct {
	ResetStatistics ActionResetStatistics `json:"#Sensor.ResetStatistics,omitempty"`
	OEM             *json.RawMessage      `json:"Oem,omitempty"`
}

// SensorPowerExcerpt -  Substruct of Outlet and other power-related objects
// From DMTF: "This resource shall be used to represent resources that
// represent the sensor data."
type SensorPowerExcerpt struct {
	ApparentVA         json.Number `json:"ApparentVA,omitempty"`
	DataSourceUri      string      `json:"DataSourceUri"`
	LoadPercent        json.Number `json:"LoadPercent,omitempty"`
	Name               string      `json:"Name"`
	PeakReading        json.Number `json:"PeakReading,omitempty"`
	PowerFactor        json.Number `json:"PowerFactor,omitempty"`
	PhysicalContext    string      `json:"PhysicalContext,omitempty"`    //enum
	PhysicalSubContext string      `json:"PhysicalSubContext,omitempty"` //enum
	ReactiveVAR        json.Number `json:"ReactiveVAR,omitempty"`
	Reading            json.Number `json:"Reading,omitempty"`
	ReadingUnits       string      `json:"ReadingUnits,omitempty"`
	Status             StatusRF    `json:"Status,omitempty"`
}

// SensorExcerpt -  Substruct of Outlet and other power-related objects
// This is the more general non-power version of SensorPowerExcerpt
type SensorExcerpt struct {
	DataSourceUri      string      `json:"DataSourceUri"`
	Name               string      `json:"Name"`
	PeakReading        json.Number `json:"PeakReading,omitempty"`
	PhysicalContext    string      `json:"PhysicalContext,omitempty"`    //enum
	PhysicalSubContext string      `json:"PhysicalSubContext,omitempty"` //enum
	Reading            json.Number `json:"Reading,omitempty"`
	ReadingUnits       string      `json:"ReadingUnits,omitempty"`
	Status             StatusRF    `json:"Status,omitempty"`
}

// Sub-struct of Sensor - Thresholds
type Thresholds struct {
	LowerCaution  Threshold `json:LowerCaution"`
	LowerCritical Threshold `json:LowerCritical"`
	LowerFatal    Threshold `json:LowerFatal"`
	UpperCaution  Threshold `json:UpperCaution"`
	UpperCritical Threshold `json:UpperCritical"`
	UpperFatal    Threshold `json:UpperFatal"`
}

// Sub-struct of Sensor - Threshold
type Threshold struct {
	Activation string      `json:"Activation"` // enum
	DwellTime  string      `json:"DwellTime"`
	Reading    json.Number `json:"Reading,omitempty"`
}

/////////////////////////////////////////////////////////////////////////////

// Redfish Alarms
//
// DMTF: "An Alarm is an entity that has a latch type behavior.
// It is designed to be used to persist sensor threshold crossing
// or to capture the momentary state of another property."
type Alarm struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	Id          string `json:"Id"`
	Description string `json:"Description"`
	Name        string `json:"Name"`

	Actions        *AlarmActions    `json:"Actions,omitempty"`
	AlarmState     string           `json:"AlarmState"`
	AutomaticReArm *bool            `json:"AutomaticReArm,omitempty"`
	OEM            *json.RawMessage `json:"Oem,omitempty"`
	Message        string           `json:"Message"`
	MessageArgs    []string         `json:"MessageArgs"`
	MessageId      string           `json:"MessageId"`
	Severity       string           `json:"Severity"`

	Links AlarmLinks `json:"Links"`
}

// Actions for Alarm - OEM only for now
type AlarmActions struct {
	OEM *json.RawMessage `json:"Oem,omitempty"`
}

// Links for Alarm
type AlarmLinks struct {
	OEM             *json.RawMessage `json:"Oem,omitempty"`
	RelatedProperty ResourceID       `json:"RelatedProperty"`
	RelatedSensor   ResourceID       `json:"RelatedSensor"`
}

/////////////////////////////////////////////////////////////////////////////

// Location
//
// Resource type.  Appears under Chassis, PowerDistribution, etc.
type Location struct {
	ContactInfo   *ContactInfo   `json:"ContactInfo,omitempty"`
	Latitude      json.Number    `json:"Latitude,omitempty"`
	Longitude     json.Number    `json:"Longitude,omitempty"`
	PartLocation  *PartLocation  `json:"PartLocation,omitempty"`
	Placement     *Placement     `json:"Placement,omitempty"`
	PostalAddress *PostalAddress `json:"PostalAddress,omitempty"`
}

// Within Location - ContactInfo
type ContactInfo struct {
	ContactName  string `json:"ContactName"`
	EmailAddress string `json:"EmailAddress"`
	PhoneNumber  string `json:"PhoneNumber,omitempty"`
}

// Within Location - PartLocation
type PartLocation struct {
	LocationOrdinalValue json.Number `json:"LocationOrdinalValue,omitempty"`
	LocationType         string      `json:"LocationType"` //enum
	Orientation          string      `json:"Orientation"`  //enum
	Reference            string      `json:"Reference"`    //enum
	ServiceLabel         string      `json:"ServiceLabel"`
}

// Within Location - PostalAddress
type PostalAddress struct {
	Country    string `json:"Country"`
	Territory  string `json:"Territory"`
	City       string `json:"City"`
	Street     string `json:"Street"`
	Name       string `json:"Name"`
	PostalCode string `json:"PostalCode"`
	Building   string `json:"Building"`
	Floor      string `json:"Floor"`
	Room       string `json:"Room"`
}

type Placement struct {
	AdditionalInfo  string      `json:"AdditionalInfo,omitempty"`
	Rack            string      `json:"Rack,omitempty"`
	RackOffset      json.Number `json:"RackOffset,omitempty"`
	RackOffsetUnits string      `json:"RackOffsetUnits,omitempty"`
	Row             string      `json:"Row,omitempty"`
}

// Identifier
//
//  Resource type.  Appears under various schemas
type Identifier struct {
	DurableName       string `json:"DurableName"`
	DurableNameFormat string `json:"DurableNameFormat"` // Enum
}

/////////////////////////////////////////////////////////////////////////////

// Power and PowerSupplies (Recitifiers)

//The Power type definition below is based on current L1 Power support, and does not have
//many of the fields defined in the DMTF Redfish Power 1.6.0 schema, specifically PowerControl.
//However, HSM does capture PowerControl info for ComputerSystems (Nodes), see ComponentSystemInfo definition.
type Power struct {
	OContext            string         `json:"@odata.context"`
	OCount              int            `json:"@odata.count"` // Oldest schemas use
	Oid                 string         `json:"@odata.id"`
	Otype               string         `json:"@odata.type"`
	Description         string         `json:"Description"`
	Name                string         `json:"Name"`
	Id                  string         `json:"Id"`
	PowerSupplies       []*PowerSupply `json:"PowerSupplies"`
	PowerSuppliesOCount int            `json:"PowerSupplies@odata.count"` // Most schemas
}

// Redfish pass-through from Redfish "PowerSupply"
// This is the set of Redfish fields for this object that HMS understands
// and/or finds useful.  Those assigned to either the *LocationInfo
// or *FRUInfo subfields constitute the type specific fields in the
// HWInventory objects that are returned in response to queries.
type PowerSupply struct {
	Oid string `json:"@odata.id"`

	// Embedded structs.
	PowerSupplyLocationInfoRF
	PowerSupplyFRUInfoRF

	Status StatusRF `json:"Status"`
}

// Location-specific Redfish properties to be stored in hardware inventory
// These are only relevant to the currently installed location of the FRU
type PowerSupplyLocationInfoRF struct {
	Name            string `json:"Name"`
	FirmwareVersion string `json:"FirmwareVersion"`
}

// Durable Redfish properties to be stored in hardware inventory as
// a specific FRU, which is then link with it's current location
// i.e. an x-name.  These properties should follow the hardware and
// allow it to be tracked even when it is removed from the system.
// TODO: How to version these (as HMS structures)
type PowerSupplyFRUInfoRF struct {
	//Manufacture Info
	Manufacturer       string `json:"Manufacturer"`
	SerialNumber       string `json:"SerialNumber"`
	Model              string `json:"Model"`
	PartNumber         string `json:"PartNumber"`
	PowerCapacityWatts int    `json:"PowerCapacityWatts"`
	PowerInputWatts    int    `json:"PowerInputWatts"`
	PowerOutputWatts   int    `json:"PowerOutputWatts"`
	PowerSupplyType    string `json:"PowerSupplyType"`
}
