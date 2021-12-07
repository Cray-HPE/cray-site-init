// +build !integration
// +build !shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
	"net"
	"testing"

	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/stretchr/testify/suite"
)

type GenSLSTestSuite struct {
	suite.Suite
}

func (suite *GenSLSTestSuite) TestConvertManagementSwitchToSLS_HappyPath() {
	tests := []struct {
		csiSwitch   csi.ManagementSwitch
		expectedSLS sls_common.GenericHardware
	}{{
		// Leaf Switch
		csiSwitch: csi.ManagementSwitch{
			Xname:               "x3000c0w14",
			Name:                "sw-leaf-bmc-001",
			ManagementInterface: net.ParseIP("10.254.0.2"),
			SwitchType:          csi.ManagementSwitchTypeLeafBMC,
			Brand:               csi.ManagementSwitchBrandAruba,
			Model:               "6300M",
		},
		expectedSLS: sls_common.GenericHardware{
			Parent:     "x3000c0",
			Xname:      "x3000c0w14",
			Type:       "comptype_mgmt_switch",
			TypeString: "MgmtSwitch",
			Class:      "River",
			ExtraPropertiesRaw: sls_common.ComptypeMgmtSwitch{
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
	}, {
		// Spine Switch
		csiSwitch: csi.ManagementSwitch{
			Xname:               "x3000c0h13s1",
			Name:                "sw-spine-001",
			ManagementInterface: net.ParseIP("10.254.0.2"),
			SwitchType:          csi.ManagementSwitchTypeSpine,
			Brand:               csi.ManagementSwitchBrandAruba,
			Model:               "8325",
		},
		expectedSLS: sls_common.GenericHardware{
			Parent:     "x3000c0h13",
			Xname:      "x3000c0h13s1",
			Type:       "comptype_hl_switch",
			TypeString: "MgmtHLSwitch",
			Class:      "River",
			ExtraPropertiesRaw: sls_common.ComptypeMgmtHLSwitch{
				IP4Addr: "10.254.0.2",
				Brand:   "Aruba",
				Model:   "8325",
				Aliases: []string{"sw-spine-001"},
			},
		},
	}, {
		// Leaf Switch
		csiSwitch: csi.ManagementSwitch{
			Xname:               "x3000c0h13s1",
			Name:                "sw-leaf-001",
			ManagementInterface: net.ParseIP("10.254.0.2"),
			SwitchType:          csi.ManagementSwitchTypeLeaf,
			Brand:               csi.ManagementSwitchBrandAruba,
			Model:               "8325",
		},
		expectedSLS: sls_common.GenericHardware{
			Parent:     "x3000c0h13",
			Xname:      "x3000c0h13s1",
			Type:       "comptype_hl_switch",
			TypeString: "MgmtHLSwitch",
			Class:      "River",
			ExtraPropertiesRaw: sls_common.ComptypeMgmtHLSwitch{
				IP4Addr: "10.254.0.2",
				Brand:   "Aruba",
				Model:   "8325",
				Aliases: []string{"sw-leaf-001"},
			},
		},
	}, {
		// CDU Mgmt Switch
		csiSwitch: csi.ManagementSwitch{
			Xname:               "d10w10",
			Name:                "sw-cdu-001",
			ManagementInterface: net.ParseIP("10.254.0.2"),
			SwitchType:          csi.ManagementSwitchTypeCDU,
			Brand:               csi.ManagementSwitchBrandDell,
			Model:               "8325",
		},
		expectedSLS: sls_common.GenericHardware{
			Parent:     "d10",
			Xname:      "d10w10",
			Type:       "comptype_cdu_mgmt_switch",
			TypeString: "CDUMgmtSwitch",
			Class:      "Mountain",
			ExtraPropertiesRaw: sls_common.ComptypeCDUMgmtSwitch{
				Brand:   "Dell",
				Model:   "8325",
				Aliases: []string{"sw-cdu-001"},
			},
		},
	}, {
		// CDU Mgmt Switch in a River cabinet for a adjacent Hill cabinet.
		csiSwitch: csi.ManagementSwitch{
			Xname:               "x3000c0h10s1",
			Name:                "sw-cdu-002",
			ManagementInterface: net.ParseIP("10.254.0.2"),
			SwitchType:          csi.ManagementSwitchTypeCDU,
			Brand:               csi.ManagementSwitchBrandDell,
			Model:               "8325",
		},
		expectedSLS: sls_common.GenericHardware{
			Parent:     "x3000c0h10",
			Xname:      "x3000c0h10s1",
			Type:       "comptype_hl_switch",
			TypeString: "MgmtHLSwitch",
			Class:      "River",
			ExtraPropertiesRaw: sls_common.ComptypeMgmtHLSwitch{
				IP4Addr: "10.254.0.2",
				Brand:   "Dell",
				Model:   "8325",
				Aliases: []string{"sw-cdu-002"},
			},
		},
	}}

	for _, test := range tests {
		slsSwitch, err := convertManagementSwitchToSLS(&test.csiSwitch)
		suite.NoError(err)
		suite.Equal(test.expectedSLS, slsSwitch)
	}
}

func (suite *GenSLSTestSuite) TestConvertManagementSwitchToSLS_InvalidSwitchType() {
	tests := []struct {
		csiSwitch     csi.ManagementSwitch
		expectedError error
	}{{
		// Missing Switch Type
		csiSwitch: csi.ManagementSwitch{
			Xname:               "x3000c0w14",
			Name:                "sw-leaf-bmc-001",
			ManagementInterface: net.ParseIP("10.254.0.2"),
			Brand:               csi.ManagementSwitchBrandAruba,
			Model:               "6300M",
		},
		expectedError: errors.New("unknown management switch type: "),
	}, {
		// Invalid switch type
		csiSwitch: csi.ManagementSwitch{
			Xname:               "x3000c0w14",
			Name:                "sw-leaf-bmc-001",
			ManagementInterface: net.ParseIP("10.254.0.2"),
			SwitchType:          csi.ManagementSwitchType("foobar"),
			Brand:               csi.ManagementSwitchBrandAruba,
			Model:               "6300M",
		},
		expectedError: errors.New("unknown management switch type: foobar"),
	}}

	for _, test := range tests {
		_, err := convertManagementSwitchToSLS(&test.csiSwitch)
		suite.Equal(test.expectedError, err)
	}
}

func (suite *GenSLSTestSuite) TestExtractSwitchesfromReservations() {
	subnet := &csi.IPV4Subnet{
		IPReservations: []csi.IPReservation{{
			Comment:   "x3000c0w14",
			Name:      "sw-leaf-bmc-001",
			IPAddress: net.ParseIP("10.254.0.2"),
		}, {
			Comment:   "x3000c0h13s1",
			Name:      "sw-spine-001",
			IPAddress: net.ParseIP("10.254.0.3"),
		}, {
			Comment:   "x3000c0h12s1",
			Name:      "sw-leaf-001",
			IPAddress: net.ParseIP("10.254.0.4"),
		}, {
			Comment:   "d10w10",
			Name:      "sw-cdu-001",
			IPAddress: net.ParseIP("10.254.0.5"),
		}},
	}

	expectedOutput := []csi.ManagementSwitch{{
		Xname:               "x3000c0w14",
		Name:                "sw-leaf-bmc-001",
		SwitchType:          "LeafBMC",
		ManagementInterface: net.ParseIP("10.254.0.2"),
	}, {
		Xname:               "x3000c0h13s1",
		Name:                "sw-spine-001",
		SwitchType:          "Spine",
		ManagementInterface: net.ParseIP("10.254.0.3"),
	}, {
		Xname:               "x3000c0h12s1",
		Name:                "sw-leaf-001",
		SwitchType:          "Leaf",
		ManagementInterface: net.ParseIP("10.254.0.4"),
	}, {
		Xname:               "d10w10",
		Name:                "sw-cdu-001",
		SwitchType:          "CDU",
		ManagementInterface: net.ParseIP("10.254.0.5"),
	}}

	extractedSwitches, err := extractSwitchesfromReservations(subnet)
	suite.NoError(err)
	suite.Equal(expectedOutput, extractedSwitches)
}

func TestGenSLSTestSuite(t *testing.T) {
	suite.Run(t, new(GenSLSTestSuite))
}
