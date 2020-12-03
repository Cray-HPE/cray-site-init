/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	shcd_parser "stash.us.cray.com/HMS/hms-shcd-parser/pkg/shcd-parser"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
	"stash.us.cray.com/MTL/csi/pkg/version"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init generates the directory structure for a new system rooted in a directory matching the system-name argument",
	Long: `init generates a scaffolding the Shasta 1.4 configuration payload.  It is based on several input files:
	1. The hmn_connections.json which describes the cabling for the BMCs on the NCNs
	2. The ncn_metadata.csv file documents the MAC addresses of the NCNs to be used in this installation
	   NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
	3. The switch_metadata.csv file which documents the Xname, Brand, Type, and Model of each switch.  Types are CDU, Leaf, Aggregation, and Spine 
	   Switch Xname,Type,Brand,Model
	
	In addition, there are many flags to impact the layout of the system.  The defaults are generally fine except for the networking flags.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		// Initialize the global viper
		v := viper.GetViper()

		// TODO: Move this to an ARG
		// We use the system-name for a directory.  Make sure it is set.
		if v.GetString("system-name") == "" {
			log.Fatalf("system-name is not set")
		}

		// Read and validate our three input files
		hmnRows, logicalNcns, switches := collectInput(v)

		// Build a set of networks we can use
		shastaNetworks, err := BuildLiveCDNetworks(v, switches)

		// Use our new networks and our list of logicalNCNs to distribute ips
		shasta.AllocateIps(logicalNcns, shastaNetworks) // This function has no return because it is working with lists of pointers.

		// Now we can finally generate the slsState
		slsState := prepareAndGenerateSLS(v, shastaNetworks, hmnRows, switches)
		// SLS can tell us which NCNs match with which Xnames, we need to update the IP Reservations
		tempNcns, err := shasta.ExtractSLSNCNs(&slsState)
		if err != nil {
			log.Panic(err) // This should never happen.  I can't really imagine how it would.
		}

		// Merge the SLS NCN list with the NCN list we got at the beginning
		for _, ncn := range logicalNcns {
			for _, tempNCN := range tempNcns {
				if ncn.Xname == tempNCN.Xname {
					// log.Printf("Found match for %v: %v \n", ncn.Xname, tempNCN)
					ncn.Hostname = tempNCN.Hostname
					ncn.Aliases = tempNCN.Aliases
					ncn.BmcPort = tempNCN.BmcPort
					// log.Println("Updated to be :", ncn)
				}
			}
		}

		// Cycle through the main networks and update the reservations and dhcp ranges as necessary
		for _, netName := range [4]string{"NMN", "HMN", "CAN", "MTL"} {

			tempSubnet, err := shastaNetworks[netName].LookUpSubnet("bootstrap_dhcp")
			if err != nil {
				log.Panic(err)
			}
			// Loop the reservations and update the NCN reservations with hostnames
			// we likely didn't have when we registered the resevation
			updateReservations(tempSubnet, logicalNcns)
			// Reset the DHCP Range to prevent overlaps
			tempSubnet.UpdateDHCPRange()
			// We expect a bootstrap_dhcp in every net, but uai_macvlan is only in
			// the NMN range for today
			if netName == "NMN" {
				tempSubnet, err = shastaNetworks[netName].LookUpSubnet("uai_macvlan")
				if err != nil {
					log.Panic(err)
				}
				updateReservations(tempSubnet, logicalNcns)
			}
		}

		// Update the SLSState with the updated network information
		_, slsState.Networks = prepareNetworkSLS(shastaNetworks)

		// Switch from a list of pointers to a list of things before we write it out
		var ncns []shasta.LogicalNCN
		for _, ncn := range logicalNcns {
			ncns = append(ncns, *ncn)
		}
		globals, err := shasta.MakeBasecampGlobals(v, shastaNetworks, "NMN", "bootstrap_dhcp", v.GetString("install-ncn"))

		writeOutput(v, shastaNetworks, slsState, ncns, switches, globals)

		// Print Summary
		fmt.Printf("\n\n===== %v Installation Summary =====\n\n", v.GetString("system-name"))
		fmt.Printf("Installation Node: %v\n", v.GetString("install-ncn"))
		fmt.Printf("Customer Access: %v GW: %v\n", v.GetString("can-cidr"), v.GetString("can-gateway"))
		fmt.Printf("\tUpstream NTP: %v\n", v.GetString("ntp-pool"))
		fmt.Printf("\tUpstream DNS: %v\n", v.GetString("ipv4-resolvers"))
		fmt.Printf("System Information\n")
		fmt.Printf("\tNCNs: %v\n", len(ncns))

		fmt.Printf("Version Information\n\t%s\n\t%s\n", version.Get().GitCommit, version.Get())
		// fmt.Printf("\tMountain Compute Cabinets: %v\n", 0) //TODO: read from SLS
		// fmt.Printf("\tRiver Compute Cabinets: %v\n", 0)    //TODO: read from SLS
		// fmt.Printf("\tHill Compute Cabinets: %v\n", 0)     //TODO: read from SLS
	},
}

func init() {
	configCmd.AddCommand(initCmd)

	// Flags with defaults for initializing a configuration

	// System Configuration Flags based on previous system_config.yml and networks_derived.yml
	initCmd.Flags().String("system-name", "sn-2024", "Name of the System")
	initCmd.Flags().String("site-domain", "dev.cray.com", "Site Domain Name")
	initCmd.Flags().String("internal-domain", "unicos.shasta", "Internal Domain Name")
	initCmd.Flags().String("ntp-pool", "time.nist.gov", "Hostname for Upstream NTP Pool")
	initCmd.MarkFlagRequired("ntp-pool")
	initCmd.Flags().String("ipv4-resolvers", "8.8.8.8, 9.9.9.9", "List of IP Addresses for DNS")
	initCmd.Flags().String("v2-registry", "https://registry.nmn/", "URL for default v2 registry used for both helm and containers")
	initCmd.Flags().String("rpm-repository", "https://packages.nmn/repository/shasta-master", "URL for default rpm repository")
	initCmd.Flags().String("can-gateway", "", "Gateway for NCNs on the CAN")
	initCmd.MarkFlagRequired("can-gateway")
	initCmd.Flags().String("ceph-cephfs-image", "dtr.dev.cray.com/cray/cray-cephfs-provisioner:0.1.0-nautilus-1.3", "The container image for the cephfs provisioner")
	initCmd.Flags().String("ceph-rbd-image", "dtr.dev.cray.com/cray/cray-rbd-provisioner:0.1.0-nautilus-1.3", "The container image for the ceph rbd provisioner")
	initCmd.Flags().String("chart-repo", "http://helmrepo.dev.cray.com:8080", "Upstream chart repo for use during the install")
	initCmd.Flags().String("docker-image-registry", "dtr.dev.cray.com", "Upstream docker registry for use during the install")
	initCmd.Flags().String("install-ncn", "ncn-m001", "Hostname of the node to be used for installation")

	// Default IPv4 Networks
	initCmd.Flags().String("nmn-cidr", shasta.DefaultNMNString, "Overall IPv4 CIDR for all Node Management subnets")
	initCmd.Flags().String("hmn-cidr", shasta.DefaultHMNString, "Overall IPv4 CIDR for all Hardware Management subnets")
	initCmd.Flags().String("can-cidr", shasta.DefaultCANString, "Overall IPv4 CIDR for all Customer Access subnets")
	initCmd.Flags().String("can-static-pool", shasta.DefaultCANStaticString, "Overall IPv4 CIDR for static Customer Access addresses")
	initCmd.Flags().String("can-dynamic-pool", shasta.DefaultCANPoolString, "Overall IPv4 CIDR for dynamic Customer Access addresses")

	initCmd.Flags().String("mtl-cidr", shasta.DefaultMTLString, "Overall IPv4 CIDR for all Provisioning subnets")
	initCmd.Flags().String("hsn-cidr", shasta.DefaultHSNString, "Overall IPv4 CIDR for all HSN subnets")

	// Bootstrap VLANS
	initCmd.Flags().Int("nmn-bootstrap-vlan", shasta.DefaultNMNVlan, "Bootstrap VLAN for the NMN")
	initCmd.Flags().Int("hmn-bootstrap-vlan", shasta.DefaultHMNVlan, "Bootstrap VLAN for the HMN")
	initCmd.Flags().Int("can-bootstrap-vlan", shasta.DefaultCANVlan, "Bootstrap VLAN for the CAN")

	// Hardware Details
	initCmd.Flags().Int("mountain-cabinets", 4, "Number of Mountain Cabinets") // 4 mountain cabinets per CDU
	initCmd.Flags().Int("starting-mountain-cabinet", 5000, "Starting ID number for Mountain Cabinets")

	initCmd.Flags().Int("river-cabinets", 1, "Number of River Cabinets")
	initCmd.Flags().Int("starting-river-cabinet", 3000, "Starting ID number for River Cabinets")

	initCmd.Flags().Int("hill-cabinets", 0, "Number of Hill Cabinets")
	initCmd.Flags().Int("starting-hill-cabinet", 9000, "Starting ID number for Hill Cabinets")

	initCmd.Flags().Int("starting-river-NID", 1, "Starting NID for Compute Nodes")
	initCmd.Flags().Int("starting-mountain-NID", 1000, "Starting NID for Compute Nodes")

	// Use these flags to prepare the basecamp metadata json
	initCmd.Flags().String("bgp-asn", "65533", "The autonomous system number for BGP conversations")
	initCmd.Flags().Int("management-net-ips", 0, "Additional number of ip addresses to reserve in each vlan for network equipment")

	// Use these flags to set the default ncn bmc credentials for bootstrap
	initCmd.Flags().String("bootstrap-ncn-bmc-user", "", "Username for connecting to the BMC on the initial NCNs")
	initCmd.MarkFlagRequired("bootstrap-ncn-bmc-user")

	initCmd.Flags().String("bootstrap-ncn-bmc-pass", "", "Password for connecting to the BMC on the initial NCNs")
	initCmd.MarkFlagRequired("bootstrap-ncn-bmc-pass")

	// Dealing with SLS precursors
	initCmd.Flags().String("hmn-connections", "hmn_connections.json", "HMN Connections JSON Location (For generating an SLS File)")
	initCmd.Flags().String("ncn-metadata", "ncn_metadata.csv", "CSV for mapping the mac addresses of the NCNs to their xnames")
	initCmd.Flags().String("switch-metadata", "switch_metadata.csv", "CSV for mapping the mac addresses of the NCNs to their xnames")

	// Loftsman Manifest Shasta-CFG
	initCmd.Flags().String("manifest-release", "", "Loftsman Manifest Release Version (leave blank to prevent manifest generation)")
	initCmd.Flags().SortFlags = false
}

func initiailzeManifestDir(url, branch, destination string) {
	// First we need a temporary directory
	dir, err := ioutil.TempDir("", "loftsman-init")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.RemoveAll(dir)
	cloneCmd := exec.Command("git", "clone", url, dir)
	out, err := cloneCmd.Output()
	if err != nil {
		log.Fatalf("cloneCommand finished with error: %s (%v)\n", out, err)
	}
	checkoutCmd := exec.Command("git", "checkout", branch)
	checkoutCmd.Dir = dir
	out, err = checkoutCmd.Output()
	if err != nil {
		if err.Error() != "exit status 1" {
			log.Fatalf("checkoutCommand finished with error: %s (%v)\n", out, err)
		}
	}
	packageCmd := exec.Command("./package/package.sh", "1.4.0")
	packageCmd.Dir = dir
	out, err = packageCmd.Output()
	if err != nil {
		log.Fatalf("packageCommand finished with error: %s (%v)\n", out, err)
	}
	targz, err := filepath.Abs(filepath.Clean(dir + "/dist/shasta-cfg-1.4.0.tgz"))
	untarCmd := exec.Command("tar", "-zxvvf", targz)
	untarCmd.Dir = destination
	out, err = untarCmd.Output()
	if err != nil {
		log.Fatalf("untarCmd finished with error: %s (%v)\n", out, err)
	}
}

func setupDirectories(systemName string, v *viper.Viper) (string, error) {
	// Set up the path for our base directory using our systemname
	basepath, err := filepath.Abs(filepath.Clean(systemName))
	if err != nil {
		return basepath, err
	}
	// Create our base directory
	if err = os.Mkdir(basepath, 0777); err != nil {
		return basepath, err
	}

	// These Directories make up the overall structure for the Configuration Payload
	// TODO: Refactor this out of the function and into defaults or some other config
	dirs := []string{
		filepath.Join(basepath, "networks"),
		filepath.Join(basepath, "manufacturing"),
		filepath.Join(basepath, "credentials"),
		filepath.Join(basepath, "dnsmasq.d"),
		filepath.Join(basepath, "cpt-files"),
		filepath.Join(basepath, "basecamp"),
	}
	// Add the Manifest directory if needed
	if v.GetString("manifest-release") != "" {
		dirs = append(dirs, filepath.Join(basepath, "loftsman-manifests"))
	}
	// Iterate through the directories and create them
	for _, dir := range dirs {
		if err := os.Mkdir(dir, 0777); err != nil {
			// log.Fatalln("Can't create directory", dir, err)
			return basepath, err
		}
	}
	return basepath, nil
}

func collectInput(v *viper.Viper) ([]shcd_parser.HMNRow, []*shasta.LogicalNCN, []*shasta.ManagementSwitch) {
	// The installation requires a set of information in order to proceed
	// First, we need some kind of representation of the physical hardware
	// That is generally represented through the hmn_connections.json file
	// which is literally a cabling map with metadata about the NCNs and
	// River Compute node BMCs, Columbia Rosetta Switches, and PDUs.
	//
	// From the hmn_connections file, we can create a set of HMNRow objects
	// to use for populating SLS.
	hmnRows, err := loadHMNConnectionsFile(v.GetString("hmn-connections"))
	if err != nil {
		log.Fatalf("unable to load hmn connections, %v \n", err)
	}
	//
	// SLS also needs to know about our networking configuration.  In order to do that,
	// we need to load the switches
	switches, err := csiFiles.ReadSwitchCSV(v.GetString("switch-metadata"))
	if err != nil {
		log.Fatalln("Couldn't extract switches", err)
	}

	// This is techincally sufficient to generate an SLSState object, but to do so now
	// would not include extended information about the NCNs and Network Switches.
	//
	// The first step in building the NCN map is to read the NCN Metadata file
	ncns, err := csiFiles.ReadNodeCSV(v.GetString("ncn-metadata"))
	if err != nil {
		log.Fatalln("Couldn't extract ncns", err)
	}
	return hmnRows, ncns, switches
}

func prepareNetworkSLS(shastaNetworks map[string]*shasta.IPV4Network) ([]shasta.IPV4Network, map[string]sls_common.Network) {
	// Fix up the network names & create the SLS friendly version of the shasta networks
	var networks []shasta.IPV4Network
	for name, network := range shastaNetworks {
		if network.Name == "" {
			network.Name = name
		}
		networks = append(networks, *network)
	}
	return networks, convertIPV4NetworksToSLS(&networks)
}

func prepareAndGenerateSLS(v *viper.Viper, shastaNetworks map[string]*shasta.IPV4Network, hmnRows []shcd_parser.HMNRow, inputSwitches []*shasta.ManagementSwitch) sls_common.SLSState {
	networks, slsNetworks := prepareNetworkSLS(shastaNetworks)

	// Generate SLS input state
	// Verify there are enough cabinet subnets
	cabinetSubnets := getCabinetSubnets(&networks)
	numRiver := v.GetInt("river-cabinets")
	numHill := v.GetInt("hill-cabinets")
	numMountain := v.GetInt("mountain-cabinets")

	numCabinets := numRiver + numHill + numMountain
	if len(cabinetSubnets) < numCabinets {
		log.Fatalln("Insufficient subnets for", numCabinets, "cabinets, has only", len(cabinetSubnets))
	} else if len(cabinetSubnets) > numCabinets {
		log.Println("Warning: Using", numCabinets, "of", len(cabinetSubnets), "available subnets")
	}

	// Management Switch Information is included in the IP Reservations for each subnet
	switchNet, err := shastaNetworks["HMN"].LookUpSubnet("bootstrap_dhcp")
	if err != nil {
		log.Fatalln("Couldn't find subnet for management switches in the HMN:", err)
	}
	reservedSwitches, _ := extractSwitchesfromReservations(switchNet)
	slsSwitches := make(map[string]sls_common.GenericHardware)
	for _, mySwitch := range reservedSwitches {
		slsSwitches[mySwitch.Xname] = convertManagemenetSwitchToSLS(&mySwitch)
	}

	// Extract Switch brands from data stored in switch_metdata.csv
	switchBrands := make(map[string]shasta.ManagementSwitchBrand)
	for _, mySwitch := range inputSwitches {
		switchBrands[mySwitch.Xname] = mySwitch.Brand
	}

	inputState := shasta.SLSGeneratorInputState{
		ManagementSwitches:     slsSwitches,
		ManagementSwitchBrands: switchBrands,
		RiverCabinets:          getCabinets(sls_common.ClassRiver, v.GetInt("starting-river-cabinet"), cabinetSubnets[0:numRiver]),
		HillCabinets:           getCabinets(sls_common.ClassHill, v.GetInt("starting-hill-cabinet"), cabinetSubnets[numRiver:numRiver+numHill]),
		MountainCabinets:       getCabinets(sls_common.ClassMountain, v.GetInt("starting-mountain-cabinet"), cabinetSubnets[numRiver+numHill:]),
		MountainStartingNid:    v.GetInt("starting-mountain-nid"),
		Networks:               slsNetworks,
	}
	slsState := shasta.GenerateSLSState(inputState, hmnRows)
	return slsState
}

func updateReservations(tempSubnet *shasta.IPV4Subnet, logicalNcns []*shasta.LogicalNCN) {
	// Loop the reservations and update the NCN reservations with hostnames
	// we likely didn't have when we registered the resevation
	for index, reservation := range tempSubnet.IPReservations {
		for _, ncn := range logicalNcns {
			if reservation.Comment == ncn.Xname {
				reservation.Name = ncn.Hostname
				reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("%v-%v", ncn.Hostname, strings.ToLower(tempSubnet.NetName)))
				reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("time-%v", strings.ToLower(tempSubnet.NetName)))
				reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("time-%v.local", strings.ToLower(tempSubnet.NetName)))
				tempSubnet.IPReservations[index] = reservation
			}
			if reservation.Comment == fmt.Sprintf("%v-mgmt", ncn.Xname) {
				reservation.Comment = reservation.Name
				reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("%v-mgmt", ncn.Hostname))
				tempSubnet.IPReservations[index] = reservation
			}
		}
		if tempSubnet.NetName == "NMN" {
			reservation.Aliases = append(reservation.Aliases, fmt.Sprintf("%v.local", reservation.Name))
			tempSubnet.IPReservations[index] = reservation
		}
	}
}

func writeOutput(v *viper.Viper, shastaNetworks map[string]*shasta.IPV4Network, slsState sls_common.SLSState, logicalNCNs []shasta.LogicalNCN, switches []*shasta.ManagementSwitch, globals interface{}) {
	basepath, _ := setupDirectories(v.GetString("system-name"), v)
	err := csiFiles.WriteJSONConfig(filepath.Join(basepath, "sls_input_file.json"), &slsState)
	if err != nil {
		log.Fatalln("Failed to encode SLS state:", err)
	}
	WriteNetworkFiles(basepath, shastaNetworks)
	v.SetConfigType("yaml")
	v.WriteConfigAs(filepath.Join(basepath, "system_config"))

	csiFiles.WriteJSONConfig(filepath.Join(basepath, "credentials/root_password.json"), shasta.DefaultRootPW)
	csiFiles.WriteJSONConfig(filepath.Join(basepath, "credentials/bmc_password.json"), shasta.DefaultBMCPW)
	csiFiles.WriteJSONConfig(filepath.Join(basepath, "credentials/mgmt_switch_password.json"), shasta.DefaultNetPW)

	for _, ncn := range logicalNCNs {
		// log.Println("Checking to see if we need CPT files for ", ncn.Hostname)
		if strings.HasPrefix(ncn.Hostname, v.GetString("install-ncn")) {
			log.Println("Generating Installer Node (CPT) interface configurations for:", ncn.Hostname)
			WriteCPTNetworkConfig(filepath.Join(basepath, "cpt-files"), ncn, shastaNetworks)
		}
	}
	WriteDNSMasqConfig(basepath, logicalNCNs, shastaNetworks)
	WriteConmanConfig(filepath.Join(basepath, "conman.conf"), logicalNCNs)
	WriteMetalLBConfigMap(basepath, v, shastaNetworks, switches)
	WriteBasecampData(filepath.Join(basepath, "basecamp/data.json"), logicalNCNs, shastaNetworks, globals)

	if v.GetString("manifest-release") != "" {
		initiailzeManifestDir(shasta.DefaultManifestURL, "release/shasta-1.4", filepath.Join(basepath, "loftsman-manifests"))
	}
}
