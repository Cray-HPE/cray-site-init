package sls

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
