// +build !integration shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const _schemaFile = "../internal/files/shcd-schema.json"

var tests = []struct {
	fixture                       string
	expectedError                 bool
	expectedErrorMsg              string
	expectedSchemaErrorMsg        string
	name                          string
	expectedSwitchMetadata        [][]string
	expectedHMNConnections        []byte
	expectedApplicationNodeConfig []byte
}{
	{
		fixture:                "../testdata/fixtures/valid_shcd.json",
		expectedError:          false,
		expectedErrorMsg:       "",
		expectedSchemaErrorMsg: "",
		name:                   "ValidFile",
		expectedSwitchMetadata: [][]string{{"Switch Xname", "Type", "Brand"}, {"x1000", "switch", "aruba"}, {"x1000", "switch", "aruba"}, {"x1000", "none", "cray"}, []string{"x3000", "server", "cray"}, []string{"x3000", "server", "cray"}, []string{"x3000", "server", "cray"}, []string{"x3000", "server", "cray"}, []string{"x3000", "server", "cray"}},
		expectedHMNConnections: []byte(`[
{
  "Source": "something",
	"SourceRack": "x1000",
	"SourceLocation": "u42",
	"SourceSubLocation": "",
	"DestinationRack": "",
	"DestinationLocation": 0,
	"DestinationPort": ""
},
{
  "Source": "another_thing",
	"SourceRack": "x1000",
	"SourceLocation": "u42",
	"SourceSubLocation": "",
	"DestinationRack": "",
	"DestinationLocation": 0,
	"DestinationPort": ""
},
{
	"Source": "thingy",
	"SourceRack": "x1000",
	"SourceLocation": "u42",
	"SourceSubLocation": "",
	"DestinationRack": "",
	"DestinationLocation": 0,
	"DestinationPort": ""
},
{
	"Source": "gateway01",
	"SourceRack": "x3000",
	"SourceLocation": "u29",
	"SourceSubLocation": "",
	"DestinationRack": "",
	"DestinationLocation": 0,
	"DestinationPort": ""
},
{
	"Source": "login02",
	"SourceRack": "x3000",
	"SourceLocation": "u28",
	"SourceSubLocation": "",
	"DestinationRack": "",
	"DestinationLocation": 0,
	"DestinationPort": ""
},
{
	"Source": "lnet01",
	"SourceRack": "x3000",
	"SourceLocation": "u27",
	"SourceSubLocation": "",
	"DestinationRack": "",
	"DestinationLocation": 0,
	"DestinationPort": ""
},
{
	"Source": "vn01",
	"SourceRack": "x3000",
	"SourceLocation": "u25",
	"SourceSubLocation": "",
	"DestinationRack": "",
	"DestinationLocation": 0,
	"DestinationPort": ""
},
{
	"Source": "uan01",
	"SourceRack": "x3000",
	"SourceLocation": "u23",
	"SourceSubLocation": "",
	"DestinationRack": "",
	"DestinationLocation": 0,
	"DestinationPort": ""
}
		]`),
		expectedApplicationNodeConfig: []byte(
			`---
# Additional application node prefixes to match in the hmn_connections.json file
prefixes:
  - gateway
  - lnet
  - login
  - vn

# Additional HSM SubRoles
prefix_hsm_subroles:
  gateway: Gateway
  lnet: LNETRouter
  login: UAN
  vn: Visualization

# Application Node aliases
aliases:
  x3000c0s23b0n0: ["uan01"]
  x3000c0s25b0n0: ["vn01"]
  x3000c0s27b0n0: ["lnet01"]
  x3000c0s28b0n0: ["login02"]
  x3000c0s29b0n0: ["gateway01"]
`),
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

func TestCreateHMNConnections(t *testing.T) {

	for _, test := range tests {

		if test.fixture == "../testdata/fixtures/valid_shcd.json" {

			t.Run(test.name, func(t *testing.T) {

				// Open the file since we know it is valid
				shcdFile, err := ioutil.ReadFile(test.fixture)

				if err != nil {
					log.Fatalf(err.Error())
				}

				shcd, err := ParseSHCD(shcdFile)

				if err != nil {
					log.Fatalf(err.Error())
				}

				// Create hmn_connections.json
				createHMNSeed(shcd, hmn_connections)

				// Validate the file was created
				assert.FileExists(t, filepath.Join(".", hmn_connections))

				// Read the csv and validate it's contents
				f, err := os.Open(filepath.Join(".", hmn_connections))

				if err != nil {
					log.Fatal("Unable to read "+filepath.Join(".", hmn_connections), err)
				}

				defer f.Close()

				hmnFile, err := os.Open(filepath.Join(".", hmn_connections))

				// if we os.Open returns an error then handle it
				if err != nil {
					fmt.Println(err)
				}

				defer hmnFile.Close()

				hmnActual, _ := ioutil.ReadAll(hmnFile)

				hmnExpected := test.expectedHMNConnections

				assert.JSONEq(t, string(hmnExpected), string(hmnActual))
			})
		}
	}
}

func TestCreateSwitchMetadata(t *testing.T) {

	for _, test := range tests {

		if test.fixture == "../testdata/fixtures/valid_shcd.json" {

			t.Run(test.name, func(t *testing.T) {

				// Open the file since we know it is valid
				shcdFile, err := ioutil.ReadFile(test.fixture)

				if err != nil {
					log.Fatalf(err.Error())
				}

				shcd, err := ParseSHCD(shcdFile)

				if err != nil {
					log.Fatalf(err.Error())
				}

				// Create switch_metadata.csv
				createSwitchSeed(shcd, switch_metadata)

				// Validate the file was created
				assert.FileExists(t, filepath.Join(".", switch_metadata))

				// Read the csv and validate it's contents
				f, err := os.Open(filepath.Join(".", switch_metadata))

				if err != nil {
					log.Fatal("Unable to read "+filepath.Join(".", switch_metadata), err)
				}

				defer f.Close()

				csvReader := csv.NewReader(f)

				content, err := csvReader.ReadAll()

				if err != nil {
					log.Fatal("Unable to parse as a CSV: "+filepath.Join(".", switch_metadata), err)
				}

				assert.Equal(t, test.expectedSwitchMetadata, content)
			})
		}
	}
}

func TestCreateApplicationNodeConfig(t *testing.T) {

	for _, test := range tests {

		if test.fixture == "../testdata/fixtures/valid_shcd.json" {

			t.Run(test.name, func(t *testing.T) {

				// Open the file since we know it is valid
				shcdFile, err := ioutil.ReadFile(test.fixture)

				if err != nil {
					log.Fatalf(err.Error())
				}

				shcd, err := ParseSHCD(shcdFile)

				if err != nil {
					log.Fatalf(err.Error())
				}

				prefixSubroleMapIn = map[string]string{
					"gateway": "Gateway",
					"login":   "UAN",
					"lnet":    "LNETRouter",
					"vn":      "Visualization",
				}

				// Create application_node_config.yaml
				createANCSeed(shcd, application_node_config)

				// Validate the file was created
				assert.FileExists(t, filepath.Join(".", application_node_config))

				// Read the yaml and validate it's contents
				f, err := os.Open(filepath.Join(".", application_node_config))

				if err != nil {
					log.Fatal("Unable to read "+filepath.Join(".", application_node_config), err)
				}

				defer f.Close()

				ancFile, err := os.Open(filepath.Join(".", application_node_config))

				// if we os.Open returns an error then handle it
				if err != nil {
					fmt.Println(err)
				}

				defer ancFile.Close()

				ancActual, _ := ioutil.ReadAll(ancFile)

				ancExpected := test.expectedApplicationNodeConfig

				assert.YAMLEq(t, string(ancExpected), string(ancActual))
			})
		}
	}
}
