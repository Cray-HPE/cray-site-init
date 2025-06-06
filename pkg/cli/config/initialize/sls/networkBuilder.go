/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

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
	"sort"
	"strings"

	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/spf13/viper"

	"github.com/Cray-HPE/cray-site-init/pkg/networking"
	"github.com/Cray-HPE/cray-site-init/pkg/sls"
)

// NetworkLayoutConfiguration is the internal configuration structure for shasta networks
type NetworkLayoutConfiguration struct {
	Template                        networking.IPV4Network
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
		DesiredBootstrapDHCPMask: net.CIDRMask(
			24,
			32,
		),
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
		DesiredBootstrapDHCPMask: net.CIDRMask(
			24,
			32,
		),
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
	_, cmnNet, _ := net.ParseCIDR(networking.DefaultCMN.CIDR)

	// Dynamically calculate the bootstrap_dhcp netmask based on number of NCNs.
	bootstrapSubnet, err := networking.SubnetWithin(
		*cmnNet,
		ncns,
	)
	if err != nil {
		log.Fatalf(
			"Failed to find a suitable subnet mask for %d NCNs within %v\n",
			ncns,
			networking.DefaultCMN.Name,
		)
	}

	// Dynamically calculate the network_hardware netmask based on number of NCNs.
	networkSubnet, err := networking.SubnetWithin(
		*cmnNet,
		switches,
	)
	if err != nil {
		log.Fatalf(
			"Failed to find a suitable subnet mask for %d switches within %v\n",
			switches,
			networking.DefaultCMN.Name,
		)
	}

	return NetworkLayoutConfiguration{
		Template:                        networking.DefaultCMN,
		SubdivideByCabinet:              false,
		SuperNetHack:                    true,
		IncludeBootstrapDHCP:            true,
		IncludeNetworkingHardwareSubnet: true,
		IncludeUAISubnet:                false,
		NetworkingHardwareNetmask:       networkSubnet.Mask,
		DesiredBootstrapDHCPMask:        bootstrapSubnet.Mask,
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
		DesiredBootstrapDHCPMask: net.CIDRMask(
			24,
			32,
		),
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
		DesiredBootstrapDHCPMask: net.CIDRMask(
			24,
			32,
		),
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
		DesiredBootstrapDHCPMask: net.CIDRMask(
			24,
			32,
		),
	}
}

// BuildCSMNetworks creates an array of IPv4 Networks based on the supplied system configuration
func BuildCSMNetworks(
	internalNetConfigs map[string]NetworkLayoutConfiguration,
	internalCabinetDetails []sls.CabinetGroupDetail,
	switches []*networking.ManagementSwitch,
) (map[string]*networking.IPV4Network, error) {
	v := viper.GetViper()
	var networkMap = make(map[string]*networking.IPV4Network)

	for name, layout := range internalNetConfigs {
		// log.Println("Building Network for ", name)
		myLayout := layout

		if name == "CHN" {
			if v.GetString("chn-cidr") == "" {
				log.Println("No CHN Network definition provided")
				continue
			}
		}

		// Update with computed fields
		myLayout.CabinetDetails = internalCabinetDetails
		myLayout.ManagementSwitches = switches

		netPtr, err := createNetFromLayoutConfig(myLayout)
		if err != nil {
			log.Fatalf(
				"Couldn't add %v Network because %v",
				name,
				err,
			)
		}
		networkMap[name] = netPtr
	}

	//
	// Start the NMN Load Balancer with our Defaults
	//
	tempNMNLoadBalancer := networking.DefaultLoadBalancerNMN
	// Add a /24 for the Load Balancers
	pool, _ := tempNMNLoadBalancer.AddSubnet(
		net.CIDRMask(
			24,
			32,
		),
		"nmn_metallb_address_pool",
		int16(v.GetInt("nmn-bootstrap-vlan")),
	)
	pool.FullName = "NMN MetalLB"
	pool.MetalLBPoolName = "node-management"
	for nme, rsrv := range networking.PinnedMetalLBReservations {
		pool.AddReservationWithPin(
			nme,
			strings.Join(
				rsrv.Aliases,
				",",
			),
			rsrv.IPByte,
		)
	}
	networkMap["NMNLB"] = &tempNMNLoadBalancer

	//
	// Start the HMN Load Balancer with our Defaults
	//
	tempHMNLoadBalancer := networking.DefaultLoadBalancerHMN
	pool, _ = tempHMNLoadBalancer.AddSubnet(
		net.CIDRMask(
			24,
			32,
		),
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
				pool.AddReservationWithPin(
					nme,
					"",
					rsrv.IPByte,
				)
			} else {
				pool.AddReservationWithPin(
					nme,
					strings.Join(
						rsrv.Aliases,
						",",
					),
					rsrv.IPByte,
				)
			}
		}
	}
	networkMap["HMNLB"] = &tempHMNLoadBalancer

	return networkMap, nil
}

func createNetFromLayoutConfig(conf NetworkLayoutConfiguration) (*networking.IPV4Network, error) {
	// log.Printf("Creating a network for %v with NetworkLayoutConfig %+v", conf.Template.Name, conf)
	var canCIDR *net.IPNet
	var cmnCIDR *net.IPNet
	var chnCIDR *net.IPNet
	// I hope this viper is temporary
	v := viper.GetViper()
	// start with the defaults
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
		_, cmnCIDR, _ = net.ParseCIDR(v.GetString("cmn-cidr"))
		conf.DesiredBootstrapDHCPMask = cmnCIDR.Mask
		_, cmnStaticPool, err := net.ParseCIDR(v.GetString("cmn-static-pool"))
		if err != nil {
			log.Printf("IP Addressing Failure\nInvalid cmn-static-pool. Cowardly refusing to create it.")
		} else {
			static, err := tempNet.AddSubnetbyCIDR(
				*cmnStaticPool,
				"cmn_metallb_static_pool",
				int16(v.GetInt("cmn-bootstrap-vlan")),
			)
			if err != nil {
				log.Fatalf(
					"IP Addressing Failure\n"+
						"Couldn't add MetalLB Static pool of %v to net %v: %v\n"+
						"Possible missing or mismatched cmn-static-pool input value.",
					v.GetString("cmn-static-pool"),
					tempNet.CIDR,
					err,
				)
			}
			static.FullName = "CMN Static Pool MetalLB"
			static.MetalLBPoolName = "customer-management-static"

			_, err = static.AddReservationWithIP(
				"external-dns",
				v.GetString("cmn-external-dns"),
				"site to system lookups",
			)
			if err != nil {
				log.Fatal(err)
			}
		}
		_, cmnDynamicPool, err := net.ParseCIDR(v.GetString("cmn-dynamic-pool"))
		if err != nil {
			log.Printf("IP Addressing Failure\nInvalid cmn-dynamic-pool. Cowardly refusing to create it.")
		} else {
			pool, err := tempNet.AddSubnetbyCIDR(
				*cmnDynamicPool,
				"cmn_metallb_address_pool",
				int16(v.GetInt("cmn-bootstrap-vlan")),
			)
			if err != nil {
				log.Fatalf(
					"IP Addressing Failure\n"+
						"Couldn't add MetalLB Dynamic pool of %v to net %v: %v\n"+
						"Possible missing or mismatched cmn-dynamic-pool input value.",
					v.GetString("cmn-dynamic-pool"),
					tempNet.CIDR,
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
			_, canCIDR, _ = net.ParseCIDR(v.GetString("can-cidr"))
			conf.DesiredBootstrapDHCPMask = canCIDR.Mask

			if v.GetString("can-static-pool") != "" {
				_, canStaticPool, err := net.ParseCIDR(v.GetString("can-static-pool"))
				if err != nil {
					log.Printf("IP Addressing Failure\nInvalid can-static-pool. Cowardly refusing to create it.")
				} else {
					static, err := tempNet.AddSubnetbyCIDR(
						*canStaticPool,
						"can_metallb_static_pool",
						int16(v.GetInt("can-bootstrap-vlan")),
					)
					if err != nil {
						log.Fatalf(
							"IP Addressing Failure\n"+
								"Couldn't add MetalLB Static pool of %v to net %v: %v\n"+
								"Possible missing or mismatched can-static-pool input value.",
							v.GetString("can-static-pool"),
							tempNet.CIDR,
							err,
						)
					}
					static.FullName = "CAN Static Pool MetalLB"
					static.MetalLBPoolName = "customer-access-static"
				}
			}
			if v.GetString("can-dynamic-pool") != "" {
				_, canDynamicPool, err := net.ParseCIDR(v.GetString("can-dynamic-pool"))
				if err != nil {
					log.Printf("IP Addressing Failure\nInvalid can-dynamic-pool. Cowardly refusing to create it.")
				} else {
					pool, err := tempNet.AddSubnetbyCIDR(
						*canDynamicPool,
						"can_metallb_address_pool",
						int16(v.GetInt("can-bootstrap-vlan")),
					)
					if err != nil {
						log.Fatalf(
							"IP Addressing Failure\n"+
								"Couldn't add MetalLB Dynamic pool of %v to net %v: %v\n"+
								"Possible missing or mismatched can-dynamic-pool value.",
							v.GetString("can-dynamic-pool"),
							tempNet.CIDR,
							err,
						)
						log.Fatalf("Possible missing or mismatched can-dynamic-pool value.")
					}
					pool.FullName = "CAN Dynamic MetalLB"
					pool.MetalLBPoolName = "customer-access"
				}
			}
		}
	}

	// Do all the special assembly for the CHN
	if tempNet.Name == "CHN" {
		if v.GetString("chn-cidr") != "" {
			_, chnCIDR, _ = net.ParseCIDR(v.GetString("chn-cidr"))
			conf.DesiredBootstrapDHCPMask = chnCIDR.Mask

			if v.GetString("chn-static-pool") != "" {
				_, chnStaticPool, err := net.ParseCIDR(v.GetString("chn-static-pool"))
				if err != nil {
					log.Printf("IP Addressing Failure\nInvalid chn-static-pool. Cowardly refusing to create it.")
				} else {
					static, err := tempNet.AddSubnetbyCIDR(
						*chnStaticPool,
						"chn_metallb_static_pool",
						int16(v.GetInt("chn-bootstrap-vlan")),
					)
					if err != nil {
						log.Fatalf(
							"IP Addressing Failure\n"+
								"Couldn't add MetalLB Static pool of %v to net %v: %v\n"+
								"Possible missing or mismatched chn-static-pool input value.",
							v.GetString("chn-static-pool"),
							tempNet.CIDR,
							err,
						)
					}
					static.FullName = "CHN Static Pool MetalLB"
					static.MetalLBPoolName = "customer-high-speed-static"
				}
			}
			if v.GetString("chn-dynamic-pool") != "" {
				_, chnDynamicPool, err := net.ParseCIDR(v.GetString("chn-dynamic-pool"))
				if err != nil {
					log.Printf("IP Addressing Failure\nInvalid chn-dynamic-pool. Cowardly refusing to create it.")
				} else {
					pool, err := tempNet.AddSubnetbyCIDR(
						*chnDynamicPool,
						"chn_metallb_address_pool",
						int16(v.GetInt("chn-bootstrap-vlan")),
					)
					if err != nil {
						log.Fatalf(
							"IP Addressing Failure\n"+
								"Couldn't add MetalLB Dynamic pool of %v to net %v: %v\n"+
								"Possible missing or mismatched chn-dynamic-pool value.",
							v.GetString("chn-dynamic-pool"),
							tempNet.CIDR,
							err,
						)
						log.Fatalf("Possible missing or mismatched chn-dynamic-pool value.")
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
		_, hsnDefaultSubnet, err := net.ParseCIDR(v.GetString("hsn-cidr"))
		if err != nil {
			log.Printf("IP Addressing Failure\nInvalid hsn-cidr. Cowardly refusing to create it.")
		} else {
			subnet, err := tempNet.AddSubnetbyCIDR(
				*hsnDefaultSubnet,
				"hsn_base_subnet",
				int16(networking.DefaultHSN.VlanRange[0]),
			)
			if err != nil {
				log.Fatalf(
					"IP Addressing Failure\nCouldn't add hsn_base_subnet of %v to net %v: %v",
					v.GetString("hsn-cidr"),
					tempNet.CIDR,
					err,
				)
			}
			subnet.FullName = "HSN Base Subnet"
		}
	}

	// Process the dedicated Networking Hardware Subnet
	if conf.IncludeNetworkingHardwareSubnet {
		// create the subnet
		hardwareSubnet, err := tempNet.AddSubnet(
			conf.NetworkingHardwareNetmask,
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
		hardwareSubnet.ReserveNetMgmtIPs(
			spineSwitches,
			leafSwitches,
			leafbmcSwitches,
			cduSwitches,
		)
	}

	// Set up the Boostrap DHCP subnet(s)
	if conf.IncludeBootstrapDHCP {
		myNet := fmt.Sprintf(
			"%s-cidr",
			netNameLower,
		)
		if v.GetString(myNet) != "" {
			var subnet *networking.IPV4Subnet
			subnet, err := tempNet.AddBiggestSubnet(
				conf.DesiredBootstrapDHCPMask,
				"bootstrap_dhcp",
				conf.BaseVlan,
			)
			if err != nil {
				return &tempNet, fmt.Errorf(
					"unable to add bootstrap_dhcp subnet to %v because %v",
					conf.Template.Name,
					err,
				)
			}
			subnet.FullName = fmt.Sprintf(
				"%v Bootstrap DHCP Subnet",
				tempNet.Name,
			)
			subnet.ParentDevice = tempNet.ParentDevice
			if tempNet.Name == "NMN" || tempNet.Name == "HMN" || tempNet.Name == "CMN" || tempNet.Name == "CAN" || tempNet.Name == "CHN" {
				if tempNet.Name == "CAN" {
					subnet.CIDR = *canCIDR
					subnet.Gateway = net.ParseIP(v.GetString("can-gateway"))
					subnet.AddReservation(
						"can-switch-1",
						"",
					)
					subnet.AddReservation(
						"can-switch-2",
						"",
					)
				} else if tempNet.Name == "CHN" {
					subnet.CIDR = *chnCIDR
					subnet.Gateway = net.ParseIP(v.GetString("chn-gateway"))
					subnet.ReserveEdgeSwitchIPs(edgeSwitches)
				} else {
					subnet.ReserveNetMgmtIPs(
						[]string{},
						[]string{},
						[]string{},
						[]string{},
					)
				}
				if tempNet.Name == "NMN" {
					subnet.AddReservation(
						"rgw-vip",
						"rgw-virtual-ip",
					)
					subnet.AddReservation(
						"kubeapi-vip",
						"k8s-virtual-ip",
					)
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
		uaisubnet, err := tempNet.AddSubnet(
			net.CIDRMask(
				23,
				32,
			),
			"uai_macvlan",
			int16(v.GetInt("nmn-bootstrap-vlan")),
		)
		_, supernetNet, _ := net.ParseCIDR(tempNet.CIDR)
		uaisubnet.Gateway = networking.Add(
			supernetNet.IP,
			1,
		)
		if err != nil {
			log.Fatalf(
				"Could not add the uai subnet to the %v Network: %v",
				tempNet.Name,
				err,
			)
		}
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
			reservation := uaisubnet.AddReservation(
				reservationName,
				strings.Join(
					reservationComment,
					",",
				),
			)
			for _, alias := range reservationComment {
				reservation.AddReservationAlias(alias)
			}
		}
		// log.Println("Added the MacVlan Subnet at ", uaisubnet.CIDR.String())
	}
	// Build out the per-cabinet subnets
	// If the networks are intended to be grouped, only do the listed cabinet type

	if conf.GroupNetworksByCabinetType && conf.SubdivideByCabinet {
		if strings.HasSuffix(
			conf.Template.Name,
			"RVR",
		) {
			tempNet.GenSubnets(
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
		}
		if strings.HasSuffix(
			conf.Template.Name,
			"MTN",
		) {
			tempNet.GenSubnets(
				conf.CabinetDetails,
				conf.CabinetCIDR,
				sls.CabinetClassFilter(slsCommon.ClassMountain),
			)
			tempNet.GenSubnets(
				conf.CabinetDetails,
				conf.CabinetCIDR,
				sls.CabinetClassFilter(slsCommon.ClassHill),
			)
		}
		// Otherwise do both
	}
	if conf.SubdivideByCabinet && !conf.GroupNetworksByCabinetType {
		tempNet.GenSubnets(
			conf.CabinetDetails,
			conf.CabinetCIDR,
			sls.CabinetClassFilter(slsCommon.ClassRiver),
		)
		tempNet.GenSubnets(
			conf.CabinetDetails,
			conf.CabinetCIDR,
			sls.CabinetClassFilter(slsCommon.ClassHill),
		)
		tempNet.GenSubnets(
			conf.CabinetDetails,
			conf.CabinetCIDR,
			sls.CabinetClassFilter(slsCommon.ClassMountain),
		)
	}

	// Apply the Supernet Hack
	if conf.SuperNetHack {
		tempNet.ApplySupernetHack()
	}
	return &tempNet, nil
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
