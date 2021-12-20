// MIT License
//
// (C) Copyright [2019, 2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package sls_common

import (
	"encoding/json"
	"net"

	base "github.com/Cray-HPE/hms-base"
)

/*
GenericHardware is the top level type for hardware in the database.  See the
Type property to cast it to whatever type it actually represets.
*/
type GenericHardware struct {
	Parent             string        `json:"Parent"`
	Children           []string      `json:"Children,omitempty"`
	Xname              string        `json:"Xname"`
	Type               HMSStringType `json:"Type"`
	Class              CabinetType   `json:"Class"`
	TypeString         base.HMSType  `json:"TypeString"`
	LastUpdated        int64         `json:"LastUpdated,omitempty"`
	LastUpdatedTime    string        `json:"LastUpdatedTime,omitempty"`
	ExtraPropertiesRaw interface{}   `json:"ExtraProperties,omitempty"`
	VaultData          interface{}   `json:"VaultData,omitempty"`
}

type GenericHardwareArray []GenericHardware

/*
GetParent returns the string xname of the parent of this object.
*/
func (gh *GenericHardware) GetParent() string {
	return gh.Parent
}
func (gh *GenericHardware) SetParent(xname string) {
	gh.Parent = xname
}

/*
GetChildren returns a slice of strings, where each entry is an xname of a
child.
*/
func (gh *GenericHardware) GetChildren() []string {
	return gh.Children
}
func (gh *GenericHardware) SetChildren(children []string) {
	gh.Children = children
}

/*
GetXname gets the xname of this object
*/
func (gh *GenericHardware) GetXname() string {
	return gh.Xname
}
func (gh *GenericHardware) SetXname(xn string) {
	gh.Xname = xn
}

func (gh *GenericHardware) GetType() HMSStringType {
	return gh.Type
}

func (gh *GenericHardware) GetClass() CabinetType {
	return gh.Class
}

func (gh *GenericHardware) GetTypeString() base.HMSType {
	return gh.TypeString
}

func (gh *GenericHardware) ToJson() (*string, error) {
	obj, err := json.Marshal(gh)
	if err != nil {
		return nil, err
	}
	strObj := string(obj)
	return &strObj, nil
}

func (gh *GenericHardware) FromJson(in string) error {
	err := json.Unmarshal([]byte(in), gh)
	return err
}

/*
ComptypeCabinet represents an object of type comptype_cabinet.
    "x3000": {
      "Class":  "River",
      "ExtraProperties": {
        "Networks": {
          "cn": {
            "HMN": {"CIDR": "10.254.0.0/22"},
            "NMN": {"CIDR": "10.252.0.0/22"}
          },
          "ncn": {
            "HMN": {"CIDR": "10.254.0.0/22"},
            "NMN": {"CIDR": "10.252.0.0/22"}
          }
        }
      },
      "Parent": "s0",
      "Type":   "comptype_cabinet",
      "Xname":  "x3000"
    },
*/
type ComptypeCabinet struct {
	// Networks has at the top the hardware type, then inside of that the network ID, then inside of that the object.
	Networks          map[string]map[string]CabinetNetworks
	DHCPRelaySwitches []string `json:",omitempty"`
}
type CabinetNetworks struct {
	CIDR       string `json:"CIDR"`
	Gateway    string `json:",omitempty"`
	VLan       int    `json:",omitempty"`
	IPv6Prefix string `json:",omitempty"`
	MACPrefix  string `json:",omitempty"`
}

/*
ComptypeCompmodPowerConnector represetns an object of type
comptype_compmod_power_connector (aka: a power connection for
a River blade)
*/
type ComptypeCompmodPowerConnector struct {
	PoweredBy []string `json:"PoweredBy"`
}

/*
ComptypeHSNConnector represents a comptype_hsn_connector, aka: a HSN
cable.
*/
type ComptypeHSNConnector struct {
	NodeNics []string `json:"NodeNics"`
}

/*
ComptypeMgmtSwitch represents a comptype_mgmt_switch (aka: a switch on the
management (NMN or HMN) network(s))
				        "IP4addr":          "127.0.0.1",
				        "Model":            "S3048T-ON",
				        "SNMPAuthPassword": "vault://hms-creds/x3000c0w22",
				        "SNMPAuthProtocol": "MD5",
				        "SNMPPrivPassword": "vault://hms-creds/x3000c0w22",
				        "SNMPPrivProtocol": "DES",
				        "SNMPUsername":     "testuser"
*/
type ComptypeMgmtSwitch struct {
	IP6Addr          string `json:"IP6addr,omitempty"`
	IP4Addr          string `json:"IP4addr,omitempty"`
	Brand            string `json:"Brand,omitempty"`
	Model            string `json:"Model,omitempty"`
	SNMPAuthPassword string `json:"SNMPAuthPassword,omitempty"`
	SNMPAuthProtocol string `json:"SNMPAuthProtocol,omitempty"`
	SNMPPrivPassword string `json:"SNMPPrivPassword,omitempty"`
	SNMPPrivProtocol string `json:"SNMPPrivProtocol,omitempty"`
	SNMPUsername     string `json:"SNMPUsername,omitempty"`

	Aliases []string `json:"Aliases,omitempty"`
}

/*
ComptypeMgmtSwitchConnector represents a comptye_mgmt_switch_connector, or
a port on a management switch
*/
type ComptypeMgmtSwitchConnector struct {
	NodeNics   []string `json:"NodeNics"`
	VendorName string
}

/*
ComptypeMgmtHLSwitch represents a comptype_hl_switch (aka: a higher level
management switch, such as Spine or Aggergation switch on the management
(NMN or HMN) network(s))
*/
type ComptypeMgmtHLSwitch struct {
	IP6Addr          string `json:"IP6addr,omitempty"`
	IP4Addr          string `json:"IP4addr,omitempty"`
	Brand            string `json:"Brand,omitempty"`
	Model            string `json:"Model,omitempty"`
	SNMPAuthPassword string `json:"SNMPAuthPassword,omitempty"`
	SNMPAuthProtocol string `json:"SNMPAuthProtocol,omitempty"`
	SNMPPrivPassword string `json:"SNMPPrivPassword,omitempty"`
	SNMPPrivProtocol string `json:"SNMPPrivProtocol,omitempty"`
	SNMPUsername     string `json:"SNMPUsername,omitempty"`

	Aliases []string `json:"Aliases,omitempty"`
}

/*
ComptypeCDUMgmtSwitch represents a comptype_cdu_mgmt_switch, (aka: a
management switch in a mountain CDU)
*/
type ComptypeCDUMgmtSwitch struct {
	Brand   string   `json:"Brand,omitempty"`
	Model   string   `json:"Model,omitempty"`
	Aliases []string `json:"Aliases,omitempty"`
}

/*
ComptypeRtrBmc represents a comptype_rtr_bmc, or the BMC of a switch
controller or TOR switch
*/
type ComptypeRtrBmc struct {
	IP6Addr  string `json:"IP6addr,omitempty"`
	IP4Addr  string `json:"IP4addr,omitempty"`
	Username string `json:"Username,omitempty"`
	Password string `json:"Password,omitempty"`
}

/*
ComptypeNodeBmc represents a comptype_nodecard, the BMC of a compute node
*/
type ComptypeNodeBmc struct {
	IP6Addr  string `json:"IP6addr,omitempty"`
	IP4Addr  string `json:"IP4addr,omitempty"`
	Username string `json:"Username,omitempty"`
	Password string `json:"Password,omitempty"`

	Aliases []string `json:"Aliases,omitempty"`
}

/*
ComptypeChassisBmc represents a comptype_chassis_bmc, the BMC of a chassis
*/
type ComptypeChassisBmc struct {
	Aliases []string `json:"Aliases,omitempty"`
}

/*
ComptypeRtrBmcNic represents a comptype_rtr_bc_nic, which is the physical
network interface port of a BMC for a switch controller or TOR switch
*/
type ComptypeRtrBmcNic struct {
	Networks []string `json:"Networks"`
	Peers    []string `json:"Peers"`
}

/*
ComptypeBmcNic represents a comptype_bmc_nic, the NIC associated with a BMC.
*/
type ComptypeBmcNic struct {
	IP6Addr  string   `json:"IP6addr,omitempty"`
	IP4Addr  string   `json:"IP4addr,omitempty"`
	Username string   `json:"Username,omitempty"`
	Password string   `json:"Password,omitempty"`
	Networks []string `json:"Networks"`
	Peers    []string `json:"Peers"`
}

/*
ComptypeCabPduNic represents a comptype_cab_pdu_nic, the NIC associated with a
cabinet-level power distribution unit (PDU)
*/
type ComptypeCabPduNic struct {
	Networks []string `json:"Networks"`
	Peers    []string `json:"Peers"`
}

/*
ComptypeNodeHsnNic represents a comptype_node_hsn_nic, the network interface to
the HSN on a compute node.
*/
type ComptypeNodeHsnNic struct {
	Networks []string `json:"Networks"`
	Peers    []string `json:"Peers"`
}

/*
ComptypeNodeNic is a comptyep_node_nic and represents a non-HSN compute node NIC
*/
type ComptypeNodeNic struct {
	Networks []string `json:"Networks"`
	Peers    []string `json:"Peers"`
}

/*
ComptypeRtrMod is a comptype_rtrmod and represents a Mountain router module
or River top-of-rack switch
*/
type ComptypeRtrMod struct {
	PowerConnector string `json:"PowerConenctor,omitempty"`
}

/*
ComptypeCompmod represents a comptype_compmod, a Mountain cabinet compute
blade slot or River rack compute blade slot.
*/
type ComptypeCompmod struct {
	PowerConnector string `json:"PowerConenctor,omitempty"`
}

/*
ComptypeNode is a comptype_node, representing a node on a Mountain compute
Blade, or node on a River node blade.
*/
type ComptypeNode struct {
	NID     int      `json:"NID,omitempty"`
	Role    string   `json:"Role,omitempty"`
	SubRole string   `json:"SubRole,omitempty"`
	Aliases []string `json:"Aliases,omitempty"`
}

/*
The Network object represents an abstract network.  For example, the
High Speed Network.
*/
type Network struct {
	Name               string      `json:"Name"`
	FullName           string      `json:"FullName"`
	IPRanges           []string    `json:"IPRanges"`
	Type               NetworkType `json:"Type"`
	LastUpdated        int64       `json:"LastUpdated,omitempty"`
	LastUpdatedTime    string      `json:"LastUpdatedTime,omitempty"`
	ExtraPropertiesRaw interface{} `json:"ExtraProperties,omitempty"`
}

// NetworkExtraProperties provides additional network information
type NetworkExtraProperties struct {
	CIDR      string  `json:"CIDR"`
	VlanRange []int16 `json:"VlanRange"`
	MTU       int16   `json:"MTU,omitempty"`
	Comment   string  `json:"Comment,omitempty"`
	PeerASN   int     `json:"PeerASN,omitempty"`
	MyASN     int     `json:"MyASN,omitempty"`

	Subnets []IPV4Subnet `json:"Subnets"`
	SystemDefaultRoute string `json:"SystemDefaultRoute,omitempty"`
}

// IPReservation is a type for managing IP Reservations
type IPReservation struct {
	Name      string   `json:"Name"`
	IPAddress net.IP   `json:"IPAddress"`
	Aliases   []string `json:"Aliases,omitempty"`

	Comment string `json:"Comment,omitempty"`
}

// IPV4Subnet is a type for managing IPv4 Subnets
type IPV4Subnet struct {
	FullName         string          `json:"FullName"`
	CIDR             string          `json:"CIDR"`
	IPReservations   []IPReservation `json:"IPReservations,omitempty"`
	Name             string          `json:"Name"`
	VlanID           int16           `json:"VlanID"`
	Gateway          net.IP          `json:"Gateway"`
	DHCPStart        net.IP          `json:"DHCPStart,omitempty"`
	DHCPEnd          net.IP          `json:"DHCPEnd,omitempty"`
	Comment          string          `json:"Comment,omitempty"`
	ReservationStart net.IP          `json:"ReservationStart,omitempty"`
	ReservationEnd   net.IP          `json:"ReservationEnd,omitempty"`
	MetalLBPoolName  string          `json:"MetalLBPoolName,omitempty"`
}

type NetworkArray []Network

// SLSGeneratorInputState is given to the SLS config generator in order to generator the SLS config file
type SLSGeneratorInputState struct {
	ManagementSwitches  map[string]GenericHardware `json:"ManagementSwitches"` // SLS Type: comptype_mgmt_switch
	RiverCabinets       map[string]GenericHardware `json:"RiverCabinets"`      // SLS Type: comptype_cabinet
	HillCabinets        map[string]GenericHardware `json:"HillCabinets"`       // SLS Type: comptype_cabinet
	MountainCabinets    map[string]GenericHardware `json:"MountainCabinets"`   // SLS Type: comptype_cabinet
	MountainStartingNid int                        `json:"MountainStartingNid"`

	Networks map[string]Network `json:"Networks"`
}
