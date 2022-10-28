/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

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

package cmd

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	"gopkg.in/yaml.v2"
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

func testProcessWithHill(t *testing.T) csi.CabinetDetailFile {
	var cabDetailFile csi.CabinetDetailFile

	err := yaml.Unmarshal([]byte(cabinetsWithHill), &cabDetailFile)
	if err != nil {
		t.Error(err)
	}
	return cabDetailFile
}

func testFindCabinetGroupDetail(t *testing.T, cabinetDetailList []csi.CabinetGroupDetail, wantedKind csi.CabinetKind) *csi.CabinetGroupDetail {
	for _, cabinetDetail := range cabinetDetailList {
		if cabinetDetail.Kind == wantedKind {
			return &cabinetDetail
		}
	}

	t.Errorf("Failed to kind %s in cabinetDetailList", wantedKind)
	return nil
}

func TestCabinetDefinitionWithHill(t *testing.T) {
	cabDetailFile := testProcessWithHill(t)
	cabDefinitions := make(map[csi.CabinetKind]cabinetDefinition)
	cabDefinitions[csi.CabinetKindHill] = cabinetDefinition{
		count:      20,
		startingID: 100,
	}
	cabDefinitions[csi.CabinetKindMountain] = cabinetDefinition{
		count:      20,
		startingID: 200,
	}
	cabDefinitions[csi.CabinetKindRiver] = cabinetDefinition{
		count:      20,
		startingID: 300,
	}

	cabinetDetailList := buildCabinetDetails(cabDefinitions, cabDetailFile)
	if len(cabinetDetailList) != len(csi.ValidCabinetTypes) {
		t.Errorf("%+v", cabinetDetailList)
	}

	// Hill
	hill := testFindCabinetGroupDetail(t, cabinetDetailList, csi.CabinetKindHill)
	if hill.Cabinets != 1 {
		t.Errorf("Expected 1 hill cabinet, but got %v", hill.Cabinets)
	}

	// Mountain
	mountain := testFindCabinetGroupDetail(t, cabinetDetailList, csi.CabinetKindMountain)
	if mountain.Cabinets != 20 {
		t.Errorf("Expected 20 mountain cabinets, but got %v", mountain.Cabinets)
	}

	// River
	river := testFindCabinetGroupDetail(t, cabinetDetailList, csi.CabinetKindRiver)
	if river.Cabinets != 1 {
		t.Errorf("Expected 1 river cabinet, but got %v", river.Cabinets)
	}

	// The EX{20,25,30,40}00 cabinet kinds are expected to be 0
	for _, cabinetKind := range []csi.CabinetKind{csi.CabinetKindEX2000, csi.CabinetKindEX2500, csi.CabinetKindEX3000, csi.CabinetKindEX4000} {
		cabinetGroupDetail := testFindCabinetGroupDetail(t, cabinetDetailList, cabinetKind)
		if cabinetGroupDetail.Cabinets != 0 {
			t.Errorf("Expected 0 %v cabinets, but got %v", cabinetKind, cabinetGroupDetail.Cabinets)
		}
	}

	// Make the cabinet detail list a bit more readable
	cabinetDetailListJSON, err := json.MarshalIndent(cabinetDetailList, "", "  ")
	if err != nil {
		panic(err)
	}

	log.Printf("%+v \n", string(cabinetDetailListJSON))
}
