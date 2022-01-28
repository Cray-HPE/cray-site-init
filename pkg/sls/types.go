package sls

import "fmt"

// It's a shame to have to do this, but, because SLS native structures use the IP type which internally is an array of
// bytes we need a more vanilla structure to allow us to work with that data. In truth this kind of feels like a bug to
// me. For some reason when mapstructure is using the reflect package to get the `Kind()` of those data defined as
// net.IP it's giving back slice instead of string.

// NetworkExtraProperties provides additional network information
type NetworkExtraProperties struct {
	CIDR      string  `json:"CIDR"`
	VlanRange []int16 `json:"VlanRange"`
	MTU       int16   `json:"MTU,omitempty"`
	Comment   string  `json:"Comment,omitempty"`
	PeerASN   int     `json:"PeerASN,omitempty"`
	MyASN     int     `json:"MyASN,omitempty"`

	Subnets []IPV4Subnet `json:"Subnets"`
}

// LookupSubnet returns a subnet by name
func (network *NetworkExtraProperties) LookupSubnet(name string) (IPV4Subnet, error) {
	var found []IPV4Subnet
	if len(network.Subnets) == 0 {
		return IPV4Subnet{}, fmt.Errorf("subnet not found \"%v\"", name)
	}
	for _, v := range network.Subnets {
		if v.Name == name {
			found = append(found, v)
		}
	}
	if len(found) == 1 {
		return found[0], nil
	}
	if len(found) > 1 {
		return found[0], fmt.Errorf("found %v subnets instead of just one", len(found))
	}
	return IPV4Subnet{}, fmt.Errorf("subnet not found \"%v\"", name)
}

// IPReservation is a type for managing IP Reservations
type IPReservation struct {
	Name      string   `json:"Name"`
	IPAddress string   `json:"IPAddress"`
	Aliases   []string `json:"Aliases,omitempty"`

	Comment string `json:"Comment,omitempty"`
}

// IPV4Subnet is a type for managing IPv4 Subnets
type IPV4Subnet struct {
	FullName        string          `json:"FullName"`
	CIDR            string          `json:"CIDR"`
	IPReservations  []IPReservation `json:"IPReservations,omitempty"`
	Name            string          `json:"Name"`
	VlanID          int16           `json:"VlanID"`
	Gateway         string          `json:"Gateway"`
	DHCPStart       string          `json:"DHCPStart,omitempty"`
	DHCPEnd         string          `json:"DHCPEnd,omitempty"`
	Comment         string          `json:"Comment,omitempty"`
	MetalLBPoolName string          `json:"MetalLBPoolName,omitempty"`
}

// ReservationsByName presents the IPReservations in a map by name
func (subnet *IPV4Subnet) ReservationsByName() map[string]IPReservation {
	reservations := make(map[string]IPReservation)
	for _, v := range subnet.IPReservations {
		reservations[v.Name] = v
	}
	return reservations
}
