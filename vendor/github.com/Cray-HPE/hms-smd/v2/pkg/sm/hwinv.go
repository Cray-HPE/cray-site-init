// MIT License
//
// (C) Copyright [2018-2023] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package sm

import (
	"encoding/json"
	"regexp"
	"strconv"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	rf "github.com/Cray-HPE/hms-smd/v2/pkg/redfish"
)

var ErrHWLocInvalid = base.NewHMSError("sm", "ID is empty or not a valid xname")
var ErrHWFRUIDInvalid = base.NewHMSError("sm", "FRUID is empty or invalid")
var ErrHWInvFmtInvalid = base.NewHMSError("sm", "Invalid HW Inventory format")
var ErrHWInvFmtNI = base.NewHMSError("sm", "HW Inv format not yet implemented")
var ErrHWInvMissingFRU = base.NewHMSError("sm", "PopulatedFRU must be populated")
var ErrHWInvMissingFRUInfo = base.NewHMSError("sm", "FRU info is empty")
var ErrHWInvMissingLoc = base.NewHMSError("sm", "Component location info is empty")

// Note most of these structures are polymorphic in the sense that they
// are stored generically in the database largely as raw json.  The non
// json parts of the struct are constant across different types.  That said,
// the client only cares about the final payload, so we give each a name
// specific to the type.  In the published spec, the json stored and
// retrieved will have to match the schema for the type in the array.
//
// Also, the hwinv location/FRU schemas the raw Redfish for the type, but
// after sorting the FRU (physical) properties and locatation (specific
// to current location) into separate embedded sub-structs (e.g. SystemFRUInfo
// and SystemLocationInfo) so we can easily snip them out separately.
// these two types of structs determine the actual schema that appears on
// the wire and gets stored as json in the DB.  The structures here should
// be largely static.  Other fields in the Redfish output that do not
// fall into one of these structures is not used for HWInventory, but is
// collected to assist in discovery or whatever other purpose.

// This is an embedded structure for HW inventory.  There should be one
// array for every hms type tracked in the inventory.  This structure
// is also reused to allow individual HWInvByLoc structures to represent
// child components for nested inventory structures.
type hmsTypeArrays struct {
	Nodes          *[]*HWInvByLoc `json:"Nodes,omitempty"`
	Cabinets       *[]*HWInvByLoc `json:"Cabinets,omitempty"`
	Chassis        *[]*HWInvByLoc `json:"Chassis,omitempty"`
	ComputeModules *[]*HWInvByLoc `json:"ComputeModules,omitempty"`
	RouterModules  *[]*HWInvByLoc `json:"RouterModules,omitempty"`
	NodeEnclosures *[]*HWInvByLoc `json:"NodeEnclosures,omitempty"`
	HSNBoards      *[]*HWInvByLoc `json:"HSNBoards,omitempty"`

	Processors *[]*HWInvByLoc `json:"Processors,omitempty"`
	Memory     *[]*HWInvByLoc `json:"Memory,omitempty"`
	Drives     *[]*HWInvByLoc `json:"Drives,omitempty"`

	CabinetPDUs                *[]*HWInvByLoc `json:"CabinetPDUs,omitempty"`
	CabinetPDUOutlets          *[]*HWInvByLoc `json:"CabinetPDUPowerConnectors,omitempty"`
	CMMRectifiers              *[]*HWInvByLoc `json:"CMMRectifiers,omitempty"`
	NodeAccels                 *[]*HWInvByLoc `json:"NodeAccels,omitempty"`
	NodeAccelRisers            *[]*HWInvByLoc `json:"NodeAccelRisers,omitempty"`
	NodeEnclosurePowerSupplies *[]*HWInvByLoc `json:"NodeEnclosurePowerSupplies,omitempty"`
	NodeHsnNICs                *[]*HWInvByLoc `json:"NodeHsnNics,omitempty"`

	// These don't have hardware inventory location/FRU info yet,
	// either because they aren't known yet or because they are manager
	// types.  Each manager (e.g. BMC) should have some kind of physical
	// enclosure, and for the purposes of HW inventory we might not need
	// both (but probably will).
	CECs           *[]*HWInvByLoc `json:"CECs,omitempty"`
	CDUs           *[]*HWInvByLoc `json:"CDUs,omitempty"`
	CabinetCDUs    *[]*HWInvByLoc `json:"CabinetCDUs,omitempty"`
	CMMFpgas       *[]*HWInvByLoc `json:"CMMFpgas,omitempty"`
	NodeFpgas      *[]*HWInvByLoc `json:"NodeFpgas,omitempty"`
	RouterFpgas    *[]*HWInvByLoc `json:"RouterFpgas,omitempty"`
	RouterTORFpgas *[]*HWInvByLoc `json:"RouterTORFpgas,omitempty"`
	HSNAsics       *[]*HWInvByLoc `json:"HSNAsics,omitempty"`

	CabinetBMCs           *[]*HWInvByLoc `json:"CabinetBMCs,omitempty"`
	CabinetPDUControllers *[]*HWInvByLoc `json:"CabinetPDUControllers,omitempty"`
	ChassisBMCs           *[]*HWInvByLoc `json:"ChassisBMCs,omitempty"`
	NodeBMCs              *[]*HWInvByLoc `json:"NodeBMCs,omitempty"`
	RouterBMCs            *[]*HWInvByLoc `json:"RouterBMCs,omitempty"`

	CabinetPDUNics      *[]*HWInvByLoc `json:"CabinetPDUNics,omitempty"`
	NodePowerConnectors *[]*HWInvByLoc `json:"NodePowerConnectors,omitempty"`
	NodeBMCNics         *[]*HWInvByLoc `json:"NodeBMCNics,omitempty"`
	NodeNICs            *[]*HWInvByLoc `json:"NodeNICs,omitempty"`
	RouterBMCNics       *[]*HWInvByLoc `json:"RouterBMCNics,omitempty"`

	MgmtSwitches    *[]*HWInvByLoc `json:"MgmtSwitches,omitempty"`
	MgmtHLSwitches  *[]*HWInvByLoc `json:"MgmtHLSwitches,omitempty"`
	CDUMgmtSwitches *[]*HWInvByLoc `json:"CDUMgmtSwitches,omitempty"`

	// Also not implemented yet.  Not clear if these will have any interesting
	// info, so they may never be,
	SMSBoxes             *[]*HWInvByLoc `json:"SMSBoxes,omitempty"`
	HSNLinks             *[]*HWInvByLoc `json:"HSNLinks,omitempty"`
	HSNConnectors        *[]*HWInvByLoc `json:"HSNConnectors,omitempty"`
	HSNConnectorPorts    *[]*HWInvByLoc `json:"HSNConnectorPorts,omitempty"`
	MgmtSwitchConnectors *[]*HWInvByLoc `json:"MgmtSwitchConnectors,omitempty"`
}

// This is a top-level hardware inventory.  We can do a flat mapping
// where every component tracked is its own top-level array, a
// completely hierarchical mapping (since the entry for a component
// can contain it's own set of hmsTypeArrays), or some combination
// (such as node subcomponents being nested, but not higher-level
// components).
type SystemHWInventory struct {
	XName  string
	Format string

	hmsTypeArrays
}

// Valid values for Format field above
const (
	HWInvFormatFullyFlat     = "FullyFlat"
	HWInvFormatHierarchical  = "Hierarchical"  // Not implemented yet.
	HWInvFormatNestNodesOnly = "NestNodesOnly" // Default
)

// Create formatted SystemHWInventory from a random array of HWInvByLoc entries.
// No sorting is done (with components of the same type), so pre/post-sort if
// needed.
// Note: entries in *HWInvByLoc are not copied if modified.  Child entries
// will be appended if format is not FullyFlat, but otherwise no changes will
// be made.
func NewSystemHWInventory(hwlocs []*HWInvByLoc, xName, format string) (*SystemHWInventory, error) {
	hwinv := new(SystemHWInventory)
	hwinv.XName = xName
	if format == HWInvFormatNestNodesOnly ||
		format == HWInvFormatHierarchical ||
		format == HWInvFormatFullyFlat {

		hwinv.Format = format
	} else {
		return nil, ErrHWInvFmtInvalid
	}
	var err error
	for _, hwloc := range hwlocs {
		switch xnametypes.ToHMSType(hwloc.Type) {
		// HWInv based on Redfish "Chassis" Type.
		case xnametypes.Cabinet:
			if hwinv.Cabinets == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Cabinets = &arr
			}
			*hwinv.Cabinets = append(*hwinv.Cabinets, hwloc)
		case xnametypes.Chassis:
			if hwinv.Chassis == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Chassis = &arr
			}
			*hwinv.Chassis = append(*hwinv.Chassis, hwloc)
		case xnametypes.ComputeModule:
			if hwinv.ComputeModules == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.ComputeModules = &arr
			}
			*hwinv.ComputeModules = append(*hwinv.ComputeModules, hwloc)
		case xnametypes.RouterModule:
			if hwinv.RouterModules == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.RouterModules = &arr
			}
			*hwinv.RouterModules = append(*hwinv.RouterModules, hwloc)
		case xnametypes.NodeEnclosure:
			if hwinv.NodeEnclosures == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.NodeEnclosures = &arr
			}
			*hwinv.NodeEnclosures = append(*hwinv.NodeEnclosures, hwloc)
		case xnametypes.HSNBoard:
			if hwinv.HSNBoards == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.HSNBoards = &arr
			}
			*hwinv.HSNBoards = append(*hwinv.HSNBoards, hwloc)
		case xnametypes.MgmtSwitch:
			if hwinv.MgmtSwitches == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.MgmtSwitches = &arr
			}
			*hwinv.MgmtSwitches = append(*hwinv.MgmtSwitches, hwloc)
		case xnametypes.MgmtHLSwitch:
			if hwinv.MgmtHLSwitches == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.MgmtHLSwitches = &arr
			}
			*hwinv.MgmtHLSwitches = append(*hwinv.MgmtHLSwitches, hwloc)
		case xnametypes.CDUMgmtSwitch:
			if hwinv.CDUMgmtSwitches == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.CDUMgmtSwitches = &arr
			}
			*hwinv.CDUMgmtSwitches = append(*hwinv.CDUMgmtSwitches, hwloc)
		case xnametypes.Node:
			if hwinv.Nodes == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Nodes = &arr
			}
			*hwinv.Nodes = append(*hwinv.Nodes, hwloc)
		case xnametypes.NodeAccel:
			if hwinv.NodeAccels == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.NodeAccels = &arr
			}
			*hwinv.NodeAccels = append(*hwinv.NodeAccels, hwloc)
		case xnametypes.Processor:
			if hwinv.Processors == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Processors = &arr
			}
			*hwinv.Processors = append(*hwinv.Processors, hwloc)
		case xnametypes.Memory:
			if hwinv.Memory == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Memory = &arr
			}
			*hwinv.Memory = append(*hwinv.Memory, hwloc)
		case xnametypes.Drive:
			if hwinv.Drives == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.Drives = &arr
			}
			*hwinv.Drives = append(*hwinv.Drives, hwloc)
		case xnametypes.NodeHsnNic:
			if hwinv.NodeHsnNICs == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.NodeHsnNICs = &arr
			}
			*hwinv.NodeHsnNICs = append(*hwinv.NodeHsnNICs, hwloc)
		case xnametypes.CabinetPDU:
			if hwinv.CabinetPDUs == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.CabinetPDUs = &arr
			}
			*hwinv.CabinetPDUs = append(*hwinv.CabinetPDUs, hwloc)
		case xnametypes.CabinetPDUOutlet:
			fallthrough
		case xnametypes.CabinetPDUPowerConnector:
			if hwinv.CabinetPDUOutlets == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.CabinetPDUOutlets = &arr
			}
			*hwinv.CabinetPDUOutlets = append(*hwinv.CabinetPDUOutlets, hwloc)
		case xnametypes.CMMRectifier:
			if hwinv.CMMRectifiers == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.CMMRectifiers = &arr
			}
			*hwinv.CMMRectifiers = append(*hwinv.CMMRectifiers, hwloc)
		case xnametypes.NodeEnclosurePowerSupply:
			if hwinv.NodeEnclosurePowerSupplies == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.NodeEnclosurePowerSupplies = &arr
			}
			*hwinv.NodeEnclosurePowerSupplies = append(*hwinv.NodeEnclosurePowerSupplies, hwloc)
		case xnametypes.NodeAccelRiser:
			if hwinv.NodeAccelRisers == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.NodeAccelRisers = &arr
			}
			*hwinv.NodeAccelRisers = append(*hwinv.NodeAccelRisers, hwloc)
		case xnametypes.NodeBMC:
			if hwinv.NodeBMCs == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.NodeBMCs = &arr
			}
			*hwinv.NodeBMCs = append(*hwinv.NodeBMCs, hwloc)
		case xnametypes.RouterBMC:
			if hwinv.RouterBMCs == nil {
				arr := make([]*HWInvByLoc, 0, 1)
				hwinv.RouterBMCs = &arr
			}
			*hwinv.RouterBMCs = append(*hwinv.RouterBMCs, hwloc)
		case xnametypes.HMSTypeInvalid:
			err = base.ErrHMSTypeInvalid
		// Not supported for this type.
		default:
			err = base.ErrHMSTypeUnsupported
		}
	}
	// If not completely "FullyFlat", start rolling up subcomponent
	// arrays into their parent components and then dropping them.
	if hwinv.Format == HWInvFormatNestNodesOnly ||
		hwinv.Format == HWInvFormatHierarchical {

		//
		// Nodes
		//

		// Roll up Node subcomponents into their parent nodes.
		// For "NestNodesOnly" this is the only roll-up step needed.

		// Avoid n^2 by first creating map to look up parent nodes in
		// constant-ish time.
		nmap := make(map[string]*HWInvByLoc)
		if hwinv.Nodes != nil {
			for _, n := range *hwinv.Nodes {
				nmap[n.ID] = n
			}
		}

		procArray := hwinv.Processors
		nodeAccelArray := hwinv.NodeAccels
		memArray := hwinv.Memory
		driveArray := hwinv.Drives
		hsnNicArray := hwinv.NodeHsnNICs
		nodeAccelRiserArray := hwinv.NodeAccelRisers
		// Moving these contents to underneath items in node array.
		// Set these arrays to nil so we won't list them twice.
		hwinv.Processors = nil
		hwinv.NodeAccels = nil
		hwinv.Memory = nil
		hwinv.Drives = nil
		hwinv.NodeHsnNICs = nil
		hwinv.NodeAccelRisers = nil

		// Processors are children of Node
		if procArray != nil {
			for _, p := range *procArray {
				parentID := xnametypes.GetHMSCompParent(p.ID)
				parent, ok := nmap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find node key %s for %s",
						parentID, p.ID)
					if hwinv.Processors == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.Processors = &arr
					}
					// Put orphan components back in their array
					*hwinv.Processors = append(*hwinv.Processors, p)
				} else {
					if parent.Processors == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.Processors = &arr
					}
					*parent.Processors = append(*parent.Processors, p)
				}
			}
		}
		// NodeAccels (GPUs) are children of Node
		if nodeAccelArray != nil {
			for _, na := range *nodeAccelArray {
				parentID := xnametypes.GetHMSCompParent(na.ID)
				parent, ok := nmap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find node key %s for %s",
						parentID, na.ID)
					if hwinv.NodeAccels == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.NodeAccels = &arr
					}
					// Put orphan components back in their array
					*hwinv.NodeAccels = append(*hwinv.NodeAccels, na)
				} else {
					if parent.NodeAccels == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.NodeAccels = &arr
					}
					*parent.NodeAccels = append(*parent.NodeAccels, na)
				}
			}
		}
		// Memory modules are children of Node
		if memArray != nil {
			for _, m := range *memArray {
				parentID := xnametypes.GetHMSCompParent(m.ID)
				parent, ok := nmap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find node key %s for %s",
						parentID, m.ID)
					if hwinv.Memory == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.Memory = &arr
					}
					// Put orphan components back in their array
					*hwinv.Memory = append(*hwinv.Memory, m)
				} else {
					if parent.Memory == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.Memory = &arr
					}
					*parent.Memory = append(*parent.Memory, m)
				}
			}
		}

		// Drives are children of Node
		if driveArray != nil {
			for _, d := range *driveArray {
				parentID := d.ID
				for xnametypes.GetHMSType(parentID) != xnametypes.Node {
					parentID = xnametypes.GetHMSCompParent(parentID)
				}
				parent, ok := nmap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find node key %s for %s",
						parentID, d.ID)
					if hwinv.Drives == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.Drives = &arr
					}
					// Put orphan components back in their array
					*hwinv.Drives = append(*hwinv.Drives, d)
				} else {
					if parent.Drives == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.Drives = &arr
					}
					*parent.Drives = append(*parent.Drives, d)
				}
			}
		}

		// HSN NICs are children of Nodes
		if hsnNicArray != nil {
			for _, n := range *hsnNicArray {
				parentID := xnametypes.GetHMSCompParent(n.ID)
				parent, ok := nmap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find node key %s for %s",
						parentID, n.ID)
					if hwinv.NodeHsnNICs == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.NodeHsnNICs = &arr
					}
					// Put orphan components back in their array
					*hwinv.NodeHsnNICs = append(*hwinv.NodeHsnNICs, n)
				} else {
					if parent.NodeHsnNICs == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.NodeHsnNICs = &arr
					}
					*parent.NodeHsnNICs = append(*parent.NodeHsnNICs, n)
				}
			}
		}

		// NodeAccelRisers are children of Nodes
		if nodeAccelRiserArray != nil {
			for _, n := range *nodeAccelRiserArray {
				parentID := xnametypes.GetHMSCompParent(n.ID)
				parent, ok := nmap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find node key %s for %s",
						parentID, n.ID)
					if hwinv.NodeAccelRisers == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.NodeAccelRisers = &arr
					}
					// Put orphan components back in their array
					*hwinv.NodeAccelRisers = append(*hwinv.NodeAccelRisers, n)
				} else {
					if parent.NodeAccelRisers == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.NodeAccelRisers = &arr
					}
					*parent.NodeAccelRisers = append(*parent.NodeAccelRisers, n)
				}
			}
		}

		//
		// PDUs, nest outlets
		//

		// Avoid n^2 by first creating map to look up parent PDU in
		// constant-ish time.
		pdumap := make(map[string]*HWInvByLoc)
		if hwinv.CabinetPDUs != nil {
			for _, pdu := range *hwinv.CabinetPDUs {
				pdumap[pdu.ID] = pdu
			}
		}

		cabPDUArray := hwinv.CabinetPDUOutlets
		// Moving these contents to underneath items in CabinetPDU array.
		// Set this array to nil so we won't list them twice.
		hwinv.CabinetPDUOutlets = nil

		// CabinetPDUOutlets are children of CabinetPDUs
		if cabPDUArray != nil {
			for _, out := range *cabPDUArray {
				parentID := xnametypes.GetHMSCompParent(out.ID)
				parent, ok := pdumap[parentID]
				if !ok {
					errlog.Printf("ERROR: Could not find pdu key %s for %s",
						parentID, out.ID)
					if hwinv.CabinetPDUOutlets == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						hwinv.CabinetPDUOutlets = &arr
					}
					// Put orphan components back in their array
					*hwinv.CabinetPDUOutlets = append(*hwinv.CabinetPDUOutlets, out)
				} else {
					if parent.CabinetPDUOutlets == nil {
						arr := make([]*HWInvByLoc, 0, 1)
						parent.CabinetPDUOutlets = &arr
					}
					*parent.CabinetPDUOutlets = append(*parent.CabinetPDUOutlets, out)
				}
			}
		}
		// Moved these contents to underneath items in CabinetPDU array.
		// Set these arrays to nil so we won't list them twice.
		hwinv.CabinetPDUOutlets = nil
	}
	if hwinv.Format == HWInvFormatHierarchical {
		// Continue rolling up components
		// TODO - need to implement controllers first.
		return hwinv, ErrHWInvFmtNI
	}
	return hwinv, err
}

// Fills out and verifies HW Inventory entries coming from external sources
func NewHWInvByLocs(hwlocs []HWInvByLoc) ([]*HWInvByLoc, error) {
	var err error
	var hls []*HWInvByLoc
	re := regexp.MustCompile(`[0-9]+$`)

	for _, hwloc := range hwlocs {
		hwloc.ID = xnametypes.NormalizeHMSCompID(hwloc.ID)
		hmsType := xnametypes.GetHMSType(hwloc.ID)
		if hmsType == xnametypes.HMSTypeInvalid {
			return hls, ErrHWLocInvalid //TODO: Define error
		}
		hwloc.Type = hmsType.String()
		ordinalStr := re.FindString(hwloc.ID)
		hwloc.Ordinal, _ = strconv.Atoi(ordinalStr)
		hwloc.Status = "Populated"
		if hwloc.PopulatedFRU == nil {
			return hls, ErrHWInvMissingFRU
		}
		hwloc.PopulatedFRU.Type = hwloc.Type
		switch hmsType {
		case xnametypes.Cabinet:
			if hwloc.HMSCabinetLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSCabinetFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocCabinet
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUCabinet
			c := new(rf.EpChassis)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ChassisRF.Manufacturer = hwloc.PopulatedFRU.HMSCabinetFRUInfo.Manufacturer
			c.ChassisRF.PartNumber = hwloc.PopulatedFRU.HMSCabinetFRUInfo.PartNumber
			c.ChassisRF.SerialNumber = hwloc.PopulatedFRU.HMSCabinetFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetChassisFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.Chassis:
			if hwloc.HMSChassisLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSChassisFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocChassis
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUChassis
			c := new(rf.EpChassis)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ChassisRF.Manufacturer = hwloc.PopulatedFRU.HMSChassisFRUInfo.Manufacturer
			c.ChassisRF.PartNumber = hwloc.PopulatedFRU.HMSChassisFRUInfo.PartNumber
			c.ChassisRF.SerialNumber = hwloc.PopulatedFRU.HMSChassisFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetChassisFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.ComputeModule:
			if hwloc.HMSComputeModuleLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSComputeModuleFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocComputeModule
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUComputeModule
			c := new(rf.EpChassis)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ChassisRF.Manufacturer = hwloc.PopulatedFRU.HMSComputeModuleFRUInfo.Manufacturer
			c.ChassisRF.PartNumber = hwloc.PopulatedFRU.HMSComputeModuleFRUInfo.PartNumber
			c.ChassisRF.SerialNumber = hwloc.PopulatedFRU.HMSComputeModuleFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetChassisFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.RouterModule:
			if hwloc.HMSRouterModuleLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSRouterModuleFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocRouterModule
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRURouterModule
			c := new(rf.EpChassis)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ChassisRF.Manufacturer = hwloc.PopulatedFRU.HMSRouterModuleFRUInfo.Manufacturer
			c.ChassisRF.PartNumber = hwloc.PopulatedFRU.HMSRouterModuleFRUInfo.PartNumber
			c.ChassisRF.SerialNumber = hwloc.PopulatedFRU.HMSRouterModuleFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetChassisFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.NodeEnclosure:
			if hwloc.HMSNodeEnclosureLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSNodeEnclosureFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocNodeEnclosure
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUNodeEnclosure
			c := new(rf.EpChassis)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ChassisRF.Manufacturer = hwloc.PopulatedFRU.HMSNodeEnclosureFRUInfo.Manufacturer
			c.ChassisRF.PartNumber = hwloc.PopulatedFRU.HMSNodeEnclosureFRUInfo.PartNumber
			c.ChassisRF.SerialNumber = hwloc.PopulatedFRU.HMSNodeEnclosureFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetChassisFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.HSNBoard:
			if hwloc.HMSHSNBoardLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSHSNBoardFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocHSNBoard
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUHSNBoard
			c := new(rf.EpChassis)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ChassisRF.Manufacturer = hwloc.PopulatedFRU.HMSHSNBoardFRUInfo.Manufacturer
			c.ChassisRF.PartNumber = hwloc.PopulatedFRU.HMSHSNBoardFRUInfo.PartNumber
			c.ChassisRF.SerialNumber = hwloc.PopulatedFRU.HMSHSNBoardFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetChassisFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.MgmtSwitch:
			if hwloc.HMSMgmtSwitchLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSMgmtSwitchFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocMgmtSwitch
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUMgmtSwitch
			c := new(rf.EpChassis)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ChassisRF.Manufacturer = hwloc.PopulatedFRU.HMSMgmtSwitchFRUInfo.Manufacturer
			c.ChassisRF.PartNumber = hwloc.PopulatedFRU.HMSMgmtSwitchFRUInfo.PartNumber
			c.ChassisRF.SerialNumber = hwloc.PopulatedFRU.HMSMgmtSwitchFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetChassisFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.MgmtHLSwitch:
			if hwloc.HMSMgmtHLSwitchLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSMgmtHLSwitchFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocMgmtHLSwitch
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUMgmtHLSwitch
			c := new(rf.EpChassis)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ChassisRF.Manufacturer = hwloc.PopulatedFRU.HMSMgmtHLSwitchFRUInfo.Manufacturer
			c.ChassisRF.PartNumber = hwloc.PopulatedFRU.HMSMgmtHLSwitchFRUInfo.PartNumber
			c.ChassisRF.SerialNumber = hwloc.PopulatedFRU.HMSMgmtHLSwitchFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetChassisFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.CDUMgmtSwitch:
			if hwloc.HMSCDUMgmtSwitchLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSCDUMgmtSwitchFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocCDUMgmtSwitch
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUCDUMgmtSwitch
			c := new(rf.EpChassis)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ChassisRF.Manufacturer = hwloc.PopulatedFRU.HMSCDUMgmtSwitchFRUInfo.Manufacturer
			c.ChassisRF.PartNumber = hwloc.PopulatedFRU.HMSCDUMgmtSwitchFRUInfo.PartNumber
			c.ChassisRF.SerialNumber = hwloc.PopulatedFRU.HMSCDUMgmtSwitchFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetChassisFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.Node:
			if hwloc.HMSNodeLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSNodeFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocNode
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUNode
			c := new(rf.EpSystem)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.SystemRF.Manufacturer = hwloc.PopulatedFRU.HMSNodeFRUInfo.Manufacturer
			c.SystemRF.PartNumber = hwloc.PopulatedFRU.HMSNodeFRUInfo.PartNumber
			c.SystemRF.SerialNumber = hwloc.PopulatedFRU.HMSNodeFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetSystemFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.NodeAccel:
			if hwloc.HMSNodeAccelLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSNodeAccelFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocNodeAccel
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUNodeAccel
			c := new(rf.EpProcessor)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ProcessorRF.Manufacturer = hwloc.PopulatedFRU.HMSNodeAccelFRUInfo.Manufacturer
			c.ProcessorRF.PartNumber = hwloc.PopulatedFRU.HMSNodeAccelFRUInfo.PartNumber
			c.ProcessorRF.SerialNumber = hwloc.PopulatedFRU.HMSNodeAccelFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetProcessorFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.Processor:
			if hwloc.HMSProcessorLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSProcessorFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocProcessor
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUProcessor
			c := new(rf.EpProcessor)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ProcessorRF.Manufacturer = hwloc.PopulatedFRU.HMSProcessorFRUInfo.Manufacturer
			c.ProcessorRF.PartNumber = hwloc.PopulatedFRU.HMSProcessorFRUInfo.PartNumber
			c.ProcessorRF.SerialNumber = hwloc.PopulatedFRU.HMSProcessorFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetProcessorFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.Memory:
			if hwloc.HMSMemoryLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSMemoryFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocMemory
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUMemory
			c := new(rf.EpMemory)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.MemoryRF.Manufacturer = hwloc.PopulatedFRU.HMSMemoryFRUInfo.Manufacturer
			c.MemoryRF.PartNumber = hwloc.PopulatedFRU.HMSMemoryFRUInfo.PartNumber
			c.MemoryRF.SerialNumber = hwloc.PopulatedFRU.HMSMemoryFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetMemoryFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.Drive:
			if hwloc.HMSDriveLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSDriveFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocDrive
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUDrive
			c := new(rf.EpDrive)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.DriveRF.Manufacturer = hwloc.PopulatedFRU.HMSDriveFRUInfo.Manufacturer
			c.DriveRF.PartNumber = hwloc.PopulatedFRU.HMSDriveFRUInfo.PartNumber
			c.DriveRF.SerialNumber = hwloc.PopulatedFRU.HMSDriveFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetDriveFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.NodeHsnNic:
			if hwloc.HMSHSNNICLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSHSNNICFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocHSNNIC
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUHSNNIC
			hwloc.PopulatedFRU.FRUID = rf.GetHSNNICFRUID(hwloc.Type, hwloc.ID, hwloc.PopulatedFRU.HMSHSNNICFRUInfo.Manufacturer, hwloc.PopulatedFRU.HMSHSNNICFRUInfo.PartNumber, hwloc.PopulatedFRU.HMSHSNNICFRUInfo.SerialNumber)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.CabinetPDU:
			if hwloc.HMSPDULocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSPDUFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocPDU
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUPDU
			c := new(rf.EpPDU)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.PowerDistributionRF.Manufacturer = hwloc.PopulatedFRU.HMSPDUFRUInfo.Manufacturer
			c.PowerDistributionRF.PartNumber = hwloc.PopulatedFRU.HMSPDUFRUInfo.PartNumber
			c.PowerDistributionRF.SerialNumber = hwloc.PopulatedFRU.HMSPDUFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetPDUFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.CabinetPDUOutlet:
			fallthrough
		case xnametypes.CabinetPDUPowerConnector:
			if hwloc.HMSOutletLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSOutletFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocOutlet
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUOutlet
			// Outlets need to be given their FRUID from info from the PDU
			// since their struct doesn't include the proper FRU info.
			if hwloc.PopulatedFRU.FRUID == "" {
				hwloc.PopulatedFRU.FRUID = "FRUIDfor" + hwloc.ID
			}
		case xnametypes.CMMRectifier:
			if hwloc.HMSCMMRectifierLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSCMMRectifierFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocCMMRectifier
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUCMMRectifier
			c := new(rf.EpPowerSupply)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.PowerSupplyRF = &rf.PowerSupply{
				PowerSupplyFRUInfoRF: rf.PowerSupplyFRUInfoRF{
					Manufacturer: hwloc.PopulatedFRU.HMSCMMRectifierFRUInfo.Manufacturer,
					SerialNumber: hwloc.PopulatedFRU.HMSCMMRectifierFRUInfo.SerialNumber,
				},
			}
			hwloc.PopulatedFRU.FRUID, err = rf.GetPowerSupplyFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.NodeEnclosurePowerSupply:
			if hwloc.HMSNodeEnclosurePowerSupplyLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSNodeEnclosurePowerSupplyFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocNodeEnclosurePowerSupply
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUNodeEnclosurePowerSupply
			c := new(rf.EpPowerSupply)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.PowerSupplyRF = &rf.PowerSupply{
				PowerSupplyFRUInfoRF: rf.PowerSupplyFRUInfoRF{
					Manufacturer: hwloc.PopulatedFRU.HMSNodeEnclosurePowerSupplyFRUInfo.Manufacturer,
					SerialNumber: hwloc.PopulatedFRU.HMSNodeEnclosurePowerSupplyFRUInfo.SerialNumber,
				},
			}
			hwloc.PopulatedFRU.FRUID, err = rf.GetPowerSupplyFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.NodeAccelRiser:
			if hwloc.HMSNodeAccelRiserLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSNodeAccelRiserFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocNodeAccelRiser
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUNodeAccelRiser
			c := new(rf.EpNodeAccelRiser) // NodeAccelRiserRF is a pointer
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.NodeAccelRiserRF = &rf.NodeAccelRiser{
				NodeAccelRiserFRUInfoRF: rf.NodeAccelRiserFRUInfoRF{
					Producer:     hwloc.PopulatedFRU.HMSNodeAccelRiserFRUInfo.Producer,
					SerialNumber: hwloc.PopulatedFRU.HMSNodeAccelRiserFRUInfo.SerialNumber,
					PartNumber:   hwloc.PopulatedFRU.HMSNodeAccelRiserFRUInfo.PartNumber,
				},
			}
			hwloc.PopulatedFRU.FRUID, err = rf.GetNodeAccelRiserFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.NodeBMC:
			if hwloc.HMSNodeBMCLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSNodeBMCFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocNodeBMC
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRUNodeBMC
			c := new(rf.EpManager)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ManagerRF.Manufacturer = hwloc.PopulatedFRU.HMSNodeBMCFRUInfo.Manufacturer
			c.ManagerRF.PartNumber = hwloc.PopulatedFRU.HMSNodeBMCFRUInfo.PartNumber
			c.ManagerRF.SerialNumber = hwloc.PopulatedFRU.HMSNodeBMCFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetManagerFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.RouterBMC:
			if hwloc.HMSRouterBMCLocationInfo == nil {
				return hls, ErrHWInvMissingLoc
			}
			if hwloc.PopulatedFRU.HMSRouterBMCFRUInfo == nil {
				return hls, ErrHWInvMissingFRUInfo
			}
			hwloc.HWInventoryByLocationType = HWInvByLocRouterBMC
			hwloc.PopulatedFRU.HWInventoryByFRUType = HWInvByFRURouterBMC
			c := new(rf.EpManager)
			c.Type = hwloc.Type
			c.ID = hwloc.ID
			c.ManagerRF.Manufacturer = hwloc.PopulatedFRU.HMSRouterBMCFRUInfo.Manufacturer
			c.ManagerRF.PartNumber = hwloc.PopulatedFRU.HMSRouterBMCFRUInfo.PartNumber
			c.ManagerRF.SerialNumber = hwloc.PopulatedFRU.HMSRouterBMCFRUInfo.SerialNumber
			hwloc.PopulatedFRU.FRUID, err = rf.GetManagerFRUID(c)
			if err != nil {
				errlog.Printf("FRUID Error: %s\n", err.Error())
				errlog.Printf("Using untrackable FRUID: %s\n", hwloc.PopulatedFRU.FRUID)
			}
		case xnametypes.HMSTypeInvalid:
			return hls, base.ErrHMSTypeInvalid
		// Not supported for this type.
		default:
			return hls, base.ErrHMSTypeUnsupported
		}
		hls = append(hls, &hwloc)
	}
	return hls, nil
}

////////////////////////////////////////////////////////////////////////////
//
// HW Inventory-by-location
//
// This is an individual component in the hardware inventory.  Or more
// accurately a location where the component is, linked to a separate
// object that describes the durable properties of the actual piece of
// physical hardware.  The latter can be tracked independently of
// its current location.
//
////////////////////////////////////////////////////////////////////////////

type HWInvByLoc struct {
	ID      string `json:"ID"`
	Type    string `json:"Type"`
	Ordinal int    `json:"Ordinal"`
	Status  string `json:"Status"`

	// This is used as a descriminator to determine the type of *Info
	// struct that will be included below.
	HWInventoryByLocationType string `json:"HWInventoryByLocationType"`

	// One of:var ErrHMSXnameInvalid = errors.New("got HMSTypeInvalid instead of valid type")
	//    HMSType                  Underlying RF Type          How named in json object
	HMSCabinetLocationInfo       *rf.ChassisLocationInfoRF   `json:"CabinetLocationInfo,omitempty"`
	HMSChassisLocationInfo       *rf.ChassisLocationInfoRF   `json:"ChassisLocationInfo,omitempty"` // Mountain chassis
	HMSComputeModuleLocationInfo *rf.ChassisLocationInfoRF   `json:"ComputeModuleLocationInfo,omitempty"`
	HMSRouterModuleLocationInfo  *rf.ChassisLocationInfoRF   `json:"RouterModuleLocationInfo,omitempty"`
	HMSNodeEnclosureLocationInfo *rf.ChassisLocationInfoRF   `json:"NodeEnclosureLocationInfo,omitempty"`
	HMSHSNBoardLocationInfo      *rf.ChassisLocationInfoRF   `json:"HSNBoardLocationInfo,omitempty"`
	HMSMgmtSwitchLocationInfo    *rf.ChassisLocationInfoRF   `json:"MgmtSwitchLocationInfo,omitempty"`
	HMSMgmtHLSwitchLocationInfo  *rf.ChassisLocationInfoRF   `json:"MgmtHLSwitchLocationInfo,omitempty"`
	HMSCDUMgmtSwitchLocationInfo *rf.ChassisLocationInfoRF   `json:"CDUMgmtSwitchLocationInfo,omitempty"`
	HMSNodeLocationInfo          *rf.SystemLocationInfoRF    `json:"NodeLocationInfo,omitempty"`
	HMSProcessorLocationInfo     *rf.ProcessorLocationInfoRF `json:"ProcessorLocationInfo,omitempty"`
	HMSNodeAccelLocationInfo     *rf.ProcessorLocationInfoRF `json:"NodeAccelLocationInfo,omitempty"`
	HMSMemoryLocationInfo        *rf.MemoryLocationInfoRF    `json:"MemoryLocationInfo,omitempty"`
	HMSDriveLocationInfo         *rf.DriveLocationInfoRF     `json:"DriveLocationInfo,omitempty"`
	HMSHSNNICLocationInfo        *rf.NALocationInfoRF        `json:"NodeHsnNicLocationInfo,omitempty"`

	HMSPDULocationInfo                      *rf.PowerDistributionLocationInfo `json:"PDULocationInfo,omitempty"`
	HMSOutletLocationInfo                   *rf.OutletLocationInfo            `json:"OutletLocationInfo,omitempty"`
	HMSCMMRectifierLocationInfo             *rf.PowerSupplyLocationInfoRF     `json:"CMMRectifierLocationInfo,omitempty"`
	HMSNodeEnclosurePowerSupplyLocationInfo *rf.PowerSupplyLocationInfoRF     `json:"NodeEnclosurePowerSupplyLocationInfo,omitempty"`
	HMSNodeBMCLocationInfo                  *rf.ManagerLocationInfoRF         `json:"NodeBMCLocationInfo,omitempty"`
	HMSRouterBMCLocationInfo                *rf.ManagerLocationInfoRF         `json:"RouterBMCLocationInfo,omitempty"`
	HMSNodeAccelRiserLocationInfo           *rf.NodeAccelRiserLocationInfoRF  `json:"NodeAccelRiserLocationInfo,omitempty"`
	// TODO: Remaining types in hmsTypeArrays

	// If status != empty, up to one of following, matching above *Info.
	PopulatedFRU *HWInvByFRU `json:"PopulatedFRU,omitempty"`

	// These are for nested references for subcomponents.
	hmsTypeArrays
}

// HWInventoryByLocationType
// TODO: Remaining types
const (
	HWInvByLocCabinet                  string = "HWInvByLocCabinet"
	HWInvByLocChassis                  string = "HWInvByLocChassis"
	HWInvByLocComputeModule            string = "HWInvByLocComputeModule"
	HWInvByLocRouterModule             string = "HWInvByLocRouterModule"
	HWInvByLocNodeEnclosure            string = "HWInvByLocNodeEnclosure"
	HWInvByLocHSNBoard                 string = "HWInvByLocHSNBoard"
	HWInvByLocMgmtSwitch               string = "HWInvByLocMgmtSwitch"
	HWInvByLocMgmtHLSwitch             string = "HWInvByLocMgmtHLSwitch"
	HWInvByLocCDUMgmtSwitch            string = "HWInvByLocCDUMgmtSwitch"
	HWInvByLocNode                     string = "HWInvByLocNode"
	HWInvByLocProcessor                string = "HWInvByLocProcessor"
	HWInvByLocNodeAccel                string = "HWInvByLocNodeAccel"
	HWInvByLocDrive                    string = "HWInvByLocDrive"
	HWInvByLocMemory                   string = "HWInvByLocMemory"
	HWInvByLocHSNNIC                   string = "HWInvByLocNodeHsnNic"
	HWInvByLocPDU                      string = "HWInvByLocPDU"
	HWInvByLocOutlet                   string = "HWInvByLocOutlet"
	HWInvByLocCMMRectifier             string = "HWInvByLocCMMRectifier"
	HWInvByLocNodeEnclosurePowerSupply string = "HWInvByLocNodeEnclosurePowerSupply"
	HWInvByLocNodeBMC                  string = "HWInvByLocNodeBMC"
	HWInvByLocRouterBMC                string = "HWInvByLocRouterBMC"
	HWInvByLocNodeAccelRiser           string = "HWInvByLocNodeAccelRiser"
)

////////////////////////////////////////////////////////////////////////////
// Encoding/decoding: HW Inventory location info
////////////////////////////////////////////////////////////////////////////

// This routine takes raw location info captured as free-form JSON (e.g.
// from a schema-free database field) and unmarshals it into the correct struct
// for the type with the proper type-specific name.
//
// NOTEs: The location info should be that produced by EncodeLocationInfo.
//        MODIFIES caller.
//
// Return: If err != nil hw is unmodified,
//         Else, the type's *LocationInfo pointer is set to the expected struct.
func (hw *HWInvByLoc) DecodeLocationInfo(locInfoJSON []byte) error {
	var (
		err                                    error
		rfChassisLocationInfo                  *rf.ChassisLocationInfoRF
		rfSystemLocationInfo                   *rf.SystemLocationInfoRF
		rfProcessorLocationInfo                *rf.ProcessorLocationInfoRF
		rfNodeAccelLocationInfo                *rf.ProcessorLocationInfoRF
		rfDriveLocationInfo                    *rf.DriveLocationInfoRF
		rfMemoryLocationInfo                   *rf.MemoryLocationInfoRF
		rfHSNNICLocationInfo                   *rf.NALocationInfoRF
		rfPDULocationInfo                      *rf.PowerDistributionLocationInfo
		rfOutletLocationInfo                   *rf.OutletLocationInfo
		rfCMMRectifierLocationInfo             *rf.PowerSupplyLocationInfoRF
		rfNodeEnclosurePowerSupplyLocationInfo *rf.PowerSupplyLocationInfoRF
		rfNodeBMCLocationInfo                  *rf.ManagerLocationInfoRF
		rfRouterBMCLocationInfo                *rf.ManagerLocationInfoRF
		rfNodeAccelRiserLocationInfo           *rf.NodeAccelRiserLocationInfoRF
	)

	switch xnametypes.ToHMSType(hw.Type) {
	// HWInv based on Redfish "Chassis" Type.  Identical structs (for now).
	case xnametypes.Cabinet:
		fallthrough
	case xnametypes.Chassis:
		fallthrough
	case xnametypes.ComputeModule:
		fallthrough
	case xnametypes.RouterModule:
		fallthrough
	case xnametypes.NodeEnclosure:
		fallthrough
	case xnametypes.HSNBoard:
		fallthrough
	case xnametypes.MgmtSwitch:
		fallthrough
	case xnametypes.MgmtHLSwitch:
		fallthrough
	case xnametypes.CDUMgmtSwitch:
		rfChassisLocationInfo = new(rf.ChassisLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfChassisLocationInfo)
		if err == nil {
			// Assign struct to appropriate name for type.
			switch xnametypes.ToHMSType(hw.Type) {
			case xnametypes.Cabinet:
				hw.HMSCabinetLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocCabinet
			case xnametypes.Chassis:
				hw.HMSChassisLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocChassis
			case xnametypes.ComputeModule:
				hw.HMSComputeModuleLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocComputeModule
			case xnametypes.RouterModule:
				hw.HMSRouterModuleLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocRouterModule
			case xnametypes.NodeEnclosure:
				hw.HMSNodeEnclosureLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocNodeEnclosure
			case xnametypes.HSNBoard:
				hw.HMSHSNBoardLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocHSNBoard
			case xnametypes.MgmtSwitch:
				hw.HMSMgmtSwitchLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocMgmtSwitch
			case xnametypes.MgmtHLSwitch:
				hw.HMSMgmtHLSwitchLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocMgmtHLSwitch
			case xnametypes.CDUMgmtSwitch:
				hw.HMSCDUMgmtSwitchLocationInfo = rfChassisLocationInfo
				hw.HWInventoryByLocationType = HWInvByLocCDUMgmtSwitch
			}
		}
	// HWInv based on Redfish "System" Type.
	case xnametypes.Node:
		rfSystemLocationInfo = new(rf.SystemLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfSystemLocationInfo)
		if err == nil {
			hw.HMSNodeLocationInfo = rfSystemLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocNode
		}
	// HWInv based on "GPU" type
	case xnametypes.NodeAccel:
		rfNodeAccelLocationInfo = new(rf.ProcessorLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfNodeAccelLocationInfo)
		if err == nil {
			hw.HMSNodeAccelLocationInfo = rfNodeAccelLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocNodeAccel
		}
	// HWInv based on Redfish "Processor" Type.
	case xnametypes.Processor:
		rfProcessorLocationInfo = new(rf.ProcessorLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfProcessorLocationInfo)
		if err == nil {
			hw.HMSProcessorLocationInfo = rfProcessorLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocProcessor
		}
	// HWInv based on Redfish "Memory" Type.
	case xnametypes.Memory:
		rfMemoryLocationInfo = new(rf.MemoryLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfMemoryLocationInfo)
		if err == nil {
			hw.HMSMemoryLocationInfo = rfMemoryLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocMemory
		}
	// HWInv based on Redfish "Drive" Type.
	case xnametypes.Drive:
		rfDriveLocationInfo = new(rf.DriveLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfDriveLocationInfo)
		if err == nil {
			hw.HMSDriveLocationInfo = rfDriveLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocDrive
		}
	// HWInv based on Redfish "HSN NIC" Type.
	case xnametypes.NodeHsnNic:
		rfHSNNICLocationInfo = new(rf.NALocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfHSNNICLocationInfo)
		if err == nil {
			hw.HMSHSNNICLocationInfo = rfHSNNICLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocHSNNIC
		}
	// HWInv based on Redfish "PowerDistribution" (aka PDU) Type.
	case xnametypes.CabinetPDU:
		rfPDULocationInfo = new(rf.PowerDistributionLocationInfo)
		err = json.Unmarshal(locInfoJSON, rfPDULocationInfo)
		if err == nil {
			hw.HMSPDULocationInfo = rfPDULocationInfo
			hw.HWInventoryByLocationType = HWInvByLocPDU
		}
	// HWInv based on Redfish "Outlet" (e.g. of a PDU) Type.
	case xnametypes.CabinetPDUOutlet:
		fallthrough
	case xnametypes.CabinetPDUPowerConnector:
		rfOutletLocationInfo = new(rf.OutletLocationInfo)
		err = json.Unmarshal(locInfoJSON, rfOutletLocationInfo)
		if err == nil {
			hw.HMSOutletLocationInfo = rfOutletLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocOutlet
		}
	case xnametypes.CMMRectifier:
		rfCMMRectifierLocationInfo = new(rf.PowerSupplyLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfCMMRectifierLocationInfo)
		if err == nil {
			hw.HMSCMMRectifierLocationInfo = rfCMMRectifierLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocCMMRectifier
		}
	case xnametypes.NodeEnclosurePowerSupply:
		rfNodeEnclosurePowerSupplyLocationInfo = new(rf.PowerSupplyLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfNodeEnclosurePowerSupplyLocationInfo)
		if err == nil {
			hw.HMSNodeEnclosurePowerSupplyLocationInfo = rfNodeEnclosurePowerSupplyLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocNodeEnclosurePowerSupply
		}
	case xnametypes.NodeAccelRiser:
		rfNodeAccelRiserLocationInfo = new(rf.NodeAccelRiserLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfNodeAccelRiserLocationInfo)
		if err == nil {
			hw.HMSNodeAccelRiserLocationInfo = rfNodeAccelRiserLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocNodeAccelRiser
		}
	case xnametypes.NodeBMC:
		rfNodeBMCLocationInfo = new(rf.ManagerLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfNodeBMCLocationInfo)
		if err == nil {
			hw.HMSNodeBMCLocationInfo = rfNodeBMCLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocNodeBMC
		}
	case xnametypes.RouterBMC:
		rfRouterBMCLocationInfo = new(rf.ManagerLocationInfoRF)
		err = json.Unmarshal(locInfoJSON, rfRouterBMCLocationInfo)
		if err == nil {
			hw.HMSRouterBMCLocationInfo = rfRouterBMCLocationInfo
			hw.HWInventoryByLocationType = HWInvByLocRouterBMC
		}
	// No match - not a valid HMSType, always an error
	case xnametypes.HMSTypeInvalid:
		err = base.ErrHMSTypeInvalid
	default:
		err = base.ErrHMSTypeUnsupported
	}
	return err
}

//
// This function encode's the hwinv's type-specific LocationInfo struct
// into a free-form JSON byte array that can be stored schema-less in the
// database.
//
// NOTE: This function is the counterpart to DecodeLocationInfo().
//
// Returns: type's location info as JSON []byte representation, err = nil
//          Else, err != nil if encoding failed (and location_info is empty)
func (hw *HWInvByLoc) EncodeLocationInfo() ([]byte, error) {
	var err error
	var locInfoJSON []byte

	switch xnametypes.ToHMSType(hw.Type) {
	// HWInv based on Redfish "Chassis" Type.
	case xnametypes.Cabinet:
		locInfoJSON, err = json.Marshal(hw.HMSCabinetLocationInfo)
	case xnametypes.Chassis:
		locInfoJSON, err = json.Marshal(hw.HMSChassisLocationInfo)
	case xnametypes.ComputeModule:
		locInfoJSON, err = json.Marshal(hw.HMSComputeModuleLocationInfo)
	case xnametypes.RouterModule:
		locInfoJSON, err = json.Marshal(hw.HMSRouterModuleLocationInfo)
	case xnametypes.NodeEnclosure:
		locInfoJSON, err = json.Marshal(hw.HMSNodeEnclosureLocationInfo)
	case xnametypes.HSNBoard:
		locInfoJSON, err = json.Marshal(hw.HMSHSNBoardLocationInfo)
	case xnametypes.MgmtSwitch:
		locInfoJSON, err = json.Marshal(hw.HMSMgmtSwitchLocationInfo)
	case xnametypes.MgmtHLSwitch:
		locInfoJSON, err = json.Marshal(hw.HMSMgmtHLSwitchLocationInfo)
	case xnametypes.CDUMgmtSwitch:
		locInfoJSON, err = json.Marshal(hw.HMSCDUMgmtSwitchLocationInfo)
	// HWInv based on Redfish "System" Type.
	case xnametypes.Node:
		locInfoJSON, err = json.Marshal(hw.HMSNodeLocationInfo)
	// HWInv based on "GPU" type
	case xnametypes.NodeAccel:
		locInfoJSON, err = json.Marshal(hw.HMSNodeAccelLocationInfo)
	// HWInv based on Redfish "Processor" Type.
	case xnametypes.Processor:
		locInfoJSON, err = json.Marshal(hw.HMSProcessorLocationInfo)
	// HWInv based on Redfish "Memory" Type.
	case xnametypes.Memory:
		locInfoJSON, err = json.Marshal(hw.HMSMemoryLocationInfo)
	// HWInv based on Redfish "Drive" Type.
	case xnametypes.Drive:
		locInfoJSON, err = json.Marshal(hw.HMSDriveLocationInfo)
	// HWInv based on Redfish "HSN NIC" Type.
	case xnametypes.NodeHsnNic:
		locInfoJSON, err = json.Marshal(hw.HMSHSNNICLocationInfo)
	// HWInv based on Redfish "PowerDistribution" (aka PDU) Type.
	case xnametypes.CabinetPDU:
		locInfoJSON, err = json.Marshal(hw.HMSPDULocationInfo)
	// HWInv based on Redfish "Outlet" (e.g. of a PDU) Type.
	case xnametypes.CabinetPDUOutlet:
		fallthrough
	case xnametypes.CabinetPDUPowerConnector:
		locInfoJSON, err = json.Marshal(hw.HMSOutletLocationInfo)
	case xnametypes.CMMRectifier:
		locInfoJSON, err = json.Marshal(hw.HMSCMMRectifierLocationInfo)
	case xnametypes.NodeEnclosurePowerSupply:
		locInfoJSON, err = json.Marshal(hw.HMSNodeEnclosurePowerSupplyLocationInfo)
	case xnametypes.NodeAccelRiser:
		locInfoJSON, err = json.Marshal(hw.HMSNodeAccelRiserLocationInfo)
	case xnametypes.NodeBMC:
		locInfoJSON, err = json.Marshal(hw.HMSNodeBMCLocationInfo)
	case xnametypes.RouterBMC:
		locInfoJSON, err = json.Marshal(hw.HMSRouterBMCLocationInfo)
	// No match - not a valid HMS Type, always an error
	case xnametypes.HMSTypeInvalid:
		err = base.ErrHMSTypeInvalid
	// Not supported for this type.
	default:
		err = base.ErrHMSTypeUnsupported
	}
	return locInfoJSON, err
}

////////////////////////////////////////////////////////////////////////////
//
// Hardware Inventory - Field Replaceable Unit data
//
//   These are the properties of components that move with the physical
//   unit and may or may not have a matching location at the moment.  These
//   will eventually have their location histories tracked.
//
////////////////////////////////////////////////////////////////////////////

type HWInvByFRU struct {
	FRUID   string `json:"FRUID"`
	Type    string `json:"Type"`
	Subtype string `json:"Subtype"`

	// This is used as a descriminator to specify the type of *Info
	// struct that will be included below.
	HWInventoryByFRUType string `json:"HWInventoryByFRUType"`

	// One of (based on HWFRUInfoType):
	//   HMSType             Underlying RF Type      How named in json object
	HMSCabinetFRUInfo       *rf.ChassisFRUInfoRF   `json:"CabinetFRUInfo,omitempty"`
	HMSChassisFRUInfo       *rf.ChassisFRUInfoRF   `json:"ChassisFRUInfo,omitempty"` // Mountain chassis
	HMSComputeModuleFRUInfo *rf.ChassisFRUInfoRF   `json:"ComputeModuleFRUInfo,omitempty"`
	HMSRouterModuleFRUInfo  *rf.ChassisFRUInfoRF   `json:"RouterModuleFRUInfo,omitempty"`
	HMSNodeEnclosureFRUInfo *rf.ChassisFRUInfoRF   `json:"NodeEnclosureFRUInfo,omitempty"`
	HMSHSNBoardFRUInfo      *rf.ChassisFRUInfoRF   `json:"HSNBoardFRUInfo,omitempty"`
	HMSMgmtSwitchFRUInfo    *rf.ChassisFRUInfoRF   `json:"MgmtSwitchFRUInfo,omitempty"`
	HMSMgmtHLSwitchFRUInfo  *rf.ChassisFRUInfoRF   `json:"MgmtHLSwitchFRUInfo,omitempty"`
	HMSCDUMgmtSwitchFRUInfo *rf.ChassisFRUInfoRF   `json:"CDUMgmtSwitchFRUInfo,omitempty"`
	HMSNodeFRUInfo          *rf.SystemFRUInfoRF    `json:"NodeFRUInfo,omitempty"`
	HMSProcessorFRUInfo     *rf.ProcessorFRUInfoRF `json:"ProcessorFRUInfo,omitempty"`
	HMSNodeAccelFRUInfo     *rf.ProcessorFRUInfoRF `json:"NodeAccelFRUInfo,omitempty"`
	HMSMemoryFRUInfo        *rf.MemoryFRUInfoRF    `json:"MemoryFRUInfo,omitempty"`
	HMSDriveFRUInfo         *rf.DriveFRUInfoRF     `json:"DriveFRUInfo,omitempty"`
	HMSHSNNICFRUInfo        *rf.NAFRUInfoRF        `json:"NodeHsnNicFRUInfo,omitempty"`

	HMSPDUFRUInfo                      *rf.PowerDistributionFRUInfo `json:"PDUFRUInfo,omitempty"`
	HMSOutletFRUInfo                   *rf.OutletFRUInfo            `json:"OutletFRUInfo,omitempty"`
	HMSCMMRectifierFRUInfo             *rf.PowerSupplyFRUInfoRF     `json:"CMMRectifierFRUInfo,omitempty"`
	HMSNodeEnclosurePowerSupplyFRUInfo *rf.PowerSupplyFRUInfoRF     `json:"NodeEnclosurePowerSupplyFRUInfo,omitempty"`
	HMSNodeBMCFRUInfo                  *rf.ManagerFRUInfoRF         `json:"NodeBMCFRUInfo,omitempty"`
	HMSRouterBMCFRUInfo                *rf.ManagerFRUInfoRF         `json:"RouterBMCFRUInfo,omitempty"`
	HMSNodeAccelRiserFRUInfo           *rf.NodeAccelRiserFRUInfoRF  `json:"NodeAccelRiserFRUInfo,omitempty"`

	// TODO: Remaining types in hmsTypeArray
}

// HWInventoryByFRUType properties.  Used to select proper subtype in
// api schema.
// TODO: Remaining types
const (
	HWInvByFRUCabinet                  string = "HWInvByFRUCabinet"
	HWInvByFRUChassis                  string = "HWInvByFRUChassis"
	HWInvByFRUComputeModule            string = "HWInvByFRUComputeModule"
	HWInvByFRURouterModule             string = "HWInvByFRURouterModule"
	HWInvByFRUNodeEnclosure            string = "HWInvByFRUNodeEnclosure"
	HWInvByFRUHSNBoard                 string = "HWInvByFRUHSNBoard"
	HWInvByFRUMgmtSwitch               string = "HWInvByFRUMgmtSwitch"
	HWInvByFRUMgmtHLSwitch             string = "HWInvByFRUMgmtHLSwitch"
	HWInvByFRUCDUMgmtSwitch            string = "HWInvByFRUCDUMgmtSwitch"
	HWInvByFRUNode                     string = "HWInvByFRUNode"
	HWInvByFRUProcessor                string = "HWInvByFRUProcessor"
	HWInvByFRUNodeAccel                string = "HWInvByFRUNodeAccel"
	HWInvByFRUMemory                   string = "HWInvByFRUMemory"
	HWInvByFRUDrive                    string = "HWInvByFRUDrive"
	HWInvByFRUHSNNIC                   string = "HWInvByFRUNodeHsnNic"
	HWInvByFRUPDU                      string = "HWInvByFRUPDU"
	HWInvByFRUOutlet                   string = "HWInvByFRUOutlet"
	HWInvByFRUCMMRectifier             string = "HWInvByFRUCMMRectifier"
	HWInvByFRUNodeEnclosurePowerSupply string = "HWInvByFRUNodeEnclosurePowerSupply"
	HWInvByFRUNodeBMC                  string = "HWInvByFRUNodeBMC"
	HWInvByFRURouterBMC                string = "HWInvByFRURouterBMC"
	HWInvByFRUNodeAccelRiser           string = "HWInvByFRUNodeAccelRiser"
)

////////////////////////////////////////////////////////////////////////////
// Encoding/decoding: HW Inventory FRU info
///////////////////////////////////////////////////////////////////////////

// This routine takes raw FRU info captured as free-form JSON (e.g.
// from a schema-free database field) and unmarshals it into the correct struct
// for the type with the proper type-specific name.
//
// NOTEs: The fruInfoJSON array should be that produced by EncodeFRUInfo.
//        MODIFIES caller.
//
// Return: If err != nil hf is unmodified and operation failed.
//         Else, the type's *FRUInfo pointer is set to the expected struct.
func (hf *HWInvByFRU) DecodeFRUInfo(fruInfoJSON []byte) error {
	var (
		err                               error = nil
		rfChassisFRUInfo                  *rf.ChassisFRUInfoRF
		rfSystemFRUInfo                   *rf.SystemFRUInfoRF
		rfProcessorFRUInfo                *rf.ProcessorFRUInfoRF
		rfNodeAccelFRUInfo                *rf.ProcessorFRUInfoRF
		rfMemoryFRUInfo                   *rf.MemoryFRUInfoRF
		rfDriveFRUInfo                    *rf.DriveFRUInfoRF
		rfHSNNICFRUInfo                   *rf.NAFRUInfoRF
		rfPDUFRUInfo                      *rf.PowerDistributionFRUInfo
		rfOutletFRUInfo                   *rf.OutletFRUInfo
		rfCMMRectifierFRUInfo             *rf.PowerSupplyFRUInfoRF
		rfNodeEnclosurePowerSupplyFRUInfo *rf.PowerSupplyFRUInfoRF
		rfNodeBMCFRUInfo                  *rf.ManagerFRUInfoRF
		rfRouterBMCFRUInfo                *rf.ManagerFRUInfoRF
		rfNodeAccelRiserFRUInfo           *rf.NodeAccelRiserFRUInfoRF
	)

	switch xnametypes.ToHMSType(hf.Type) {
	// HWInv based on Redfish "Chassis" Type.  Identical structs (for now).
	case xnametypes.Cabinet:
		fallthrough
	case xnametypes.Chassis:
		fallthrough
	case xnametypes.ComputeModule:
		fallthrough
	case xnametypes.RouterModule:
		fallthrough
	case xnametypes.NodeEnclosure:
		fallthrough
	case xnametypes.HSNBoard:
		fallthrough
	case xnametypes.MgmtSwitch:
		fallthrough
	case xnametypes.MgmtHLSwitch:
		fallthrough
	case xnametypes.CDUMgmtSwitch:
		rfChassisFRUInfo = new(rf.ChassisFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfChassisFRUInfo)
		if err == nil {
			// Assign struct to appropriate name for type.
			switch xnametypes.ToHMSType(hf.Type) {
			case xnametypes.Cabinet:
				hf.HMSCabinetFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUCabinet
			case xnametypes.Chassis:
				hf.HMSChassisFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUChassis
			case xnametypes.ComputeModule:
				hf.HMSComputeModuleFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUComputeModule
			case xnametypes.RouterModule:
				hf.HMSRouterModuleFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRURouterModule
			case xnametypes.NodeEnclosure:
				hf.HMSNodeEnclosureFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUNodeEnclosure
			case xnametypes.HSNBoard:
				hf.HMSHSNBoardFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUHSNBoard
			case xnametypes.MgmtSwitch:
				hf.HMSMgmtSwitchFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUMgmtSwitch
			case xnametypes.MgmtHLSwitch:
				hf.HMSMgmtHLSwitchFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUMgmtHLSwitch
			case xnametypes.CDUMgmtSwitch:
				hf.HMSCDUMgmtSwitchFRUInfo = rfChassisFRUInfo
				hf.HWInventoryByFRUType = HWInvByFRUCDUMgmtSwitch
			}
		}
	// HWInv based on Redfish "System" Type.
	case xnametypes.Node:
		rfSystemFRUInfo = new(rf.SystemFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfSystemFRUInfo)
		if err == nil {
			hf.HMSNodeFRUInfo = rfSystemFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUNode
		}
	// HWInv based on "GPU" type
	case xnametypes.NodeAccel:
		rfNodeAccelFRUInfo = new(rf.ProcessorFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfNodeAccelFRUInfo)
		if err == nil {
			hf.HMSNodeAccelFRUInfo = rfNodeAccelFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUNodeAccel
		}
	// HWInv based on Redfish "Processor" Type.
	case xnametypes.Processor:
		rfProcessorFRUInfo = new(rf.ProcessorFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfProcessorFRUInfo)
		if err == nil {
			hf.HMSProcessorFRUInfo = rfProcessorFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUProcessor
		}
	// HWInv based on Redfish "Memory" Type.
	case xnametypes.Memory:
		rfMemoryFRUInfo = new(rf.MemoryFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfMemoryFRUInfo)
		if err == nil {
			hf.HMSMemoryFRUInfo = rfMemoryFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUMemory
		}
	// HWInv based on Redfish "Drive" Type.
	case xnametypes.Drive:
		rfDriveFRUInfo = new(rf.DriveFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfDriveFRUInfo)
		if err == nil {
			hf.HMSDriveFRUInfo = rfDriveFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUDrive
		}
	// HWInv based on Redfish "Memory" Type.
	case xnametypes.NodeHsnNic:
		rfHSNNICFRUInfo = new(rf.NAFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfHSNNICFRUInfo)
		if err == nil {
			hf.HMSHSNNICFRUInfo = rfHSNNICFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUHSNNIC
		}
	// HWInv based on Redfish "PowerDistribution" Type.
	case xnametypes.CabinetPDU:
		rfPDUFRUInfo = new(rf.PowerDistributionFRUInfo)
		err = json.Unmarshal(fruInfoJSON, rfPDUFRUInfo)
		if err == nil {
			hf.HMSPDUFRUInfo = rfPDUFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUPDU
		}
	// HWInv based on Redfish "Outlet" (e.g. of a PDU) Type.
	case xnametypes.CabinetPDUOutlet:
		fallthrough
	case xnametypes.CabinetPDUPowerConnector:
		rfOutletFRUInfo = new(rf.OutletFRUInfo)
		err = json.Unmarshal(fruInfoJSON, rfOutletFRUInfo)
		if err == nil {
			hf.HMSOutletFRUInfo = rfOutletFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUOutlet
		}
	// HWInv based on Redfish "PowerSupply" Type.
	case xnametypes.CMMRectifier:
		rfCMMRectifierFRUInfo = new(rf.PowerSupplyFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfCMMRectifierFRUInfo)
		if err == nil {
			hf.HMSCMMRectifierFRUInfo = rfCMMRectifierFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUCMMRectifier
		}
	// HWInv based on Redfish "PowerSupply" Type.
	case xnametypes.NodeEnclosurePowerSupply:
		rfNodeEnclosurePowerSupplyFRUInfo = new(rf.PowerSupplyFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfNodeEnclosurePowerSupplyFRUInfo)
		if err == nil {
			hf.HMSNodeEnclosurePowerSupplyFRUInfo = rfNodeEnclosurePowerSupplyFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUNodeEnclosurePowerSupply
		}
	// HWInv based on Redfish "NodeAccelRiser" Type.
	case xnametypes.NodeAccelRiser:
		rfNodeAccelRiserFRUInfo = new(rf.NodeAccelRiserFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfNodeAccelRiserFRUInfo)
		if err == nil {
			hf.HMSNodeAccelRiserFRUInfo = rfNodeAccelRiserFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUNodeAccelRiser
		}
	// HWInv based on Redfish "Manager" Type.
	case xnametypes.NodeBMC:
		rfNodeBMCFRUInfo = new(rf.ManagerFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfNodeBMCFRUInfo)
		if err == nil {
			hf.HMSNodeBMCFRUInfo = rfNodeBMCFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRUNodeBMC
		}
	// HWInv based on Redfish "Manager" Type.
	case xnametypes.RouterBMC:
		rfRouterBMCFRUInfo = new(rf.ManagerFRUInfoRF)
		err = json.Unmarshal(fruInfoJSON, rfRouterBMCFRUInfo)
		if err == nil {
			hf.HMSRouterBMCFRUInfo = rfRouterBMCFRUInfo
			hf.HWInventoryByFRUType = HWInvByFRURouterBMC
		}
	// No match - not a valid HMSType, always an error
	case xnametypes.HMSTypeInvalid:
		err = base.ErrHMSTypeInvalid
	default:
		err = base.ErrHMSTypeUnsupported
	}
	return err
}

//
// This function encode's the hwinv's type-specific FRU info struct
// into a free-form JSON byte array that can be stored schema-less in the
// database.
//
// NOTE: This function is the counterpart to DecodeFRUInfo().
//
// Returns: FRU's info as JSON []byte representation, err = nil
//          Else, err != nil if encoding failed (plus, []byte value is empty)
func (hf *HWInvByFRU) EncodeFRUInfo() ([]byte, error) {
	var err error
	var fruInfoJSON []byte

	switch xnametypes.ToHMSType(hf.Type) {
	// HWInv based on Redfish "Chassis" Type.
	case xnametypes.Cabinet:
		fruInfoJSON, err = json.Marshal(hf.HMSCabinetFRUInfo)
	case xnametypes.Chassis:
		fruInfoJSON, err = json.Marshal(hf.HMSChassisFRUInfo)
	case xnametypes.ComputeModule:
		fruInfoJSON, err = json.Marshal(hf.HMSComputeModuleFRUInfo)
	case xnametypes.RouterModule:
		fruInfoJSON, err = json.Marshal(hf.HMSRouterModuleFRUInfo)
	case xnametypes.NodeEnclosure:
		fruInfoJSON, err = json.Marshal(hf.HMSNodeEnclosureFRUInfo)
	case xnametypes.HSNBoard:
		fruInfoJSON, err = json.Marshal(hf.HMSHSNBoardFRUInfo)
	case xnametypes.MgmtSwitch:
		fruInfoJSON, err = json.Marshal(hf.HMSMgmtSwitchFRUInfo)
	case xnametypes.MgmtHLSwitch:
		fruInfoJSON, err = json.Marshal(hf.HMSMgmtHLSwitchFRUInfo)
	case xnametypes.CDUMgmtSwitch:
		fruInfoJSON, err = json.Marshal(hf.HMSCDUMgmtSwitchFRUInfo)
	// HWInv based on Redfish "System" Type.
	case xnametypes.Node:
		fruInfoJSON, err = json.Marshal(hf.HMSNodeFRUInfo)
	// HWInv based on "GPU" type
	case xnametypes.NodeAccel:
		fruInfoJSON, err = json.Marshal(hf.HMSNodeAccelFRUInfo)
	// HWInv based on Redfish "Processor" Type.
	case xnametypes.Processor:
		fruInfoJSON, err = json.Marshal(hf.HMSProcessorFRUInfo)
	// HWInv based on Redfish "Memory" Type.
	case xnametypes.Memory:
		fruInfoJSON, err = json.Marshal(hf.HMSMemoryFRUInfo)
	// HWInv based on Redfish "Processor" Type.
	case xnametypes.Drive:
		fruInfoJSON, err = json.Marshal(hf.HMSDriveFRUInfo)
	// HWInv based on Redfish "HSN NIC" Type.
	case xnametypes.NodeHsnNic:
		fruInfoJSON, err = json.Marshal(hf.HMSHSNNICFRUInfo)
	// HWInv based on Redfish "PowerDistribution" (aka PDU) Type.
	case xnametypes.CabinetPDU:
		fruInfoJSON, err = json.Marshal(hf.HMSPDUFRUInfo)
	// HWInv based on Redfish "Outlet" (e.g. of a PDU) Type.
	case xnametypes.CabinetPDUOutlet:
		fallthrough
	case xnametypes.CabinetPDUPowerConnector:
		fruInfoJSON, err = json.Marshal(hf.HMSOutletFRUInfo)
	// HWInv based on Redfish "PowerSupply" Type.
	case xnametypes.CMMRectifier:
		fruInfoJSON, err = json.Marshal(hf.HMSCMMRectifierFRUInfo)
	// HWInv based on Redfish "PowerSupply" Type.
	case xnametypes.NodeEnclosurePowerSupply:
		fruInfoJSON, err = json.Marshal(hf.HMSNodeEnclosurePowerSupplyFRUInfo)
	// HWInv based on Redfish "NodeAccelRiser" Type.
	case xnametypes.NodeAccelRiser:
		fruInfoJSON, err = json.Marshal(hf.HMSNodeAccelRiserFRUInfo)
	// HWInv based on Redfish "Manager" Type.
	case xnametypes.NodeBMC:
		fruInfoJSON, err = json.Marshal(hf.HMSNodeBMCFRUInfo)
	// HWInv based on Redfish "Manager" Type.
	case xnametypes.RouterBMC:
		fruInfoJSON, err = json.Marshal(hf.HMSRouterBMCFRUInfo)
	// No match - not a valid HMS Type, always an error
	case xnametypes.HMSTypeInvalid:
		err = base.ErrHMSTypeInvalid
	// Not supported for this type.
	default:
		err = base.ErrHMSTypeUnsupported
	}
	return fruInfoJSON, err
}
