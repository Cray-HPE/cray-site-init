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
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"os"

	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/gocarina/gocsv"
)

// LogicalNCN is the main struct for NCNs
// It replaces the deprecated BootstrapNCNMetadata
// and still matches the ncn_metadata.csv file as
// NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
type LogicalNCN struct {
	Role             string       `yaml:"role" json:"role" csv:"NCN Role"`
	Subrole          string       `yaml:"subrole" json:"subrole" csv:"NCN Subrole"`
	BmcMac           string       `yaml:"bmc-mac" json:"bmc-mac" csv:"BMC MAC"`
	BmcPort          string       `yaml:"bmc-port" json:"bmc-port" csv:"BMC Switch Port"`
	NmnMac           string       `yaml:"nmn-mac" json:"nmn-mac" csv:"NMN MAC"`
	NmnPort          string       `yaml:"nmn-port" json:"nmn-port" csv:"NMN Switch Port"`
	NmnIP            string       `yaml:"nmn-ip" json:"nmn-ip" csv:"-"`
	HmnIP            string       `yaml:"hmn-ip" json:"hmn-ip" csv:"-"`
	MtlIP            string       `yaml:"mtl-ip" json:"mtl-ip" csv:"-"`
	CmnIP            string       `yaml:"cmn-ip" json:"cmn-ip" csv:"-"`
	CanIP            string       `yaml:"can-ip" json:"can-ip" csv:"-"`
	Xname            string       `yaml:"xname" json:"xname" csv:"NCN xname"`
	Hostname         string       `yaml:"hostname" json:"hostname" csv:"-"`
	InstanceID       string       `yaml:"instance-id" json:"instance-id" csv:"-"` // should be unique for the life of the image
	Region           string       `yaml:"region" json:"region" csv:"-"`
	AvailabilityZone string       `yaml:"availability-zone" json:"availability-zone" csv:"-"`
	ShastaRole       string       `yaml:"shasta-role" json:"shasta-role" csv:"-"` // map to HSM Subrole
	Aliases          []string     `yaml:"aliases" json:"aliases" csv:"-"`
	Networks         []NCNNetwork `yaml:"networks" json:"networks" csv:"-"`
	BmcIP            string       `yaml:"bmc-ip" json:"bmc-ip" csv:"-"`
	Bond0Mac0        string       `yaml:"bond0-mac0" json:"bond0-mac0" csv:"-"`
	Bond0Mac1        string       `yaml:"bond0-mac1" json:"bond0-mac1" csv:"-"`
	Cabinet          string       `yaml:"cabinet" json:"cabinet" csv:"-"` // Use to establish availability zone
}

// NewBootstrapNCNMetadata is a type that matches the updated ncn_metadata.csv file as
// Xname,Role,Subrole,BMC MAC,Bootstrap MAC,Bond0 MAC0,Bond0 Mac1
// It is probable that on many machines bootstrap mac will be the same as one of the bond macs
// Do not be alarmed.
type NewBootstrapNCNMetadata struct {
	Xname        string `json:"xname" csv:"Xname"`
	Role         string `json:"role" csv:"Role"`
	Subrole      string `json:"subrole" csv:"Subrole"`
	BmcMac       string `json:"bmc-mac" csv:"BMC MAC"`
	BootstrapMac string `json:"bootstrap-mac" csv:"Bootstrap MAC"`
	Bond0Mac0    string `json:"bond0-mac0" csv:"Bond0 MAC0"`
	Bond0Mac1    string `json:"bond0-mac1" csv:"Bond0 MAC1"`
}

// LogicalUAN is like LogicalNCN, but for UANs
type LogicalUAN struct {
	Xname    string   `yaml:"xname" json:"xname" csv:"NCN xname"`
	Hostname string   `yaml:"hostname" json:"hostname" csv:"-"`
	Role     string   `yaml:"role" json:"role" csv:"NCN Role"`
	Subrole  string   `yaml:"subrole" json:"subrole" csv:"NCN Subrole"`
	CanIP    net.IP   `yaml:"bmc-ip" json:"bmc-ip" csv:"-"`
	Aliases  []string `yaml:"aliases" json:"aliases" csv:"-"`
}

// NCNNetwork holds information about networks in the NCN context
type NCNNetwork struct {
	NetworkName         string       `json:"network-name"`
	FullName            string       `json:"full-name"`
	IPv4Address         netip.Addr   `json:"ipv4-address"`
	IPv6Address         netip.Addr   `json:"ipv6-address,omitempty"`
	InterfaceName       string       `json:"net-device"`
	InterfaceMac        string       `json:"mac-address"`
	ParentInterfaceName string       `json:"parent-interface-name"`
	Vlan                int          `json:"vlan"`
	CIDR4               netip.Prefix `json:"cidr4"`
	CIDR6               netip.Prefix `json:"cidr6,omitempty"`
	Gateway4            netip.Addr   `json:"gateway4"`
	Gateway6            netip.Addr   `json:"gateway6,omitempty"`
}

// Validate is a validator that checks for a minimum set of info
func (lncn *LogicalNCN) Validate() (err error) {
	xname := lncn.Xname

	// First off verify that this is a valid xname
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf(
			"invalid xname for NCN: %s",
			xname,
		)
	}

	// Next, verify that the xname is type of Node
	if xnametypes.GetHMSType(xname) != xnametypes.Node {
		return fmt.Errorf(
			"invalid type %s for NCN xname: %s",
			xnametypes.GetHMSTypeString(xname),
			xname,
		)
	}

	if lncn.Role == "" {
		// TODO Verify the role against the listing of valid Roles
		return fmt.Errorf("empty role")
	}
	if lncn.Subrole == "" {
		// TODO Verify the role against the listing of valid SubRoles
		return fmt.Errorf("empty sub-role")
	}
	return err
}

// GetHostname returns an explicit hostname if possible, otherwise the Xname, otherwise an empty string
func (lncn LogicalNCN) GetHostname() string {
	if lncn.Hostname == "" {
		return lncn.Xname
	}
	return lncn.Hostname
}

// Normalize the values of a LogicalNCN
func (lncn *LogicalNCN) Normalize() error {
	// Right now we only need to the normalize the xname for the switch. IE strip any leading 0s
	lncn.Xname = xnametypes.NormalizeHMSCompID(lncn.Xname)

	return nil
}

// GetIP takes in a netname and returns an IP address for that netname
func (lncn *LogicalNCN) GetIP(netName string) (addr netip.Addr) {
	for _, inet := range lncn.Networks {
		if inet.NetworkName == netName {
			addr = inet.IPv4Address
			break
		}
	}
	return addr
}

// GenerateInstanceID creates an instance-id fit for use in the instance metadata
func GenerateInstanceID() (id string) {
	b := make(
		[]byte,
		4,
	)
	_, err := rand.Read(b)
	if err != nil {
		return id
	}
	id = fmt.Sprintf(
		"i-%X",
		b,
	)
	return id
}

// ReadNodeCSV parses a CSV file into a list of NCN_bootstrap nodes for use by the installer
func ReadNodeCSV(filename string) (
	[]*LogicalNCN, error,
) {
	var nodes []*LogicalNCN
	var newNodes []*NewBootstrapNCNMetadata

	ncnMetadataFile, err := os.OpenFile(
		filename,
		os.O_RDWR|os.O_CREATE,
		os.ModePerm,
	)
	if err != nil {
		return nodes, err
	}
	defer ncnMetadataFile.Close()
	// In 1.4, we have a new format for this file. Try that first and then fall back to the older style if necessary
	newErr := gocsv.UnmarshalFile(
		ncnMetadataFile,
		&newNodes,
	)
	if newErr == nil {
		for _, node := range newNodes {
			nodes = append(
				nodes,
				&LogicalNCN{
					Xname:     node.Xname,
					Role:      node.Role,
					Subrole:   node.Subrole,
					BmcMac:    node.BmcMac,
					NmnMac:    node.BootstrapMac,
					Bond0Mac0: node.Bond0Mac0,
					Bond0Mac1: node.Bond0Mac1,
				},
			)
		}
		return nodes, nil
	}

	ncnMetadataFile.Seek(
		0,
		io.SeekStart,
	)
	err = gocsv.UnmarshalFile(
		ncnMetadataFile,
		&nodes,
	)
	if err == nil { // Load nodes from file
		return nodes, nil
	}

	if newErr != nil {
		if err != nil {
			log.Println(
				"Unable to parse ncn_metadata with new style because ",
				newErr,
			)
			log.Fatal(
				"Unable to parse ncn_metadata with old format because ",
				err,
			)
		}
		log.Fatal(
			"Unable to parse ncn_metadata with new style because ",
			newErr,
		)
	}

	return nodes, err
}
