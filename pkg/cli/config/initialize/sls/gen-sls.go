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
	"strings"

	"github.com/spf13/cobra"

	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"

	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

// OutputFile name of the output file for generate SLS.
const OutputFile = "sls_input_file.json"

// NewCommand represents the sls command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "gen-sls [options] <path>",
		Short: "Generates SLS input file",
		Long: `Generates SLS input file based on a Shasta configuration and
	HMN connections file. By default, cabinets are assumed to be one River, the
	rest Mountain.`,
		Args: cobra.RangeArgs(
			0,
			1,
		),
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Deprecated
			log.Println("This command has been deprecated")
		},
	}
	c.Flags().Int16(
		"river-cabinets",
		1,
		"Number of River cabinets",
	)
	c.Flags().Int(
		"hill-cabinets",
		0,
		"Number of River cabinets",
	)
	return c
}

// GenCabinetMap creates a map of cabinets.
func GenCabinetMap(
	cd []sls.CabinetGroupDetail, shastaNetworks map[string]*networking.IPNetwork,
) map[slsCommon.CabinetType]map[string]CabinetTemplate {
	// Use information from CabinetGroupDetails and shastaNetworks to generate
	// Cabinet information for SLS
	cabinets := make(map[sls.CabinetKind][]int) // key => kind, value => list of cabinet_ids
	cabinetDetails := make(map[sls.CabinetKind]map[int]sls.CabinetDetail)
	for _, cabinetGroup := range cd {
		cabinetIDs := cabinetGroup.CabinetIDs()
		cabinets[cabinetGroup.Kind] = cabinetIDs
		cabinetDetails[cabinetGroup.Kind] = cabinetGroup.GetCabinetDetails()
	}

	// Iterate through the cabinets of each kind and build structures that work for SLS Generation
	slsCabinetMap := make(map[sls.CabinetKind]map[string]CabinetTemplate)
	for cabinetKind, cabIds := range cabinets {
		class, err := cabinetKind.Class()
		if err != nil {
			log.Fatalf(
				"Unable to determine cabinet class for cabinet kind (%v)",
				cabinetKind,
			)
		}

		tmpCabinets := make(map[string]CabinetTemplate)
		for _, id := range cabIds {
			// Find the NMN and HMN networks for each cabinet
			networks := make(map[string]slsCommon.CabinetNetworks)
			for _, netName := range []string{
				"NMN",
				"HMN",
				"NMN_MTN",
				"HMN_MTN",
				"NMN_RVR",
				"HMN_RVR",
			} {
				if shastaNetworks[netName] != nil {
					subnet := shastaNetworks[netName].SubnetByName(
						fmt.Sprintf(
							"cabinet_%d",
							id,
						),
					)
					if subnet.CIDR != "<nil>" {
						networks[strings.TrimSuffix(
							strings.TrimSuffix(
								netName,
								"_MTN",
							),
							"_RVR",
						)] = slsCommon.CabinetNetworks{
							CIDR:    subnet.CIDR,
							Gateway: subnet.Gateway.String(),
							VLan:    int(subnet.VlanID),
						}
					}
				}
			}

			// Build out the sls cabinet structure
			cabinetTemplate := CabinetTemplate{
				Xname: xnames.Cabinet{
					Cabinet: id,
				},
				Class: class,
				CabinetNetworks: map[string]map[string]slsCommon.CabinetNetworks{
					"cn": networks,
				},
			}

			// If the cabinet kind is actually a special model of cabinet, add it to the template.
			if cabinetKind.IsModel() {
				cabinetTemplate.Model = string(cabinetKind)
			}

			// Do the stuff specific to each kind (within the context of a single cabinet)
			switch class {
			case slsCommon.ClassRiver:
				cabinetTemplate.CabinetNetworks["ncn"] = networks
				cabinetTemplate.AirCooledChassisList = DefaultRiverChassisList
				cabinetTemplate.LiquidCooledChassisList = []int{}

				if cabinetDetail, present := cabinetDetails[cabinetKind][id]; present && cabinetDetail.ChassisCount != nil {
					log.Fatalf(
						"Overriding air or liquid cooled chassis counts is not permitted for river cabinets (%s). Refusing to continue",
						cabinetTemplate.Xname.String(),
					)
				}
			case slsCommon.ClassHill:
				// Start with a EX2000 Hill cabinet as the default
				cabinetTemplate.Class = slsCommon.ClassHill
				cabinetTemplate.AirCooledChassisList = []int{}
				cabinetTemplate.LiquidCooledChassisList = DefaultHillChassisList

				// Find any cabinet specific overrides
				if cabinetDetail, present := cabinetDetails[cabinetKind][id]; present && cabinetKind != sls.CabinetKindEX2500 {
					// This is a normal EX2000 hill cabinet
					if cabinetDetail.ChassisCount != nil {
						log.Fatalf(
							"Overriding air or liquid cooled chassis counts is not permitted for hill (EX2000) cabinets (%s). Refusing to continue",
							cabinetTemplate.Xname.String(),
						)
					}
				} else if cabinetDetail, present := cabinetDetails[cabinetKind][id]; present && cabinetKind == sls.CabinetKindEX2500 {
					// So we have a EX2500 Cabinet, as chassis overrides has been specified
					// Here are following allowed configurations for it to be considered a hill cabinet
					// - 0 air-cooled chassis, with 1, 2, or 3 liquid cooled chassis
					// - 1 air-cooled chassis, with 1 liquid-cooled chassis
					if cabinetDetail.ChassisCount == nil {
						msg := "EX2500 cabinets require chassis counts to be specified via cabinets.yaml\n"
						msg += "The following is an example how to specify chassis counts in cabinets.yaml:\n"
						msg += "    - id: 8001\n"
						msg += "      hmn-vlan: 3001\n"
						msg += "      nmn-vlan: 2001\n"
						msg += "      chassis-count:\n"
						msg += "          air-cooled: 0\n"
						msg += "          liquid-cooled: 3"
						log.Fatal(msg)
					}

					switch cabinetDetail.ChassisCount.AirCooled {
					case 0:
						if cabinetDetail.ChassisCount.LiquidCooled < 1 || 3 < cabinetDetail.ChassisCount.LiquidCooled {
							log.Fatalf(
								"Invalid liquid-cooled chassis count specified for hill (EX2500) cabinet %s. Given %d, expected between 1 and 3. Refusing to continue",
								cabinetTemplate.Xname.String(),
								cabinetDetail.ChassisCount.AirCooled,
							)
						}

						// We are assuming the following chassis configuration
						// 1 Chassis - c0
						// 2 Chassis - c0, c1
						// 3 Chassis - c0, c1, c2
						liquidCooledChassisList := []int{}
						for i := 0; i < cabinetDetail.ChassisCount.LiquidCooled; i++ {
							liquidCooledChassisList = append(
								liquidCooledChassisList,
								i,
							)
						}
						cabinetTemplate.LiquidCooledChassisList = liquidCooledChassisList

					case 1:
						switch cabinetDetail.ChassisCount.LiquidCooled {
						case 0:
							// The EX2500 cabinet can also be like a standard 19 inch server rack. In that case it should be treated like a normal river cabinet
							// This is for compatibility with Slingshot tooling that a river chassis in a EX2500 cabinet are always c4.
							cabinetTemplate.AirCooledChassisList = []int{4}
							cabinetTemplate.LiquidCooledChassisList = []int{}
						case 1:
							// Valid configuration
							cabinetTemplate.AirCooledChassisList = []int{4}
							cabinetTemplate.LiquidCooledChassisList = []int{0}
						default:
							// Invalid configuration
							log.Fatalf(
								"Invalid liquid-cooled chassis count specified for hill (EX2500) cabinet %s. Given %d, expected 1. EX2500 cabinets with 1 air-cooled chassis can only have 1 liquid-cooled chassis. Refusing to continue",
								cabinetTemplate.Xname.String(),
								cabinetDetail.ChassisCount.LiquidCooled,
							)
						}
					default:
						log.Fatalf(
							"Invalid air-cooled chassis count specified for hill (EX2500) cabinet %s. Given %d, expected 0 or 1. Refusing to continue",
							cabinetTemplate.Xname.String(),
							cabinetDetail.ChassisCount.AirCooled,
						)
					}
				}

			case slsCommon.ClassMountain:
				cabinetTemplate.Class = slsCommon.ClassMountain
				cabinetTemplate.AirCooledChassisList = []int{}
				cabinetTemplate.LiquidCooledChassisList = DefaultMountainChassisList

				if cabinetDetail, present := cabinetDetails[cabinetKind][id]; present && cabinetDetail.ChassisCount != nil {
					log.Fatalf(
						"Overriding air or liquid cooled chassis counts is not permitted for mountain cabinets (%s). Refusing to continue",
						cabinetTemplate.Xname.String(),
					)
				}
			default:
				log.Fatalf(
					"Unknown cabinet class (%v) for cabinet x%v with cabient kind %v",
					class,
					id,
					cabinetKind,
				)
			}
			// Validate that our cabinet will be addressable as a valid Xname
			if err := cabinetTemplate.Xname.Validate(); err != nil {
				log.Fatalf(
					"%s is not a valid Xname for a cabinet. Error %v. Refusing to continue.",
					cabinetTemplate.Xname.String(),
					err,
				)
			}
			tmpCabinets[cabinetTemplate.Xname.String()] = cabinetTemplate
		}
		slsCabinetMap[cabinetKind] = tmpCabinets
	}

	// Build up the result map, by sorting cabinets by their class
	result := map[slsCommon.CabinetType]map[string]CabinetTemplate{}
	for _, cabinets := range slsCabinetMap {
		for xname, template := range cabinets {
			class := template.Class

			if _, present := result[class]; !present {
				result[class] = map[string]CabinetTemplate{}
			}

			result[class][xname] = template
		}
	}

	return result
}

// ConvertManagementSwitchToSLS converts a management switch to a format compatible with SLS.
func ConvertManagementSwitchToSLS(s *networking.ManagementSwitch) (slsCommon.GenericHardware, error) {
	switch s.SwitchType {
	case networking.ManagementSwitchTypeLeafBMC:
		return slsCommon.GenericHardware{
			Parent:     xnametypes.GetHMSCompParent(s.Xname),
			Xname:      s.Xname,
			Type:       slsCommon.MgmtSwitch,
			TypeString: xnametypes.MgmtSwitch,
			Class:      slsCommon.ClassRiver,
			ExtraPropertiesRaw: slsCommon.ComptypeMgmtSwitch{
				IP4Addr: s.ManagementInterface.String(),
				Brand:   s.Brand.String(),
				Model:   s.Model,
				SNMPAuthPassword: fmt.Sprintf(
					"vault://hms-creds/%s",
					s.Xname,
				),
				SNMPAuthProtocol: "MD5",
				SNMPPrivPassword: fmt.Sprintf(
					"vault://hms-creds/%s",
					s.Xname,
				),
				SNMPPrivProtocol: "DES",
				SNMPUsername:     "testuser",

				Aliases: []string{s.Name},
			},
		}, nil
	case networking.ManagementSwitchTypeLeaf:
		fallthrough
	case networking.ManagementSwitchTypeSpine:
		return slsCommon.GenericHardware{
			Parent:     xnametypes.GetHMSCompParent(s.Xname),
			Xname:      s.Xname,
			Type:       slsCommon.MgmtHLSwitch,
			TypeString: xnametypes.MgmtHLSwitch,
			Class:      slsCommon.ClassRiver,
			ExtraPropertiesRaw: slsCommon.ComptypeMgmtHLSwitch{
				IP4Addr: s.ManagementInterface.String(),
				Brand:   s.Brand.String(),
				Model:   s.Model,
				Aliases: []string{s.Name},
			},
		}, nil

	case networking.ManagementSwitchTypeCDU:
		if xnametypes.GetHMSType(s.Xname) == xnametypes.MgmtHLSwitch {
			// This is a CDU switch in the River cabinet that is adjacent to the Hill cabinet. Use the MgmtHLSwitch type instead
			return slsCommon.GenericHardware{
				Parent:     xnametypes.GetHMSCompParent(s.Xname),
				Xname:      s.Xname,
				Type:       slsCommon.MgmtHLSwitch,
				TypeString: xnametypes.MgmtHLSwitch,
				Class:      slsCommon.ClassRiver,
				ExtraPropertiesRaw: slsCommon.ComptypeMgmtHLSwitch{
					IP4Addr: s.ManagementInterface.String(),
					Brand:   s.Brand.String(),
					Model:   s.Model,
					Aliases: []string{s.Name},
				},
			}, nil
		}

		return slsCommon.GenericHardware{
			Parent:     xnametypes.GetHMSCompParent(s.Xname),
			Xname:      s.Xname,
			Type:       slsCommon.CDUMgmtSwitch,
			TypeString: xnametypes.CDUMgmtSwitch,
			Class:      slsCommon.ClassMountain,
			ExtraPropertiesRaw: slsCommon.ComptypeCDUMgmtSwitch{
				Brand:   s.Brand.String(),
				Model:   s.Model,
				Aliases: []string{s.Name},
			},
		}, nil
	}

	return slsCommon.GenericHardware{}, fmt.Errorf(
		"unknown management switch type: %s",
		s.SwitchType,
	)
}

// ExtractSwitchesfromReservations extracts all the switches from an IP network.
func ExtractSwitchesfromReservations(subnet *slsCommon.IPSubnet) ([]networking.ManagementSwitch, error) {
	var switches []networking.ManagementSwitch
	for _, reservation := range subnet.IPReservations {
		if strings.HasPrefix(
			reservation.Name,
			"sw-spine",
		) {
			switches = append(
				switches,
				networking.ManagementSwitch{
					Xname:               reservation.Comment,
					Name:                reservation.Name,
					SwitchType:          networking.ManagementSwitchTypeSpine,
					ManagementInterface: reservation.IPAddress,
				},
			)
		}
		if strings.HasPrefix(
			reservation.Name,
			"sw-leaf",
		) && !strings.HasPrefix(
			reservation.Name,
			"sw-leaf-bmc",
		) {
			switches = append(
				switches,
				networking.ManagementSwitch{
					Xname:               reservation.Comment,
					Name:                reservation.Name,
					SwitchType:          networking.ManagementSwitchTypeLeaf,
					ManagementInterface: reservation.IPAddress,
				},
			)
		}
		if strings.HasPrefix(
			reservation.Name,
			"sw-leaf-bmc",
		) {
			switches = append(
				switches,
				networking.ManagementSwitch{
					Xname:               reservation.Comment,
					Name:                reservation.Name,
					SwitchType:          networking.ManagementSwitchTypeLeafBMC,
					ManagementInterface: reservation.IPAddress,
				},
			)
		}
		if strings.HasPrefix(
			reservation.Name,
			"sw-cdu",
		) {
			switches = append(
				switches,
				networking.ManagementSwitch{
					Xname:               reservation.Comment,
					Name:                reservation.Name,
					SwitchType:          networking.ManagementSwitchTypeCDU,
					ManagementInterface: reservation.IPAddress,
				},
			)
		}
	}

	return switches, nil
}

// ConvertIPNetworksToSLS converts IP network definitions to a compatible format for SLS.
func ConvertIPNetworksToSLS(networks *[]networking.IPNetwork) map[string]slsCommon.Network {
	slsNetworks := make(
		map[string]slsCommon.Network,
		len(*networks),
	)

	for _, network := range *networks {
		// TODO enforce the network name to have no spaces
		slsNetwork := convertIPNetworkToSLS(&network)
		slsNetworks[slsNetwork.Name] = slsNetwork
	}

	return slsNetworks
}

func convertIPNetworkToSLS(n *networking.IPNetwork) (slsNetwork slsCommon.Network) {
	subnets := make(
		[]slsCommon.IPSubnet,
		len(n.Subnets),
	)
	for i, subnet := range n.Subnets {
		subnets[i] = convertIPSubnetToSLS(subnet)
	}
	slsNetwork = slsCommon.Network{
		Name:     n.Name,
		FullName: n.FullName,
		Type:     n.NetType,
		IPRanges: []string{n.CIDR4},
	}
	slsExtraProperties := slsCommon.NetworkExtraProperties{
		Comment:            n.Comment,
		CIDR:               n.CIDR4,
		CIDR6:              n.CIDR6,
		MTU:                n.MTU,
		VlanRange:          n.VlanRange,
		PeerASN:            n.PeerASN,
		MyASN:              n.MyASN,
		Subnets:            subnets,
		SystemDefaultRoute: n.SystemDefaultRoute,
	}
	slsNetwork.ExtraPropertiesRaw = slsExtraProperties
	return slsNetwork
}

func convertIPSubnetToSLS(s *slsCommon.IPSubnet) (ipSubnet slsCommon.IPSubnet) {
	ipReservations := make(
		[]slsCommon.IPReservation,
		len(s.IPReservations),
	)
	for i, ipReservation := range s.IPReservations {
		ipReservations[i] = convertIPReservationToSLS(&ipReservation)
	}
	ipSubnet = slsCommon.IPSubnet{
		Name:             s.Name,
		FullName:         s.FullName,
		CIDR:             s.CIDR,
		CIDR6:            s.CIDR6,
		VlanID:           s.VlanID,
		Comment:          s.Comment,
		Gateway:          s.Gateway,
		Gateway6:         s.Gateway6,
		DHCPStart:        s.DHCPStart,
		DHCPEnd:          s.DHCPEnd,
		ReservationStart: s.ReservationStart,
		ReservationEnd:   s.ReservationEnd,
		IPReservations:   ipReservations,
		MetalLBPoolName:  s.MetalLBPoolName,
	}
	return ipSubnet
}

func convertIPReservationToSLS(s *slsCommon.IPReservation) (ipReservation slsCommon.IPReservation) {
	ipReservation = slsCommon.IPReservation{
		IPAddress:  s.IPAddress,
		IPAddress6: s.IPAddress6,
		Name:       s.Name,
		Comment:    s.Comment,
		Aliases:    s.Aliases,
	}
	return ipReservation
}
