/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"net"
)

// IPReservation is a type for managing IP Reservations
type IPReservation struct {
	IPAddress net.IPAddr
	Name      string
	Comment   string
}

// IPV4Network is a type for managing IPv4 Networks
type IPV4Network struct {
	FullName       string          `yaml:"full_name"`
	CIDR           string          `yaml:"cidr"`
	IPReservations []IPReservation `yaml:"ip_reservations"`
	Subnets        []IPV4Subnet    `yaml:"subnets"`
	Name           string          `yaml:"name"`
	Vlan           int16           `yaml:"vlan_id"`
	MTU            int16           `yaml:"mtu"`
	NetType        string          `yaml:"type"` // ethernet or ???
	Comment        string          `yaml:"comment"`
}

// IPV4Subnet is a type for managing IPv4 Subnets
type IPV4Subnet struct {
	FullName       string          `yaml:"full_name"`
	CIDR           string          `yaml:"cidr"` // Convert to net.IPNet on use
	IPReservations []IPReservation `yaml:"ip_reservations"`
	Name           string          `yaml:"name"`
	Comment        string          `yaml:"comment"`
	Gateway        string          `yaml:"gateway"` // Convert to net.IPAddr on use
	DHCPStart      net.IPAddr      `yaml:"iprange"`
	DHCPEnd        net.IPAddr      `yaml:"iprange"`
}
