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

// Package networking provides IP address management functionality.
// Mostly a copy/paste from https://github.com/giantswarm/ipam without proprietary service/error handling
package networking

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"net"
	"net/netip"
	"slices"
	"strings"

	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/spf13/viper"
)

const (

	// IPv4Size is the size of an IPv4 address in bits.
	IPv4Size = 32

	// IPv6Size is the size of an IPv6 address in bits.
	IPv6Size = 128

	// IPv6MinimumSize is the largest prefix that a network may be in bits.
	IPv6MinimumSize = 64
)

// ContainsSubnet returns true when the subnet is a part of the network, false
// otherwise.
func ContainsSubnet(network netip.Prefix, subnet netip.Prefix) bool {
	subnetRange := newIPRange(subnet)
	return network.Contains(subnetRange.start) && network.Contains(subnetRange.end)
}

// Add increments the given IP by the number.
//
// examples:
//
//		Add(10.0.4.0/24, 1) -> 10.0.4.1
//		Add(10.0.4.0/24, 300) -> 10.0.4.255
//	 Add(fdf8:413:de2c:204::/64, 18446744073709551615) -> fdf8:413:de2c:204:ffff:ffff:ffff:ffff
//
// Negative numbers are ignored.
func Add(prefix netip.Prefix, number uint64) (newIP netip.Addr) {
	if number < 1 {
		return prefix.Addr()
	}
	startingIP := prefix.Addr().AsSlice()
	ipInt := new(big.Int).SetBytes(startingIP)
	toAdd, ok := new(big.Int).SetString(
		fmt.Sprintf(
			"%0d",
			number,
		),
		10,
	)
	if !ok {
		log.Fatalf(
			"Error adding %s to IP address: %d",
			prefix.String(),
			number,
		)
	}
	ipInt.Add(
		ipInt,
		toAdd,
	)
	ipBytes := ipInt.Bytes()
	var resultIP net.IP
	if prefix.Addr().Is4() {
		resultIP = ipBytes
	} else if prefix.Addr().Is6() {
		if len(ipBytes) < net.IPv6len {
			paddedBytes := make(
				[]byte,
				net.IPv6len,
			)
			copy(
				paddedBytes[net.IPv6len-len(ipBytes):],
				ipBytes,
			)
			ipBytes = paddedBytes
		}
		resultIP = ipBytes
	}
	if resultIP == nil {
		log.Fatalf(
			"Failed to resolve IP address from adding %d to %s\n",
			number,
			prefix.String(),
		)
	}
	newIP, err := netip.ParseAddr(resultIP.String())
	if !prefix.Contains(newIP) {
		lastIP, err := Broadcast(prefix)
		if err == nil {
			log.Printf(
				"Tried adding %d to %s but the result was out-of-range.\n Returning %s\n",
				number,
				prefix.String(),
				lastIP,
			)
			newIP = lastIP
		}
	}
	if err != nil {
		log.Fatalf(
			"Failed to add %v to %v because %v.\n",
			number,
			prefix.String(),
			err,
		)
	}
	return newIP
}

/*
AddReservation
Adds and returns a new IPReservation to the given subnet. If an IPReservation exists in the given
subnet for the given name, the existing IPReservation is returned.

Both an IPv4 and IPv6 address will be reserved, IPv6 is only reserved if the given subnet has a valid IPv6 CIDR.
*/
func AddReservation(subnet *slsCommon.IPSubnet, name string, comment string) (IPReservation *slsCommon.IPReservation, err error) {
	ipv4, ipv6, err4, err6 := FindFreeIPAddress(subnet)
	if err4 != nil {
		return IPReservation, fmt.Errorf(
			"error finding a free IP address in %s (ipv4: %s, ipv6: %s) because %v",
			subnet.Name,
			subnet.CIDR,
			subnet.CIDR6,
			err4,
		)
	}
	if err6 != nil {
		if subnet.CIDR6 != "" {
			return IPReservation, fmt.Errorf(
				"error finding a free IP address in %s (ipv6: %s) because %v",
				subnet.Name,
				subnet.CIDR6,
				err6,
			)
		}
	}
	newIPReservation := slsCommon.IPReservation{
		Name:    name,
		Comment: comment,
	}
	if ipv4.IsValid() && !ipv4.IsUnspecified() {
		newIPReservation.IPAddress = ipv4.AsSlice()
	}
	if err6 == nil && ipv6.IsValid() && !ipv6.IsUnspecified() {
		newIPReservation.IPAddress6 = ipv6.AsSlice()
	}
	subnet.IPReservations = append(
		subnet.IPReservations,
		newIPReservation,
	)
	IPReservation = &subnet.IPReservations[len(subnet.IPReservations)-1]
	return IPReservation, err
}

/*
FindFreeIPAddress finds the first free, usable IP addresses in a given subnet.
Returns an IPv4 and IPv6 address, each with their own error.
*/
func FindFreeIPAddress(subnet *slsCommon.IPSubnet) (ipv4 netip.Addr, ipv6 netip.Addr, err4 error, err6 error) {
	IPAddresses, IPAddresses6 := getIPAddressesFromSubnet(subnet)
	ipv4, err4 = findFirstFreeIPv4Address(
		subnet,
		IPAddresses,
	)
	if err4 != nil && subnet.CIDR != "" {
		err4 = fmt.Errorf(
			"failed to find a free IPv4 address in %s because %v",
			subnet.Name,
			err4,
		)
	}
	ipv6, err6 = findFirstFreeIPv6Address(
		subnet,
		IPAddresses6,
	)
	if err6 != nil && subnet.CIDR6 != "" {
		err6 = fmt.Errorf(
			"failed to find a free IPv6 address in %s because %v",
			subnet.Name,
			err6,
		)
	}
	return ipv4, ipv6, err4, err6
}

func getIPAddressesFromSubnet(subnet *slsCommon.IPSubnet) (IPAddresses []netip.Addr, IPAddresses6 []netip.Addr) {
	for _, reservation := range subnet.IPReservations {
		IPReservation, err := netip.ParseAddr(reservation.IPAddress.String())
		if err == nil {
			IPAddresses = append(
				IPAddresses,
				IPReservation,
			)
		}
		IPReservation6, err := netip.ParseAddr(reservation.IPAddress6.String())
		if err == nil {
			IPAddresses6 = append(
				IPAddresses6,
				IPReservation6,
			)
		}
	}
	return IPAddresses, IPAddresses6
}

func findFirstFreeIPv4Address(subnet *slsCommon.IPSubnet, addresses []netip.Addr) (address netip.Addr, err error) {
	prefix, err := netip.ParsePrefix(subnet.CIDR)
	if err != nil {
		return address, fmt.Errorf(
			"error parsing CIDR '%s' because %v",
			subnet.CIDR,
			err,
		)
	}

	address = prefix.Addr()

	/*
		In IPv4, we will avoid leasing IPs to the
		- Root address (e.g. 192.168.0.0 for 192.168.0.0/24)
		- The gateway address (e.g. 192.168.0.1 for 192.168.0.0/24)
		- The broadcast address (e.g. 192.168.0.255 for 192.168.0.0/24)

		If we resolve an address that does not belong to the subnet the function returns with an error.
	*/

	subnetRoot, err := FindCIDRRootIP(prefix)
	if err != nil {
		return address, fmt.Errorf(
			"error finding subnet root IP address for %s subnet because %v",
			prefix.String(),
			err,
		)
	}
	addresses = append(
		addresses,
		subnetRoot,
	)
	gateway, err := netip.ParseAddr(subnet.Gateway.String())
	if err != nil {
		return address, fmt.Errorf(
			"error reading gateway IP for %s subnet because %v",
			subnet.Gateway.String(),
			err,
		)
	}
	addresses = append(
		addresses,
		gateway,
	)
	broadcast, err := Broadcast(prefix)
	if err != nil {
		return address, fmt.Errorf(
			"error resolving broadcast address for %s subnet because %v",
			prefix.String(),
			err,
		)
	}
	addresses = append(
		addresses,
		broadcast,
	)
	for slices.Contains(
		addresses,
		address,
	) {
		address = address.Next()
	}
	if !prefix.Contains(address) {
		address = netip.IPv4Unspecified()
		err = fmt.Errorf(
			"%s %s subnet has exhausted its available IPv4 addresses, failed to find a free address",
			subnet.Name,
			prefix.String(),
		)
	}
	return address, err
}

func findFirstFreeIPv6Address(subnet *slsCommon.IPSubnet, addresses []netip.Addr) (address netip.Addr, err error) {
	prefix, err := netip.ParsePrefix(subnet.CIDR6)
	if err != nil {
		return address, fmt.Errorf(
			"error parsing CIDR '%s': %s",
			subnet.CIDR6,
			err,
		)
	}

	address = prefix.Addr()

	/*
		In IPv6, we will avoid leasing IPs to the
		- Root address (e.g. 192.168.0.0 for 192.168.0.0/24)
		- The gateway address (e.g. 192.168.0.1 for 192.168.0.0/24)

		If we resolve an address that does not belong to the subnet the function returns with an error.
	*/
	subnetRoot, err := FindCIDRRootIP(prefix)
	addresses = append(
		addresses,
		subnetRoot,
	)
	if err != nil {
		return address, fmt.Errorf(
			"error finding subnet root IP address for %s subnet because %v",
			prefix.String(),
			err,
		)
	}
	gateway, err := netip.ParseAddr(subnet.Gateway6.String())
	if err != nil {
		return address, fmt.Errorf(
			"error resolving gateway address for %s subnet because %v",
			subnet.Gateway6.String(),
			err,
		)
	}
	addresses = append(
		addresses,
		gateway,
	)
	for slices.Contains(
		addresses,
		address,
	) {
		address = address.Next()
	}
	if !prefix.Contains(address) {
		address = netip.IPv6Unspecified()
		err = fmt.Errorf(
			"%s %s subnet has exhausted its available IPv6 addresses, failed to find a free address",
			subnet.Name,
			prefix.String(),
		)
	}
	return address, err
}

/*
UpdateReservation will modify an existing reservation. This is useful when the subnet's CIDR changes or new CIDRs

If IPv6Only is set to true, only IPSubnet.IPAddress6 values are updated. This is useful when IPv6 is being added to an
existing SLS subnet, and we do not want to change the IPv4 reservations.

If the subnet does not have IPv6 defined, then IPv6Only has no effect.
*/

func UpdateReservation(subnet *slsCommon.IPSubnet, IPReservation slsCommon.IPReservation, IPv6Only bool) (newIPReservation slsCommon.IPReservation, err error) {
	ipv4, ipv6, err4, err6 := FindFreeIPAddress(subnet)
	if err4 != nil {
		return newIPReservation, fmt.Errorf(
			"error finding a free IP address in %s because %v",
			subnet.Name,
			err,
		)
	}
	newIPReservation = IPReservation
	if !IPv6Only && ipv4.IsValid() && !ipv4.IsUnspecified() {
		newIPReservation.IPAddress = ipv4.AsSlice()
	}
	if err6 == nil && ipv6.IsValid() && !ipv6.IsUnspecified() {
		newIPReservation.IPAddress6 = ipv6.AsSlice()
	}
	return newIPReservation, err
}

// AddReservationWithIP adds a reservation with a specific ip address.
func AddReservationWithIP(subnet *slsCommon.IPSubnet, name string, addr netip.Addr, comment string) (
	reservation *slsCommon.IPReservation, err error,
) {
	if addr.Is4() {
		prefix, err := netip.ParsePrefix(subnet.CIDR)
		if err != nil {
			return nil, fmt.Errorf(
				"error parsing CIDR '%s': %s",
				subnet.CIDR,
				err,
			)
		}
		if prefix.Contains(addr) {
			newReservation := slsCommon.IPReservation{
				IPAddress: addr.AsSlice(),
				Name:      name,
				Comment:   comment,
			}
			for _, reservation := range subnet.IPReservations {
				if reservation.IPAddress.Equal(addr.AsSlice()) {
					return nil, fmt.Errorf(
						"failed to reserve IPv4 address for %s, address already reserved for %s",
						newReservation.Name,
						reservation.Name,
					)
				}
			}
			subnet.IPReservations = append(
				subnet.IPReservations,
				newReservation,
			)
		} else {
			return nil, fmt.Errorf(
				"cannot add \"%v\" to %v subnet as %v. %v is not part of %v",
				name,
				subnet.Name,
				addr,
				addr,
				prefix.String(),
			)
		}
	} else if addr.Is6() {
		prefix, err := netip.ParsePrefix(subnet.CIDR6)
		if err != nil {
			return nil, fmt.Errorf(
				"error parsing CIDR '%s': %s",
				subnet.CIDR6,
				err,
			)
		}
		if prefix.Contains(addr) {
			newReservation := slsCommon.IPReservation{
				IPAddress6: addr.AsSlice(),
				Name:       name,
				Comment:    comment,
			}
			for _, reservation := range subnet.IPReservations {
				if reservation.IPAddress6.Equal(addr.AsSlice()) {
					return nil, fmt.Errorf(
						"failed to reserve IPv4 address for %s, address already reserved for %s",
						newReservation.Name,
						reservation.Name,
					)
				}
			}
			subnet.IPReservations = append(
				subnet.IPReservations,
				newReservation,
			)
		} else {
			return nil, fmt.Errorf(
				"cannot add \"%v\" to %v subnet as %v. %v is not part of %v",
				name,
				subnet.Name,
				addr,
				addr,
				prefix.String(),
			)
		}
	}
	return &subnet.IPReservations[len(subnet.IPReservations)-1], nil
}

// AddReservationWithPin adds a new IPv4 reservation to the subnet with the last octet pinned.
func AddReservationWithPin(
	subnet *slsCommon.IPSubnet, name string, comment string, pin uint8,
) (*slsCommon.IPReservation, error) {
	// Grab the "floor" of the subnet and alter the last byte to match the pinned byte
	// modulo 4/16 bit ip addresses
	newIP := make(
		net.IP,
		4,
	)
	_, cidr, err := net.ParseCIDR(subnet.CIDR)
	if err != nil {
		return nil, err
	}
	if len(cidr.IP) == 4 {
		newIP[0] = cidr.IP[0]
		newIP[1] = cidr.IP[1]
		newIP[2] = cidr.IP[2]
		newIP[3] = pin
	}
	if len(cidr.IP) == 16 {
		newIP[0] = cidr.IP[12]
		newIP[1] = cidr.IP[13]
		newIP[2] = cidr.IP[14]
		newIP[3] = pin
	}
	if comment != "" {
		subnet.IPReservations = append(
			subnet.IPReservations,
			slsCommon.IPReservation{
				IPAddress: newIP,
				Name:      name,
				Comment:   comment,
				Aliases: strings.Split(
					comment,
					",",
				),
			},
		)
	} else {
		subnet.IPReservations = append(
			subnet.IPReservations,
			slsCommon.IPReservation{
				IPAddress: newIP,
				Name:      name,
			},
		)
	}
	return &subnet.IPReservations[len(subnet.IPReservations)-1], nil
}

// Broadcast returns the broadcast IP of a given subnet, or 0.0.0.0, :: if unspecified.
func Broadcast(subnet netip.Prefix) (broadcast netip.Addr, err error) {

	var bits int
	if subnet.Addr().Is4() {
		bits = IPv4Size
	} else if subnet.Addr().Is6() {
		bits = IPv6Size
	}
	mask := net.CIDRMask(
		subnet.Bits(),
		bits,
	)

	binarySubnetAddr, err := subnet.Addr().MarshalBinary()
	if err != nil {
		return broadcast, fmt.Errorf(
			"error serializing subnet address: %v",
			err,
		)
	}
	binarySubnetMask, err := netip.MustParseAddr(net.IP(mask).String()).MarshalBinary()
	if err != nil {
		return broadcast, fmt.Errorf(
			"error parsing mashaling binary address %v",
			err,
		)
	}

	var binaryBrdCast []byte
	for i, subnetOctet := range binarySubnetAddr {
		maskOctet := binarySubnetMask[i]
		binaryBrdCast = append(
			binaryBrdCast,
			subnetOctet|^maskOctet,
		)
	}
	broadcast, _ = netip.AddrFromSlice(binaryBrdCast)
	return broadcast, nil
}

// PrefixLengthToSubnetMask returns a dot-decimal subnet mask based on the given number of ones (CIDR prefix length)
// and the total number of bits in the address.
//
// examples:
//
//   - (IPv4) 24 ones in a 32-bit address (/24) returns 255.255.255.0
//   - (IPv6) 65 ones in a 128-bit address (/65) returns ffff:ffff:ffff:ffff:8000::
func PrefixLengthToSubnetMask(prefixLength int, bits int) (mask netip.Addr, err error) {
	cidrMask := net.CIDRMask(
		prefixLength,
		bits,
	)
	subnetMask := net.IP(cidrMask)
	mask, err = netip.ParseAddr(subnetMask.String())
	return mask, err
}

// FindCIDRRootIP returns the subnet's IP address for an IP in CIDR notation.
//
// examples:
//
//   - (IPv4) 10.120.234.45/27 returns 10.120.234.32
//   - (IPv6) fdf8:413:de2c:205::223/64 returns fdf8:413:de2c:205::
func FindCIDRRootIP(prefix netip.Prefix) (addr netip.Addr, err error) {
	var subnetMask netip.Addr
	if prefix.Addr().Is4() {
		subnetMask, err = PrefixLengthToSubnetMask(
			prefix.Bits(),
			IPv4Size,
		)
		if err != nil {
			return addr, err
		}
	} else if prefix.Addr().Is6() {
		subnetMask, err = PrefixLengthToSubnetMask(
			prefix.Bits(),
			IPv6Size,
		)
		if err != nil {
			return addr, err
		}
	}
	binaryPrefixAddr, err := prefix.Addr().MarshalBinary()
	if err != nil {
		return addr, err
	}
	binarySubnet, err := subnetMask.MarshalBinary()
	if err != nil {
		return addr, err
	}
	var binarySubnetAddr []byte
	for i, subnetOctet := range binarySubnet {
		binarySubnetAddr = append(
			binarySubnetAddr,
			subnetOctet&binaryPrefixAddr[i],
		)
	}
	addr, _ = netip.AddrFromSlice(binarySubnetAddr)
	return addr, err
}

/*
FindGatewayIP looks for the first usable IP in a netip.Prefix. NOTE: this is
a GENERIC function, it does not pay any concession to the program's parameters.
In other words, if --can-gateway is set on the command line, this function should still
return the gateway based on the same logic. That way we can use this function for any
network (even those without a --gateway parameter), and all we need to run the function is a
netip.Prefix.

Networks that have their own --gateway parameter handle that separately throughout the program.
*/
func FindGatewayIP(prefix netip.Prefix) (gateway netip.Addr) {
	// If our subnet started at the beginning of the network, bump the IP by one.
	subnetIP, err := FindCIDRRootIP(prefix)
	if err != nil {
		log.Fatalf(
			"error finding root IP for %s because %v",
			prefix.String(),
			err,
		)
	}
	gateway = subnetIP.Next()
	return gateway
}

/*
FindGatewayIP returns the specified gateway for the network if one was given. Otherwise, a gateway is assumed
by finding the first IP in the network (using FindGatewayIP).
*/
func (network *IPNetwork) FindGatewayIP(prefix netip.Prefix) (gateway netip.Addr, err error) {
	if network.CIDR4 != prefix.String() && network.CIDR6 != prefix.String() {
		return gateway, fmt.Errorf(
			"network gateway resolution was called on non-contained network prefix: %s",
			prefix.String(),
		)
	}
	v := viper.GetViper()
	var gatewayString string
	if prefix.Addr().Is4() {
		gw4Key := fmt.Sprintf(
			"%s-gateway4",
			strings.ToLower(network.Name),
		)
		gwKey := fmt.Sprintf(
			"%s-gateway",
			strings.ToLower(network.Name),
		)
		if v.IsSet(gw4Key) {
			gatewayString = v.GetString(gw4Key)
		} else if v.IsSet(gw4Key) {
			gatewayString = v.GetString(gwKey)
		}

	} else if prefix.Addr().Is6() {
		gw6Key := fmt.Sprintf(
			"%s-gateway6",
			strings.ToLower(network.Name),
		)
		if v.IsSet(gw6Key) {
			gatewayString = v.GetString(gw6Key)
		}
	}
	if gatewayString != "" {
		gateway, err = netip.ParseAddr(gatewayString)
		if err != nil {
			gateway = FindGatewayIP(prefix)
			log.Printf(
				"WARNING: %s had a Gateway flag present but an invalid value for %s\nDefaulting to %s\n",
				prefix.String(),
				gatewayString,
				gateway.String(),
			)
		}
	} else {
		gateway = FindGatewayIP(prefix)
	}
	return gateway, err
}

/*
UsableHostAddresses returns the number of usable addresses for a given CIDR. This does not take into account
the IP of the CIDR (e.g. 10.0.0.100/24 is treated as 10.0.0.0/24).

For IPv4, the "usable" addresses starts at the root and excludes the broadcast address.
of the subnet (e.g. 192.168.0.0/24 the resulting number of addresses would be 254 ((2 ** (32 - 24)) - 2), we subtract two,
one to move the index to 0, and another for excluding the broadcast address.

For IPv6, the "usable" addresses can be very large, we accommodate this in two ways:

  - We use the newer big.Int interface for handling large numbers
  - We ignore masks smaller than 64 bits, a /63 or /20 will return the result for a /64.

For example, fdf8:413:de2c:204::/64 will return 18446744073709551615 (2 ** (128 - 64) - 1), we subtract to move our index to 0.
*/
func UsableHostAddresses(prefix netip.Prefix) (numHosts uint64, err error) {
	var maxBits int
	if prefix.Addr().Is4() {
		// var hosts float64 // can't be uint32, since 2^32 is 1 more than uint32.
		maxBits = IPv4Size
		if prefix.Bits() == IPv4Size-1 || prefix.Bits() == IPv4Size {
			numHosts = 0
		} else {
			hostBits := maxBits - prefix.Bits()
			// Subtract 2 because our IP addresses index at 0 and we need to exclude the broadcast IP on IPv4.
			numHosts = uint64(
				math.Pow(
					2,
					float64(hostBits),
				) - 2,
			)
		}
	} else if prefix.Addr().Is6() {
		maxBits = IPv6Size
		var tempNum = big.NewInt(2)
		var prefixBits int

		// 2^64 is our largest subnet, anything more is just multiples of 2^64 (e.g. /63 is two subnets of 2^64).
		if prefix.Bits() < IPv6MinimumSize {
			prefixBits = IPv6MinimumSize
		} else {
			prefixBits = prefix.Bits()
		}
		hostBits := maxBits - prefixBits
		tempHosts := tempNum.Exp(
			big.NewInt(2),
			big.NewInt(int64(hostBits)),
			nil,
		)
		tempHosts = tempHosts.Sub(
			tempNum,
			big.NewInt(1),
		)
		numHosts = tempHosts.Uint64()
	}
	return numHosts, err
}
