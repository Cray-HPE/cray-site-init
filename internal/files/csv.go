/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"os"

	"github.com/gocarina/gocsv"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// ReadSwitchCSV parses a CSV file into a list of ManagementSwitch structs
func ReadSwitchCSV(filename string) ([]*shasta.ManagementSwitch, error) {
	switches := []*shasta.ManagementSwitch{}
	return switches, nil
}

// ReadNodeCSV parses a CSV file into a list of NCN_bootstrap nodes for use by the installer
func ReadNodeCSV(filename string) ([]*shasta.BootstrapNCNMetadata, error) {
	nodes := []*shasta.BootstrapNCNMetadata{}

	ncnMetadataFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nodes, err
	}
	defer ncnMetadataFile.Close()

	if err := gocsv.UnmarshalFile(ncnMetadataFile, &nodes); err != nil { // Load nodes from file
		return nodes, err
	}
	return nodes, nil
}
