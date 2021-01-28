/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	shcd_parser "stash.us.cray.com/HMS/hms-shcd-parser/pkg/shcd-parser"
	"stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/csi"
)

type cabinetDefinition struct {
	count      int
	startingID int
}

func collectInput(v *viper.Viper) ([]shcd_parser.HMNRow, []*csi.LogicalNCN, []*csi.ManagementSwitch, csi.SLSGeneratorApplicationNodeConfig, []csi.CabinetGroupDetail) {
	// The installation requires a set of information in order to proceed
	// First, we need some kind of representation of the physical hardware
	// That is generally represented through the hmn_connections.json file
	// which is literally a cabling map with metadata about the NCNs and
	// River Compute node BMCs, Columbia Rosetta Switches, and PDUs.
	//
	// From the hmn_connections file, we can create a set of HMNRow objects
	// to use for populating SLS.
	hmnRows, err := loadHMNConnectionsFile(v.GetString("hmn-connections"))
	if err != nil {
		log.Fatalf("unable to load hmn connections, %v \n", err)
	}

	// SLS also needs to know about our networking configuration.  In order to do that,
	// we need to load the switches
	switches, err := csi.ReadSwitchCSV(v.GetString("switch-metadata"))
	if err != nil {
		log.Fatalln("Couldn't extract switches", err)
	}

	// Normalize the management switch data, before validation
	for _, mySwitch := range switches {
		mySwitch.Normalize()
	}

	if err := validateSwitchInput(switches); err != nil {
		log.Println("Unable to get reasonable Switches from your csv")
		log.Println("Does your header match the preferred style? Switch Xname,Type,Brand")
		log.Fatal("CSV Parsing failed.  Can't continue.")
	}

	// This is techincally sufficient to generate an SLSState object, but to do so now
	// would not include extended information about the NCNs and Network Switches.
	//
	// The first step in building the NCN map is to read the NCN Metadata file
	ncns, err := csi.ReadNodeCSV(v.GetString("ncn-metadata"))
	if err != nil {
		log.Fatalln("Couldn't extract ncns", err)
	}

	// Normalize the ncn data, before validation
	for _, ncn := range ncns {
		ncn.Normalize()
	}

	if err := validateNCNInput(ncns); err != nil {
		log.Println("Unable to get reasonable NCNs from your csv")
		log.Println("Does your header match the preferred style? Xname,Role,Subrole,BMC MAC,Bootstrap MAC,Bond0 MAC0,Bond0 MAC1")
		log.Fatal("CSV Parsing failed.  Can't continue.")
	}

	// Cabinet Map Configuration
	// This is an optional input file
	var cabDetailFile csi.CabinetDetailFile
	if v.IsSet("cabinets-yaml") {
		err := files.ReadYAMLConfig(v.GetString("cabinets-yaml"), &cabDetailFile)
		if err != nil {
			log.Fatalf("Unable to parse cabinets-yaml file: %s\nError: %v", v.GetString("cabinets-yaml"), err)
		}
	}

	cabDefinitions := make(map[string]cabinetDefinition)
	for _, cabType := range []string{"river", "mountain", "hill"} {
		cabDefinitions[cabType] = cabinetDefinition{
			count:      v.GetInt(fmt.Sprintf("%s-cabinets", cabType)),
			startingID: v.GetInt(fmt.Sprintf("starting-%s-cabinet", cabType))}
	}
	cabinetDetailList := buildCabinetDetails(cabDefinitions, cabDetailFile)

	// Application Node configration for SLS Config Generator
	// This is an optional input file
	var applicationNodeConfig csi.SLSGeneratorApplicationNodeConfig
	if v.IsSet("application-node-config-yaml") {
		applicationNodeConfigPath := v.GetString("application-node-config-yaml")

		log.Printf("Using application node config: %s\n", applicationNodeConfigPath)
		err := files.ReadYAMLConfig(applicationNodeConfigPath, &applicationNodeConfig)
		if err != nil {
			log.Fatalf("Unable to parse application-node-config file: %s\nError: %v", applicationNodeConfigPath, err)
		}
	}

	// Normalize application node config
	if err := applicationNodeConfig.Normalize(); err != nil {
		log.Fatalf("Failed to normalize application node config. Error: %s", err)
	}

	// Validate Application node config
	if err := applicationNodeConfig.Validate(); err != nil {
		log.Fatalf("Failed to validate application node config. Error: %s", err)
	}

	return hmnRows, ncns, switches, applicationNodeConfig, cabinetDetailList
}

func validateSwitchInput(switches []*csi.ManagementSwitch) error {
	// Validate that there is an non-zero number of NCNs extracted from ncn_metadata.csv
	if len(switches) == 0 {
		return fmt.Errorf("unable to extract Switches from switch metadata csv")
	}

	// Validate each Switch
	var mustFail = false
	for _, mySwitch := range switches {
		if err := mySwitch.Validate(); err != nil {
			mustFail = true
			log.Println("Switch from csv is invalid:", err)
		}
	}

	if mustFail {
		return fmt.Errorf("switch_metadata.csv contains invalid switch data")
	}

	return nil
}

func validateNCNInput(ncns []*csi.LogicalNCN) error {
	// Validate that there is an non-zero number of NCNs extracted from ncn_metadata.csv
	if len(ncns) == 0 {
		return fmt.Errorf("unable to extract NCNs from ncn metadata csv")
	}

	// Validate each NCN
	var mustFail = false
	for _, ncn := range ncns {
		if err := ncn.Validate(); err != nil {
			mustFail = true
			log.Println("NCN from csv is invalid", ncn, err)
		}
	}

	if mustFail {
		return fmt.Errorf("ncn_metadata.csv contains invalid NCN data")
	}

	return nil
}

func buildCabinetDetails(cabinetDefinitions map[string]cabinetDefinition, cabDetailFile csi.CabinetDetailFile) []csi.CabinetGroupDetail {
	for _, cabType := range []string{"river", "mountain", "hill"} {
		pos, err := positionInCabinetList(cabType, cabDetailFile.Cabinets)
		if err != nil {
			var tmpCabinet csi.CabinetGroupDetail
			tmpCabinet.Kind = cabType
			tmpCabinet.Cabinets = cabinetDefinitions[cabType].count
			tmpCabinet.StartingCabinet = cabinetDefinitions[cabType].startingID
			tmpCabinet.PopulateIds()
			cabDetailFile.Cabinets = append(cabDetailFile.Cabinets, tmpCabinet)
		} else {
			cabDetailFile.Cabinets[pos].Cabinets = cabDetailFile.Cabinets[pos].Length()
			cabDetailFile.Cabinets[pos].PopulateIds()
		}
	}
	return cabDetailFile.Cabinets
}

func positionInCabinetList(kind string, cabs []csi.CabinetGroupDetail) (int, error) {
	for i, cab := range cabs {
		if cab.Kind == kind {
			return i, nil
		}
	}
	return 0, fmt.Errorf("%s not found", kind)
}
