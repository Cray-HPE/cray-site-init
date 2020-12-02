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

func (suite *InitCmdTestSuite) TestValidateSwitchInput() {

}

func (suite *InitCmdTestSuite) TestValidateNCNInput_HappyPath() {
	ncns := []*shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0"},
		{Xname: "x3000c0s2b0n0"},
		{Xname: "x3000c0s3b0n0"},
	}

	err := validateNCNInput(ncns)
	suite.NoError(err)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_InvalidXName() {
	ncns := []*shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0"},
		{Xname: "foo"},
		{Xname: "x3000c0s3b0n0"},
	}

	err := validateNCNInput(ncns)
	suite.Error(errors.New("invalid xname for NCN: foo"), err)
}

func (suite *InitCmdTestSuite) TestValidateNCNInput_WrongXNameType() {
	ncns := []*shasta.LogicalNCN{
		{Xname: "x3000c0s1b0n0"},
		{Xname: "x3000c0s2b0n0"},
		{Xname: "x3000c0s3b0"},
	}

	err := validateNCNInput(ncns)
	suite.Error(errors.New("invalid type NodeBMC for NCN xname: x3000c0s3b0"), err)
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
	suite.Error(errors.New("failed to find NCN from ncn-metadata in generated SLS State: x3000c0s7b0n0"), err)
}

func TestInitCmdTestSuite(t *testing.T) {
	suite.Run(t, new(InitCmdTestSuite))
}
