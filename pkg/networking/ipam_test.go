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

package networking

import (
	"errors"
	"fmt"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/suite"
)

type NetworksTestSuite struct {
	suite.Suite
}

func (suite *NetworksTestSuite) TestAdd() {

	tests := []struct {
		prefix   netip.Prefix
		toAdd    uint64
		expected netip.Addr
	}{
		{
			prefix:   netip.MustParsePrefix("10.0.0.0/24"),
			toAdd:    1,
			expected: netip.MustParseAddr("10.0.0.1"),
		},
		{
			prefix:   netip.MustParsePrefix("10.0.4.0/24"),
			toAdd:    300,
			expected: netip.MustParseAddr("10.0.4.255"),
		},
		{
			prefix:   netip.MustParsePrefix("fdf8:413:de2c:204::/64"),
			toAdd:    18446744073709551615,
			expected: netip.MustParseAddr("fdf8:413:de2c:204:ffff:ffff:ffff:ffff"),
		},
	}
	for _, test := range tests {
		actual := Add(
			test.prefix,
			test.toAdd,
		)
		suite.Equal(
			test.expected,
			actual,
			fmt.Sprintf(
				"Adding %d to %s expected %s, actual %s",
				test.toAdd,
				test.prefix.String(),
				test.expected.String(),
				actual.String(),
			),
		)
	}
}

func (suite *NetworksTestSuite) TestUsableHostAddresses() {

	tests := []struct {
		prefix   netip.Prefix
		expected uint64
	}{
		{
			prefix:   netip.MustParsePrefix("10.1.0.0/24"),
			expected: 254,
		},
		{
			prefix:   netip.MustParsePrefix("10.1.0.0/16"),
			expected: 65534,
		},
		{
			prefix:   netip.MustParsePrefix("10.1.0.0/8"),
			expected: 16777214,
		},
		{
			prefix:   netip.MustParsePrefix("192.168.4.3/17"),
			expected: 32766,
		},
		{
			prefix:   netip.MustParsePrefix("fdf8:413:de2c:204::/64"),
			expected: 18446744073709551615,
		},
		{
			// Test that CIDRs larger than /64 return /64 because they're too big.
			prefix:   netip.MustParsePrefix("fdf8:413:de2c:204::/63"),
			expected: 18446744073709551615,
		},
		{
			prefix:   netip.MustParsePrefix("fdf8:413:de2c:204::/126"),
			expected: 3,
		},
	}
	for _, test := range tests {
		actual, err := UsableHostAddresses(test.prefix)
		suite.Equal(
			test.expected,
			actual,
			fmt.Sprintf(
				"UsableHostAddresses(%x) expected %d, actual %d",
				test.prefix.String(),
				test.expected,
				actual,
			),
		)
		suite.Nil(err)
	}
}

func (suite *NetworksTestSuite) TestBroadcast() {

	tests := []struct {
		prefix   netip.Prefix
		expected netip.Addr
	}{
		{
			prefix:   netip.MustParsePrefix("10.1.0.0/24"),
			expected: netip.MustParseAddr("10.1.0.255"),
		},
		{
			prefix:   netip.MustParsePrefix("10.1.0.0/16"),
			expected: netip.MustParseAddr("10.1.255.255"),
		},
		{
			prefix:   netip.MustParsePrefix("10.1.20.30/16"),
			expected: netip.MustParseAddr("10.1.255.255"),
		},
		{
			prefix:   netip.MustParsePrefix("10.21.20.30/8"),
			expected: netip.MustParseAddr("10.255.255.255"),
		},
	}
	for _, test := range tests {
		actual, err := Broadcast(test.prefix)
		suite.Equal(
			test.expected,
			actual,
			fmt.Sprintf(
				"Broadcast(%x) expected %s, actual %s",
				test.prefix.String(),
				test.expected,
				actual.String(),
			),
		)
		suite.Nil(err)
	}
}

func (suite *NetworksTestSuite) TestPrefixLengthToSubnetMask() {

	tests := []struct {
		prefixLength int
		bits         int
		expected     netip.Addr
	}{
		{
			prefixLength: 24,
			bits:         IPv4Size,
			expected:     netip.MustParseAddr("255.255.255.0"),
		},
		{
			prefixLength: 65,
			bits:         IPv6Size,
			expected:     netip.MustParseAddr("ffff:ffff:ffff:ffff:8000::"),
		},
	}
	for _, test := range tests {
		actual, err := PrefixLengthToSubnetMask(
			test.prefixLength,
			test.bits,
		)
		suite.Equal(
			test.expected,
			actual,
			fmt.Sprintf(
				"TestPrefixLengthToSubnetMask(%x) expected %s, actual %s",
				test.prefixLength,
				test.expected.String(),
				actual.String(),
			),
		)
		suite.Nil(err)
	}
}

func (suite *NetworksTestSuite) TestFindCIDRRootIP() {

	tests := []struct {
		prefix   netip.Prefix
		expected netip.Addr
	}{
		{
			prefix:   netip.MustParsePrefix("10.1.100.2/24"),
			expected: netip.MustParseAddr("10.1.100.0"),
		},
		{
			prefix:   netip.MustParsePrefix("10.120.234.45/27"),
			expected: netip.MustParseAddr("10.120.234.32"),
		},
		{
			prefix:   netip.MustParsePrefix("fdf8:413:de2c:205::223/64"),
			expected: netip.MustParseAddr("fdf8:413:de2c:205::"),
		},
	}
	for _, test := range tests {
		actual, err := FindCIDRRootIP(test.prefix)
		suite.Equal(
			test.expected,
			actual,
			fmt.Sprintf(
				"FindCIDRRootIP(%x) expected %s, actual %s",
				test.prefix.String(),
				test.expected.String(),
				actual.String(),
			),
		)
		suite.Nil(err)
	}
}

func (suite *NetworksTestSuite) TestFindGatewayIP() {

	tests := []struct {
		prefix   netip.Prefix
		expected netip.Addr
	}{
		{
			prefix:   netip.MustParsePrefix("10.1.100.2/24"),
			expected: netip.MustParseAddr("10.1.100.1"),
		},
		{
			prefix:   netip.MustParsePrefix("10.120.234.45/27"),
			expected: netip.MustParseAddr("10.120.234.33"),
		},
		{
			prefix:   netip.MustParsePrefix("fdf8:413:de2c:205::223/64"),
			expected: netip.MustParseAddr("fdf8:413:de2c:205::1"),
		},
	}
	for _, test := range tests {
		actual := FindGatewayIP(test.prefix)
		suite.Equal(
			test.expected,
			actual,
			fmt.Sprintf(
				"FindGatewayIP(%x) expected %s, actual %s",
				test.prefix.String(),
				test.expected.String(),
				actual.String(),
			),
		)
	}
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
		val, err := IsVLANAllocated(uint16(test.vlan))
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
			// VLAN 0 is untagged (not a real VLAN).
			vlan:          0,
			expectedError: nil,
		},
		{
			// VLAN 7 not allocated and this should work
			vlan:          7,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		err := AllocateVLAN(uint16(test.vlan))
		suite.Equal(
			test.expectedError,
			err,
		)
	}
}

func (suite *NetworksTestSuite) TestVlanAllocate() {
	// Allocate VLAN 2 the first time should be success
	var vlan uint16 = 2
	suite.Equal(
		nil,
		AllocateVLAN(vlan),
	)
	// Allocate VLAN 2 the 2nd time should fail as it's already used
	suite.Equal(
		errors.New("VLAN already used"),
		AllocateVLAN(vlan),
	)
}

func (suite *NetworksTestSuite) TestVlanAllocateRange() {
	// Bad start and end range
	suite.Equal(
		errors.New("VLAN out of range"),
		AllocateVlanRange(
			4092,
			4099,
		),
	)

	var startVlan int16 = 200
	var endVlan int16 = 299

	// Bad start and end range
	suite.Equal(
		errors.New("VLAN range is bad - start is larger than end"),
		AllocateVlanRange(
			endVlan,
			startVlan,
		),
	)

	// VLANs already in use in the range cannot be allocated, including ends
	_ = AllocateVLAN(200)
	_ = AllocateVLAN(207)
	_ = AllocateVLAN(213)
	_ = AllocateVLAN(299)
	suite.Equal(
		errors.New("VLANs already used: [200 207 213 299]"),
		AllocateVlanRange(
			startVlan,
			endVlan,
		),
	)

	// Successfully allocate a range of VLANs
	FreeVLAN(200)
	FreeVLAN(207)
	FreeVLAN(213)
	FreeVLAN(299)
	suite.Equal(
		nil,
		AllocateVlanRange(
			startVlan,
			endVlan,
		),
	)
}

func (suite *NetworksTestSuite) TestVlanFree() {
	// Allocate VLAN 4
	var vlan uint16 = 4
	err := AllocateVLAN(vlan)
	suite.Nil(
		err,
		fmt.Sprintf(
			"AllocateVLAN(%d) returned an error: %v",
			vlan,
			err,
		),
	)
	suite.True(
		VLANs[vlan],
		fmt.Sprintf(
			"expected VLAN %d to be allocated but allocation status was %v",
			vlan,
			VLANs[vlan],
		),
	)

	// Deallocate VLAN 4 the 1st time should succeed
	FreeVLAN(vlan)
	suite.False(
		VLANs[vlan],
		fmt.Sprintf(
			"expected VLAN %d to be freed but allocation status was %v",
			vlan,
			VLANs[vlan],
		),
	)

	FreeVLAN(vlan)
	// Deallocate VLAN 4 the 2nd time should succeed
	suite.False(
		VLANs[vlan],
		fmt.Sprintf(
			"expected VLAN %d to be freed but its allocation status was %v",
			vlan,
			VLANs[vlan],
		),
	)
}

func (suite *NetworksTestSuite) TestVlanFreeRange() {
	var startVlan uint16 = 400
	var endVlan uint16 = 499
	err := freeVLANRange(
		endVlan,
		startVlan,
	)
	suite.Error(
		err,
		fmt.Sprintf(
			"freeVLANRange(%d,%d) did not return an error when it should have.",
			endVlan,
			startVlan,
		),
	)

	err = AllocateVLAN(startVlan)
	suite.Nil(
		err,
		fmt.Sprintf(
			"AllocateVLAN(%d) returned an error: %v",
			startVlan,
			err,
		),
	)

	err = AllocateVLAN(endVlan)
	suite.Nil(
		err,
		fmt.Sprintf(
			"AllocateVLAN(%d) returned an error: %v",
			endVlan,
			err,
		),
	)

	// Deallocate the range the first time - should succeed

	err = freeVLANRange(
		startVlan,
		endVlan,
	)
	suite.Nil(
		err,
		fmt.Sprintf(
			"freeVLANRange(%d,%d) returned an error: %v",
			startVlan,
			endVlan,
			err,
		),
	)

	startAllocation, err := IsVLANAllocated(uint16(startVlan))
	suite.Nil(
		err,
		fmt.Sprintf(
			"IsVLANAllocated(%d) returned an error: %v",
			startVlan,
			err,
		),
	)
	suite.False(
		startAllocation,
	)

	endAllocation, err := IsVLANAllocated(uint16(endVlan))
	suite.Nil(
		err,
		fmt.Sprintf(
			"IsVLANAllocated(%d) returned an error: %v",
			endVlan,
			err,
		),
	)
	suite.False(
		endAllocation,
	)

	// Deallocate the range the second time - should still succeed
	err = freeVLANRange(
		startVlan,
		endVlan,
	)
	suite.Nil(
		err,
		fmt.Sprintf(
			"freeVLANRange(%d,%d) returned an error: %v",
			startVlan,
			endVlan,
			err,
		),
	)
}

func TestNetworksTestSuite(t *testing.T) {
	suite.Run(
		t,
		new(NetworksTestSuite),
	)
}
