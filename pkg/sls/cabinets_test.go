/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

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

package sls

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
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
			Kind: CabinetKind(
				fmt.Sprintf(
					"kind_%v",
					i,
				),
			),
			Cabinets:        numCabinets,
			StartingCabinet: startingCabinet,
		}
		tmpCab.PopulateIds()
		cabinetFile.Cabinets = append(
			cabinetFile.Cabinets,
			tmpCab,
		)
	}
	return cabinetFile
}

func TestMarshalCabinetsFile(t *testing.T) {
	cabinetFile := genRandomCabinetDetailStruct()
	doc, err := yaml.Marshal(cabinetFile)
	if err != nil {
		log.Fatalln(
			"Unable to Marshal",
			cabinetFile,
		)
	}
	t.Logf(
		"%v",
		string(doc[:]),
	)
}

func TestUnMarshalCabinetsFile(t *testing.T) {

	var cabinetFile CabinetDetailFile
	err := yaml.Unmarshal(
		fakeYaml,
		&cabinetFile,
	)
	if err != nil {
		log.Fatalln(
			"Unable to Unmarshal the fake Yaml",
			err,
		)
	}
}

type CabinetFilterFuncTestSuite struct {
	suite.Suite

	groupDetails map[CabinetKind]CabinetGroupDetail

	riverCabinetDetail                          CabinetDetail
	hillCabinetDetail                           CabinetDetail
	hillEX2500CabinetDetail                     CabinetDetail
	hillEX2500CabinetWithAirCooledChassisDetail CabinetDetail
	mountainCabinetDetail                       CabinetDetail
}

func (suite *CabinetFilterFuncTestSuite) SetupSuite() {
	cabinetKinds := []CabinetKind{
		CabinetKindRiver,
		CabinetKindHill,
		CabinetKindMountain,
		CabinetKindEX2000,
		CabinetKindEX2500,
		CabinetKindEX3000,
		CabinetKindEX4000,
	}

	suite.groupDetails = map[CabinetKind]CabinetGroupDetail{}
	for _, cabinetKind := range cabinetKinds {
		suite.groupDetails[cabinetKind] = CabinetGroupDetail{Kind: cabinetKind}
	}

	suite.riverCabinetDetail = CabinetDetail{ID: 3000}
	suite.hillCabinetDetail = CabinetDetail{ID: 9000}
	suite.hillEX2500CabinetDetail = CabinetDetail{
		ID: 9001,
		ChassisCount: &ChassisCount{
			LiquidCooled: 3,
			AirCooled:    0,
		},
	}
	suite.hillEX2500CabinetWithAirCooledChassisDetail = CabinetDetail{
		ID: 9002,
		ChassisCount: &ChassisCount{
			LiquidCooled: 1,
			AirCooled:    1,
		},
	}
	suite.mountainCabinetDetail = CabinetDetail{ID: 1000}

}

func (suite *CabinetFilterFuncTestSuite) TestCabinetKindSelector_River() {
	cabinetFilter := CabinetKindFilter(CabinetKindRiver)

	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func (suite *CabinetFilterFuncTestSuite) TestCabinetKindSelector_Hill() {
	cabinetFilter := CabinetKindFilter(CabinetKindHill)

	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func (suite *CabinetFilterFuncTestSuite) TestCabinetKindSelector_Mountain() {
	cabinetFilter := CabinetKindFilter(CabinetKindMountain)

	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func (suite *CabinetFilterFuncTestSuite) TestCabinetClassFilter_River() {
	cabinetFilter := CabinetClassFilter(slsCommon.ClassRiver)

	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func (suite *CabinetFilterFuncTestSuite) TestCabinetClassFilter_Hill() {
	cabinetFilter := CabinetClassFilter(slsCommon.ClassHill)

	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func (suite *CabinetFilterFuncTestSuite) TestCabinetClassFilter_Mountain() {
	cabinetFilter := CabinetClassFilter(slsCommon.ClassMountain)

	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func (suite *CabinetFilterFuncTestSuite) TestCabinetChassisCountsFilter() {
	cabinetFilter := CabinetAirCooledChassisCountFilter(1)

	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func (suite *CabinetFilterFuncTestSuite) TestAndCabinetFilter_EX2500CabinetWithAirCooledChassis() {
	cabinetFilter := AndCabinetFilter(
		CabinetKindFilter(CabinetKindEX2500),
		CabinetAirCooledChassisCountFilter(1),
	)

	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func (suite *CabinetFilterFuncTestSuite) TestAndCabinetFilter_EX2500CabinetWith3LiquidCooledChassis() {
	cabinetFilter := AndCabinetFilter(
		CabinetKindFilter(CabinetKindEX2500),
		CabinetAirCooledChassisCountFilter(0),
		CabinetLiquidCooledChassisCountFilter(3),
	)

	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func (suite *CabinetFilterFuncTestSuite) TEstOrCabinetFilter() {
	cabinetFilter := OrCabinetFilter(
		// Standard River Cabinet
		CabinetClassFilter(slsCommon.ClassRiver),

		// Or the special case where special case for EX2500 cabinets with both liquid and air cooled chassis
		AndCabinetFilter(
			CabinetKindFilter(CabinetKindEX2500),
			CabinetAirCooledChassisCountFilter(1),
		),
	)

	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindRiver],
			suite.riverCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindHill],
			suite.hillCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindMountain],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetDetail,
		),
	)
	suite.True(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX2500],
			suite.hillEX2500CabinetWithAirCooledChassisDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX3000],
			suite.mountainCabinetDetail,
		),
	)
	suite.False(
		cabinetFilter(
			suite.groupDetails[CabinetKindEX4000],
			suite.mountainCabinetDetail,
		),
	)
}

func TestCabinetFilter(t *testing.T) {
	suite.Run(
		t,
		new(CabinetFilterFuncTestSuite),
	)
}

type CabinetGroupDetailTestSuite struct {
	suite.Suite
}

func (suite *CabinetGroupDetailTestSuite) TestCabinetClassRiver() {
	cgd := CabinetGroupDetail{Kind: CabinetKindRiver}

	class, err := cgd.Kind.Class()
	suite.NoError(err)
	suite.Equal(
		slsCommon.ClassRiver,
		class,
	)
}

func (suite *CabinetGroupDetailTestSuite) TestCabinetClassHill() {
	kinds := []CabinetKind{
		CabinetKindHill,
		CabinetKindEX2000,
		CabinetKindEX2500,
	}

	for _, kind := range kinds {
		cgd := CabinetGroupDetail{Kind: kind}

		class, err := cgd.Kind.Class()
		suite.NoError(err)
		suite.Equal(
			slsCommon.ClassHill,
			class,
		)
	}
}

func (suite *CabinetGroupDetailTestSuite) TestCabinetClassMountain() {
	kinds := []CabinetKind{
		CabinetKindMountain,
		CabinetKindEX3000,
		CabinetKindEX4000,
	}

	for _, kind := range kinds {
		cgd := CabinetGroupDetail{Kind: kind}

		class, err := cgd.Kind.Class()
		suite.NoError(err)
		suite.Equal(
			slsCommon.ClassMountain,
			class,
		)
	}
}

func (suite *CabinetGroupDetailTestSuite) TestCabinetClassInvalidKind() {
	cgd := CabinetGroupDetail{Kind: "foobar"}

	_, err := cgd.Kind.Class()
	suite.EqualError(
		err,
		"unknown cabinet kind (foobar)",
	)
}

func TestCabinetGroupDetail(t *testing.T) {
	suite.Run(
		t,
		new(CabinetGroupDetailTestSuite),
	)
}
