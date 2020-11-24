/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	base "stash.us.cray.com/HMS/hms-base"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
)

// initCmd represents the init command
var genSLSCmd = &cobra.Command{
	Use:   "gen-sls [options] <path>",
	Short: "Generates SLS input file",
	Long: `Generates SLS input file based on a Shasta 1.4 configuration and 
	HMN connections file. By default, cabinets are assumed to be one River, the 
	rest Mountain.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		var basepath string
		if len(args) < 1 {
			basepath, err = os.Getwd()
		} else {
			basepath, err = filepath.Abs(filepath.Clean(args[0]))
		}
		if err != nil {
			log.Fatalln(err)
		}

		// Load system configuration
		sysconfig, err := loadSystemConfig(filepath.Join(basepath, "system_config.yaml"))
		if err != nil {
			log.Fatalln(err)
		}

		// Load networks
		networks, err := extractNetworks(filepath.Join(basepath, "networks"))
		if err != nil {
			log.Fatalln(err)
		}

		// Load HMN Connections
		hmnRows, err := loadHMNConnectionsFile(filepath.Join(basepath, "hmn_connections.json"))
		if err != nil {
			log.Fatalln("Failed to load HMN Connections:", err)
		}

		// Initialize the global viper
		v := viper.GetViper()

		// Determine number of each type of cabinet
		numRiver := v.GetInt("river-cabinets")
		numHill := v.GetInt("hill-cabinets")
		numMountain := int(sysconfig.Cabinets) - numRiver - numHill

		if numMountain < 0 {
			log.Fatalln("Exceeded maximum of", sysconfig.Cabinets, "cabinets")
		}

		log.Println("Total Cabinets:", sysconfig.Cabinets)
		log.Println("  River:", numRiver)
		log.Println("  Hill:", numHill)
		log.Println("  Mountain:", numMountain)

		// Verify there are enough cabinet subnets
		cabinetSubnets := getCabinetSubnets(&networks)
		numCabinets := numRiver + numHill + numMountain
		if len(cabinetSubnets) < numCabinets {
			log.Fatalln("Insufficient subnets for", numCabinets, "cabinets, has only", len(cabinetSubnets))
		} else if len(cabinetSubnets) > numCabinets {
			log.Println("Warning: Using", numCabinets, "of", len(cabinetSubnets), "available subnets")
		}

		// Generate SLS input state
		inputState := shasta.SLSGeneratorInputState{
			// TODO What about the ManagementSwitch?
			// ManagementSwitches: ,
			RiverCabinets:       getCabinets(sls_common.ClassRiver, 1004, cabinetSubnets[0:numRiver]),
			HillCabinets:        getCabinets(sls_common.ClassHill, 3000, cabinetSubnets[numRiver:numRiver+numHill]),
			MountainCabinets:    getCabinets(sls_common.ClassMountain, 5000, cabinetSubnets[numRiver+numHill:]),
			MountainStartingNid: sysconfig.StartingNID,
			Networks:            convertIPV4NetworksToSLS(&networks),
		}

		slsState := shasta.GenerateSLSState(inputState, hmnRows)

		err = csiFiles.WriteJSONConfig("-", &slsState)
		if err != nil {
			log.Fatalln("Failed to encode SLS state:", err)
		}
	},
}

func init() {
	configCmd.AddCommand(genSLSCmd)
	genSLSCmd.Flags().Int16("river-cabinets", 1, "Number of River cabinets")
	genSLSCmd.Flags().Int("hill-cabinets", 0, "Number of River cabinets")

}

func getCabinetSubnets(networks *[]shasta.IPV4Network) []map[string]*shasta.IPV4Subnet {
	// Collect all subnets for each cabinet
	cabinetSubnets := make(map[string]map[string]*shasta.IPV4Subnet)
	for _, network := range *networks {
		for _, subnet := range network.Subnets {
			if strings.HasPrefix(subnet.Name, "cabinet_") {
				if _, ok := cabinetSubnets[subnet.Name]; !ok {
					cabinetSubnets[subnet.Name] = make(map[string]*shasta.IPV4Subnet)
				}
				//cabinetSubnets[subnet.Name][network.Name] = convertIPV4SubnetToSLS(&subnet)
				cabinetSubnets[subnet.Name][network.Name] = subnet
			}
		}
	}

	// Get slice of cabinet names and sort them
	cabinets := make([]string, len(cabinetSubnets))
	i := 0
	for name := range cabinetSubnets {
		cabinets[i] = name
		i++
	}

	// Assemble ordered list of cabinet subnets
	subnets := make([]map[string]*shasta.IPV4Subnet, len(cabinetSubnets))
	for i, name := range cabinets {
		subnets[i] = cabinetSubnets[name]
	}

	return subnets
}

func getCabinets(cabinetClass sls_common.CabinetType, startingIndex int, cabinetSubnets []map[string]*shasta.IPV4Subnet) map[string]sls_common.GenericHardware {
	cabinets := make(map[string]sls_common.GenericHardware)

	for i, subnets := range cabinetSubnets {
		// Convert subnets to CabinetNetworks struct
		networks := make(map[string]sls_common.CabinetNetworks, len(subnets))
		for name, subnet := range subnets {
			networks[name] = sls_common.CabinetNetworks{
				CIDR:    subnet.CIDR.String(),
				Gateway: subnet.Gateway.String(),
				VLan:    int(subnet.VlanID),
			}
		}

		cabinet := sls_common.GenericHardware{
			Parent:     "s0",
			Xname:      fmt.Sprintf("x%d", startingIndex+i),
			Class:      cabinetClass,
			Type:       sls_common.Cabinet,
			TypeString: base.Cabinet,
			ExtraPropertiesRaw: sls_common.ComptypeCabinet{
				Networks: map[string]map[string]sls_common.CabinetNetworks{"cn": networks},
			},
		}

		// River cabinets get an "ncn" Networks entry?
		if cabinet.Class == sls_common.ClassRiver {
			cabinet.ExtraPropertiesRaw.(sls_common.ComptypeCabinet).Networks["ncn"] = networks
		}

		cabinets[cabinet.Xname] = cabinet
	}

	return cabinets
}

func convertManagemenetSwitchToSLS(s *shasta.ManagementSwitch) sls_common.GenericHardware {
	return sls_common.GenericHardware{
		Parent:     base.GetHMSCompParent(s.Xname),
		Xname:      s.Xname,
		Type:       sls_common.MgmtSwitch,
		TypeString: base.MgmtSwitch,
		Class:      sls_common.ClassRiver,
		ExtraPropertiesRaw: sls_common.ComptypeMgmtSwitch{
			IP4Addr:          s.ManagementInterface.String(), // TODO Test
			Model:            s.Model,
			SNMPAuthPassword: fmt.Sprintf("vault://hms-creds/%s", s.Xname),
			SNMPAuthProtocol: "MD5",
			SNMPPrivPassword: fmt.Sprintf("vault://hms-creds/%s", s.Xname),
			SNMPPrivProtocol: "DES",
			SNMPUsername:     "testuser",

			Aliases: []string{s.Name},
		},
	}
}

func extractSwitchesfromReservations(subnet *shasta.IPV4Subnet) ([]shasta.ManagementSwitch, error) {
	var switches []shasta.ManagementSwitch
	for _, reservation := range subnet.IPReservations {
		if strings.HasPrefix(reservation.Name, "sw-spine") {
			switches = append(switches, shasta.ManagementSwitch{
				Xname:               reservation.Comment,
				Name:                reservation.Name,
				SwitchType:          "spine",
				ManagementInterface: reservation.IPAddress,
			})
		}
		if strings.HasPrefix(reservation.Name, "sw-leaf") {
			switches = append(switches, shasta.ManagementSwitch{
				Xname:               reservation.Comment,
				Name:                reservation.Name,
				SwitchType:          "leaf",
				ManagementInterface: reservation.IPAddress,
			})
		}
	}

	return switches, nil
}

func convertIPV4NetworksToSLS(networks *[]shasta.IPV4Network) map[string]sls_common.Network {
	slsNetworks := make(map[string]sls_common.Network, len(*networks))

	for _, network := range *networks {
		// TODO enforce the network name to have no spaces
		slsNetwork := convertIPV4NetworkToSLS(&network)
		slsNetworks[slsNetwork.Name] = slsNetwork
	}

	return slsNetworks
}

func convertIPV4NetworkToSLS(n *shasta.IPV4Network) sls_common.Network {
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

func convertIPV4SubnetToSLS(s *shasta.IPV4Subnet) sls_common.IPV4Subnet {
	ipReservations := make([]sls_common.IPReservation, len(s.IPReservations))
	for i, ipReservation := range s.IPReservations {
		ipReservations[i] = convertIPReservationToSLS(&ipReservation)
	}

	return sls_common.IPV4Subnet{
		Name:           s.Name,
		FullName:       s.FullName,
		CIDR:           s.CIDR.String(),
		VlanID:         s.VlanID,
		Comment:        s.Comment,
		Gateway:        s.Gateway,
		DHCPStart:      s.DHCPStart,
		DHCPEnd:        s.DHCPEnd,
		IPReservations: ipReservations,
	}
}

func convertIPReservationToSLS(s *shasta.IPReservation) sls_common.IPReservation {
	return sls_common.IPReservation{
		IPAddress: s.IPAddress,
		Name:      s.Name,
		Comment:   s.Comment,
		Aliases:   s.Aliases,
	}
}
