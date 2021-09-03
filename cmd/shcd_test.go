// +build !integration shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	fixture          string
	expectedError    bool
	expectedErrorMsg string
	name             string
}{
	{
		fixture:          "../testdata/fixtures/valid_shcd.json",
		expectedError:    false,
		expectedErrorMsg: "",
		name:             "ValidFile",
	},
	{
		fixture:          "../testdata/fixtures/invalid_shcd.json",
		expectedError:    true,
		expectedErrorMsg: "invalid character ',' after top-level value",
		name:             "MissingBracketFile",
	},
	{
		fixture:          "../testdata/fixtures/invalid_data_types_shcd.json",
		expectedError:    true,
		expectedErrorMsg: "json: cannot unmarshal string into Go struct field Id.id of type int",
		name:             "InvalidDataTypeFile",
	},
}

// Test different JSON input files
func TestValidSHCDJSONTest(t *testing.T) {

	expectedType := Shcd{}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			// Open the file
			f, err := ioutil.ReadFile(test.fixture)

			if err != nil {
				t.Fatalf("%v", err)
			}

			// Test the shcd file to see if it is parsed properly
			shcd, err := ParseSHCD(f)

			// returnedErr := err != nil

			if test.expectedError == false {
				// A valid, machine-readable shcd should return no errors
				assert.NoError(t, err)
				// and be of type Shcd
				assert.IsType(t, expectedType, shcd)
			} else {
				if assert.Error(t, err) {
					assert.EqualError(t, err, test.expectedErrorMsg)
				}
			}
		})
	}
}
