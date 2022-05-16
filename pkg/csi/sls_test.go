//go:build !integration && !shcd
// +build !integration,!shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

import (
	"testing"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/stretchr/testify/suite"
)

type SLSTestSuite struct {
	suite.Suite
}

func (suite *SLSTestSuite) TestGetSLSCabinets() {
	slsState := sls_common.SLSState{
		Hardware: map[string]sls_common.GenericHardware{
			// Cabinets: 1 River, 2 Hill, 3 Mountain
			"x3000": {Xname: "x3000", Type: sls_common.Cabinet, Class: sls_common.ClassRiver},
			"x5000": {Xname: "x5000", Type: sls_common.Cabinet, Class: sls_common.ClassHill},
			"x5001": {Xname: "x5001", Type: sls_common.Cabinet, Class: sls_common.ClassHill},
			"x1000": {Xname: "x1000", Type: sls_common.Cabinet, Class: sls_common.ClassMountain},
			"x1001": {Xname: "x1001", Type: sls_common.Cabinet, Class: sls_common.ClassMountain},
			"x1002": {Xname: "x1002", Type: sls_common.Cabinet, Class: sls_common.ClassMountain},

			// Extra SLS Data, to ignore
			"x3000c0s1b0n0": {Xname: "x3000c0s4b0n0", Type: sls_common.Node, Class: sls_common.ClassRiver},
			"x5000c0s1b0n0": {Xname: "x5000c0s1b0n0", Type: sls_common.Node, Class: sls_common.ClassHill},
			"x9000c0s1b0n0": {Xname: "x9000c0s1b0n0", Type: sls_common.Node, Class: sls_common.ClassMountain},
		},
	}

	suite.Len(GetSLSCabinets(slsState, sls_common.ClassRiver), 1, "River Cabinets")
	suite.Len(GetSLSCabinets(slsState, sls_common.ClassHill), 2, "Hill Cabinets")
	suite.Len(GetSLSCabinets(slsState, sls_common.ClassMountain), 3, "Mountain Cabinets")
}

func TestSLSTestSuite(t *testing.T) {
	suite.Run(t, new(SLSTestSuite))
}
