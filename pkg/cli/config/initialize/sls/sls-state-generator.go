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
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	shcdParser "github.com/Cray-HPE/hms-shcd-parser/pkg/shcd-parser"
	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"

	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Cray-HPE/cray-site-init/pkg/csm"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
)

var (
	// DefaultMountainChassisList contains the default list of liquid cooled chassis for a TODO MODEL cabinet
	DefaultMountainChassisList = []int{
		0,
		1,
		2,
		3,
		4,
		5,
		6,
		7,
	}
	// DefaultHillChassisList contains the default list of liquid cooled chassis for a TODO MODEL cabinet
	DefaultHillChassisList = []int{
		1,
		3,
	}
	// DefaultRiverChassisList contains the default list of air cooled chassis for standard 19 inch rack
	DefaultRiverChassisList = []int{0}

	// Regular expressions to get around humans.
	portRegex        = regexp.MustCompile(`[a-zA-Z]*(\d+)`)
	uRegex           = regexp.MustCompile(`[a-zA-Z]*(\d+)([a-zA-Z]*)`)
	computeNodeRegex = regexp.MustCompile(`(\d+$)`)
	pduuRegex        = regexp.MustCompile(`(x\d+p|pdu)(\d+)`) // Match on x3000p0 and pdu0
)

// CabinetTemplate Should this be merged with Cabinet detail?
type CabinetTemplate struct {
	Xname           xnames.Cabinet
	Model           string
	Class           slsCommon.CabinetType
	CabinetNetworks map[string]map[string]slsCommon.CabinetNetworks

	LiquidCooledChassisList []int
	AirCooledChassisList    []int
}

func (ct *CabinetTemplate) buildExtraProperties() slsCommon.ComptypeCabinet {
	return slsCommon.ComptypeCabinet{
		Model:    ct.Model,
		Networks: ct.CabinetNetworks,
	}
}

// GeneratorInputState is given to the SLS config generator in order to generate the SLS config file
type GeneratorInputState struct {
	ApplicationNodeConfig GeneratorApplicationNodeConfig `json:"ApplicationNodeConfig"`

	ManagementSwitches  map[string]slsCommon.GenericHardware `json:"ManagementSwitches"` // SLS Type: comptype_mgmt_switch
	RiverCabinets       map[string]CabinetTemplate           `json:"RiverCabinets"`      // SLS Type: comptype_cabinet
	HillCabinets        map[string]CabinetTemplate           `json:"HillCabinets"`       // SLS Type: comptype_cabinet
	MountainCabinets    map[string]CabinetTemplate           `json:"MountainCabinets"`   // SLS Type: comptype_cabinet
	MountainStartingNid int                                  `json:"MountainStartingNid"`

	Networks map[string]slsCommon.Network `json:"Networks"`
}

// GeneratorApplicationNodeConfig is given to the SLS config generator to control the application node generation in SLS
type GeneratorApplicationNodeConfig struct {
	Prefixes          []string          `yaml:"prefixes"`
	PrefixHSMSubroles map[string]string `yaml:"prefix_hsm_subroles"`

	Aliases map[string][]string `yaml:"aliases"`
}

// Validate GeneratorApplicationNodeConfig contents
func (applicationNodeConfig *GeneratorApplicationNodeConfig) Validate() error {
	// Verify that all keys in the Alias map are valid xnames
	for xname := range applicationNodeConfig.Aliases {
		// First off verify that this is valid xname
		if !xnametypes.IsHMSCompIDValid(xname) {
			return fmt.Errorf(
				"invalid xname for application node used as key in Aliases map: %s",
				xname,
			)
		}

		// Next, verify that the xname is type of node
		if xnametypes.GetHMSType(xname) != xnametypes.Node {
			return fmt.Errorf(
				"invalid type %s for Application xname in Aliases map: %s",
				xnametypes.GetHMSTypeString(xname),
				xname,
			)
		}
	}

	// Verify that no nodes share the same alias
	allAliases := map[string]string{}
	for xname, aliases := range applicationNodeConfig.Aliases {
		for _, alias := range aliases {
			if allAliases[alias] != "" {
				return fmt.Errorf(
					"found duplicate application node alias: %s for xnames %s %s",
					alias,
					allAliases[alias],
					xname,
				)
			}
			allAliases[alias] = xname
		}
	}

	// Verify that there are no subrole placeholders that need replacing.
	prefixErr := make(
		[]string,
		0,
		1,
	)
	for prefix, subrole := range applicationNodeConfig.PrefixHSMSubroles {
		if subrole == networking.SubrolePlaceHolder {
			prefixErr = append(
				prefixErr,
				prefix,
			)
		}
	}
	if len(prefixErr) > 1 {
		return fmt.Errorf(
			"prefixes, '%v', have no subrole mapping. Replace `%s` placeholders with valid subroles in the Application Node Config file",
			prefixErr,
			networking.SubrolePlaceHolder,
		)
	} else if len(prefixErr) == 1 {
		return fmt.Errorf(
			"prefix, '%v', has no subrole mapping. Replace `%s` placeholder with a valid subrole in the Application Node Config file",
			prefixErr,
			networking.SubrolePlaceHolder,
		)
	}

	return nil
}

// Normalize the values of the GeneratorApplicationNodeConfig structure
func (applicationNodeConfig *GeneratorApplicationNodeConfig) Normalize() error {
	// All prefixes should be lower case
	normalizedPrefixes := []string{}
	for _, prefix := range applicationNodeConfig.Prefixes {
		normalizedPrefixes = append(
			normalizedPrefixes,
			strings.ToLower(prefix),
		)
	}

	// All keys in PrefixSubroles should be lowercase
	normalizedPrefixSubroles := map[string]string{}
	for prefix, subRole := range applicationNodeConfig.PrefixHSMSubroles {
		normalizedPrefix := strings.ToLower(prefix)

		if _, present := normalizedPrefixSubroles[normalizedPrefix]; present {
			// Found a duplicate prefix, after normalization
			return fmt.Errorf(
				"found a duplicate application node prefix after normalization - Prefix: %s, Normalized Prefix: %s",
				prefix,
				normalizedPrefix,
			)
		}

		normalizedPrefixSubroles[normalizedPrefix] = subRole
	}

	// Normalize xnames in the Aliases map
	normalizedAliases := map[string][]string{}
	for xname, aliases := range applicationNodeConfig.Aliases {
		normalizedXname := xnametypes.NormalizeHMSCompID(xname)

		if _, present := normalizedAliases[normalizedXname]; present {
			// Found a duplicate xname, after normalization
			return fmt.Errorf(
				"found a duplicate application node xname after normalization - Xname: %s, Normalized Xname: %s",
				xname,
				normalizedXname,
			)
		}

		normalizedAliases[normalizedXname] = aliases
	}

	applicationNodeConfig.PrefixHSMSubroles = normalizedPrefixSubroles
	applicationNodeConfig.Prefixes = normalizedPrefixes
	applicationNodeConfig.Aliases = normalizedAliases

	return nil
}

// StateGenerator is a utility that can take an GeneratorInputState to create a valid SLSState
type StateGenerator struct {
	logger     *zap.Logger
	inputState GeneratorInputState
	hmnRows    []shcdParser.HMNRow

	// Need a universal structure that keeps track of parents for nodes.
	nodeParents map[string]int

	// Management nodes need NIDs too.
	currentManagementNID int
	currentMountainNID   int
}

// NewStateGenerator create a new instances of the state generator
func NewStateGenerator(logger *zap.Logger, inputState GeneratorInputState, hmnRows []shcdParser.HMNRow) StateGenerator {
	return StateGenerator{
		logger:               logger,
		inputState:           inputState,
		hmnRows:              hmnRows,
		currentManagementNID: 100001,
	}
}

// GenerateSLSState will generate the SLSState
func (g *StateGenerator) GenerateSLSState() slsCommon.SLSState {
	// Build the sections
	allHardware := g.buildHardwareSection()
	allNetworks := g.buildNetworksSection()

	// Finally assemble the whole JSON payload.
	return slsCommon.SLSState{
		Hardware: allHardware,
		Networks: allNetworks,
	}
}

func (g *StateGenerator) buildHardwareSection() (allHardware map[string]slsCommon.GenericHardware) {
	logger := g.logger

	/*
		Now begins the task of putting meaning to these rows. For the most part this is a simple process, the source
		column tells you what type of hardware it is and any index it might need, the source rack and location are the
		majority of what's necessary for the xname, and the destination fields tell you how to construct the connection
		objects.

		The only real trick here is the source parent field. That indicates two things:
		  1) A grouping of nodes that are physically located in the same chassis.
		  2) There is another device that needs to be treated differently (a CMC on a Gigabyte node is the only example
		     of this at the time of writing.)
	*/

	// We maintain 4 GenericHardware maps to keep the lookups simple.
	cabinetHardwareMap := make(map[string]slsCommon.GenericHardware)
	nodeHardwareMap := make(map[string]slsCommon.GenericHardware)
	connectionHardwareMap := make(map[string]slsCommon.GenericHardware)
	switchHardwareMap := g.inputState.ManagementSwitches

	// Verify that all of the management switches have a corresponding river cabinet
	for _, mySwitch := range switchHardwareMap {
		if mySwitch.Class != slsCommon.ClassRiver {
			// Right now we only care about verifying that the river management switches have a corresponding cabinet,
			// This means we are not doing any checking for the parent CDU for CDU switches.
			continue
		}

		switch mySwitch.Type {
		case slsCommon.MgmtSwitch:
			// xname: xXcCwW
			parentCabinet := xnametypes.GetHMSCompParent(mySwitch.Parent)

			if _, err := g.canCabinetContainAirCooledHardware(parentCabinet); err != nil {
				logger.Fatal(
					"Parent cabinet for MgmtHLSwitch can not contain air-cooled hardware",
					zap.Any(
						"switch",
						mySwitch,
					),
					zap.String(
						"parentCabinet",
						parentCabinet,
					),
					zap.String(
						"xname",
						mySwitch.Xname,
					),
					zap.Error(err),
				)
			}
		case slsCommon.MgmtHLSwitch:
			// xname: xXcChHsS
			parentChassis := xnametypes.GetHMSCompParent(mySwitch.Parent)
			parentCabinet := xnametypes.GetHMSCompParent(parentChassis)

			if _, err := g.canCabinetContainAirCooledHardware(parentCabinet); err != nil {
				logger.Fatal(
					"Parent cabinet for MgmtHLSwitch can not contain air-cooled hardware",
					zap.Any(
						"switch",
						mySwitch,
					),
					zap.String(
						"parentCabinet",
						parentCabinet,
					),
					zap.String(
						"xname",
						mySwitch.Xname,
					),
					zap.Error(err),
				)
			}
		default:
			logger.Fatal(
				"Unknown river management switch type",
				zap.Any(
					"switch",
					mySwitch,
				),
			)
		}
	}

	//
	// First off lets, build up the river hardware
	//

	// Create River Cabinets
	for _, cabinetTemplate := range g.inputState.RiverCabinets {
		riverCabinet := g.buildSLSHardware(
			cabinetTemplate.Xname,
			cabinetTemplate.Class,
			cabinetTemplate.buildExtraProperties(),
		)

		cabinetHardwareMap[cabinetTemplate.Xname.String()] = riverCabinet
	}

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
			logger.Debug(
				"Found empty hardware, ignoring...",
				zap.Any(
					"row",
					row,
				),
			)
			continue
		}

		// Ensure that the cabinet exists
		if _, err := g.canCabinetContainAirCooledHardware(row.SourceRack); err != nil {
			logger.Fatal(
				"Parent cabinet can not contain air-cooled hardware",
				zap.Any(
					"row",
					row,
				),
				zap.String(
					"parentCabinet",
					row.SourceRack,
				), // This value is normally used to construct the xname
				zap.String(
					"xname",
					nodeHardware.Xname,
				),
				zap.Error(err),
			)
		}

		nodeHardwareMap[nodeHardware.Xname] = nodeHardware

		// Finally generate the network connection if there is one.
		if strings.TrimSpace(row.DestinationPort) != "" && strings.TrimSpace(row.DestinationPort) != "0" {
			nodeConnection := g.getSwitchConnectionForHardware(
				nodeHardware,
				row,
			)
			connectionHardwareMap[nodeConnection.Xname] = nodeConnection

			// Make sure the switch exists.
			_, switchExists := switchHardwareMap[nodeConnection.Parent]
			if !switchExists {
				logger.Fatal(
					"Failed to find switch in SLS Input State!",
					zap.String(
						"switchXname",
						nodeConnection.Parent,
					),
				)
			}
		}
	}

	//
	// Next, build Up Hill Hardware
	//
	g.currentMountainNID = g.inputState.MountainStartingNid
	hillCabinets := g.getSortedCabinetXNames(g.inputState.HillCabinets)
	for _, xname := range hillCabinets {
		cabinetTemplate := g.inputState.HillCabinets[xname]

		hillHardware := g.getLiquidCooledHardwareForCabinet(cabinetTemplate)

		for _, hardware := range hillHardware {
			nodeHardwareMap[hardware.Xname] = hardware
		}
	}

	//
	// Finally, build up Mountain Hardware
	//
	mountainCabinets := g.getSortedCabinetXNames(g.inputState.MountainCabinets)
	for _, xname := range mountainCabinets {
		cabinetTemplate := g.inputState.MountainCabinets[xname]

		mountainHardware := g.getLiquidCooledHardwareForCabinet(cabinetTemplate)
		for _, hardware := range mountainHardware {
			nodeHardwareMap[hardware.Xname] = hardware
		}
	}

	// Combine the maps.
	allHardware = make(map[string]slsCommon.GenericHardware)
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

func (g *StateGenerator) getSortedCabinetXNames(cabinets map[string]CabinetTemplate) []string {
	xnames := []string{}
	for _, cab := range cabinets {
		xnames = append(
			xnames,
			cab.Xname.String(),
		)
	}

	sort.Strings(xnames)

	return xnames
}

func (g *StateGenerator) canCabinetContainAirCooledHardware(cabinetXname string) (bool, error) {
	if _, ok := g.inputState.RiverCabinets[cabinetXname]; ok {
		// River Cabinets can of course hold air-cooled hardware
		return true, nil
	} else if cabinetTemplate, ok := g.inputState.HillCabinets[cabinetXname]; ok {
		if cabinetTemplate.Model == "EX2500" {
			if len(cabinetTemplate.AirCooledChassisList) >= 1 {
				// This is an EX2500 cabinet with a air cooled chassis in it
				return true, nil
			}

			// This ia an EX2500 cabinet with no air-cooled chassis
			return false, fmt.Errorf(
				"hill cabinet (EX2500) %s does not contain any air-cooled chassis",
				cabinetXname,
			)
		}

		// Traditional Hill cabinet
		return false, fmt.Errorf(
			"hill cabinet (non EX2500) %s cannot contain air-cooled hardware",
			cabinetXname,
		)

	} else if _, ok := g.inputState.MountainCabinets[cabinetXname]; ok {
		return false, fmt.Errorf(
			"mountain cabinet %s cannot contain air-cooled hardware",
			cabinetXname,
		)
	} else {
		return false, fmt.Errorf(
			"unknown cabinet %s",
			cabinetXname,
		)
	}
}

func (g *StateGenerator) determineRiverChassis(cabinet xnames.Cabinet) (xnames.Chassis, error) {
	// Check to see if this is even a cabinet that can have river hardware
	_, err := g.canCabinetContainAirCooledHardware(cabinet.String())
	if err != nil {
		return xnames.Chassis{}, err
	}

	// Next, determine if this is a standard river cabinet for a EX2500 cabinet
	hillCabinetTemplate, hillCabinetExists := g.inputState.HillCabinets[cabinet.String()]

	chassisInteger := 0
	if hillCabinetExists {
		// This is a EX2500 cabinet with a air cooled chassis
		chassisInteger = hillCabinetTemplate.AirCooledChassisList[0]
	}

	return cabinet.Chassis(chassisInteger), nil
}

func (g *StateGenerator) parseSourceCabinetFromRow(row shcdParser.HMNRow) (xnames.Cabinet, error) {
	cabinetString := strings.TrimPrefix(
		strings.ToLower(row.SourceRack),
		"x",
	)
	cabinetInteger, err := strconv.Atoi(cabinetString)
	if err != nil {
		return xnames.Cabinet{}, err
	}

	return xnames.Cabinet{Cabinet: cabinetInteger}, nil
}

// River hardware
func (g *StateGenerator) getRiverHardwareFromRow(row shcdParser.HMNRow) (hardware slsCommon.GenericHardware) {
	sourceLowerCase := strings.ToLower(row.Source)

	// General idea here is to look for exceptions to this being a compute of any kind and handle those.
	if sourceLowerCase == "columbia" || strings.HasPrefix(
		sourceLowerCase,
		"sw-hsn",
	) {
		return g.getTORHardwareFromRow(row)
	}

	// Check for PDU
	pduMatches := pduuRegex.FindStringSubmatch(sourceLowerCase)
	if len(pduMatches) == 3 {
		pduNumberString := pduMatches[2]

		return g.getPDUControllerHardwareFromRow(
			row,
			pduNumberString,
		)
	}

	// Cooling door
	if strings.Contains(
		sourceLowerCase,
		"door",
	) {
		return g.getDoorHardwareFromRow(row)
	}

	// Management switches
	if strings.HasPrefix(
		sourceLowerCase,
		"sw-leaf",
	) || strings.HasPrefix(
		sourceLowerCase,
		"sw-25g",
	) || strings.HasPrefix(
		sourceLowerCase,
		"sw-40g",
	) || strings.HasPrefix(
		sourceLowerCase,
		"sw-leaf-bmc",
	) {
		return g.getManagementSwitchHardwareFrom(row)
	}
	// Management switches deprecated naming
	if strings.HasPrefix(
		sourceLowerCase,
		"sw-agg",
	) || strings.HasPrefix(
		sourceLowerCase,
		"sw-smn",
	) {
		return g.getManagementSwitchHardwareFrom(row)
	}
	// Default to node.
	return g.getNodeHardwareFromRow(row)
}

func (g *StateGenerator) getTORHardwareFromRow(row shcdParser.HMNRow) (hardware slsCommon.GenericHardware) {
	logger := g.logger

	// First determine the cabinet
	cabinet, err := g.parseSourceCabinetFromRow(row)
	if err != nil {
		g.logger.Fatal(
			"Failed to parse source cabinet from row!",
			zap.Error(err),
			zap.Any(
				"row",
				row,
			),
		)
	}

	// Determine the chassis
	chassis, err := g.determineRiverChassis(cabinet)
	if err != nil {
		g.logger.Fatal(
			"Failed to determine the chassis integer for rosetta!", // Find a better name for this
			zap.Error(err),
			zap.String(
				"row.SourceRack",
				row.SourceRack,
			),
			zap.Any(
				"row",
				row,
			),
		)
	}

	// Determine the rack U and BMC ordinals
	uSubmatches := uRegex.FindStringSubmatch(row.SourceLocation)
	if len(uSubmatches) < 2 {
		logger.Fatal(
			"Attempted to run regex on source location but did not find U number!",
			zap.Any(
				"uSubmatches",
				uSubmatches,
			),
		)
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
	bmcNumber := 0
	if strings.ToLower(row.SourceSubLocation) == "l" || danglingUBits == "l" {
		bmcNumber = 1
	} else if strings.ToLower(row.SourceSubLocation) == "r" || danglingUBits == "r" {
		bmcNumber = 2
	}

	// Determine the rack U of the rosetta
	uInteger, err := strconv.Atoi(uString)
	if err != nil {
		logger.Fatal(
			"Failed to parse U number string to integer!",
			zap.Error(err),
			zap.String(
				"uString",
				uString,
			),
		)
	}

	// Finally build up the xname!
	tor := chassis.RouterModule(uInteger).RouterBMC(bmcNumber)

	hardware = g.buildSLSHardware(
		tor,
		slsCommon.ClassRiver,
		slsCommon.ComptypeRtrBmc{
			Username: fmt.Sprintf(
				"vault://hms-creds/%s",
				tor.String(),
			),
			Password: fmt.Sprintf(
				"vault://hms-creds/%s",
				tor.String(),
			),
		},
	)

	return
}

func (g *StateGenerator) getPDUControllerHardwareFromRow(
	row shcdParser.HMNRow, pduNumberString string,
) (hardware slsCommon.GenericHardware) {
	logger := g.logger

	pduInteger, err := strconv.Atoi(pduNumberString)
	if err != nil {
		logger.Fatal(
			"Failed to parse PDU number string to integer!",
			zap.Error(err),
			zap.String(
				"pduNumberString",
				pduNumberString,
			),
		)
	}

	// Note: the PDU integer is being treated as PDU Cabinet controller number
	// Which in this case make sense, as a controlling PDU is connected to the HMN network
	pduXname := fmt.Sprintf(
		"%sm%d",
		row.SourceRack,
		pduInteger,
	)

	hardware = slsCommon.GenericHardware{
		Parent:     row.SourceRack,
		Xname:      pduXname,
		Type:       slsCommon.CabinetPDUController,
		Class:      slsCommon.ClassRiver,
		TypeString: xnametypes.CabinetPDUController,
	}

	return
}

func (g *StateGenerator) getDoorHardwareFromRow(row shcdParser.HMNRow) (hardware slsCommon.GenericHardware) {
	g.logger.Warn(
		"Cooling door found, but xname does not yet exist for cooling doors!",
		zap.Any(
			"row",
			row,
		),
	)

	return
}

func (g *StateGenerator) getManagementSwitchHardwareFrom(row shcdParser.HMNRow) (hardware slsCommon.GenericHardware) {
	// Not all SHCDs have the management switch connection information in the HMN connections tables,
	// and we are provided switch information via switch_metadata
	// The HMN connection information is not required for discovery.
	// sw-leaf, sw-25g, sw-40g, sw-leafbmc, or deprecated sw-agg, sw-smn
	g.logger.Warn(
		"Ignoring management Switch found in hmn_connections, management switch information is solely from from switch_metadata.csv",
		zap.Any(
			"row",
			row,
		),
	)

	return
}

func (g *StateGenerator) isApplicationNode(sourceLowerCase string) (isApplicationNode bool, subRole string) {
	applicationNodeConfig := g.inputState.ApplicationNodeConfig

	// Merge default Application node prefixes with the user provided prefixes.
	prefixes := []string{}
	prefixes = append(
		prefixes,
		applicationNodeConfig.Prefixes...,
	)
	prefixes = append(
		prefixes,
		networking.DefaultApplicationNodePrefixes...,
	)

	// Merge default Application node subroles with the user provided subroles. User provided subroles can override the default subroles
	subRoles := map[string]string{}
	for prefix, subRole := range networking.DefaultApplicationNodeSubroles {
		subRoles[prefix] = subRole
	}
	for prefix, subRole := range applicationNodeConfig.PrefixHSMSubroles {
		subRoles[prefix] = subRole
	}

	// Check source to see if it matches any know application node prefix
	for _, prefix := range prefixes {
		if strings.HasPrefix(
			sourceLowerCase,
			prefix,
		) {
			// Found an application node!
			isApplicationNode = true
			subRole = subRoles[prefix]
			return
		}
	}

	// Not an application node
	return false, ""
}

func (g *StateGenerator) getApplicationNodeAlias(xname string) []string {
	// Get the aliases for the application node (if it exists)
	return g.inputState.ApplicationNodeConfig.Aliases[xname]
}

func (g *StateGenerator) getNodeHardwareFromRow(row shcdParser.HMNRow) (hardware slsCommon.GenericHardware) {
	logger := g.logger

	sourceLowerCase := strings.ToLower(row.Source)
	role := "Compute"
	subRole := ""
	thisNodeExtraProperties := slsCommon.ComptypeNode{}

	// First things first: figure out what this thing is.
	if strings.HasPrefix(
		sourceLowerCase,
		"mn",
	) {
		role = "Management"
		subRole = "Master"

		indexString := strings.TrimPrefix(
			sourceLowerCase,
			"mn",
		)

		indexNumber, err := strconv.Atoi(indexString)
		if err != nil {
			logger.Fatal(
				"Failed to parse index number string to integer!",
				zap.Error(err),
				zap.String(
					"indexString",
					indexString,
				),
			)
		}

		managementAlias := fmt.Sprintf(
			"ncn-m%03d",
			indexNumber,
		)

		thisNodeExtraProperties.NID = g.currentManagementNID
		thisNodeExtraProperties.Aliases = append(
			thisNodeExtraProperties.Aliases,
			managementAlias,
		)

		g.currentManagementNID++
	} else if strings.HasPrefix(
		sourceLowerCase,
		"wn",
	) {
		role = "Management"
		subRole = "Worker"

		indexString := strings.TrimPrefix(
			sourceLowerCase,
			"wn",
		)

		indexNumber, err := strconv.Atoi(indexString)
		if err != nil {
			logger.Fatal(
				"Failed to parse index number string to integer!",
				zap.Error(err),
				zap.String(
					"indexString",
					indexString,
				),
			)
		}

		managementAlias := fmt.Sprintf(
			"ncn-w%03d",
			indexNumber,
		)

		thisNodeExtraProperties.NID = g.currentManagementNID
		thisNodeExtraProperties.Aliases = append(
			thisNodeExtraProperties.Aliases,
			managementAlias,
		)

		g.currentManagementNID++
	} else if strings.HasPrefix(
		sourceLowerCase,
		"sn",
	) {
		role = "Management"
		subRole = "Storage"

		indexString := strings.TrimPrefix(
			sourceLowerCase,
			"sn",
		)

		indexNumber, err := strconv.Atoi(indexString)
		if err != nil {
			logger.Fatal(
				"Failed to parse index number string to integer!",
				zap.Error(err),
				zap.String(
					"indexString",
					indexString,
				),
			)
		}

		managementAlias := fmt.Sprintf(
			"ncn-s%03d",
			indexNumber,
		)

		thisNodeExtraProperties.NID = g.currentManagementNID
		thisNodeExtraProperties.Aliases = append(
			thisNodeExtraProperties.Aliases,
			managementAlias,
		)

		g.currentManagementNID++
	} else if strings.HasPrefix(
		sourceLowerCase,
		"fmn",
	) {
		// FabricManager nodes are only supported in CSM 1.7 and later
		_, oneSevenCSM := csm.CompareMajorMinor("1.7")
		if oneSevenCSM == -1 {
			logger.Info(
				"Skipping FabricManager node - requires CSM 1.7 or later",
				zap.String("source", row.Source),
			)
			return hardware
		}

		role = "Management"
		subRole = "FabricManager"

		indexString := strings.TrimPrefix(
			sourceLowerCase,
			"fmn",
		)

		indexNumber, err := strconv.Atoi(indexString)
		if err != nil {
			logger.Fatal(
				"Failed to parse index number string to integer!",
				zap.Error(err),
				zap.String(
					"indexString",
					indexString,
				),
			)
		}

		managementAlias := fmt.Sprintf(
			"fmn%03d",
			indexNumber,
		)

		thisNodeExtraProperties.NID = g.currentManagementNID
		thisNodeExtraProperties.Aliases = append(
			thisNodeExtraProperties.Aliases,
			managementAlias,
		)

		g.currentManagementNID++

	} else if strings.HasPrefix(
		sourceLowerCase,
		"nid",
	) || strings.HasPrefix(
		sourceLowerCase,
		"cn",
	) {
		// Computes are the hardest it would seem. They can be either nid000001 or cn-01 or cn01...maddening.
		// Even more regular expressions!
		nidNumberMatches := computeNodeRegex.FindStringSubmatch(row.Source)
		if len(nidNumberMatches) < 2 {
			logger.Fatal(
				"Attempted to run regex on source location but did not find NID number!",
				zap.Any(
					"nidNumberMatches",
					nidNumberMatches,
				),
			)
		}
		nidNumberString := nidNumberMatches[1]

		nidNumber, err := strconv.Atoi(nidNumberString)
		if err != nil {
			logger.Fatal(
				"Failed to parse NID number string to integer!",
				zap.Error(err),
				zap.String(
					"nidNumberString",
					nidNumberString,
				),
			)
		}

		thisNodeExtraProperties.NID = nidNumber

		nidAlias := fmt.Sprintf(
			"nid%06d",
			nidNumber,
		)
		thisNodeExtraProperties.Aliases = append(
			thisNodeExtraProperties.Aliases,
			nidAlias,
		)
	} else if isApplicationNode, appSubrole := g.isApplicationNode(sourceLowerCase); isApplicationNode {
		role = "Application"
		subRole = appSubrole
	} else if strings.Contains(
		sourceLowerCase,
		"cmc",
	) {
		role = "System"
	} else {
		logger.Warn(
			"Found unknown source prefix! If this is expected to be an Application node, please update application_node_config.yaml",
			zap.Any(
				"row",
				row,
			),
		)
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
			if parentRow == (shcdParser.HMNRow{}) {
				logger.Fatal(
					"Failed to find matching row for specified parent!",
					zap.Any(
						"row",
						row,
					),
				)
			}

			// Get the U number and add it to the lookup.
			uString := strings.TrimPrefix(
				parentRow.SourceLocation,
				"u",
			)

			var err error
			uInteger, err = strconv.Atoi(uString)
			if err != nil {
				logger.Fatal(
					"Failed to parse parent U number string to integer!",
					zap.Error(err),
					zap.String(
						"uString",
						uString,
					),
				)
			}

			g.nodeParents[row.SourceParent] = uInteger
		}

		// Now the BMC number.
		bmcNumber = ((thisNodeExtraProperties.NID - 1) % 4) + 1
	} else {
		uSubmatches := uRegex.FindStringSubmatch(row.SourceLocation)
		if len(uSubmatches) < 2 {
			logger.Fatal(
				"Attempted to run regex on source location but did not find U number!",
				zap.Any(
					"uSubmatches",
					uSubmatches,
				),
			)
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
			logger.Fatal(
				"Failed to parse U number string to integer!",
				zap.Error(err),
				zap.String(
					"uString",
					uString,
				),
			)
		}
	}

	// Now we need to determine
	cabinet, err := g.parseSourceCabinetFromRow(row)
	if err != nil {
		g.logger.Fatal(
			"Failed to parse source cabinet from row!",
			zap.Error(err),
			zap.Any(
				"row",
				row,
			),
		)
	}

	// Next determine the chassis
	chassis, err := g.determineRiverChassis(cabinet)
	if err != nil {
		g.logger.Fatal(
			"Failed to determine the chassis integer for node!",
			zap.Error(err),
			zap.String(
				"row.SourceRack",
				row.SourceRack,
			),
			zap.Any(
				"row",
				row,
			),
		)
	}

	// At this point we either have a genuine node or we have a parent of some sort (i.e., a CMC for a Gigabyte node).
	// We need to distinguish that as it has an impact on the type. We also want to make sure it's actually plugged in.

	// Start by seeing if this is a parent to something else.
	_, isAParent := g.nodeParents[row.Source]
	if isAParent {
		// If it is, then the type is actually comptype_chassis_bmc.
		cmc := chassis.ComputeModule(uInteger).NodeBMC(999)

		hardware = g.buildSLSHardware(
			cmc,
			slsCommon.ClassRiver,
			nil,
		)
	} else {
		node := chassis.ComputeModule(uInteger).NodeBMC(bmcNumber).Node(0)

		if thisNodeExtraProperties.Role == "Application" {
			// If this is an Application node lets get its aliases of it (if they exist)
			aliases := g.getApplicationNodeAlias(node.String())
			thisNodeExtraProperties.Aliases = append(
				thisNodeExtraProperties.Aliases,
				aliases...,
			)
		}

		hardware = g.buildSLSHardware(
			node,
			slsCommon.ClassRiver,
			thisNodeExtraProperties,
		)
	}

	return
}

func (g *StateGenerator) getSwitchConnectionForHardware(
	hardware slsCommon.GenericHardware, row shcdParser.HMNRow,
) (connection slsCommon.GenericHardware) {

	// Determine the xname of the device that this MgmtSwitchConnector will connect to
	var destinationXname string
	if xnametypes.IsHMSTypeController(hardware.TypeString) {
		// This this type *IS* the BMC or PDU, then don't use the parent, use the xname.
		destinationXname = hardware.Xname
	} else {
		destinationXname = hardware.Parent
	}

	//
	// Build up the xname for the MgmtSwitchConnector
	//

	// Determine the cabinet integer
	cabinetString := strings.TrimPrefix(
		strings.ToLower(row.DestinationRack),
		"x",
	)
	cabinetInteger, err := strconv.Atoi(cabinetString)
	if err != nil {
		g.logger.Fatal(
			"Failed to parse destination cabinet number string to integer!",
			zap.Error(err),
			zap.String(
				"cabinetString",
				cabinetString,
			),
			zap.Any(
				"row",
				row,
			),
		)
	}

	cabinet := xnames.Cabinet{Cabinet: cabinetInteger}

	// Determine the chassis integer
	chassis, err := g.determineRiverChassis(cabinet)
	if err != nil {
		g.logger.Fatal(
			"Failed to determine the chassis integer for node!",
			zap.Error(err),
			zap.String(
				"cabinet",
				cabinet.String(),
			),
			zap.Any(
				"row",
				row,
			),
		)
	}

	// Determine the slot/rack u
	destinationUString := strings.TrimPrefix(
		strings.ToLower(row.DestinationLocation),
		"u",
	)
	destinationUInteger, err := strconv.Atoi(destinationUString)
	if err != nil {
		g.logger.Fatal(
			"Failed to parse destination location number string to integer!",
			zap.Error(err),
			zap.String(
				"destinationUInteger",
				destinationUString,
			),
			zap.Any(
				"row",
				row,
			),
		)
	}

	mgmtSwitch := chassis.MgmtSwitch(destinationUInteger)

	// Because of "reasons" the port/jack string is either prefixed with a `j` or a `p`. To combat this, use regex.
	portSubmatches := portRegex.FindStringSubmatch(row.DestinationPort)
	if len(portSubmatches) < 2 {
		g.logger.Fatal(
			"Attempted to run regex on destination port but did not find port number!",
			zap.Any(
				"portSubmatches",
				portSubmatches,
			),
			zap.Any(
				"row",
				row,
			),
		)
	}
	destinationPortString := portSubmatches[1]

	destinationPortInteger, err := strconv.Atoi(destinationPortString)
	if err != nil {
		g.logger.Fatal(
			"Failed to parse destination port number string to integer!",
			zap.Error(err),
			zap.String(
				"destinationPortString",
				destinationPortString,
			),
			zap.Any(
				"row",
				row,
			),
		)
	}

	mgmtSwitchConnector := mgmtSwitch.MgmtSwitchConnector(destinationPortInteger)

	// Get the brand for this switch
	slsMgmtSwitch, ok := g.inputState.ManagementSwitches[mgmtSwitch.String()]
	if !ok {
		g.logger.Fatal(
			"Unable to find management switch",
			zap.String(
				"switchName",
				mgmtSwitch.String(),
			),
			zap.String(
				"connectorXname",
				mgmtSwitchConnector.String(),
			),
			zap.String(
				"destinationXname",
				destinationXname,
			),
		)
	}

	ep, ok := slsMgmtSwitch.ExtraPropertiesRaw.(slsCommon.ComptypeMgmtSwitch)
	if !ok {
		g.logger.Fatal(
			"Unable to get management switch extra properties",
			zap.String(
				"switchName",
				mgmtSwitch.String(),
			),
			zap.String(
				"connectorXname",
				mgmtSwitchConnector.String(),
			),
			zap.String(
				"destinationXname",
				destinationXname,
			),
		)
	}
	switchBrand := ep.Brand

	if switchBrand == "" {
		g.logger.Fatal(
			"Management Switch brand found not provided for switch",
			zap.String(
				"switchName",
				mgmtSwitch.String(),
			),
			zap.String(
				"connectorXname",
				mgmtSwitchConnector.String(),
			),
			zap.String(
				"destinationXname",
				destinationXname,
			),
		)
	}

	// Calculate the vendor name for the ethernet interfaces
	// Dell switches use: ethernet1/1/1
	// Aruba switches use: 1/1/1
	var vendorName string
	switch switchBrand {
	case networking.ManagementSwitchBrandDell.String():
		vendorName = fmt.Sprintf(
			"ethernet1/1/%d",
			destinationPortInteger,
		)
	case networking.ManagementSwitchBrandAruba.String():
		vendorName = fmt.Sprintf(
			"1/1/%d",
			destinationPortInteger,
		)
	case networking.ManagementSwitchBrandMellanox.String():
		// This should only occur when the HMN connections says that a BMC is connected to the
		// spine/leaf switch. Which should not happen.
		g.logger.Fatal(
			"Currently do no support MgmtSwitchConnector for Mellanox switches",
			zap.Any(
				"switchBrand",
				switchBrand,
			),
			zap.String(
				"switchName",
				mgmtSwitch.String(),
			),
			zap.String(
				"connectorXname",
				mgmtSwitchConnector.String(),
			),
			zap.String(
				"destinationXname",
				destinationXname,
			),
		)
	default:
		g.logger.Fatal(
			"Unknown Management Switch brand found for switch",
			zap.Any(
				"switchBrand",
				switchBrand,
			),
			zap.String(
				"switchName",
				mgmtSwitch.String(),
			),
			zap.String(
				"connectorXname",
				mgmtSwitchConnector.String(),
			),
			zap.String(
				"destinationXname",
				destinationXname,
			),
		)
	}

	connection = g.buildSLSHardware(
		mgmtSwitchConnector,
		slsCommon.ClassRiver,
		slsCommon.ComptypeMgmtSwitchConnector{
			NodeNics:   []string{destinationXname},
			VendorName: vendorName,
		},
	)

	return
}

func (g *StateGenerator) findRowWithSource(sourceParent string) shcdParser.HMNRow {
	sourceParentLowerCase := strings.ToLower(sourceParent)
	for _, row := range g.hmnRows {
		if strings.ToLower(row.Source) == sourceParentLowerCase {
			return row
		}
	}

	return shcdParser.HMNRow{}
}

// Mountain and Hill hardware
func (g *StateGenerator) getLiquidCooledHardwareForCabinet(cabinetTemplate CabinetTemplate) (hardware []slsCommon.GenericHardware) {
	// Start with the Cabinet
	cabinetXname := cabinetTemplate.Xname

	slsCabinet := g.buildSLSHardware(
		cabinetXname,
		cabinetTemplate.Class,
		cabinetTemplate.buildExtraProperties(),
	)
	hardware = append(
		hardware,
		slsCabinet,
	)

	for _, chassisOrdinal := range cabinetTemplate.LiquidCooledChassisList {
		chassisXname := cabinetXname.Chassis(chassisOrdinal)

		// Start with the Chassis
		slsChassis := g.buildSLSHardware(
			chassisXname,
			cabinetTemplate.Class,
			nil,
		)
		hardware = append(
			hardware,
			slsChassis,
		)

		// Next the CMM
		slsChassisBMC := g.buildSLSHardware(
			chassisXname.ChassisBMC(0),
			cabinetTemplate.Class,
			nil,
		)
		hardware = append(
			hardware,
			slsChassisBMC,
		)

		for slotOrdinal := 0; slotOrdinal < 8; slotOrdinal++ {
			for bmcOrdinal := 0; bmcOrdinal < 2; bmcOrdinal++ {
				for nodeOrdinal := 0; nodeOrdinal < 2; nodeOrdinal++ {
					// Construct the xname for the node
					nodeXname := chassisXname.ComputeModule(slotOrdinal).NodeBMC(bmcOrdinal).Node(nodeOrdinal)

					node := g.buildSLSHardware(
						nodeXname,
						cabinetTemplate.Class,
						slsCommon.ComptypeNode{
							NID:  g.currentMountainNID,
							Role: "Compute",
							Aliases: []string{
								fmt.Sprintf(
									"nid%06d",
									g.currentMountainNID,
								),
							},
						},
					)

					hardware = append(
						hardware,
						node,
					)

					g.currentMountainNID++
				}
			}
		}
	}

	return
}

func (g *StateGenerator) buildSLSHardware(
	xname xnames.Xname, class slsCommon.CabinetType, extraProperties interface{},
) slsCommon.GenericHardware {
	if xname == nil {
		panic("nil xname provided")
	}

	return slsCommon.GenericHardware{
		Parent:             xname.ParentInterface().String(),
		Xname:              xname.String(),
		Type:               slsCommon.HMSTypeToHMSStringType(xname.Type()),
		Class:              class,
		TypeString:         xname.Type(),
		ExtraPropertiesRaw: extraProperties,
	}
}

// Networks
func (g *StateGenerator) buildNetworksSection() (allNetworks map[string]slsCommon.Network) {
	allNetworks = g.inputState.Networks

	// This would be a good place to do any modifications to the given network data.
	// For right now, we leave them be.

	return
}

// GenerateSLSState generates new SLSState object from an input state and hmn-connections file.
func GenerateSLSState(inputState GeneratorInputState, hmnRows []shcdParser.HMNRow) slsCommon.SLSState {
	atomicLevel := zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			atomicLevel,
		),
	)

	atomicLevel.SetLevel(zap.InfoLevel)

	logger.Info("Beginning SLS configuration generation.")
	g := NewStateGenerator(
		logger,
		inputState,
		hmnRows,
	)
	return g.GenerateSLSState()
}
