//
//  MIT License
//
//  (C) Copyright 2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//
//go:build !integration && !shcd
// +build !integration,!shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"
	"testing"

	"github.com/Cray-HPE/csm-common/go/pkg/csi"
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

func TestCabinetDefinitionWithHill(t *testing.T) {
	cabDetailFile := testProcessWithHill(t)
	cabDefinitions := make(map[string]cabinetDefinition)
	cabDefinitions["hill"] = cabinetDefinition{
		count:      20,
		startingID: 100,
	}
	cabDefinitions["mountain"] = cabinetDefinition{
		count:      20,
		startingID: 200,
	}
	cabDefinitions["river"] = cabinetDefinition{
		count:      20,
		startingID: 300,
	}

	cabinetDetailList := buildCabinetDetails(cabDefinitions, cabDetailFile)
	if len(cabinetDetailList) != 3 {
		t.Errorf("%+v", cabinetDetailList)
	}
	log.Printf("%+v \n", cabinetDetailList)
}
