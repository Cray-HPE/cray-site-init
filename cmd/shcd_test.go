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

const _schemaFile = "../internal/files/shcd-schema.json"

var tests = []struct {
	fixture                string
	expectedError          bool
	expectedErrorMsg       string
	expectedSchemaErrorMsg string
	name                   string
}{
	{
		fixture:                "../testdata/fixtures/valid_shcd.json",
		expectedError:          false,
		expectedErrorMsg:       "",
		expectedSchemaErrorMsg: "",
		name:                   "ValidFile",
	},
	{
		fixture:                "../testdata/fixtures/invalid_shcd.json",
		expectedError:          true,
		expectedErrorMsg:       "invalid character ',' after top-level value",
		expectedSchemaErrorMsg: "SHCD schema error: (root): Invalid type. Expected: array, given: object",
		name:                   "MissingBracketFile",
	},
	{
		fixture:                "../testdata/fixtures/invalid_data_types_shcd.json",
		expectedError:          true,
		expectedErrorMsg:       "json: cannot unmarshal string into Go struct field Id.id of type int",
		expectedSchemaErrorMsg: "SHCD schema error: 0.id: Invalid type. Expected: integer, given: string",
		name:                   "InvalidDataTypeFile",
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

func TestSHCDAgainstSchema(t *testing.T) {

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			// Validate the file passed against the pre-defined schema
			validSHCD, err := ValidateSchema(test.fixture, _schemaFile)

			if test.expectedError == false {

				// If it meets the schema, it should return true
				assert.Equal(t, validSHCD, true)

			} else {

				// Otherwise, check the error message
				if assert.Error(t, err) {
					assert.EqualError(t, err, test.expectedSchemaErrorMsg)
				}

			}
		})
	}
}
