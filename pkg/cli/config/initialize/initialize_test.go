/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

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

package initialize

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

type InitCmdTestSuite struct {
	suite.Suite
}

func (suite *InitCmdTestSuite) TestValidateSwitchInput_HappyPath() {
	switches := []*networking.ManagementSwitch{
		{
			// LeafBMC
			Xname:      "x3000c0w14",
			SwitchType: networking.ManagementSwitchTypeLeafBMC,
			Brand:      networking.ManagementSwitchBrandAruba,
		},
		{
			// Spine Switch
			Xname:      "x3000c0h13s1",
			SwitchType: networking.ManagementSwitchTypeSpine,
			Brand:      networking.ManagementSwitchBrandAruba,
		},
		{
			// Leaf Switch
			Xname:      "x3000c0h12s1",
			SwitchType: networking.ManagementSwitchTypeLeaf,
			Brand:      networking.ManagementSwitchBrandAruba,
		},
		{
			// CDU Switch located in a CDU
			Xname:      "d10w10",
			SwitchType: networking.ManagementSwitchTypeCDU,
			Brand:      networking.ManagementSwitchBrandDell,
		},
		{
			// CDU Switch located the River cabinet adjacent to a Hill cabinet
			Xname:      "x3000c0h10s1",
			SwitchType: networking.ManagementSwitchTypeCDU,
			Brand:      networking.ManagementSwitchBrandDell,
		},
		{
			// Edge Switch
			Xname:      "x3000c0h18s1",
			SwitchType: networking.ManagementSwitchTypeEdge,
			Brand:      networking.ManagementSwitchBrandArista,
		},
	}

	err := validateSwitchInput(switches)
	suite.NoError(err)
}

func (suite *InitCmdTestSuite) TestValidateSwitchInput_InvalidXname() {
	switches := []*networking.ManagementSwitch{
		{
			// Valid Xname
			Xname:      "x3000c0w14",
			SwitchType: networking.ManagementSwitchTypeLeafBMC,
			Brand:      networking.ManagementSwitchBrandAruba,
		},
		{
			// Invalid Xname
			Xname:      "x3000c0w15L",
			SwitchType: networking.ManagementSwitchTypeSpine,
			Brand:      networking.ManagementSwitchBrandAruba,
		},
	}

	err := validateSwitchInput(switches)
	suite.Equal(
		errors.New("switch_metadata.csv contains invalid switch Data"),
		err,
	)
}

func (suite *InitCmdTestSuite) TestValidateSwitchInput_WrongXNameTypes() {
	// Test validate with valid xnames, but check that we are enforcing that the
	// different switch types are using hte correct names.
	//
	// The validateSwitchInput function reports the same error if 1 or more switches
	// had validation issues.

	tests := []struct {
		mySwitch      networking.ManagementSwitch
		expectedError error
	}{
		{
			// Spine using MgmtSwitch, should be using MgmtHLSwitch
			mySwitch: networking.ManagementSwitch{
				Xname:      "x10c0w14",
				SwitchType: networking.ManagementSwitchTypeSpine,
				Brand:      networking.ManagementSwitchBrandAruba,
			},
			expectedError: errors.New("switch_metadata.csv contains invalid switch Data"),
		},
		{
			// Spine using CDUMgmtSwitch, should be using MgmtHLSwitch
			mySwitch: networking.ManagementSwitch{
				Xname:      "d10w14",
				SwitchType: networking.ManagementSwitchTypeSpine,
				Brand:      networking.ManagementSwitchBrandAruba,
			},
			expectedError: errors.New("switch_metadata.csv contains invalid switch Data"),
		},
		{
			// Leaf using MgmtSwitch, should be using MgmtHLSwitch
			mySwitch: networking.ManagementSwitch{
				Xname:      "x20c0w14",
				SwitchType: networking.ManagementSwitchTypeLeaf,
				Brand:      networking.ManagementSwitchBrandAruba,
			},
			expectedError: errors.New("switch_metadata.csv contains invalid switch Data"),
		},
		{
			// Leaf using CDUMgmtSwitch, should be using MgmtHLSwitch
			mySwitch: networking.ManagementSwitch{
				Xname:      "d20w14",
				SwitchType: networking.ManagementSwitchTypeLeaf,
				Brand:      networking.ManagementSwitchBrandAruba,
			},
			expectedError: errors.New("switch_metadata.csv contains invalid switch Data"),
		},
		{
			// CDU using MgmtHLSwitch, should be using CDUMgmtSwitch
			mySwitch: networking.ManagementSwitch{
				Xname:      "x30c0w14",
				SwitchType: networking.ManagementSwitchTypeCDU,
				Brand:      networking.ManagementSwitchBrandAruba,
			},
			expectedError: errors.New("switch_metadata.csv contains invalid switch Data"),
		},
	}

	for _, test := range tests {
		switches := []*networking.ManagementSwitch{&test.mySwitch}
		err := validateSwitchInput(switches)
		suite.Equal(
			test.expectedError,
			err,
		)
	}
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_HappyPath() {
	ncns := []*LogicalNCN{
		{
			Xname:   "x3000c0s1b0n0",
			Role:    "Management",
			Subrole: "Master",
		},
		{
			Xname:   "x3000c0s2b0n0",
			Role:    "Management",
			Subrole: "Worker",
		},
		{
			Xname:   "x3000c0s3b0n0",
			Role:    "Management",
			Subrole: "Storage",
		},
	}

	err := validateNCNInput(ncns)
	suite.NoError(err)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_InvalidXName() {
	ncns := []*LogicalNCN{
		{
			Xname:   "x3000c0s1b0n0",
			Role:    "Management",
			Subrole: "Master",
		},
		{
			Xname:   "foo",
			Role:    "Management",
			Subrole: "Worker",
		},
		{
			Xname:   "x3000c0s3b0n0",
			Role:    "Management",
			Subrole: "Storage",
		},
	}

	err := validateNCNInput(ncns)
	suite.Equal(
		errors.New("ncn_metadata.csv contains invalid NCN Data"),
		err,
	)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_WrongXNameType() {
	ncns := []*LogicalNCN{
		{
			Xname:   "x3000c0s1b0n0",
			Role:    "Management",
			Subrole: "Master",
		},
		{
			Xname:   "x3000c0s2b0n0",
			Role:    "Management",
			Subrole: "Worker",
		},
		{
			Xname:   "x3000c0s3b0",
			Role:    "Management",
			Subrole: "Storage",
		},
	}

	err := validateNCNInput(ncns)
	suite.Equal(
		errors.New("ncn_metadata.csv contains invalid NCN Data"),
		err,
	)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_ZeroNCNs() {
	ncns := []*LogicalNCN{}

	err := validateNCNInput(ncns)
	suite.Equal(
		errors.New("unable to extract NCNs from ncn metadata csv"),
		err,
	)
}

func (suite *InitCmdTestSuite) TestMergeNCNs_HappyPath() {
	ncns := []*LogicalNCN{
		{
			Xname:   "x3000c0s1b0n0",
			Role:    "Management",
			Subrole: "Master",
		},
		{
			Xname:   "x3000c0s4b0n0",
			Role:    "Management",
			Subrole: "Worker",
		},
		{
			Xname:   "x3000c0s7b0n0",
			Role:    "Management",
			Subrole: "Storage",
		},
	}

	slsNCNs := []LogicalNCN{
		{
			Xname:    "x3000c0s1b0n0",
			Hostname: "ncn-m001",
			Aliases:  []string{"ncn-m001"},
			BmcPort:  "x3000c0w14:1/1/31",
		},
		{
			Xname:    "x3000c0s4b0n0",
			Hostname: "ncn-w001",
			Aliases:  []string{"ncn-w001"},
			BmcPort:  "x3000c0w14:1/1/32",
		},
		{
			Xname:    "x3000c0s7b0n0",
			Hostname: "ncn-s001",
			Aliases:  []string{"ncn-s001"},
			BmcPort:  "x3000c0w14:1/1/33",
		},
	}

	expectedMergeList := []*LogicalNCN{
		{
			Xname:    "x3000c0s1b0n0",
			Role:     "Management",
			Subrole:  "Master",
			Hostname: "ncn-m001",
			Aliases:  []string{"ncn-m001"},
			BmcPort:  "x3000c0w14:1/1/31",
		},
		{
			Xname:    "x3000c0s4b0n0",
			Role:     "Management",
			Subrole:  "Worker",
			Hostname: "ncn-w001",
			Aliases:  []string{"ncn-w001"},
			BmcPort:  "x3000c0w14:1/1/32",
		},
		{
			Xname:    "x3000c0s7b0n0",
			Role:     "Management",
			Subrole:  "Storage",
			Hostname: "ncn-s001",
			Aliases:  []string{"ncn-s001"},
			BmcPort:  "x3000c0w14:1/1/33",
		},
	}

	mergedNCNs, err := mergeNCNs(
		ncns,
		slsNCNs,
	)
	suite.NoError(err)

	suite.Equal(
		expectedMergeList,
		mergedNCNs,
	)
}

func (suite *InitCmdTestSuite) TestMergeNCNs_MissingXnameInSLS() {
	ncns := []*LogicalNCN{
		{
			Xname:   "x3000c0s1b0n0",
			Role:    "Management",
			Subrole: "Master",
		},
		{
			Xname:   "x3000c0s4b0n0",
			Role:    "Management",
			Subrole: "Worker",
		},
		{
			Xname:   "x3000c0s7b0n0",
			Role:    "Management",
			Subrole: "Storage",
		},
	}

	slsNCNs := []LogicalNCN{
		{
			Xname:    "x3000c0s1b0n0",
			Hostname: "ncn-m001",
			Aliases:  []string{"ncn-m001"},
			BmcPort:  "x3000c0w14:1/1/31",
		},
		{
			Xname:    "x3000c0s4b0n0",
			Hostname: "ncn-w001",
			Aliases:  []string{"ncn-w001"},
			BmcPort:  "x3000c0w14:1/1/32",
		},
	}

	_, err := mergeNCNs(
		ncns,
		slsNCNs,
	)
	suite.Equal(
		errors.New("failed to find NCN from ncn-metadata in generated SLS State: x3000c0s7b0n0"),
		err,
	)
}

func TestInitCmdTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(InitCmdTestSuite),
	)
}
