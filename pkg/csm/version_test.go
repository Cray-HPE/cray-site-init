/*
 MIT License

 (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP

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

package csm

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// csmVersionStrings Valid CSM release strings that should always be detected. Any semver should be
// detected, however CSM releases are given with or without the V depending on context (e.g. on or off
// a system).
var csmVersionStrings = []struct {
	in  string
	out string
}{
	{
		"1",
		"v1",
	},
	{
		"1.0",
		"v1.0",
	},
	{
		"1.0.0",
		"v1.0.0",
	},
	{
		"1.0.0-alpha.1",
		"v1.0.0-alpha.1",
	},
	{
		"1.0.0-beta.11",
		"v1.0.0-beta.11",
	},
	{
		"1.0.0-beta.2",
		"v1.0.0-beta.2",
	},
	{
		"1.0.0-rc.1",
		"v1.0.0-rc.1",
	},
	{
		"1.2",
		"v1.2",
	},
	{
		"1.2.0",
		"v1.2.0",
	},
}

// versionCompare are versions and their expected outcomes when compared against the CurrentVersion.
var versionCompare = []struct {
	in  string
	out int
}{
	{
		"v1.0.0",
		1,
	},
	{
		"v2.0.0",
		-1,
	},
}

type VersionTestSuite struct {
	suite.Suite
	ConfigThatShouldStartAsNew          *viper.Viper
	VersionThatShouldBeCompatible       string
	VersionThatShouldBeCompatibleSemVer string
	VersionThatShouldBeRandom           float64
}

func (suite *VersionTestSuite) SetupTest() {
	os.Clearenv()
	viper.Reset()
	suite.ConfigThatShouldStartAsNew = viper.GetViper()
	f, _ := strconv.ParseFloat(
		MinimumVersion[1:],
		32,
	)
	suite.VersionThatShouldBeCompatible = strconv.FormatFloat(
		f+0.1,
		'f',
		1,
		64,
	)
	suite.VersionThatShouldBeCompatibleSemVer = fmt.Sprintf(
		"v%s",
		suite.VersionThatShouldBeCompatible,
	)
}

func (suite *VersionTestSuite) TestVersionUnset() {
	actual, err := currentVersion()
	assert.Error(
		suite.T(),
		err,
	)
	assert.Empty(
		suite.T(),
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionNil() {
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		nil,
	)
	actual, err := currentVersion()
	assert.Error(
		suite.T(),
		err,
	)
	assert.Empty(
		suite.T(),
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionEmpty() {
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		"",
	)
	actual, err := currentVersion()
	assert.Error(
		suite.T(),
		err,
	)
	assert.Empty(
		suite.T(),
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionFloat() {
	f, err := strconv.ParseFloat(
		suite.VersionThatShouldBeCompatible,
		64,
	)
	assert.Nil(
		suite.T(),
		err,
	)
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		f,
	)
	actual, err := currentVersion()
	expected := suite.VersionThatShouldBeCompatibleSemVer
	assert.Nil(
		suite.T(),
		err,
	)
	assert.Equal(
		suite.T(),
		expected,
		actual,
		"",
	)
}

func (suite *VersionTestSuite) TestVersionString() {
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		suite.VersionThatShouldBeCompatible,
	)
	actual, err := currentVersion()
	expected := suite.VersionThatShouldBeCompatibleSemVer
	assert.Nil(
		suite.T(),
		err,
	)
	assert.Equal(
		suite.T(),
		expected,
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionUnknownCompatible() {
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		nil,
	)
	actual, err := IsCompatible()
	assert.Error(
		suite.T(),
		err,
	)
	assert.Empty(
		suite.T(),
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionIsIncompatible() {
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		"0.0",
	)
	actual, err := IsCompatible()
	expected := suite.VersionThatShouldBeCompatibleSemVer
	assert.Error(
		suite.T(),
		err,
	)
	assert.NotEqual(
		suite.T(),
		expected,
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionIsCompatibleFloat() {
	f, err := strconv.ParseFloat(
		suite.VersionThatShouldBeCompatible,
		64,
	)
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		f,
	)
	assert.Nil(
		suite.T(),
		err,
	)
	actual, err := IsCompatible()
	expected := suite.VersionThatShouldBeCompatibleSemVer
	assert.Nil(
		suite.T(),
		err,
	)
	assert.Equal(
		suite.T(),
		expected,
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionIsCompatibleString() {
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		suite.VersionThatShouldBeCompatible,
	)
	actual, err := IsCompatible()
	expected := suite.VersionThatShouldBeCompatibleSemVer
	assert.Equal(
		suite.T(),
		expected,
		actual,
	)
	assert.Nil(
		suite.T(),
		err,
	)
}

func (suite *VersionTestSuite) TestVersionDetectedVersionUnset() {
	actual, err := DetectedVersion()
	assert.NotNil(
		suite.T(),
		err,
	)
	assert.Empty(
		suite.T(),
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionDetectedVersionNotSemver() {
	osErr := os.Setenv(
		APIEnvName,
		"foo",
	)
	assert.Nil(
		suite.T(),
		osErr,
	)
	actual, err := DetectedVersion()
	assert.NotNil(
		suite.T(),
		err,
	)
	assert.Empty(
		suite.T(),
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionDetectedVersionSemver() {
	for _, v := range csmVersionStrings {
		osErr := os.Setenv(
			APIEnvName,
			v.in,
		)
		assert.Nil(
			suite.T(),
			osErr,
		)
		actual, err := DetectedVersion()
		expected := v.out
		assert.Nil(
			suite.T(),
			err,
		)
		assert.Equal(
			suite.T(),
			expected,
			actual,
		)
	}
}

func (suite *VersionTestSuite) TestVersionCompare() {
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		suite.VersionThatShouldBeCompatible,
	)
	expected := suite.VersionThatShouldBeCompatibleSemVer
	for _, v := range versionCompare {
		actual, eval := Compare(v.in)
		assert.Equal(
			suite.T(),
			expected,
			actual,
		)
		assert.Equal(
			suite.T(),
			v.out,
			eval,
		)
	}
	actual, eval := Compare(suite.VersionThatShouldBeCompatible)
	assert.Zero(
		suite.T(),
		eval,
	)
	assert.Equal(
		suite.T(),
		expected,
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionCompareMajorMinor() {
	suite.ConfigThatShouldStartAsNew.Set(
		APIKeyName,
		suite.VersionThatShouldBeCompatible,
	)
	expected := suite.VersionThatShouldBeCompatibleSemVer
	for _, v := range versionCompare {
		actual, eval := CompareMajorMinor(v.in)
		assert.Equal(
			suite.T(),
			expected,
			actual,
		)
		assert.Equal(
			suite.T(),
			v.out,
			eval,
		)
	}
	actual, eval := CompareMajorMinor(suite.VersionThatShouldBeCompatible)
	assert.Zero(
		suite.T(),
		eval,
	)
	assert.Equal(
		suite.T(),
		expected,
		actual,
	)
}

func (suite *VersionTestSuite) TestVersionVshastaV2() {
	for _, v := range csmVersionStrings {
		osErr := os.Setenv(
			APIEnvName,
			fmt.Sprintf(
				"https://artifactory.example.com/artifactory/csm-releases/csm/%s/csm-%s.tar.gz",
				v.in,
				v.in,
			),
		)
		assert.Nil(
			suite.T(),
			osErr,
		)
		actual, err := DetectedVersion()
		assert.Nil(
			suite.T(),
			err,
		)
		assert.Equal(
			suite.T(),
			v.out,
			actual,
		)
	}
}

func TestVersionSuite(t *testing.T) {
	suite.Run(
		t,
		new(VersionTestSuite),
	)
}
