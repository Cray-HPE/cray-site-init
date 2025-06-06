/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
*/

package sls

import (
	"fmt"
	"os"
	"testing"

	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	shcdParser "github.com/Cray-HPE/hms-shcd-parser/pkg/shcd-parser"

	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

/*
 * NOTE: You have to be really careful adding or modifying this test structure below. This config generator has to make
 * a lot of assumptions, so there are a lot of implicit ordering constraints that need to be honored. Also, if you
 * look at this data and then look at the tests you'll probably think well that just doesn't make any sense for a
 * couple of them. Case in point, anything at U20...that will have a slot number of 19. That's just the way the naming
 * convention works and actually the reason for the test.
 */

var HMNConnections = []shcdParser.HMNRow{
	{
		Source:              "mn01",
		SourceRack:          "x3000",
		SourceLocation:      "u01",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p25",
	},
	{
		Source:         "wn01",
		SourceRack:     "x3000",
		SourceLocation: "u07",
	},
	{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p28",
	},
	{
		Source:              "sn01",
		SourceRack:          "x3000",
		SourceLocation:      "u13",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p30",
	},
	{
		Source:              "nid000001",
		SourceRack:          "x3000",
		SourceLocation:      "u19",
		SourceSubLocation:   "R",
		SourceParent:        "SubRack-001-cmc",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p33",
	},
	{
		Source:              "nid000002",
		SourceRack:          "x3000",
		SourceLocation:      "U19",
		SourceSubLocation:   "L",
		SourceParent:        "SubRack-001-cmc",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p34",
	},
	{
		Source:              "cn-03",
		SourceRack:          "x3000",
		SourceLocation:      "u20",
		SourceSubLocation:   "R",
		SourceParent:        "SubRack-001-cmc",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p35",
	},
	{
		Source:              "cn04",
		SourceRack:          "x3000",
		SourceLocation:      "u20",
		SourceSubLocation:   "L",
		SourceParent:        "SubRack-001-cmc",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p36",
	},
	{
		Source:              "nid000005",
		SourceRack:          "x3000",
		SourceLocation:      "u21",
		SourceSubLocation:   "R",
		SourceParent:        "SubRack-002-cmc",
		DestinationRack:     "x3001",
		DestinationLocation: "u21",
		DestinationPort:     "p21",
	},
	{
		Source:              "SubRack-001-cmc",
		SourceRack:          "x3000",
		SourceLocation:      "u19",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p38",
	},
	{
		Source:              "SubRack-002-cmc",
		SourceRack:          "x3000",
		SourceLocation:      "u21",
		DestinationRack:     "x3001",
		DestinationLocation: "u21",
		DestinationPort:     "p22",
	},
	{
		// Application Node - UAN: x3000c0s26b0n0
		Source:              "UAN",
		SourceRack:          "x3000",
		SourceLocation:      "u26",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p37",
	},
	{
		// Application Node - Login Node: x3000c0s27b0n0
		Source:              "Ln01",
		SourceRack:          "x3000",
		SourceLocation:      "u27",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p38",
	},
	{
		// Application Node - Gatway Node: x3000c0s28b0n0
		Source:              "Gn01",
		SourceRack:          "x3000",
		SourceLocation:      "u28",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p38",
	},
	{
		// Application Node - Visualization Node: x3000c0s29b0n0
		Source:              "vn01",
		SourceRack:          "x3000",
		SourceLocation:      "u29",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p38",
	},
	{
		// Application Node - LNet Node: x3000c0s30b0n0
		Source:              "Lnet01",
		SourceRack:          "x3000",
		SourceLocation:      "u30",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p38",
	},
	{
		// Application Node - LNet Node: x3000c0s31b0n0
		Source:              "Lnet02",
		SourceRack:          "x3000",
		SourceLocation:      "u31",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p38",
	},
	{
		// Application Node - UAN Node: x3000c0s32b0n0
		Source:              "uan02",
		SourceRack:          "x3000",
		SourceLocation:      "u32",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p38",
	},
	{
		Source:              "sw-hsn001",
		SourceRack:          "x3000",
		SourceLocation:      "u22",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p47",
	},
	{
		Source:              "Columbia",
		SourceRack:          "x3000",
		SourceLocation:      "u24",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p48",
	},
	{
		Source:              "x3000p0",
		SourceRack:          "x3000",
		SourceLocation:      " ",
		DestinationRack:     "x3000",
		DestinationLocation: "u38",
		DestinationPort:     "j41",
	},
	{
		Source:              "x3000door-Motiv",
		SourceRack:          "x3000",
		SourceLocation:      " ",
		DestinationRack:     "x3000",
		DestinationLocation: "u36",
		DestinationPort:     "j27",
	},
	{
		Source:              "CAN",
		SourceRack:          "cfcan",
		SourceLocation:      " ",
		DestinationRack:     "x3000",
		DestinationLocation: "u38",
		DestinationPort:     "j49",
	},
	{
		Source:              "x3000p0",
		SourceRack:          "x3000",
		SourceLocation:      "p0",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "j48",
	},
	{
		Source:              "x3001p1",
		SourceRack:          "x3001",
		SourceLocation:      "p0",
		DestinationRack:     "x3001",
		DestinationLocation: "u42",
		DestinationPort:     "j48",
	},
	{
		Source:              "pdu0",
		SourceRack:          "x3001",
		SourceLocation:      "pdu0",
		DestinationRack:     "x3001",
		DestinationLocation: "u42",
		DestinationPort:     "j27",
	},
	{
		Source:              "pdu2",
		SourceRack:          "x3001",
		SourceLocation:      "pdu0",
		DestinationRack:     "x3001",
		DestinationLocation: "u42",
		DestinationPort:     "j27",
	},

	//
	// River hardware in a EX 2500 Lite cabinet
	//
	// 4 compute nodes BMCs and a CMC connecting to a management switch (u16) within the same cabinet
	{
		Source:              "nid000101",
		SourceRack:          "x5004",
		SourceLocation:      "u17",
		SourceSubLocation:   "R",
		SourceParent:        "SubRack-004-CMC",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j31",
	},
	{
		Source:              "nid000102",
		SourceRack:          "x5004",
		SourceLocation:      "u18",
		SourceSubLocation:   "R",
		SourceParent:        "SubRack-004-CMC",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j32",
	},
	{
		Source:              "nid000103",
		SourceRack:          "x5004",
		SourceLocation:      "u18",
		SourceSubLocation:   "L",
		SourceParent:        "SubRack-004-CMC",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j33",
	},
	{
		Source:              "nid000104",
		SourceRack:          "x5004",
		SourceLocation:      "u17",
		SourceSubLocation:   "L",
		SourceParent:        "SubRack-004-CMC",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j34",
	},
	{
		Source:              "SubRack-004-CMC",
		SourceRack:          "x5004",
		SourceLocation:      "u17",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j30",
	},

	// Management NCN - Worker connected to a management switch (u16) within the same cabinet
	{
		Source:              "wn50",
		SourceRack:          "x5004",
		SourceLocation:      "u19",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j35",
	},

	// Application Node - UAN - connecting to a management switch (u16) within the same cabinet
	{
		Source:              "uan50",
		SourceRack:          "x5004",
		SourceLocation:      "u20",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j36",
	},

	// Application Node - LNETRouter - connecting to a management switch (u16) in a different cabinet
	{
		Source:              "lnet50",
		SourceRack:          "x5004",
		SourceLocation:      "u21",
		DestinationRack:     "x3001",
		DestinationLocation: "u42",
		DestinationPort:     "j20",
	},

	// HSN switch
	{
		Source:              "sw-hsn50",
		SourceRack:          "x5004",
		SourceLocation:      "u1",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j37",
	},

	// Cabinet PDUs
	{
		Source:              "x5004p0",
		SourceRack:          "x5004",
		SourceLocation:      "p0",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j38",
	},
	{
		Source:              "pdu1",
		SourceRack:          "x5004",
		SourceLocation:      "p1",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j39",
	},
}

// TestSLSInputState contains a data to generate a configuration with the following configuration
// - 2 River Cabinets (x3000 & x3001)
// - 4 LeafBMC switches (2 in each river cabinet)
//   - sw-leaf-bmc-00[1-3] are Dell
//   - sw-leaf-bmc-004 is Aruba
//
// - 1 Hill Cabinet x5000
// - 1 Mountain Cabinet x1000
var TestSLSInputState = GeneratorInputState{
	ApplicationNodeConfig: GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"Lnet",
			"Login",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":    "Visualization",
			"Login": "UAN",
			"Lnet":  "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0": {"uan-01"},
			"x3000c0s28b0n0": {"gateway-01"},
			"x3000c0s29b0n0": {"visualization-01"},
			"x3000c0s30b0n0": {"lnet-01"},
			"x3000c0s31b0n0": {"lnet-02"},
			"x3000c0s32b0n0": {"uan-02"},
			"x5004c4s20b0n0": {"uan-50"},
			"x5004c4s21b0n0": {"lnet-50"},
		},
	},

	ManagementSwitches: map[string]slsCommon.GenericHardware{
		"x3000c0w22": buildMgmtSwitch(
			"x3000c0w22",
			"sw-leaf-bmc-01",
			"10.254.0.2",
			networking.ManagementSwitchBrandDell,
		),
		"x3000c0w38": buildMgmtSwitch(
			"x3000c0w38",
			"sw-leaf-bmc-02",
			"10.254.0.3",
			networking.ManagementSwitchBrandDell,
		),
		"x3001c0w21": buildMgmtSwitch(
			"x3001c0w21",
			"sw-leaf-bmc-03",
			"10.254.0.4",
			networking.ManagementSwitchBrandDell,
		),
		"x3001c0w42": buildMgmtSwitch(
			"x3001c0w42",
			"sw-leaf-bmc-04",
			"10.254.0.42",
			networking.ManagementSwitchBrandAruba,
		),
		"x5004c4w16": buildMgmtSwitch(
			"x5004c4w16",
			"sw-leaf-bmc-05",
			"10.254.0.43",
			networking.ManagementSwitchBrandAruba,
		),
	},

	RiverCabinets: map[string]CabinetTemplate{
		"x3000": {
			Xname: xnames.Cabinet{
				Cabinet: 3000,
			},
			Class:                   slsCommon.ClassRiver, // Model: , Not applicable???
			AirCooledChassisList:    DefaultRiverChassisList,
			LiquidCooledChassisList: []int{},
			CabinetNetworks: map[string]map[string]slsCommon.CabinetNetworks{
				"cn": {
					"HMN": {
						CIDR:    "10.107.0.0/22",
						Gateway: "10.107.0.1",
						VLan:    1513,
					},
					"NMN": {
						CIDR:    "10.106.0.0/22",
						Gateway: "10.106.0.1",
						VLan:    1770,
					},
				},
				"ncn": {
					"HMN": {
						CIDR:    "10.107.0.0/22",
						Gateway: "10.107.0.1",
						VLan:    1513,
					},
					"NMN": {
						CIDR:    "10.106.0.0/22",
						Gateway: "10.106.0.1",
						VLan:    1770,
					},
				},
			},
		},
		"x3001": {
			Xname: xnames.Cabinet{
				Cabinet: 3001,
			},
			Class:                   slsCommon.ClassRiver, // Model: , Not applicable???
			AirCooledChassisList:    DefaultRiverChassisList,
			LiquidCooledChassisList: []int{},
			CabinetNetworks: map[string]map[string]slsCommon.CabinetNetworks{
				"cn": {
					"HMN": {
						CIDR:    "10.107.2.0/22",
						Gateway: "10.107.2.1",
						VLan:    1514,
					},
					"NMN": {
						CIDR:    "10.106.2.0/22",
						Gateway: "10.106.2.1",
						VLan:    1771,
					},
				},
				"ncn": {
					"HMN": {
						CIDR:    "10.107.2.0/22",
						Gateway: "10.107.2.1",
						VLan:    1514,
					},
					"NMN": {
						CIDR:    "10.106.2.0/22",
						Gateway: "10.106.2.1",
						VLan:    1771,
					},
				},
			},
		},
	},

	HillCabinets: map[string]CabinetTemplate{
		// EX2000 - Traditional Hill cabinet
		"x5000": {
			Xname: xnames.Cabinet{
				Cabinet: 5000,
			},
			Class:                   slsCommon.ClassHill, // Model: , // TODO
			AirCooledChassisList:    []int{},
			LiquidCooledChassisList: DefaultHillChassisList,
			CabinetNetworks: map[string]map[string]slsCommon.CabinetNetworks{
				"cn": {
					"HMN": {
						CIDR:    "10.108.4.0/22",
						Gateway: "10.108.4.1",
						VLan:    2000,
					},
					"NMN": {
						CIDR:    "10.107.4.0/22",
						Gateway: "10.107.4.1",
						VLan:    3000,
					},
				},
			},
		},

		// EX2500 - 1 Liquid Cooled Chassis
		"x5001": {
			Xname: xnames.Cabinet{
				Cabinet: 5001,
			},
			Class:                   slsCommon.ClassHill,
			Model:                   "EX2500",
			AirCooledChassisList:    []int{},
			LiquidCooledChassisList: []int{0},
		}, // EX2500 - 2 Liquid Cooled Chassis
		"x5002": {
			Xname: xnames.Cabinet{
				Cabinet: 5002,
			},
			Class:                slsCommon.ClassHill,
			Model:                "EX2500",
			AirCooledChassisList: []int{},
			LiquidCooledChassisList: []int{
				0,
				1,
			},
		}, // EX2500 - 3 Liquid Cooled Chassis
		"x5003": {
			Xname: xnames.Cabinet{
				Cabinet: 5003,
			},
			Class:                slsCommon.ClassHill,
			Model:                "EX2500",
			AirCooledChassisList: []int{},
			LiquidCooledChassisList: []int{
				0,
				1,
				2,
			},
		}, // EX2500 - 1 Liquid Cooled Chassis and 1 air cooled chassis
		"x5004": {
			Xname: xnames.Cabinet{
				Cabinet: 5004,
			},
			Class:                   slsCommon.ClassHill,
			Model:                   "EX2500",
			AirCooledChassisList:    []int{4},
			LiquidCooledChassisList: []int{0},
		},
	},

	MountainCabinets: map[string]CabinetTemplate{
		"x1000": {
			Xname: xnames.Cabinet{
				Cabinet: 1000,
			},
			Class:                   slsCommon.ClassMountain, // Model: , // TODO
			AirCooledChassisList:    []int{},
			LiquidCooledChassisList: DefaultMountainChassisList,
			CabinetNetworks: map[string]map[string]slsCommon.CabinetNetworks{
				"cn": {
					"HMN": {
						CIDR:    "10.108.6.0/22",
						Gateway: "10.108.6.1",
						VLan:    2001,
					},
					"NMN": {
						CIDR:    "10.107.6.0/22",
						Gateway: "10.107.6.1",
						VLan:    3001,
					},
				},
			},
		},
	},

	MountainStartingNid: 1000,
}

func buildMgmtSwitch(
	xname, name, ipAddress string, brand networking.ManagementSwitchBrand,
) slsCommon.GenericHardware {
	return slsCommon.GenericHardware{
		Parent:     xnametypes.GetHMSCompParent(xname),
		Xname:      xname,
		Type:       slsCommon.MgmtSwitch,
		Class:      slsCommon.ClassRiver,
		TypeString: xnametypes.MgmtSwitch,
		ExtraPropertiesRaw: slsCommon.ComptypeMgmtSwitch{
			IP4Addr:          ipAddress,
			Brand:            brand.String(),
			Model:            "S3048T-ON",
			SNMPAuthPassword: "vault://hms-creds/" + xname,
			SNMPAuthProtocol: "MD5",
			SNMPPrivPassword: "vault://hms-creds/" + xname,
			SNMPPrivProtocol: "DES",
			SNMPUsername:     "testuser",
			Aliases:          []string{name},
		},
	}
}

func stringArrayContains(
	array []string, needle string,
) bool {
	for _, item := range array {
		if item == needle {
			return true
		}
	}

	return false
}

type ConfigGeneratorTestSuite struct {
	suite.Suite

	generator   StateGenerator
	allHardware map[string]slsCommon.GenericHardware
}

func (suite *ConfigGeneratorTestSuite) SetupSuite() {
	// Setup logger for testing
	encoderCfg := zap.NewProductionEncoderConfig()
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
	)

	// Normalize and validate the application node config
	err := TestSLSInputState.ApplicationNodeConfig.Normalize()
	suite.NoError(err)

	err = TestSLSInputState.ApplicationNodeConfig.Validate()
	suite.NoError(err)

	g := NewStateGenerator(
		logger,
		TestSLSInputState,
		HMNConnections,
	)

	suite.allHardware = g.buildHardwareSection()
	suite.generator = g
}

func (suite *ConfigGeneratorTestSuite) TestVerifyNoEmptyHardware() {
	for xname, hardware := range suite.allHardware {
		suite.NotEmpty(
			xname,
			"xname key is empty",
		)
		suite.NotEmpty(
			hardware.Xname,
			"Xname is empty for %s",
			xname,
		)
		suite.NotEmpty(
			hardware.Parent,
			"Parent is empty for %s",
			xname,
		)
		suite.NotEmpty(
			hardware.Type,
			"Type is empty for %s",
			xname,
		)
		suite.NotEmpty(
			hardware.TypeString,
			"TypeString is empty for %s",
			xname,
		)
		suite.NotEmpty(
			hardware.Class,
			"Class is empty for %s",
			xname,
		)

		// Note: The extra properties field maybe empty for some component types
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyXnameMatches() {
	for xname, hardware := range suite.allHardware {
		suite.Equal(
			xname,
			hardware.Xname,
			"Found inconsistent xname in the hardware map, key does not match SLS object",
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyXnameValid() {
	for _, hardware := range suite.allHardware {
		suite.True(
			xnametypes.IsHMSCompIDValid(hardware.Xname),
			"Found invalid xname %s",
			hardware.Xname,
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyParentXname() {
	for _, hardware := range suite.allHardware {
		expectedParentXname := xnametypes.GetHMSCompParent(hardware.Xname)

		suite.Equal(
			expectedParentXname,
			hardware.Parent,
			"Found inconsistent parent xname for %s",
			hardware.Xname,
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyTypeValid() {
	for _, hardware := range suite.allHardware {
		expectedType := slsCommon.HMSTypeToHMSStringType(xnametypes.GetHMSType(hardware.Xname))

		suite.Equal(
			expectedType,
			hardware.Type,
			"Found inconsistent type for %s",
			hardware.Xname,
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyTypeStringValid() {
	for _, hardware := range suite.allHardware {
		expectedTypeString := xnametypes.GetHMSType(hardware.Xname)

		suite.Equal(
			expectedTypeString,
			hardware.TypeString,
			"Found inconsistent type string for %s",
			hardware.Xname,
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestMasterNode() {
	/*
	  "x3000c0s1b0n0": {
	    "Parent": "x3000c0s1b0",
	    "Xname": "x3000c0s1b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 100001,
	      "Role": "Management",
	      "SubRole": "Master",
	      "Aliases": [
	        "ncn-m001"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s1b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s1b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s1b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.Role,
		"Management",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"Master",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"ncn-m001",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestCANWorkerNode() {
	/*
	  "x3000c0s7b0n0": {
	    "Parent": "x3000c0s7b0",
	    "Xname": "x3000c0s7b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 100002,
	      "Role": "Management",
	      "SubRole": "Worker",
	      "Aliases": [
	        "ncn-w001"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s7b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s7b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s7b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.Role,
		"Management",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"Worker",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"ncn-w001",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestWorkerNode() {
	/*
	  "x3000c0s9b0n0": {
	    "Parent": "x3000c0s9b0",
	    "Xname": "x3000c0s9b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 100003,
	      "Role": "Management",
	      "SubRole": "Worker",
	      "Aliases": [
	        "ncn-w002"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s9b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s9b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s9b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.Role,
		"Management",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"Worker",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"ncn-w002",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestStorageNode() {
	/*
	  "x3000c0s13b0n0": {
	    "Parent": "x3000c0s13b0",
	    "Xname": "x3000c0s13b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 100004,
	      "Role": "Management",
	      "SubRole": "Storage",
	      "Aliases": [
	        "ncn-s001"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s13b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s13b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s13b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.Role,
		"Management",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"Storage",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"ncn-s001",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestCompute_NID() {
	/*
	  "x3000c0s19b1n0": {
	    "Parent": "x3000c0s19b1",
	    "Xname": "x3000c0s19b1n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 1,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000001"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b1n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s19b1",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s19b1n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NID,
		1,
	)
	suite.Equal(
		hardwareExtraProperties.Role,
		"Compute",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"nid000001",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestCompute_NID_CapitolSourceU() {
	/*
	  "x3000c0s19b2n0": {
	    "Parent": "x3000c0s19b2",
	    "Xname": "x3000c0s19b2n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 2,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000002"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b2n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s19b2",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s19b2n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NID,
		2,
	)
	suite.Equal(
		hardwareExtraProperties.Role,
		"Compute",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"nid000002",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestCompute_CN_WithHyphen() {
	/*
	  "x3000c0s19b3n0": {
	    "Parent": "x3000c0s19b3",
	    "Xname": "x3000c0s19b3n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 3,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000003"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b3n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s19b3",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s19b3n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NID,
		3,
	)
	suite.Equal(
		hardwareExtraProperties.Role,
		"Compute",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"nid000003",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestCompute_CN_WithoutHyphen() {
	/*
	  "x3000c0s19b4n0": {
	    "Parent": "x3000c0s19b4",
	    "Xname": "x3000c0s19b4n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 4,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000004"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b4n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s19b4",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s19b4n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NID,
		4,
	)
	suite.Equal(
		hardwareExtraProperties.Role,
		"Compute",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"nid000004",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestCompute_SwitchDifferentCabinet() {
	/*
		  {
		    "Parent": "x3000c0s21b1",
			"Xname": "x3000c0s21b1n0",
			"Type": "comptype_node",
			"Class": "River",
			"TypeString": "Node",
			"ExtraProperties": {
			  "NID": 5,
			  "Role": "Compute",
			  "Aliases": [
			    "nid000005"
			  ]
		    }
		  }
	*/
	hardware, ok := suite.allHardware["x3000c0s21b1n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s21b1",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s21b1n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NID,
		5,
	)
	suite.Equal(
		hardwareExtraProperties.Role,
		"Compute",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"nid000005",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestCompute_CMC() {
	/*
	  "x3000c0s19b999": {
	    "Parent": "x3000c0s19",
	    "Xname": "x3000c0s19b999",
	    "Type": "comptype_ncard",
	    "Class": "River",
	    "TypeString": "NodeBMC"
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b999"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s19",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s19b999",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_ncard"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("NodeBMC"),
	)
}

func (suite *ConfigGeneratorTestSuite) TestUAN() {
	/*
		  "x3000c0s26b0n0": {
		    "Parent": "x3000c0s26b0",
		    "Xname": "x3000c0s26b0n0",
		    "Type": "comptype_node",
		    "Class": "River",
		    "TypeString": "Node",
		    "ExtraProperties": {
		      "Role": "Application",
			  "SubRole": "UAN"
			  "Aliases": [
				  "uan-01"
			  ]
		    }
		  },
	*/
	hardware, ok := suite.allHardware["x3000c0s26b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s26b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s26b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	// TODO: CASMHMS-3598
	// suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(
		hardwareExtraProperties.Role,
		"Application",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"UAN",
	)
	suite.Equal(
		hardwareExtraProperties.Aliases,
		[]string{"uan-01"},
	)
}

func (suite *ConfigGeneratorTestSuite) TestLoginNode() {
	/*
		  "x3000c0s27b0n0": {
		    "Parent": "x3000c0s27b0",
		    "Xname": "x3000c0s27b0n0",
		    "Type": "comptype_node",
		    "Class": "River",
		    "TypeString": "Node",
		    "ExtraProperties": {
		      "Role": "Application",
			  "SubRole": "UAN"
		    }
		  },
	*/
	hardware, ok := suite.allHardware["x3000c0s27b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s27b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s27b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	// TODO: CASMHMS-3598
	// suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(
		hardwareExtraProperties.Role,
		"Application",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"UAN",
	)
	suite.Empty(hardwareExtraProperties.Aliases) // This login node was intentionally not given a Alias
}

func (suite *ConfigGeneratorTestSuite) TestGatewayNode() {
	/*
		  "x3000c0s28b0n0": {
		    "Parent": "x3000c0s28b0",
		    "Xname": "x3000c0s28b0n0",
		    "Type": "comptype_node",
		    "Class": "River",
		    "TypeString": "Node",
		    "ExtraProperties": {
		      "Role": "Application",
			  "SubRole": "Gateway"
		    }
		  },
	*/
	hardware, ok := suite.allHardware["x3000c0s28b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s28b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s28b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	// TODO: CASMHMS-3598
	// suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(
		hardwareExtraProperties.Role,
		"Application",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"Gateway",
	)
	suite.Equal(
		hardwareExtraProperties.Aliases,
		[]string{"gateway-01"},
	)
}

func (suite *ConfigGeneratorTestSuite) TestVisualizationNode() {
	/*
		  "x3000c0s29b0n0": {
		    "Parent": "x3000c0s29b0",
		    "Xname": "x3000c0s29b0n0",
		    "Type": "comptype_node",
		    "Class": "River",
		    "TypeString": "Node",
		    "ExtraProperties": {
		      "Role": "Application",
			  "SubRole": "Visualization"
		    }
		  },
	*/
	hardware, ok := suite.allHardware["x3000c0s29b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s29b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s29b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	// TODO: CASMHMS-3598
	// suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(
		hardwareExtraProperties.Role,
		"Application",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"Visualization",
	)
	suite.Equal(
		hardwareExtraProperties.Aliases,
		[]string{"visualization-01"},
	)
}

func (suite *ConfigGeneratorTestSuite) TestNodeLNet01() {
	/*
		  "x3000c0s30b0n0": {
		    "Parent": "x3000c0s30b0",
		    "Xname": "x3000c0s30b0n0",
		    "Type": "comptype_node",
		    "Class": "River",
		    "TypeString": "Node",
		    "ExtraProperties": {
		      "Role": "Application",
			  "SubRole": "LNETRouter"
		    }
		  },
	*/
	hardware, ok := suite.allHardware["x3000c0s30b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s30b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s30b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	// TODO: CASMHMS-3598
	// suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(
		hardwareExtraProperties.Role,
		"Application",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"LNETRouter",
	)
	suite.Equal(
		hardwareExtraProperties.Aliases,
		[]string{"lnet-01"},
	)
}

func (suite *ConfigGeneratorTestSuite) TestNodeLNet02() {
	/*
		  "x3000c0s31b0n0": {
		    "Parent": "x3000c0s31b0",
		    "Xname": "x3000c0s31b0n0",
		    "Type": "comptype_node",
		    "Class": "River",
		    "TypeString": "Node",
		    "ExtraProperties": {
		      "Role": "Application",
			  "SubRole": "LNETRouter"
		    }
		  },
	*/
	hardware, ok := suite.allHardware["x3000c0s31b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s31b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s31b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	// TODO: CASMHMS-3598
	// suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(
		hardwareExtraProperties.Role,
		"Application",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"LNETRouter",
	)
	suite.Equal(
		hardwareExtraProperties.Aliases,
		[]string{"lnet-02"},
	)
}

func (suite *ConfigGeneratorTestSuite) TestNodeUAN2() {
	/*
		  "x3000c0s32b0n0": {
		    "Parent": "x3000c0s32b0",
		    "Xname": "x3000c0s32b0n0",
		    "Type": "comptype_node",
		    "Class": "River",
		    "TypeString": "Node",
		    "ExtraProperties": {
		      "Role": "Application",
			  "SubRole": "UAN"
		    }
		  },
	*/
	hardware, ok := suite.allHardware["x3000c0s32b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0s32b0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0s32b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	// TODO: CASMHMS-3598
	// suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(
		hardwareExtraProperties.Role,
		"Application",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"UAN",
	)
	suite.Equal(
		hardwareExtraProperties.Aliases,
		[]string{"uan-02"},
	)
}

func (suite *ConfigGeneratorTestSuite) TestManagementSwitch() {
	/*
	  "x3000c0w22": {
	    "Parent": "x3000",
	    "Xname": "x3000c0w22",
	    "Type": "comptype_mgmt_switch",
	    "Class": "River",
	    "TypeString": "MgmtSwitch",
	    "ExtraProperties": {
	      "IP4addr": "10.254.0.2",
	      "Model": "S3048T-ON",
	      "SNMPAuthPassword": "vault://hms-creds/x3000c0w22",
	      "SNMPAuthProtocol": "MD5",
	      "SNMPPrivPassword": "vault://hms-creds/x3000c0w22",
	      "SNMPPrivProtocol": "DES",
	      "SNMPUsername": "testuser"
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0w22"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0w22",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitch"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitch)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.IP4Addr,
		"10.254.0.2",
	)
	suite.Equal(
		hardwareExtraProperties.Model,
		"S3048T-ON",
	)
	suite.Equal(
		hardwareExtraProperties.SNMPAuthPassword,
		"vault://hms-creds/x3000c0w22",
	)
	suite.Equal(
		hardwareExtraProperties.SNMPAuthProtocol,
		"MD5",
	)
	suite.Equal(
		hardwareExtraProperties.SNMPPrivPassword,
		"vault://hms-creds/x3000c0w22",
	)
	suite.Equal(
		hardwareExtraProperties.SNMPPrivProtocol,
		"DES",
	)
	suite.Equal(
		hardwareExtraProperties.SNMPUsername,
		"testuser",
	)
}

func (suite *ConfigGeneratorTestSuite) TestHSNSwitch_HSN() {
	/*
	  "x3000c0r22b0": {
	    "Parent": "x3000c0r22",
	    "Xname": "x3000c0r22b0",
	    "Type": "comptype_rtr_bmc",
	    "Class": "River",
	    "TypeString": "RouterBMC",
	    "ExtraProperties": {
	      "Username": "vault://hms-creds/x3000c0r22b0",
	      "Password": "vault://hms-creds/x3000c0r22b0"
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0r22b0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0r22",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0r22b0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_rtr_bmc"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("RouterBMC"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeRtrBmc)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.Username,
		"vault://hms-creds/x3000c0r22b0",
	)
	suite.Equal(
		hardwareExtraProperties.Password,
		"vault://hms-creds/x3000c0r22b0",
	)
}

func (suite *ConfigGeneratorTestSuite) TestHSNSwitch_Columbia() {
	/*
	  "x3000c0r24b0": {
	    "Parent": "x3000c0r24",
	    "Xname": "x3000c0r24b0",
	    "Type": "comptype_rtr_bmc",
	    "Class": "River",
	    "TypeString": "RouterBMC",
	    "ExtraProperties": {
	      "Username": "vault://hms-creds/x3000c0r24b0",
	      "Password": "vault://hms-creds/x3000c0r24b0"
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0r24b0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0r24",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0r24b0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_rtr_bmc"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("RouterBMC"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeRtrBmc)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.Username,
		"vault://hms-creds/x3000c0r24b0",
	)
	suite.Equal(
		hardwareExtraProperties.Password,
		"vault://hms-creds/x3000c0r24b0",
	)
}

func (suite *ConfigGeneratorTestSuite) TestCabinetPDUController() {
	/*
	   "x3000m0": {
	     "Parent": "x3000",
	     "Xname": "x3000m0",
	     "Type": "comptype_cab_pdu_controller",
	     "Class": "River",
	     "TypeString": "CabinetPDUController",
	   }
	*/

	hardware, ok := suite.allHardware["x3000m0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000",
	)
	suite.Equal(
		hardware.Xname,
		"x3000m0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_cab_pdu_controller"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("CabinetPDUController"),
	)

	suite.Nil(
		hardware.ExtraPropertiesRaw,
		"ExtraProperties type is not nil",
	)
}

func (suite *ConfigGeneratorTestSuite) TestCabinetPDUController_pduPrefix0() {
	/*
	   "x3001m0": {
	     "Parent": "x3001",
	     "Xname": "x3000m0",
	     "Type": "comptype_cab_pdu_controller",
	     "Class": "River",
	     "TypeString": "CabinetPDUController",
	   }
	*/

	hardware, ok := suite.allHardware["x3001m0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3001",
	)
	suite.Equal(
		hardware.Xname,
		"x3001m0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_cab_pdu_controller"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("CabinetPDUController"),
	)

	suite.Nil(
		hardware.ExtraPropertiesRaw,
		"ExtraProperties type is not nil",
	)
}

func (suite *ConfigGeneratorTestSuite) TestCabinetPDUController_pduPrefix2() {
	/*
	   "x3001m2": {
	     "Parent": "x3001",
	     "Xname": "x3000m2",
	     "Type": "comptype_cab_pdu_controller",
	     "Class": "River",
	     "TypeString": "CabinetPDUController",
	   }
	*/

	hardware, ok := suite.allHardware["x3001m2"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3001",
	)
	suite.Equal(
		hardware.Xname,
		"x3001m2",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_cab_pdu_controller"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("CabinetPDUController"),
	)

	suite.Nil(
		hardware.ExtraPropertiesRaw,
		"ExtraProperties type is not nil",
	)
}

func (suite *ConfigGeneratorTestSuite) TestMgmtSwitchConnector_CabinetPDUController() {
	/*
		"x3000c0w22j48": {
		"Parent": "x3000c0w22",
		"Xname": "x3000c0w22j48",
		"Type": "comptype_mgmt_switch_connector",
		"Class": "River",
		"TypeString": "MgmtSwitchConnector",
		"ExtraProperties": {
			"NodeNics": [
				"x3000m0"
			],
			"VendorName": "ethernet1/1/48"
		}
	*/
	hardware, ok := suite.allHardware["x3000c0w22j48"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3000c0w22",
	)
	suite.Equal(
		hardware.Xname,
		"x3000c0w22j48",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x3000m0"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"ethernet1/1/48",
	)
}

func (suite *ConfigGeneratorTestSuite) TestMgmtSwitchConnector_ComputeSwitchDifferentCabinet() {
	/*
		  {
		    "Parent": "x3001c0w21",
		    "Xname": "x3001c0w21j21",
		    "Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
			  "NodeNics": [
			    "x3000c0s21b1"
			  ],
			  "VendorName": "ethernet1/1/21"
		 	  }
			}
		  }
	*/
	hardware, ok := suite.allHardware["x3001c0w21j21"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3001c0w21",
	)
	suite.Equal(
		hardware.Xname,
		"x3001c0w21j21",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x3000c0s21b1"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"ethernet1/1/21",
	)
}

func (suite *ConfigGeneratorTestSuite) TestMgmtSwitchConnector_ArubaSwitch() {
	/*
		  {
		    "Parent": "x3001c0w42",
		    "Xname": "x3001c0w42j48",
		    "Type": "comptype_mgmt_switch_connector",
		    "Class": "River",
		    "TypeString": "MgmtSwitchConnector",
		    "ExtraProperties": {
		      "NodeNics": [
			    "x3001m1"
			  ],
			  "VendorName": "1/1/48"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x3001c0w42j48"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3001c0w42",
	)
	suite.Equal(
		hardware.Xname,
		"x3001c0w42j48",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x3001m1"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/48",
	)
}

func (suite *ConfigGeneratorTestSuite) TestCabinet_River() {
	/*
		{
			"Parent": "s0",
			"Xname": "x3001",
			"Type": "comptype_cabinet",
			"Class": "River",
			"TypeString": "Cabinet",
			"ExtraProperties": {}
		}
	*/

	hardware, ok := suite.allHardware["x3001"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"s0",
	)
	suite.Equal(
		hardware.Xname,
		"x3001",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_cabinet"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Cabinet"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeCabinet)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		slsCommon.ComptypeCabinet{
			Networks: map[string]map[string]slsCommon.CabinetNetworks{
				"cn": {
					"HMN": {
						CIDR:    "10.107.2.0/22",
						Gateway: "10.107.2.1",
						VLan:    1514,
					},
					"NMN": {
						CIDR:    "10.106.2.0/22",
						Gateway: "10.106.2.1",
						VLan:    1771,
					},
				},
				"ncn": {
					"HMN": {
						CIDR:    "10.107.2.0/22",
						Gateway: "10.107.2.1",
						VLan:    1514,
					},
					"NMN": {
						CIDR:    "10.106.2.0/22",
						Gateway: "10.106.2.1",
						VLan:    1771,
					},
				},
			},
		},
		hardwareExtraProperties,
	)
}

func (suite *ConfigGeneratorTestSuite) TestCabinet_Hill() {
	/*
		{
			"Parent": "s0",
			"Xname": "x5000",
			"Type": "comptype_cabinet",
			"Class": "Hill",
			"TypeString": "Cabinet",
			"ExtraProperties": {}
		}
	*/

	hardware, ok := suite.allHardware["x5000"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"s0",
	)
	suite.Equal(
		hardware.Xname,
		"x5000",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_cabinet"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("Hill"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Cabinet"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeCabinet)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		slsCommon.ComptypeCabinet{
			Networks: map[string]map[string]slsCommon.CabinetNetworks{
				"cn": {
					"HMN": slsCommon.CabinetNetworks{
						CIDR:    "10.108.4.0/22",
						Gateway: "10.108.4.1",
						VLan:    2000,
					},
					"NMN": slsCommon.CabinetNetworks{
						CIDR:    "10.107.4.0/22",
						Gateway: "10.107.4.1",
						VLan:    3000,
					},
				},
			},
		},
		hardwareExtraProperties,
	)
}

func (suite *ConfigGeneratorTestSuite) TestVerifyChassis_Hill() {
	chassiss := []string{
		"x5000c1",
		"x5000c3",
	}

	// Verify Hill Chassis
	for _, chassis := range chassiss {
		/*
			{
				"Parent": "x5000",
				"Xname": "x5000c0",
				"Type": "comptype_chassis",
				"Class": "Hill",
				"TypeString": "Chassis"
			}
		*/
		hardware, ok := suite.allHardware[chassis]
		suite.True(
			ok,
			"Unable to find xname.",
		) // TODO update all to specify the xname they are missing

		suite.Equal(
			hardware.Parent,
			"x5000",
		)
		suite.Equal(
			hardware.Xname,
			chassis,
		)
		suite.Equal(
			hardware.Type,
			slsCommon.HMSStringType("comptype_chassis"),
		)
		suite.Equal(
			hardware.Class,
			slsCommon.CabinetType("Hill"),
		)
		suite.Equal(
			hardware.TypeString,
			xnametypes.HMSType("Chassis"),
		)

		suite.Nil(
			hardware.ExtraPropertiesRaw,
			"ExtraProperties type is not nil",
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyChassisBMC_Hill() {
	chassisBMCs := []string{
		"x5000c1b0",
		"x5000c3b0",
	}

	// Verify Hill Chassis BMCs
	for _, chassisBMC := range chassisBMCs {
		/*
			{
				"Parent": "x5000c0",
				"Xname": "x5000c0b0",
				"Type": "comptype_chassis_bmc",
				"Class": "Hill",
				"TypeString": "ChassisBMC"
			}
		*/
		hardware, ok := suite.allHardware[chassisBMC]
		suite.True(
			ok,
			"Unable to find xname.",
		)

		suite.Equal(
			hardware.Parent,
			xnametypes.GetHMSCompParent(chassisBMC),
		)
		suite.Equal(
			hardware.Xname,
			chassisBMC,
		)
		suite.Equal(
			hardware.Type,
			slsCommon.HMSStringType("comptype_chassis_bmc"),
		)
		suite.Equal(
			hardware.Class,
			slsCommon.CabinetType("Hill"),
		)
		suite.Equal(
			hardware.TypeString,
			xnametypes.HMSType("ChassisBMC"),
		)

		suite.Nil(
			hardware.ExtraPropertiesRaw,
			"ExtraProperties type is not nil",
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyComputeNodes_Hill() {
	startingNID := TestSLSInputState.MountainStartingNid

	suite.verifyLiquidCooledComputeNodes(
		[]string{
			"x5000c1",
			"x5000c3",
		},
		startingNID,
		slsCommon.ClassHill,
	)
}

func (suite *ConfigGeneratorTestSuite) TestCabinet_Mountain() {
	/*
		{
			"Parent": "s0",
			"Xname": "x1000",
			"Type": "comptype_cabinet",
			"Class": "Mountain",
			"TypeString": "Cabinet",
			"ExtraProperties": {}
		}
	*/

	hardware, ok := suite.allHardware["x1000"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"s0",
	)
	suite.Equal(
		hardware.Xname,
		"x1000",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_cabinet"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("Mountain"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Cabinet"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeCabinet)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		slsCommon.ComptypeCabinet{
			Networks: map[string]map[string]slsCommon.CabinetNetworks{
				"cn": {
					"HMN": slsCommon.CabinetNetworks{
						CIDR:    "10.108.6.0/22",
						Gateway: "10.108.6.1",
						VLan:    2001,
					},
					"NMN": slsCommon.CabinetNetworks{
						CIDR:    "10.107.6.0/22",
						Gateway: "10.107.6.1",
						VLan:    3001,
					},
				},
			},
		},
		hardwareExtraProperties,
	)
}

func (suite *ConfigGeneratorTestSuite) TestVerifyChassis_Mountain() {
	chassiss := []string{
		"x1000c0",
		"x1000c1",
		"x1000c2",
		"x1000c3",
		"x1000c4",
		"x1000c5",
		"x1000c6",
		"x1000c7",
	}

	// Verify Mountain Chassis
	for _, chassis := range chassiss {
		/*
			{
				"Parent": "x1000",
				"Xname": "x1000c0",
				"Type": "comptype_chassis",
				"Class": "Mountain",
				"TypeString": "Chassis"
			}
		*/
		hardware, ok := suite.allHardware[chassis]
		suite.True(
			ok,
			"Unable to find xname %s",
			chassis,
		)

		suite.Equal(
			hardware.Parent,
			"x1000",
		)
		suite.Equal(
			hardware.Xname,
			chassis,
		)
		suite.Equal(
			hardware.Type,
			slsCommon.HMSStringType("comptype_chassis"),
		)
		suite.Equal(
			hardware.Class,
			slsCommon.CabinetType("Mountain"),
		)
		suite.Equal(
			hardware.TypeString,
			xnametypes.HMSType("Chassis"),
		)

		suite.Nil(
			hardware.ExtraPropertiesRaw,
			"ExtraProperties type is not nil",
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyChassisBMC_Mountain() {
	chassisBMCs := []string{
		"x1000c0b0",
		"x1000c1b0",
		"x1000c2b0",
		"x1000c3b0",
		"x1000c4b0",
		"x1000c5b0",
		"x1000c6b0",
		"x1000c7b0",
	}

	// Verify Mountain Chassis BMCs
	for _, chassisBMC := range chassisBMCs {
		/*
			{
				"Parent": "x1000c0",
				"Xname": "x1000c0b0",
				"Type": "comptype_chassis_bmc",
				"Class": "Mountain",
				"TypeString": "ChassisBMC"
			}
		*/
		hardware, ok := suite.allHardware[chassisBMC]
		suite.True(
			ok,
			"Unable to find xname.",
		)

		suite.Equal(
			hardware.Parent,
			xnametypes.GetHMSCompParent(chassisBMC),
		)
		suite.Equal(
			hardware.Xname,
			chassisBMC,
		)
		suite.Equal(
			hardware.Type,
			slsCommon.HMSStringType("comptype_chassis_bmc"),
		)
		suite.Equal(
			hardware.Class,
			slsCommon.CabinetType("Mountain"),
		)
		suite.Equal(
			hardware.TypeString,
			xnametypes.HMSType("ChassisBMC"),
		)

		suite.Nil(
			hardware.ExtraPropertiesRaw,
			"ExtraProperties type is not nil",
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyComputeNodes_Mountain() {
	chassiss := []string{
		"x1000c0",
		"x1000c1",
		"x1000c2",
		"x1000c3",
		"x1000c4",
		"x1000c5",
		"x1000c6",
		"x1000c7",
	}

	hillCabinetOffset := 288 // Nids for Mountain Hardware are generated after Hill
	startingNID := TestSLSInputState.MountainStartingNid + hillCabinetOffset

	// Verify Mountain Compute Nodes
	suite.verifyLiquidCooledComputeNodes(
		chassiss,
		startingNID,
		slsCommon.ClassMountain,
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_NID101() {
	/*
	  "x5004c4s17b1n0": {
	    "Parent": "x50004c3s17b1",
	    "Xname": "x50004c3s17b1n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 101,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000101"
	      ]
	    }
	  },
	*/

	hardware, ok := suite.allHardware["x5004c4s17b1n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4s17b1",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4s17b1n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NID,
		101,
	)
	suite.Equal(
		hardwareExtraProperties.Role,
		"Compute",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"nid000101",
		),
	)

}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_NID101_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j31": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j31",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x50004c3s17b1"
				],
				"VendorName": "ethernet1/1/31"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j31"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j31",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004c4s17b1"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/31",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_NID102() {
	/*
	  "x50004c3s17b2n0": {
	    "Parent": "x50004c3s17b2",
	    "Xname": "x50004c3s17b2n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 102,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000102"
	      ]
	    }
	  },
	*/

	hardware, ok := suite.allHardware["x5004c4s17b2n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4s17b2",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4s17b2n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NID,
		102,
	)
	suite.Equal(
		hardwareExtraProperties.Role,
		"Compute",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"nid000102",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_NID102_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j32": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j32",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x50004c3s17b2"
				],
				"VendorName": "ethernet1/1/32"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j32"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j32",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004c4s17b2"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/32",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_NID103() {
	/*
	  "x50004c3s17b3n0": {
	    "Parent": "x50004c3s17b3",
	    "Xname": "x50004c3s17b3n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 103,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000103"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x5004c4s17b3n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4s17b3",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4s17b3n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NID,
		103,
	)
	suite.Equal(
		hardwareExtraProperties.Role,
		"Compute",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"nid000103",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_NID103_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j33": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j33",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x50004c3s17b3"
				],
				"VendorName": "ethernet1/1/33"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j33"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j33",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004c4s17b3"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/33",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_NID104() {
	/*
	  "x5004c4s17b4n0": {
	    "Parent": "x50004c3s17b4",
	    "Xname": "x50004c3s17b4n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 104,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000104"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x5004c4s17b4n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4s17b4",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4s17b4n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NID,
		104,
	)
	suite.Equal(
		hardwareExtraProperties.Role,
		"Compute",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"nid000104",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_NID104_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j34": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j34",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x50004c3s17b4"
				],
				"VendorName": "ethernet1/1/34"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j34"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j34",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004c4s17b4"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/34",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_CMC() {
	/*
	  "x50004c3s17b999": {
	    "Parent": "x50004c3s17",
	    "Xname": "x50004c3s17b999",
	    "Type": "comptype_ncard",
	    "Class": "River",
	    "TypeString": "NodeBMC"
	  },
	*/
	hardware, ok := suite.allHardware["x5004c4s17b999"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4s17",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4s17b999",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_ncard"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("NodeBMC"),
	)

	suite.Nil(
		hardware.ExtraPropertiesRaw,
		"ExtraProperties type is not nil",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverComputeNode_CMC_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j30": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j30",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x50004c3s17b999"
				],
				"VendorName": "ethernet1/1/30"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j30"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j30",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004c4s17b999"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/30",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverWorkerNode() {
	/*
	  "x5004c4s19b0n0": {
	    "Parent": "x5004c4s19b0",
	    "Xname": "x5004c4s19b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": TODO,
	      "Role": "Management",
	      "SubRole": "Worker",
	      "Aliases": [
	        "ncn-w050"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x5004c4s19b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4s19b0",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4s19b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.Role,
		"Management",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"Worker",
	)
	suite.True(
		stringArrayContains(
			hardwareExtraProperties.Aliases,
			"ncn-w050",
		),
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverWorkerNode_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j35": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j35",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x5004c4s19b0"
				],
				"VendorName": "ethernet1/1/35"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j35"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j35",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004c4s19b0"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/35",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverApplicationNode_UAN() {
	/*
		  "x5004c4s20b0n0": {
		    "Parent": "x5004c4s20b0",
		    "Xname": "x5004c4s20b0n0",
		    "Type": "comptype_node",
		    "Class": "River",
		    "TypeString": "Node",
		    "ExtraProperties": {
		      "Role": "Application",
			  "SubRole": "UAN"
			  "Aliases": [
				  "uan-50"
			  ]
		    }
		  },
	*/
	hardware, ok := suite.allHardware["x5004c4s20b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4s20b0",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4s20b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	// TODO: CASMHMS-3598
	// suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(
		hardwareExtraProperties.Role,
		"Application",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"UAN",
	)
	suite.Equal(
		hardwareExtraProperties.Aliases,
		[]string{"uan-50"},
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverApplicationNode_UAN_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j36": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j36",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x5004c4s20b0"
				],
				"VendorName": "ethernet1/1/36"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j36"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j36",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004c4s20b0"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/36",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverApplicationNode_LNetRouter() {
	/*
		  "x5004c4s21b0n0": {
		    "Parent": "x5004c4s21b0",
		    "Xname": "x5004c4s21b0n0",
		    "Type": "comptype_node",
		    "Class": "River",
		    "TypeString": "Node",
		    "ExtraProperties": {
		      "Role": "Application",
			  "SubRole": "LNETRouter"
			  "Aliases": [
				  "lnet-50"
			  ]
		    }
		  },
	*/
	hardware, ok := suite.allHardware["x5004c4s21b0n0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4s21b0",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4s21b0n0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_node"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("Node"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	// TODO: CASMHMS-3598
	// suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(
		hardwareExtraProperties.Role,
		"Application",
	)
	suite.Equal(
		hardwareExtraProperties.SubRole,
		"LNETRouter",
	)
	suite.Equal(
		hardwareExtraProperties.Aliases,
		[]string{"lnet-50"},
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverApplicationNode_LNetRouter_MgmtSwitchConnector_DifferentCabinet() {
	// This Management switch connector is located in a river cabinet

	/*
		  "x3001c0w42j20": {
			"Parent": "x3001c0w42",
			"Xname": "x3001c0w42j20",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x5004c4s21b0"
				],
				"VendorName": "ethernet1/1/20"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x3001c0w42j20"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x3001c0w42",
	)
	suite.Equal(
		hardware.Xname,
		"x3001c0w42j20",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004c4s21b0"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/20",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverRosetta() {
	/*
	  "x5004c4r1b0": {
	    "Parent": "x5004c4r1",
	    "Xname": "x5004c4r1b0",
	    "Type": "comptype_rtr_bmc",
	    "Class": "River",
	    "TypeString": "RouterBMC",
	    "ExtraProperties": {
	      "Username": "vault://hms-creds/x5004c4r1b0",
	      "Password": "vault://hms-creds/x5004c4r1b0"
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x5004c4r1b0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4r1",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4r1b0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_rtr_bmc"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("RouterBMC"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeRtrBmc)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.Username,
		"vault://hms-creds/x5004c4r1b0",
	)
	suite.Equal(
		hardwareExtraProperties.Password,
		"vault://hms-creds/x5004c4r1b0",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_RiverRosetta_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j37": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j37",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x5004c4s19b0"
				],
				"VendorName": "ethernet1/1/35"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j37"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j37",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004c4r1b0"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/37",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_PDU0() {
	/*
	   "x5004m0": {
	     "Parent": "x5004",
	     "Xname": "x5004m0",
	     "Type": "comptype_cab_pdu_controller",
	     "Class": "River",
	     "TypeString": "CabinetPDUController",
	   }
	*/

	hardware, ok := suite.allHardware["x5004m0"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004",
	)
	suite.Equal(
		hardware.Xname,
		"x5004m0",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_cab_pdu_controller"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("CabinetPDUController"),
	)

	suite.Nil(
		hardware.ExtraPropertiesRaw,
		"ExtraProperties type is not nil",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_PDU0_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j38": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j38",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x5004m0"
				],
				"VendorName": "ethernet1/1/38"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j38"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j38",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004m0"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/38",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_PDU1() {
	/*
	   "x5004m1": {
	     "Parent": "x5004",
	     "Xname": "x5004m1",
	     "Type": "comptype_cab_pdu_controller",
	     "Class": "River",
	     "TypeString": "CabinetPDUController",
	   }
	*/

	hardware, ok := suite.allHardware["x5004m1"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004",
	)
	suite.Equal(
		hardware.Xname,
		"x5004m1",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_cab_pdu_controller"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("CabinetPDUController"),
	)

	suite.Nil(
		hardware.ExtraPropertiesRaw,
		"ExtraProperties type is not nil",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_PDU1_MgmtSwitchConnector() {
	/*
		  "x5004c4w16j39": {
			"Parent": "x5004c4w16",
			"Xname": "x5004c4w16j39",
			"Type": "comptype_mgmt_switch_connector",
			"Class": "River",
			"TypeString": "MgmtSwitchConnector",
			"ExtraProperties": {
				"NodeNics": [
					"x5004m1"
				],
				"VendorName": "ethernet1/1/39"
			}
		  }
	*/
	hardware, ok := suite.allHardware["x5004c4w16j39"]
	suite.True(
		ok,
		"Unable to find xname.",
	)

	suite.Equal(
		hardware.Parent,
		"x5004c4w16",
	)
	suite.Equal(
		hardware.Xname,
		"x5004c4w16j39",
	)
	suite.Equal(
		hardware.Type,
		slsCommon.HMSStringType("comptype_mgmt_switch_connector"),
	)
	suite.Equal(
		hardware.Class,
		slsCommon.CabinetType("River"),
	)
	suite.Equal(
		hardware.TypeString,
		xnametypes.HMSType("MgmtSwitchConnector"),
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		hardwareExtraProperties.NodeNics,
		[]string{"x5004m1"},
	)
	suite.Equal(
		hardwareExtraProperties.VendorName,
		"1/1/39",
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_VerifyChassis() {
	chassiss := []string{
		// 1 Liquid-cooled chassis
		"x5001c0",

		// 2 Liquid-cooled chassis
		"x5002c0",
		"x5002c1",

		// 3 Liquid-cooled chassis
		"x5003c0",
		"x5003c1",
		"x5003c2",

		// 1 Liquid-cooled chassis and 1 air cooled chassis
		// do not expect to see a chassis bmc for c3
		// TODO should there be a "River" chassis within the cabinet?
		"x5004c0",
	}

	// Verify Chassis
	for _, chassis := range chassiss {
		/*
			{
				"Parent": "x5000",
				"Xname": "x5000c0",
				"Type": "comptype_chassis",
				"Class": "Hill",
				"TypeString": "Chassis"
			}
		*/
		hardware, ok := suite.allHardware[chassis]
		suite.True(
			ok,
			"Unable to find xname.",
		) // TODO update all to specify the xname they are missing

		suite.Equal(
			hardware.Parent,
			xnametypes.GetHMSCompParent(chassis),
		)
		suite.Equal(
			hardware.Xname,
			chassis,
		)
		suite.Equal(
			hardware.Type,
			slsCommon.HMSStringType("comptype_chassis"),
		)
		suite.Equal(
			hardware.Class,
			slsCommon.CabinetType("Hill"),
		)
		suite.Equal(
			hardware.TypeString,
			xnametypes.HMSType("Chassis"),
		)

		suite.Nil(
			hardware.ExtraPropertiesRaw,
			"ExtraProperties type is not nil",
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_VerifyChassisBMC() {
	chassisBMCs := []string{
		// 1 Liquid-cooled chassis
		"x5001c0b0",

		// 2 Liquid-cooled chassis
		"x5002c0b0",
		"x5002c1b0",

		// 3 Liquid-cooled chassis
		"x5003c0b0",
		"x5003c1b0",
		"x5003c2b0",

		// 1 Liquid-cooled chassis and 1 air cooled chassis
		// do not expect to see a chassis bmc for c3
		// TODO should there be a "River" chassis within the cabinet?
		"x5004c0b0",
	}

	// Verify Hill Chassis BMCs
	for _, chassisBMC := range chassisBMCs {
		/*
			{
				"Parent": "x5000c0",
				"Xname": "x5000c0b0",
				"Type": "comptype_chassis_bmc",
				"Class": "Hill",
				"TypeString": "ChassisBMC"
			}
		*/
		hardware, ok := suite.allHardware[chassisBMC]
		suite.True(
			ok,
			"Unable to find xname.",
		)

		suite.Equal(
			hardware.Parent,
			xnametypes.GetHMSCompParent(chassisBMC),
		)
		suite.Equal(
			hardware.Xname,
			chassisBMC,
		)
		suite.Equal(
			hardware.Type,
			slsCommon.HMSStringType("comptype_chassis_bmc"),
		)
		suite.Equal(
			hardware.Class,
			slsCommon.CabinetType("Hill"),
		)
		suite.Equal(
			hardware.TypeString,
			xnametypes.HMSType("ChassisBMC"),
		)

		suite.Nil(
			hardware.ExtraPropertiesRaw,
			"ExtraProperties type is not nil",
		)
	}
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_VerifyComputeNodes_5001() {
	startingNID := TestSLSInputState.MountainStartingNid + 64

	suite.verifyLiquidCooledComputeNodes(
		[]string{"x5001c0"},
		startingNID,
		slsCommon.ClassHill,
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_VerifyComputeNodes_5002() {
	startingNID := TestSLSInputState.MountainStartingNid + 96

	suite.verifyLiquidCooledComputeNodes(
		[]string{
			"x5002c0",
			"x5002c1",
		},
		startingNID,
		slsCommon.ClassHill,
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_VerifyComputeNodes_5003() {
	startingNID := TestSLSInputState.MountainStartingNid + 160

	suite.verifyLiquidCooledComputeNodes(
		[]string{
			"x5003c0",
			"x5003c1",
			"x5003c2",
		},
		startingNID,
		slsCommon.ClassHill,
	)
}

func (suite *ConfigGeneratorTestSuite) TestEX2500_VerifyComputeNodes_5004() {
	startingNID := TestSLSInputState.MountainStartingNid + 256

	suite.verifyLiquidCooledComputeNodes(
		[]string{"x5004c0"},
		startingNID,
		slsCommon.ClassHill,
	)
}

// verifyLiquidCooledComputeNodes is a helper function to verify the liquid cooled nodes in a given chassis
func (suite *ConfigGeneratorTestSuite) verifyLiquidCooledComputeNodes(
	chassiss []string, startingNID int, class slsCommon.CabinetType,
) {
	nodeBMCs := []string{
		"s0b0",
		"s0b1", // Slot 0
		"s1b0",
		"s1b1", // Slot 1
		"s2b0",
		"s2b1", // Slot 2
		"s3b0",
		"s3b1", // Slot 3
		"s4b0",
		"s4b1", // Slot 4
		"s5b0",
		"s5b1", // Slot 5
		"s6b0",
		"s6b1", // Slot 6
		"s7b0",
		"s7b1", // Slot 7
	}

	nodes := []string{
		"n0",
		"n1",
	}

	expectedNid := startingNID

	for _, chassis := range chassiss {
		for _, nodeBMCSuffix := range nodeBMCs {
			for _, node := range nodes {
				/*
				   "x5000c1s0b0n0": {
				     "Parent": "x5000c1s0b0",
				     "Xname": "x5000c1s0b0n0",
				     "Type": "comptype_node",
				     "Class": "Hill",
				     "TypeString": "Node",
				     "ExtraProperties": {
				       "NID": 1000,
				       "Role": "Compute",
				       "Aliases": [
				         "nid001000"
				       ]
				     }
				   }
				*/

				// Calculate xnames for BMC and node
				nodeBMC := chassis + nodeBMCSuffix
				nodeXname := nodeBMC + node

				hardware, ok := suite.allHardware[nodeXname]
				suite.True(
					ok,
					"Unable to find xname.",
				)

				suite.Equal(
					hardware.Parent,
					nodeBMC,
				)
				suite.Equal(
					hardware.Xname,
					nodeXname,
				)
				suite.Equal(
					hardware.Type,
					slsCommon.HMSStringType("comptype_node"),
				)
				suite.Equal(
					hardware.Class,
					class,
				)
				suite.Equal(
					hardware.TypeString,
					xnametypes.HMSType("Node"),
				)

				hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
				suite.True(
					ok,
					"ExtraProperties type is not expected type.",
				)

				suite.Equal(
					hardwareExtraProperties.Role,
					"Compute",
				)
				suite.Equal(
					hardwareExtraProperties.NID,
					expectedNid,
				)

				expectedAlias := fmt.Sprintf(
					"nid%06d",
					expectedNid,
				)
				suite.Equal(
					hardwareExtraProperties.Aliases,
					[]string{expectedAlias},
				)

				expectedNid++
			}
		}
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyNoUnexpectedLiquidCooledChassis() {
	expectedChassis := map[string]bool{}

	// Hill
	expectedChassis["x5000c1"] = true
	expectedChassis["x5000c3"] = true

	// EX25000
	// 1 Liquid-cooled chassis
	expectedChassis["x5001c0"] = true

	// 2 Liquid-cooled chassis
	expectedChassis["x5002c0"] = true
	expectedChassis["x5002c1"] = true

	// 3 Liquid-cooled chassis
	expectedChassis["x5003c0"] = true
	expectedChassis["x5003c1"] = true
	expectedChassis["x5003c2"] = true

	// 1 Liquid-cooled chassis and 1 air cooled chassis
	// do not expect to see a chassis bmc for c3
	// TODO should there be a "River" chassis within the cabinet?
	expectedChassis["x5004c0"] = true

	// Mountain
	expectedChassis["x1000c0"] = true
	expectedChassis["x1000c1"] = true
	expectedChassis["x1000c2"] = true
	expectedChassis["x1000c3"] = true
	expectedChassis["x1000c4"] = true
	expectedChassis["x1000c5"] = true
	expectedChassis["x1000c6"] = true
	expectedChassis["x1000c7"] = true

	for xname, hardware := range suite.allHardware {
		if xnametypes.GetHMSType(xname) != xnametypes.Chassis {
			continue
		}

		if present := expectedChassis[xname]; !present && hardware.Class != slsCommon.ClassRiver {
			suite.Fail(
				"Found unexpected liquid-cooled chassis",
				xname,
			)
		}
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyNoUnexpectedLiquidCooledChassisBMCs() {
	expectedChassis := map[string]bool{}

	// Hill
	expectedChassis["x5000c1b0"] = true
	expectedChassis["x5000c3b0"] = true

	// EX25000
	// 1 Liquid-cooled chassis
	expectedChassis["x5001c0b0"] = true

	// 2 Liquid-cooled chassis
	expectedChassis["x5002c0b0"] = true
	expectedChassis["x5002c1b0"] = true

	// 3 Liquid-cooled chassis
	expectedChassis["x5003c0b0"] = true
	expectedChassis["x5003c1b0"] = true
	expectedChassis["x5003c2b0"] = true

	// 1 Liquid-cooled chassis and 1 air cooled chassis
	// do not expect to see a chassis bmc for c4
	expectedChassis["x5004c0b0"] = true

	// Mountain
	expectedChassis["x1000c0b0"] = true
	expectedChassis["x1000c1b0"] = true
	expectedChassis["x1000c2b0"] = true
	expectedChassis["x1000c3b0"] = true
	expectedChassis["x1000c4b0"] = true
	expectedChassis["x1000c5b0"] = true
	expectedChassis["x1000c6b0"] = true
	expectedChassis["x1000c7b0"] = true

	for xname, hardware := range suite.allHardware {
		if xnametypes.GetHMSType(xname) != xnametypes.ChassisBMC {
			continue
		}

		if present := expectedChassis[xname]; !present && hardware.Class != slsCommon.ClassRiver {
			suite.Fail(
				"Found unexpected liquid-cooled chassis bmc",
				xname,
			)
		}
	}
}

func (suite *ConfigGeneratorTestSuite) TestBuildSLSHardware() {
	slsHardware := suite.generator.buildSLSHardware(
		xnames.Cabinet{Cabinet: 1234},
		slsCommon.ClassMountain,
		nil,
	)
	suite.T().Log(slsHardware)
}

func (suite *ConfigGeneratorTestSuite) TestGetLiquidCooledHardwareForCabinet_Mountain() {
	cabinetTemplate := CabinetTemplate{
		Xname: xnames.Cabinet{
			Cabinet: 1000,
		},
		Class:                   slsCommon.ClassMountain,
		LiquidCooledChassisList: DefaultMountainChassisList,
	}

	slsHardware := suite.generator.getLiquidCooledHardwareForCabinet(cabinetTemplate)
	suite.NotEmpty(slsHardware)
	for _, hardware := range slsHardware {
		suite.T().Log(hardware.Xname)
	}
}

func TestConfigGeneratorTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(ConfigGeneratorTestSuite),
	)
}

type ApplicationNodeConfigTestSuite struct {
	suite.Suite
}

func (suite *ApplicationNodeConfigTestSuite) TestApplicationNodeConfigNormalize_NormalizedInput() {
	// This application node config already contains normalized data
	applicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0": {"uan-01"},
			"x3000c0s28b0n0": {"gateway-01"},
			"x3000c0s29b0n0": {"visualization-01"},
		},
	}

	expectedApplicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0": {"uan-01"},
			"x3000c0s28b0n0": {"gateway-01"},
			"x3000c0s29b0n0": {"visualization-01"},
		},
	}

	err := applicationNodeConfig.Normalize()
	suite.NoError(err)
	suite.Equal(
		expectedApplicationNodeConfig,
		applicationNodeConfig,
	)
}

func (suite *ApplicationNodeConfigTestSuite) TestApplicationNodeConfigNormalize_UnNormalizedInput() {
	applicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"Vn",
			"LNET",
		},
		PrefixHSMSubroles: map[string]string{
			"Vn":   "Visualization",
			"lNet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x03000c0s026b00n00":  {"uan-01"},
			"x03000c00s028b00n00": {"gateway-01"},
			"x03000c00s029b00n00": {"visualization-01"},
		},
	}

	expectedApplicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0": {"uan-01"},
			"x3000c0s28b0n0": {"gateway-01"},
			"x3000c0s29b0n0": {"visualization-01"},
		},
	}

	err := applicationNodeConfig.Normalize()
	suite.NoError(err)
	suite.Equal(
		expectedApplicationNodeConfig,
		applicationNodeConfig,
	)
}

func (suite *ApplicationNodeConfigTestSuite) TestApplicationNodeConfigNormalize_DuplicatePrefixSubroleKeys() {
	applicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
			"LNET": "LNETRouter2",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0": {"uan-01"},
			"x3000c0s28b0n0": {"gateway-01"},
			"x3000c0s29b0n0": {"visualization-01"},
		},
	}

	// When normalize fails is is expected that the
	expectedApplicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
			"LNET": "LNETRouter2",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0": {"uan-01"},
			"x3000c0s28b0n0": {"gateway-01"},
			"x3000c0s29b0n0": {"visualization-01"},
		},
	}

	err := applicationNodeConfig.Normalize()
	suite.Error(err) // TODO check error contents
	suite.Equal(
		expectedApplicationNodeConfig,
		applicationNodeConfig,
	)
}

func (suite *ApplicationNodeConfigTestSuite) TestApplicationNodeConfigNormalize_DuplicateXnameAliasKeys() {
	applicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0":  {"uan-01"},
			"x3000c0s28b0n0":  {"gateway-01"},
			"x3000c0s29b0n0":  {"visualization-01"},
			"x3000c0s029b0n0": {"visualization-01"}, // Has extra zero padding for the slot
		},
	}

	// When normalize fails is is expected that the
	expectedApplicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0":  {"uan-01"},
			"x3000c0s28b0n0":  {"gateway-01"},
			"x3000c0s29b0n0":  {"visualization-01"},
			"x3000c0s029b0n0": {"visualization-01"},
		},
	}

	err := applicationNodeConfig.Normalize()
	suite.Error(err) // TODO check error contents
	suite.Equal(
		expectedApplicationNodeConfig,
		applicationNodeConfig,
	)
}

func (suite *ApplicationNodeConfigTestSuite) TestApplicationNodeConfigValidate_HappyPath() {
	applicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0": {"uan-01"},
			"x3000c0s28b0n0": {"gateway-01"},
			"x3000c0s29b0n0": {"visualization-01"},
		},
	}

	err := applicationNodeConfig.Validate()
	suite.NoError(err)
}

func (suite *ApplicationNodeConfigTestSuite) TestApplicationNodeConfigValidate_InvalidXname() {
	applicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000f0s26b0n0": {"uan-01"}, // Invalid Xname
			"x3000c0s28b0n0": {"gateway-01"},
			"x3000c0s29b0n0": {"visualization-01"},
		},
	}

	err := applicationNodeConfig.Validate()
	suite.EqualError(
		err,
		"invalid xname for application node used as key in Aliases map: x3000f0s26b0n0",
	)
}

func (suite *ApplicationNodeConfigTestSuite) TestApplicationNodeConfigValidate_WrongXNameType() {
	applicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0":   {"uan-01"}, // Xname for BMC type
			"x3000c0s28b0n0": {"gateway-01"},
			"x3000c0s29b0n0": {"visualization-01"},
		},
	}

	err := applicationNodeConfig.Validate()
	suite.EqualError(
		err,
		"invalid type NodeBMC for Application xname in Aliases map: x3000c0s26b0",
	)
}

func (suite *ApplicationNodeConfigTestSuite) TestApplicationNodeConfigValidate_DuplicateAlias() {
	applicationNodeConfig := GeneratorApplicationNodeConfig{
		Prefixes: []string{
			"vn",
			"lnet",
		},
		PrefixHSMSubroles: map[string]string{
			"vn":   "Visualization",
			"lnet": "LNETRouter",
		},
		Aliases: map[string][]string{
			"x3000c0s26b0n0": {"uan-01"},
			"x3000c0s28b0n0": {"uan-01"},
			"x3000c0s29b0n0": {"visualization-01"},
		},
	}

	err := applicationNodeConfig.Validate()
	suite.Error(err)
	// Since we are iterating over maps, the key order is not guaranteed, so the following condition some times fails
	// suite.EqualError(err, "found duplicate application node alias: uan-01 for xnames x3000c0s26b0n0 x3000c0s28b0n0")
}

func TestApplicationNodeConfigTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(ApplicationNodeConfigTestSuite),
	)
}

type ParseSourceCabinetFromRowTestSuite struct {
	suite.Suite
}

func (suite *ParseSourceCabinetFromRowTestSuite) TestExpectedPath() {
	g := StateGenerator{}

	cabinet, err := g.parseSourceCabinetFromRow(shcdParser.HMNRow{SourceRack: "x3000"})
	suite.NoError(err)
	suite.Equal(
		3000,
		cabinet.Cabinet,
	)
}

func (suite *ParseSourceCabinetFromRowTestSuite) TestCaptialX() {
	g := StateGenerator{}

	cabinet, err := g.parseSourceCabinetFromRow(shcdParser.HMNRow{SourceRack: "X3000"})
	suite.NoError(err)
	suite.Equal(
		3000,
		cabinet.Cabinet,
	)
}

func (suite *ParseSourceCabinetFromRowTestSuite) TestMissingX() {
	g := StateGenerator{}

	cabinet, err := g.parseSourceCabinetFromRow(shcdParser.HMNRow{SourceRack: "3000"})
	suite.NoError(err)
	suite.Equal(
		3000,
		cabinet.Cabinet,
	)
}

func (suite *ParseSourceCabinetFromRowTestSuite) TestMalformedString() {
	g := StateGenerator{}

	_, err := g.parseSourceCabinetFromRow(shcdParser.HMNRow{SourceRack: "foo"})
	suite.Error(err)
}

func TestParseSourceCabinetFromRowTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(ParseSourceCabinetFromRowTestSuite),
	)
}

type GetNodeHardwareFromRowTestSuite struct {
	suite.Suite

	generator StateGenerator
}

func (suite *GetNodeHardwareFromRowTestSuite) SetupSuite() {
	// Setup logger for testing
	encoderCfg := zap.NewProductionEncoderConfig()
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
	)

	// Normalize and validate the application node config
	err := TestSLSInputState.ApplicationNodeConfig.Normalize()
	suite.NoError(err)

	err = TestSLSInputState.ApplicationNodeConfig.Validate()
	suite.NoError(err)

	// Only specify the information we need for testing. We will be passing in our own HMNRows in later
	suite.generator = NewStateGenerator(
		logger,
		TestSLSInputState,
		[]shcdParser.HMNRow{},
	)
}

func (suite *GetNodeHardwareFromRowTestSuite) TestEX2500_UAN() {
	expectedXname := "x5004c4s20b0n0"

	// As a sanity check, lets see if we can get the expected node alias
	suite.Equal(
		[]string{"uan-50"},
		suite.generator.getApplicationNodeAlias(expectedXname),
	)
	suite.True(suite.generator.canCabinetContainAirCooledHardware("x5004"))

	row := shcdParser.HMNRow{
		Source:              "uan50",
		SourceRack:          "x5004",
		SourceLocation:      "u20",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j36",
	}

	// Process the row
	hardware := suite.generator.getNodeHardwareFromRow(row)
	suite.NotEmpty(hardware)

	suite.Equal(
		"x5004c4s20b0",
		hardware.Parent,
	)
	suite.Equal(
		"x5004c4s20b0n0",
		hardware.Xname,
	)
	suite.Equal(
		slsCommon.Node,
		hardware.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		hardware.Class,
	)
	suite.Equal(
		xnametypes.Node,
		hardware.TypeString,
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		"Application",
		hardwareExtraProperties.Role,
	)
	suite.Equal(
		"UAN",
		hardwareExtraProperties.SubRole,
	)
	suite.Equal(
		[]string{"uan-50"},
		hardwareExtraProperties.Aliases,
	)
}

func (suite *GetNodeHardwareFromRowTestSuite) TestEX2500_LNETRouter() {
	expectedXname := "x5004c4s21b0n0"

	// As a sanity check, lets see if we can get the expected node alias
	suite.Equal(
		[]string{"lnet-50"},
		suite.generator.getApplicationNodeAlias(expectedXname),
	)
	suite.True(suite.generator.canCabinetContainAirCooledHardware("x5004"))

	row := shcdParser.HMNRow{
		Source:              "lnet50",
		SourceRack:          "x5004",
		SourceLocation:      "u21",
		DestinationRack:     "x3001",
		DestinationLocation: "u42",
		DestinationPort:     "j20",
	}

	// Process the row
	hardware := suite.generator.getNodeHardwareFromRow(row)
	suite.NotEmpty(hardware)

	suite.Equal(
		"x5004c4s21b0",
		hardware.Parent,
	)
	suite.Equal(
		"x5004c4s21b0n0",
		hardware.Xname,
	)
	suite.Equal(
		slsCommon.Node,
		hardware.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		hardware.Class,
	)
	suite.Equal(
		xnametypes.Node,
		hardware.TypeString,
	)

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(slsCommon.ComptypeNode)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		"Application",
		hardwareExtraProperties.Role,
	)
	suite.Equal(
		"LNETRouter",
		hardwareExtraProperties.SubRole,
	)
	suite.Equal(
		[]string{"lnet-50"},
		hardwareExtraProperties.Aliases,
	)
}

func TestGetNodeHardwareFromRowTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(GetNodeHardwareFromRowTestSuite),
	)
}

type GetSwitchConnectionForHardwareTestSuite struct {
	suite.Suite

	generator StateGenerator
}

func (suite *GetSwitchConnectionForHardwareTestSuite) SetupSuite() {
	// Setup logger for testing
	encoderCfg := zap.NewProductionEncoderConfig()
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
	)

	// Normalize and validate the application node config
	err := TestSLSInputState.ApplicationNodeConfig.Normalize()
	suite.NoError(err)

	err = TestSLSInputState.ApplicationNodeConfig.Validate()
	suite.NoError(err)

	// Only specify the information we need for testing. We will be passing in our own HMNRows in later
	suite.generator = NewStateGenerator(
		logger,
		TestSLSInputState,
		[]shcdParser.HMNRow{},
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestRiverCabinet_CMC() {
	row := shcdParser.HMNRow{
		Source:              "SubRack-001-cmc",
		SourceRack:          "x3000",
		SourceLocation:      "u19",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p38",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s19b999"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	suite.Equal(
		"x3000c0w22",
		connection.Parent,
	)
	suite.Equal(
		"x3000c0w22j38",
		connection.Xname,
	)
	suite.Equal(
		slsCommon.MgmtSwitchConnector,
		connection.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		connection.Class,
	)
	suite.Equal(
		xnametypes.MgmtSwitchConnector,
		connection.TypeString,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		connectionExtraProperties.NodeNics,
		[]string{"x3000c0s19b999"},
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"ethernet1/1/38",
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestRiverCabinet_PDU() {
	row := shcdParser.HMNRow{
		Source:              "x3000p0",
		SourceRack:          "x3000",
		SourceLocation:      "p0",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "j48",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000m0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	suite.Equal(
		"x3000c0w22",
		connection.Parent,
	)
	suite.Equal(
		"x3000c0w22j48",
		connection.Xname,
	)
	suite.Equal(
		slsCommon.MgmtSwitchConnector,
		connection.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		connection.Class,
	)
	suite.Equal(
		xnametypes.MgmtSwitchConnector,
		connection.TypeString,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		connectionExtraProperties.NodeNics,
		[]string{"x3000m0"},
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"ethernet1/1/48",
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestRiverCabinet_NodeBMC() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	suite.Equal(
		"x3000c0w22",
		connection.Parent,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
	suite.Equal(
		slsCommon.MgmtSwitchConnector,
		connection.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		connection.Class,
	)
	suite.Equal(
		xnametypes.MgmtSwitchConnector,
		connection.TypeString,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		connectionExtraProperties.NodeNics,
		[]string{"x3000c0s9b0"},
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"ethernet1/1/28",
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestRiverCabinet_RouterBMC() {
	row := shcdParser.HMNRow{
		Source:              "sw-hsn001",
		SourceRack:          "x3000",
		SourceLocation:      "u22",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p47",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0r22b0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	suite.Equal(
		"x3000c0w22",
		connection.Parent,
	)
	suite.Equal(
		"x3000c0w22j47",
		connection.Xname,
	)
	suite.Equal(
		slsCommon.MgmtSwitchConnector,
		connection.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		connection.Class,
	)
	suite.Equal(
		xnametypes.MgmtSwitchConnector,
		connection.TypeString,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		connectionExtraProperties.NodeNics,
		[]string{"x3000c0r22b0"},
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"ethernet1/1/47",
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestEXCabinet_CMC() {
	row := shcdParser.HMNRow{
		Source:              "SubRack-004-CMC",
		SourceRack:          "x5004",
		SourceLocation:      "u17",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j30",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x5004c4s17b999"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	suite.Equal(
		"x5004c4w16",
		connection.Parent,
	)
	suite.Equal(
		"x5004c4w16j30",
		connection.Xname,
	)
	suite.Equal(
		slsCommon.MgmtSwitchConnector,
		connection.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		connection.Class,
	)
	suite.Equal(
		xnametypes.MgmtSwitchConnector,
		connection.TypeString,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		connectionExtraProperties.NodeNics,
		[]string{"x5004c4s17b999"},
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"1/1/30",
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestEXCabinet_PDU() {
	row := shcdParser.HMNRow{
		Source:              "x5004p0",
		SourceRack:          "x5004",
		SourceLocation:      "p0",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j38",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x5004m0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	suite.Equal(
		"x5004c4w16",
		connection.Parent,
	)
	suite.Equal(
		"x5004c4w16j38",
		connection.Xname,
	)
	suite.Equal(
		slsCommon.MgmtSwitchConnector,
		connection.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		connection.Class,
	)
	suite.Equal(
		xnametypes.MgmtSwitchConnector,
		connection.TypeString,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		connectionExtraProperties.NodeNics,
		[]string{"x5004m0"},
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"1/1/38",
	)
}
func (suite *GetSwitchConnectionForHardwareTestSuite) TestEXCabinet_NodeBMC() {
	row := shcdParser.HMNRow{
		Source:              "wn50",
		SourceRack:          "x5004",
		SourceLocation:      "u19",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j35",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x5004c4s19b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	suite.Equal(
		"x5004c4w16",
		connection.Parent,
	)
	suite.Equal(
		"x5004c4w16j35",
		connection.Xname,
	)
	suite.Equal(
		slsCommon.MgmtSwitchConnector,
		connection.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		connection.Class,
	)
	suite.Equal(
		xnametypes.MgmtSwitchConnector,
		connection.TypeString,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		connectionExtraProperties.NodeNics,
		[]string{"x5004c4s19b0"},
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"1/1/35",
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestEXCabinet_RouterBMC() {
	row := shcdParser.HMNRow{
		Source:              "sw-hsn50",
		SourceRack:          "x5004",
		SourceLocation:      "u1",
		DestinationRack:     "x5004",
		DestinationLocation: "u16",
		DestinationPort:     "j37",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x5004c4r1b0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	suite.Equal(
		"x5004c4w16",
		connection.Parent,
	)
	suite.Equal(
		"x5004c4w16j37",
		connection.Xname,
	)
	suite.Equal(
		slsCommon.MgmtSwitchConnector,
		connection.Type,
	)
	suite.Equal(
		slsCommon.ClassRiver,
		connection.Class,
	)
	suite.Equal(
		xnametypes.MgmtSwitchConnector,
		connection.TypeString,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)

	suite.Equal(
		connectionExtraProperties.NodeNics,
		[]string{"x5004c4r1b0"},
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"1/1/37",
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestinationRack_UppercaseXCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "X3000",
		DestinationLocation: "u22",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)

}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestinationRack_LowercaseXCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestinationRack_MissingXCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "3000",
		DestinationLocation: "u22",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestionLocation_UppercaseUCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "U22",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestionLocation_LowercaseUCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestionLocation_MissingUCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "22",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestionPort_UppercaseJCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "J28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}
func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestionPort_UppercasePCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "P28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}
func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestionPort_LowercaseJCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "j28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}
func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestionPort_LowercasePCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestDestionPort_MissingCharacter() {
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)

	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)
	suite.Equal(
		"x3000c0w22j28",
		connection.Xname,
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestDellMgmtSwitch() {
	// x3000c0w22 is Dell
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "22",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)
	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"ethernet1/1/28",
	)
}

func (suite *GetSwitchConnectionForHardwareTestSuite) TestArubaMgmtSwitch() {
	// x3001c0w42 is Aruba
	row := shcdParser.HMNRow{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3001",
		DestinationLocation: "42",
		DestinationPort:     "p28",
	}

	hardware := suite.generator.buildSLSHardware(
		xnames.FromString("x3000c0s9b0n0"),
		slsCommon.ClassRiver,
		nil,
	)
	connection := suite.generator.getSwitchConnectionForHardware(
		hardware,
		row,
	)

	connectionExtraProperties, ok := connection.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitchConnector)
	suite.True(
		ok,
		"ExtraProperties type is not expected type.",
	)
	suite.Equal(
		connectionExtraProperties.VendorName,
		"1/1/28",
	)
}

func TestGetSwitchConnectionForHardwareTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(GetSwitchConnectionForHardwareTestSuite),
	)
}

type CabinetHelpersTestSuite struct {
	suite.Suite

	generator StateGenerator
}

func (suite *CabinetHelpersTestSuite) SetupSuite() {
	// Setup logger for testing
	encoderCfg := zap.NewProductionEncoderConfig()
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
	)

	// Normalize and validate the application node config
	err := TestSLSInputState.ApplicationNodeConfig.Normalize()
	suite.NoError(err)

	err = TestSLSInputState.ApplicationNodeConfig.Validate()
	suite.NoError(err)

	// Only specify the information we need for testing. We will be passing in our own HMNRows in later
	suite.generator = NewStateGenerator(
		logger,
		TestSLSInputState,
		[]shcdParser.HMNRow{},
	)
}

func (suite *CabinetHelpersTestSuite) TestDetermineRiverChassis_RiverCabinet() {
	chassis, err := suite.generator.determineRiverChassis(xnames.Cabinet{Cabinet: 3000})
	suite.NoError(err)
	suite.Equal(
		xnames.FromString("x3000c0"),
		chassis,
	)
}

func (suite *CabinetHelpersTestSuite) TestDetermineRiverChassis_HillCabinet() {
	_, err := suite.generator.determineRiverChassis(xnames.Cabinet{Cabinet: 5000})
	suite.EqualError(
		err,
		"hill cabinet (non EX2500) x5000 cannot contain air-cooled hardware",
	)
}

func (suite *CabinetHelpersTestSuite) TestDetermineRiverChassis_EX2500Cabinet() {
	chassis, err := suite.generator.determineRiverChassis(xnames.Cabinet{Cabinet: 5004})
	suite.NoError(err)
	suite.Equal(
		xnames.FromString("x5004c4"),
		chassis,
	)
}

func (suite *CabinetHelpersTestSuite) TestDetermineRiverChassis_MountainCabinet() {
	_, err := suite.generator.determineRiverChassis(xnames.Cabinet{Cabinet: 1000})
	suite.EqualError(
		err,
		"mountain cabinet x1000 cannot contain air-cooled hardware",
	)
}

func (suite *CabinetHelpersTestSuite) TestDetermineRiverChassis_InvalidCabinet() {
	_, err := suite.generator.determineRiverChassis(xnames.Cabinet{Cabinet: 1234})
	suite.Error(err)
}

func (suite *CabinetHelpersTestSuite) TestCanCabinetContainAirCooledHardware_RiverCabinet() {
	ok, err := suite.generator.canCabinetContainAirCooledHardware("x3000")
	suite.NoError(err)
	suite.True(ok)
}

func (suite *CabinetHelpersTestSuite) TestCanCabinetContainAirCooledHardware_MountainCabinet() {
	ok, err := suite.generator.canCabinetContainAirCooledHardware("x1000")
	suite.EqualError(
		err,
		"mountain cabinet x1000 cannot contain air-cooled hardware",
	)
	suite.False(ok)
}

func (suite *CabinetHelpersTestSuite) TestCanCabinetContainAirCooledHardware_HillCabinet() {
	ok, err := suite.generator.canCabinetContainAirCooledHardware("x5000")
	suite.EqualError(
		err,
		"hill cabinet (non EX2500) x5000 cannot contain air-cooled hardware",
	)
	suite.False(ok)
}

func (suite *CabinetHelpersTestSuite) TestCanCabinetContainAirCooledHardware_EX2500_NoAirCooledChassis() {
	ok, err := suite.generator.canCabinetContainAirCooledHardware("x5001")
	suite.EqualError(
		err,
		"hill cabinet (EX2500) x5001 does not contain any air-cooled chassis",
	)
	suite.False(ok)
}

func (suite *CabinetHelpersTestSuite) TestCanCabinetContainAirCooledHardware_EX2500_AirCooledChassis() {
	ok, err := suite.generator.canCabinetContainAirCooledHardware("x5004")
	suite.NoError(err)
	suite.True(ok)
}

func (suite *CabinetHelpersTestSuite) TestCanCabinetContainAirCooledHardware_UnknownCabinet() {
	ok, err := suite.generator.canCabinetContainAirCooledHardware("x1234")
	suite.Error(err)
	suite.False(ok)
}

func (suite *CabinetHelpersTestSuite) TestGetSortedCabinetXNames() {
	cabinetXnames := []xnames.Cabinet{
		{Cabinet: 3000},
		{Cabinet: 9000},
		{Cabinet: 5001},
		{Cabinet: 0},
		{Cabinet: 100},
		{Cabinet: 110},
		{Cabinet: 111},
		{Cabinet: 10},
	}

	// Build up the list of cabinets from the list of xnames. We only care about the xname of the cabinet
	// when sorting.
	cabinets := map[string]CabinetTemplate{}
	for _, xname := range cabinetXnames {
		cab := CabinetTemplate{
			Xname: xname,
		}

		cabinets[xname.String()] = cab
	}

	sortedCabinets := suite.generator.getSortedCabinetXNames(cabinets)

	expected := []string{
		"x0",
		"x10",
		"x100",
		"x110",
		"x111",
		"x3000",
		"x5001",
		"x9000",
	}

	suite.Equal(
		expected,
		sortedCabinets,
	)
}

func TestCabinetHelpersTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(CabinetHelpersTestSuite),
	)
}
