//go:build !integration || shcd
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

const _schema = "shcd-schema.json"

var _schemaFile = filepath.Join("../internal/files", _schema)

var switchMetaExpected = "../testdata/expected/" + switchMetadata
var hmnConnExpected = "../testdata/expected/" + hmnConnections
var appNodeExpected = "../testdata/expected/" + applicationNodeConfig
var ncnMetaExpected = "../testdata/expected/" + ncnMetadata

// Generate shcd.json example:
// canu validate shcd -a Full --shcd shcd.xlsx --tabs 10G_25G_40G_100G,NMN,HMN,MTN_TDS --corners I37,T125,J15,T24,J20,U51,K15,U36 --out shcd.json
var tests = []struct {
	fixture                       string
	expectedError                 bool
	expectedErrorMsg              string
	expectedSchemaErrorMsg        string
	name                          string
	expectedSwitchMetadata        string
	expectedHMNConnections        string
	expectedApplicationNodeConfig string
	expectedNCNMetadata           string
}{
	{
		fixture:                       "../testdata/fixtures/valid_shcd.json",
		expectedError:                 false,
		expectedErrorMsg:              "",
		expectedSchemaErrorMsg:        "",
		name:                          "ValidFile",
		expectedSwitchMetadata:        switchMetaExpected,
		expectedHMNConnections:        hmnConnExpected,
		expectedApplicationNodeConfig: appNodeExpected,
		expectedNCNMetadata:           ncnMetaExpected,
	},
	{
		fixture:                "../testdata/fixtures/invalid_shcd.json",
		expectedError:          true,
		expectedErrorMsg:       "invalid character ':' after top-level value",
		expectedSchemaErrorMsg: "SHCD schema error: (root): Invalid type. Expected: object, given: string",
		name:                   "MissingBracketFile",
	},
	{
		fixture:                "../testdata/fixtures/invalid_data_types_shcd.json",
		expectedError:          true,
		expectedErrorMsg:       "json: cannot unmarshal string into Go struct field ID.topology.id of type int",
		expectedSchemaErrorMsg: "SHCD schema error: topology.0.id: Invalid type. Expected: integer, given: string",
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
				createHMNSeed(shcd, hmnConnections)

				// Validate the file was created
				assert.FileExists(t, filepath.Join(".", hmnConnections))

				// Read the generated json and validate it's contents
				hmnGenerated, err := os.Open(filepath.Join(".", hmnConnections))

				if err != nil {
					log.Fatal(err)
				}

				defer hmnGenerated.Close()

				hmnExpected, err := os.Open(test.expectedHMNConnections)

				// if we os.Open returns an error then handle it
				if err != nil {
					log.Fatal(err)
				}

				defer hmnExpected.Close()

				hmnActual, _ := ioutil.ReadAll(hmnGenerated)

				hmnConnections, err := ioutil.ReadAll(hmnExpected)

				if err != nil {
					log.Fatal(err)
				}

				assert.JSONEq(t, string(hmnConnections), string(hmnActual))
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
				createSwitchSeed(shcd, switchMetadata)

				// Validate the file was created
				assert.FileExists(t, filepath.Join(".", switchMetadata))

				// Read the csv and validate it's contents
				generated, err := os.Open(filepath.Join(".", switchMetadata))

				if err != nil {
					log.Fatal("Unable to read "+filepath.Join(".", switchMetadata), err)
				}

				defer generated.Close()

				smGenerated := csv.NewReader(generated)

				actual, err := smGenerated.ReadAll()

				if err != nil {
					log.Fatal("Unable to parse as a CSV: "+filepath.Join(".", switchMetadata), err)
				}

				// Read the csv and validate it's contents
				expected, err := os.Open(filepath.Join(".", switchMetadata))

				if err != nil {
					log.Fatal("Unable to read "+filepath.Join(".", switchMetadata), err)
				}

				defer expected.Close()

				csvReader := csv.NewReader(expected)

				smExpected, err := csvReader.ReadAll()

				if err != nil {
					log.Fatal("Unable to parse as a CSV: "+test.expectedSwitchMetadata, err)
				}

				assert.Equal(t, smExpected, actual)
			})
		}
	}
}

func TestCreateNCNMetadata(t *testing.T) {

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

				// Create ncn_metadata.csv
				createNCNSeed(shcd, ncnMetadata)

				// Validate the file was created
				assert.FileExists(t, filepath.Join(".", ncnMetadata))

				// Read the csv and validate it's contents
				generated, err := os.Open(filepath.Join(".", ncnMetadata))

				if err != nil {
					log.Fatal("Unable to read "+filepath.Join(".", ncnMetadata), err)
				}

				defer generated.Close()

				ncnGenerated := csv.NewReader(generated)

				actual, err := ncnGenerated.ReadAll()

				if err != nil {
					log.Fatal("Unable to parse as a CSV: "+filepath.Join(".", ncnMetadata), err)
				}

				// Read the csv and validate it's contents
				expected, err := os.Open(filepath.Join(".", ncnMetadata))

				if err != nil {
					log.Fatal("Unable to read "+filepath.Join(".", ncnMetadata), err)
				}

				defer expected.Close()

				csvReader := csv.NewReader(expected)

				ncnExpected, err := csvReader.ReadAll()

				if err != nil {
					log.Fatal("Unable to parse as a CSV: "+test.expectedNCNMetadata, err)
				}

				assert.Equal(t, ncnExpected, actual)
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
				createANCSeed(shcd, applicationNodeConfig)

				// Validate the file was created
				assert.FileExists(t, filepath.Join(".", applicationNodeConfig))

				// Read the yaml and validate it's contents
				ancGenerated, err := os.Open(filepath.Join(".", applicationNodeConfig))

				if err != nil {
					log.Fatal("Unable to read "+filepath.Join(".", applicationNodeConfig), err)
				}

				defer ancGenerated.Close()

				ancExpected, err := os.Open(test.expectedApplicationNodeConfig)

				// if we os.Open returns an error then handle it
				if err != nil {
					fmt.Println(err)
				}

				defer ancExpected.Close()

				ancActual, _ := ioutil.ReadAll(ancGenerated)

				appNodeConfig, err := ioutil.ReadAll(ancExpected)

				if err != nil {
					fmt.Println(err)
				}

				assert.YAMLEq(t, string(appNodeConfig), string(ancActual))
			})
		}
	}
}

func TestGenerateHMNSourceName(t *testing.T) {
	testCases := []struct {
		desc       string
		commonName string
		want       string
	}{
		{
			desc:       "Common Name bogus returns bogus",
			commonName: "bogus",
			want:       "bogus",
		},
		{
			desc:       "Common Name ncn-m001 returns nm01",
			commonName: "ncn-m001",
			want:       "mn01",
		},
		{
			desc:       "Common Name ncn-w002 returns wn02",
			commonName: "ncn-w002",
			want:       "wn02",
		},
		{
			desc:       "Common Name ncn-s003 returns sn03",
			commonName: "ncn-s003",
			want:       "sn03",
		},
		{
			desc:       "Common Name uan001 returns uan01",
			commonName: "uan001",
			want:       "uan01",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ID := ID{CommonName: tC.commonName}
			got := ID.GenerateHMNSourceName()
			if tC.want != got {
				t.Errorf("want common name %q, got %q", tC.want, got)
			}
		})
	}
}
