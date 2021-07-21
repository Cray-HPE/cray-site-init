/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	shcd_parser "github.com/Cray-HPE/hms-shcd-parser/pkg/shcd-parser"
	"github.com/spf13/viper"
)

type cabinetDefinition struct {
	count      int
	startingID int
}

func collectHMNRows(v *viper.Viper) []shcd_parser.HMNRow {
	seedFileHmnConnections := filepath.Dir(viper.ConfigFileUsed()) + "/" + v.GetString("hmn-connections")
	hmnRows, err := loadHMNConnectionsFile(seedFileHmnConnections)
	if err != nil {
		log.Fatalf("unable to load hmn connections, %v \n", err)
	}
	return hmnRows
}

func collectNCNMeta(v *viper.Viper) []*csi.LogicalNCN {
	seedFileNcnMetadata := filepath.Dir(viper.ConfigFileUsed()) + "/" + v.GetString("ncn-metadata")
	ncns, err := csi.ReadNodeCSV(seedFileNcnMetadata)
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

	return ncns
}

func collectSwitches(v *viper.Viper) []*csi.ManagementSwitch {
	seedFileSwitchMetadata := filepath.Dir(v.ConfigFileUsed()) + "/" + v.GetString("switch-metadata")

	switches, err := csi.ReadSwitchCSV(seedFileSwitchMetadata)
	if err != nil {
		log.Fatalf("Couldn't extract switches, %v", err)
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

	return switches
}

func collectApplicationNodeConfig(v *viper.Viper) csi.SLSGeneratorApplicationNodeConfig {
	var applicationNodeConfig csi.SLSGeneratorApplicationNodeConfig
	if v.IsSet("application-node-config-yaml") {
		seedFileAppNodeConfig := filepath.Dir(viper.ConfigFileUsed()) + "/" + v.GetString("application-node-config-yaml")

		log.Printf("Using application node config: %s\n", seedFileAppNodeConfig)
		err := files.ReadYAMLConfig(seedFileAppNodeConfig, &applicationNodeConfig)
		if err != nil {
			log.Fatalf("Unable to parse application-node-config file: %s\nError: %v", seedFileAppNodeConfig, err)
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

	return applicationNodeConfig
}

func collectCabinets(v *viper.Viper) []csi.CabinetGroupDetail {
	var cabDetailFile csi.CabinetDetailFile
	if v.IsSet("cabinets-yaml") {
		seedFileCabinets := filepath.Dir(viper.ConfigFileUsed()) + "/" + v.GetString("cabinets-yaml")
		err := files.ReadYAMLConfig(seedFileCabinets, &cabDetailFile)
		if err != nil {
			log.Fatalf("Unable to parse cabinets-yaml file: %s\nError: %v", v.GetString("cabinets-yaml"), err)
		}
	}

	cabDefinitions := make(map[string]cabinetDefinition)
	for _, cabType := range csi.ValidCabinetTypes {
		cabDefinitions[cabType] = cabinetDefinition{
			count:      v.GetInt(fmt.Sprintf("%s-cabinets", cabType)),
			startingID: v.GetInt(fmt.Sprintf("starting-%s-cabinet", cabType))}
	}
	cabinetDetailList := buildCabinetDetails(cabDefinitions, cabDetailFile)

	return cabinetDetailList
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
	// this seedfile should be in the same place as the config, so use that to craft the path
	hmnRows := collectHMNRows(v)

	// This is techincally sufficient to generate an SLSState object, but to do so now
	// would not include extended information about the NCNs and Network Switches.
	//
	// The first step in building the NCN map is to read the NCN Metadata file
	ncns := collectNCNMeta(v)

	// SLS also needs to know about our networking configuration.  In order to do that,
	// we need to load the switches
	switches := collectSwitches(v)

	// Application Node configration for SLS Config Generator
	// This is an optional input file
	applicationNodeConfig := collectApplicationNodeConfig(v)

	// Cabinet Map Configuration
	// This is an optional input file
	cabinetDetailList := collectCabinets(v)

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
	for _, cabType := range csi.ValidCabinetTypes {
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
