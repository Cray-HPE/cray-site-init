// +build !integration
// +build !shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type NetworkingTestSuite struct {
	suite.Suite
}

func (suite *NetworkingTestSuite) TestIsManagementSwitchTypeValid() {
	switchTypes := []ManagementSwitchType{
		ManagementSwitchTypeAggregation,
		ManagementSwitchTypeCDU,
		ManagementSwitchTypeLeaf,
		ManagementSwitchTypeSpine,
	}

	for _, switchType := range switchTypes {
		valid := IsManagementSwitchTypeValid(switchType)
		suite.True(valid, "Switch type: %s", switchType)
	}
}

func (suite *NetworkingTestSuite) TestIsManagementSwitchTypeValid_InvalidType() {
	switchType := ManagementSwitchType("foo")

	valid := IsManagementSwitchTypeValid(switchType)
	suite.False(valid, "Switch type: %s", switchType)
}

func (suite *NetworkingTestSuite) TestValidateSwitch_InvalidXname() {
	mySwitch := ManagementSwitch{
		Xname:      "x3000c0w15L", // Invalid Xname
		SwitchType: ManagementSwitchTypeSpine,
		Brand:      ManagementSwitchBrandAruba,
	}

	err := mySwitch.Validate()
	suite.Equal(errors.New("invalid xname for Switch: x3000c0w15L"), err)
}

func (suite *NetworkingTestSuite) TestValidateSwitch_WrongXnameTypes() {
	// Test validate with valid xnames, but check that we are enforcing that the
	// different switch types are using hte correct names

	tests := []struct {
		mySwitch      ManagementSwitch
		expectedError error
	}{{
		// Spine using MgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "x10c0w14",
			SwitchType: ManagementSwitchTypeSpine,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Aggregation switch: x10c0w14, should use xXcChHsS format"),
	}, {
		// Spine using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "d10w14",
			SwitchType: ManagementSwitchTypeSpine,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Aggregation switch: d10w14, should use xXcChHsS format"),
	}, {
		// Aggregation using MgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "x20c0w14",
			SwitchType: ManagementSwitchTypeAggregation,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Aggregation switch: x20c0w14, should use xXcChHsS format"),
	}, {
		// Aggregation using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "d20w14",
			SwitchType: ManagementSwitchTypeAggregation,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Aggregation switch: d20w14, should use xXcChHsS format"),
	}, {
		// CDU using MgmtHLSwitch, should be using CDUMgmtSwitch
		mySwitch: ManagementSwitch{
			Xname:      "x30c0w14",
			SwitchType: ManagementSwitchTypeCDU,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for CDU switch: x30c0w14, should use dDwW format (if in an adjacent river cabinet to a TBD cabinet use the xXcChHsS format)"),
	}}

	for _, test := range tests {
		err := test.mySwitch.Validate()
		suite.Equal(test.expectedError, err)
	}
}

func (suite *NetworkingTestSuite) TestNormalizeSwitch() {

	tests := []struct {
		mySwitch      ManagementSwitch
		expectedXname string
		expectedError error
	}{{
		// Already normalized MgmtSwitch Xname
		mySwitch: ManagementSwitch{
			Xname: "x3000c0w10",
		},
		expectedXname: "x3000c0w10",
		expectedError: nil,
	}, {
		// Already normalized MgmtHLSwitch Xname
		mySwitch: ManagementSwitch{
			Xname: "x3000c0h10s1",
		},
		expectedXname: "x3000c0h10s1",
		expectedError: nil,
	}, {
		// Already normalized CDUMgmtSwitch Xname
		mySwitch: ManagementSwitch{
			Xname: "d10w14",
		},
		expectedXname: "d10w14",
		expectedError: nil,
	}, {
		// Un-normalized MgmtSwitch Xname
		mySwitch: ManagementSwitch{
			Xname: "x03000c00w010",
		},
		expectedXname: "x3000c0w10",
		expectedError: nil,
	}, {
		// Un-normalized MgmtHLSwitch Xname
		mySwitch: ManagementSwitch{
			Xname: "x03000c00h010s01",
		},
		expectedXname: "x3000c0h10s1",
		expectedError: nil,
	}, {
		// Un-normalized CDUMgmtSwitch Xname
		mySwitch: ManagementSwitch{
			Xname: "d010w014",
		},
		expectedXname: "d10w14",
		expectedError: nil,
	}}

	for _, test := range tests {
		err := test.mySwitch.Normalize()
		suite.Equal(test.expectedError, err)
		suite.Equal(test.expectedXname, test.mySwitch.Xname)
	}
}

func TestNetworkingTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkingTestSuite))
}
