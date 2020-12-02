/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

import (
	"fmt"
	"log"
	"net"

	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	"stash.us.cray.com/MTL/csi/pkg/ipam"
)

// ManagementSwitchBrand known list of Management switch brands
type ManagementSwitchBrand string

// ManagementSwitchBrandAruba for Aruba Management switches
const ManagementSwitchBrandAruba ManagementSwitchBrand = "Aruba"

// ManagementSwitchBrandDell for Dell Management switches
const ManagementSwitchBrandDell ManagementSwitchBrand = "Dell"

// ManagementSwitchBrandMellanox for Mellanox Management switches
const ManagementSwitchBrandMellanox ManagementSwitchBrand = "Mellanox"

func (msb ManagementSwitchBrand) String() string {
	return string(msb)
}

// ManagementSwitchType the type of management switch CDU/Leaf/Spine/Aggregation
type ManagementSwitchType string

// ManagementSwitchTypeCDU is the type for CDU Management switches
const ManagementSwitchTypeCDU ManagementSwitchType = "CDU"

// ManagementSwitchTypeLeaf is the type for Leaf Management switches
const ManagementSwitchTypeLeaf ManagementSwitchType = "Leaf"

// ManagementSwitchTypeSpine is the type for Spine Management switches
const ManagementSwitchTypeSpine ManagementSwitchType = "Spine"

// ManagementSwitchTypeAggregation is the type for Aggregation Management switches
const ManagementSwitchTypeAggregation ManagementSwitchType = "Aggregation"

// IsManagementSwitchTypeValid validates the given ManagementSwitchType
func IsManagementSwitchTypeValid(mst ManagementSwitchType) bool {
	switch mst {
	case ManagementSwitchTypeAggregation:
		fallthrough
	case ManagementSwitchTypeCDU:
		fallthrough
	case ManagementSwitchTypeLeaf:
		fallthrough
	case ManagementSwitchTypeSpine:
		return true
	}

	return false
}

// IPReservation is a type for managing IP Reservations
type IPReservation struct {
	IPAddress net.IP   `yaml:"ip_address"`
	Name      string   `yaml:"name"`
	Comment   string   `yaml:"comment"`
	Aliases   []string `yaml:"aliases"`
}

// IPV4Network is a type for managing IPv4 Networks
type IPV4Network struct {
	FullName  string                 `yaml:"full_name"`
	CIDR      string                 `yaml:"cidr"`
	Subnets   []*IPV4Subnet          `yaml:"subnets"`
	Name      string                 `yaml:"name"`
	VlanRange []int16                `yaml:"vlan_range"`
	MTU       int16                  `yaml:"mtu"`
	NetType   sls_common.NetworkType `yaml:"type"`
	Comment   string                 `yaml:"comment"`
}

// IPV4Subnet is a type for managing IPv4 Subnets
type IPV4Subnet struct {
	FullName       string          `yaml:"full_name" form:"full_name" mapstructure:"full_name"`
	CIDR           net.IPNet       `yaml:"cidr"`
	IPReservations []IPReservation `yaml:"ip_reservations"`
	Name           string          `yaml:"name" form:"name" mapstructure:"name"`
	NetName        string          `yaml:"net-name"`
	VlanID         int16           `yaml:"vlan_id" form:"vlan_id" mapstructure:"vlan_id"`
	Comment        string          `yaml:"comment"`
	Gateway        net.IP          `yaml:"gateway"`
	DHCPStart      net.IP          `yaml:"iprange-start"`
	DHCPEnd        net.IP          `yaml:"iprange-end"`
}

// ManagementSwitch is a type for managing Management switches
type ManagementSwitch struct {
	Xname               string                `json:"xname" mapstructure:"xname" csv:"Switch Xname"` // Required for SLS
	Name                string                `json:"name" mapstructure:"name" csv:"-"`              // Required for SLS to update DNS
	Brand               ManagementSwitchBrand `json:"brand" mapstructure:"brand" csv:"Brand"`
	Model               string                `json:"model" mapstructure:"model" csv:"Model"`
	Os                  string                `json:"operating-system" mapstructure:"operating-system" csv:"-"`
	Firmware            string                `json:"firmware" mapstructure:"firmware" csv:"-"`
	SwitchType          ManagementSwitchType  `json:"type" mapstructure:"type" csv:"Type"` //"CDU/Leaf/Spine/Aggregation"
	ManagementInterface net.IP                `json:"ip" mapstructure:"ip" csv:"-"`        // SNMP/REST interface IP (not a distinct BMC)  // Required for SLS
}

// GenSubnets subdivides a network into a set of subnets
func (iNet *IPV4Network) GenSubnets(cabinetDetails []CabinetDetail, mask net.IPMask) error {

	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	mySubnets := iNet.AllocatedSubnets()
	myIPv4Subnets := iNet.Subnets

	for _, cabinetDetail := range cabinetDetails {

		for i := 0; i < int(cabinetDetail.Cabinets); i++ {
			newSubnet, err := ipam.Free(*myNet, mask, mySubnets)
			mySubnets = append(mySubnets, newSubnet)
			if err != nil {
				log.Printf("Gensubnets couldn't add subnet because %v \n", err)
				return err
			}
			tempSubnet := IPV4Subnet{
				CIDR:    newSubnet,
				Name:    fmt.Sprintf("cabinet_%v", cabinetDetail.StartingCabinet+i),
				Gateway: ipam.Add(newSubnet.IP, 1),
				// Reserving the first vlan in the range for a non-cabinet aligned vlan if needed in the future.
				VlanID: iNet.VlanRange[1] + int16(i),
			}
			// Bump the DHCP Start IP past the gateway
			tempSubnet.DHCPStart = ipam.Add(tempSubnet.CIDR.IP, len(tempSubnet.IPReservations)+2)
			tempSubnet.DHCPEnd = ipam.Add(ipam.Broadcast(tempSubnet.CIDR), -1)
			myIPv4Subnets = append(myIPv4Subnets, &tempSubnet)
		}
	}
	iNet.Subnets = myIPv4Subnets
	return nil
}

// AllocatedSubnets returns a list of the allocated subnets
func (iNet IPV4Network) AllocatedSubnets() []net.IPNet {
	var myNets []net.IPNet
	for _, v := range iNet.Subnets {
		myNets = append(myNets, v.CIDR)
	}
	return myNets
}

// AddSubnetbyCIDR allocates a new subnet
func (iNet *IPV4Network) AddSubnetbyCIDR(desiredNet net.IPNet, name string, vlanID int16) (*IPV4Subnet, error) {
	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	if ipam.Contains(*myNet, desiredNet) {
		iNet.Subnets = append(iNet.Subnets, &IPV4Subnet{
			CIDR:    desiredNet,
			Name:    name,
			Gateway: ipam.Add(desiredNet.IP, 1),
			VlanID:  vlanID,
		})
		return iNet.Subnets[len(iNet.Subnets)-1], nil
	}
	return &IPV4Subnet{}, fmt.Errorf("subnet %v is not part of %v", desiredNet.String(), myNet.String())
}

// AddSubnet allocates a new subnet
func (iNet *IPV4Network) AddSubnet(mask net.IPMask, name string, vlanID int16) (*IPV4Subnet, error) {
	var tempSubnet IPV4Subnet
	_, myNet, _ := net.ParseCIDR(iNet.CIDR)
	newSubnet, err := ipam.Free(*myNet, mask, iNet.AllocatedSubnets())
	if err != nil {
		return &tempSubnet, err
	}
	iNet.Subnets = append(iNet.Subnets, &IPV4Subnet{
		CIDR:    newSubnet,
		Name:    name,
		NetName: iNet.Name,
		Gateway: ipam.Add(newSubnet.IP, 1),
		VlanID:  vlanID,
	})
	return iNet.Subnets[len(iNet.Subnets)-1], nil
}

// AddBiggestSubnet allocates the largest subnet possible within the requested network and mask
func (iNet *IPV4Network) AddBiggestSubnet(mask net.IPMask, name string, vlanID int16) (*IPV4Subnet, error) {
	// Try for the largest available and go smaller if needed
	maskSize, _ := mask.Size() // the second output of this function is 32 for ipv4 or 64 for ipv6
	for i := maskSize; i < 29; i++ {
		// log.Printf("Trying to find room for a /%d mask in %v \n", i, iNet.Name)
		newSubnet, err := iNet.AddSubnet(net.CIDRMask(i, 32), name, vlanID)
		if err == nil {
			return newSubnet, nil
		}
	}
	return &IPV4Subnet{}, fmt.Errorf("no room for %v subnet within %v (tried from /%d to /29)", name, iNet.Name, maskSize)
}

// LookUpSubnet returns a subnet by name
func (iNet *IPV4Network) LookUpSubnet(name string) (*IPV4Subnet, error) {
	var found []*IPV4Subnet
	if len(iNet.Subnets) == 0 {
		return &IPV4Subnet{}, fmt.Errorf("subnet not found %v", name)
	}
	for _, v := range iNet.Subnets {
		if v.Name == name {
			found = append(found, v)
		}
	}
	if len(found) == 1 {
		return found[0], nil
	}
	if len(found) > 1 {
		log.Printf("Found %v subnets named %v in the %v network instead of just one \n", len(found), name, iNet.Name)
		return found[0], fmt.Errorf("found %v subnets instead of just one", len(found))
	}
	return &IPV4Subnet{}, fmt.Errorf("subnet not found %v", name)
}

// SubnetbyName Return a copy of the subnet by name or a blank subnet if it doesn't exists
func (iNet IPV4Network) SubnetbyName(name string) IPV4Subnet {
	for _, v := range iNet.Subnets {
		if v.Name == name {
			return *v
		}
	}
	return IPV4Subnet{}
}

// ReserveNetMgmtIPs reserves (n) IP addresses for management networking equipment
func (iSubnet *IPV4Subnet) ReserveNetMgmtIPs(spines []string, leafs []string, aggs []string, cdus []string, additional int) {
	for i := 0; i < len(spines); i++ {
		name := fmt.Sprintf("sw-spine-%03d", i+1)
		iSubnet.AddReservation(name, spines[i])
	}
	for i := 0; i < len(leafs); i++ {
		name := fmt.Sprintf("sw-leaf-%03d", i+1)
		iSubnet.AddReservation(name, leafs[i])
	}
	for i := 0; i < len(aggs); i++ {
		name := fmt.Sprintf("sw-agg-%03d", i+1)
		iSubnet.AddReservation(name, aggs[i])
	}
	for i := 0; i < len(cdus); i++ {
		name := fmt.Sprintf("sw-cdu-%03d", i+1)
		iSubnet.AddReservation(name, cdus[i])
	}
	for i := 0; i < additional; i++ {
		name := fmt.Sprintf("mgmt-net-stub-%03d", i+1)
		iSubnet.AddReservation(name, "")
	}
}

// ReservedIPs returns a list of IPs already reserved within the subnet
func (iSubnet *IPV4Subnet) ReservedIPs() []net.IP {
	var addresses []net.IP
	for _, v := range iSubnet.IPReservations {
		addresses = append(addresses, v.IPAddress)
	}
	return addresses
}

// ReservationsByName presents the IPReservations in a map by name
func (iSubnet *IPV4Subnet) ReservationsByName() map[string]IPReservation {
	reservations := make(map[string]IPReservation)
	for _, v := range iSubnet.IPReservations {
		reservations[v.Name] = v
	}
	return reservations
}

// UpdateDHCPRange resets the DHCPStart to exclude all IPReservations
func (iSubnet *IPV4Subnet) UpdateDHCPRange() {
	myReservedIPs := iSubnet.ReservedIPs()
	ip := ipam.Add(iSubnet.CIDR.IP, len(myReservedIPs)+2)
	iSubnet.DHCPStart = ip
	// log.Printf("Inside UpdateDHCPRange and ip = %v which is at %v in list\n", ip, netIPInSlice(ip, myReservedIPs))
	for ipam.NetIPInSlice(ip, myReservedIPs) > 0 {
		iSubnet.DHCPStart = ipam.Add(ip, 2)
		log.Printf("Dealing with DHCPStart as %v \n", iSubnet.DHCPStart)
		ip = ipam.Add(ip, 1)
	}
	iSubnet.DHCPEnd = ipam.Add(ipam.Broadcast(iSubnet.CIDR), -1)
}

// AddReservation adds a new IP reservation to the subnet
func (iSubnet *IPV4Subnet) AddReservation(name, comment string) *IPReservation {
	myReservedIPs := iSubnet.ReservedIPs()
	floor := iSubnet.CIDR.IP.Mask(iSubnet.CIDR.Mask)
	if !floor.Equal(iSubnet.CIDR.IP) {
		log.Printf("VERY BAD - In reservation. CIDR.IP = %v and floor is %v", iSubnet.CIDR.IP.String(), floor)
	}
	// Start counting from the bottom knowing the gateway is on the bottom
	tempIP := ipam.Add(iSubnet.CIDR.IP, 2)
	for {
		for _, v := range myReservedIPs {
			if tempIP.Equal(v) {
				tempIP = ipam.Add(tempIP, 1)
			}
		}
		iSubnet.IPReservations = append(iSubnet.IPReservations, IPReservation{
			IPAddress: tempIP,
			Name:      name,
			Comment:   comment,
		})
		return &iSubnet.IPReservations[len(iSubnet.IPReservations)-1]
	}

}

// NetworkLayoutConfiguration is the internal configuration structure for shasta networks
type NetworkLayoutConfiguration struct {
	Template                        IPV4Network
	ReservationHostnames            []string
	IncludeBootstrapDHCP            bool
	DesiredBootstrapDHCPMask        net.IPMask
	IncludeNetworkingHardwareSubnet bool
	AdditionalNetworkingSpace       int
	NetworkingHardwareNetmask       net.IPMask
	BaseVlan                        int16
	SubdivideByCabinet              bool
	IncludeUAISubnet                bool
	CabinetDetails                  []CabinetDetail
	CabinetCIDR                     net.IPMask
	ManagementSwitches              []*ManagementSwitch
}

// IsValid provides feedback about any problems with the configuration
func (nlc *NetworkLayoutConfiguration) IsValid() (bool, error) {
	if nlc.IncludeNetworkingHardwareSubnet {
		if len(nlc.ManagementSwitches) < 1 {
			return false, fmt.Errorf("can't build networking hardware subnets without ManagementSwitches")
		}
	}
	if nlc.SubdivideByCabinet {
		if len(nlc.CabinetDetails) < 1 {
			return false, fmt.Errorf("can't build per cabinet subnets without a list of cabinet details")
		}
	}
	return true, nil
}

// GenLayoutConfiguration creates a configuration from a default template and booleans
func GenLayoutConfiguration(template IPV4Network, IncludeBootstrapDHCP bool, IncludeNetworkingHardware bool, SubdivideByCabinet bool) NetworkLayoutConfiguration {
	return NetworkLayoutConfiguration{
		Template:                        template,
		IncludeBootstrapDHCP:            IncludeBootstrapDHCP,
		IncludeNetworkingHardwareSubnet: IncludeBootstrapDHCP,
		SubdivideByCabinet:              SubdivideByCabinet,
	}
}
