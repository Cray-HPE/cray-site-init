/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

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
			Xname: "x10c0w14", SwitchType: ManagementSwitchTypeSpine,
			Brand: ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Aggergation switch: x10c0w14, should use xXcChHsS format"),
	}, {
		// Spine using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "d10w14",
			SwitchType: ManagementSwitchTypeSpine,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Aggergation switch: d10w14, should use xXcChHsS format"),
	}, {
		// Aggergation using MgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "x20c0w14",
			SwitchType: ManagementSwitchTypeAggregation,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Aggergation switch: x20c0w14, should use xXcChHsS format"),
	}, {
		// Aggergation using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "d20w14",
			SwitchType: ManagementSwitchTypeAggregation,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Aggergation switch: d20w14, should use xXcChHsS format"),
	}, {
		// CDU using MgmtHLSwitch, should be using CDUMgmtSwitch
		mySwitch: ManagementSwitch{
			Xname:      "x30c0w14",
			SwitchType: ManagementSwitchTypeCDU,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for CDU switch: x30c0w14, should use dDwW format"),
	}, {
		// CDU using MgmtSwitch, should be using CDUMgmtSwitch
		mySwitch: ManagementSwitch{
			Xname:      "x30c0h14s1",
			SwitchType: ManagementSwitchTypeCDU,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for CDU switch: x30c0h14s1, should use dDwW format"),
	}}

	for _, test := range tests {
		err := test.mySwitch.Validate()
		suite.Equal(test.expectedError, err)
	}
}

func TestNetworkingTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkingTestSuite))
}
