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

package initialize

import (
	"fmt"
	"log"
	"path/filepath"

	shcdParser "github.com/Cray-HPE/hms-shcd-parser/pkg/shcd-parser"
	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/internal/files"
	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	slsInit "github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
	"github.com/Cray-HPE/cray-site-init/pkg/sls"
)

type cabinetDefinition struct {
	count      int
	startingID int
}

func collectHMNRows(v *viper.Viper) []shcdParser.HMNRow {
	seedFileHmnConnections := filepath.Dir(viper.ConfigFileUsed()) + "/" + v.GetString("hmn-connections")
	hmnRows, err := loadHMNConnectionsFile(seedFileHmnConnections)
	if err != nil {
		log.Fatalf(
			"unable to load hmn connections, %v \n",
			err,
		)
	}
	return hmnRows
}

func collectNCNMeta(v *viper.Viper) []*LogicalNCN {
	seedFileNcnMetadata := filepath.Dir(viper.ConfigFileUsed()) + "/" + v.GetString("ncn-metadata")
	ncns, err := ReadNodeCSV(seedFileNcnMetadata)
	if err != nil {
		log.Fatalln(
			"Couldn't extract ncns",
			err,
		)
	}

	// Normalize the ncn data, before validation
	for _, ncn := range ncns {
		ncn.Normalize()
	}

	if err := validateNCNInput(ncns); err != nil {
		log.Println("Unable to get reasonable NCNs from your csv")
		log.Println("Does your header match the preferred style? Xname,Role,Subrole,BMC MAC,Bootstrap MAC,Bond0 MAC0,Bond0 MAC1")
		log.Fatal("CSV Parsing failed. Can't continue.")
	}

	return ncns
}

func collectSwitches(v *viper.Viper) []*networking.ManagementSwitch {
	seedFileSwitchMetadata := filepath.Dir(v.ConfigFileUsed()) + "/" + v.GetString("switch-metadata")

	switches, err := networking.ReadSwitchCSV(seedFileSwitchMetadata)
	if err != nil {
		log.Fatalf(
			"Couldn't extract switches, %v",
			err,
		)
	}

	// Normalize the management switch data, before validation
	for _, mySwitch := range switches {
		mySwitch.Normalize()
	}

	if err := validateSwitchInput(switches); err != nil {
		log.Println("Unable to get reasonable Switches from your csv")
		log.Println("Does your header match the preferred style? Switch Xname,Type,Brand")
		log.Fatal("CSV Parsing failed. Can't continue.")
	}

	return switches
}

func collectApplicationNodeConfig(v *viper.Viper) slsInit.GeneratorApplicationNodeConfig {
	var applicationNodeConfig slsInit.GeneratorApplicationNodeConfig
	if v.IsSet("application-node-config-yaml") && (v.GetString("application-node-config-yaml") != "") {
		seedFileAppNodeConfig := filepath.Dir(viper.ConfigFileUsed()) + "/" + v.GetString("application-node-config-yaml")

		log.Printf(
			"Using application node config: %s\n",
			seedFileAppNodeConfig,
		)
		err := files.ReadYAMLConfig(
			seedFileAppNodeConfig,
			&applicationNodeConfig,
		)
		if err != nil {
			log.Fatalf(
				"Unable to parse application-node-config file: %s\nError: %v",
				seedFileAppNodeConfig,
				err,
			)
		}
	}

	// Normalize application node config
	if err := applicationNodeConfig.Normalize(); err != nil {
		log.Fatalf(
			"Failed to normalize application node config. Error: %s",
			err,
		)
	}

	// Validate Application node config
	if err := applicationNodeConfig.Validate(); err != nil {
		log.Fatalf(
			"Failed to validate application node config. Error: %s",
			err,
		)
	}

	return applicationNodeConfig
}

func collectCabinets(v *viper.Viper) []sls.CabinetGroupDetail {
	var cabDetailFile sls.CabinetDetailFile
	if v.IsSet("cabinets-yaml") && (v.GetString("cabinets-yaml") != "") {
		seedFileCabinets := filepath.Dir(viper.ConfigFileUsed()) + "/" + v.GetString("cabinets-yaml")
		err := files.ReadYAMLConfig(
			seedFileCabinets,
			&cabDetailFile,
		)
		if err != nil {
			log.Fatalf(
				"Unable to parse cabinets-yaml file: %s\nError: %v",
				v.GetString("cabinets-yaml"),
				err,
			)
		}
	}

	cabDefinitions := make(map[sls.CabinetKind]cabinetDefinition)
	for _, cabType := range sls.ValidCabinetTypes {
		cabDefinitions[cabType] = cabinetDefinition{
			count: v.GetInt(
				fmt.Sprintf(
					"%s-cabinets",
					cabType,
				),
			),
			startingID: v.GetInt(
				fmt.Sprintf(
					"starting-%s-cabinet",
					cabType,
				),
			),
		}
	}
	cabinetDetailList := buildCabinetDetails(
		cabDefinitions,
		cabDetailFile,
	)

	// Verify no duplicate cabinet ids are present
	knownCabinetIDs := map[int]bool{}
	for _, cabinetGroupDetail := range cabinetDetailList {
		for _, id := range cabinetGroupDetail.CabinetIDs() {
			if knownCabinetIDs[id] {
				log.Fatalf(
					"Found duplicate cabinet id: %v",
					id,
				)
			}

			knownCabinetIDs[id] = true
		}
	}

	return cabinetDetailList
}

func collectInput(v *viper.Viper) (
	[]shcdParser.HMNRow, []*LogicalNCN, []*networking.ManagementSwitch, slsInit.GeneratorApplicationNodeConfig,
	[]sls.CabinetGroupDetail,
) {
	// The installation requires a set of information in order to proceed
	// First, we need some kind of representation of the physical hardware
	// That is generally represented through the hmn_connections.json file
	// which is literally a cabling map with metadata about the NCNs and
	// River Compute node BMCs, Columbia Rosetta Switches, and PDUs.
	//
	// From the hmn_connections file, we can create a set of HMNRow objects
	// to use for populating
	// this seedfile should be in the same place as the config, so use that to craft the path
	hmnRows := collectHMNRows(v)

	// This is technically sufficient to generate an SLSState object, but to do so now
	// would not include extended information about the NCNs and Network Switches.
	//
	// The first step in building the NCN map is to read the NCN Metadata file
	ncns := collectNCNMeta(v)

	// SLS also needs to know about our networking configuration. In order to do that,
	// we need to load the switches
	switches := collectSwitches(v)

	// Application Node configuration for SLS Config Generator
	// This is an optional input file
	applicationNodeConfig := collectApplicationNodeConfig(v)

	// Cabinet Map Configuration
	// This is an optional input file
	cabinetDetailList := collectCabinets(v)

	return hmnRows, ncns, switches, applicationNodeConfig, cabinetDetailList
}

func validateSwitchInput(switches []*networking.ManagementSwitch) error {
	// Validate that there is an non-zero number of NCNs extracted from ncn_metadata.csv
	if len(switches) == 0 {
		return fmt.Errorf("unable to extract Switches from switch metadata csv")
	}

	// Validate each Switch
	var mustFail = false
	for _, mySwitch := range switches {
		if err := mySwitch.Validate(); err != nil {
			mustFail = true
			log.Println(
				"Switch from csv is invalid:",
				err,
			)
		}
	}

	if mustFail {
		return fmt.Errorf("switch_metadata.csv contains invalid switch data")
	}

	return nil
}

func validateNCNInput(ncns []*LogicalNCN) error {
	// Validate that there is an non-zero number of NCNs extracted from ncn_metadata.csv
	if len(ncns) == 0 {
		return fmt.Errorf("unable to extract NCNs from ncn metadata csv")
	}

	// Validate each NCN
	var mustFail = false
	for _, ncn := range ncns {
		if err := ncn.Validate(); err != nil {
			mustFail = true
			log.Println(
				"NCN from csv is invalid",
				ncn,
				err,
			)
		}
	}

	if mustFail {
		return fmt.Errorf("ncn_metadata.csv contains invalid NCN data")
	}

	return nil
}

func buildCabinetDetails(
	cabinetDefinitions map[sls.CabinetKind]cabinetDefinition, cabDetailFile sls.CabinetDetailFile,
) []sls.CabinetGroupDetail {
	for _, cabType := range sls.ValidCabinetTypes {
		pos, err := positionInCabinetList(
			cabType,
			cabDetailFile.Cabinets,
		)
		if err != nil {
			var tmpCabinet sls.CabinetGroupDetail
			tmpCabinet.Kind = cabType
			tmpCabinet.Cabinets = cabinetDefinitions[cabType].count
			tmpCabinet.StartingCabinet = cabinetDefinitions[cabType].startingID
			tmpCabinet.PopulateIds()
			cabDetailFile.Cabinets = append(
				cabDetailFile.Cabinets,
				tmpCabinet,
			)
		} else {
			cabDetailFile.Cabinets[pos].Cabinets = cabDetailFile.Cabinets[pos].Length()
			cabDetailFile.Cabinets[pos].PopulateIds()
		}
	}
	return cabDetailFile.Cabinets
}

func loadHMNConnectionsFile(path string) (
	rows []shcdParser.HMNRow, err error,
) {
	err = csiFiles.ReadJSONConfig(
		path,
		&rows,
	)
	return
}

func positionInCabinetList(
	kind sls.CabinetKind, cabs []sls.CabinetGroupDetail,
) (
	int, error,
) {
	for i, cab := range cabs {
		if cab.Kind == kind {
			return i, nil
		}
	}
	return 0, fmt.Errorf(
		"%s not found",
		kind,
	)
}
