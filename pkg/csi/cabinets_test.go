//go:build !integration && !shcd
// +build !integration,!shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

var fakeYaml = []byte(`
cabinets:
    - type: kind_1
      total_number: 3
      starting_id: 9
      ids:
      - 9
      - 10
      - 11
    - type: kind_2
      total_number: 10
      starting_id: 58
      ids:
      - 58
      - 59
      - 60
      - 61
      - 62
      - 63
      - 64
      - 65
      - 66
      - 67
    - type: kind_3
      total_number: 6
      starting_id: 174
      ids:
      - 174
      - 175
      - 176
      - 177
      - 178
      - 179
    - type: kind_4
      total_number: 3
      starting_id: 183
      ids:
      - 183
      - 184
      - 185
`)

func genRandomCabinetDetailStruct() CabinetDetailFile {
	var cabinetFile CabinetDetailFile
	rand.Seed(12)
	for i := 1; i < 5; i++ {
		startingCabinet := rand.Intn(200)
		numCabinets := rand.Intn(20)
		tmpCab := CabinetGroupDetail{
			Kind:            fmt.Sprintf("kind_%v", i),
			Cabinets:        numCabinets,
			StartingCabinet: startingCabinet,
		}
		tmpCab.PopulateIds()
		cabinetFile.Cabinets = append(cabinetFile.Cabinets, tmpCab)
	}
	return cabinetFile
}

func TestMarshalCabinetsFile(t *testing.T) {
	cabinetFile := genRandomCabinetDetailStruct()
	doc, err := yaml.Marshal(cabinetFile)
	if err != nil {
		log.Fatalln("Unable to Marshal", cabinetFile)
	}
	t.Logf("%v", string(doc[:]))
}

func TestUnMarshalCabinetsFile(t *testing.T) {

	var cabinetFile CabinetDetailFile
	err := yaml.Unmarshal(fakeYaml, &cabinetFile)
	if err != nil {
		log.Fatalln("Unable to Unmarshal the fake Yaml", err)
	}
}

type CabinetFilterFuncTestSuite struct {
	suite.Suite

	riverGroupDetail    CabinetGroupDetail
	hillGroupDetail     CabinetGroupDetail
	mountainGroupDetail CabinetGroupDetail

	riverCabinetDetail                          CabinetDetail
	hillCabinetDetail                           CabinetDetail
	hillEX2500CabinetDetail                     CabinetDetail
	hillEX2500CabinetWithAirCooledChassisDetail CabinetDetail
	mountainCabinetDetail                       CabinetDetail
}

func (suite *CabinetFilterFunc) SetupSuite() {
	suite.riverGroupDetail = CabinetGroupDetail{Kind: "river"}
	suite.hillGroupDetail = CabinetGroupDetail{Kind: "hill"}
	suite.mountainGroupDetail = CabinetGroupDetail{Kind: "mountain"}

	suite.riverCabinetDetail = CabinetDetail{ID: 3000}
	suite.hillCabinetDetail = CabinetDetail{ID: 9000}
	suite.hillEX2500CabinetDetail = CabinetDetail{
		ID:    9001,
		Model: "EX2500",
		ChassisCount: &ChassisCount{
			LiquidCooled: 3,
			AirCooled:    0,
		},
	}
	suite.hillEX2500CabinetWithAirCooledChassisDetail = CabinetDetail{
		ID:    9002,
		Model: "EX2500",
		ChassisCount: &ChassisCount{
			LiquidCooled: 1,
			AirCooled:    1,
		},
	}
	suite.mountainCabinetDetail = CabinetDetail{ID: 1000}

}

func (suite *CabinetFilterFuncTestSuite) TestCabinetKindSelector_River() {
	cabinetSelector := CabinetKindFilter("river")

	suite.True(cabinetSelector(suite.riverGroupDetail, suite.riverCabinetDetail))
	suite.False(cabinetSelector(suite.hillGroupDetail, suite.hillCabinetDetail))
	suite.False(cabinetSelector(suite.hillGroupDetail, suite.hillEX2500CabinetDetail))
	suite.False(cabinetSelector(suite.hillGroupDetail, suite.hillEX2500CabinetWithAirCooledChassisDetail))
	suite.False(cabinetSelector(suite.mountainGroupDetail, suite.mountainCabinetDetail))
}

func (suite *CabinetFilterFuncTestSuite) TestCabinetKindSelector_Hill() {
	cabinetSelector := CabinetKindFilter("hill")

	suite.False(cabinetSelector(suite.riverGroupDetail, suite.riverCabinetDetail))
	suite.True(cabinetSelector(suite.hillGroupDetail, suite.hillCabinetDetail))
	suite.True(cabinetSelector(suite.hillGroupDetail, suite.hillEX2500CabinetDetail))
	suite.True(cabinetSelector(suite.hillGroupDetail, suite.hillEX2500CabinetWithAirCooledChassisDetail))
	suite.False(cabinetSelector(suite.mountainGroupDetail, suite.mountainCabinetDetail))
}

func (suite *CabinetFilterFuncTestSuite) TestCabinetKindSelector_Mountain() {
	cabinetSelector := CabinetKindFilter("mountain")

	suite.False(cabinetSelector(suite.riverGroupDetail, suite.riverCabinetDetail))
	suite.False(cabinetSelector(suite.hillGroupDetail, suite.hillCabinetDetail))
	suite.False(cabinetSelector(suite.hillGroupDetail, suite.hillEX2500CabinetDetail))
	suite.False(cabinetSelector(suite.hillGroupDetail, suite.hillEX2500CabinetWithAirCooledChassisDetail))
	suite.True(cabinetSelector(suite.mountainGroupDetail, suite.mountainCabinetDetail))
}

func (suite *CabinetFilterFuncTestSuite) TestCabinetEX2500AirCooledChassisSelector_Hill_EX2500() {
	cabinetSelector := CabinetEX2500AirCooledChassisFilter()

	suite.False(cabinetSelector(suite.riverGroupDetail, suite.riverCabinetDetail))
	suite.False(cabinetSelector(suite.hillGroupDetail, suite.hillCabinetDetail))
	suite.False(cabinetSelector(suite.hillGroupDetail, suite.hillEX2500CabinetDetail))
	suite.True(cabinetSelector(suite.hillGroupDetail, suite.hillEX2500CabinetWithAirCooledChassisDetail))
	suite.False(cabinetSelector(suite.mountainGroupDetail, suite.mountainCabinetDetail))
}

func TestCabinetFilter(t *testing.T) {
	suite.Run(t, new(CabinetFilterFuncTestSuite))
}
