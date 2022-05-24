// Copyright 2021 Hewlett Packard Enterprise Development LP

package csi

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	base "github.com/Cray-HPE/hms-base"
	shcd_parser "github.com/Cray-HPE/hms-shcd-parser/pkg/shcd-parser"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	mountainChassisList = []string{"c0", "c1", "c2", "c3", "c4", "c5", "c6", "c7"}
	tdsChassisList      = []string{"c1", "c3"}

	// Regular expressions to get around humans.
	portRegex        = regexp.MustCompile(`[a-zA-Z]*(\d+)`)
	uRegex           = regexp.MustCompile(`[a-zA-Z]*(\d+)([a-zA-Z]*)`)
	computeNodeRegex = regexp.MustCompile(`(\d+$)`)
	pduuRegex        = regexp.MustCompile(`(x\d+p|pdu)(\d+)`) // Match on x3000p0 and pdu0
)

// SLSGeneratorInputState is given to the SLS config generator in order to generate the SLS config file
type SLSGeneratorInputState struct {
	ApplicationNodeConfig SLSGeneratorApplicationNodeConfig `json:"ApplicationNodeConfig"`

	ManagementSwitches  map[string]sls_common.GenericHardware `json:"ManagementSwitches"` // SLS Type: comptype_mgmt_switch
	RiverCabinets       map[string]sls_common.GenericHardware `json:"RiverCabinets"`      // SLS Type: comptype_cabinet
	HillCabinets        map[string]sls_common.GenericHardware `json:"HillCabinets"`       // SLS Type: comptype_cabinet
	MountainCabinets    map[string]sls_common.GenericHardware `json:"MountainCabinets"`   // SLS Type: comptype_cabinet
	MountainStartingNid int                                   `json:"MountainStartingNid"`

	Networks map[string]sls_common.Network `json:"Networks"`
}

// SLSGeneratorApplicationNodeConfig is given to the SLS config generator to control the application node generation in SLS
type SLSGeneratorApplicationNodeConfig struct {
	Prefixes          []string          `yaml:"prefixes"`
	PrefixHSMSubroles map[string]string `yaml:"prefix_hsm_subroles"`

	Aliases map[string][]string `yaml:"aliases"`
}

// Validate SLSGeneratorApplicationNodeConfig contents
func (applicationNodeConfig *SLSGeneratorApplicationNodeConfig) Validate() error {
	// Verify that all keys in the Alias map are valid xnames
	for xname := range applicationNodeConfig.Aliases {
		// First off verify that this is valid xname
		if !base.IsHMSCompIDValid(xname) {
			return fmt.Errorf("invalid xname for application node used as key in Aliases map: %s", xname)
		}

		// Next, verify that the xname is type of node
		if base.GetHMSType(xname) != base.Node {
			return fmt.Errorf("invalid type %s for Application xname in Aliases map: %s", base.GetHMSTypeString(xname), xname)
		}
	}

	// Verify that no nodes share the same alias
	allAliases := map[string]string{}
	for xname, aliases := range applicationNodeConfig.Aliases {
		for _, alias := range aliases {
			if allAliases[alias] != "" {
				return fmt.Errorf("found duplicate application node alias: %s for xnames %s %s", alias, allAliases[alias], xname)
			}
			allAliases[alias] = xname
		}
	}

	// Verify that there are no subrole placeholders that need replacing.
	prefixErr := make([]string, 0, 1)
	for prefix, subrole := range applicationNodeConfig.PrefixHSMSubroles {
		if subrole == SubrolePlaceHolder {
			prefixErr = append(prefixErr, prefix)
		}
	}
	if len(prefixErr) > 1 {
		return fmt.Errorf("prefixes, '%v', have no subrole mapping. Replace `%s` placeholders with valid subroles in the Application Node Config file", prefixErr, SubrolePlaceHolder)
	} else if len(prefixErr) == 1 {
		return fmt.Errorf("prefix, '%v', has no subrole mapping. Replace `%s` placeholder with a valid subrole in the Application Node Config file", prefixErr, SubrolePlaceHolder)
	}

	return nil
}

// Normalize the values of the SLSGeneratorApplicationNodeConfig structure
func (applicationNodeConfig *SLSGeneratorApplicationNodeConfig) Normalize() error {
	// All prefixes should be lower case
	normalizedPrefixes := []string{}
	for _, prefix := range applicationNodeConfig.Prefixes {
		normalizedPrefixes = append(normalizedPrefixes, strings.ToLower(prefix))
	}

	// All keys in PrefixSubroles should be lowercase
	normalizedPrefixSubroles := map[string]string{}
	for prefix, subRole := range applicationNodeConfig.PrefixHSMSubroles {
		normalizedPrefix := strings.ToLower(prefix)

		if _, present := normalizedPrefixSubroles[normalizedPrefix]; present {
			// Found a duplicate prefix, after normalization
			return fmt.Errorf("found a duplicate application node prefix after normalization - Prefix: %s, Normalized Prefix: %s", prefix, normalizedPrefix)
		}

		normalizedPrefixSubroles[normalizedPrefix] = subRole
	}

	// Normalize xnames in the Aliases map
	normalizedAliases := map[string][]string{}
	for xname, aliases := range applicationNodeConfig.Aliases {
		normalizedXname := base.NormalizeHMSCompID(xname)

		if _, present := normalizedAliases[normalizedXname]; present {
			// Found a duplicate xname, after normalization
			return fmt.Errorf("found a duplicate application node xname after normalization - Xname: %s, Normalized Xname: %s", xname, normalizedXname)
		}

		normalizedAliases[normalizedXname] = aliases
	}

	applicationNodeConfig.PrefixHSMSubroles = normalizedPrefixSubroles
	applicationNodeConfig.Prefixes = normalizedPrefixes
	applicationNodeConfig.Aliases = normalizedAliases

	return nil
}

// SLSStateGenerator is a utility that can take an SLSGeneratorInputState to create a valid SLSState
type SLSStateGenerator struct {
	logger     *zap.Logger
	inputState SLSGeneratorInputState
	hmnRows    []shcd_parser.HMNRow

	// Need a universal structure that keeps track of parents for nodes.
	nodeParents map[string]int

	// Management nodes need NIDs too.
	currentManagementNID int
	currentMountainNID   int
}

// NewSLSStateGenerator create a new instances of the state generator
func NewSLSStateGenerator(logger *zap.Logger, inputState SLSGeneratorInputState, hmnRows []shcd_parser.HMNRow) SLSStateGenerator {
	return SLSStateGenerator{
		logger:               logger,
		inputState:           inputState,
		hmnRows:              hmnRows,
		currentManagementNID: 100001,
	}
}

// GenerateSLSState will generate the SLSState
func (g *SLSStateGenerator) GenerateSLSState() sls_common.SLSState {
	// Build the sections
	allHardware := g.buildHardwareSection()
	allNetworks := g.buildNetworksSection()

	// Finally assemble the whole JSON payload.
	return sls_common.SLSState{
		Hardware: allHardware,
		Networks: allNetworks,
	}
}

func (g *SLSStateGenerator) buildHardwareSection() (allHardware map[string]sls_common.GenericHardware) {
	logger := g.logger

	/*
		Now begins the task of putting meaning to these rows. For the most part this is a simple process, the source
		column tells you what type of hardware it is and any index it might need, the source rack and location are the
		majority of what's necessary for the xname, and the destination fields tell you how to construct the connection
		objects.

		The only real trick here is the source parent field. That indicates two things:
		  1) A grouping of nodes that are phycsially located in the same chassis.
		  2) There is another device that needs to be treated differently (a CMC on a Gigabyte node is the only example
		     of this at the time of writing.)
	*/

	// We maintain 4 GenericHardware maps to keep the lookups simple.
	cabinetHardwareMap := make(map[string]sls_common.GenericHardware)
	nodeHardwareMap := make(map[string]sls_common.GenericHardware)
	connectionHardwareMap := make(map[string]sls_common.GenericHardware)
	switchHardwareMap := g.inputState.ManagementSwitches

	// Verify that all of the management switches have a corrensponding river cabinet
	for _, mySwitch := range switchHardwareMap {
		if mySwitch.Class != sls_common.ClassRiver {
			// Right now we only care about verifying that the river management switches have a corrensponding cabinet,
			// This means we are not doing any checking for the parent CDU for CDU switches.
			continue
		}

		switch mySwitch.Type {
		case sls_common.MgmtSwitch:
			// xname: xXcCwW
			parentCabinet := base.GetHMSCompParent(mySwitch.Parent)
			if _, ok := g.inputState.RiverCabinets[parentCabinet]; !ok {
				logger.Fatal("Parent river cabinet does not exist for MgmtSwitch",
					zap.Any("switch", mySwitch),
					zap.String("parentCabinet", parentCabinet),
					zap.String("xname", mySwitch.Xname),
				)
			}
		case sls_common.MgmtHLSwitch:
			// xname: xXcChHsS
			parentChassis := base.GetHMSCompParent(mySwitch.Parent)
			parentCabinet := base.GetHMSCompParent(parentChassis)
			if _, ok := g.inputState.RiverCabinets[parentCabinet]; !ok {
				logger.Fatal("Parent river cabinet does not exist for MgmtHLSwitch",
					zap.Any("switch", mySwitch),
					zap.String("parentCabinet", parentCabinet),
					zap.String("xname", mySwitch.Xname),
				)
			}
		default:
			logger.Fatal("Unknown river management switch type",
				zap.Any("switch", mySwitch))
		}
	}

	//
	// First off lets, build up the river hardware
	//

	// We need to run through the HMN connections file and build up the list of parents first.
	g.nodeParents = map[string]int{}
	for _, row := range g.hmnRows {
		// To make it so that we immediately know what parents there are, add all of them to the map
		// but with a bogus U number.
		if row.SourceParent != "" {
			g.nodeParents[row.SourceParent] = -1
		}
	}

	// River nodes and other devices connected to the HMN
	for _, row := range g.hmnRows {
		// Generate the node
		nodeHardware := g.getRiverHardwareFromRow(row)
		if nodeHardware.Xname == "" {
			logger.Debug("Found empty hardware, ignoring...", zap.Any("row", row))
			continue
		}

		// Ensure that the cabinet exists
		if _, ok := g.inputState.RiverCabinets[row.SourceRack]; !ok {
			logger.Fatal("Parent river cabinet does not exist",
				zap.Any("row", row),
				zap.String("parentCabinet", row.SourceRack), // This value is normally used to construct the xname
				zap.String("xname", nodeHardware.Xname),
			)
		}

		nodeHardwareMap[nodeHardware.Xname] = nodeHardware

		// Finally generate the network connection if there is one.
		if strings.TrimSpace(row.DestinationPort) != "" {
			nodeConnection := g.getConnectionForNode(nodeHardware, row)
			connectionHardwareMap[nodeConnection.Xname] = nodeConnection

			// Make sure the switch exists.
			_, switchExists := switchHardwareMap[nodeConnection.Parent]
			if !switchExists {
				destinationUString := strings.TrimPrefix(row.DestinationLocation, "u")
				switchXname := fmt.Sprintf("%sc0w%s", row.SourceRack, destinationUString)

				logger.Fatal("Failed to find switch in SLS Input State!",
					zap.String("switchXname", switchXname),
				)
			}
		}
	}

	// Lastly add the River Cabinets
	for xname, cab := range g.inputState.RiverCabinets {
		cabinetHardwareMap[xname] = cab
	}

	//
	// Next, build Up Hill Hardware
	//
	g.currentMountainNID = g.inputState.MountainStartingNid
	hillCabinets := g.getSortedCabinetXNames(g.inputState.HillCabinets)
	for _, xname := range hillCabinets {
		cab := g.inputState.HillCabinets[xname]

		cabinetHardwareMap[cab.Xname] = cab
		hillHardware := g.getHardwareForMountainCab(cab.Xname, sls_common.ClassHill)
		for _, hardware := range hillHardware {
			nodeHardwareMap[hardware.Xname] = hardware
		}
	}

	//
	// Finally, build up Mountain Hardware
	//
	mountainCabinets := g.getSortedCabinetXNames(g.inputState.MountainCabinets)
	for _, xname := range mountainCabinets {
		cab := g.inputState.MountainCabinets[xname]
		cabinetHardwareMap[xname] = cab

		mountainHardware := g.getHardwareForMountainCab(cab.Xname, sls_common.ClassMountain)
		for _, hardware := range mountainHardware {
			nodeHardwareMap[hardware.Xname] = hardware
		}
	}

	// Combine the maps.
	allHardware = make(map[string]sls_common.GenericHardware)
	for xname, hardware := range cabinetHardwareMap {
		if xname == "" {
			logger.Fatal("Found nil hardware in cabinets")
		}
		allHardware[xname] = hardware
	}
	for xname, hardware := range nodeHardwareMap {
		if xname == "" {
			logger.Fatal("Found nil hardware in node hardware")
		}
		allHardware[xname] = hardware
	}
	for xname, hardware := range connectionHardwareMap {
		if xname == "" {
			logger.Fatal("Found nil hardware in connection hardware")
		}
		allHardware[xname] = hardware
	}
	for xname, hardware := range switchHardwareMap {
		if xname == "" {
			logger.Fatal("Found nil hardware in switch hardware hardware")
		}
		allHardware[xname] = hardware
	}

	return
}

func (g *SLSStateGenerator) getSortedCabinetXNames(cabinets map[string]sls_common.GenericHardware) []string {
	xnames := []string{}
	for _, cab := range cabinets {
		xnames = append(xnames, cab.Xname)
	}

	sort.Strings(xnames)

	return xnames
}

//
// River hardware
//
func (g *SLSStateGenerator) getRiverHardwareFromRow(row shcd_parser.HMNRow) (hardware sls_common.GenericHardware) {
	sourceLowerCase := strings.ToLower(row.Source)

	// General idea here is to look for exceptions to this being a compute of any kind and handle those.
	if sourceLowerCase == "columbia" || strings.HasPrefix(sourceLowerCase, "sw-hsn") {
		return g.getTORHardwareFromRow(row)
	}

	// Check for PDU
	pduMatches := pduuRegex.FindStringSubmatch(sourceLowerCase)
	if len(pduMatches) == 3 {
		pduNumberString := pduMatches[2]

		return g.getPDUControllerHardwareFromRow(row, pduNumberString)
	}

	// Cooling door
	if strings.Contains(sourceLowerCase, "door") {
		return g.getDoorHardwareFromRow(row)
	}

	// Management switches
	if strings.HasPrefix(sourceLowerCase, "sw-leaf") || strings.HasPrefix(sourceLowerCase, "sw-25g") || strings.HasPrefix(sourceLowerCase, "sw-40g") || strings.HasPrefix(sourceLowerCase, "sw-leaf-bmc") {
		return g.getManagementSwitchHardwareFrom(row)
	}
	// Management switches deprecated naming
	if strings.HasPrefix(sourceLowerCase, "sw-agg") || strings.HasPrefix(sourceLowerCase, "sw-smn") {
		return g.getManagementSwitchHardwareFrom(row)
	}
	// Default to node.
	return g.getNodeHardwareFromRow(row)
}

func (g *SLSStateGenerator) getTORHardwareFromRow(row shcd_parser.HMNRow) (hardware sls_common.GenericHardware) {
	logger := g.logger

	var uInteger int
	bmcNumber := 0

	uSubmatches := uRegex.FindStringSubmatch(row.SourceLocation)
	if len(uSubmatches) < 2 {
		logger.Fatal("Attempted to run regex on source location but did not find U number!",
			zap.Any("uSubmatches", uSubmatches))
	}
	uString := uSubmatches[1]

	// Sometimes people like to not follow their own conventions (because Excel!!!!) and they tack the L or R
	// right onto the end of the U. Cool!
	danglingUBits := ""
	if len(uSubmatches) == 3 {
		danglingUBits = strings.ToLower(uSubmatches[2])
	}

	// This is also a hack, but to prevent a sheet that doesn't have parent information from messing things up,
	// look to the sublocation for offset.
	if strings.ToLower(row.SourceSubLocation) == "l" || danglingUBits == "l" {
		bmcNumber = 1
	} else if strings.ToLower(row.SourceSubLocation) == "r" || danglingUBits == "r" {
		bmcNumber = 2
	}

	var err error
	uInteger, err = strconv.Atoi(uString)
	if err != nil {
		logger.Fatal("Failed to parse U number string to integer!",
			zap.Error(err), zap.String("uString", uString))
	}

	torXname := fmt.Sprintf("%sc0r%db%d", row.SourceRack, uInteger, bmcNumber)

	hardware = sls_common.GenericHardware{
		Parent:     row.SourceRack,
		Xname:      torXname,
		Type:       "comptype_rtr_bmc",
		Class:      "River",
		TypeString: "RouterBMC",
		ExtraPropertiesRaw: sls_common.ComptypeRtrBmc{
			Username: fmt.Sprintf("vault://hms-creds/%s", torXname),
			Password: fmt.Sprintf("vault://hms-creds/%s", torXname),
		},
	}

	return
}

func (g *SLSStateGenerator) getPDUControllerHardwareFromRow(row shcd_parser.HMNRow, pduNumberString string) (hardware sls_common.GenericHardware) {
	logger := g.logger

	pduInteger, err := strconv.Atoi(pduNumberString)
	if err != nil {
		logger.Fatal("Failed to parse PDU number string to integer!",
			zap.Error(err), zap.String("pduNumberString", pduNumberString))
	}

	// Note: the PDU integer is being treated as PDU Cabinet controller number
	// Which in this case make sense, as a controling PDU is connected to the HMN network
	pduXname := fmt.Sprintf("%sm%d", row.SourceRack, pduInteger)

	hardware = sls_common.GenericHardware{
		Parent:     row.SourceRack,
		Xname:      pduXname,
		Type:       sls_common.CabinetPDUController,
		Class:      sls_common.ClassRiver,
		TypeString: base.CabinetPDUController,
	}

	return
}

func (g *SLSStateGenerator) getDoorHardwareFromRow(row shcd_parser.HMNRow) (hardware sls_common.GenericHardware) {
	g.logger.Warn("Cooling door found, but xname does not yet exist for cooling doors!", zap.Any("row", row))

	return
}

func (g *SLSStateGenerator) getManagementSwitchHardwareFrom(row shcd_parser.HMNRow) (hardware sls_common.GenericHardware) {
	// Not all SHCDs have the management switch connection information in the HMN connections tables,
	// and we are provided switch information via switch_metadata
	// The HMN connection information is not required for discovery.
	// sw-leaf, sw-25g, sw-40g, sw-leafbmc, or deprecated sw-agg, sw-smn
	g.logger.Warn("Ignoring management Switch found in hmn_connections, management switch information is solely from from switch_metadata.csv", zap.Any("row", row))

	return
}

func (g *SLSStateGenerator) isApplicationNode(sourceLowerCase string) (isApplicationNode bool, subRole string) {
	applicationNodeConfig := g.inputState.ApplicationNodeConfig

	// Merge default Application node prefixes with the user provided prefixes.
	prefixes := []string{}
	prefixes = append(prefixes, applicationNodeConfig.Prefixes...)
	prefixes = append(prefixes, DefaultApplicationNodePrefixes...)

	// Merge default Application node subroles with the user provided subroles. User provided subroles can override the default subroles
	subRoles := map[string]string{}
	for prefix, subRole := range DefaultApplicationNodeSubroles {
		subRoles[prefix] = subRole
	}
	for prefix, subRole := range applicationNodeConfig.PrefixHSMSubroles {
		subRoles[prefix] = subRole
	}

	// Check source to see if it matches any know application node prefix
	for _, prefix := range prefixes {
		if strings.HasPrefix(sourceLowerCase, prefix) {
			// Found an application node!
			isApplicationNode = true
			subRole = subRoles[prefix]
			return
		}
	}

	// Not an application node
	return false, ""
}

func (g *SLSStateGenerator) getApplicationNodeAlias(xname string) []string {
	// Get the aliases for the application node (if it exists)
	return g.inputState.ApplicationNodeConfig.Aliases[xname]
}

func (g *SLSStateGenerator) getNodeHardwareFromRow(row shcd_parser.HMNRow) (hardware sls_common.GenericHardware) {
	logger := g.logger

	sourceLowerCase := strings.ToLower(row.Source)
	role := "Compute"
	subRole := ""
	thisNodeExtraProperties := sls_common.ComptypeNode{}

	// First things first: figure out what this thing is.
	if strings.HasPrefix(sourceLowerCase, "mn") {
		role = "Management"
		subRole = "Master"

		indexString := strings.TrimPrefix(sourceLowerCase, "mn")

		indexNumber, err := strconv.Atoi(indexString)
		if err != nil {
			logger.Fatal("Failed to parse index number string to integer!",
				zap.Error(err), zap.String("indexString", indexString))
		}

		managementAlias := fmt.Sprintf("ncn-m%03d", indexNumber)

		thisNodeExtraProperties.NID = g.currentManagementNID
		thisNodeExtraProperties.Aliases = append(thisNodeExtraProperties.Aliases, managementAlias)

		g.currentManagementNID++
	} else if strings.HasPrefix(sourceLowerCase, "wn") {
		role = "Management"
		subRole = "Worker"

		indexString := strings.TrimPrefix(sourceLowerCase, "wn")

		indexNumber, err := strconv.Atoi(indexString)
		if err != nil {
			logger.Fatal("Failed to parse index number string to integer!",
				zap.Error(err), zap.String("indexString", indexString))
		}

		managementAlias := fmt.Sprintf("ncn-w%03d", indexNumber)

		thisNodeExtraProperties.NID = g.currentManagementNID
		thisNodeExtraProperties.Aliases = append(thisNodeExtraProperties.Aliases, managementAlias)

		g.currentManagementNID++
	} else if strings.HasPrefix(sourceLowerCase, "sn") {
		role = "Management"
		subRole = "Storage"

		indexString := strings.TrimPrefix(sourceLowerCase, "sn")

		indexNumber, err := strconv.Atoi(indexString)
		if err != nil {
			logger.Fatal("Failed to parse index number string to integer!",
				zap.Error(err), zap.String("indexString", indexString))
		}

		managementAlias := fmt.Sprintf("ncn-s%03d", indexNumber)

		thisNodeExtraProperties.NID = g.currentManagementNID
		thisNodeExtraProperties.Aliases = append(thisNodeExtraProperties.Aliases, managementAlias)

		g.currentManagementNID++
	} else if strings.HasPrefix(sourceLowerCase, "nid") || strings.HasPrefix(sourceLowerCase, "cn") {
		// Computes are the hardest it would seem. They can be either nid000001 or cn-01 or cn01...maddening.
		// Even more regular expressions!
		nidNumberMatches := computeNodeRegex.FindStringSubmatch(row.Source)
		if len(nidNumberMatches) < 2 {
			logger.Fatal("Attempted to run regex on source location but did not find NID number!",
				zap.Any("nidNumberMatches", nidNumberMatches))
		}
		nidNumberString := nidNumberMatches[1]

		nidNumber, err := strconv.Atoi(nidNumberString)
		if err != nil {
			logger.Fatal("Failed to parse NID number string to integer!",
				zap.Error(err), zap.String("nidNumberString", nidNumberString))
		}

		thisNodeExtraProperties.NID = nidNumber

		nidAlias := fmt.Sprintf("nid%06d", nidNumber)
		thisNodeExtraProperties.Aliases = append(thisNodeExtraProperties.Aliases, nidAlias)
	} else if isApplicationNode, appSubrole := g.isApplicationNode(sourceLowerCase); isApplicationNode {
		role = "Application"
		subRole = appSubrole
	} else if strings.Contains(sourceLowerCase, "cmc") {
		role = "System"
	} else {
		logger.Warn("Found unknown source prefix! If this is expected to be an Application node, please update application_node_config.yaml",
			zap.Any("row", row))
		return
	}

	// These are generic.
	thisNodeExtraProperties.Role = role
	thisNodeExtraProperties.SubRole = subRole

	// Now we have to check to see if this node has a "parent" entity.
	// If it does, then the BMC number will not just be 0. It's a bit of a hack, but we'll define the BMC number to
	// be the modulo of the NID number and 4 (which is how many nodes are currently in the multi-node enclosures
	// ...like I said, hack). And of course the U number just becomes that of the parent.
	var uInteger int
	bmcNumber := 0
	if strings.TrimSpace(row.SourceParent) != "" {
		// First find the slot number.
		parentU, sourceParentExists := g.nodeParents[row.SourceParent]
		if sourceParentExists && parentU != -1 {
			uInteger = parentU
		} else {
			// Find the row with the parent.
			parentRow := g.findRowWithSource(row.SourceParent)
			if parentRow == (shcd_parser.HMNRow{}) {
				logger.Fatal("Failed to find matching row for specified parent!",
					zap.Any("row", row))
			}

			// Get the U number and add it to the lookup.
			uString := strings.TrimPrefix(parentRow.SourceLocation, "u")

			var err error
			uInteger, err = strconv.Atoi(uString)
			if err != nil {
				logger.Fatal("Failed to parse parent U number string to integer!",
					zap.Error(err), zap.String("uString", uString))
			}

			g.nodeParents[row.SourceParent] = uInteger
		}

		// Now the BMC number.
		bmcNumber = ((thisNodeExtraProperties.NID - 1) % 4) + 1
	} else {
		uSubmatches := uRegex.FindStringSubmatch(row.SourceLocation)
		if len(uSubmatches) < 2 {
			logger.Fatal("Attempted to run regex on source location but did not find U number!",
				zap.Any("uSubmatches", uSubmatches))
		}
		uString := uSubmatches[1]

		// Sometimes people like to not follow their own conventions (because Excel!!!!) and they tack the L or R
		// right onto the end of the U. Cool!
		danglingUBits := ""
		if len(uSubmatches) == 3 {
			danglingUBits = strings.ToLower(uSubmatches[2])
		}

		// This is also a hack, but to prevent a sheet that doesn't have parent information from messing things up,
		// look to the sublocation for offset.
		if strings.ToLower(row.SourceSubLocation) == "l" || danglingUBits == "l" {
			bmcNumber = 1
		} else if strings.ToLower(row.SourceSubLocation) == "r" || danglingUBits == "r" {
			bmcNumber = 2
		}

		var err error
		uInteger, err = strconv.Atoi(uString)
		if err != nil {
			logger.Fatal("Failed to parse U number string to integer!",
				zap.Error(err), zap.String("uString", uString))
		}
	}

	// At this point we either have a genuine node or we have a parent of some sort (i.e., a CMC for a Gigabyte node).
	// We need to distinguish that as it has an impact on the type. We also want to make sure it's actually plugged in.

	// Start by seeing if this is a parent to something else.
	_, isAParent := g.nodeParents[row.Source]
	if isAParent {
		// If it is, then the type is actually comptype_chassis_bmc.
		hardware = sls_common.GenericHardware{
			Parent:     row.SourceRack,
			Xname:      fmt.Sprintf("%sc0s%db999", row.SourceRack, uInteger),
			Type:       "comptype_chassis_bmc",
			Class:      "River",
			TypeString: "ChassisBMC",
		}
	} else {
		nodeXname := fmt.Sprintf("%sc0s%db%dn0", row.SourceRack, uInteger, bmcNumber)

		if thisNodeExtraProperties.Role == "Application" {
			// If this is an Application node lets get its aliases of it (if they exist)
			aliases := g.getApplicationNodeAlias(nodeXname)
			thisNodeExtraProperties.Aliases = append(thisNodeExtraProperties.Aliases, aliases...)
		}

		hardware = sls_common.GenericHardware{
			Parent:             fmt.Sprintf("%sc0s%db%d", row.SourceRack, uInteger, bmcNumber),
			Xname:              nodeXname,
			Type:               "comptype_node",
			Class:              "River",
			TypeString:         "Node",
			ExtraPropertiesRaw: thisNodeExtraProperties,
		}
	}

	return
}

func (g *SLSStateGenerator) getConnectionForNode(node sls_common.GenericHardware, row shcd_parser.HMNRow) (
	connection sls_common.GenericHardware) {
	destinationUString := strings.TrimPrefix(row.DestinationLocation, "u")

	// Because of "reasons" the port/jack string is either prefixed with a `j` or a `p`. To combat this, use regex.
	portSubmatches := portRegex.FindStringSubmatch(row.DestinationPort)
	if len(portSubmatches) < 2 {
		g.logger.Fatal("Attempted to run regex on destination port but did not find port number!",
			zap.Any("portSubmatches", portSubmatches),
			zap.Any("row", row))
	}
	destinationJackString := portSubmatches[1]

	var destinationXname string
	if strings.HasSuffix(string(node.Type), "bmc") || node.Type == sls_common.CabinetPDUController {
		// This this type *IS* the BMC or PDU, then don't use the parent, use the xname.
		destinationXname = node.Xname
	} else {
		destinationXname = node.Parent
	}

	// Determine Switch and MgmtSwitchConnector xnames
	switchName := fmt.Sprintf("%sc0w%s", row.DestinationRack, destinationUString)
	connectorXname := fmt.Sprintf("%sc0w%sj%s", row.DestinationRack, destinationUString, destinationJackString)

	// Get the brand for this switch
	mgmtSwitch, ok := g.inputState.ManagementSwitches[switchName]
	if !ok {
		g.logger.Fatal("Unable to find management switch",
			zap.String("switchName", switchName),
			zap.String("connectorXname", switchName),
			zap.String("destinationXname", destinationXname))
	}

	ep, ok := mgmtSwitch.ExtraPropertiesRaw.(sls_common.ComptypeMgmtSwitch)
	if !ok {
		g.logger.Fatal("Unable to get management switch extra properties",
			zap.String("switchName", switchName),
			zap.String("connectorXname", switchName),
			zap.String("destinationXname", destinationXname))
	}
	switchBrand := ep.Brand

	if switchBrand == "" {
		g.logger.Fatal("Management Switch brand found not provided for switch",
			zap.String("switchName", switchName),
			zap.String("connectorXname", switchName),
			zap.String("destinationXname", destinationXname))
	}

	// Calculate the vendor name for the ethernet interfaces
	// Dell switches use: ethernet1/1/1
	// Aruba switches use: 1/1/1
	var vendorName string
	switch switchBrand {
	case ManagementSwitchBrandDell.String():
		vendorName = fmt.Sprintf("ethernet1/1/%s", destinationJackString)
	case ManagementSwitchBrandAruba.String():
		vendorName = fmt.Sprintf("1/1/%s", destinationJackString)
	case ManagementSwitchBrandMellanox.String():
		// This should only occur when the HMN connections says that a BMC is connected to the
		// spine/leaf switch. Which should not happen.
		g.logger.Fatal("Currently do no support MgmtSwitchConnector for Mellanox switches",
			zap.Any("switchBrand", switchBrand),
			zap.String("switchName", switchName),
			zap.String("connectorXname", switchName),
			zap.String("destinationXname", destinationXname))
	default:
		g.logger.Fatal("Unknown Management Switch brand found for switch",
			zap.Any("switchBrand", switchBrand),
			zap.String("switchName", switchName),
			zap.String("connectorXname", switchName),
			zap.String("destinationXname", destinationXname))
	}

	connection = sls_common.GenericHardware{
		Parent:     switchName,
		Xname:      connectorXname,
		Type:       "comptype_mgmt_switch_connector",
		Class:      "River",
		TypeString: "MgmtSwitchConnector",
		ExtraPropertiesRaw: sls_common.ComptypeMgmtSwitchConnector{
			NodeNics:   []string{destinationXname},
			VendorName: vendorName,
		},
	}

	return
}

func (g *SLSStateGenerator) findRowWithSource(sourceParent string) shcd_parser.HMNRow {
	sourceParentLowerCase := strings.ToLower(sourceParent)
	for _, row := range g.hmnRows {
		if strings.ToLower(row.Source) == sourceParentLowerCase {
			return row
		}
	}

	return shcd_parser.HMNRow{}
}

//
// Mountain and Hill hardware
//
func (g *SLSStateGenerator) getHardwareForMountainCab(cabXname string, cabClass sls_common.CabinetType) (nodes []sls_common.GenericHardware) {
	logger := g.logger

	var chassisList []string
	switch cabClass {
	case sls_common.ClassMountain:
		chassisList = mountainChassisList
	case sls_common.ClassHill:
		chassisList = tdsChassisList
	default:
		logger.Fatal("Unable to genreate mountain hardware for cabinet class",
			zap.Any("cabClass", cabClass),
			zap.String("cabNname", cabXname),
		)
	}

	for _, chassis := range chassisList {
		// Start with the CMM
		cmm := sls_common.GenericHardware{
			Parent:     cabXname,
			Xname:      fmt.Sprintf("%s%s", cabXname, chassis),
			Type:       "comptype_chassis_bmc",
			Class:      sls_common.CabinetType(cabClass),
			TypeString: "ChassisBMC",
		}
		nodes = append(nodes, cmm)

		for slot := 0; slot < 8; slot++ {
			for bmc := 0; bmc < 2; bmc++ {
				for node := 0; node < 2; node++ {
					newNode := sls_common.GenericHardware{
						Parent:     fmt.Sprintf("%s%ss%db%d", cabXname, chassis, slot, bmc),
						Xname:      fmt.Sprintf("%s%ss%db%dn%d", cabXname, chassis, slot, bmc, node),
						Type:       "comptype_node",
						Class:      sls_common.CabinetType(cabClass),
						TypeString: "Node",
						ExtraPropertiesRaw: sls_common.ComptypeNode{
							NID:     g.currentMountainNID,
							Role:    "Compute",
							Aliases: []string{fmt.Sprintf("nid%06d", g.currentMountainNID)},
						},
					}
					nodes = append(nodes, newNode)

					g.currentMountainNID++
				}
			}
		}
	}

	return
}

//
// Networks
//
func (g *SLSStateGenerator) buildNetworksSection() (allNetworks map[string]sls_common.Network) {
	allNetworks = g.inputState.Networks

	// This would be a good place to do any modifications to the given network data.
	// For right now, we leave them be.

	return
}

// GenerateSLSState generates new SLSState object from an input state and hmn-connections file.
func GenerateSLSState(inputState SLSGeneratorInputState, hmnRows []shcd_parser.HMNRow) sls_common.SLSState {
	atomicLevel := zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atomicLevel,
	))

	atomicLevel.SetLevel(zap.InfoLevel)

	logger.Info("Beginning SLS configuration generation.")

	g := NewSLSStateGenerator(logger, inputState, hmnRows)
	return g.GenerateSLSState()
}
