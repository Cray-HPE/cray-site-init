// MIT License
//
// (C) Copyright [2019, 2021-2022] Hewlett Packard Enterprise Development LP
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

package sls_common

import (
	"strings"

	"github.com/Cray-HPE/hms-xname/xnametypes"
)

type SLSVersion struct {
	Counter     int    `json:"Counter"`     //Value associated with last update
	LastUpdated string `json:"LastUpdated"` //ISO 8601 timestamp
}

type SLSState struct {
	Hardware map[string]GenericHardware `json:"Hardware"`
	Networks map[string]Network         `json:"Networks"`
}

/*
CabinetType tells us what physical hardware profile is in use.  One of
River, Mountain or Hill.
*/
type CabinetType string

/*
Valid CabinetType values
*/
const (
	ClassRiver    CabinetType = "River"    // River rack
	ClassMountain CabinetType = "Mountain" // Mountain rack
	ClassHill     CabinetType = "Hill"     // Hill (combined) rack
)

var hmsHWClassMap = map[string]CabinetType{
	"river":    ClassRiver,
	"mountain": ClassMountain,
	"hill":     ClassHill,
}

/*
NetworkType represents valid values for the Type field of the network
object
*/
type NetworkType string

func (nwt NetworkType) String() string {
	return string(nwt)
}

/*
Valid values for NetworkType
*/
const (
	NetworkTypeSS10       NetworkType = "slingshot10"
	NetworkTypeCassini    NetworkType = "cassini"
	NetworkTypeEthernet   NetworkType = "ethernet"
	NetworkTypeOPA        NetworkType = "opa"
	NetworkTypeInfiniband NetworkType = "infiniband"
	NetworkTypeMixed      NetworkType = "mixed"
)

/*
HMSStringType represents the type of a piece of hardware in the "HSS
Component Naming Convention" format
*/
type HMSStringType string

func (hts HMSStringType) String() string {
	return string(hts)
}

/*

 */
const (
	CDU                      HMSStringType = "comptype_cdu"                   // dD
	CDUMgmtSwitch            HMSStringType = "comptype_cdu_mgmt_switch"       // dDwW
	CabinetCDU               HMSStringType = "comptype_cab_cdu"               // xXdD
	Cabinet                  HMSStringType = "comptype_cabinet"               // xX
	CabinetPDUController     HMSStringType = "comptype_cab_pdu_controller"    // xXmM
	CabinetPDU               HMSStringType = "comptype_cab_pdu"               // xXmMpP
	CabinetPDUNic            HMSStringType = "comptype_cab_pdu_nic"           // xXmMiI
	CabinetPDUOutlet         HMSStringType = "comptype_cab_pdu_outlet"        // xXmMpPjJ DEPRECATED
	CabinetPDUPowerConnector HMSStringType = "comptype_cab_pdu_pwr_connector" // xXmMpPvV

	Chassis            HMSStringType = "comptype_chassis"                 // xXcC
	ChassisBMC         HMSStringType = "comptype_chassis_bmc"             // xXcCbB
	CMMRectifier       HMSStringType = "comptype_cmm_rectifier"           // xXcCtT
	CMMFpga            HMSStringType = "comptype_cmm_fpga"                // xXcCfF
	CEC                HMSStringType = "comptype_cec"                     // xXeE
	ComputeModule      HMSStringType = "comptype_compmod"                 // xXcCsS
	RouterModule       HMSStringType = "comptype_rtrmod"                  // xXcCrR
	NodeBMC            HMSStringType = "comptype_ncard"                   // xXcCsSbB
	NodeBMCNic         HMSStringType = "comptype_bmc_nic"                 // xXcCsSbBiI
	NodeEnclosure      HMSStringType = "comptype_node_enclosure"          // xXcCsSeE
	NodePowerConnector HMSStringType = "comptype_compmod_power_connector" // xXcCsSvV
	Node               HMSStringType = "comptype_node"                    // xXcCsSbBnN
	Processor          HMSStringType = "comptype_node_processor"          // xXcCsSbBnNpP
	NodeNIC            HMSStringType = "comptype_node_nic"                // xXcCsSbBnNiI
	NodeHsnNIC         HMSStringType = "comptype_node_hsn_nic"            // xXcCsSbBnNhH
	Memory             HMSStringType = "comptype_dimm"                    // xXcCsSbBnNdD
	NodeAccel          HMSStringType = "comptype_node_accel"              // xXcCsSbBnNaA
	NodeFpga           HMSStringType = "comptype_node_fpga"               // xXcCsSbBfF
	HSNAsic            HMSStringType = "comptype_hsn_asic"                // xXcCrRaA
	RouterFpga         HMSStringType = "comptype_rtr_fpga"                // xXcCrRfF
	RouterTORFpga      HMSStringType = "comptype_rtr_tor_fpga"            // xXcCrRtTfF
	RouterBMC          HMSStringType = "comptype_rtr_bmc"                 // xXcCrRbB
	RouterBMCNic       HMSStringType = "comptype_rtr_bmc_nic"             // xXcCrRbBiI

	HSNBoard            HMSStringType = "comptype_hsn_board"             // xXcCrReE
	HSNLink             HMSStringType = "comptype_hsn_link"              // xXcCrRaAlL
	HSNConnector        HMSStringType = "comptype_hsn_connector"         // xXcCrRjJ
	HSNConnectorPort    HMSStringType = "comptype_hsn_connector_port"    // xXcCrRjJpP
	MgmtSwitch          HMSStringType = "comptype_mgmt_switch"           // xXcCwW
	MgmtSwitchConnector HMSStringType = "comptype_mgmt_switch_connector" // xXcCwWjJ
	MgmtHLSwitch        HMSStringType = "comptype_hl_switch"             // xXcChHsS

	// Special types and wildcards
	SMSBox         HMSStringType = "comptype_ncn_box"   // smsN
	Partition      HMSStringType = "comptype_partition" // pH.S
	System         HMSStringType = "any"                // s0
	HMSTypeAll     HMSStringType = "comptype_all"       // all
	HMSTypeAllComp HMSStringType = "comptype_all_comp"  // all_comp
	HMSTypeAllSvc  HMSStringType = "comptype_all_svc"   // all_svc
	HMSTypeInvalid HMSStringType = "INVALID"            // Not a valid type/xname
)

type hmsTypeConverter struct {
	Name          string
	HMSType       xnametypes.HMSType
	HMSStringType HMSStringType
}

var hmsTypeHMSStringTypeTable = map[string]hmsTypeConverter{
	"invalid": {
		"invalid",
		xnametypes.HMSTypeInvalid,
		HMSTypeInvalid,
	},
	"hmstypeall": {
		"hmstypeall",
		xnametypes.HMSTypeAll,
		HMSTypeAll,
	},
	"hmstypeallsvc": {
		"hmstypeallsvc",
		xnametypes.HMSTypeAllSvc,
		HMSTypeAllSvc,
	},
	"hmstypeallcomp": {
		"hmstypeallcomp",
		xnametypes.HMSTypeAllComp,
		HMSTypeAllComp,
	},
	"partition": {
		"partition",
		xnametypes.Partition,
		Partition,
	},
	"system": {
		"system",
		xnametypes.System,
		System,
	},
	"smsbox": {
		"smsbox",
		xnametypes.SMSBox,
		SMSBox,
	},
	"cdu": {
		"cdu",
		xnametypes.CDU,
		CDU,
	},
	"cdumgmtswitch": {
		"cdumgmtswitch",
		xnametypes.CDUMgmtSwitch,
		CDUMgmtSwitch,
	},
	"cabinetcdu": {
		"cabinetcdu",
		xnametypes.CabinetCDU,
		CabinetCDU,
	},
	"cabinetpducontroller": {
		"cabinetpducontroller",
		xnametypes.CabinetPDUController,
		CabinetPDUController,
	},
	"cabinetpdu": {
		"cabinetpdu",
		xnametypes.CabinetPDU,
		CabinetPDU,
	},
	"cabinetpdunic": {
		"cabinetpdunic",
		xnametypes.CabinetPDUNic,
		CabinetPDUNic,
	},
	"cabinetpduoutlet": {
		"cabinetpduoutlet",
		xnametypes.CabinetPDUOutlet,
		CabinetPDUOutlet,
	},
	"cabinetpdupowerconnector": {
		"cabinetpdupowerconnector",
		xnametypes.CabinetPDUPowerConnector,
		CabinetPDUPowerConnector,
	},
	"cec": {
		"cec",
		xnametypes.CEC,
		CEC,
	},
	"cabinet": {
		"cabinet",
		xnametypes.Cabinet,
		Cabinet,
	},
	"chassis": {
		"chassis",
		xnametypes.Chassis,
		Chassis,
	},
	"chassisbmc": {
		"chassisbmc",
		xnametypes.ChassisBMC,
		ChassisBMC,
	},
	"cmmfpga": {
		"cmmfpga",
		xnametypes.CMMFpga,
		CMMFpga,
	},
	"cmmrectifier": {
		"cmmrectifier",
		xnametypes.CMMRectifier,
		CMMRectifier,
	},
	"computemodule": {
		"computemodule",
		xnametypes.ComputeModule,
		ComputeModule,
	},
	"nodefpga": {
		"nodefpga",
		xnametypes.NodeFpga,
		NodeFpga,
	},
	"nodebmc": {
		"nodebmc",
		xnametypes.NodeBMC,
		NodeBMC,
	},
	"nodebmcnic": {
		"nodebmcnic",
		xnametypes.NodeBMCNic,
		NodeBMCNic,
	},
	"nodeenclosure": {
		"nodeenclosure",
		xnametypes.NodeEnclosure,
		NodeEnclosure,
	},
	"nodepowerconnector": {
		"nodepowerconnector",
		xnametypes.NodePowerConnector,
		NodePowerConnector,
	},
	"hsnboard": {
		"hsnboard",
		xnametypes.HSNBoard,
		HSNBoard,
	},
	"node": {
		"node",
		xnametypes.Node,
		Node,
	},
	"nodenic": {
		"nodenic",
		xnametypes.NodeNic,
		NodeNIC,
	},
	"nodehsnnic": {
		"nodehsnnic",
		xnametypes.NodeHsnNic,
		NodeHsnNIC,
	},
	"nodeaccel": {
		"nodeaccel",
		xnametypes.NodeAccel,
		NodeAccel,
	},
	"memory": {
		"memory",
		xnametypes.Memory,
		Memory,
	},
	"processor": {
		"processor",
		xnametypes.Processor,
		Processor,
	},
	"routermodule": {
		"routermodule",
		xnametypes.RouterModule,
		RouterModule,
	},
	"routerfpga": {
		"routerfpga",
		xnametypes.RouterFpga,
		RouterFpga,
	},
	"routertorfpga": {
		"routertorfpga",
		xnametypes.RouterTORFpga,
		RouterTORFpga,
	},
	"routerbmc": {
		"routerbmc",
		xnametypes.RouterBMC,
		RouterBMC,
	},
	"routerbmcnic": {
		"routerbmcnic",
		xnametypes.RouterBMCNic,
		RouterBMCNic,
	},
	"hsnasic": {
		"hsnasic",
		xnametypes.HSNAsic,
		HSNAsic,
	},
	"hsnconnector": {
		"hsnconnector",
		xnametypes.HSNConnector,
		HSNConnector,
	},
	"hsnconnectorport": {
		"hsnconnectorport",
		xnametypes.HSNConnectorPort,
		HSNConnectorPort,
	},
	"hsnlink": {
		"hsnlink",
		xnametypes.HSNLink,
		HSNLink,
	},
	"mgmtswitch": {
		"mgmtswitch",
		xnametypes.MgmtSwitch,
		MgmtSwitch,
	},
	"mgmtswitchconnector": {
		"mgmtswitchconnector",
		xnametypes.MgmtSwitchConnector,
		MgmtSwitchConnector,
	},
	"mgmthlswitch": {
		"mgmthlswitch",
		xnametypes.MgmtHLSwitch,
		MgmtHLSwitch,
	},
}

/*
HMSStringTypeToHMSType converts an HMSStringType (from this module) into an HMSType (from hmstypes.go)
*/
func HMSStringTypeToHMSType(str HMSStringType) xnametypes.HMSType {
	for _, tabEntry := range hmsTypeHMSStringTypeTable {
		if str == tabEntry.HMSStringType {
			return tabEntry.HMSType
		}
	}

	return hmsTypeHMSStringTypeTable["invalid"].HMSType
}

/*
HMSTypeToHMSStringType converts an HMSType (from hmstypes.go) into an HMSStringType (from this module)
*/
func HMSTypeToHMSStringType(str xnametypes.HMSType) HMSStringType {
	for _, tabEntry := range hmsTypeHMSStringTypeTable {
		if str == tabEntry.HMSType {
			return tabEntry.HMSStringType
		}
	}

	return hmsTypeHMSStringTypeTable["invalid"].HMSStringType
}

/*
Verify that a cabinet type/class is valid.
*/
func IsCabinetTypeValid(class CabinetType) bool {
	_, ok := hmsHWClassMap[strings.ToLower(string(class))]
	return ok
}
