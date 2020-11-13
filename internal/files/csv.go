/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"io"
	"os"

	"github.com/gocarina/gocsv"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
)

// ReadSwitchCSV parses a CSV file into a list of ManagementSwitch structs
func ReadSwitchCSV(filename string) ([]*shasta.ManagementSwitch, error) {
	switches := []*shasta.ManagementSwitch{}
	return switches, nil
}

// ReadNodeCSV parses a CSV file into a list of NCN_bootstrap nodes for use by the installer
func ReadNodeCSV(filename string) ([]*shasta.BootstrapNCNMetadata, error) {
	nodes := []*shasta.BootstrapNCNMetadata{}
	newNodes := []*shasta.NewBootstrapNCNMetadata{}

	ncnMetadataFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nodes, err
	}
	defer ncnMetadataFile.Close()
	// In 1.4, we have a new format for this file.  Try that first and then fall back to the older style if necessary
	err = gocsv.UnmarshalFile(ncnMetadataFile, &newNodes)
	if err != nil {
		for _, node := range newNodes {
			nodes = append(nodes, &shasta.BootstrapNCNMetadata{
				Xname:     node.Xname,
				Role:      node.Role,
				Subrole:   node.Subrole,
				BmcMac:    node.BmcMac,
				NmnMac:    node.BootstrapMac,
				Bond0Macs: []string{node.Bond0Mac0, node.Bond0Mac1},
			})
		}
		return nodes, nil
	}

	// Be Kind Rewind https://www.imdb.com/title/tt0799934/
	ncnMetadataFile.Seek(0, io.SeekStart)
	err = gocsv.UnmarshalFile(ncnMetadataFile, &nodes)
	if err != nil { // Load nodes from file
		return nodes, nil
	}
	return nodes, err
}
