/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"os"

	"github.com/gocarina/gocsv"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// ReadCSV parses a CSV file into a list of NCN_bootstrap nodes for use by the installer
func ReadCSV(filename string) []*shasta.BootstrapNCNMetadata {
	ncnMetadataFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer ncnMetadataFile.Close()

	nodes := []*shasta.BootstrapNCNMetadata{}

	if err := gocsv.UnmarshalFile(ncnMetadataFile, &nodes); err != nil { // Load nodes from file
		panic(err)
	}
	return nodes
}
