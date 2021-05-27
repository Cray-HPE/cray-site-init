/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

import (
	"fmt"
	"net"
	"os"

	base "github.com/Cray-HPE/hms-base"
	"github.com/gocarina/gocsv"
)

// ManagementSwitchBrand known list of Management switch brands
type ManagementSwitchBrand string

func (msb ManagementSwitchBrand) String() string {
	return string(msb)
}

// ManagementSwitchBrandAruba for Aruba Management switches
const ManagementSwitchBrandAruba ManagementSwitchBrand = "Aruba"

// ManagementSwitchBrandDell for Dell Management switches
const ManagementSwitchBrandDell ManagementSwitchBrand = "Dell"

// ManagementSwitchBrandMellanox for Mellanox Management switches
const ManagementSwitchBrandMellanox ManagementSwitchBrand = "Mellanox"

// ManagementSwitchType the type of management switch CDU/Leaf/Spine/Aggregation
type ManagementSwitchType string

// ManagementSwitchTypeCDU is the type for CDU Management switches
const ManagementSwitchTypeCDU ManagementSwitchType = "CDU"

// ManagementSwitchTypeLeaf is the type for Leaf Management switches
const ManagementSwitchTypeLeaf ManagementSwitchType = "Leaf"

// ManagementSwitchTypeSpine is the type for Spine Management switches
const ManagementSwitchTypeSpine ManagementSwitchType = "Spine"

// ManagementSwitchTypeAggregation is the type for Aggregation Management switches
const ManagementSwitchTypeAggregation ManagementSwitchType = "Aggregation"

func (mst ManagementSwitchType) String() string {
	return string(mst)
}

// IsManagementSwitchTypeValid validates the given ManagementSwitchType
func IsManagementSwitchTypeValid(mst ManagementSwitchType) bool {
	switch mst {
	case ManagementSwitchTypeAggregation:
		fallthrough
	case ManagementSwitchTypeCDU:
		fallthrough
	case ManagementSwitchTypeLeaf:
		fallthrough
	case ManagementSwitchTypeSpine:
		return true
	}

	return false
}

// ManagementSwitch is a type for managing Management switches
type ManagementSwitch struct {
	Xname               string                `json:"xname" mapstructure:"xname" csv:"Switch Xname"` // Required for SLS
	Name                string                `json:"name" mapstructure:"name" csv:"-"`              // Required for SLS to update DNS
	Brand               ManagementSwitchBrand `json:"brand" mapstructure:"brand" csv:"Brand"`
	Model               string                `json:"model" mapstructure:"model" csv:"Model"`
	Os                  string                `json:"operating-system" mapstructure:"operating-system" csv:"-"`
	Firmware            string                `json:"firmware" mapstructure:"firmware" csv:"-"`
	SwitchType          ManagementSwitchType  `json:"type" mapstructure:"type" csv:"Type"` //"CDU/Leaf/Spine/Aggregation"
	ManagementInterface net.IP                `json:"ip" mapstructure:"ip" csv:"-"`        // SNMP/REST interface IP (not a distinct BMC)  // Required for SLS
}

// Validate ManagementSwitch contents
func (mySwitch *ManagementSwitch) Validate() error {
	// Validate the data that was read in switch_metadata.csv. We are inforcing 3 constaints:
	// 1. Validate the xname is valid
	// 2. The specified switch type is valid
	// 3. The HMS type for the xname matches the type of switch being used

	xname := mySwitch.Xname
	// Verify xname is valid
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid xname for Switch: %s", xname)
	}

	// Verify that the specify management switch type is one of the known values
	if !IsManagementSwitchTypeValid(mySwitch.SwitchType) {
		return fmt.Errorf("invalid management switch type: %s %s", xname, mySwitch.SwitchType)
	}

	// Now we need to verify that the correct switch xname format was used for the different
	// types of management switches.
	hmsType := base.GetHMSType(xname)
	switch mySwitch.SwitchType {
	case ManagementSwitchTypeLeaf:
		if hmsType != base.MgmtSwitch {
			return fmt.Errorf("invalid xname used for Leaf switch: %s,  should use xXcCwW format", xname)
		}
	case ManagementSwitchTypeSpine:
		fallthrough
	case ManagementSwitchTypeAggregation:
		if hmsType != base.MgmtHLSwitch {
			return fmt.Errorf("invalid xname used for Spine/Aggregation switch: %s, should use xXcChHsS format", xname)
		}
	case ManagementSwitchTypeCDU:
		// CDU Management switches can be under different switch types
		// dDwW - This is normally used for mountain systems, and Hill systems that have CDU switches getting
		// power from the Hill cabinet.
		//
		// xXcChHsS - This is normally for Aggregation and Spine switches, but some Hill cabinets have the
		// CDU switches powered/racked into the adjacent river cabinet.

		if hmsType != base.CDUMgmtSwitch && hmsType != base.MgmtHLSwitch {
			return fmt.Errorf("invalid xname used for CDU switch: %s, should use dDwW format (if in an adjacent river cabinet to a TBD cabinet use the xXcChHsS format)", xname)
		}
	default:
		return fmt.Errorf("invalid switch type for xname: %s", xname)
	}

	return nil
}

// Normalize the values of a Management switch
func (mySwitch *ManagementSwitch) Normalize() error {
	// Right now we only need to the normalize the xname for the switch. IE strip any leading 0s
	mySwitch.Xname = base.NormalizeHMSCompID(mySwitch.Xname)

	return nil
}

// ReadSwitchCSV parses a CSV file into a list of ManagementSwitch structs
func ReadSwitchCSV(filename string) ([]*ManagementSwitch, error) {
	switches := []*ManagementSwitch{}
	switchMetadataFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return switches, err
	}
	defer switchMetadataFile.Close()
	err = gocsv.UnmarshalFile(switchMetadataFile, &switches)
	if err != nil { // Load switches from file
		return switches, err
	}
	return switches, nil
}
