/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

import (
	"fmt"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

// CabinetKind is the type of the cabinet. This can either be a generic identifier like river,
// hill, or mountain. It can also be a cabinet model number like EX2000, EX25000, EX3000, or EX4000.
type CabinetKind string

// Enumerations of CabinetKinds for valid cabinet types.
const (
	CabinetKindRiver    = "river"
	CabinetKindHill     = "hill"
	CabinetKindMountain = "mountain"
	CabinetKindEX2000   = "EX2000"
	CabinetKindEX2500   = "EX2500"
	CabinetKindEX3000   = "EX3000"
	CabinetKindEX4000   = "EX4000"
)

// IsModel will return true if this cabinet type is the actual model of the cabinet.
func (ck CabinetKind) IsModel() bool {
	if ck == CabinetKindRiver || ck == CabinetKindHill || ck == CabinetKindMountain {
		return false
	}

	return true
}

// Class will determine the SLS cabinet class of this Cabinet group
func (ck CabinetKind) Class() (sls_common.CabinetType, error) {
	switch ck {
	case CabinetKindRiver:
		return sls_common.ClassRiver, nil
	case CabinetKindEX2000:
		fallthrough
	case CabinetKindEX2500:
		fallthrough
	case CabinetKindHill:
		return sls_common.ClassHill, nil
	case CabinetKindEX3000:
		fallthrough
	case CabinetKindEX4000:
		fallthrough
	case CabinetKindMountain:
		return sls_common.ClassMountain, nil
	default:
		return "", fmt.Errorf("unknown cabinet kind (%s)", ck)
	}
}

// CabinetGroupDetail stores information that can only come from Manufacturing
type CabinetGroupDetail struct {
	Kind            CabinetKind     `mapstructure:"cabinet-type" yaml:"type" valid:"-"`
	Cabinets        int             `mapstructure:"number" yaml:"total_number" valid:"-"`
	StartingCabinet int             `mapstructure:"starting-cabinet" yaml:"starting_id" valid:"-"`
	CabinetDetails  []CabinetDetail `mapstructure:"cabinets" yaml:"cabinets" valid:"-"`
}

// CabinetDetail stores information about individual cabinets
type CabinetDetail struct {
	ID           int           `mapstructure:"id" yaml:"id" valid:"numeric"`
	ChassisCount *ChassisCount `mapstructure:"chassis-count" yaml:"chassis-count" valid:"-"` // This field is only respected for EX2500 cabinets with variable chassis counts
	NMNSubnet    string        `mapstructure:"nmn-subnet" yaml:"nmn-subnet" valid:"-"`
	NMNVlanID    int16         `mapstructure:"nmn-vlan" yaml:"nmn-vlan" valid:"numeric"`
	HMNSubnet    string        `mapstructure:"hmn-subnet" yaml:"hmn-subnet" valid:"-"`
	HMNVlanID    int16         `mapstructure:"hmn-vlan" yaml:"hmn-vlan" valid:"numeric"`
}

// ChassisCount stores optional information about the chassis composition of the cabinet
type ChassisCount struct {
	LiquidCooled int `mapstructure:"liquid-cooled" yaml:"liquid-cooled" valid:"numeric"`
	AirCooled    int `mapstructure:"air-cooled" yaml:"air-cooled" valid:"numeric"`
}

// CabinetIDs returns the list of all cabinet ids
func (cgd *CabinetGroupDetail) CabinetIDs() []int {
	var cabinetIds []int
	for _, cab := range cgd.CabinetDetails {
		cabinetIds = append(cabinetIds, cab.ID)
	}
	return cabinetIds
}

// PopulateIds fills out the cabinet ids by doing simple math
func (cgd *CabinetGroupDetail) PopulateIds() {
	if len(cgd.CabinetDetails) < cgd.Cabinets {
		for cabIndex := 0; cabIndex < cgd.Cabinets; cabIndex++ {
			var tmpCabinet CabinetDetail
			if cabIndex < len(cgd.CabinetDetails) {
				tmpCabinet = cgd.CabinetDetails[cabIndex]
			} else {
				tmpCabinet = CabinetDetail{}
				cgd.CabinetDetails = append(cgd.CabinetDetails, tmpCabinet)
			}
			if tmpCabinet.ID == 0 {
				tmpCabinet.ID = cgd.StartingCabinet + cabIndex
			}
			cgd.CabinetDetails[cabIndex] = tmpCabinet
		}
	}
}

// GetCabinetDetails will retrieve all cabinets that have cabinet specific overrides present
func (cgd *CabinetGroupDetail) GetCabinetDetails() map[int]CabinetDetail {
	details := map[int]CabinetDetail{}

	for _, cab := range cgd.CabinetDetails {
		details[cab.ID] = cab
	}

	return details
}

// Length returns the expected number of cabinets from the total_number passed in or the length of the cabinet_ids array
func (cgd *CabinetGroupDetail) Length() int {
	if len(cgd.CabinetDetails) == 0 {
		return cgd.Cabinets
	}
	return len(cgd.CabinetDetails)
}

// CabinetTypes returns a list of cabinet types from the file
func (cdf *CabinetDetailFile) CabinetTypes() []CabinetKind {
	var out []CabinetKind
	for _, cd := range cdf.Cabinets {
		out = append(out, cd.Kind)
	}
	return out
}

// CabinetDetailFile is a struct that matches the syntax of the configuration file for non-sequential cabinet ids
type CabinetDetailFile struct {
	Cabinets []CabinetGroupDetail `yaml:"cabinets"`
}

// LoadCabinetDetailFile loads the cabinet details from the filesystem
func LoadCabinetDetailFile(path string) (CabinetDetailFile, error) {
	var cabDetailFile CabinetDetailFile
	err := csiFiles.ReadYAMLConfig(path, &cabDetailFile)
	return cabDetailFile, err
}

// CabinetFilterFunc is a function type that functions can implement to define rules if a cabinet with its its
// CabinetGroupDetail and CabinetDetail should be filtered out or not.
type CabinetFilterFunc func(CabinetGroupDetail, CabinetDetail) bool

// CabinetKindFilter returns true when a CabinetGroupDetail is of the specified kind.
// For example, CabinetKindSelector(CabinetKindRiver) would match for river cabinets.
func CabinetKindFilter(kind CabinetKind) CabinetFilterFunc {
	return func(groupDetail CabinetGroupDetail, cabinetDetail CabinetDetail) bool {
		return groupDetail.Kind == kind
	}
}

// CabinetClassFilter returns true when a CabinetGroupDetail is of the specified kind.
// For example, CabinetClassFilter(sls_common.ClassRiver) would match for river cabinets.
func CabinetClassFilter(expectedClass sls_common.CabinetType) CabinetFilterFunc {
	return func(groupDetail CabinetGroupDetail, cabinetDetail CabinetDetail) bool {
		class, _ := groupDetail.Kind.Class()
		return class == expectedClass
	}
}

// CabinetAirCooledChassisCountFilter returns true when a CabinetDetail has a matching number of air-cooled
// chassis in a ChassisCount structure.
func CabinetAirCooledChassisCountFilter(airCooledChassisCount int) CabinetFilterFunc {
	return func(groupDetail CabinetGroupDetail, cabinetDetail CabinetDetail) bool {
		if cabinetDetail.ChassisCount != nil {
			return cabinetDetail.ChassisCount.AirCooled == airCooledChassisCount
		}

		return false
	}
}

// CabinetLiquidCooledChassisCountFilter returns true when a CabinetDetail has a matching number of liquid-cooled
// chassis in a ChassisCount structure.
func CabinetLiquidCooledChassisCountFilter(liquidCooledChassisCount int) CabinetFilterFunc {
	return func(groupDetail CabinetGroupDetail, cabinetDetail CabinetDetail) bool {
		if cabinetDetail.ChassisCount != nil {
			return cabinetDetail.ChassisCount.LiquidCooled == liquidCooledChassisCount
		}

		return false
	}
}

// AndCabinetFilter allows for multiple cabinets filters to be chained together, and all must pass.
func AndCabinetFilter(cabinetFilters ...CabinetFilterFunc) CabinetFilterFunc {
	return func(groupDetail CabinetGroupDetail, cabinetDetail CabinetDetail) bool {

		// Loop through the cabinet filters in the order they were provided an perform the test
		for _, cabinetFilter := range cabinetFilters {
			if !cabinetFilter(groupDetail, cabinetDetail) {
				// This filter does not match
				return false
			}
		}

		// All of the filters have passed, this must be a match!
		return true
	}
}

// OrCabinetFilter allows for multiple cabinets filters to be chained together, and all must pass.
func OrCabinetFilter(cabinetFilters ...CabinetFilterFunc) CabinetFilterFunc {
	return func(groupDetail CabinetGroupDetail, cabinetDetail CabinetDetail) bool {

		// Loop through the cabinet filters in the order they were provided and perform the test
		for _, cabinetFilter := range cabinetFilters {
			if cabinetFilter(groupDetail, cabinetDetail) {
				// This filter matches!
				return true
			}
		}

		// None of the filters have passed, this is not a match.
		return false
	}
}
