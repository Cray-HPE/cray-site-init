/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

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

package cmd

import (
	"errors"
	"testing"

	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	"github.com/stretchr/testify/suite"
)

type InitCmdTestSuite struct {
	suite.Suite
}

func (suite *InitCmdTestSuite) TestValidateSwitchInput_HappyPath() {
	switches := []*csi.ManagementSwitch{
		{
			// LeafBMC
			Xname: "x3000c0w14", SwitchType: csi.ManagementSwitchTypeLeafBMC,
			Brand: csi.ManagementSwitchBrandAruba,
		}, {
			// Spine Switch
			Xname:      "x3000c0h13s1",
			SwitchType: csi.ManagementSwitchTypeSpine,
			Brand:      csi.ManagementSwitchBrandAruba,
		}, {
			// Leaf Switch
			Xname:      "x3000c0h12s1",
			SwitchType: csi.ManagementSwitchTypeLeaf,
			Brand:      csi.ManagementSwitchBrandAruba,
		}, {
			// CDU Switch located in a CDU
			Xname:      "d10w10",
			SwitchType: csi.ManagementSwitchTypeCDU,
			Brand:      csi.ManagementSwitchBrandDell,
		}, {
			// CDU Switch located the River cabinet adjacent to a Hill cabinet
			Xname:      "x3000c0h10s1",
			SwitchType: csi.ManagementSwitchTypeCDU,
			Brand:      csi.ManagementSwitchBrandDell,
		}, {
			// Edge Switch
			Xname:      "x3000c0h18s1",
			SwitchType: csi.ManagementSwitchTypeEdge,
			Brand:      csi.ManagementSwitchBrandArista,
		},
	}

	err := validateSwitchInput(switches)
	suite.NoError(err)
}

func (suite *InitCmdTestSuite) TestValidateSwitchInput_InvalidXname() {
	switches := []*csi.ManagementSwitch{
		{ // Valid Xname
			Xname: "x3000c0w14", SwitchType: csi.ManagementSwitchTypeLeafBMC,
			Brand: csi.ManagementSwitchBrandAruba,
		}, { // Invalid Xname
			Xname:      "x3000c0w15L",
			SwitchType: csi.ManagementSwitchTypeSpine,
			Brand:      csi.ManagementSwitchBrandAruba,
		},
	}

	err := validateSwitchInput(switches)
	suite.Equal(errors.New("switch_metadata.csv contains invalid switch data"), err)
}

func (suite *InitCmdTestSuite) TestValidateSwitchInput_WrongXNameTypes() {
	// Test validate with valid xnames, but check that we are enforcing that the
	// different switch types are using hte correct names.
	//
	// The validateSwitchInput function reports the same error if 1 or more switches
	// had validation issues.

	tests := []struct {
		mySwitch      csi.ManagementSwitch
		expectedError error
	}{{
		// Spine using MgmtSwitch, should be using MgmtHLSwitch
		mySwitch: csi.ManagementSwitch{
			Xname: "x10c0w14", SwitchType: csi.ManagementSwitchTypeSpine,
			Brand: csi.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}, {
		// Spine using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: csi.ManagementSwitch{
			Xname:      "d10w14",
			SwitchType: csi.ManagementSwitchTypeSpine,
			Brand:      csi.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}, {
		// Leaf using MgmtSwitch, should be using MgmtHLSwitch
		mySwitch: csi.ManagementSwitch{
			Xname:      "x20c0w14",
			SwitchType: csi.ManagementSwitchTypeLeaf,
			Brand:      csi.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}, {
		// Leaf using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: csi.ManagementSwitch{
			Xname:      "d20w14",
			SwitchType: csi.ManagementSwitchTypeLeaf,
			Brand:      csi.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}, {
		// CDU using MgmtHLSwitch, should be using CDUMgmtSwitch
		mySwitch: csi.ManagementSwitch{
			Xname:      "x30c0w14",
			SwitchType: csi.ManagementSwitchTypeCDU,
			Brand:      csi.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}}

	for _, test := range tests {
		switches := []*csi.ManagementSwitch{&test.mySwitch}
		err := validateSwitchInput(switches)
		suite.Equal(test.expectedError, err)
	}
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_HappyPath() {
	ncns := []*csi.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "x3000c0s2b0n0", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s3b0n0", Role: "Management", Subrole: "Storage"},
	}

	err := validateNCNInput(ncns)
	suite.NoError(err)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_InvalidXName() {
	ncns := []*csi.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "foo", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s3b0n0", Role: "Management", Subrole: "Storage"},
	}

	err := validateNCNInput(ncns)
	suite.Equal(errors.New("ncn_metadata.csv contains invalid NCN data"), err)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_WrongXNameType() {
	ncns := []*csi.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "x3000c0s2b0n0", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s3b0", Role: "Management", Subrole: "Storage"},
	}

	err := validateNCNInput(ncns)
	suite.Equal(errors.New("ncn_metadata.csv contains invalid NCN data"), err)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_ZeroNCNs() {
	ncns := []*csi.LogicalNCN{}

	err := validateNCNInput(ncns)
	suite.Equal(errors.New("unable to extract NCNs from ncn metadata csv"), err)
}

func (suite *InitCmdTestSuite) TestMergeNCNs_HappyPath() {
	ncns := []*csi.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "x3000c0s4b0n0", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s7b0n0", Role: "Management", Subrole: "Storage"},
	}

	slsNCNs := []csi.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Hostname: "ncn-m001", Aliases: []string{"ncn-m001"}, BmcPort: "x3000c0w14:1/1/31"},
		{Xname: "x3000c0s4b0n0", Hostname: "ncn-w001", Aliases: []string{"ncn-w001"}, BmcPort: "x3000c0w14:1/1/32"},
		{Xname: "x3000c0s7b0n0", Hostname: "ncn-s001", Aliases: []string{"ncn-s001"}, BmcPort: "x3000c0w14:1/1/33"},
	}

	expectedMergeList := []*csi.LogicalNCN{
		{
			Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master",
			Hostname: "ncn-m001", Aliases: []string{"ncn-m001"}, BmcPort: "x3000c0w14:1/1/31",
		}, {
			Xname: "x3000c0s4b0n0", Role: "Management", Subrole: "Worker",
			Hostname: "ncn-w001", Aliases: []string{"ncn-w001"}, BmcPort: "x3000c0w14:1/1/32",
		}, {
			Xname: "x3000c0s7b0n0", Role: "Management", Subrole: "Storage",
			Hostname: "ncn-s001", Aliases: []string{"ncn-s001"}, BmcPort: "x3000c0w14:1/1/33",
		},
	}

	err := mergeNCNs(ncns, slsNCNs)
	suite.NoError(err)

	suite.Equal(expectedMergeList, ncns)
}

func (suite *InitCmdTestSuite) TestMergeNCNs_MissingXnameInSLS() {
	ncns := []*csi.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "x3000c0s4b0n0", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s7b0n0", Role: "Management", Subrole: "Storage"},
	}

	slsNCNs := []csi.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Hostname: "ncn-m001", Aliases: []string{"ncn-m001"}, BmcPort: "x3000c0w14:1/1/31"},
		{Xname: "x3000c0s4b0n0", Hostname: "ncn-w001", Aliases: []string{"ncn-w001"}, BmcPort: "x3000c0w14:1/1/32"},
	}

	err := mergeNCNs(ncns, slsNCNs)
	suite.Equal(errors.New("failed to find NCN from ncn-metadata in generated SLS State: x3000c0s7b0n0"), err)
}

func TestInitCmdTestSuite(t *testing.T) {
	suite.Run(t, new(InitCmdTestSuite))
}
