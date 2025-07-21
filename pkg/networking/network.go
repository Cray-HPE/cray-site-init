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

package networking

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"slices"
	"sort"
	"strings"

	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/sls"
	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/spf13/viper"
)

// VLANs accounts for all used VLANs during a run of cray-site-init.
var VLANs = [MaxVLAN]bool{MaxUsableVLAN: true}

// IPNetwork is a type for managing IP Networks.
type IPNetwork struct {
	FullName           string                `yaml:"full_name"`
	CIDR4              string                `yaml:"cidr4"`
	CIDR6              string                `yaml:"cidr6,omitempty"`
	Subnets            []*slsCommon.IPSubnet `yaml:"subnets"`
	Name               string                `yaml:"name"`
	VlanRange          []int16               `yaml:"vlan_range"`
	MTU                int16                 `yaml:"mtu"`
	NetType            slsCommon.NetworkType `yaml:"type"`
	Comment            string                `yaml:"comment"`
	PeerASN            int                   `yaml:"peer-asn"`
	MyASN              int                   `yaml:"my-asn"`
	SystemDefaultRoute string                `yaml:"system_default_route"`
	ParentDevice       string                `yaml:"parent-device"`
	PITServer          string                `yaml:"pit-server"`
	DNSServer          string                `yaml:"dns-server"`
}

type IPNetworks []*IPNetwork

type NetworkMap map[string]*IPNetwork

func (s IPNetworks) Len() int {
	return len(s)
}

func (s IPNetworks) Less(i, j int) bool {
	return s[i].FullName < s[j].FullName
}

func (s IPNetworks) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// IPRange defines a pair of IPs, over a range.
type IPRange struct {
	start netip.Addr
	end   netip.Addr
}

// PinnedReservation is a simple struct to work with our abomination of a PinnedMetalLBReservations.
type PinnedReservation struct {
	IPByte  uint8
	Aliases []string
}

// IPNets is a helper type for sorting net.IPNets.
type IPNets []netip.Prefix

func (s IPNets) Len() int {
	return len(s)
}

func (s IPNets) Less(i, j int) bool {
	return s[i].Addr().Less(s[j].Addr())
}

func (s IPNets) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

/*
IsVLANAllocated takes an int16 and tests if a given VLAN is already allocated and managed.
If the VLAN is above our MaxUsableVLAN or below the MinVLAN an error is returned.
*/
func IsVLANAllocated(vlan uint16) (bool, error) {
	if vlan > MaxUsableVLAN || vlan < MinVLAN {
		return true, errors.New("VLAN out of range")
	}
	return VLANs[vlan], nil
}

// AllocateVLAN takes an uint16 and manages a single VLAN.
func AllocateVLAN(vlan uint16) error {
	allocated, err := IsVLANAllocated(vlan)
	if allocated {
		if err != nil {
			return err
		}
		return errors.New("VLAN already used")
	}
	VLANs[vlan] = true
	return nil
}

// FreeVLAN frees a given VLAN by setting its allocation status to "false".
func FreeVLAN(vlan uint16) {
	VLANs[vlan] = false
}

// freeVLANRange is strictly for expediting tests. This will free a chunk of VLANs
func freeVLANRange(startVLAN uint16, endVLAN uint16) (err error) {
	if startVLAN > endVLAN {
		return fmt.Errorf(
			"VLAN range is bad - start is larger than end (%d !> %d)",
			startVLAN,
			endVLAN,
		)
	}
	for vlan := startVLAN; vlan <= endVLAN; vlan++ {
		FreeVLAN(vlan)
	}
	return err
}

// AllocateVlanRange takes two int16 and manages a range of VLANs.
func AllocateVlanRange(startVLAN int16, endVLAN int16) error {
	if startVLAN > endVLAN {
		return errors.New("VLAN range is bad - start is larger than end")
	}
	// Pre-test all VLANs for previous allocation
	hasAllocationErrors := false
	var allocatedVlans []uint16
	for vlan := uint16(startVLAN); vlan <= uint16(endVLAN); vlan++ {
		allocated, err := IsVLANAllocated(vlan)
		if err != nil {
			return err
		}
		if allocated {
			hasAllocationErrors = true
			allocatedVlans = append(
				allocatedVlans,
				vlan,
			)
		}
	}
	if hasAllocationErrors {
		return fmt.Errorf(
			"VLANs already used: %v",
			allocatedVlans,
		)
	}

	for vlan := uint16(startVLAN); vlan <= uint16(endVLAN); vlan++ {
		err := AllocateVLAN(vlan)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
SupernetSubnets is a list of subnets that should report the real subnet mask of their parent network (instead of their
own SLS subnet mask).

See "ApplySupernetHack"
*/
var SupernetSubnets = []string{
	"bootstrap_dhcp",
	"network_hardware",
	"can_metallb_static_pool",
	"can_metallb_address_pool",
}

/*
ApplySupernetHack
The IP allocation code in this application derives subnets within each network, these subnets prevent overlapping IP
reservations (e.g. network_hardware will use a segment of IPs before servers, and servers use a different segment).

Example:

The metal network is 10.1.0.0/16, this app will divide it into /24 chunks for each group of hardware:

  - network switches get 10.1.0.0/24
  - NCNs get 10.1.1.0/24

In the end, these /24 subnets do not exist in the network configuration.

This is where the "supernet" comes in. This function goes through each subnet's datastruct and updates the CIDR
to match that of the parent network.

Example:

- network switches were allocated IPs in 10.1.0.0/24 but their CIDR now shows 10.1.0.0/16
- NCNs were allocated IPs in 10.1.1.0/24 but their CIDR now shows 10.1.1.0/16

This allows us to reflect the real network that those devices live in as well as their starting address,.
*/
func (network *IPNetwork) ApplySupernetHack() {
	net4, err := netip.ParsePrefix(network.CIDR4)
	if err != nil && network.CIDR4 != "" {
		log.Fatalf(
			"couldn't parse the IPv4 CIDR for %s %s because %v",
			network.Name,
			network.CIDR4,
			err,
		)
	}
	gw4, err := network.FindGatewayIP(net4)
	if err != nil && net4.IsValid() {
		log.Fatalf(
			"couldn't find the IPv4 gateway for %s %s because %v",
			network.Name,
			network.CIDR4,
			err,
		)
	}

	net6, err := netip.ParsePrefix(network.CIDR6)
	if err != nil && network.CIDR6 != "" {
		log.Fatalf(
			"couldn't parse the IPv6 CIDR for %s %s because %v",
			network.Name,
			network.CIDR6,
			err,
		)
	}
	gw6, err := network.FindGatewayIP(net6)
	if err != nil && net6.IsValid() {
		log.Fatalf(
			"couldn't find the IPv6 gateway for %s %s because %v",
			network.Name,
			network.CIDR6,
			err,
		)
	}
	for _, subnetName := range SupernetSubnets {

		subnet, err := network.LookUpSubnet(subnetName)
		if err != nil {
			continue
		}

		if net4.IsValid() && gw4.IsValid() {
			subnet.Gateway = gw4.AsSlice()
			prefix4, err := netip.ParsePrefix(subnet.CIDR)
			if err != nil {
				if !net4.IsValid() {
					log.Fatalf(
						"Failed to parse the subnet prefix for %s because %s",
						subnetName,
						err,
					)
				}
			}
			subnet.CIDR = netip.PrefixFrom(
				prefix4.Addr(),
				net4.Bits(),
			).String()
		}

		// Since IPv6 networks are optional, we'll only set the override if they exist (without failing otherwise).
		if net6.IsValid() && gw6.IsValid() {
			subnet.Gateway6 = gw6.AsSlice()
			prefix6, err := netip.ParsePrefix(subnet.CIDR6)
			if err == nil {
				if !net6.IsValid() {
					log.Fatalf(
						"subnet [%s] had a IPv6 CIDR set [%s] but its parent network [%s] did not!",
						subnetName,
						subnet.CIDR6,
						network.Name,
					)
				}
				subnet.CIDR6 = netip.PrefixFrom(
					prefix6.Addr(),
					net6.Bits(),
				).String()
			}
		}
	}
}

// GenSubnets subdivides a network into a set of subnets
func (network *IPNetwork) GenSubnets(
	cabinetDetails []sls.CabinetGroupDetail, cidr net.IPMask, cabinetFilter sls.CabinetFilterFunc,
) error {
	networkPrefix, err := netip.ParsePrefix(network.CIDR4)
	if err != nil {
		return err
	}
	subnets := network.AllocatedIPv4Subnets()
	networkSubnets := network.Subnets
	var minVlan, maxVlan int16 = MaxUsableVLAN, MinVLAN

	// IPv6 subnetting.
	var prefix6 netip.Prefix
	if network.CIDR6 != "" {
		prefix6, err = netip.ParsePrefix(network.CIDR6)
		if err != nil {
			log.Printf(
				"Network %s had an unparseable IPv6 CIDR set: %s",
				network.Name,
				err,
			)
		}
	}

	for _, cabinetDetail := range cabinetDetails {
		for j, i := range cabinetDetail.CabinetDetails {
			if cabinetFilter(
				cabinetDetail,
				i,
			) {
				newSubnet, err := free(
					networkPrefix,
					cidr,
					subnets,
				)
				if err != nil {
					return fmt.Errorf(
						"couldn't add subnet because %v",
						err,
					)
				}
				subnets = append(
					subnets,
					newSubnet,
				)
				var tmpVlanID int16
				if strings.HasPrefix(
					strings.ToUpper(network.Name),
					"NMN",
				) {
					tmpVlanID = i.NMNVlanID
				}
				if strings.HasPrefix(
					strings.ToUpper(network.Name),
					"HMN",
				) {
					tmpVlanID = i.HMNVlanID
				}
				if tmpVlanID == 0 {
					tmpVlanID = int16(j) + network.VlanRange[0]
				}
				tempSubnet := slsCommon.IPSubnet{
					CIDR: newSubnet.String(),
					Name: fmt.Sprintf(
						"cabinet_%d",
						i.ID,
					),
					Gateway: newSubnet.Addr().Next().AsSlice(),
					VlanID:  tmpVlanID,
				}

				err = UpdateDHCPRange(
					&tempSubnet,
					false,
				)
				if err != nil {
					err = fmt.Errorf(
						"couldn't update DHCP range for %s because %v",
						tempSubnet.Name,
						err,
					)
					return err
				}

				// IPv6 subnetting.
				if prefix6.IsValid() {
					tempSubnet.CIDR6 = prefix6.String()
					tempSubnet.Gateway6 = prefix6.Addr().Next().AsSlice()
				}

				// Add the new subnet and move the VLANs along.
				networkSubnets = append(
					networkSubnets,
					&tempSubnet,
				)
				if tmpVlanID < minVlan {
					minVlan = tmpVlanID
				}
				if tmpVlanID > maxVlan {
					maxVlan = tmpVlanID
				}
			}
		}
	}
	network.VlanRange[0] = minVlan
	network.VlanRange[1] = maxVlan
	network.Subnets = networkSubnets
	return err
}

// AllocatedIPv4Subnets returns a list of the allocated IPv4 CIDRs.
func (network *IPNetwork) AllocatedIPv4Subnets() (subnets []netip.Prefix) {
	for _, v := range network.Subnets {
		prefix, err := netip.ParsePrefix(v.CIDR)
		if err != nil {
			log.Fatalf(
				"Failed to parse CIDR for %s because %v \n",
				v.Name,
				err,
			)
		}
		subnets = append(
			subnets,
			prefix,
		)
	}
	return subnets
}

// AllocatedIPv6Subnets returns a list of the allocated IPv6 CIDRs.
func (network *IPNetwork) AllocatedIPv6Subnets() (subnets []netip.Prefix) {
	for _, v := range network.Subnets {
		if v.CIDR6 == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(v.CIDR6)
		if err != nil {
			log.Fatalf(
				"Failed to parse CIDR6 for %s because %v",
				v.Name,
				err,
			)
		}
		subnets = append(
			subnets,
			prefix,
		)
	}
	return subnets
}

// AllocatedVLANs returns a list of all allocated vlan ids.
func (network *IPNetwork) AllocatedVLANs() (vlans []int16) {
	for _, v := range network.Subnets {
		if v.VlanID > 0 {
			vlans = append(
				vlans,
				v.VlanID,
			)
		}
	}
	return vlans
}

/*
CreateSubnetByCIDR returns an SLS subnet based on the given CIDRs.
*/
func (network *IPNetwork) CreateSubnetByCIDR(
	subnetPrefix netip.Prefix, name string, vlanID int16, andAppend bool,
) (
	subnet *slsCommon.IPSubnet, err error,
) {
	if !subnetPrefix.IsValid() {
		err = fmt.Errorf("invalid subnet prefix given to CreateSubnetByCIDR")
		return subnet, err
	}
	networkPrefix, err := netip.ParsePrefix(network.CIDR4)
	if err != nil {
		return subnet, err
	}
	if ContainsSubnet(
		networkPrefix,
		subnetPrefix,
	) {
		var newSubnet = slsCommon.IPSubnet{
			Name:   name,
			VlanID: vlanID,
		}
		subnet = &newSubnet
		network.SetSubnetIP(
			subnet,
			subnetPrefix,
		)
	} else {
		err = fmt.Errorf(
			"subnet %v is not part of %v",
			subnetPrefix.String(),
			networkPrefix.String(),
		)
		return subnet, err
	}
	if andAppend {
		return network.AppendSubnet(subnet), err
	}
	return subnet, err
}

/*
SetSubnetIP sets the CIDR and Gateway in a given subnet object.
Works for IPv4 and IPv6.
*/
func (network *IPNetwork) SetSubnetIP(subnet *slsCommon.IPSubnet, prefix netip.Prefix) {
	if !prefix.IsValid() {
		return
	}
	gateway := FindGatewayIP(prefix)
	if prefix.Addr().Is4() {
		subnet.CIDR = prefix.String()
		subnet.Gateway = gateway.AsSlice()
	} else if prefix.Addr().Is6() {
		subnet.CIDR6 = prefix.String()
		subnet.Gateway6 = gateway.AsSlice()
	}
}

/*
CreateSubnetByMask
Creates a viable SLS subnet based on the given net.IPMasks that fits within the network and appends that subnet to the
network's list of subnets. The created subnet is returned.
*/
func (network *IPNetwork) CreateSubnetByMask(
	v4mask net.IPMask, v6mask net.IPMask, name string, vlanID int16,
) (
	subnet *slsCommon.IPSubnet, err error,
) {
	var newSubnet = slsCommon.IPSubnet{
		FullName: network.Name,
		Name:     name,
		VlanID:   vlanID,
	}
	subnet = &newSubnet
	if v4mask != nil {
		var prefix4 netip.Prefix
		prefix4, err = netip.ParsePrefix(network.CIDR4)
		if err != nil {
			return subnet, fmt.Errorf(
				"couldn't parse %s CIDR %s because %v",
				network.Name,
				network.CIDR4,
				err,
			)
		}
		var freeIPv4Subnet netip.Prefix
		freeIPv4Subnet, err = free(
			prefix4,
			v4mask,
			network.AllocatedIPv4Subnets(),
		)
		if err == nil {
			network.SetSubnetIP(
				subnet,
				freeIPv4Subnet,
			)
		}
	}

	if v6mask != nil {
		var prefix6 netip.Prefix
		prefix6, err = netip.ParsePrefix(network.CIDR6)
		if err != nil {
			return subnet, fmt.Errorf(
				"couldn't parse %s CIDR6 %s because %v",
				network.Name,
				network.CIDR6,
				err,
			)
		}
		var freeIPv6Subnet netip.Prefix
		freeIPv6Subnet, err = free(
			prefix6,
			v6mask,
			network.AllocatedIPv6Subnets(),
		)
		if err == nil {
			network.SetSubnetIP(
				subnet,
				freeIPv6Subnet,
			)
		}
	}
	return network.AppendSubnet(subnet), err
}

func (network *IPNetwork) AppendSubnet(subnet *slsCommon.IPSubnet) *slsCommon.IPSubnet {
	network.Subnets = append(
		network.Subnets,
		subnet,
	)
	return network.Subnets[len(network.Subnets)-1]
}

// FindBiggestIPv4Subnet finds the largest, free IPv4 subnet on the network that for the given mask.
func (network *IPNetwork) FindBiggestIPv4Subnet(
	mask net.IPMask,
) (
	prefix netip.Prefix, err error,
) {
	if network.CIDR4 == "" {
		return prefix, fmt.Errorf(
			"no IPv4 definition found for %v",
			network.Name,
		)
	}
	if mask == nil {
		return
	}
	maskSize, _ := mask.Size()
	networkPrefix, err := netip.ParsePrefix(network.CIDR4)
	if err != nil {
		return prefix, fmt.Errorf(
			"couldn't parse %s CIDR [%s] because %v",
			network.Name,
			network.CIDR4,
			err,
		)
	}
	allocatedSubnets := network.AllocatedIPv4Subnets()
	for i := maskSize; i < SmallestIPv4Block; i++ {
		newMask := net.CIDRMask(
			i,
			IPv4Size,
		)
		prefix, err = free(
			networkPrefix,
			newMask,
			allocatedSubnets,
		)
		if err == nil {
			return prefix, err
		}
	}
	return prefix, fmt.Errorf(
		"no room for %v subnet within %v (tried from /%d to /%d)",
		network.Name,
		network.CIDR4,
		maskSize,
		SmallestIPv4Block,
	)
}

// FindBiggestIPv6Subnet finds the largest, free IPv6 subnet on the network that for the given mask.
func (network *IPNetwork) FindBiggestIPv6Subnet(
	mask net.IPMask,
) (
	prefix netip.Prefix, err error,
) {
	if network.CIDR6 == "" {
		return prefix, fmt.Errorf(
			"no IPv6 definition found for %v",
			network.Name,
		)
	}
	if mask == nil {
		return
	}
	maskSize, _ := mask.Size()
	networkPrefix, err := netip.ParsePrefix(network.CIDR6)
	if err != nil {
		return prefix, fmt.Errorf(
			"couldn't parse %s CIDR6 [%s] because %v",
			network.Name,
			network.CIDR6,
			err,
		)
	}
	allocatedSubnets := network.AllocatedIPv6Subnets()
	for i := maskSize; i < SmallestIPv6Block; i++ {
		newMask := net.CIDRMask(
			i,
			IPv6Size,
		)
		prefix, err = free(
			networkPrefix,
			newMask,
			allocatedSubnets,
		)
		if err == nil {
			return prefix, err
		}
	}
	return prefix, fmt.Errorf(
		"no room for %v subnet within %v (tried from /%d to /%d)",
		network.Name,
		network.CIDR6,
		maskSize,
		SmallestIPv6Block,
	)
}

// LookUpSubnet returns a subnet by name
func (network *IPNetwork) LookUpSubnet(name string) (
	subnet *slsCommon.IPSubnet, err error,
) {
	if len(network.Subnets) == 0 {
		return subnet, fmt.Errorf(
			"no subnets defined for %s - failed to lookup %s",
			network.Name,
			name,
		)
	}
	for _, v := range network.Subnets {
		if v.Name == name {
			subnet = v
			return subnet, err
		}
	}
	return subnet, fmt.Errorf(
		"subnet not found \"%v\"",
		name,
	)
}

// SubnetWithin returns the smallest subnet than can contain (size) hosts
func (network *IPNetwork) SubnetWithin(desiredHosts uint64) (smallestSubnet4 netip.Prefix, smallestSubnet6 netip.Prefix, errors []error) {
	prefix4, err := netip.ParsePrefix(network.CIDR4)
	if err == nil {
		smallestSubnet4, err = smallestIPv4SubnetWithin(
			prefix4,
			desiredHosts,
		)
		if err != nil {
			errors = append(
				errors,
				err,
			)
		}
	}
	prefix6, err := netip.ParsePrefix(network.CIDR6)
	if err == nil {
		smallestSubnet6, err = smallestIPv6SubnetWithin(
			prefix6,
			desiredHosts,
		)
		if err != nil {
			errors = append(
				errors,
				err,
			)
		}
	}
	return smallestSubnet4, smallestSubnet6, errors
}

func smallestIPv4SubnetWithin(prefix netip.Prefix, desiredHosts uint64) (smallestSubnet netip.Prefix, err error) {
	if !prefix.IsValid() {
		err = fmt.Errorf(
			"invalid network: %s",
			prefix.String(),
		)
		return prefix, err
	}
	if prefix.IsSingleIP() {
		err = fmt.Errorf(
			"network [%s] is of only one IP address with 0 available",
			prefix.String(),
		)
		return prefix, err
	}
	if !prefix.Addr().Is4() {
		err = fmt.Errorf(
			"network [%s] is not IPv4 address",
			prefix.String(),
		)
		return prefix, err
	}
	var smallestPrefix int
	for i := prefix.Bits(); i < int(IPv4Size); i++ {
		smallestSubnet = netip.PrefixFrom(
			prefix.Addr(),
			i,
		)
		usableHosts, _ := UsableHostAddresses(smallestSubnet)
		if usableHosts > desiredHosts {
			smallestPrefix = i
		}
	}
	smallestSubnet = netip.PrefixFrom(
		prefix.Addr(),
		smallestPrefix,
	)
	return smallestSubnet, err
}

func smallestIPv6SubnetWithin(prefix netip.Prefix, desiredHosts uint64) (smallestSubnet netip.Prefix, err error) {
	if !prefix.IsValid() {
		err = fmt.Errorf(
			"invalid network: %s",
			prefix.String(),
		)
		return prefix, err
	}
	if prefix.IsSingleIP() {
		err = fmt.Errorf(
			"network [%s] is of only one IP address with 0 available",
			prefix.String(),
		)
		return prefix, err
	}
	if !prefix.Addr().Is6() {
		err = fmt.Errorf(
			"network [%s] is not IPv6 address",
			prefix.String(),
		)
		return prefix, err
	}
	var smallestPrefix int
	for i := prefix.Bits(); i < int(IPv6Size); i++ {
		smallestSubnet = netip.PrefixFrom(
			prefix.Addr(),
			i,
		)
		usableHosts, _ := UsableHostAddresses(smallestSubnet)
		if usableHosts > desiredHosts {
			smallestPrefix = i
		}
	}
	smallestSubnet = netip.PrefixFrom(
		prefix.Addr(),
		smallestPrefix,
	)
	return smallestSubnet, err
}

// SubnetByName Return a copy of the subnet by name or a blank subnet if it doesn't exist.
func (network *IPNetwork) SubnetByName(name string) slsCommon.IPSubnet {
	for _, v := range network.Subnets {
		if strings.EqualFold(
			v.Name,
			name,
		) {
			return *v
		}
	}
	return slsCommon.IPSubnet{}
}

// ReserveEdgeSwitchIPs reserves (n) IP addresses for edge switches
func ReserveEdgeSwitchIPs(subnet *slsCommon.IPSubnet, edges []string) (err error) {
	for i := 0; i < len(edges); i++ {
		name := fmt.Sprintf(
			"chn-switch-%01d",
			i+1,
		)

		_, err := AddReservation(
			subnet,
			name,
			edges[i],
		)
		if err != nil {
			return err
		}
	}
	return err
}

// ReserveNetMgmtIPs reserves (n) IP addresses for management networking equipment
func ReserveNetMgmtIPs(
	subnet *slsCommon.IPSubnet, spines []string, leafs []string, leafbmcs []string, cdus []string,
) (err error) {
	for i := 0; i < len(spines); i++ {
		name := fmt.Sprintf(
			"sw-spine-%03d",
			i+1,
		)
		_, err := AddReservation(
			subnet,
			name,
			spines[i],
		)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(leafs); i++ {
		name := fmt.Sprintf(
			"sw-leaf-%03d",
			i+1,
		)
		_, err := AddReservation(
			subnet,
			name,
			leafs[i],
		)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(leafbmcs); i++ {
		name := fmt.Sprintf(
			"sw-leaf-bmc-%03d",
			i+1,
		)
		_, err := AddReservation(
			subnet,
			name,
			leafbmcs[i],
		)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(cdus); i++ {
		name := fmt.Sprintf(
			"sw-cdu-%03d",
			i+1,
		)
		_, err := AddReservation(
			subnet,
			name,
			cdus[i],
		)
		if err != nil {
			return err
		}
	}
	return err
}

// UpdateDHCPRange resets the DHCPStart and DHCPEnd to exclude all IPReservations.
func UpdateDHCPRange(subnet *slsCommon.IPSubnet, applySupernetHack bool) (err error) {

	myReservedIPs := subnet.ReservedIPs()
	subnetPrefix, _ := netip.ParsePrefix(subnet.CIDR)
	usable, err := UsableHostAddresses(subnetPrefix)
	if err != nil {
		log.Printf(
			"Error checking for usable addresses in subnet %s: %v",
			subnet.CIDR,
			err,
		)
		return
	}
	if uint64(len(myReservedIPs)) > usable {
		return fmt.Errorf(
			"could not create %s subnet in %s. There are %d reservations and only %d usable ip addresses in the subnet %v",
			subnet.FullName,
			subnet.Name,
			len(myReservedIPs),
			usable,
			subnet.CIDR,
		)
	}

	// Bump the DHCP Start IP past the gateway
	// At least ten IPs are needed, but more if required
	staticLimit := Add(
		subnetPrefix,
		10,
	)
	dynamicLimit := Add(
		subnetPrefix,
		uint64(len(subnet.IPReservations)+2),
	)
	if dynamicLimit.Compare(staticLimit) == -1 {
		if subnet.Name == "uai_macvlan" {
			subnet.ReservationStart = staticLimit.AsSlice()
		} else {
			subnet.DHCPStart = staticLimit.AsSlice()
		}
	} else {
		if subnet.Name == "uai_macvlan" {
			subnet.ReservationStart = dynamicLimit.AsSlice()
		} else {
			subnet.DHCPStart = dynamicLimit.AsSlice()
		}
	}

	if applySupernetHack {
		if subnet.Name == "uai_macvlan" {
			dhcpStartAddr, err := netip.ParseAddr(subnet.DHCPStart.String())
			if err != nil {
				return fmt.Errorf(
					"error parsing DHCP start address in subnet for supernethack %s: %v",
					subnetPrefix,
					err,
				)
			}
			reservationEnd := dhcpStartAddr.AsSlice()
			subnet.ReservationEnd = reservationEnd
		} else {
			dhcpStartAddr, err := netip.ParseAddr(subnet.DHCPStart.String())
			if err != nil {
				return fmt.Errorf(
					"error parsing DHCP end address in subnet for supernethack %s: %v",
					subnetPrefix,
					err,
				)
			}
			dhcpEndAddr := Add(
				netip.PrefixFrom(
					dhcpStartAddr,
					subnetPrefix.Bits(),
				),
				200,
			)
			subnet.DHCPEnd = dhcpEndAddr.AsSlice()
		}
	} else {
		broadcast, err := Broadcast(subnetPrefix)
		if err != nil {
			return fmt.Errorf(
				"error obtaining broadcast address for subnet %s: %v",
				subnetPrefix,
				err,
			)
		}
		if subnet.Name == "uai_macvlan" {
			subnet.ReservationEnd = broadcast.Prev().AsSlice()
		} else {
			subnet.DHCPEnd = broadcast.Prev().AsSlice()
		}
	}
	return err
}

// GenInterfaceName generates the network interface name for a subnet.
func (network *IPNetwork) GenInterfaceName(subnet *slsCommon.IPSubnet) (interfaceName string, err error) {
	if len(subnet.Name) > 15 {
		err = fmt.Errorf(
			"network name [%s] is greater than 15 bytes",
			subnet.Name,
		)
	}
	if subnet.Name == "" {
		err = fmt.Errorf(
			"network name [%s] is empty/nil",
			subnet.Name,
		)
	}

	if subnet.VlanID <= FirstVLAN {
		interfaceName = network.ParentDevice
	} else {
		index := 0 // In the future, if we ever need to change this we can handle it with a loop and setting a range in IPSubnet.
		interfaceName = fmt.Sprintf(
			"%s.%s%d",
			network.ParentDevice,
			strings.ToLower(network.Name),
			index,
		)
	}
	return interfaceName, err
}

/*
free
takes a network, a mask, and a list of subnets.
An available network, within the first network, is returned.
*/
func free(network netip.Prefix, mask net.IPMask, subnets []netip.Prefix) (freeNetwork netip.Prefix, err error) {

	maskOnes, _ := mask.Size()
	networkCapacity, _ := UsableHostAddresses(network)
	maskPrefix, _ := netip.ParsePrefix(mask.String())
	maskCapacity, _ := UsableHostAddresses(maskPrefix)
	if networkCapacity < maskCapacity {
		return freeNetwork, fmt.Errorf(
			"prefix was %s, mask requested did not fit /%v (bit)",
			network.String(),
			maskOnes,
		)
	}

	for _, subnet := range subnets {
		if !network.Contains(subnet.Addr()) {
			return freeNetwork, fmt.Errorf(
				"%v is not contained by %v",
				subnet.String(),
				network.String(),
			)
		}
	}

	sort.Sort(IPNets(subnets))

	freeIPRanges, err := freeIPRanges(
		network,
		subnets,
	)

	if err != nil {
		return freeNetwork, fmt.Errorf(
			"failed to find any free IP ranges: %v",
			err,
		)
	}
	// Attempt to find a free space, of the required size.
	freeNetwork, err = space(
		freeIPRanges,
		mask,
	)
	if err != nil {
		return freeNetwork, fmt.Errorf(
			"failed to find any free space in network [%v]. Error: %v",
			network.String(),
			err,
		)
	}

	// Invariant: The IP of the network returned should be contained
	// within the network supplied.
	if !network.Contains(freeNetwork.Addr()) {
		return freeNetwork, fmt.Errorf(
			"%v is not contained by %v",
			freeNetwork.Addr().String(),
			network,
		)
	}

	// Invariant: The mask of the network returned should be equal to
	// the mask supplied as an argument.
	if freeNetwork.Bits() != maskOnes {
		return freeNetwork, fmt.Errorf(
			"have: %v, requested: %v",
			freeNetwork.Bits(),
			mask,
		)
	}

	return freeNetwork, err
}

// freeIPRanges takes a network, and a list of subnets.
// It calculates available IPRanges, within the original network.
func freeIPRanges(network netip.Prefix, subnets []netip.Prefix) (freeSubnets []IPRange, err error) {
	networkRange := newIPRange(network)

	// If no subnets, return the entire network range.
	if len(subnets) == 0 {
		freeSubnets = append(
			freeSubnets,
			networkRange,
		)
		return freeSubnets, err
	}

	{
		// Check space between start of network and first subnet.
		firstSubnetRange := newIPRange(subnets[0])
		// Check if we have a free-range between the network's start and the start of the first subnet.
		if networkRange.start != firstSubnetRange.start {
			freeSubnets = append(
				freeSubnets,
				IPRange{
					start: networkRange.start,
					end:   firstSubnetRange.start.Prev(),
				},
			)
		}
	}

	{
		// Check space between each subnet.
		for i := 0; i < len(subnets)-1; i++ {
			currentSubnetRange := newIPRange(subnets[i])
			nextSubnetRange := newIPRange(subnets[i+1])
			// If the two subnets are not contiguous, then there is a free-range between them.
			if currentSubnetRange.end.Next().Less(nextSubnetRange.start.Prev()) {
				freeSubnets = append(
					freeSubnets,
					IPRange{
						start: currentSubnetRange.end.Next(),
						end:   nextSubnetRange.start.Prev(),
					},
				)
			}
		}
	}

	{
		// Check space between last subnet and end of network.
		lastSubnetRange := newIPRange(subnets[len(subnets)-1])
		// Check the last subnet doesn't end at the end of the network.
		if lastSubnetRange.end.Less(networkRange.end) {
			// It doesn't, so we have a free-range between the end of the
			// last subnet, and the end of the network.
			freeSubnets = append(
				freeSubnets,
				IPRange{
					start: lastSubnetRange.end.Next(),
					end:   networkRange.end,
				},
			)
		}
	}

	return freeSubnets, nil
}

func newIPRange(network netip.Prefix) (iprange IPRange) {
	usableIPs, err := UsableHostAddresses(network)
	if err != nil {
		return iprange
	}

	// UsableHostAddresses will return "2^(bits - prefix) - 2" for IPv4, automatically subtracting the root and broadcast
	// address. However, for newIPRange we want to include these. We already include the root address via our returned
	// iprange, but we need to add one for the would-be broadcast address.
	if network.Addr().Is4() {
		usableIPs++
	}
	endAddress := Add(
		network,
		usableIPs,
	)
	startAddress, err := FindCIDRRootIP(network)
	if err != nil {
		log.Fatalf(
			"failed to find subnet root IP: %v",
			err,
		)
	}
	if startAddress == endAddress {
		return iprange
	}
	iprange = IPRange{
		start: startAddress,
		end:   endAddress,
	}
	return iprange
}

// space takes a list of free ip ranges, and a mask, returning the first eligible block that could hold any IPs.
func space(freeIPRanges []IPRange, mask net.IPMask) (firstFree netip.Prefix, err error) {

	var start netip.Addr
	prefixLength, _ := mask.Size()
	for _, freeIPRange := range freeIPRanges {
		start = freeIPRange.start
		end := freeIPRange.end
		for start.Less(end) {
			prefix := netip.PrefixFrom(
				start,
				prefixLength,
			)
			if !prefix.IsValid() {
				return firstFree, fmt.Errorf(
					"no blocks found for %s - %s",
					freeIPRange.start,
					end,
				)
			}
			subnetIP, err := FindCIDRRootIP(prefix)
			if err != nil {
				return firstFree, err
			}

			/*
				When subnet allocations contain various different subnet sizes, it can be
				that free IP range starts from smaller network than what we are finding
				for. Therefore, we must first adjust the start IP such that it can hold the
				whole network that we are looking space for.

				Example: free IP range starts at 10.1.2.192 and ends 10.1.255.255, and
				we're looking for the first available /24.

				We look for next available /24 network so first suitable
				start IP for this would be 10.1.3.0.
			*/
			if subnetIP.Compare(start) == -1 {
				broadcast, err := Broadcast(prefix)
				if err != nil {
					return prefix, fmt.Errorf(
						"failed to find broadcast address for %s because %v",
						prefix.String(),
						err,
					)
				}
				start = broadcast.Next()
			} else {
				firstFree = netip.PrefixFrom(
					subnetIP,
					prefixLength,
				)
				break
			}
		}

		if firstFree.Bits() == -1 {
			continue
		}

		// Check that our firstFree network fits our desired mask's addresses.
		startBinary, _ := start.MarshalBinary()
		endBinary, _ := end.MarshalBinary()
		var resultBinary []byte
		for i, startOctet := range startBinary {
			endOctet := endBinary[i]
			resultBinary = append(
				resultBinary,
				endOctet-startOctet,
			)
		}
		resultAddr, _ := netip.AddrFromSlice(resultBinary)
		resultPrefix := netip.PrefixFrom(
			resultAddr,
			prefixLength,
		)
		usableHosts, err := UsableHostAddresses(resultPrefix)
		if err != nil {
			return firstFree, err
		}
		maxHosts, err := UsableHostAddresses(
			netip.PrefixFrom(
				start,
				prefixLength,
			),
		)
		if err != nil {
			return firstFree, err
		}
		if usableHosts >= maxHosts {
			return firstFree, err
		}
	}
	err = fmt.Errorf(
		"tried to fit a /%v",
		prefixLength,
	)
	return firstFree, err
}

/*
CheckCIDROverlap takes a list of CIDR flags, and verifies if their input values from the command line overlap
with any other of the given input values.

Requires a list of flags to check.
*/
func CheckCIDROverlap(cidrFlags []string) (errors []error) {
	v := viper.GetViper()

	for _, cidrFlag := range cidrFlags {
		cidr, err := netip.ParsePrefix(v.GetString(cidrFlag))
		if err != nil {
			continue
		}
		var compareCIDRFlagName string
		var compareCIDR netip.Prefix
		hasOverlap := slices.ContainsFunc(
			cidrFlags,
			func(flag string) bool {

				// Ignore the flag we're actively comparing to.
				if flag == cidrFlag {
					return false
				}

				compareCIDRFlagName = flag
				compareCIDR, err = netip.ParsePrefix(v.GetString(flag))
				if err != nil || !compareCIDR.IsValid() {
					return false
				}
				return cidr.Overlaps(compareCIDR)
			},
		)
		if hasOverlap {
			errors = append(
				errors,
				fmt.Errorf(
					"%-10s [%s] overlaps with %-10s [%s]",
					cidrFlag,
					cidr.String(),
					compareCIDRFlagName,
					compareCIDR.String(),
				),
			)
		}
	}
	return errors
}
