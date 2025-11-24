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

package sls

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"sort"
	"strings"

	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/pkg/csm"
	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

// NetworkLayoutConfiguration is the internal configuration structure for shasta networks
type NetworkLayoutConfiguration struct {
	Template                        networking.IPNetwork
	ReservationHostnames            []string
	IncludeBootstrapDHCP            bool
	DesiredBootstrapDHCPMask        net.IPMask
	IncludeNetworkingHardwareSubnet bool
	SuperNetHack                    bool
	AdditionalNetworkingSpace       int
	NetworkingHardwareNetmask       net.IPMask
	BaseVlan                        int16
	SubdivideByCabinet              bool
	GroupNetworksByCabinetType      bool
	IncludeUAISubnet                bool
	CabinetDetails                  []sls.CabinetGroupDetail
	CabinetCIDR                     net.IPMask
	ManagementSwitches              []*networking.ManagementSwitch
}

// GenDefaultBICANConfig returns the set of defaults for mapping the BICAN toggle
func GenDefaultBICANConfig(systemDefaultRoute string) NetworkLayoutConfiguration {

	networking.DefaultBICAN.SystemDefaultRoute = systemDefaultRoute
	return NetworkLayoutConfiguration{
		Template:                        networking.DefaultBICAN,
		SubdivideByCabinet:              false,
		IncludeBootstrapDHCP:            false,
		IncludeNetworkingHardwareSubnet: false,
		IncludeUAISubnet:                false,
	}
}

// GenDefaultHMNConfig is the set of defaults for mapping the HMN
func GenDefaultHMNConfig() NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        networking.DefaultHMN,
		SubdivideByCabinet:              false,
		GroupNetworksByCabinetType:      true,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: true,
		SuperNetHack:                    true,
		IncludeUAISubnet:                false,
		CabinetCIDR:                     networking.DefaultCabinetMask,
		NetworkingHardwareNetmask:       networking.DefaultNetworkingHardwareMask,
	}
}

// GenDefaultNMNConfig returns the set of defaults for mapping the NMN
func GenDefaultNMNConfig() NetworkLayoutConfiguration {
	return NetworkLayoutConfiguration{
		Template:                        networking.DefaultNMN,
		SubdivideByCabinet:              false,
		GroupNetworksByCabinetType:      true,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: true,
		SuperNetHack:                    true,
		IncludeUAISubnet:                true,
		CabinetCIDR:                     networking.DefaultCabinetMask,
		NetworkingHardwareNetmask:       networking.DefaultNetworkingHardwareMask,
	}
}

// GenDefaultHSNConfig returns the set of defaults for mapping the HSN
func GenDefaultHSNConfig() NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        networking.DefaultHSN,
		SubdivideByCabinet:              false,
		IncludeBootstrapDHCP:            false,
		IncludeNetworkingHardwareSubnet: false,
		IncludeUAISubnet:                false,
	}
}

// GenDefaultCMNConfig returns the set of defaults for mapping the CMN
func GenDefaultCMNConfig(ncns int, switches int) NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        networking.DefaultCMN,
		SubdivideByCabinet:              false,
		SuperNetHack:                    true,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: true,
		IncludeUAISubnet:                false,
	}
}

// GenDefaultCANConfig returns the set of defaults for mapping the CAN
func GenDefaultCANConfig() NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        networking.DefaultCAN,
		SubdivideByCabinet:              false,
		SuperNetHack:                    false,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: false,
		IncludeUAISubnet:                false,
	}
}

// GenDefaultCHNConfig returns the set of defaults for mapping the CHN
func GenDefaultCHNConfig() NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        networking.DefaultCHN,
		SubdivideByCabinet:              false,
		SuperNetHack:                    false,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: false,
		IncludeUAISubnet:                false,
	}
}

// GenDefaultMTLConfig returns the set of defaults for mapping the MTL
func GenDefaultMTLConfig() NetworkLayoutConfiguration {

	return NetworkLayoutConfiguration{
		Template:                        networking.DefaultMTL,
		SubdivideByCabinet:              false,
		SuperNetHack:                    true,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: true,
		IncludeUAISubnet:                false,
		NetworkingHardwareNetmask:       networking.DefaultNetworkingHardwareMask,
	}
}

// BuildCSMNetworks creates an array of IPNetworks based on the supplied system configuration
func BuildCSMNetworks(
	internalNetConfigs map[string]NetworkLayoutConfiguration,
	internalCabinetDetails []sls.CabinetGroupDetail,
	switches []*networking.ManagementSwitch,
) (networkMap networking.NetworkMap, err error) {
	v := viper.GetViper()
	networkMap = make(networking.NetworkMap)

	for name, layout := range internalNetConfigs {
		myLayout := layout

		if strings.ToUpper(name) == "CHN" {
			if v.GetString("chn-cidr4") == "" {
				log.Println("No CHN Network definition provided")
				continue
			}
		}

		// Update with computed fields
		myLayout.CabinetDetails = internalCabinetDetails
		myLayout.ManagementSwitches = switches
		netPtr, err := createNetFromLayoutConfig(myLayout)
		if err != nil {
			return nil, err
		}
		networkMap[name] = netPtr
	}

	//
	// Start the NMN Load Balancer with our Defaults
	//
	tempNMNLoadBalancer := networking.DefaultLoadBalancerNMN
	// Add a /24 for the Load Balancers
	pool, err := tempNMNLoadBalancer.CreateSubnetByMask(
		net.CIDRMask(
			networking.DefaultIPv4Block,
			networking.IPv4Size,
		),
		nil,
		"nmn_metallb_address_pool",
		int16(v.GetInt("nmn-bootstrap-vlan")),
	)
	if err != nil {
		return nil, err
	}
	pool.FullName = "NMN MetalLB"
	pool.MetalLBPoolName = "node-management"
	for nme, rsrv := range networking.PinnedMetalLBReservations {
		_, err = networking.AddReservationWithPin(
			pool,
			nme,
			strings.Join(
				rsrv.Aliases,
				",",
			),
			rsrv.IPByte,
		)
		if err != nil {
			return nil, err
		}
	}
	networkMap["NMNLB"] = &tempNMNLoadBalancer

	//
	// Start the HMN Load Balancer with our Defaults
	//
	tempHMNLoadBalancer := networking.DefaultLoadBalancerHMN
	pool, _ = tempHMNLoadBalancer.CreateSubnetByMask(
		net.CIDRMask(
			24,
			networking.IPv4Size,
		),
		nil,
		"hmn_metallb_address_pool",
		int16(v.GetInt("hmn-bootstrap-vlan")),
	)
	pool.FullName = "HMN MetalLB"
	pool.MetalLBPoolName = "hardware-management"
	for nme, rsrv := range networking.PinnedMetalLBReservations {
		// // Because of the hack to pin ip addresses, we've got an overloaded datastructure in defaults.
		// // We need to prune it here before we write it out. It's pretty ugly, but we plan to throw all of this code away when ip pinning is no longer necessary
		if nme != "istio-ingressgateway-local" {
			if nme == "istio-ingressgateway" {
				_, err = networking.AddReservationWithPin(
					pool,
					nme,
					"",
					rsrv.IPByte,
				)
				if err != nil {
					return nil, err
				}
			} else {
				_, err = networking.AddReservationWithPin(
					pool,
					nme,
					strings.Join(
						rsrv.Aliases,
						",",
					),
					rsrv.IPByte,
				)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	networkMap["HMNLB"] = &tempHMNLoadBalancer

	return networkMap, err
}

func createNetFromLayoutConfig(conf NetworkLayoutConfiguration) (network *networking.IPNetwork, err error) {

	var canCIDR netip.Prefix
	var cmnCIDR netip.Prefix
	var chnCIDR netip.Prefix

	v := viper.GetViper()

	tempNet := conf.Template
	netNameLower := strings.ToLower(tempNet.Name)

	// figure out what switches we have
	leafbmcSwitches := switchXnamesByType(
		conf.ManagementSwitches,
		"LeafBMC",
	)
	spineSwitches := switchXnamesByType(
		conf.ManagementSwitches,
		"Spine",
	)
	leafSwitches := switchXnamesByType(
		conf.ManagementSwitches,
		"Leaf",
	)
	cduSwitches := switchXnamesByType(
		conf.ManagementSwitches,
		"CDU",
	)
	edgeSwitches := switchXnamesByType(
		conf.ManagementSwitches,
		"Edge",
	)

	// Do all the special assembly for the CMN
	if tempNet.Name == "CMN" {
		if v.IsSet("cmn-cidr4") {
			cmnCIDR, err = netip.ParsePrefix(v.GetString("cmn-cidr4"))
			if err != nil {
				return nil, fmt.Errorf(
					"couldn't parse CMN CIDR [%s]: %v",
					v.GetString("cmn-cidr4"),
					err,
				)
			}
		}
		conf.DesiredBootstrapDHCPMask = net.CIDRMask(
			cmnCIDR.Bits(),
			networking.IPv4Size,
		)
		cmnStaticPool, err := netip.ParsePrefix(v.GetString("cmn-static-pool"))
		if err != nil {
			log.Printf("IP Addressing Failure\nInvalid cmn-static-pool. Cowardly refusing to create it.")
		} else {
			static, err := tempNet.CreateSubnetByCIDR(
				cmnStaticPool,
				"cmn_metallb_static_pool",
				int16(v.GetInt("cmn-bootstrap-vlan")),
				true,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"couldn't add MetalLB Static pool of %v to net %v because %v",
					v.GetString("cmn-static-pool"),
					tempNet.CIDR4,
					err,
				)
			}
			static.FullName = "CMN Static Pool MetalLB"
			static.MetalLBPoolName = "customer-management-static"
			externalDNS, _ := netip.ParseAddr(v.GetString("cmn-external-dns"))
			_, err = networking.AddReservationWithIP(
				static,
				"external-dns",
				externalDNS,
				"site to system lookups",
			)

			if err != nil {
				log.Fatal(err)
			}
		}
		cmnDynamicPool, err := netip.ParsePrefix(v.GetString("cmn-dynamic-pool"))
		if err != nil {
			log.Printf("IP Addressing Failure\nInvalid cmn-dynamic-pool. Cowardly refusing to create it.")
		} else {
			pool, err := tempNet.CreateSubnetByCIDR(
				cmnDynamicPool,
				"cmn_metallb_address_pool",
				int16(v.GetInt("cmn-bootstrap-vlan")),
				true,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"couldn't add MetalLB Dynamic pool of %v to net %v because %v",
					v.GetString("cmn-dynamic-pool"),
					tempNet.CIDR4,
					err,
				)
			}
			pool.FullName = "CMN Dynamic MetalLB"
			pool.MetalLBPoolName = "customer-management"

		}
	}

	// Do all the special assembly for the CAN
	if tempNet.Name == "CAN" {
		if v.GetString("can-cidr") != "" {
			canCIDR, err = netip.ParsePrefix(v.GetString("can-cidr"))
			if err != nil {
				return nil, fmt.Errorf(
					"invalid can-cidr %v because %v",
					canCIDR,
					err,
				)
			}
			conf.DesiredBootstrapDHCPMask = net.CIDRMask(
				canCIDR.Bits(),
				networking.IPv4Size,
			)

			if v.GetString("can-static-pool") != "" {
				canStaticPool, err := netip.ParsePrefix(v.GetString("can-static-pool"))
				if err != nil {
					log.Printf("IP Addressing Failure\nInvalid can-static-pool. Cowardly refusing to create it.")
				} else {
					static, err := tempNet.CreateSubnetByCIDR(
						canStaticPool,
						"can_metallb_static_pool",
						int16(v.GetInt("can-bootstrap-vlan")),
						true,
					)
					if err != nil {
						return nil, fmt.Errorf(
							"couldn't add MetalLB Static pool of %v to net %v because %v",
							v.GetString("can-static-pool"),
							tempNet.CIDR4,
							err,
						)
					}
					static.FullName = "CAN Static Pool MetalLB"
					static.MetalLBPoolName = "customer-access-static"
				}
			}
			if v.GetString("can-dynamic-pool") != "" {
				canDynamicPool, err := netip.ParsePrefix(v.GetString("can-dynamic-pool"))
				if err != nil {
					log.Printf("IP Addressing Failure\nInvalid can-dynamic-pool. Cowardly refusing to create it.")
				} else {
					pool, err := tempNet.CreateSubnetByCIDR(
						canDynamicPool,
						"can_metallb_address_pool",
						int16(v.GetInt("can-bootstrap-vlan")),
						true,
					)
					if err != nil {
						return nil, fmt.Errorf(
							"couldn't add MetalLB Dynamic pool of %v to net %v because %v",
							v.GetString("can-dynamic-pool"),
							tempNet.CIDR4,
							err,
						)
					}
					pool.FullName = "CAN Dynamic MetalLB"
					pool.MetalLBPoolName = "customer-access"
				}
			}
		}
	}

	// Do all the special assembly for the CHN
	if tempNet.Name == "CHN" {
		if v.GetString("chn-cidr4") != "" {
			chnCIDR, err = netip.ParsePrefix(v.GetString("chn-cidr4"))
			if err != nil {
				return nil, fmt.Errorf(
					"invalid chn-cidr4 %s because %v",
					chnCIDR,
					err,
				)
			}
			conf.DesiredBootstrapDHCPMask = net.CIDRMask(
				chnCIDR.Bits(),
				networking.IPv4Size,
			)

			if v.GetString("chn-static-pool") != "" {
				chnStaticPool, err := netip.ParsePrefix(v.GetString("chn-static-pool"))
				if err != nil {
					log.Printf("IP Addressing Failure\nInvalid chn-static-pool. Cowardly refusing to create it.")
				} else {
					static, err := tempNet.CreateSubnetByCIDR(
						chnStaticPool,
						"chn_metallb_static_pool",
						int16(v.GetInt("chn-bootstrap-vlan")),
						true,
					)
					if err != nil {
						return nil, fmt.Errorf(
							"couldn't add MetalLB Static pool of %v to net %v because %v",
							v.GetString("chn-static-pool"),
							tempNet.CIDR4,
							err,
						)
					}
					static.FullName = "CHN Static Pool MetalLB"
					static.MetalLBPoolName = "customer-high-speed-static"
				}
			}
			if v.GetString("chn-dynamic-pool") != "" {
				chnDynamicPool, err := netip.ParsePrefix(v.GetString("chn-dynamic-pool"))
				if err != nil {
					log.Printf("IP Addressing Failure\nInvalid chn-dynamic-pool. Cowardly refusing to create it.")
				} else {
					pool, err := tempNet.CreateSubnetByCIDR(
						chnDynamicPool,
						"chn_metallb_address_pool",
						int16(v.GetInt("chn-bootstrap-vlan")),
						true,
					)
					if err != nil {
						return nil, fmt.Errorf(
							"couldn't add MetalLB Dynamic pool of %v to net %v because %v",
							v.GetString("chn-dynamic-pool"),
							tempNet.CIDR4,
							err,
						)
					}
					pool.FullName = "CHN Dynamic MetalLB"
					pool.MetalLBPoolName = "customer-high-speed"
				}
			}
		}
	}

	// Initialize the required subnet for the HSN
	// This will be the entire network but is required to store IPReservations for DNS naming
	if tempNet.Name == "HSN" {
		hsnDefaultSubnet, err := netip.ParsePrefix(v.GetString("hsn-cidr"))

		if err != nil {
			log.Printf("IP Addressing Failure\nInvalid hsn-cidr. Cowardly refusing to create it.\n")
		} else {
			subnet, err := tempNet.CreateSubnetByCIDR(
				hsnDefaultSubnet,
				"hsn_base_subnet",
				networking.DefaultHSN.VlanRange[0],
				true,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"couldn't add hsn_base_subnet of %v to net %v because %v",
					v.GetString("hsn-cidr"),
					tempNet.CIDR4,
					err,
				)
			}
			subnet.FullName = "HSN Base Subnet"
		}
	}

	// Process the dedicated Networking Hardware Subnet
	if conf.IncludeNetworkingHardwareSubnet {
		// create the subnet
		smallestSubnet4, smallestSubnet6, _ := tempNet.SubnetWithin(
			uint64(len(conf.ManagementSwitches)),
		)
		if !smallestSubnet4.IsValid() {
			return nil, fmt.Errorf(
				"failed to find an IPv4 subnet for the management switches in %s",
				tempNet.CIDR4,
			)
		}
		hardwareSubnet, err := tempNet.CreateSubnetByMask(
			net.CIDRMask(
				smallestSubnet4.Bits(),
				networking.IPv4Size,
			),
			net.CIDRMask(
				smallestSubnet6.Bits(),
				networking.IPv6Size,
			),
			"network_hardware",
			conf.BaseVlan,
		)
		if err != nil {
			return &tempNet, fmt.Errorf(
				"unable to add network hardware subnet to %v because %v",
				conf.Template.Name,
				err,
			)
		}
		// populate it with base information
		hardwareSubnet.FullName = fmt.Sprintf(
			"%v Management Network Infrastructure",
			tempNet.Name,
		)
		err = networking.ReserveNetMgmtIPs(
			hardwareSubnet,
			spineSwitches,
			leafSwitches,
			leafbmcSwitches,
			cduSwitches,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to reserve management IPs for %s because %v",
				hardwareSubnet.FullName,
				err,
			)
		}
	}

	// Set up the Boostrap DHCP subnet(s).
	if conf.IncludeBootstrapDHCP {

		cidr6Key := fmt.Sprintf(
			"%s-cidr6",
			netNameLower,
		)
		cidr4Key := fmt.Sprintf(
			"%s-cidr4",
			netNameLower,
		)
		cidrKey := fmt.Sprintf(
			"%s-cidr",
			netNameLower,
		)

		// Handle IPv4 CIDRs, networks with IPv6 will use a different key for their cidr4.
		var cidr4 string
		if v.IsSet(cidr4Key) {
			cidr4 = v.GetString(cidr4Key)
		} else if v.IsSet(cidrKey) {
			cidr4 = v.GetString(cidrKey)
		}
		var cidr6 string
		if v.IsSet(cidr6Key) {
			cidr6 = v.GetString(cidr6Key)
		}

		var mask4, mask6 net.IPMask
		if cidr4 == "" {
			return &tempNet, fmt.Errorf("failed to find a CIDR v4 IP for bootstrap_dhcp")
		}
		if conf.DesiredBootstrapDHCPMask == nil {
			mask4 = net.CIDRMask(
				networking.DefaultIPv4Block,
				networking.IPv4Size,
			)
		} else {
			mask4 = conf.DesiredBootstrapDHCPMask
		}
		if cidr6 != "" {
			mask6 = net.CIDRMask(
				networking.DefaultIPv6Block,
				networking.IPv6Size,
			)
		}
		biggestMask4, err := tempNet.FindBiggestIPv4Subnet(
			mask4,
		)
		if err != nil {
			return &tempNet, fmt.Errorf(
				"failed to find a big enough subnet for bootstrap_dhcp in %s because %v",
				tempNet.CIDR4,
				err,
			)
		}
		subnet, err := tempNet.CreateSubnetByCIDR(
			biggestMask4,
			"bootstrap_dhcp",
			conf.BaseVlan,
			false,
		)
		if err != nil {
			return &tempNet, fmt.Errorf(
				"unable to add bootstrap_dhcp subnet to %v because %v",
				conf.Template.Name,
				err,
			)
		}
		biggestMask6, err := tempNet.FindBiggestIPv6Subnet(
			mask6,
		)
		if err != nil {
			if tempNet.CIDR6 != "" {
				return &tempNet, fmt.Errorf(
					"network %s had a CIDR6 set but failed to create a IPv6 subnet for bootstrap_dhcp because %v",
					tempNet.Name,
					err,
				)
			}
		} else {
			tempNet.SetSubnetIP(
				subnet,
				biggestMask6,
			)
		}
		tempNet.AppendSubnet(subnet)
		subnet.FullName = fmt.Sprintf(
			"%v Bootstrap DHCP Subnet",
			tempNet.Name,
		)
		if tempNet.Name == "NMN" || tempNet.Name == "HMN" || tempNet.Name == "CMN" || tempNet.Name == "CAN" || tempNet.Name == "CHN" {
			switch tempNet.Name {
			case "CAN":
				subnet.CIDR = canCIDR.String()
				canGateway := net.ParseIP(v.GetString("can-gateway"))
				if canGateway == nil {
					return nil, fmt.Errorf("chn-gateway4 was not a valid IPv4 address")
				}
				subnet.Gateway = canGateway
				_, err := networking.AddReservation(
					subnet,
					"can-switch-1",
					"",
				)
				if err != nil {
					return nil, err
				}
				_, err = networking.AddReservation(
					subnet,
					"can-switch-2",
					"",
				)
				if err != nil {
					return nil, err
				}
			case "CHN":
				subnet.CIDR = chnCIDR.String()
				chnGateway := net.ParseIP(v.GetString("chn-gateway4"))
				if chnGateway == nil {
					return nil, fmt.Errorf("chn-gateway4 was not a valid IPv4 address")
				}
				subnet.Gateway = chnGateway

				err := networking.ReserveEdgeSwitchIPs(
					subnet,
					edgeSwitches,
				)
				if err != nil {
					return nil, err
				}
			default:
				err := networking.ReserveNetMgmtIPs(
					subnet,
					[]string{},
					[]string{},
					[]string{},
					[]string{},
				)
				if err != nil {
					return nil, fmt.Errorf(
						"failed to reserve management IPs for %s because %v",
						subnet.FullName,
						err,
					)
				}
			}
			if tempNet.Name == "NMN" {
				_, err = networking.AddReservation(
					subnet,
					"rgw-vip",
					"rgw-virtual-ip",
				)
				if err != nil {
					return nil, err
				}
				_, err = networking.AddReservation(
					subnet,
					"kubeapi-vip",
					"k8s-virtual-ip",
				)
				if err != nil {
					return nil, err
				}
				// FabricManager VIP is only supported in CSM 1.7 and later
				_, oneSevenCSM := csm.CompareMajorMinor("1.7")
				if oneSevenCSM != -1 {
					_, err = networking.AddReservation(
						subnet,
						"fmn-vip",
						"fmn-virtual-ip",
					)
					if err != nil {
						return nil, err
					}
				}
			}
			if tempNet.Name == "HMN" {
				// FabricManager VIP is only supported in CSM 1.7 and later
				_, oneSevenCSM := csm.CompareMajorMinor("1.7")
				if oneSevenCSM != -1 {
					_, err = networking.AddReservation(
						subnet,
						"fmn-vip",
						"fmn-virtual-ip",
					)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	// Set up the ASNs
	myASN := fmt.Sprintf(
		"bgp-%s-asn",
		netNameLower,
	)
	if v.GetString(myASN) != "" {
		tempNet.PeerASN = v.GetInt("bgp-asn")
		tempNet.MyASN = v.GetInt(myASN)
	}

	// Add the macvlan/uai subnet(s)
	if conf.IncludeUAISubnet {
		// Use the NMN vlan for uai_macvlan
		uaisubnet, err := tempNet.CreateSubnetByMask(
			net.CIDRMask(
				23,
				networking.IPv4Size,
			),
			nil,
			"uai_macvlan",
			int16(v.GetInt("nmn-bootstrap-vlan")),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to add uai_macvlan subnet to %v because %v",
				tempNet.CIDR4,
				err,
			)
		}
		supernet, err := netip.ParsePrefix(tempNet.CIDR4)
		if err != nil {
			return nil, fmt.Errorf(
				"could not create supernet from subnet %s because %v",
				tempNet.CIDR4,
				err,
			)
		}
		uaisubnet.Gateway = supernet.Addr().Next().AsSlice()
		uaisubnet.FullName = "NMN UAIs"

		// Add the UAI reservations in order so they are consistent
		var keys []string
		for k := range networking.DefaultUAISubnetReservations {
			keys = append(
				keys,
				k,
			)
		}
		sort.Strings(keys)

		for _, reservationName := range keys {
			var reservationComment = networking.DefaultUAISubnetReservations[reservationName]
			reservation, err := networking.AddReservation(
				uaisubnet,
				reservationName,
				strings.Join(
					reservationComment,
					",",
				),
			)
			if err != nil {
				return nil, fmt.Errorf(
					"could not add %v to the %v network because %v",
					reservationName,
					tempNet.Name,
					err,
				)
			}
			for _, alias := range reservationComment {
				reservation.AddReservationAlias(alias)
			}
		}
	}
	// Build out the per-cabinet subnets
	// If the networks are intended to be grouped, only do the listed cabinet type

	if conf.GroupNetworksByCabinetType && conf.SubdivideByCabinet {
		if strings.HasSuffix(
			conf.Template.Name,
			"RVR",
		) {
			err = tempNet.GenSubnets(
				conf.CabinetDetails,
				conf.CabinetCIDR,
				sls.OrCabinetFilter(
					// Standard River Cabinet
					sls.CabinetClassFilter(slsCommon.ClassRiver),

					// Or the special case where special case for EX2500 cabinets with both liquid and air cooled chassis
					sls.AndCabinetFilter(
						sls.CabinetKindFilter(sls.CabinetKindEX2500),
						sls.CabinetAirCooledChassisCountFilter(1),
					),
				),
			)
			if err != nil {
				return nil, err
			}
		}
		if strings.HasSuffix(
			conf.Template.Name,
			"MTN",
		) {
			err = tempNet.GenSubnets(
				conf.CabinetDetails,
				conf.CabinetCIDR,
				sls.CabinetClassFilter(slsCommon.ClassMountain),
			)
			if err != nil {
				return nil, err
			}
			err = tempNet.GenSubnets(
				conf.CabinetDetails,
				conf.CabinetCIDR,
				sls.CabinetClassFilter(slsCommon.ClassHill),
			)
			if err != nil {
				return nil, err
			}
		}
		// Otherwise do both
	}
	if conf.SubdivideByCabinet && !conf.GroupNetworksByCabinetType {
		err := tempNet.GenSubnets(
			conf.CabinetDetails,
			conf.CabinetCIDR,
			sls.CabinetClassFilter(slsCommon.ClassRiver),
		)
		if err != nil {
			return nil, err
		}
		err = tempNet.GenSubnets(
			conf.CabinetDetails,
			conf.CabinetCIDR,
			sls.CabinetClassFilter(slsCommon.ClassHill),
		)
		if err != nil {
			return nil, err
		}
		err = tempNet.GenSubnets(
			conf.CabinetDetails,
			conf.CabinetCIDR,
			sls.CabinetClassFilter(slsCommon.ClassMountain),
		)
		if err != nil {
			return nil, err
		}
	}

	// Apply the Supernet Hack
	if conf.SuperNetHack {
		tempNet.ApplySupernetHack()
	}

	network = &tempNet
	return network, err
}

func switchXnamesByType(switches []*networking.ManagementSwitch, switchType networking.ManagementSwitchType) []string {
	var xnames []string
	for _, mswitch := range switches {
		if mswitch.SwitchType == switchType {
			xnames = append(
				xnames,
				mswitch.Xname,
			)
		}
	}
	return xnames
}
