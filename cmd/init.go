/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	sicFiles "stash.us.cray.com/MTL/sic/internal/files"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init generates the directory structure for a new system rooted in a directory matching the system-name argument",
	Long:  `init generates a scaffolding the Shasta 1.4 configuration payload.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		// Initialize the global viper
		v := viper.GetViper()
		var conf shasta.SystemConfig

		// TODO: Move this to an ARG
		// We use the system-name for a directory.  Make sure it is set.
		if v.GetString("system-name") == "" {
			log.Fatalf("system-name is not set")
		}
		basepath, err := setupDirectories(v.GetString("system-name"), v)
		// Everything should be set up properly in the global viper now.
		// Unmarshaling it to a struct at this point also verifies that the
		// resulting struct is valid.

		err = v.Unmarshal(&conf)
		if err != nil {
			log.Fatalf("unable to decode configuration into struct, %v \n", err)
		}

		// The installation requires a set of information in order to proceed
		// First, we need some kind of representation of the physical hardware
		// That is generally represented through the hmn_connections.json file
		// which is literally a cabling map with metadata about the NCNs.
		//
		// From the hmn_connections file, we can create a set of HMNRow objects
		// to use for populating SLS.
		hmnRows, err := loadHMNConnectionsFile(v.GetString("hmn-connections"))
		if err != nil {
			log.Fatalf("unable to load hmn connections, %v \n", err)
		}
		//
		// SLS also needs to know about our networking configuration.  In order to do that,
		// we need to split one or more large CIDRs into subnets per cabinet and add that to
		// the SLS configuration.  We have sensible defaults, but most of this needs to be
		// handled through ip address math with out internal ipam library
		shastaNetworks, err := BuildLiveCDNetworks(conf, v)
		// This is techincally sufficient to generate an SLSState object, but to do so now
		// would not include extended information about the NCNs and Network Switches.
		//
		// The first step in building the NCN map is to read the NCN Metadata file
		ncnMeta, err := sicFiles.ReadNodeCSV(v.GetString("ncn-metadata"))
		log.Printf("ncnMeta: %v\n", ncnMeta)
		// *** Loading Data Complete **** //
		// *** Begin Enrichment *** //
		// Alone, this metadata isn't enough.  We need to enrich it by converting from the
		// simple metadata structure to a more useful shasta.LogicalNCN structure
		var ncns []*shasta.LogicalNCN
		for _, node := range ncnMeta {
			ncns = append(ncns, node.AsLogicalNCN()) // Conversion logic in here is simplistic
		}
		//
		// Now, we have all the raw data we need to build out our system map/configuration
		// before committing it to disk as the various configuration files we need.
		//
		// To enrich our data, we need to allocate IPs for the management network switches and
		// NCNs and pair MACs with IPs in the NCN structures
		shasta.AllocateIps(ncns, shastaNetworks) // This function has no return because it is working with lists of pointers.
		log.Printf("ncns: %v\n", ncns)
		// Finally, the data is properly enriched and we can begin shaping it for use.
		// *** Enrichment Complete *** //
		// *** Commence Shaping for use *** //
		// First, our sls generator needs a list of networks and not a map of them
		var networks []shasta.IPV4Network
		for name, network := range shastaNetworks {
			if network.Name == "" {
				network.Name = name
			}
			networks = append(networks, *network)
		}
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
		switchNet, err := shastaNetworks["NMN"].LookUpSubnet("bootstrap_dhcp")
		switches, _ := extractSwitchesfromReservations(switchNet)
		log.Println("Found Switches:", switches)
		slsSwitches := make(map[string]sls_common.GenericHardware)
		for _, mySwitch := range switches {
			log.Println("Found Switch:", mySwitch.Xname)
			slsSwitches[mySwitch.Xname] = convertManagemenetSwitchToSLS(&mySwitch)
		}

		inputState := shasta.SLSGeneratorInputState{
			// TODO What about the ManagementSwitch?
			// ManagementSwitches: should be an array of sls_common.Hardware xname and ip addr are crucial
			ManagementSwitches:  slsSwitches,
			RiverCabinets:       getCabinets(sls_common.ClassRiver, v.GetInt("starting-river-cabinet"), cabinetSubnets[0:numRiver]),
			HillCabinets:        getCabinets(sls_common.ClassHill, v.GetInt("starting-hill-cabinet"), cabinetSubnets[numRiver:numRiver+numHill]),
			MountainCabinets:    getCabinets(sls_common.ClassMountain, v.GetInt("starting-mountain-cabinet"), cabinetSubnets[numRiver+numHill:]),
			MountainStartingNid: v.GetInt("starting-mountain-nid"),
			Networks:            convertIPV4NetworksToSLS(&networks),
		}
		slsState := shasta.GenerateSLSState(inputState, hmnRows)

		err = sicFiles.WriteJSONConfig(filepath.Join(basepath, "sls_input_file.json"), &slsState)
		if err != nil {
			log.Fatalln("Failed to encode SLS state:", err)
		}

		// Now that SLS can tell us which NCNs match with which Xnames, we need to update the IP Reservations
		tempNcns, err := shasta.ExtractSLSNCNs(&slsState)
		if err != nil {
			log.Panic(err)
		}
		tempSubnet, err := shastaNetworks["NMN"].LookUpSubnet("bootstrap_dhcp")
		if err != nil {
			log.Panic(err)
		} else {
			for _, reservation := range tempSubnet.IPReservations {
				for index, ncn := range tempNcns {
					if reservation.Comment == ncn.Xname {
						reservation.Name = ncn.Hostnames[0]
						log.Printf("Setting hostname to %v for %v. \n", reservation.Name, reservation.Comment)
						tempSubnet.IPReservations[index] = reservation
					}
				}
			}
		}

		conf.IPV4Resolvers = strings.Split(viper.GetString("ipv4-resolvers"), ",")
		conf.SiteServices.NtpPoolHostname = conf.NtpPoolHostname
		sicFiles.WriteYAMLConfig(filepath.Join(basepath, "system_config.yaml"), conf)

		sicFiles.WriteJSONConfig(filepath.Join(basepath, "credentials/root_password.json"), shasta.DefaultRootPW)
		sicFiles.WriteJSONConfig(filepath.Join(basepath, "credentials/bmc_password.json"), shasta.DefaultBMCPW)
		sicFiles.WriteJSONConfig(filepath.Join(basepath, "credentials/mgmt_switch_password.json"), shasta.DefaultNetPW)

		WriteDNSMasqConfig(basepath, ncns, shastaNetworks)
		WriteConmanConfig(filepath.Join(basepath, "conman.conf"), ncns, conf)
		WriteMetalLBConfigMap(basepath, conf, shastaNetworks)
		WriteBaseCampData(filepath.Join(basepath, "data.json"), conf, &slsState, ncnMeta)
		WriteNetworkFiles(basepath, shastaNetworks)

		if v.GetString("manifest-release") != "" {
			initiailzeManifestDir(shasta.DefaultManifestURL, "release/shasta-1.4", filepath.Join(basepath, "loftsman-manifests"))
		}
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
	initCmd.Flags().String("ipv4-resolvers", "8.8.8.8, 9.9.9.9", "List of IP Addresses for DNS")
	initCmd.Flags().String("v2-registry", "https://packages.local/", "URL for default v2 registry used for both helm and containers")
	initCmd.Flags().String("rpm-repository", "https://packages.local/repository/shasta-master", "URL for default rpm repository")

	// Default IPv4 Networks
	initCmd.Flags().String("nmn-cidr", shasta.DefaultNMNString, "Overall IPv4 CIDR for all Node Management subnets")
	initCmd.Flags().String("hmn-cidr", shasta.DefaultHMNString, "Overall IPv4 CIDR for all Hardware Management subnets")
	initCmd.Flags().String("can-cidr", shasta.DefaultCANString, "Overall IPv4 CIDR for all Customer Access subnets")
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
	initCmd.Flags().String("spine-switch-xnames", "", "Comma separated list of xnames for spine switches")
	initCmd.Flags().String("leaf-switch-xnames", "", "Comma separated list of xnames for leaf switches")
	initCmd.Flags().String("bgp-asn", "65533", "The autonomous system number for BGP conversations")
	initCmd.Flags().Int("management-net-ips", 20, "Number of ip addresses to reserve in each vlan for the management network")

	// Use these flags to set the default ncn bmc credentials for bootstrap
	initCmd.Flags().String("bootstrap-ncn-bmc-user", "", "Username for connecting to the BMC on the initial NCNs")
	initCmd.Flags().String("bootstrap-ncn-bmc-pass", "", "Password for connecting to the BMC on the initial NCNs")

	// Dealing with an SLS file
	initCmd.Flags().String("from-sls-file", "", "SLS File Location")

	// Dealing with SLS precursors
	initCmd.Flags().String("hmn-connections", "hmn_connections.json", "HMN Connections JSON Location (For generating an SLS File)")
	initCmd.Flags().String("ncn-metadata", "ncn_metadata.csv", "CSV for mapping the mac addresses of the NCNs to their xnames")

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

/*
	// Handle an SLSFile if one is provided
	var slsState sls_common.SLSState
	if v.GetString("from-sls-file") != "" {
		log.Println("Loading from ", v.GetString("from-sls-file"))
		slsState, err = shasta.ParseSLSFile(v.GetString("from-sls-file"))
		if err != nil {
			log.Fatal("Error loading the sls-file: ", err)
		}
		// Confirm that our slsState is valid
		networks, err := shasta.ExtractSLSNetworks(&slsState)
		if err != nil {
			log.Printf("Couldn't extract networks because: %d \n", err)
		} else {
			log.Printf("Extracted %d networks \n", len(networks))
		}
		switches, err := shasta.ExtractSLSSwitches(&slsState)
		if err != nil {
			log.Printf("Couldn't extract switches because: %d \n", err)
		} else {
			log.Printf("Extracted %d switches \n", len(switches))
		}
		ncns, err := shasta.ExtractSLSNCNs(&slsState)
		if err != nil {
			log.Printf("Couldn't extract ncns because: %d \n", err)
		} else {
			log.Printf("Extracted %d management ncns \n", len(ncns))
		}

	}
*/
