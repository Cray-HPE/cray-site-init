/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"log"
	"testing"

	"gopkg.in/yaml.v2"
	"github.com/Cray-HPE/cray-site-init/pkg/csi"
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
