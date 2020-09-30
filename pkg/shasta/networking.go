/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"net"

	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
)

// IPReservation is a type for managing IP Reservations
type IPReservation struct {
	IPAddress net.IPAddr `yaml:"ip_address"`
	Name      string     `yaml:"name"`
	Comment   string     `yaml:"comment"`
}

// IPV4Network is a type for managing IPv4 Networks
type IPV4Network struct {
	FullName       string                 `yaml:"full_name"`
	CIDR           string                 `yaml:"cidr"`
	IPReservations []IPReservation        `yaml:"ip_reservations"`
	Subnets        []IPV4Subnet           `yaml:"subnets"`
	Name           string                 `yaml:"name"`
	VlanRange      []int16                `yaml:"vlan_range"`
	MTU            int16                  `yaml:"mtu"`
	NetType        sls_common.NetworkType `yaml:"type"`
	Comment        string                 `yaml:"comment"`
}

// IPV4Subnet is a type for managing IPv4 Subnets
type IPV4Subnet struct {
	FullName       string          `yaml:"full_name" form:"full_name" mapstructure:"full_name"`
	CIDR           net.IPNet       `yaml:"cidr"` // Convert to net.IPNet on use
	IPReservations []IPReservation `yaml:"ip_reservations"`
	Name           string          `yaml:"name" form:"name" mapstructure:"name"`
	VlanID         int16           `yaml:"vlan_id" form:"vlan_id" mapstructure:"vlan_id"`
	Comment        string          `yaml:"comment"`
	Gateway        net.IP          `yaml:"gateway"`
	DHCPStart      net.IP          `yaml:"iprange-start"`
	DHCPEnd        net.IP          `yaml:"iprange-end"`
}

// ManagementSwitch is a type for managing Management switches
type ManagementSwitch struct {
	Xname               string `form:"xname" mapstructure:"xname"`
	Name                string `form:"name" mapstructure:"name"`
	Brand               string `form:"brand" mapstructure:"brand"`
	Model               string `form:"model" mapstructure:"model"`
	Os                  string `form:"operating-system" mapstructure:"operating-system"`
	Firmware            string `form:"firmware" mapstructure:"firmware"`
	SwitchType          string //"CDU/Leaf/Spine"
	ManagementInterface net.IPAddr
}
