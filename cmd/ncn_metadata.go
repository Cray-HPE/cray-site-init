/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

// BootstrapNCNMetadata is a type that matches the ncn_metadata.csv file as
// NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
type BootstrapNCNMetadata struct {
	Xname   string `csv:"NCN xname"`
	Role    string `csv:"NCN Role"`
	Subrole string `csv:"NCN Subrole"`
	BmcMac  string `csv:"BMC MAC"`
	BmcPort string `csv:"BMC Switch Port"`
	NmnPac  string `csv:"NMN MAC"`
	NmnPort string `csv:"NMN Switch Port"`
}

// ReadCSV parses a CSV file into a list of NCN_bootstrap nodes for use by the installer
func ReadCSV(filename string) []*BootstrapNCNMetadata {
	ncnMetadataFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer ncnMetadataFile.Close()

	nodes := []*BootstrapNCNMetadata{}

	if err := gocsv.UnmarshalFile(ncnMetadataFile, &nodes); err != nil { // Load nodes from file
		panic(err)
	}
	for _, node := range nodes {
		fmt.Println("Hello", node.Xname)
	}

	return nodes
}
