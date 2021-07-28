/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	base "stash.us.cray.com/HMS/hms-base"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
)

// initCmd represents the init command
var genSLSCmd = &cobra.Command{
	Use:   "gen-sls [options] <path>",
	Short: "Generates SLS input file",
	Long: `Generates SLS input file based on a Shasta configuration and
	HMN connections file. By default, cabinets are assumed to be one River, the
	rest Mountain.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		// Deprecated
		log.Println("This command has been deprecated")

	},
}

func init() {
	configCmd.AddCommand(genSLSCmd)
	genSLSCmd.DisableAutoGenTag = true
	genSLSCmd.Flags().Int16("river-cabinets", 1, "Number of River cabinets")
	genSLSCmd.Flags().Int("hill-cabinets", 0, "Number of River cabinets")

}

func genCabinetMap(cd []csi.CabinetGroupDetail, shastaNetworks map[string]*csi.IPV4Network) map[string]map[string]sls_common.GenericHardware {
	// Use information from CabinetGroupDetails and shastaNetworks to generate
	// Cabinet information for SLS
	cabinets := make(map[string][]int) // key => kind, value => list of cabinet_ids
	for _, cab := range cd {
		cabinets[strings.ToLower(cab.Kind)] = cab.CabinetIDs()
	}

	// Iterate through the cabinets of each kind and build structures that work for SLS Generation
	slsCabinetMap := make(map[string]map[string]sls_common.GenericHardware)
	for kind, cabIds := range cabinets {
		tmpCabinets := make(map[string]sls_common.GenericHardware)
		for _, id := range cabIds {
			// Find the NMN and HMN networks for each cabinet
			networks := make(map[string]sls_common.CabinetNetworks)
			for _, netName := range []string{"NMN", "HMN", "NMN_MTN", "HMN_MTN", "NMN_RVR", "HMN_RVR"} {
				if shastaNetworks[netName] != nil {
					subnet := shastaNetworks[netName].SubnetbyName(fmt.Sprintf("cabinet_%d", id))
					if subnet.CIDR.String() != "<nil>" {
						networks[strings.TrimSuffix(strings.TrimSuffix(netName, "_MTN"), "_RVR")] = sls_common.CabinetNetworks{
							CIDR:    subnet.CIDR.String(),
							Gateway: subnet.Gateway.String(),
							VLan:    int(subnet.VlanID),
						}
					}
				}
			}
			// Build out the sls cabinet structure
			cabinet := sls_common.GenericHardware{
				Parent:     "s0",
				Xname:      fmt.Sprintf("x%d", id),
				Type:       sls_common.Cabinet,
				TypeString: base.Cabinet,
				ExtraPropertiesRaw: sls_common.ComptypeCabinet{
					Networks: map[string]map[string]sls_common.CabinetNetworks{"cn": networks},
				},
			}
			// Do the stuff specific to each kind (within the context of a single cabinet)
			if kind == "river" {
				cabinet.Class = sls_common.ClassRiver
				cabinet.ExtraPropertiesRaw.(sls_common.ComptypeCabinet).Networks["ncn"] = networks
			}
			if kind == "hill" {
				cabinet.Class = sls_common.ClassHill
			}
			if kind == "mountain" {
				cabinet.Class = sls_common.ClassMountain
			}
			// Validate that our cabinet will be addressable as a valid Xname
			if base.GetHMSType(cabinet.Xname) != base.Cabinet {
				log.Fatalf("%s is not a valid Xname for a cabinet.  Refusing to continue.", cabinet.Xname)
			}
			tmpCabinets[cabinet.Xname] = cabinet
		}
		slsCabinetMap[kind] = tmpCabinets
	}
	return slsCabinetMap
}

func convertManagementSwitchToSLS(s *csi.ManagementSwitch) (sls_common.GenericHardware, error) {
	switch s.SwitchType {
	case csi.ManagementSwitchTypeLeaf:
		return sls_common.GenericHardware{
			Parent:     base.GetHMSCompParent(s.Xname),
			Xname:      s.Xname,
			Type:       sls_common.MgmtSwitch,
			TypeString: base.MgmtSwitch,
			Class:      sls_common.ClassRiver,
			ExtraPropertiesRaw: sls_common.ComptypeMgmtSwitch{
				IP4Addr:          s.ManagementInterface.String(),
				Brand:            s.Brand.String(),
				Model:            s.Model,
				SNMPAuthPassword: fmt.Sprintf("vault://hms-creds/%s", s.Xname),
				SNMPAuthProtocol: "MD5",
				SNMPPrivPassword: fmt.Sprintf("vault://hms-creds/%s", s.Xname),
				SNMPPrivProtocol: "DES",
				SNMPUsername:     "testuser",

				Aliases: []string{s.Name},
			},
		}, nil
	case csi.ManagementSwitchTypeAggregation:
		fallthrough
	case csi.ManagementSwitchTypeSpine:
		return sls_common.GenericHardware{
			Parent:     base.GetHMSCompParent(s.Xname),
			Xname:      s.Xname,
			Type:       sls_common.MgmtHLSwitch,
			TypeString: base.MgmtHLSwitch,
			Class:      sls_common.ClassRiver,
			ExtraPropertiesRaw: sls_common.ComptypeMgmtHLSwitch{
				IP4Addr: s.ManagementInterface.String(),
				Brand:   s.Brand.String(),
				Model:   s.Model,
				Aliases: []string{s.Name},
			},
		}, nil

	case csi.ManagementSwitchTypeCDU:
		if base.GetHMSType(s.Xname) == base.MgmtHLSwitch {
			// This is a CDU switch in the River cabinet that is adjacent to the Hill cabinet. Use the MgmtHLSwitch type instead
			return sls_common.GenericHardware{
				Parent:     base.GetHMSCompParent(s.Xname),
				Xname:      s.Xname,
				Type:       sls_common.MgmtHLSwitch,
				TypeString: base.MgmtHLSwitch,
				Class:      sls_common.ClassRiver,
				ExtraPropertiesRaw: sls_common.ComptypeMgmtHLSwitch{
					IP4Addr: s.ManagementInterface.String(),
					Brand:   s.Brand.String(),
					Model:   s.Model,
					Aliases: []string{s.Name},
				},
			}, nil
		}

		return sls_common.GenericHardware{
			Parent:     base.GetHMSCompParent(s.Xname),
			Xname:      s.Xname,
			Type:       sls_common.CDUMgmtSwitch,
			TypeString: base.CDUMgmtSwitch,
			Class:      sls_common.ClassMountain,
			ExtraPropertiesRaw: sls_common.ComptypeCDUMgmtSwitch{
				Brand:   s.Brand.String(),
				Model:   s.Model,
				Aliases: []string{s.Name},
			},
		}, nil
	}

	return sls_common.GenericHardware{}, fmt.Errorf("unknown management switch type: %s", s.SwitchType)
}

func extractSwitchesfromReservations(subnet *csi.IPV4Subnet) ([]csi.ManagementSwitch, error) {
	var switches []csi.ManagementSwitch
	for _, reservation := range subnet.IPReservations {
		if strings.HasPrefix(reservation.Name, "sw-spine") {
			switches = append(switches, csi.ManagementSwitch{
				Xname:               reservation.Comment,
				Name:                reservation.Name,
				SwitchType:          csi.ManagementSwitchTypeSpine,
				ManagementInterface: reservation.IPAddress,
			})
		}
		if strings.HasPrefix(reservation.Name, "sw-agg") {
			switches = append(switches, csi.ManagementSwitch{
				Xname:               reservation.Comment,
				Name:                reservation.Name,
				SwitchType:          csi.ManagementSwitchTypeAggregation,
				ManagementInterface: reservation.IPAddress,
			})
		}
		if strings.HasPrefix(reservation.Name, "sw-leaf") {
			switches = append(switches, csi.ManagementSwitch{
				Xname:               reservation.Comment,
				Name:                reservation.Name,
				SwitchType:          csi.ManagementSwitchTypeLeaf,
				ManagementInterface: reservation.IPAddress,
			})
		}
		if strings.HasPrefix(reservation.Name, "sw-cdu") {
			switches = append(switches, csi.ManagementSwitch{
				Xname:               reservation.Comment,
				Name:                reservation.Name,
				SwitchType:          csi.ManagementSwitchTypeCDU,
				ManagementInterface: reservation.IPAddress,
			})
		}
	}

	return switches, nil
}

func convertIPV4NetworksToSLS(networks *[]csi.IPV4Network) map[string]sls_common.Network {
	slsNetworks := make(map[string]sls_common.Network, len(*networks))

	for _, network := range *networks {
		// TODO enforce the network name to have no spaces
		slsNetwork := convertIPV4NetworkToSLS(&network)
		slsNetworks[slsNetwork.Name] = slsNetwork
	}

	return slsNetworks
}

func convertIPV4NetworkToSLS(n *csi.IPV4Network) sls_common.Network {
	subnets := make([]sls_common.IPV4Subnet, len(n.Subnets))
	for i, subnet := range n.Subnets {
		subnets[i] = convertIPV4SubnetToSLS(subnet)
	}

	return sls_common.Network{
		Name:     n.Name,
		FullName: n.FullName,
		Type:     n.NetType,
		IPRanges: []string{n.CIDR},
		ExtraPropertiesRaw: sls_common.NetworkExtraProperties{
			Comment:   n.Comment,
			CIDR:      n.CIDR,
			MTU:       n.MTU,
			VlanRange: n.VlanRange,
			Subnets:   subnets,
		},
	}
}

func convertIPV4SubnetToSLS(s *csi.IPV4Subnet) sls_common.IPV4Subnet {
	ipReservations := make([]sls_common.IPReservation, len(s.IPReservations))
	for i, ipReservation := range s.IPReservations {
		ipReservations[i] = convertIPReservationToSLS(&ipReservation)
	}

	return sls_common.IPV4Subnet{
		Name:             s.Name,
		FullName:         s.FullName,
		CIDR:             s.CIDR.String(),
		VlanID:           s.VlanID,
		Comment:          s.Comment,
		Gateway:          s.Gateway,
		DHCPStart:        s.DHCPStart,
		DHCPEnd:          s.DHCPEnd,
		ReservationStart: s.ReservationStart,
		ReservationEnd:   s.ReservationEnd,
		IPReservations:   ipReservations,
	}
}

func convertIPReservationToSLS(s *csi.IPReservation) sls_common.IPReservation {
	return sls_common.IPReservation{
		IPAddress: s.IPAddress,
		Name:      s.Name,
		Comment:   s.Comment,
		Aliases:   s.Aliases,
	}
}
