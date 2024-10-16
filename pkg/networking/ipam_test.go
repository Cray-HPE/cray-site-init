/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

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

package networking

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type NetworksTestSuite struct {
	suite.Suite
}

func (suite *NetworksTestSuite) TestIsVlanAllocatedBadVlans() {
	tests := []struct {
		vlan          int16
		expectedBool  bool
		expectedError error
	}{
		{
			// VLAN out of range too small
			vlan:          -1,
			expectedBool:  true,
			expectedError: errors.New("VLAN out of range"),
		},
		{
			// VLAN out of range too big
			vlan:          4096,
			expectedBool:  true,
			expectedError: errors.New("VLAN out of range"),
		},
	}

	for _, test := range tests {
		val, err := IsVlanAllocated(test.vlan)
		suite.Equal(
			val,
			test.expectedBool,
		)
		suite.Equal(
			test.expectedError,
			err,
		)
	}
}

func (suite *NetworksTestSuite) TestAllocateVlanBadVlans() {
	tests := []struct {
		vlan          int16
		expectedError error
	}{
		{
			// VLAN out of range too small
			vlan:          -1,
			expectedError: errors.New("VLAN out of range"),
		},
		{
			// VLAN out of range too big
			vlan:          4096,
			expectedError: errors.New("VLAN out of range"),
		},
		{
			// VLAN 7 not allocated and this should work
			vlan:          7,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		err := AllocateVlan(test.vlan)
		suite.Equal(
			test.expectedError,
			err,
		)
	}
}

func (suite *NetworksTestSuite) TestVlanAllocate() {
	// Allocate VLAN 2 the first time should be success
	var vlan int16 = 2
	suite.Equal(
		nil,
		AllocateVlan(vlan),
	)
	//Allocate VLAN 2 the 2nd time should fail as it's already used
	suite.Equal(
		errors.New("VLAN already used"),
		AllocateVlan(vlan),
	)
}

func (suite *NetworksTestSuite) TestVlanAllocateRange() {
	// Bad start and end range
	suite.Equal(
		errors.New("VLAN out of range"),
		AllocateVlanRange(4092, 4099),
	)

	var startVlan int16 = 200
	var endVlan int16 = 299

	// Bad start and end range
	suite.Equal(
		errors.New("VLAN range is bad - start is larger than end"),
		AllocateVlanRange(endVlan, startVlan),
	)

	// VLANs already in use in the range cannot be allocated, including ends
	AllocateVlan(200)
	AllocateVlan(207)
	AllocateVlan(213)
	AllocateVlan(299)
	suite.Equal(
		errors.New("VLANs already used: [200 207 213 299]"),
		AllocateVlanRange(startVlan, endVlan),
	)

	// Successfully allocate a range of VLANs
	FreeVlan(200)
	FreeVlan(207)
	FreeVlan(213)
	FreeVlan(299)
	suite.Equal(
		nil,
		AllocateVlanRange(startVlan, endVlan),
	)
}

func (suite *NetworksTestSuite) TestVlanFree() {
	// Allocate VLAN 4
	var vlan int16 = 4
	suite.Equal(
		nil,
		AllocateVlan(vlan),
	)

	// Deallocate VLAN 4 the 1st time should succeed
	suite.Equal(
		nil,
		FreeVlan(vlan),
	)

	// Deallocate VLAN 4 the 2nd time should succeed
	suite.Equal(
		nil,
		FreeVlan(vlan),
	)
}

func (suite *NetworksTestSuite) TestVlanFreeRange() {
	var startVlan int16 = 400
	var endVlan int16 = 499
	suite.Equal(
		errors.New("VLAN range is bad - start is larger than end"),
		FreeVlanRange(endVlan, startVlan),
	)

	AllocateVlan(startVlan)
	AllocateVlan(endVlan)
	// Deallocate the range the first time - should succeed
	suite.Equal(
		nil,
		FreeVlanRange(startVlan, endVlan),
	)

	startAllocation, _ := IsVlanAllocated(startVlan)
	suite.Equal(
		false,
		startAllocation,
	)

	endAllocation, _ := IsVlanAllocated(endVlan)
	suite.Equal(
		false,
		endAllocation,
	)

	// Deallocate the range the second time - should still succeed
	suite.Equal(
		nil,
		FreeVlanRange(startVlan, endVlan),
	)

}
func TestNetworksTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(NetworksTestSuite),
	)
}
