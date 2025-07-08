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
	"fmt"
	"log"
	"os"
	"path/filepath"

	shcdParser "github.com/Cray-HPE/hms-shcd-parser/pkg/shcd-parser"
	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/internal/files"
	slsInit "github.com/Cray-HPE/cray-site-init/pkg/cli/config/initialize/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

const (
	// DefaultApplicationNodeConfigFilename is the default filename for the application node configuration file.
	DefaultApplicationNodeConfigFilename = "application_node_config.yaml"
	// DefaultCabinetsFilename is the default filename for the cabinents file.
	DefaultCabinetsFilename = "cabinets.yaml"
	// DefaultHMNConnectionsFilename is the default filename for the HMN connections file.
	DefaultHMNConnectionsFilename = "hmn_connections.json"
	// DefaultNCNMetadataFilename is the default filename for NCN metadata file.
	DefaultNCNMetadataFilename = "ncn_metadata.csv"
	// DefaultSwitchMetadataFilename is the default filename for switch metadata file.
	DefaultSwitchMetadataFilename = "switch_metadata.csv"
)

var (
	// ApplicationNodeConfigFile is the resolved file for the application node configuration.
	ApplicationNodeConfigFile string
	// CabinetsFile is the resolved file for the cabinents configuration.
	CabinetsFile string
	// HMNConnectionsFile is the resolved file for the HMN connections configuration.
	HMNConnectionsFile string
	// NCNMetadataFile is the resolved file for NCN metadata configuration.
	NCNMetadataFile string
	// SwitchMetadataFile is the resolved file for switch metadata configuration.
	SwitchMetadataFile string
)

type cabinetDefinition struct {
	count      int
	startingID int
}

func getFile(name string) (path string, err error) {
	v := viper.GetViper()
	inputDir := v.GetString("input-dir")
	if inputDir == "" {
		path = fmt.Sprintf(
			"%s%s%s",
			filepath.Dir(v.ConfigFileUsed()),
			string(os.PathSeparator),
			name,
		)
	} else {
		path = fmt.Sprintf(
			"%s%s%s",
			inputDir,
			string(os.PathSeparator),
			name,
		)
	}
	_, err = os.Stat(path)
	return path, err
}

func collectHMNRows(v *viper.Viper) (hmnRows []shcdParser.HMNRow, err error) {
	hmnConnectionsFile := v.GetString("hmn-connections")
	if hmnConnectionsFile == "" {
		HMNConnectionsFile = DefaultHMNConnectionsFilename
	} else {
		HMNConnectionsFile = hmnConnectionsFile
	}
	seedFileHmnConnections, err := getFile(HMNConnectionsFile)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading hmn-connections: %w",
			err,
		)
	}
	hmnRows, err = loadHMNConnectionsFile(seedFileHmnConnections)
	if err != nil {
		return nil, fmt.Errorf(
			"error loading hmn-connections file: %v",
			err,
		)
	}
	return hmnRows, err
}

func collectNCNMeta(v *viper.Viper) (ncns []*LogicalNCN, err error) {
	ncnFile := v.GetString("ncn-file")
	if ncnFile == "" {
		NCNMetadataFile = DefaultNCNMetadataFilename
	} else {
		NCNMetadataFile = ncnFile
	}
	seedFileNcnMetadata, err := getFile(NCNMetadataFile)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading ncn-metadata file because %v",
			err,
		)
	}
	ncns, err = ReadNodeCSV(seedFileNcnMetadata)
	if err != nil {
		return nil, fmt.Errorf(
			"couldn't extract ncns: %v",
			err,
		)
	}

	// Normalize the ncn Data, before validation
	for _, ncn := range ncns {
		err := ncn.Normalize()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to normalize NCN Data because %v",
				err,
			)
		}
	}

	if err := validateNCNInput(ncns); err != nil {
		return nil, fmt.Errorf(
			"unable to get reasonable NCNs from your csv because %v",
			err,
		)
	}

	return ncns, err
}

func collectSwitches(v *viper.Viper) (switches []*networking.ManagementSwitch, err error) {
	switchMetadataFile := v.GetString("switch-metadata")
	if switchMetadataFile == "" {
		SwitchMetadataFile = DefaultSwitchMetadataFilename
	} else {
		SwitchMetadataFile = switchMetadataFile
	}
	seedFileSwitchMetadata, err := getFile(SwitchMetadataFile)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading switch-metadata file because %v",
			err,
		)
	}

	switches, err = networking.ReadSwitchCSV(seedFileSwitchMetadata)
	if err != nil {
		return nil, fmt.Errorf(
			"couldn't extract switches because %v",
			err,
		)
	}

	// Normalize the management switch Data, before validation
	errors := make(
		[]error,
		0,
	)
	for _, mySwitch := range switches {
		mySwitch.Normalize()
	}
	if len(errors) != 0 {
		return nil, fmt.Errorf(
			"couldn't extract switches from CSV because %v",
			errors,
		)
	}

	if err = validateSwitchInput(switches); err != nil {
		return nil, fmt.Errorf(
			"unable to get reasonable Switches from your csv because %v",
			err,
		)
	}

	return switches, err
}

func collectApplicationNodeConfig(v *viper.Viper) (applicationNodeConfig slsInit.GeneratorApplicationNodeConfig, err error) {
	if !v.IsSet("application-node-config-yaml") {
		return applicationNodeConfig, nil
	}
	applicationNodeConfigFile := v.GetString("application-node-config-yaml")
	if applicationNodeConfigFile == "" {
		ApplicationNodeConfigFile = DefaultApplicationNodeConfigFilename
		_, err := os.Stat(ApplicationNodeConfigFile)
		if err != nil {
			return applicationNodeConfig, nil
		}
	} else {
		ApplicationNodeConfigFile = applicationNodeConfigFile
	}
	seedFileAppNodeConfig, err := getFile(ApplicationNodeConfigFile)
	if err != nil {
		return applicationNodeConfig, fmt.Errorf(
			"error reading application-node-config-yaml because %v",
			v.GetString("application-node-config-yaml"),
		)
	}

	log.Printf(
		"Using application node config: %s\n",
		seedFileAppNodeConfig,
	)
	err = files.ReadYAMLConfig(
		seedFileAppNodeConfig,
		&applicationNodeConfig,
	)
	if err != nil {
		return applicationNodeConfig, fmt.Errorf(
			"unable to parse application-node-config file [%s] because %v",
			seedFileAppNodeConfig,
			err,
		)
	}

	// Normalize application node config
	if err = applicationNodeConfig.Normalize(); err != nil {
		return applicationNodeConfig, fmt.Errorf(
			"failed to normalize application node config because %v",
			err,
		)
	}

	// Validate Application node config
	if err = applicationNodeConfig.Validate(); err != nil {
		return applicationNodeConfig, fmt.Errorf(
			"failed to validate application node config because %v",
			err,
		)
	}

	return applicationNodeConfig, err
}

func collectCabinets(v *viper.Viper) (cabinetDetailList []sls.CabinetGroupDetail, err error) {
	var cabDetailFile sls.CabinetDetailFile
	if v.IsSet("cabinets-yaml") {
		cabinetsFile := v.GetString("cabinets-yaml")
		if cabinetsFile == "" {
			CabinetsFile = DefaultCabinetsFilename
		} else {
			CabinetsFile = cabinetsFile
			seedFileCabinets, err := getFile(v.GetString("cabinets-yaml"))
			if err != nil {
				return nil, fmt.Errorf(
					"error reading cabinets-yaml file because %v",
					err,
				)
			}
			err = files.ReadYAMLConfig(
				seedFileCabinets,
				&cabDetailFile,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"unable to parse cabinets-yaml file [%s] because %v",
					v.GetString("cabinets-yaml"),
					err,
				)
			}
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
	cabinetDetailList = buildCabinetDetails(
		cabDefinitions,
		cabDetailFile,
	)

	// Verify no duplicate cabinet ids are present
	knownCabinetIDs := map[int]bool{}
	for _, cabinetGroupDetail := range cabinetDetailList {
		for _, id := range cabinetGroupDetail.CabinetIDs() {
			if knownCabinetIDs[id] {
				return nil, fmt.Errorf(
					"found duplicate cabinet id: %v",
					id,
				)
			}

			knownCabinetIDs[id] = true
		}
	}

	return cabinetDetailList, err
}

func collectInput(v *viper.Viper) (
	hmnRows []shcdParser.HMNRow,
	ncns []*LogicalNCN,
	switches []*networking.ManagementSwitch,
	applicationNodeConfig slsInit.GeneratorApplicationNodeConfig,
	cabinetDetailList []sls.CabinetGroupDetail,
	errs []error,
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
	hmnRows, err := collectHMNRows(v)
	if err != nil {
		errs = append(
			errs,
			err,
		)
	}

	// This is technically sufficient to generate an SLSState object, but to do so now
	// would not include extended information about the NCNs and Network Switches.
	//
	// The first step in building the NCN map is to read the NCN Metadata file
	ncns, err = collectNCNMeta(v)
	if err != nil {
		errs = append(
			errs,
			err,
		)
	}

	// SLS also needs to know about our networking configuration. In order to do that,
	// we need to load the switches
	switches, err = collectSwitches(v)
	if err != nil {
		errs = append(
			errs,
			err,
		)
	}

	// Application Node configuration for SLS Config Generator
	// This is an optional input file
	applicationNodeConfig, err = collectApplicationNodeConfig(v)
	if err != nil {
		errs = append(
			errs,
			err,
		)
	}

	// Cabinet Map Configuration
	// This is an optional input file
	cabinetDetailList, err = collectCabinets(v)
	if err != nil {
		errs = append(
			errs,
			err,
		)
	}

	return hmnRows, ncns, switches, applicationNodeConfig, cabinetDetailList, errs
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
		return fmt.Errorf("switch_metadata.csv contains invalid switch Data")
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
		return fmt.Errorf("ncn_metadata.csv contains invalid NCN Data")
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
	err = files.ReadJSONConfig(
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
