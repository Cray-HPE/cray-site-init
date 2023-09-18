// MIT License
//
// (C) Copyright 2018-2023 Hewlett Packard Enterprise Development LP
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
package xnametypes

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

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
	VirtualNode              HMSType = "VirtualNode"              // xXcCsSbBnNvV
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
	RouterTOR                HMSType = "RouterTOR"                // xXcCrRtT
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

type HMSCompRecognitionEntry struct {
	Type          HMSType
	ParentType    HMSType
	ExampleString string
	Regex         *regexp.Regexp
	GenStr        string
	NumArgs       int
}

// Component recognition table keyed by normalized (i.e. all lowercase)
// component name.
// WARNING: if you modify this map you MUST regenerate the xnames, see https://github.com/Cray-HPE/hms-xname#code-generation
var hmsCompRecognitionTable = map[string]HMSCompRecognitionEntry{
	"invalid": {
		HMSTypeInvalid,
		HMSTypeInvalid,
		"INVALID",
		regexp.MustCompile("INVALID"),
		"INVALID",
		0,
	},
	"hmstypeall": {
		HMSTypeAll,
		HMSTypeInvalid,
		"all",
		regexp.MustCompile("^all$"),
		"all",
		0,
	},
	"hmstypeallsvc": {
		HMSTypeAllSvc,
		HMSTypeInvalid,
		"all_svc",
		regexp.MustCompile("^all_svc$"),
		"all_svc",
		0,
	},
	"hmstypeallcomp": {
		HMSTypeAllComp,
		HMSTypeInvalid,
		"all_comp",
		regexp.MustCompile("^all_comp$"),
		"all_comp",
		0,
	},
	"partition": {
		Partition,
		HMSTypeInvalid,
		"pH.S",
		regexp.MustCompile("^p([0-9]+)(.([0-9]+))?$"),
		"p%d.%d",
		2,
	},
	"system": {
		System,
		HMSTypeInvalid,
		"sS",
		regexp.MustCompile("^s0$"),
		"s0",
		0,
	},
	"smsbox": {
		SMSBox,
		HMSTypeInvalid,
		"smsN",
		regexp.MustCompile("^sms([0-9]+)$"),
		"sms%d",
		1,
	},
	"cdu": {
		CDU,
		System,
		"dD",
		regexp.MustCompile("^d([0-9]+)$"),
		"d%d",
		1,
	},
	"cdumgmtswitch": {
		CDUMgmtSwitch,
		CDU,
		"dDwW",
		regexp.MustCompile("^d([0-9]+)w([0-9]+)$"),
		"d%dw%d",
		2,
	},
	"cabinetcdu": {
		CabinetCDU,
		Cabinet,
		"xXdD",
		regexp.MustCompile("^x([0-9]{1,4})d([0-1])$"),
		"x%dd%d",
		2,
	},
	"cabinetpducontroller": {
		CabinetPDUController,
		Cabinet,
		"xXmM",
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])$"),
		"x%dm%d",
		2,
	},
	"cabinetpdu": {
		CabinetPDU,
		CabinetPDUController,
		"xXmMpP",
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])p([0-7])$"),
		"x%dm%dp%d",
		3,
	},
	"cabinetpdunic": {
		CabinetPDUNic,
		CabinetPDUController,
		"xXmMpPiI",
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])i([0-3])$"),
		"x%dm%di%d",
		3,
	},
	"cabinetpduoutlet": {
		CabinetPDUOutlet,
		CabinetPDU,
		"xXmMpPjJ",
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])p([0-7])j([1-9][0-9]*)$"),
		"x%dm%dp%dj%d",
		4,
	},
	"cabinetpdupowerconnector": {
		CabinetPDUPowerConnector,
		CabinetPDU,
		"xXmMpPvV",
		regexp.MustCompile("^x([0-9]{1,4})m([0-3])p([0-7])v([1-9][0-9]*)$"),
		"x%dm%dp%dv%d",
		4,
	},
	"cec": {
		CEC,
		Cabinet,
		"xXeE",
		regexp.MustCompile("^x([0-9]{1,4})e([0-1])$"),
		"x%de%d",
		2,
	},
	"cabinet": {
		Cabinet,
		System,
		"xX",
		regexp.MustCompile("^x([0-9]{1,4})$"),
		"x%d",
		1,
	},
	"cabinetbmc": {
		CabinetBMC,
		Cabinet,
		"xXbB",
		regexp.MustCompile("^x([0-9]{1,4})b([0])$"),
		"x%db%d",
		2,
	},
	"chassis": {
		Chassis,
		Cabinet,
		"xXcC",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])$"),
		"x%dc%d",
		2,
	},
	"chassisbmc": {
		ChassisBMC,
		Chassis,
		"xXcCbB",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])b([0])$"),
		"x%dc%db%d",
		3,
	},
	"chassisbmcnic": {
		ChassisBMCNic,
		ChassisBMC,
		"xXcCbBiI",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])b([0])i([0-3])$"),
		"x%dc%db%di%d",
		4,
	},
	"cmmfpga": {
		CMMFpga,
		Chassis,
		"xXcCfF",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])f([0])$"),
		"x%dc%df%d",
		3,
	},
	"cmmrectifier": {
		CMMRectifier,
		Chassis,
		"xXcCtT",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])t([0-9])$"),
		"x%dc%dt%d",
		3,
	},
	"computemodule": {
		ComputeModule,
		Chassis,
		"xXcCsS",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)$"),
		"x%dc%ds%d",
		3,
	},
	"storagegroup": {
		StorageGroup,
		Node,
		"xXcCsSbBnNgG",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)g([0-9]+)$"),
		"x%dc%ds%db%dn%dg%d",
		6,
	},
	"drive": {
		Drive,
		StorageGroup,
		"xXcCsSbBnNgGkK",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)g([0-9]+)k([0-9]+)$"),
		"x%dc%ds%db%dn%dg%dk%d",
		7,
	},
	"nodefpga": {
		NodeFpga,
		NodeEnclosure,
		"xXcCsSbBfF",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)f([0])$"),
		"x%dc%ds%db%df%d",
		5,
	},
	"nodebmc": {
		NodeBMC,
		ComputeModule,
		"xXcCsSbB",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)$"),
		"x%dc%ds%db%d",
		4,
	},
	"nodebmcnic": {
		NodeBMCNic,
		NodeBMC,
		"xXcCsSbBiI",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)i([0-3])$"),
		"x%dc%ds%db%di%d",
		5,
	},
	"nodeenclosure": {
		NodeEnclosure,
		ComputeModule,
		"xXcCsSbBeE",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)e([0-9]+)$"),
		"x%dc%ds%de%d",
		4,
	},
	"nodeenclosurepowersupply": {
		NodeEnclosurePowerSupply,
		NodeEnclosure,
		"xXcCsSbBeEtT",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)e([0-9]+)t([0-9]+)$"),
		"x%dc%ds%de%dt%d",
		5,
	},
	"nodepowerconnector": { // 'j' is deprecated, should be 'v'
		NodePowerConnector,
		ComputeModule,
		"xXcCsSv",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)[jv]([1-2])$"),
		"x%dc%ds%dv%d",
		4,
	},
	"hsnboard": {
		HSNBoard,
		RouterModule,
		"xXcCrReE",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)e([0-9]+)$"),
		"x%dc%dr%de%d",
		4,
	},
	"node": {
		Node,
		NodeBMC, // Controlling entity is an nC or COTS BMC
		"xXcCsSbBnN",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)$"),
		"x%dc%ds%db%dn%d",
		5,
	},
	"virtualnode": {
		VirtualNode,
		Node, // The hypervisor
		"xXcCsSbBnNvV",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)v([0-9]+)$"),
		"x%dc%ds%db%dn%dv%d",
		6,
	},
	"nodenic": {
		NodeNic,
		Node,
		"xXcCsSbBnNiI",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)i([0-3])$"),
		"x%dc%ds%db%dn%di%d",
		6,
	},
	"nodehsnnic": {
		NodeHsnNic,
		Node,
		"xXcCsSbBnNhH",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)h([0-3])$"),
		"x%dc%ds%db%dn%dh%d",
		6,
	},
	"nodeaccel": {
		NodeAccel,
		Node,
		"xXcCsSbBnNaA",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)a([0-9]+)$"),
		"x%dc%ds%db%dn%da%d",
		6,
	},
	"nodeaccelriser": {
		NodeAccelRiser,
		Node,
		"xXcCsSbBnNrR",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)r([0-7])$"),
		"x%dc%ds%db%dn%dr%d",
		6,
	},
	"memory": {
		Memory,
		Node, // Parent is actually a socket but we'll use node
		"xXcCsSbBnNdD",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)d([0-9]+)$"),
		"x%dc%ds%db%dn%dd%d",
		6,
	},
	"processor": {
		Processor,
		Node,
		"xXcCsSbBnNpP",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)p([0-3])$"),
		"x%dc%ds%db%dn%dp%d",
		6,
	},
	"routermodule": {
		RouterModule,
		Chassis,
		"xXcCrR",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)$"),
		"x%dc%dr%d",
		3,
	},
	"routerfpga": {
		RouterFpga,
		RouterModule,
		"xXcCrRfF",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)f([01])$"),
		"x%dc%dr%df%d",
		4,
	},
	"routertor": {
		RouterTOR,
		RouterModule,
		"xXcCrRtT",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)t([0-9]+)$"),
		"x%dc%dr%dt%d",
		4,
	},
	"routertorfpga": {
		RouterTORFpga,
		RouterTOR,
		"xXcCrRtTfF",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)t([0-9]+)f([0-1])$"),
		"x%dc%dr%dt%df%d",
		5,
	},
	"routerbmc": {
		RouterBMC,
		RouterModule,
		"xXcCrRbB",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)b([0-9]+)$"),
		"x%dc%dr%db%d",
		4,
	},
	"routerbmcnic": {
		RouterBMCNic,
		RouterBMC,
		"xXcCrRbBiI",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)b([0-9]+)i([0-3])$"),
		"x%dc%dr%db%di%d",
		5,
	},
	"routerpowerconnector": {
		RouterPowerConnector,
		RouterModule,
		"xXcCrRvV",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)v([1-2])$"),
		"x%dc%dr%dv%d",
		4,
	},
	"hsnasic": {
		HSNAsic,
		RouterModule,
		"xXcCrRaA",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)a([0-3])$"),
		"x%dc%dr%da%d",
		4,
	},
	"hsnconnector": {
		HSNConnector,
		RouterModule,
		"xXcCrRjJ",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)j([1-9][0-9]*)$"),
		"x%dc%dr%dj%d",
		4,
	},
	"hsnconnectorport": {
		HSNConnectorPort,
		HSNConnector,
		"xXcCrRjJpP",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)j([1-9][0-9]*)p([012])$"),
		"x%dc%dr%dj%dp%d",
		5,
	},
	"hsnlink": {
		HSNLink,
		HSNAsic,
		"xXcCrRaAlL",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])r([0-9]+)a([0-3])l([0-9]+)$"),
		"x%dc%dr%da%dl%d",
		5,
	},
	"mgmtswitch": {
		MgmtSwitch,
		Chassis,
		"xXcCwW",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])w([1-9][0-9]*)$"),
		"x%dc%dw%d",
		3,
	},
	"mgmtswitchconnector": {
		MgmtSwitchConnector,
		MgmtSwitch,
		"xXcCwWjJ",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])w([1-9][0-9]*)j([1-9][0-9]*)$"),
		"x%dc%dw%dj%d",
		4,
	},
	"mgmthlswitchenclosure": {
		MgmtHLSwitchEnclosure,
		Chassis,
		"xXcChH",
		regexp.MustCompile("^x([0-9]{1,4})c([0-7])h([1-9][0-9]*)$"),
		"x%dc%dh%d",
		3,
	},
	"mgmthlswitch": {
		MgmtHLSwitch,
		MgmtHLSwitchEnclosure,
		"xXcChHsS",
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

func GetHMSCompRecognitionTable() map[string]HMSCompRecognitionEntry {
	copy := map[string]HMSCompRecognitionEntry{}

	for k, v := range hmsCompRecognitionTable {
		copy[k] = v
	}

	return copy
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

// GetHMSTypeFormatString for a given HMSType will return the corresponding
// fmt.Sprintf compatible format string, and the number of verbs are required
// for the format string.
func GetHMSTypeFormatString(hmsType HMSType) (string, int, error) {
	typeLower := strings.ToLower(hmsType.String())
	if value, ok := hmsCompRecognitionTable[typeLower]; ok {
		return value.GenStr, value.NumArgs, nil
	}

	return "", 0, fmt.Errorf("unknown HMSType: %s", hmsType)
}

// GetHMSTypeRegex for a given HMSType will return the regular expression
// that matches to match xnames of that HMSType.
func GetHMSTypeRegex(hmsType HMSType) (*regexp.Regexp, error) {
	typeLower := strings.ToLower(hmsType.String())
	if value, ok := hmsCompRecognitionTable[typeLower]; ok {
		return value.Regex, nil
	}

	return nil, fmt.Errorf("unknown HMSType: %s", hmsType)
}

// Allow HMSType to be treated as a standard string type.
func (t HMSType) String() string { return string(t) }

// Given a properly formatted xname, get its immediate parent.
//
//	i.e. x0c0s22b11 would become x0c0s22
func GetHMSCompParent(xname string) string {
	hmsType := GetHMSType(xname)
	if hmsType == CDU || hmsType == Cabinet {
		return "s0"
	}

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

// Remove leading zeros, i.e. for each run of numbers, trim off leading
// zeros so each run starts with either non-zero, or is a single zero.
// This has been duplicated from hms-base, but it allows the packages to be independent.
func RemoveLeadingZeros(s string) string {
	//var b strings.Builder // Go 1.10
	b := []byte("")

	// base case
	length := len(s)
	if length < 2 {
		return s
	}
	// Look for 0 after letter and before number. Skip these and
	// pretend the previous value was still a letter for the next
	// round, to catch multiple leading zeros.
	i := 0
	lastLetter := true
	for ; i < length-1; i++ {
		if s[i] == '0' && lastLetter == true {
			if unicode.IsNumber(rune(s[i+1])) {
				// leading zero
				continue
			}
		}
		if unicode.IsNumber(rune(s[i])) {
			lastLetter = false
		} else {
			lastLetter = true
		}
		// b.WriteByte(s[i]) // Go 1.10
		b = append(b, s[i])
	}
	//b.WriteByte(s[i]) // Go 1.10
	//return b.String()
	b = append(b, s[i])
	return string(b)
}
