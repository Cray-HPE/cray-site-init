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
