/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

// Package ipam provides IP address management functionality.
// Mostly a copy/paste from https://github.com/giantswarm/ipam without prioprietary service/error handling
package ipam

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"net"
	"reflect"
	"sort"
)

// IPRange defines a pair of IPs, over a range.
type IPRange struct {
	start net.IP
	end   net.IP
}

// IPNets is a helper type for sorting net.IPNets.
type IPNets []net.IPNet

func (s IPNets) Len() int {
	return len(s)
}

func (s IPNets) Less(i, j int) bool {
	return ipToDecimal(s[i].IP) < ipToDecimal(s[j].IP)
}

func (s IPNets) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// CalculateSubnetMask calculates new subnet mask to accommodate n subnets.
func CalculateSubnetMask(networkMask net.IPMask, n uint) (net.IPMask, error) {
	if n == 0 {
		return nil, errors.New("divide by zero")
	}

	// Calculate amount of bits needed to accommodate at least N subnets.
	subnetBitsNeeded := bits.Len(n - 1)

	maskOnes, maskBits := networkMask.Size()
	if subnetBitsNeeded > maskBits-maskOnes {
		return nil, fmt.Errorf("no room in network mask %s to accommodate %d subnets", networkMask.String(), n)
	}

	return net.CIDRMask(maskOnes+subnetBitsNeeded, maskBits), nil
}

// CanonicalizeSubnets iterates over subnets and returns deduplicated list of
// networks that belong to networkRange. Subnets that overlap each other but
// aren't exactly the same are not removed. Subnets are returned in the same
// order as they appear in input.
//
// Example:
//	  networkRange: 192.168.2.0/24
//	  subnets: [172.168.2.0/25, 192.168.2.0/25, 192.168.3.128/25, 192.168.2.0/25, 192.168.2.128/25]
//	  returned: [192.168.2.0/25, 192.168.2.128/25]
//
// Example 2:
//	  networkRange: 10.0.0.0/8
//	  subnets: [10.1.0.0/16, 10.1.0.0/24, 10.1.1.0/24]
//	  returned: [10.1.0.0/16, 10.1.0.0/24, 10.1.1.0/24]
//
func CanonicalizeSubnets(networkRange net.IPNet, subnets []net.IPNet) []net.IPNet {
	// Naive deduplication as net.IPNet cannot be used as key for map. This
	// should be ok for current foreseeable future.
	for i := 0; i < len(subnets); i++ {
		// Remove subnets that don't belong to our desired network.
		if !networkRange.Contains(subnets[i].IP) {
			subnets = append(subnets[:i], subnets[i+1:]...)
			i--
			continue
		}

		// Remove duplicates.
		for j := i + 1; j < len(subnets); j++ {
			if reflect.DeepEqual(subnets[i], subnets[j]) {
				subnets = append(subnets[:j], subnets[j+1:]...)
				j--
			}
		}
	}

	return subnets
}

// Contains returns true when the subnet is a part of the network, false
// otherwise.
func Contains(network, subnet net.IPNet) bool {
	subnetRange := NewIPRange(subnet)
	return network.Contains(subnetRange.start) && network.Contains(subnetRange.end)
}

// Free takes a network, a mask, and a list of subnets.
// An available network, within the first network, is returned.
func Free(network net.IPNet, mask net.IPMask, subnets []net.IPNet) (net.IPNet, error) {
	if size(network.Mask) < size(mask) {
		return net.IPNet{},
			fmt.Errorf("have: %v, requested: %v", network.Mask, mask)
	}

	for _, subnet := range subnets {
		if !network.Contains(subnet.IP) {
			return net.IPNet{},
				fmt.Errorf("%v is not contained by %v", subnet.IP, network)
		}
	}

	sort.Sort(IPNets(subnets))

	// Find all the free IP ranges.
	freeIPRanges, err := freeIPRanges(network, subnets)
	if err != nil {
		return net.IPNet{}, err
	}

	// Attempt to find a free space, of the required size.
	freeIP, err := space(freeIPRanges, mask)
	if err != nil {
		return net.IPNet{}, err
	}

	// Invariant: The IP of the network returned should not be nil.
	if freeIP == nil {
		return net.IPNet{}, errors.New("no available ips")
	}

	freeNetwork := net.IPNet{IP: freeIP, Mask: mask}

	// Invariant: The IP of the network returned should be contained
	// within the network supplied.
	if !network.Contains(freeNetwork.IP) {
		return net.IPNet{},
			fmt.Errorf("%v is not contained by %v", freeNetwork.IP, network)
	}

	// Invariant: The mask of the network returned should be equal to
	// the mask supplied as an argument.
	if !bytes.Equal(mask, freeNetwork.Mask) {
		return net.IPNet{},
			fmt.Errorf("have: %v, requested: %v", freeNetwork.Mask, mask)
	}

	return freeNetwork, nil
}

// Half takes a network and returns two subnets which split the network in
// half.
func Half(network net.IPNet) (first, second net.IPNet, err error) {
	ones, parts := network.Mask.Size()
	if ones == parts {
		return net.IPNet{}, net.IPNet{}, fmt.Errorf("single IP mask %q is not allowed", network.Mask.String())
	}

	// Bit shift is dividing by 2.
	ones++
	mask := net.CIDRMask(ones, parts)

	// Compute first half.
	first, err = Free(network, mask, nil)
	if err != nil {
		return net.IPNet{}, net.IPNet{}, err
	}

	// Second half is computed by getting next free.
	second, err = Free(network, mask, []net.IPNet{first})
	if err != nil {
		return net.IPNet{}, net.IPNet{}, err
	}

	return first, second, nil
}

// Split returns n subnets from network.
func Split(network net.IPNet, n uint) ([]net.IPNet, error) {
	mask, err := CalculateSubnetMask(network.Mask, n)
	if err != nil {
		return nil, err
	}

	var subnets []net.IPNet
	for i := uint(0); i < n; i++ {
		subnet, err := Free(network, mask, subnets)
		if err != nil {
			return nil, err
		}

		subnets = append(subnets, subnet)
	}

	return subnets, nil
}

// Add increments the given IP by the number.
// e.g: add(10.0.4.0, 1) -> 10.0.4.1.
// Negative values are allowed for decrementing.
func Add(ip net.IP, number int) net.IP {
	return decimalToIP(ipToDecimal(ip) + number)
}

// decimalToIP converts an int to a net.IP.
func decimalToIP(ip int) net.IP {
	t := make(net.IP, 4)
	binary.BigEndian.PutUint32(t, uint32(ip))

	return t
}

// freeIPRanges takes a network, and a list of subnets.
// It calculates available IPRanges, within the original network.
func freeIPRanges(network net.IPNet, subnets []net.IPNet) ([]IPRange, error) {
	freeSubnets := []IPRange{}
	networkRange := NewIPRange(network)

	if len(subnets) == 0 {
		freeSubnets = append(freeSubnets, networkRange)
		return freeSubnets, nil
	}

	{
		// Check space between start of network and first subnet.
		firstSubnetRange := NewIPRange(subnets[0])

		// Check the first subnet doesn't start at the start of the network.
		if !networkRange.start.Equal(firstSubnetRange.start) {
			// It doesn't, so we have a free range between the start
			// of the network, and the start of the first subnet.
			end := Add(firstSubnetRange.start, -1)
			freeSubnets = append(freeSubnets,
				IPRange{start: networkRange.start, end: end},
			)
		}
	}

	{
		// Check space between each subnet.
		for i := 0; i < len(subnets)-1; i++ {
			currentSubnetRange := NewIPRange(subnets[i])
			nextSubnetRange := NewIPRange(subnets[i+1])

			// If the two subnets are not contiguous,
			if ipToDecimal(currentSubnetRange.end)+1 != ipToDecimal(nextSubnetRange.start) {
				// Then there is a free range between them.
				start := Add(currentSubnetRange.end, 1)
				end := Add(nextSubnetRange.start, -1)
				freeSubnets = append(freeSubnets, IPRange{start: start, end: end})
			}
		}
	}

	{
		// Check space between last subnet and end of network.
		lastSubnetRange := NewIPRange(subnets[len(subnets)-1])

		// Check the last subnet doesn't end at the end of the network.
		if !lastSubnetRange.end.Equal(networkRange.end) {
			// It doesn't, so we have a free range between the end of the
			// last subnet, and the end of the network.
			start := Add(lastSubnetRange.end, 1)
			freeSubnets = append(freeSubnets,
				IPRange{start: start, end: networkRange.end},
			)
		}
	}

	return freeSubnets, nil
}

// ipToDecimal converts a net.IP to an int.
func ipToDecimal(ip net.IP) int {
	t := ip
	if len(ip) == 16 {
		t = ip[12:16]
	}

	return int(binary.BigEndian.Uint32(t))
}

// NewIPRange takes an IPNet, and returns the ipRange of the network.
func NewIPRange(network net.IPNet) IPRange {
	start := network.IP
	end := Add(network.IP, size(network.Mask)-1)

	return IPRange{start: start, end: end}
}

// size takes a mask, and returns the number of addresses.
func size(mask net.IPMask) int {
	ones, _ := mask.Size()
	size := int(math.Pow(2, float64(32-ones)))

	return size
}

// Broadcast takes a net.IPNet and returns the broadcast address as net.IP
func Broadcast(network net.IPNet) net.IP {
	return Add(network.IP, size(network.Mask)-1)
}

// space takes a list of free ip ranges, and a mask,
// and returns the start IP of the first range that could fit the mask.
func space(freeIPRanges []IPRange, mask net.IPMask) (net.IP, error) {
	for _, freeIPRange := range freeIPRanges {
		start := ipToDecimal(freeIPRange.start)
		end := ipToDecimal(freeIPRange.end)

		// When subnet allocations contain various different subnet sizes, it can be
		// that free IP range starts from smaller network than what we are finding
		// for. Therefore we must first adjust the start IP such that it can hold the
		// whole network that we are looking space for.
		//
		// Example: Free IP range starts at 10.1.2.192 and ends 10.1.255.255.
		//          We look for next available /24 network so first suitable
		//          start IP for this would be 10.1.3.0.
		//
		ones, _ := mask.Size()
		trailingZeros := bits.TrailingZeros32(uint32(start))
		for (start < end) && (ones < (32 - trailingZeros)) {
			var mask uint32
			for i := 0; i < trailingZeros; i++ {
				mask |= 1 << uint32(i)
			}

			start = int(uint32(start) | mask)
			start++
			trailingZeros = bits.TrailingZeros32(uint32(start))
		}

		if end-start+1 >= size(mask) {
			return decimalToIP(start), nil
		}
	}

	return nil, fmt.Errorf("tried to fit: %v", mask)
}

// SubnetWithin returns the smallest subnet than can contain (size) hosts
func SubnetWithin(network net.IPNet, hostNumber int) (net.IPNet, error) {
	var n net.IPNet
	ip := network.IP.String()

	// sort the map
	keys := make([]int, 0)
	for k := range netmasks {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	// run through the sorted map
	for _, k := range keys {
		subnet := netmasks[k]
		if k > hostNumber {
			_, mynet, err := net.ParseCIDR(fmt.Sprintf("%v%v", ip, subnet))
			return *mynet, err
		}
	}
	return n, nil
}

// NetIPInSlice makes it easy to assess if an IP address is present in a list of ips
func NetIPInSlice(a net.IP, list []net.IP) int {
	for index, b := range list {
		if b.Equal(a) {
			return index
		}
	}
	return 0
}

// IPLessThan compare two ip addresses
// by section left-most is most significant
func IPLessThan(a, b net.IP) bool {
	for i := range a { // go left to right and compare each one
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false // they are equal
}

// Quick const map for mapping the number of hosts to the netmask shorthand
// I'm bad at math.
var netmasks = map[int]string{
	2:     "/30",
	6:     "/29",
	14:    "/28",
	30:    "/27",
	62:    "/26",
	126:   "/25",
	254:   "/24",
	510:   "/23",
	1022:  "/22",
	2046:  "/21",
	4094:  "/20",
	8190:  "/19",
	16382: "/18",
	32766: "/17",
	65534: "/16",
}
