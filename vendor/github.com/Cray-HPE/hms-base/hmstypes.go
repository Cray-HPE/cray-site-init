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

package base

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Use HMS-wrapped errors.  Subsequent errors will be children of this one.
var e = NewHMSError("hms", "GenericError")

//
// HMS Component type.  This is the top-level classification.  There may be
// subtypes as required by HMS.  This is above the Redfish type and there is
// no particular relationship to anything in Redfish (or whatever else at a
// lower level) implied.
//
// 1.0.0
//

type HMSType string

// This is an enum (though they have no numeric values in ReST and we don't
// want to store them that way since we could get bit by renumbering.
// the down-side is that we will have to normalize capitalization
// differences.  This isn't a huge problem since the type field should only
// be modified by discovery.  It shouldn't be changed by the user, since
// it must map to a corresponding cname and we can't change its hierarchical
// relationship.
const (
	CDU                      HMSType = "CDU"                      // dD
	CDUMgmtSwitch            HMSType = "CDUMgmtSwitch"            // dDwW
	CabinetCDU               HMSType = "CabinetCDU"               // xXdD
	Cabinet                  HMSType = "Cabinet"                  // xX
	CabinetBMC               HMSType = "CabinetBMC"               // xXbB
	CabinetPDUController     HMSType = "CabinetPDUController"     // xXmM
	CabinetPDU               HMSType = "CabinetPDU"               // xXmMpP
	CabinetPDUNic            HMSType = "CabinetPDUNic"            // xXmMiI
	CabinetPDUOutlet         HMSType = "CabinetPDUOutlet"         // xXmMpPjJ DEPRECATED
	CabinetPDUPowerConnector HMSType = "CabinetPDUPowerConnector" // xXmMpPvV

	Chassis                  HMSType = "Chassis"                  // xXcC
	ChassisBMC               HMSType = "ChassisBMC"               // xXcCbB
	ChassisBMCNic            HMSType = "ChassisBMCNic"            // xXcCbBiI
	CMMRectifier             HMSType = "CMMRectifier"             // xXcCtT
	CMMFpga                  HMSType = "CMMFpga"                  // xXcCfF
	CEC                      HMSType = "CEC"                      // xXeE
	ComputeModule            HMSType = "ComputeModule"            // xXcCsS
	RouterModule             HMSType = "RouterModule"             // xXcCrR
	NodeBMC                  HMSType = "NodeBMC"                  // xXcCsSbB
	NodeBMCNic               HMSType = "NodeBMCNic"               // xXcCsSbBiI
	NodeEnclosure            HMSType = "NodeEnclosure"            // xXcCsSeE
	NodeEnclosurePowerSupply HMSType = "NodeEnclosurePowerSupply" // xXcCsSeEtT
	NodePowerConnector       HMSType = "NodePowerConnector"       // xXcCsSjJ
	Node                     HMSType = "Node"                     // xXcCsSbBnN
	Processor                HMSType = "Processor"                // xXcCsSbBnNpP
	StorageGroup             HMSType = "StorageGroup"             // xXcCsSbBnNgG
	Drive                    HMSType = "Drive"                    // xXcCsSbBnNgGkK
	NodeNic                  HMSType = "NodeNic"                  // xXcCsSbBnNiI
	NodeHsnNic               HMSType = "NodeHsnNic"               // xXcCsSbBnNhH
	Memory                   HMSType = "Memory"                   // xXcCsSbBnNdD
	NodeAccel                HMSType = "NodeAccel"                // xXcCsSbBnNaA
	NodeAccelRiser           HMSType = "NodeAccelRiser"           // xXcCsSbBnNrR
	NodeFpga                 HMSType = "NodeFpga"                 // xXcCsSbBfF
	HSNAsic                  HMSType = "HSNAsic"                  // xXcCrRaA
	RouterFpga               HMSType = "RouterFpga"               // xXcCrRfF
	RouterTORFpga            HMSType = "RouterTORFpga"            // xXcCrRtTfF
	RouterBMC                HMSType = "RouterBMC"                // xXcCrRbB
	RouterBMCNic             HMSType = "RouterBMCNic"             // xXcCrRbBiI
	RouterPowerConnector     HMSType = "RouterPowerConnector"     // xXcCrRvV

	HSNBoard              HMSType = "HSNBoard"              // xXcCrReE
	HSNLink               HMSType = "HSNLink"               // xXcCrRaAlL
	HSNConnector          HMSType = "HSNConnector"          // xXcCrRjJ
	HSNConnectorPort      HMSType = "HSNConnectorPort"      // xXcCrRjJpP
	MgmtSwitch            HMSType = "MgmtSwitch"            // xXcCwW
	MgmtHLSwitchEnclosure HMSType = "MgmtHLSwitchEnclosure" // xXcChH
	MgmtHLSwitch          HMSType = "MgmtHLSwitch"          // xXcChHsS
	MgmtSwitchConnector   HMSType = "MgmtSwitchConnector"   // xXcCwWjJ

	// Special types and wildcards
	SMSBox         HMSType = "SMSBox"    // smsN
	Partition      HMSType = "Partition" // pH.S
	System         HMSType = "System"    // s0
	HMSTypeAll     HMSType = "All"       // all
	HMSTypeAllComp HMSType = "AllComp"   // all_comp
	HMSTypeAllSvc  HMSType = "AllSvc"    // all_svc
	HMSTypeInvalid HMSType = "INVALID"   // Not a valid type/xname
)

var ErrHMSTypeInvalid = e.NewChild("got HMSTypeInvalid instead of valid type")
var ErrHMSTypeUnsupported = e.NewChild("HMSType value not supported for this operation")

type hmsCompRecognitionEntry struct {
	Type       HMSType
	ParentType HMSType
	Regex      *regexp.Regexp
	GenStr     string
	NumArgs    int
}

// Component recognition table keyed by normalized (i.e. all lowercase)
// component name.
var hmsCompRecognitionTable = map[string]hmsCompRecognitionEntry{
	"invalid": {
		HMSTypeInvalid,
		HMSTypeInvalid,
		regexp.MustCompile("INVALID"),
		"INVALID",
		0,
	},
	"hmstypeall": {
		HMSTypeAll,
		HMSTypeInvalid,
		regexp.MustCompile("^all$"),
		"all",
		0,
	},
	"hmstypeallsvc": {
		HMSTypeAllSvc,
		HMSTypeInvalid,
		regexp.MustCompile("^all_svc$"),
		"all_svc",
		0,
	},
	"hmstypeallcomp": {
		HMSTypeAllComp,
		HMSTypeInvalid,
		regexp.MustCompile("^all_comp$"),
		"all_comp",
		0,
	},
	"partition": {
		Partition,
		HMSTypeInvalid,
		regexp.MustCompile("^p([0-9]+)(.([0-9]+))?$"),
		"p%d.%d",
		2,
	},
	"system": {
		System,
		HMSTypeInvalid,
		regexp.MustCompile("^s0$"),
		"s0",
		0,
	},
	"smsbox": {
		SMSBox,
		HMSTypeInvalid,
		regexp.MustCompile("^sms([0-9]+)$"),
		"sms%d",
		1,
	},
	"cdu": {
		CDU,
		HMSTypeInvalid, //TODO: what's the CDU's parent? System, right?
		regexp.MustCompile("^d([0-9]+)$"),
		"d%d",
		1,
	},
	"cdumgmtswitch": {
		CDUMgmtSwitch,
		CDU,
		regexp.MustCompile("^d([0-9]+)w([0-9]+)$"),
		"d%dw%d",
		2,
	},
	"cabinetcdu": {
		CabinetCDU,
		Cabinet,
		regexp.MustCompile("^x([0-9]{1,4})d([0-1])$"),
		"x%dd%d",
		2,
	},
	"cabinetpducontroller": {
		CabinetPDUController,
		Cabinet,
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])$"),
		"x%dm%d",
		2,
	},
	"cabinetpdu": {
		CabinetPDU,
		CabinetPDUController,
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])p([0-7])$"),
		"x%dm%dp%d",
		3,
	},
	"cabinetpdunic": {
		CabinetPDUNic,
		CabinetPDUController,
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])i([0-3])$"),
		"x%dm%dp%di%d",
		3,
	},
	"cabinetpduoutlet": {
		CabinetPDUOutlet,
		CabinetPDU,
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])p([0-7])j([1-9][0-9]*)$"),
		"x%dm%dp%dj%d",
		4,
	},
	"cabinetpdupowerconnector": {
		CabinetPDUPowerConnector,
		CabinetPDU,
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])p([0-7])v([1-9][0-9]*)$"),
		"x%dm%dp%dv%d",
		4,
	},
	"cec": {
		CEC,
		Cabinet,
		regexp.MustCompile("^x([0-9]{1,4})e([0-1])$"),
		"x%de%d",
		2,
	},
	"cabinet": {
		Cabinet,
		System,
		regexp.MustCompile("^x([0-9]{1,4})$"),
		"x%d",
		1,
	},
	"cabinetbmc": {
		CabinetBMC,
		Cabinet,
		regexp.MustCompile("^x([0-9]{1,4})b([0])$"),
		"x%db%d",
		2,
	},
	"chassis": {
		Chassis,
		Cabinet,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])$"),
		"x%dc%d",
		2,
	},
	"chassisbmc": {
		ChassisBMC,
		Chassis,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])b([0])$"),
		"x%dc%db%d",
		3,
	},
	"chassisbmcnic": {
		ChassisBMCNic,
		ChassisBMC,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])b([0])i([0-3])$"),
		"x%dc%db%di%d",
		3,
	},
	"cmmfpga": {
		CMMFpga,
		Chassis,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])f([0])$"),
		"x%dc%df%d",
		3,
	},
	"cmmrectifier": {
		CMMRectifier,
		Chassis,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])t([0-9])$"),
		"x%dc%dt%d",
		3,
	},
	"computemodule": {
		ComputeModule,
		Chassis,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)$"),
		"x%dc%ds%d",
		3,
	},
	"storagegroup": {
		StorageGroup,
		Node,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)g([0-9]+)$"),
		"x%dc%ds%db%dn%dg%d",
		6,
	},
	"drive": {
		Drive,
		StorageGroup,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)g([0-9]+)k([0-9]+)$"),
		"x%dc%ds%db%dn%dg%dk%d",
		7,
	},
	"nodefpga": {
		NodeFpga,
		NodeEnclosure,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)f([0])$"),
		"x%dc%ds%db%df%d",
		5,
	},
	"nodebmc": {
		NodeBMC,
		ComputeModule,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)$"),
		"x%dc%ds%db%d",
		4,
	},
	"nodebmcnic": {
		NodeBMCNic,
		NodeBMC,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)i([0-3])$"),
		"x%dc%ds%db%di%d",
		5,
	},
	"nodeenclosure": {
		NodeEnclosure,
		ComputeModule,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)e([0-9]+)$"),
		"x%dc%ds%de%d",
		4,
	},
	"nodeenclosurepowersupply": {
		NodeEnclosurePowerSupply,
		NodeEnclosure,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)e([0-9]+)t([0-9]+)$"),
		"x%dc%ds%de%dt%d",
		5,
	},
	"nodepowerconnector": { //'j' is deprecated, should be 'v'
		NodePowerConnector,
		ComputeModule,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)[jv]([1-2])$"),
		"x%dc%ds%dv%d",
		4,
	},
	"hsnboard": {
		HSNBoard,
		RouterBMC,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)e([0-9]+)$"),
		"x%dc%dr%de%d",
		4,
	},
	"node": {
		Node,
		NodeBMC, // Controlling entity is an nC or COTS BMC
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)$"),
		"x%dc%ds%db%dn%d",
		5,
	},
	"nodenic": {
		NodeNic,
		Node,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)i([0-3])$"),
		"x%dc%ds%db%dn%di%d",
		6,
	},
	"nodehsnnic": {
		NodeHsnNic,
		Node,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)h([0-3])$"),
		"x%dc%ds%db%dn%dh%d",
		6,
	},
	"nodeaccel": {
		NodeAccel,
		Node,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)a([0-9]+)$"),
		"x%dc%ds%db%dn%da%d",
		6,
	},
	"nodeaccelriser": {
		NodeAccelRiser,
		Node,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)r([0-7])$"),
		"x%dc%ds%db%dn%dr%d",
		6,
	},
	"memory": {
		Memory,
		Node, //parent is actually a socket but we'll use node
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)d([0-9]+)$"),
		"x%dc%ds%db%dn%dd%d",
		6,
	},
	"processor": {
		Processor,
		Node,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)p([0-3])$"),
		"x%dc%ds%db%dn%dp%d",
		6,
	},
	"routermodule": {
		RouterModule,
		Chassis,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)$"),
		"x%dc%dr%d",
		3,
	},
	"routerfpga": {
		RouterFpga,
		RouterModule,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)f([01])$"),
		"x%dc%dr%df%d",
		4,
	},
	"routertorfpga": {
		RouterTORFpga,
		RouterModule,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)t([0-9]+)f([0-1])$"),
		"x%dc%dr%dt%df%d",
		5,
	},
	"routerbmc": {
		RouterBMC,
		RouterModule,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)b([0-9]+)$"),
		"x%dc%dr%db%d",
		4,
	},
	"routerbmcnic": {
		RouterBMCNic,
		RouterBMC,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)b([0-9]+)i([0-3])$"),
		"x%dc%dr%db%di%d",
		5,
	},
	"routerpowerconnector": {
		RouterPowerConnector,
		RouterModule,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)v([1-2])$"),
		"x%dc%dr%dv%d",
		4,
	},
	"hsnasic": {
		HSNAsic,
		RouterModule,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)a([0-3])$"),
		"x%dc%dr%da%d",
		4,
	},
	"hsnconnector": {
		HSNConnector,
		RouterModule,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)j([1-9][0-9]*)$"),
		"x%dc%dr%dj%d",
		4,
	},
	"hsnconnectorport": {
		HSNConnectorPort,
		HSNConnector,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)j([1-9][0-9]*)p([012])$"),
		"x%dc%dr%dj%dp%d",
		5,
	},
	"hsnlink": {
		HSNLink,
		HSNAsic,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)a([0-3])l([0-9]+)$"),
		"x%dc%dr%da%dl%d",
		5,
	},
	"mgmtswitch": {
		MgmtSwitch,
		Chassis,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])w([1-9][0-9]*)$"),
		"x%dc%dw%d",
		3,
	},
	"mgmtswitchconnector": {
		MgmtSwitchConnector,
		MgmtSwitch,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])w([1-9][0-9]*)j([1-9][0-9]*)$"),
		"x%dc%dw%dj%d",
		4,
	},
	"mgmthlswitchenclosure": {
		MgmtHLSwitchEnclosure,
		Chassis,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])h([1-9][0-9]*)$"),
		"x%dc%dh%d",
		3,
	},
	"mgmthlswitch": {
		MgmtHLSwitch,
		MgmtHLSwitchEnclosure,
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])h([1-9][0-9]*)s([1-9])$"),
		"x%dc%dh%ds%d",
		4,
	},
	//	Module:  {  // ComputeModule or RouterModule.  Can we support this?
	//		Module,hss/base hss/smd hss/sm hss/hmsds
	//		Chassis,
	//		"^x([0-9]{1,4})c([0-7])([sr])([0-9]+)$",
	//		"x%dc%d%s%d"
	//	},
}


// Get the HMSType for a given xname, based on its pattern in the recognition
// table above.
// If no string matches, HMSTypeInvalid is returned.
func GetHMSType(xname string) HMSType {
	for _, compEntry := range hmsCompRecognitionTable {
		if compEntry.Regex.MatchString(xname) {
			return compEntry.Type
		}
	}
	return HMSTypeInvalid
}

// Get the HMSType for a given xname, based on its pattern in the recognition
// table above (string version).
// If no type matches, the empty string is returned.
func GetHMSTypeString(xname string) string {
	hmsType := GetHMSType(xname)
	if hmsType == HMSTypeInvalid {
		return ""
	}
	return hmsType.String()
}

func GetHMSTypeList() []string {
	hmsTypeList := []string{}
	for _, compEntry := range hmsCompRecognitionTable {
		hmsTypeList = append(hmsTypeList, compEntry.Type.String())
	}
	return hmsTypeList
}

// Returns whether hmsType is a controller type, i.e. that
// would host a Redfish entry point
func IsHMSTypeController(hmsType HMSType) bool {
	switch hmsType {
	case ChassisBMC:
		fallthrough
	case RouterBMC:
		fallthrough
	case NodeBMC:
		fallthrough
	case CabinetPDUController:
		return true
	default:
		return false
	}
}

// Returns whether hmsTypeStr matches a controller type, i.e. that
// would host a Redfish entry point
func IsHMSTypeStrController(hmsTypeStr string) bool {
	hmsType := ToHMSType(hmsTypeStr)
	return IsHMSTypeController(hmsType)
}

// Returns whether hmsType does not expect System collections to be
// present, or if so, non-empty.
func ControllerHasSystems(hmsType HMSType) bool {
	switch hmsType {
	case NodeBMC:
		return true
	default:
		return false
	}
}

// String version of above.
func ControllerHasSystemsStr(hmsTypeStr string) bool {
	hmsType := ToHMSType(hmsTypeStr)
	return ControllerHasSystems(hmsType)
}

// Normally every controller should has 1+ Chassis, but certain special
// types may not, and we don't use them if even if they are there.
func ControllerHasChassis(hmsType HMSType) bool {
	switch hmsType {
	case CabinetPDUController:
		return false
	default:
		return true
	}
}

// String version of above.
func ControllerHasChassisStr(hmsTypeStr string) bool {
	hmsType := ToHMSType(hmsTypeStr)
	return ControllerHasChassis(hmsType)
}

// Returns whether hmsType is a container type, i.e. that
// contains other components but is not a logical type (i.e. Node,
// that may or may not be a discrete hardware module), or a
// specialized terminal subcomponent (HSNAsic, Processor, etc.)
//
// NOTE: This is used specifically for Redfish "Chassis" components.
// so non-Chassis (e.g. PDUs) don't apply (different type, outlets are
// not really separate physical pieces)
func IsHMSTypeContainer(hmsType HMSType) bool {
	switch hmsType {
	case Cabinet:
		fallthrough
	case Chassis:
		fallthrough
	case ComputeModule:
		fallthrough
	case RouterModule:
		fallthrough
	case NodeEnclosure:
		return true
	case HSNBoard:
		return true
	default:
		return false
	}
}

// Returns whether hmsTypeStr matches a container type, i.e. that
// contains other components but is not a logical type (i.e. Node,
// that may or may not be a discrete hardware module), or a
// specialized terminal subcomponent (HSNAsic, Processor, etc.)
func IsHMSTypeStrContainer(hmsTypeStr string) bool {
	hmsType := ToHMSType(hmsTypeStr)
	return IsHMSTypeContainer(hmsType)
}

// Returns the given HMS type string (adjusting any capitalization differences),
// if a valid HMS component type was given.  Else, return the empty string.
func VerifyNormalizeType(typeStr string) string {
	typeLower := strings.ToLower(typeStr)
	value, ok := hmsCompRecognitionTable[typeLower]
	if ok != true {
		return ""
	} else {
		return value.Type.String()
	}
}

// Returns the given HMSType (adjusting any capitalization differences),
// if a valid HMS component type string was given. Else, return HMSTypeINVALID.
func ToHMSType(typeStr string) HMSType {
	typeLower := strings.ToLower(typeStr)
	value, ok := hmsCompRecognitionTable[typeLower]
	if ok != true {
		return HMSTypeInvalid
	} else {
		return value.Type
	}
}

// TODO
func GetHMSTypeFormatString(hmsType HMSType) (string, int, error) {
	typeLower := strings.ToLower(hmsType.String())
	if value, ok := hmsCompRecognitionTable[typeLower]; ok {
		return value.GenStr, value.NumArgs, nil
	}
	
	return "", 0, fmt.Errorf("unknown HMSType: %s", typeLower)
}

// Allow HMSType to be treated as a standard string type.
func (t HMSType) String() string { return string(t) }

//
// State field used in component, set in response to events by state manager.
// 1.0.0
//
type HMSState string

// Valid state values for components - should refect hardware state
// Enabled/Disabled is a separate boolean field, as the component should
// still have it's actual physical state known and tracked at all times, so
// we know what it is when it is enabled.  It also avoids the primary case
// where admins need to modify the state field manually.
//
// NOTE: there will be no state between on and ready.  If the managed plane
// software does not have heartbeats, On is as high as it will ever get.
// So "active" is not useful.   'Paused' is not in scope now that the software
// status field exists.
const (
	StateUnknown   HMSState = "Unknown"   // The State is unknown.  Appears missing but has not been confirmed as empty.
	StateEmpty     HMSState = "Empty"     // The location is not populated with a component
	StatePopulated HMSState = "Populated" // Present (not empty), but no further track can or is being done.
	StateOff       HMSState = "Off"       // Present but powered off
	StateOn        HMSState = "On"        // Powered on.  If no heartbeat mechanism is available, it's software state may be unknown.

	StateStandby HMSState = "Standby" // No longer Ready and presumed dead.  It typically means HB has been lost (w/alert).
	StateHalt    HMSState = "Halt"    // No longer Ready and halted.  OS has been gracefully shutdown or panicked (w/ alert).
	StateReady   HMSState = "Ready"   // Both On and Ready to provide its expected services, i.e. used for jobs.

	//  Retired (actually never used) states:
	// StateActive    HMSState = "Active"    // If level-2 systems without hb monitoring can make a distinction between on and booting/booted.
	// StatePaused    HMSState = "Paused"    // Was in a Ready state, but is temporarily unavailable due to admin action or a transient issue.
)

var ErrHMSStateInvalid = e.NewChild("was not a valid HMS state")
var ErrHMSStateUnsupported = e.NewChild("HMSState value not supported for this operation")
var ErrHMSNeedForce = e.NewChild("operation not allowed and not forced.")

// For case-insensitive verification and normalization of state strings
var hmsStateMap = map[string]HMSState{
	"unknown":   StateUnknown,
	"empty":     StateEmpty,
	"populated": StatePopulated,
	"off":       StateOff,
	"on":        StateOn,
	"standby":   StateStandby,
	"halt":      StateHalt,
	"ready":     StateReady,
}

func GetHMSStateList() []string {
	hmsStateList := []string{}
	for _, state := range hmsStateMap {
		hmsStateList = append(hmsStateList, state.String())
	}
	return hmsStateList
}

// Returns the given state string (adjusting any capitalization differences),
// if a valid state is given.  Else, return the empty string.
func VerifyNormalizeState(stateStr string) string {
	stateLower := strings.ToLower(stateStr)
	value, ok := hmsStateMap[stateLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// Specifies valid STARTING states before changing to the indicated state,
// at least without forcing the change, which would normally be a bad idea.
// An empty array means "None without forcing.
var hmsValidStartStatesMap = map[string][]string{
	"unknown":   []string{}, // Force/HSM only
	"empty":     []string{}, // Force/HSM only
	"populated": []string{}, // Force/HSM only
	"off":       []string{string(StateOff), string(StateOn), string(StateStandby), string(StateHalt), string(StateReady)},
	"on":        []string{string(StateOn), string(StateOff), string(StateStandby), string(StateHalt)},
	"standby":   []string{string(StateStandby), string(StateReady)},
	"halt":      []string{string(StateHalt), string(StateReady)},
	"ready":     []string{string(StateReady), string(StateOn), string(StateOff), string(StateStandby), string(StateHalt)}, // Last three are needed (for now) if RF events break.
}

// If ok == true, beforeStates contain valid current states a
// component can be in if it is being transitioned to afterState without
// being forced (either because it is a bad idea, or the state should
// only be set by HSM and not by other software).  An empty array means 'None
// without force=true
//
// If ok == false, afterState matched no valid HMS State (case insensitive)
func GetValidStartStates(afterState string) (beforeStates []string, ok bool) {
	stateLower := strings.ToLower(afterState)
	beforeStates, ok = hmsValidStartStatesMap[stateLower]
	return
}

// Same as above, but with force flag.  If not found, returns
// ErrHMSStateInvalid.  If can only be forced, and force = false,
// error will be ErrHMSNeedForce.   Otherwise list of starting states.
// If force = true and no errors, an empty array means no restrictions.
func GetValidStartStateWForce(
	afterState string,
	force bool,
) (beforeStates []string, err error) {

	beforeStates = []string{}
	// See if transition is valid.
	if force == false {
		var ok bool
		beforeStates, ok = GetValidStartStates(afterState)
		if !ok {
			err = ErrHMSStateInvalid
		} else if len(beforeStates) == 0 {
			err = ErrHMSNeedForce
		}
	}
	return
}

// Check to see if the state is above on (on is the highest we will get
// from Redfish, so these are state set by higher software layers)
func IsPostBootState(stateStr string) bool {
	stateLower := strings.ToLower(stateStr)
	value, ok := hmsStateMap[stateLower]
	if ok != true {
		return false
	} else {
		switch value {
		//case StateActive:
		//	fallthrough
		case StateStandby:
			fallthrough
		case StateHalt:
			fallthrough
		case StateReady:
			return true
		//case StatePaused:
		//	return true
		default:
			return false
		}
	}
}

// Allow HMSState to be treated as a standard string type.
func (s HMSState) String() string { return string(s) }

//
// Flag field used in component, set in response to events by state manager.
// 1.0.0
//

type HMSFlag string

// Valid flag values.
const (
	FlagUnknown HMSFlag = "Unknown"
	FlagOK      HMSFlag = "OK"      // Functioning properly
	FlagWarning HMSFlag = "Warning" // Continues to operate, but has an issue that may require attention.
	FlagAlert   HMSFlag = "Alert"   // No longer operating as expected.  The state may also have changed due to error.
	FlagLocked  HMSFlag = "Locked"  // Another service has reserved this component.
)

// For case-insensitive verification and normalization of flag strings
var hmsFlagMap = map[string]HMSFlag{
	"unknown": FlagUnknown,
	"ok":      FlagOK,
	"warning": FlagWarning,
	"warn":    FlagWarning,
	"alert":   FlagAlert,
	"locked":  FlagLocked,
}

// Get a list of all valid HMS flags
func GetHMSFlagList() []string {
	hmsFlagList := []string{}
	for _, flag := range hmsFlagMap {
		hmsFlagList = append(hmsFlagList, flag.String())
	}
	return hmsFlagList
}

// Returns the given flag string (adjusting any capitalization differences),
// if a valid flag was given.  Else, return the empty string.
func VerifyNormalizeFlag(flagStr string) string {
	flagLower := strings.ToLower(flagStr)
	value, ok := hmsFlagMap[flagLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// As above, but if flag is the empty string, then return FlagOK.
// If non-empty and invalid, return the empty string.
func VerifyNormalizeFlagOK(flag string) string {
	if flag == "" {
		return FlagOK.String()
	}
	return VerifyNormalizeFlag(flag)
}

// Allow HMSFlag to be treated as a standard string type.
func (f HMSFlag) String() string { return string(f) }

//
// Role of component
// 1.0.0
//

type HMSRole string

// Valid role values.
const (
	RoleCompute     HMSRole = "Compute"
	RoleService     HMSRole = "Service"
	RoleSystem      HMSRole = "System"
	RoleApplication HMSRole = "Application"
	RoleStorage     HMSRole = "Storage"
	RoleManagement  HMSRole = "Management"
)

// For case-insensitive verification and normalization of role strings
var defaultHMSRoleMap = map[string]string{
	"compute":     RoleCompute.String(),
	"service":     RoleService.String(),
	"system":      RoleSystem.String(),
	"application": RoleApplication.String(),
	"storage":     RoleStorage.String(),
	"management":  RoleManagement.String(),
}

var hmsRoleMap = defaultHMSRoleMap

// Get a list of all valid HMS roles
func GetHMSRoleList() []string {
	hmsRoleList := []string{}
	for _, role := range hmsRoleMap {
		hmsRoleList = append(hmsRoleList, role)
	}
	return hmsRoleList
}

// Returns the given role string (adjusting any capitalization differences),
// if a valid role was given.  Else, return the empty string.
func VerifyNormalizeRole(roleStr string) string {
	roleLower := strings.ToLower(roleStr)
	value, ok := hmsRoleMap[roleLower]
	if ok != true {
		return ""
	} else {
		return value
	}
}

// Allow HMSRole to be treated as a standard string type.
func (r HMSRole) String() string { return string(r) }

//
// SubRole of component
// 1.0.0
//

type HMSSubRole string

// Valid SubRole values.
const (
	SubRoleMaster  HMSSubRole = "Master"
	SubRoleWorker  HMSSubRole = "Worker"
	SubRoleStorage HMSSubRole = "Storage"
)

// For case-insensitive verification and normalization of SubRole strings
var defaultHMSSubRoleMap = map[string]string{
	"master":  SubRoleMaster.String(),
	"worker":  SubRoleWorker.String(),
	"storage": SubRoleStorage.String(),
}

var hmsSubRoleMap = defaultHMSSubRoleMap

// Get a list of all valid HMS subroles
func GetHMSSubRoleList() []string {
	hmsSubRoleList := []string{}
	for _, subrole := range hmsSubRoleMap {
		hmsSubRoleList = append(hmsSubRoleList, subrole)
	}
	return hmsSubRoleList
}

// Returns the given SubRole string (adjusting any capitalization differences),
// if a valid SubRole was given.  Else, return the empty string.
func VerifyNormalizeSubRole(subRoleStr string) string {
	subRoleLower := strings.ToLower(subRoleStr)
	value, ok := hmsSubRoleMap[subRoleLower]
	if ok != true {
		return ""
	} else {
		return value
	}
}

// Allow HMSSubRole to be treated as a standard string type.
func (r HMSSubRole) String() string { return string(r) }

//
// HMSNetType - type of high speed network
// 1.0.0
//

type HMSNetType string

const (
	NetSling      HMSNetType = "Sling"
	NetInfiniband HMSNetType = "Infiniband"
	NetEthernet   HMSNetType = "Ethernet"
	NetOEM        HMSNetType = "OEM" // Placeholder for non-slingshot
	NetNone       HMSNetType = "None"
)

// For case-insensitive verification and normalization of HSN network types
var hmsNetTypeMap = map[string]HMSNetType{
	"sling":      NetSling,
	"infiniband": NetInfiniband,
	"ethernet":   NetEthernet,
	"oem":        NetOEM,
	"none":       NetNone,
}

// Get a list of all valid HMS NetTypes
func GetHMSNetTypeList() []string {
	hmsNetTypeList := []string{}
	for _, netType := range hmsNetTypeMap {
		hmsNetTypeList = append(hmsNetTypeList, netType.String())
	}
	return hmsNetTypeList
}

// Returns the given net type string (adjusting any capitalization differences),
// if a valid netType was given.  Else, return the empty string.
func VerifyNormalizeNetType(netTypeStr string) string {
	netTypeLower := strings.ToLower(netTypeStr)
	value, ok := hmsNetTypeMap[netTypeLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// Allow HMSNetType to be treated as a standard string type.
func (r HMSNetType) String() string { return string(r) }

//
// HMSArch - binary type needed for component
// 1.0.0
//

type HMSArch string

const (
	ArchX86     HMSArch = "X86"
	ArchARM     HMSArch = "ARM"
	ArchUnknown HMSArch = "UNKNOWN"
	ArchOther   HMSArch = "Other"
)

// For case-insensitive verification and normalization of HSN network types
var hmsArchMap = map[string]HMSArch{
	"x86":     ArchX86,
	"arm":     ArchARM,
	"unknown": ArchUnknown,
	"other":   ArchOther,
}

// Get a list of all valid HMS Arch
func GetHMSArchList() []string {
	hmsArchList := []string{}
	for _, arch := range hmsArchMap {
		hmsArchList = append(hmsArchList, arch.String())
	}
	return hmsArchList
}

// Returns the given arch string (adjusting any capitalization differences),
// if a valid arch was given.  Else, return the empty string.
func VerifyNormalizeArch(archStr string) string {
	archLower := strings.ToLower(archStr)
	value, ok := hmsArchMap[archLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// Allow HMSArch to be treated as a standard string type.
func (r HMSArch) String() string { return string(r) }

//
// HMSClass - Physical hardware profile
// 1.0.0
//

type HMSClass string

const (
	ClassRiver    HMSClass = "River"
	ClassMountain HMSClass = "Mountain"
	ClassHill     HMSClass = "Hill"
)

// For case-insensitive verification and normalization of HMS Class
var hmsClassMap = map[string]HMSClass{
	"river":    ClassRiver,
	"mountain": ClassMountain,
	"hill":     ClassHill,
}

// Get a list of all valid HMS Class
func GetHMSClassList() []string {
	hmsClassList := []string{}
	for _, class := range hmsClassMap {
		hmsClassList = append(hmsClassList, class.String())
	}
	return hmsClassList
}

// Returns the given class string (adjusting any capitalization differences),
// if a valid class was given.  Else, return the empty string.
func VerifyNormalizeClass(classStr string) string {
	classLower := strings.ToLower(classStr)
	value, ok := hmsClassMap[classLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// Allow HMSClass to be treated as a standard string type.
func (r HMSClass) String() string { return string(r) }

//
// This is the equivalent to rs_node_t in Cascade.  It is the minimal
// amount of of information for tracking component state and other vital
// info at an abstract level.  The hwinv is for component-type specific
// fields and detailed HW attributes, i.e. just like XC.
//
// For most HMS operations (and non-inventory ones in the managed plane)
// this info should be sufficient.  We want to keep it minimal for speed.
// Those fields that are not fixed at discovery should be those that can
// change outside of discovery in response to system activity, i.e. hwinv
// should contain only fields that are basically static between discoveries
// of the endpoint.   Things like firmware versions might be an exception,
// but that would be a separate process SM would
//
// 1.0.0
//
type Component struct {
	ID                  string      `json:"ID"`
	Type                string      `json:"Type"`
	State               string      `json:"State,omitempty"`
	Flag                string      `json:"Flag,omitempty"`
	Enabled             *bool       `json:"Enabled,omitempty"`
	SwStatus            string      `json:"SoftwareStatus,omitempty"`
	Role                string      `json:"Role,omitempty"`
	SubRole             string      `json:"SubRole,omitempty"`
	NID                 json.Number `json:"NID,omitempty"`
	Subtype             string      `json:"Subtype,omitempty"`
	NetType             string      `json:"NetType,omitempty"`
	Arch                string      `json:"Arch,omitempty"`
	Class               string      `json:"Class,omitempty"`
	ReservationDisabled bool        `json:"ReservationDisabled,omitempty"`
	Locked              bool        `json:"Locked,omitempty"`
}

// A collection of 0-n Components.  It could just be an ordinary
// array but we want to save the option to have indentifying info, etc.
// packaged with it, e.g. the query parameters or options that produced it,
// especially if there are fewer fields than normal being included.
type ComponentArray struct {
	Components []*Component `json:"Components"`
}

// Given a properly formatted xname, get its immediate parent.
//  i.e. x0c0s22b11 would become x0c0s22
func GetHMSCompParent(xname string) string {
	// Trim all trailing numbers, then in the result, trim all trailing
	// letters.
	pstr := strings.TrimRightFunc(xname,
		func(r rune) bool { return unicode.IsNumber(r) })
	pstr = strings.TrimRightFunc(pstr,
		func(r rune) bool { return unicode.IsLetter(r) })
	return pstr
}

// This lower-cases the xname id does other normalization so we
// always represent the same location with the same string.
// NOTE: this does not validate the xname, so do not use it
// to see if it is valid.  However, any string post
// normalization should still be as invalid or valid as an xname as it
// was prior to this call.
func NormalizeHMSCompID(xname string) string {
	xnameNorm := RemoveLeadingZeros(strings.TrimSpace(xname))
	return strings.ToLower(xnameNorm)
}

// Returns true if xname is valid, i.e. matches some format for a
// valid HMS component type.  False if it is invalid.
func IsHMSCompIDValid(xname string) bool {
	hmsType := GetHMSType(xname)
	if hmsType == HMSTypeInvalid {
		return false
	}
	return true
}

// Returns the given xname ID string (adjusting for any capitalization and
// whitespace and leading zero differences) if a valid xname was given.
// Else, return the empty string.
func VerifyNormalizeCompID(idStr string) string {
	idNorm := NormalizeHMSCompID(idStr)
	if IsHMSCompIDValid(idNorm) == true {
		return idNorm
	} else {
		return ""
	}
}

// ValidateCompIDs validates an array of component IDs (xnames) into
// valid and invalid arrays optionally adding duplicates to the invalid
// array or discarding them.
func ValidateCompIDs(compIDs []string, dupsValid bool) ([]string, []string) {
	seen := make(map[string]bool)
	valid := []string{}
	invalid := []string{}

	for _, compID := range compIDs {
		if _, s := seen[compID]; !s {
			seen[compID] = true
			if IsHMSCompIDValid(compID) {
				valid = append(valid, compID)
			} else {
				invalid = append(invalid, compID)
			}
		} else {
			if !dupsValid {
				invalid = append(invalid, compID)
			}
		}
	}

	return valid, invalid
}
