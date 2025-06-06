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
	"errors"
	"net"
	"testing"

	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"

	"github.com/stretchr/testify/suite"

	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

type GenSLSTestSuite struct {
	suite.Suite
}

func (suite *GenSLSTestSuite) TestConvertManagementSwitchToSLS_HappyPath() {
	tests := []struct {
		csiSwitch   networking.ManagementSwitch
		expectedSLS slsCommon.GenericHardware
	}{
		{
			// Leaf Switch
			csiSwitch: networking.ManagementSwitch{
				Xname:               "x3000c0w14",
				Name:                "sw-leaf-bmc-001",
				ManagementInterface: net.ParseIP("10.254.0.2"),
				SwitchType:          networking.ManagementSwitchTypeLeafBMC,
				Brand:               networking.ManagementSwitchBrandAruba,
				Model:               "6300M",
			},
			expectedSLS: slsCommon.GenericHardware{
				Parent:     "x3000c0",
				Xname:      "x3000c0w14",
				Type:       "comptype_mgmt_switch",
				TypeString: "MgmtSwitch",
				Class:      "River",
				ExtraPropertiesRaw: slsCommon.ComptypeMgmtSwitch{
					IP4Addr:          "10.254.0.2",
					Brand:            "Aruba",
					Model:            "6300M",
					SNMPAuthPassword: "vault://hms-creds/x3000c0w14",
					SNMPAuthProtocol: "MD5",
					SNMPPrivPassword: "vault://hms-creds/x3000c0w14",
					SNMPPrivProtocol: "DES",
					SNMPUsername:     "testuser",
					Aliases:          []string{"sw-leaf-bmc-001"},
				},
			},
		},
		{
			// Spine Switch
			csiSwitch: networking.ManagementSwitch{
				Xname:               "x3000c0h13s1",
				Name:                "sw-spine-001",
				ManagementInterface: net.ParseIP("10.254.0.2"),
				SwitchType:          networking.ManagementSwitchTypeSpine,
				Brand:               networking.ManagementSwitchBrandAruba,
				Model:               "8325",
			},
			expectedSLS: slsCommon.GenericHardware{
				Parent:     "x3000c0h13",
				Xname:      "x3000c0h13s1",
				Type:       "comptype_hl_switch",
				TypeString: "MgmtHLSwitch",
				Class:      "River",
				ExtraPropertiesRaw: slsCommon.ComptypeMgmtHLSwitch{
					IP4Addr: "10.254.0.2",
					Brand:   "Aruba",
					Model:   "8325",
					Aliases: []string{"sw-spine-001"},
				},
			},
		},
		{
			// Leaf Switch
			csiSwitch: networking.ManagementSwitch{
				Xname:               "x3000c0h13s1",
				Name:                "sw-leaf-001",
				ManagementInterface: net.ParseIP("10.254.0.2"),
				SwitchType:          networking.ManagementSwitchTypeLeaf,
				Brand:               networking.ManagementSwitchBrandAruba,
				Model:               "8325",
			},
			expectedSLS: slsCommon.GenericHardware{
				Parent:     "x3000c0h13",
				Xname:      "x3000c0h13s1",
				Type:       "comptype_hl_switch",
				TypeString: "MgmtHLSwitch",
				Class:      "River",
				ExtraPropertiesRaw: slsCommon.ComptypeMgmtHLSwitch{
					IP4Addr: "10.254.0.2",
					Brand:   "Aruba",
					Model:   "8325",
					Aliases: []string{"sw-leaf-001"},
				},
			},
		},
		{
			// CDU Mgmt Switch
			csiSwitch: networking.ManagementSwitch{
				Xname:               "d10w10",
				Name:                "sw-cdu-001",
				ManagementInterface: net.ParseIP("10.254.0.2"),
				SwitchType:          networking.ManagementSwitchTypeCDU,
				Brand:               networking.ManagementSwitchBrandDell,
				Model:               "8325",
			},
			expectedSLS: slsCommon.GenericHardware{
				Parent:     "d10",
				Xname:      "d10w10",
				Type:       "comptype_cdu_mgmt_switch",
				TypeString: "CDUMgmtSwitch",
				Class:      "Mountain",
				ExtraPropertiesRaw: slsCommon.ComptypeCDUMgmtSwitch{
					Brand:   "Dell",
					Model:   "8325",
					Aliases: []string{"sw-cdu-001"},
				},
			},
		},
		{
			// CDU Mgmt Switch in a River cabinet for a adjacent Hill cabinet.
			csiSwitch: networking.ManagementSwitch{
				Xname:               "x3000c0h10s1",
				Name:                "sw-cdu-002",
				ManagementInterface: net.ParseIP("10.254.0.2"),
				SwitchType:          networking.ManagementSwitchTypeCDU,
				Brand:               networking.ManagementSwitchBrandDell,
				Model:               "8325",
			},
			expectedSLS: slsCommon.GenericHardware{
				Parent:     "x3000c0h10",
				Xname:      "x3000c0h10s1",
				Type:       "comptype_hl_switch",
				TypeString: "MgmtHLSwitch",
				Class:      "River",
				ExtraPropertiesRaw: slsCommon.ComptypeMgmtHLSwitch{
					IP4Addr: "10.254.0.2",
					Brand:   "Dell",
					Model:   "8325",
					Aliases: []string{"sw-cdu-002"},
				},
			},
		},
	}

	for _, test := range tests {
		slsSwitch, err := ConvertManagementSwitchToSLS(&test.csiSwitch)
		suite.NoError(err)
		suite.Equal(
			test.expectedSLS,
			slsSwitch,
		)
	}
}

func (suite *GenSLSTestSuite) TestConvertManagementSwitchToSLS_InvalidSwitchType() {
	tests := []struct {
		csiSwitch     networking.ManagementSwitch
		expectedError error
	}{
		{
			// Missing Switch Type
			csiSwitch: networking.ManagementSwitch{
				Xname:               "x3000c0w14",
				Name:                "sw-leaf-bmc-001",
				ManagementInterface: net.ParseIP("10.254.0.2"),
				Brand:               networking.ManagementSwitchBrandAruba,
				Model:               "6300M",
			},
			expectedError: errors.New("unknown management switch type: "),
		},
		{
			// Invalid switch type
			csiSwitch: networking.ManagementSwitch{
				Xname:               "x3000c0w14",
				Name:                "sw-leaf-bmc-001",
				ManagementInterface: net.ParseIP("10.254.0.2"),
				SwitchType:          networking.ManagementSwitchType("foobar"),
				Brand:               networking.ManagementSwitchBrandAruba,
				Model:               "6300M",
			},
			expectedError: errors.New("unknown management switch type: foobar"),
		},
	}

	for _, test := range tests {
		_, err := ConvertManagementSwitchToSLS(&test.csiSwitch)
		suite.Equal(
			test.expectedError,
			err,
		)
	}
}

func (suite *GenSLSTestSuite) TestExtractSwitchesfromReservations() {
	subnet := &networking.IPV4Subnet{
		IPReservations: []networking.IPReservation{
			{
				Comment:   "x3000c0w14",
				Name:      "sw-leaf-bmc-001",
				IPAddress: net.ParseIP("10.254.0.2"),
			},
			{
				Comment:   "x3000c0h13s1",
				Name:      "sw-spine-001",
				IPAddress: net.ParseIP("10.254.0.3"),
			},
			{
				Comment:   "x3000c0h12s1",
				Name:      "sw-leaf-001",
				IPAddress: net.ParseIP("10.254.0.4"),
			},
			{
				Comment:   "d10w10",
				Name:      "sw-cdu-001",
				IPAddress: net.ParseIP("10.254.0.5"),
			},
		},
	}

	expectedOutput := []networking.ManagementSwitch{
		{
			Xname:               "x3000c0w14",
			Name:                "sw-leaf-bmc-001",
			SwitchType:          "LeafBMC",
			ManagementInterface: net.ParseIP("10.254.0.2"),
		},
		{
			Xname:               "x3000c0h13s1",
			Name:                "sw-spine-001",
			SwitchType:          "Spine",
			ManagementInterface: net.ParseIP("10.254.0.3"),
		},
		{
			Xname:               "x3000c0h12s1",
			Name:                "sw-leaf-001",
			SwitchType:          "Leaf",
			ManagementInterface: net.ParseIP("10.254.0.4"),
		},
		{
			Xname:               "d10w10",
			Name:                "sw-cdu-001",
			SwitchType:          "CDU",
			ManagementInterface: net.ParseIP("10.254.0.5"),
		},
	}

	extractedSwitches, err := ExtractSwitchesfromReservations(subnet)
	suite.NoError(err)
	suite.Equal(
		expectedOutput,
		extractedSwitches,
	)
}

func TestGenSLSTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(GenSLSTestSuite),
	)
}
