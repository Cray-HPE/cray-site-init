//
//  MIT License
//
//  (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.

//go:build !integration && !shcd
// +build !integration,!shcd

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
		ManagementSwitchTypeLeaf,
		ManagementSwitchTypeCDU,
		ManagementSwitchTypeLeafBMC,
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
		expectedError: errors.New("invalid xname used for Spine/Leaf/Edge switch: x10c0w14, should use xXcChHsS format"),
	}, {
		// Spine using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "d10w14",
			SwitchType: ManagementSwitchTypeSpine,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Leaf/Edge switch: d10w14, should use xXcChHsS format"),
	}, {
		// Leaf using MgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "x20c0w14",
			SwitchType: ManagementSwitchTypeLeaf,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Leaf/Edge switch: x20c0w14, should use xXcChHsS format"),
	}, {
		// Leaf using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: ManagementSwitch{
			Xname:      "d20w14",
			SwitchType: ManagementSwitchTypeLeaf,
			Brand:      ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("invalid xname used for Spine/Leaf/Edge switch: d20w14, should use xXcChHsS format"),
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
