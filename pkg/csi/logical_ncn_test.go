//go:build !integration && !shcd
// +build !integration,!shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type NCNBootStrapTestSuite struct {
	suite.Suite
}

func (suite *NCNBootStrapTestSuite) TestValidateNCNInput_HappyPath() {
	ncns := []LogicalNCN{
		{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: "Master"},
		{Xname: "x3000c0s2b0n0", Role: "Management", Subrole: "Worker"},
		{Xname: "x3000c0s3b0n0", Role: "Management", Subrole: "Storage"},
	}

	for _, ncn := range ncns {
		err := ncn.Validate()
		suite.NoError(err)
	}
}

func (suite *NCNBootStrapTestSuite) TestValidateNCNInput_InvalidXName() {
	tests := []struct {
		ncn           LogicalNCN
		expectedError error
	}{{
		ncn:           LogicalNCN{Xname: "foo", Role: "Management", Subrole: "Master"},
		expectedError: errors.New("invalid xname for NCN: foo"),
	}, {
		ncn:           LogicalNCN{Xname: "x3000c0x3b0n0", Role: "Management", Subrole: "Master"},
		expectedError: errors.New("invalid xname for NCN: x3000c0x3b0n0"),
	}}

	for _, test := range tests {
		err := test.ncn.Validate()
		suite.Equal(test.expectedError, err)
	}
}

func (suite *NCNBootStrapTestSuite) TestValidateNCNInput_WrongXNameType() {
	tests := []struct {
		ncn           LogicalNCN
		expectedError error
	}{{
		ncn:           LogicalNCN{Xname: "x3000", Role: "Management", Subrole: "Master"},
		expectedError: errors.New("invalid type Cabinet for NCN xname: x3000"),
	}, {
		ncn:           LogicalNCN{Xname: "x3000c0s3b0", Role: "Management", Subrole: "Master"},
		expectedError: errors.New("invalid type NodeBMC for NCN xname: x3000c0s3b0"),
	}}

	for _, test := range tests {
		err := test.ncn.Validate()
		suite.Equal(test.expectedError, err)
	}
}

func (suite *NCNBootStrapTestSuite) TestValidateNCNInput_EmptyRole() {
	tests := []struct {
		ncn           LogicalNCN
		expectedError error
	}{{
		ncn:           LogicalNCN{Xname: "x3000c0s1b0n0", Role: "", Subrole: "Master"},
		expectedError: errors.New("empty role"),
	}, {
		ncn:           LogicalNCN{Xname: "x3000c0s2b0n0", Role: "", Subrole: "Master"},
		expectedError: errors.New("empty role"),
	}}

	for _, test := range tests {
		err := test.ncn.Validate()
		suite.Equal(test.expectedError, err)
	}
}

func (suite *NCNBootStrapTestSuite) TestValidateNCNInput_EmptySubRole() {
	tests := []struct {
		ncn           LogicalNCN
		expectedError error
	}{{
		ncn:           LogicalNCN{Xname: "x3000c0s1b0n0", Role: "Management", Subrole: ""},
		expectedError: errors.New("empty sub-role"),
	}, {
		ncn:           LogicalNCN{Xname: "x3000c0s2b0n0", Role: "Management", Subrole: ""},
		expectedError: errors.New("empty sub-role"),
	}}

	for _, test := range tests {
		err := test.ncn.Validate()
		suite.Equal(test.expectedError, err)
	}
}

func (suite *NCNBootStrapTestSuite) TestLogicalNCNNormalize() {
	tests := []struct {
		ncn           LogicalNCN
		expectedXname string
		expectedError error
	}{{
		// Already normalized Node xname
		ncn:           LogicalNCN{Xname: "x3000c0s17b0n0"},
		expectedXname: "x3000c0s17b0n0",
		expectedError: nil,
	}, {
		// Un-normalized Node xname with 0 padding in the slot
		ncn:           LogicalNCN{Xname: "x3000c0s05b0n0"},
		expectedXname: "x3000c0s5b0n0",
		expectedError: nil,
	}, {
		// Un-normalized Node xname with 0 padding in each number
		ncn:           LogicalNCN{Xname: "x03000c00s005b000n00"},
		expectedXname: "x3000c0s5b0n0",
		expectedError: nil,
	}}

	for _, test := range tests {
		err := test.ncn.Normalize()
		suite.Equal(test.expectedError, err)
		suite.Equal(test.expectedXname, test.ncn.Xname)
	}
}

func TestNCNBootStrapTestSuite(t *testing.T) {
	suite.Run(t, new(NCNBootStrapTestSuite))
}
