/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NetworkingTestSuite struct {
	suite.Suite
}

func (suite *NetworkingTestSuite) TestIsManagementSwitchTypeValid() {
	switchTypes := []ManagementSwitchType{
		ManagementSwitchTypeAggregation,
		ManagementSwitchTypeCDU,
		ManagementSwitchTypeLeaf,
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

func TestNetworkingTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkingTestSuite))
}
