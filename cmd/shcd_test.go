//go:build !integration || shcd
// +build !integration shcd

/*
 *
 *  MIT License
 *
 *  (C) Copyright 2022 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
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

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

var (
	switchMetaExpected = "../testdata/expected/" + switchMetadata
	hmnConnExpected    = "../testdata/expected/" + hmnConnections
	appNodeExpected    = "../testdata/expected/" + applicationNodeConfig
	ncnMetaExpected    = "../testdata/expected/" + ncnMetadata
)

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
		expectedSchemaErrorMsg: "(root): Invalid type. Expected: object, given: string",
		name:                   "MissingBracketFile",
	},
	{
		fixture:                "../testdata/fixtures/invalid_data_types_shcd.json",
		expectedError:          true,
		expectedErrorMsg:       "json: cannot unmarshal string into Go struct field ID.topology.id of type int",
		expectedSchemaErrorMsg: "topology.0.id: Invalid type. Expected: integer, given: string",
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
			err := ValidateSchema(test.fixture, shcdSchemaFile)
			if test.expectedError == true {
				// Otherwise, check the error message
				if assert.Error(t, err) {
					assert.EqualError(t, err, test.expectedSchemaErrorMsg)
				}
			}
		})
	}
}

func TestCreateHMNConnections(t *testing.T) {
	t.Parallel()
	shcdFile, err := ioutil.ReadFile("../testdata/fixtures/valid_shcd.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	shcd, err := ParseSHCD(shcdFile)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Create hmn_connections.json
	err = createHMNSeed(shcd.Topology)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Validate the file was created
	assert.FileExists(t, filepath.Join(".", hmnConnections))
	// Read the generated json and validate it's contents
	hmnGenerated, err := os.Open(filepath.Join(".", hmnConnections))
	if err != nil {
		t.Fatal(err)
	}
	defer hmnGenerated.Close()
	hmnExpected, err := os.Open(hmnConnExpected)
	if err != nil {
		t.Fatal(err)
	}
	defer hmnExpected.Close()
	hmnActual, err := ioutil.ReadAll(hmnGenerated)
	if err != nil {
		t.Fatal(err)
	}
	hmnConnections, err := ioutil.ReadAll(hmnExpected)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, string(hmnConnections), string(hmnActual))

}

func TestCreateSwitchMetadata(t *testing.T) {
	t.Parallel()
	jsonFilePath := "../testdata/fixtures/valid_shcd.json"
	// Open the file without validating it since we know it is valid
	shcdFile, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatal(err)
	}
	shcd, err := ParseSHCD(shcdFile)
	if err != nil {
		log.Fatal(err)
	}
	err = createSwitchSeed(shcd.Topology)
	if err != nil {
		log.Fatal(err)
	}
	// Read the csv and validate it's contents
	generated, err := os.Open(filepath.Join(".", switchMetadata))
	if err != nil {
		log.Fatalf("Unable to read %s: %+v", filepath.Join(".", switchMetadata), err)
	}
	defer generated.Close()
	smGenerated := csv.NewReader(generated)
	actual, err := smGenerated.ReadAll()
	if err != nil {
		log.Fatalf("Unable to read %s: %+v", filepath.Join(".", switchMetadata), err)
	}
	// Read the csv and validate it's contents
	expected, err := os.Open(filepath.Join(".", switchMetadata))
	if err != nil {
		log.Fatalf("Unable to read %s: %+v", filepath.Join(".", switchMetadata), err)
	}
	defer expected.Close()
	csvReader := csv.NewReader(expected)
	smExpected, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("Unable to parse %q as a CSV: %+v", filepath.Join(".", switchMetadata), err)
	}
	assert.Equal(t, smExpected, actual)
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
		{
			desc:       "Common Name pdu-x3000-001 returns x3000p1",
			commonName: "pdu-x3000-001",
			want:       "x3000p1",
		},
		{
			desc:       "Common Name sw-hsn-001 returns sw-hsn01",
			commonName: "sw-hsn-001",
			want:       "sw-hsn01",
		},
		{
			desc:       "Common Name cn005 returns cn05",
			commonName: "cn005",
			want:       "cn05",
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

func TestFilterByTypeSwitch_ReturnsNoItemsIfNoSwitches(t *testing.T) {
	t.Parallel()

	want := []ID{}
	topology := []ID{
		{
			CommonName: "ncn-m001",
			Type:       "server",
		},
		{
			CommonName: "ncn-m002",
			Type:       "bogus",
		},
	}
	got := FilterByType(topology, "switch")
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFilterByTypeSwitch_ReturnsCorrectItems(t *testing.T) {
	t.Parallel()
	want := []ID{
		{
			CommonName: "sw-spine-001",
			Type:       "switch",
		},
		{
			CommonName: "sw-spine-002",
			Type:       "switch",
		},
		{
			CommonName: "sw-leaf-bmc-001",
			Type:       "switch",
		},
	}

	topology := []ID{
		{
			CommonName: "ncn-m001",
			Type:       "server",
		},
		{
			CommonName: "sw-spine-001",
			Type:       "switch",
		},
		{
			CommonName: "sw-spine-002",
			Type:       "switch",
		},
		{
			CommonName: "ncn-m002",
			Type:       "server",
		},
		{
			CommonName: "sw-leaf-bmc-001",
			Type:       "switch",
		},
	}

	got := FilterByType(topology, "switch")
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFilterByTypeServer_ReturnsNoItemsIfNoServers(t *testing.T) {
	t.Parallel()

	want := []ID{}
	topology := []ID{
		{
			CommonName: "sw-leaf-bmc-001",
			Type:       "switch",
		},
		{
			CommonName: "ncn-m002",
			Type:       "bogus",
		},
	}
	got := FilterByType(topology, "server")
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFilterByTypeServer_ReturnsCorrectItems(t *testing.T) {
	t.Parallel()
	want := []ID{
		{
			CommonName: "ncn-m002",
			Type:       "server",
		},
		{
			CommonName: "ncn-s005",
			Type:       "server",
		},
	}

	topology := []ID{
		{
			CommonName: "ncn-m002",
			Type:       "server",
		},
		{
			CommonName: "sw-spine-001",
			Type:       "switch",
		},
		{
			CommonName: "ncn-s005",
			Type:       "server",
		},
	}

	got := FilterByType(topology, "server")
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGenerateXNameGeneratesCorrectNameForCDUSwitch(t *testing.T) {
	t.Parallel()

	want := "d0w2"
	id := ID{
		Architecture: "mountain_compute_leaf",
		CommonName:   "sw-cdu-002",
		Type:         "switch",
		Vendor:       "aruba",
	}
	got := id.GenerateXname()
	if want != got {
		t.Fatalf("want xname %q got %q", want, got)
	}
}

func TestGenerateXNameGeneratesCorrectNameForSpineSwitch(t *testing.T) {
	t.Parallel()

	want := "x3000c0h38s1"
	id := ID{
		Architecture: "spine",
		CommonName:   "sw-spine-003",
		Location: Location{
			Elevation: "u38",
			Rack:      "x3000",
		},
		Model:  "8325_JL627A",
		Type:   "switch",
		Vendor: "aruba",
	}
	got := id.GenerateXname()
	if want != got {
		t.Fatalf("want xname %q got %q", want, got)
	}
}

func TestGenerateXNameGeneratesCorrectNameForRedbullSpineSwitch(t *testing.T) {
	t.Parallel()

	want := "x3000c0h19s2"
	id := ID{
		Architecture: "spine",
		CommonName:   "sw-spine-002",
		Location: Location{
			Elevation: "u19",
			Rack:      "x3000",
		},
		Model:  "8325_JL627A",
		Type:   "switch",
		Vendor: "mellanox",
	}
	got := id.GenerateXname()
	if want != got {
		t.Fatalf("want xname %q got %q", want, got)
	}
}
