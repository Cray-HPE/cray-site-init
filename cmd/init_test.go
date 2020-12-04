/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
)

type InitCmdTestSuite struct {
	suite.Suite
}

func (suite *InitCmdTestSuite) TestValidateSwitchInput_HappyPath() {
	switches := []*shasta.ManagementSwitch{
		{
			Xname: "x3000c0w14", SwitchType: shasta.ManagementSwitchTypeLeaf,
			Brand: shasta.ManagementSwitchBrandAruba,
		}, {
			Xname:      "x3000c0h13s1",
			SwitchType: shasta.ManagementSwitchTypeSpine,
			Brand:      shasta.ManagementSwitchBrandAruba,
		}, {
			Xname:      "x3000c0h12s1",
			SwitchType: shasta.ManagementSwitchTypeAggregation,
			Brand:      shasta.ManagementSwitchBrandAruba,
		}, {
			Xname:      "d10w10",
			SwitchType: shasta.ManagementSwitchTypeCDU,
			Brand:      shasta.ManagementSwitchBrandDell,
		},
	}

	err := validateSwitchInput(switches)
	suite.NoError(err)
}

func (suite *InitCmdTestSuite) TestValidateSwitchInput_InvalidXname() {
	switches := []*shasta.ManagementSwitch{
		{ // Valid Xname
			Xname: "x3000c0w14", SwitchType: shasta.ManagementSwitchTypeLeaf,
			Brand: shasta.ManagementSwitchBrandAruba,
		}, { // Invalid Xname
			Xname:      "x3000c0w15L",
			SwitchType: shasta.ManagementSwitchTypeSpine,
			Brand:      shasta.ManagementSwitchBrandAruba,
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
		mySwitch      shasta.ManagementSwitch
		expectedError error
	}{{
		// Spine using MgmtSwitch, should be using MgmtHLSwitch
		mySwitch: shasta.ManagementSwitch{
			Xname: "x10c0w14", SwitchType: shasta.ManagementSwitchTypeSpine,
			Brand: shasta.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}, {
		// Spine using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: shasta.ManagementSwitch{
			Xname:      "d10w14",
			SwitchType: shasta.ManagementSwitchTypeSpine,
			Brand:      shasta.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}, {
		// Aggergation using MgmtSwitch, should be using MgmtHLSwitch
		mySwitch: shasta.ManagementSwitch{
			Xname:      "x20c0w14",
			SwitchType: shasta.ManagementSwitchTypeAggregation,
			Brand:      shasta.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}, {
		// Aggergation using CDUMgmtSwitch, should be using MgmtHLSwitch
		mySwitch: shasta.ManagementSwitch{
			Xname:      "d20w14",
			SwitchType: shasta.ManagementSwitchTypeAggregation,
			Brand:      shasta.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}, {
		// CDU using MgmtHLSwitch, should be using CDUMgmtSwitch
		mySwitch: shasta.ManagementSwitch{
			Xname:      "x30c0w14",
			SwitchType: shasta.ManagementSwitchTypeCDU,
			Brand:      shasta.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}, {
		// CDU using MgmtSwitch, should be using CDUMgmtSwitch
		mySwitch: shasta.ManagementSwitch{
			Xname:      "x30c0h14s1",
			SwitchType: shasta.ManagementSwitchTypeCDU,
			Brand:      shasta.ManagementSwitchBrandAruba,
		},
		expectedError: errors.New("switch_metadata.csv contains invalid switch data"),
	}}

	for _, test := range tests {
		switches := []*shasta.ManagementSwitch{&test.mySwitch}
		err := validateSwitchInput(switches)
		suite.Equal(test.expectedError, err)
	}
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_HappyPath() {
	ncns := []*shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "x3000c0s2b0n0", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s3b0n0", Role: "Management", Subrole: "Storage"},
	}

	err := validateNCNInput(ncns)
	suite.NoError(err)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_InvalidXName() {
	ncns := []*shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "foo", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s3b0n0", Role: "Management", Subrole: "Storage"},
	}

	err := validateNCNInput(ncns)
	suite.Equal(errors.New("ncn_metadata.csv contains invalid NCN data"), err)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_WrongXNameType() {
	ncns := []*shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "x3000c0s2b0n0", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s3b0", Role: "Management", Subrole: "Storage"},
	}

	err := validateNCNInput(ncns)
	suite.Equal(errors.New("ncn_metadata.csv contains invalid NCN data"), err)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_ZeroNCNs() {
	ncns := []*shasta.LogicalNCN{}

	err := validateNCNInput(ncns)
	suite.Equal(errors.New("Unable to extract NCNs from ncn metadata csv"), err)
}

func (suite *InitCmdTestSuite) TestMergeNCNs_HappyPath() {
	ncns := []*shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "x3000c0s4b0n0", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s7b0n0", Role: "Management", Subrole: "Storage"},
	}

	slsNCNs := []shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Hostname: "ncn-m001", Aliases: []string{"ncn-m001"}, BmcPort: "x3000c0w14:1/1/31"},
		{Xname: "x3000c0s4b0n0", Hostname: "ncn-w001", Aliases: []string{"ncn-w001"}, BmcPort: "x3000c0w14:1/1/32"},
		{Xname: "x3000c0s7b0n0", Hostname: "ncn-s001", Aliases: []string{"ncn-s001"}, BmcPort: "x3000c0w14:1/1/33"},
	}

	expectedMergeList := []*shasta.LogicalNCN{
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
	ncns := []*shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "x3000c0s4b0n0", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s7b0n0", Role: "Management", Subrole: "Storage"},
	}

	slsNCNs := []shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0", Hostname: "ncn-m001", Aliases: []string{"ncn-m001"}, BmcPort: "x3000c0w14:1/1/31"},
		{Xname: "x3000c0s4b0n0", Hostname: "ncn-w001", Aliases: []string{"ncn-w001"}, BmcPort: "x3000c0w14:1/1/32"},
	}

	err := mergeNCNs(ncns, slsNCNs)
	suite.Equal(errors.New("failed to find NCN from ncn-metadata in generated SLS State: x3000c0s7b0n0"), err)
}

func TestInitCmdTestSuite(t *testing.T) {
	suite.Run(t, new(InitCmdTestSuite))
}
