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

package initialize

import (
	"encoding/json"
	"log"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/sls"
)

var cabinetsWithHill = `
cabinets:
- type: river
  cabinets:
    - id: 3000

- type: hill
  cabinets:
    - id: 9000
      subnet: 10.104.0.1/22
      vlan: 2000
`

func testProcessWithHill(t *testing.T) sls.CabinetDetailFile {
	var cabDetailFile sls.CabinetDetailFile

	err := yaml.Unmarshal(
		[]byte(cabinetsWithHill),
		&cabDetailFile,
	)
	if err != nil {
		t.Error(err)
	}
	return cabDetailFile
}

func testFindCabinetGroupDetail(
	t *testing.T, cabinetDetailList []sls.CabinetGroupDetail, wantedKind sls.CabinetKind,
) *sls.CabinetGroupDetail {
	for _, cabinetDetail := range cabinetDetailList {
		if cabinetDetail.Kind == wantedKind {
			return &cabinetDetail
		}
	}

	t.Errorf(
		"Failed to kind %s in cabinetDetailList",
		wantedKind,
	)
	return nil
}

func TestCabinetDefinitionWithHill(t *testing.T) {
	cabDetailFile := testProcessWithHill(t)
	cabDefinitions := make(map[sls.CabinetKind]cabinetDefinition)
	cabDefinitions[sls.CabinetKindHill] = cabinetDefinition{
		count:      20,
		startingID: 100,
	}
	cabDefinitions[sls.CabinetKindMountain] = cabinetDefinition{
		count:      20,
		startingID: 200,
	}
	cabDefinitions[sls.CabinetKindRiver] = cabinetDefinition{
		count:      20,
		startingID: 300,
	}

	cabinetDetailList := buildCabinetDetails(
		cabDefinitions,
		cabDetailFile,
	)
	if len(cabinetDetailList) != len(sls.ValidCabinetTypes) {
		t.Errorf(
			"%+v",
			cabinetDetailList,
		)
	}

	// Hill
	hill := testFindCabinetGroupDetail(
		t,
		cabinetDetailList,
		sls.CabinetKindHill,
	)
	if hill.Cabinets != 1 {
		t.Errorf(
			"Expected 1 hill cabinet, but got %v",
			hill.Cabinets,
		)
	}

	// Mountain
	mountain := testFindCabinetGroupDetail(
		t,
		cabinetDetailList,
		sls.CabinetKindMountain,
	)
	if mountain.Cabinets != 20 {
		t.Errorf(
			"Expected 20 mountain cabinets, but got %v",
			mountain.Cabinets,
		)
	}

	// River
	river := testFindCabinetGroupDetail(
		t,
		cabinetDetailList,
		sls.CabinetKindRiver,
	)
	if river.Cabinets != 1 {
		t.Errorf(
			"Expected 1 river cabinet, but got %v",
			river.Cabinets,
		)
	}

	// The EX{20,25,30,40}00 cabinet kinds are expected to be 0
	for _, cabinetKind := range []sls.CabinetKind{
		sls.CabinetKindEX2000,
		sls.CabinetKindEX2500,
		sls.CabinetKindEX3000,
		sls.CabinetKindEX4000,
	} {
		cabinetGroupDetail := testFindCabinetGroupDetail(
			t,
			cabinetDetailList,
			cabinetKind,
		)
		if cabinetGroupDetail.Cabinets != 0 {
			t.Errorf(
				"Expected 0 %v cabinets, but got %v",
				cabinetKind,
				cabinetGroupDetail.Cabinets,
			)
		}
	}

	// Make the cabinet detail list a bit more readable
	cabinetDetailListJSON, err := json.MarshalIndent(
		cabinetDetailList,
		"",
		"  ",
	)
	if err != nil {
		panic(err)
	}

	log.Printf(
		"%+v \n",
		string(cabinetDetailListJSON),
	)
}
