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
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/Cray-HPE/cray-site-init/pkg/shcd"
	"github.com/google/go-cmp/cmp"
)

func TestCreateHMNConnections(t *testing.T) {
	t.Parallel()
	shcdFile, err := ioutil.ReadFile("../testdata/fixtures/valid_shcd.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	shcd, err := shcd.ParseSHCD(shcdFile)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Create hmn_connections.json
	err = createHMNSeed(shcd.Topology)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Validate the file was created
	_, err = os.Stat(filepath.Join(".", hmnConnections))
	if err != nil {
		t.Fatal(err)
	}
	// Read the generated json and validate it's contents
	hmnGenerated, err := os.Open(filepath.Join(".", hmnConnections))
	if err != nil {
		t.Fatal(err)
	}
	defer hmnGenerated.Close()
	hmnExpected, err := os.Open("../testdata/expected/" + hmnConnections)
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
	if !cmp.Equal(string(hmnConnections), string(hmnActual)) {
		t.Fatal(cmp.Diff(string(hmnConnections), string(hmnActual)))
	}
}

func TestCreateSwitchMetadata(t *testing.T) {
	t.Parallel()
	jsonFilePath := "../testdata/fixtures/valid_shcd.json"
	// Open the file without validating it since we know it is valid
	shcdFile, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatal(err)
	}
	shcd, err := shcd.ParseSHCD(shcdFile)
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
	if !cmp.Equal(smExpected, actual) {
		t.Fatal(cmp.Diff(smExpected, actual))
	}
}

func TestCreateNCNMetadata(t *testing.T) {
	t.Parallel()
	// Open the file without validating it since we know it is valid
	shcdFile, err := ioutil.ReadFile("../testdata/fixtures/valid_shcd.json")
	if err != nil {
		t.Fatal(err.Error())
	}
	shcd, err := shcd.ParseSHCD(shcdFile)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = createNCNSeed(shcd.Topology)
	if err != nil {
		t.Fatal(err.Error())
	}
	// Validate the file was created
	_, err = os.Stat(filepath.Join(".", ncnMetadata))
	if err != nil {
		t.Fatal(err)
	}
	// Read the csv and validate it's contents
	generated, err := os.Open(filepath.Join(".", ncnMetadata))
	if err != nil {
		t.Fatalf("Unable to read %q: %+v\n", filepath.Join(".", ncnMetadata), err)
	}
	defer generated.Close()
	ncnGenerated := csv.NewReader(generated)
	actual, err := ncnGenerated.ReadAll()
	if err != nil {
		t.Fatalf("Unable to parse file %q as a CSV: %+v", filepath.Join(".", ncnMetadata), err)
	}
	// Read the csv and validate it's contents
	expected, err := os.Open(filepath.Join(".", ncnMetadata))
	if err != nil {
		t.Fatalf("Unable to read %q: %+v", filepath.Join(".", ncnMetadata), err)
	}
	defer expected.Close()
	csvReader := csv.NewReader(expected)
	ncnExpected, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Unable to parse file %q as a CSV: %+v", "../testdata/expected/"+ncnMetadata, err)
	}
	if !cmp.Equal(ncnExpected, actual) {
		t.Fatal(cmp.Diff(ncnExpected, actual))
	}
}

func TestCreateApplicationNodeConfig(t *testing.T) {
	t.Parallel()
	// Open the file without validating it since we know it is valid
	shcdFile, err := ioutil.ReadFile("../testdata/fixtures/odin_valid_ccj.json")
	if err != nil {
		t.Fatal(err.Error())
	}
	shcd, err := shcd.ParseSHCD(shcdFile)
	if err != nil {
		t.Fatal(err.Error())
	}
	prefixSubroleMapIn = map[string]string{
		"gateway": "Gateway",
		"login":   "UAN",
		"lnet":    "LNETRouter",
		"vn":      "Visualization",
	}
	// Create application_node_config.yaml
	err = createANCSeed(shcd.Topology)
	if err != nil {
		t.Fatal(err)
	}
	// Validate the file was created
	_, err = os.Stat(filepath.Join(".", applicationNodeConfig))
	if err != nil {
		t.Fatal(err)
	}
	// Read the yaml and validate it's contents
	ancGenerated, err := os.Open(filepath.Join(".", applicationNodeConfig))
	if err != nil {
		t.Fatalf("Unable to read %q: %+v", filepath.Join(".", applicationNodeConfig), err)
	}
	defer ancGenerated.Close()
	ancExpected, err := os.Open("../testdata/expected/" + applicationNodeConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer ancExpected.Close()
	ancActual, err := ioutil.ReadAll(ancGenerated)
	if err != nil {
		t.Fatal(err)
	}
	appNodeConfig, err := ioutil.ReadAll(ancExpected)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(string(appNodeConfig), string(ancActual)) {
		t.Fatal(cmp.Diff(string(appNodeConfig), string(ancActual)))
	}
}

func TestGenerateSourceName(t *testing.T) {
	t.Parallel()
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
		{
			desc:       "Common Name gateway001 returns gateway01",
			commonName: "gateway001",
			want:       "gateway01",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ID := shcd.ID{CommonName: tC.commonName}
			got := ID.GenerateSourceName()
			if tC.want != got {
				t.Errorf("want common name %q, got %q", tC.want, got)
			}
		})
	}
}

func TestFilterByTypeSwitch_ReturnsNoItemsIfNoSwitches(t *testing.T) {
	t.Parallel()
	want := []shcd.ID{}
	topology := []shcd.ID{
		{
			CommonName: "ncn-m001",
			Type:       "server",
		},
		{
			CommonName: "ncn-m002",
			Type:       "bogus",
		},
	}
	got := shcd.FilterByType(topology, "switch")
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFilterByTypeSwitch_ReturnsCorrectItems(t *testing.T) {
	t.Parallel()
	want := []shcd.ID{
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

	topology := []shcd.ID{
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

	got := shcd.FilterByType(topology, "switch")
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFilterByTypeServer_ReturnsNoItemsIfNoServers(t *testing.T) {
	t.Parallel()
	want := []shcd.ID{}
	topology := []shcd.ID{
		{
			CommonName: "sw-leaf-bmc-001",
			Type:       "switch",
		},
		{
			CommonName: "ncn-m002",
			Type:       "bogus",
		},
	}
	got := shcd.FilterByType(topology, "server")
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFilterByTypeServer_ReturnsCorrectItems(t *testing.T) {
	t.Parallel()
	want := []shcd.ID{
		{
			CommonName: "ncn-m002",
			Type:       "server",
		},
		{
			CommonName: "ncn-s005",
			Type:       "server",
		},
	}

	topology := []shcd.ID{
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

	got := shcd.FilterByType(topology, "server")
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGenerateXNameGeneratesCorrectNameForCDUSwitch(t *testing.T) {
	t.Parallel()
	want := "d0w2"
	id := shcd.ID{
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
	id := shcd.ID{
		Architecture: "spine",
		CommonName:   "sw-spine-003",
		Location: shcd.Location{
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
	id := shcd.ID{
		Architecture: "spine",
		CommonName:   "sw-spine-002",
		Location: shcd.Location{
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

func TestNewShcdErrorsIfFileDoesNotExist(t *testing.T) {
	t.Parallel()
	_, err := shcd.NewShcd("bogus-file-name-bogus.bogus-bogus")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("want error os.ErrNotExist got %+v", err)
	}
}

func TestNewShcdErrorsIfShcdFileSyntaxIsBroken(t *testing.T) {
	t.Parallel()
	// we are testing that the syntax fails, so we must be sure the file
	// exists or this is going to be the erro and we don't get into the syntax validation
	_, err := os.Stat("../testdata/fixtures/invalid_shcd.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = shcd.NewShcd("../testdata/fixtures/invalid_shcd.json")
	if err == nil {
		t.Fatal("want error to be not nil")
	}
}

func TestNewShcdErrorsIfShcdFileHasWrongDataType(t *testing.T) {
	t.Parallel()
	// we are testing that the syntax fails, so we must be sure the file
	// exists or this is going to be the erro and we don't get into the syntax validation
	_, err := os.Stat("../testdata/fixtures/invalid_data_types_shcd.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = shcd.NewShcd("../testdata/fixtures/invalid_data_types_shcd.json")
	if err == nil {
		t.Fatal("want error to be not nil")
	}
}

func TestNewShcdReturnsCorrectShcdData(t *testing.T) {
	t.Parallel()
	want := &shcd.Shcd{
		Topology: []shcd.ID{{
			Architecture: "river_ncn_node_4_port",
			CommonName:   "ncn-m001",
			ID:           31,
			Location: shcd.Location{
				Elevation: "u01",
				Rack:      "x3000",
			},
			Model: "river_ncn_node_4_port",
			Ports: []shcd.Port{
				{
					DestNodeID: 17,
					DestPort:   1,
					Port:       1,
					Slot:       "pcie-slot1",
					Speed:      25,
				}},
			Type:   "server",
			Vendor: "hpe",
		}},
	}
	got, err := shcd.NewShcd("../testdata/fixtures/simple_shcd.json")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}
