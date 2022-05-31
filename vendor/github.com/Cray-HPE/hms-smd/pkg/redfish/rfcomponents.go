// MIT License
//
// (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
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

package rf

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	base "github.com/Cray-HPE/hms-base"
)

/////////////////////////////////////////////////////////////////////////////
//
//
// Redfish top-level "Component" discovery (Chassis, ComputerSystem, Manager)
//
//
/////////////////////////////////////////////////////////////////////////////

type ComponentDescription struct {
	ID             string `json:"ID"`   // Key, HMS ID, i.e. xname
	Type           string `json:"Type"` // Key, HMS Type
	Domain         string `json:"Domain,omitempty"`
	FQDN           string `json:"FQDN,omitempty"`
	RedfishType    string `json:"RedfishType"`
	RedfishSubtype string `json:"RedfishSubtype"`
	MACAddr        string `json:"MACAddr,omitempty"`
	UUID           string `json:"UUID,omitempty"`
	OdataID        string `json:"OdataID"`
	RfEndpointID   string `json:"RedfishEndpointID"`
}

type InventoryData struct {
	Ordinal        int    `json:"Ordinal"`
	RawOrdinal     int    `json:"-"`
	Status         string `json:"Status"`
	State          string `json:"State"`
	Flag           string `json:"Flag"`
	Arch           string `json:"Arch"`
	NetType        string `json:"NetType"`
	DefaultRole    string `json:"DefaultRole"`
	DefaultSubRole string `json:"DefaultSubRole"`
	DefaultClass   string `json:"DefaultClass"`
	FRUID          string `json:"FRUID"`
	Subtype        string `json:"Subtype"` // HMS Subtype, NYI
}

// Type specific info for Redfish Chassis components
type ComponentChassisInfo struct {
	Name    string          `json:"Name,omitempty"`
	Actions *ChassisActions `json:"Actions,omitempty"`
}

// Type specific info for Redfish ComputerSystem components
type ComponentSystemInfo struct {
	Name       string                 `json:"Name,omitempty"`
	Actions    *ComputerSystemActions `json:"Actions,omitempty"`
	EthNICInfo []*EthernetNICInfo     `json:"EthernetNICInfo,omitempty"`
	PowerCtlInfo
	Controls   []*Control             `json:"Controls,omitempty"`
}

type ComponentManagerInfo struct {
	Name       string             `json:"Name,omitempty"`
	Actions    *ManagerActions    `json:"Actions,omitempty"`
	EthNICInfo []*EthernetNICInfo `json:"EthernetNICInfo,omitempty"`
}

type ComponentPDUInfo struct {
	Name    string                    `json:"Name,omitempty"`
	Actions *PowerDistributionActions `json:"Actions,omitempty"`
}

type ComponentOutletInfo struct {
	Name    string         `json:"Name,omitempty"`
	Actions *OutletActions `json:"Actions,omitempty"`
}

type EthernetNICInfo struct {
	RedfishId           string `json:"RedfishId"`
	Oid                 string `json:"@odata.id"`
	Description         string `json:"Description,omitempty"`
	FQDN                string `json:"FQDN,omitempty"`
	Hostname            string `json:"Hostname,omitempty"`
	InterfaceEnabled    *bool  `json:"InterfaceEnabled,omitempty"`
	MACAddress          string `json:"MACAddress"`
	PermanentMACAddress string `json:"PermanentMACAddress,omitempty"`
}

type PowerCtlInfo struct {
	PowerURL string          `json:"PowerURL,omitempty"`
	PowerCtl []*PowerControl `json:"PowerControl,omitempty"`
}

type PowerControl struct {
	ResourceID
	MemberId           string        `json:"MemberId,omitempty"`
	Name               string        `json:"Name,omitempty"`
	PowerCapacityWatts int           `json:"PowerCapacityWatts,omitempty"`
	OEM                *PwrCtlOEM    `json:"OEM,omitempty"`
	RelatedItem        []*ResourceID `json:"RelatedItem,omitempty"`
}

type PwrCtlOEM struct {
	Cray *PwrCtlOEMCray `json:"Cray,omitempty"`
	HPE  *PwrCtlOEMHPE  `json:"HPE,omitempty"`
}

type PwrCtlOEMCray struct {
	PowerIdleWatts  int           `json:"PowerIdleWatts,omitempty"`
	PowerLimit      *CrayPwrLimit `json:"PowerLimit,omitempty"`
	PowerResetWatts int           `json:"PowerResetWatts,omitempty"`
}

type CrayPwrLimit struct {
	Min int `json:"Min,omitempty"`
	Max int `json:"Max,omitempty"`
}

type PwrCtlOEMHPE struct {
	PowerLimit CrayPwrLimit     `json:"PowerLimit"`
	PowerRegulationEnabled bool `json:"PowerRegulationEnabled"`
	Status     string           `json:"Status"`
	Target     string           `json:"Target"`
}

type PwrCtlRelatedItem struct {
	Oid string `json:"@odata.id"`
}

type Control struct {
	URL     string    `json:"URL"`
	Control RFControl `json:"Control"`
}

/////////////////////////////////////////////////////////////////////////////
//
// Redfish Chassis component discovery
//
/////////////////////////////////////////////////////////////////////////////

// These are the discovered attributes for a particular entry in the /Chassis/
// URL for a given RF endpoint.  The chassis is important to understand where
// a particular system may be in an enclosure (if it is not the only one) and
// what type it is.  Additionally, power operations may be carried out at the
// chassis level and there may be some nesting of chassis (e.g. blade and cage)
// that we may need to understand.
type EpChassis struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	// Embedded struct - Chassis specific info
	ComponentChassisInfo

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	BaseOdataID string `json:"BaseOdataID"`
	ChassisURL  string `json:"ChassisURL"` // Full URL to this Chassis

	PChassisOID  string `json:"pChassisOID"`  // odata.id for parent, if nested
	PChassisType string `json:"pChassisType"` // ChassisType enum from parent

	// The Oids here are map keys stored by the endpoint.  Via the
	// epRF ptr we can obtain the Go EpManager or EpSystem structs.
	//
	// We will want to store some higher-level information when we know
	// more about the system type, but for now they are just for navigation.
	//
	// Note that since there is always only a single "ContainedBy" link, we
	// use this, if it exists, to set the parent chassis OdataID field above
	// (e.g. PChassisOID/PChassesType).
	ManagedBy    []ResourceID `json:"managedBy"`
	ChildChassis []ResourceID `json:"childChassis"`
	ChildSystems []ResourceID `json:"childSystems"`

	LastStatus string `json:"lastStatus"`

	ChassisRF     Chassis          `json:"chassisRF"`
	chassisURLRaw *json.RawMessage //`json:"chassisURLRaw"`

	Power         *EpPower        `json:"Power"`
	PowerSupplies EpPowerSupplies `json:"PowerSupplies"`

	epRF *RedfishEP // Backpointer, for connection details, etc.
}

// Set of EpChassis, representing a Redfish "chassis" under some RF endpoint.
type EpChassisSet struct {
	Num  int                   `json:"num"`
	OIDs map[string]*EpChassis `json:"oids"`
}

// Initializes EpSystem struct with minimal information needed to
// discover it, i.e. endpoint info and the OdataID of the system to look at.
// This should be the only way this struct is created.
func NewEpChassis(epRF *RedfishEP, odataID ResourceID, rawOrdinal int) *EpChassis {
	c := new(EpChassis)
	c.Type = base.HMSTypeInvalid.String() // Must be updated later
	c.OdataID = odataID.Oid
	c.BaseOdataID = odataID.Basename()
	c.RedfishType = ChassisType
	c.RfEndpointID = epRF.ID
	c.LastStatus = NotYetQueried
	c.Ordinal = -1
	c.RawOrdinal = rawOrdinal
	c.epRF = epRF
	return c
}

// Makes contact with remote endpoints to discover basic information about
// all Redfish Chassis in EpChassisSet.  Typically this would be all
// Chassis under a particular Redfish entry point.  Phase1 discovery
// fetches all relevant data from the RF entry point but does not fully
// discover all info.  This is left for Phase2, which is intended to be
// run only after ALL components (managers, chassis, systems, etc.) have
// completed phase 1 under a particular endpoint.
func (cs *EpChassisSet) discoverRemotePhase1() {
	for _, c := range cs.OIDs {
		c.discoverRemotePhase1()
	}
}

// Makes contact with remote endpoint to discover basic information about
// the EpChassis (i.e. a particular Redfish Chassis).  This is the first
// step once the struct is initialized.  After Phase1 discovery
// is complete, phase 2 discovery can be performed.  The second phase does
// not obtain new data but rather explores data from phase1 further, and
// should be done after all components under a particular RF entry point
// have completed Phase 1 (managers, chassis, systems, etc.).
func (c *EpChassis) discoverRemotePhase1() {
	// Should never happen
	if c.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for Chassis odataID: %s\n",
			c.OdataID)
		c.LastStatus = EndpointInvalid
		return
	}
	// Workaround - DST1372
	if c.OdataID == "/redfish/v1/Chassis/RackMount/HSBackplane" {
		c.LastStatus = RedfishSubtypeNoSupport
		c.RedfishSubtype = RFSubtypeUnknown
		return
	}
	c.ChassisURL = c.epRF.FQDN + c.OdataID

	path := c.OdataID
	url := c.epRF.FQDN + path
	topURL := url
	chassisURLJSON, err := c.epRF.GETRelative(path)
	if err != nil || chassisURLJSON == nil {
		if err == ErrRFDiscURLNotFound {
			errlog.Printf("%s: Redfish bug! Link %s was dead (404).  "+
				"Will try to continue.  No component will be created.",
				c.epRF.ID, path)
			c.LastStatus = RedfishSubtypeNoSupport
			c.RedfishSubtype = RFSubtypeUnknown
		} else {
			c.LastStatus = HTTPsGetFailed
		}
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, chassisURLJSON)
	}
	c.chassisURLRaw = &chassisURLJSON
	c.LastStatus = HTTPsGetOk

	// Decode JSON into Chassis structure
	if err := json.Unmarshal(*c.chassisURLRaw, &c.ChassisRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			c.LastStatus = EPResponseFailedDecode
			return
		}
	}
	c.PChassisOID = c.ChassisRF.Links.ContainedBy.Oid
	c.RedfishSubtype = c.ChassisRF.ChassisType
	c.ManagedBy = c.ChassisRF.Links.ManagedBy
	c.ChildChassis = c.ChassisRF.Links.Contains
	c.ChildSystems = c.ChassisRF.Links.ComputerSystems
	if c.ChassisRF.Actions != nil {
		c.Actions = c.ChassisRF.Actions
	}

	// Workaround CASMHMS-4954 Apollo 6500 Enclosures missing Model.
	// Use Name in place of Model.
	if c.ChassisRF.Model == "" {
		c.ChassisRF.Model = c.ChassisRF.Name
	}

	//
	// Get link to Chassis' Power object
	//

	if c.ChassisRF.Power.Oid == "" {
		//errlog.Printf("%s: No Power obj found.\n", topURL)
		c.PowerSupplies.Num = 0
		c.PowerSupplies.OIDs = make(map[string]*EpPowerSupply)
	} else {
		//create a new EpPower object using chassis and Power.OID
		c.Power = NewEpPower(c, ResourceID{c.ChassisRF.Power.Oid})
		//retrieve the Power RF
		c.Power.discoverRemotePhase1()
		//discover any PowerSupplies

		if len(c.Power.PowerRF.PowerSupplies) > 0 {
			if c.Power.PowerRF.PowerSuppliesOCount > 0 && c.Power.PowerRF.PowerSuppliesOCount != len(c.Power.PowerRF.PowerSupplies) {
				errlog.Printf("%s: PowerSupplies@odata.count != PowerSupplies array len\n", url)
			} else {
				c.PowerSupplies.Num = len(c.Power.PowerRF.PowerSupplies)
				c.PowerSupplies.OIDs = make(map[string]*EpPowerSupply)
				//FIX: this will not result in the PowerSupplies being sorted
				//sort.Sort(ResourceIDSlice(c.Power.PowerRF.PowerSupplies))
				for i, powerSupply := range c.Power.PowerRF.PowerSupplies {
					pID := powerSupply.Oid
					c.PowerSupplies.OIDs[pID] = NewEpPowerSupply(c.Power, ResourceID{pID}, i)
				}
				//invoke the series of discoverRemotePhase1 calls for each PowerSupply
				c.PowerSupplies.discoverRemotePhase1()
			}
		}

	}

	c.LastStatus = VerifyingData
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(c, "", "   ")
		errlog.Printf("%s: %s\n", topURL, jout)
	}
}

// To be completed when all Chassis info has been retrieved via Redfish
// from the parent endpoint.  Establishes HMS data to provide higher-level
// data structures that can be integrated into a complete view of the
// system.
func (cs *EpChassisSet) discoverLocalPhase2() error {
	var savedError error
	for i, c := range cs.OIDs {
		c.discoverLocalPhase2()
		if c.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF ChassisType not supported: %s",
				i, c.RedfishSubtype)
		} else if c.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, c.LastStatus)
			errlog.Printf("Chassis discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// This is the second discovery phase, after all information from
// the parent endpoint has been gathered.  This is not really intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the endpoint first, so that it is available during later steps.
func (c *EpChassis) discoverLocalPhase2() {
	// Should never happen
	if c.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for system odataID: %s\n",
			c.OdataID)
		c.LastStatus = EndpointInvalid
		return
	}
	if c.LastStatus != VerifyingData {
		return
	}
	// There may be chassis types that are not supported.
	c.Type = c.epRF.getChassisHMSType(c)
	if c.Type == base.HMSTypeInvalid.String() {
		c.LastStatus = RedfishSubtypeNoSupport
		return
	}
	c.Ordinal = c.epRF.getChassisOrdinal(c)
	c.ID = c.epRF.getChassisHMSID(c, c.Type, c.Ordinal)
	if c.ID == "" {
		c.LastStatus = RedfishSubtypeNoSupport
		return
	}
	c.Name = c.ChassisRF.Name

	// Sets up HMS state fields using Status/State/Health info from Redfish
	c.discoverComponentState()

	// TODO: actually discover these
	c.Arch = base.ArchX86.String()
	c.NetType = base.NetSling.String()

	// Check if we have something valid to insert into the data store
	hmsType := base.GetHMSType(c.ID)
	if !base.IsHMSTypeContainer(hmsType) || c.Type != hmsType.String() {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s' != %s) for %s\n",
			c.ID, c.Type, hmsType.String(), c.ChassisURL)
		c.LastStatus = VerificationFailed
		return
	}

	// Complete discovery and verify subcomponents
	var childStatus string = DiscoverOK
	if err := c.PowerSupplies.discoverLocalPhase2(); err != nil {
		fmt.Printf("c.PowerSupplies.discoverLocalPhase2(): returned err %v", err)
		childStatus = ChildVerificationFailed
	}

	c.LastStatus = childStatus
}

// Sets up HMS state fields using Status/State/Health info from Redfish
func (c *EpChassis) discoverComponentState() {
	// HSNBoard here is a workaround, should never be legitmately absent.
	if c.ChassisRF.Status.State != "Absent" ||
		c.Type == base.HSNBoard.String() {

		// Chassis status is too unpredictable and no clear what Ready
		// means anyways since it's not a node or a controller.  So just
		// leave everything at max of on, but set warning for
		// weird state or health.  Mountain components never get
		// higher than On currently anyways.
		c.Status = "Populated"
		c.State = base.StatePopulated.String()
		c.Flag = base.FlagOK.String()
		if c.ChassisRF.PowerState != "" {
			if c.ChassisRF.PowerState == POWER_STATE_OFF ||
				c.ChassisRF.PowerState == POWER_STATE_POWERING_ON {
				c.State = base.StateOff.String()
			} else if c.ChassisRF.PowerState == POWER_STATE_ON ||
				c.ChassisRF.PowerState == POWER_STATE_POWERING_OFF {
				c.State = base.StateOn.String()
			}
		}
		// Don't set flags on Populated components - not tracked.
		// If somehow something that should be tracked is set to
		// populated, it can be fixed only via rediscovery.
		if c.State != base.StatePopulated.String() {
			if c.ChassisRF.Status.State != "Enabled" {
				c.Flag = base.FlagWarning.String()
			}
			if c.ChassisRF.Status.Health == "Warning" {
				c.Flag = base.FlagWarning.String()
			} else if c.ChassisRF.Status.Health == "Critical" {
				c.Flag = base.FlagAlert.String()
			}
		}
		generatedFRUID, err := GetChassisFRUID(c)
		if err != nil {
			errlog.Printf("FRUID Error: %s\n", err.Error())
			errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
		}
		c.FRUID = generatedFRUID
	} else {
		c.Status = "Empty"
		c.State = base.StateEmpty.String()
		//the state of the component is known (empty), it is not locked, does not have an alert or warning, so therefore Flag defaults to OK.
		c.Flag = base.FlagOK.String()
	}
}

/////////////////////////////////////////////////////////////////////////////
//
// Redfish Manager component discovery
//
/////////////////////////////////////////////////////////////////////////////

// These are the discovered attributes for a particular entry in the /Manager/
// URL for a given RF endpoint.  The manager is important to understand as
// they will represent important components such as BMCs and a top-level
// manager (if there is more than one) will implement the Redfish entry point
// we will be interacting with.
type EpManager struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	// Embedded struct: Manager specific data
	ComponentManagerInfo

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	ManagerURL  string `json:"ManagerURL"` // Full URL to this Manager
	BaseOdataID string `json:"BaseOdataID"`

	ChassisOID  string `json:"ChassisOID"`  // odata.id for containing chassis
	ChassisType string `json:"ChassisType"` // ChassisType enum from container

	LastStatus string `json:"LastStatus"`

	// The oids here are map keys for structs stored in the EpEndpoint.  Via
	// the epRF ptr we can obtain the correspodning Go EpSystem or EpChassis
	// structs.  While a manager will usually manage a single system or
	// multi-system enclosure or chassis, it is not necessarily a strict
	// one-for-one mapping in either direction.
	//
	// We will want to store some higher-level information when we know
	// more about the Manager type, but for now they are just for navigation.
	//
	// Note that since there is always only a single "ManagerInChassis" link,
	// use this, if it exists, to set the ChassisOID field above.
	ManagerInChassis ResourceID   `json:"managerInChassis"`
	ManagedChassis   []ResourceID `json:"managedChassis"`
	ManagedSystems   []ResourceID `json:"managedSystems"`

	ManagerRF     Manager          `json:"managerRF"`
	managerURLRaw *json.RawMessage //`json:"managerURLRaw"`

	// Ethernet Interfaces are children of Managers or Systems, not top
	// level components like Chassis or Managers.  Therefore we put the
	// local collection here.  Chassis or Managers are consolodated at the
	// EpEndpoint level in maps and we will use the odataID recorded here to
	// reference these via the epRF pointer.
	ENetInterfaces EpEthInterfaces `json:"enetInterfaces"`

	epRF *RedfishEP // Backpointer, for connection details, etc.
}

// Set of EpManager, representing a Redfish "Manager" under some RF endpoint.
type EpManagers struct {
	Num  int                   `json:"num"`
	OIDs map[string]*EpManager `json:"oids"`
}

// Initializes EpSystem struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the system to look at.
// This should be the only way this struct is created.
func NewEpManager(epRF *RedfishEP, odataID ResourceID, rawOrdinal int) *EpManager {
	m := new(EpManager)
	m.Type = base.HMSTypeInvalid.String() // Must be set properly later
	m.OdataID = odataID.Oid
	m.BaseOdataID = odataID.Basename()
	m.RedfishType = ManagerType
	m.LastStatus = NotYetQueried
	m.Ordinal = -1
	m.RawOrdinal = rawOrdinal
	m.RfEndpointID = epRF.ID
	m.epRF = epRF
	return m
}

// Makes contact with remote endpoints to discover basic information about
// all Redfish managers in EpManagers.  Typically this would be all
// Managers under a particular Redfish entry point.  Phase1 discovery
// fetches all relevant data from the RF entry point but does not fully
// discover all info.  This is left for Phase2, which is intended to be
// run only after ALL components (managers, chassis, systems, etc.) have
// completed phase 1 under a particular endpoint.
func (ms *EpManagers) discoverRemotePhase1() {
	for _, m := range ms.OIDs {
		m.discoverRemotePhase1()
	}
}

// Makes contact with remote endpoint to discover basic information about
// this EpManager (i.e. a particular Redfish "Manager", e.g. BMC).  This is
// the first step once the struct is initialized.  After Phase1 discovery
// is complete, phase 2 discovery can be performed.  The second phase does
// not obtain new data but rather explores data from phase1 further, and
// should be done after all components under a particular RF entry point
// have completed Phase 1.
func (m *EpManager) discoverRemotePhase1() {
	// Should never happen
	if m.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for Manager odataID: %s\n",
			m.OdataID)
		m.LastStatus = EndpointInvalid
		return
	}
	m.ManagerURL = m.epRF.FQDN + m.OdataID

	path := m.OdataID
	url := m.epRF.FQDN + path
	topURL := url
	managerURLJSON, err := m.epRF.GETRelative(path)
	if err != nil || managerURLJSON == nil {
		m.LastStatus = HTTPsGetFailed
		return
	}
	m.managerURLRaw = &managerURLJSON
	m.LastStatus = HTTPsGetOk

	// Decode JSON into Manager structure
	if err := json.Unmarshal(*m.managerURLRaw, &m.ManagerRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			m.LastStatus = EPResponseFailedDecode
			return
		}
	}
	m.RedfishSubtype = m.ManagerRF.ManagerType
	m.UUID = m.ManagerRF.UUID

	m.ManagerInChassis = m.ManagerRF.Links.ManagerInChassis
	m.ManagedChassis = m.ManagerRF.Links.ManagerForChassis
	m.ManagedSystems = m.ManagerRF.Links.ManagerForServers
	if m.ManagerRF.Actions != nil {
		m.Actions = m.ManagerRF.Actions
	}

	// Get link to Manager's ethernet interfaces
	if m.ManagerRF.EthernetInterfaces.Oid == "" {
		errlog.Printf("%s: No EthernetInterfaces Found.\n", topURL)
		m.ENetInterfaces.Num = 0
		m.ENetInterfaces.OIDs = make(map[string]*EpEthInterface)
	} else {
		path = m.ManagerRF.EthernetInterfaces.Oid
		url = m.epRF.FQDN + path
		ethIfacesJSON, err := m.epRF.GETRelative(path)
		if err != nil || ethIfacesJSON == nil {
			m.LastStatus = HTTPsGetFailed
			return
		}
		if rfDebug > 0 {
			errlog.Printf("%s: %s\n", url, ethIfacesJSON)
		}
		m.LastStatus = HTTPsGetOk

		var ethInfo EthernetInterfaceCollection
		if err := json.Unmarshal(ethIfacesJSON, &ethInfo); err != nil {
			errlog.Printf("Failed to decode %s: %s\n", url, err)
			m.LastStatus = EPResponseFailedDecode
		}
		if ethInfo.MembersOCount > 0 && ethInfo.MembersOCount != len(ethInfo.Members) {
			errlog.Printf("%s: Member@odata.count != Member array len\n", url)
		} else if ethInfo.OCount > 0 && ethInfo.OCount != len(ethInfo.Members) {
			errlog.Printf("%s: odata.count != Member array len\n", url)
		}
		m.ENetInterfaces.Num = len(ethInfo.Members)
		m.ENetInterfaces.OIDs = make(map[string]*EpEthInterface)

		sort.Sort(ResourceIDSlice(ethInfo.Members))
		for i, eOID := range ethInfo.Members {
			eID := eOID.Basename()
			m.ENetInterfaces.OIDs[eID] = NewEpEthInterface(m.epRF, m.OdataID, m.RedfishType, eOID, i)
		}
		m.ENetInterfaces.discoverRemotePhase1()
	}
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(m, "", "   ")
		errlog.Printf("%s: %s\n", topURL, jout)
	}
	m.LastStatus = VerifyingData
}

// To be completed when all Manager info has been retrieved via Redfish
// from the parent endpoint.  Establishes HMS data to provide higher-level
// data structures that can be integrated into a complete view of the
// system.
func (ms *EpManagers) discoverLocalPhase2() error {
	var savedError error
	for i, m := range ms.OIDs {
		m.discoverLocalPhase2()
		if m.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF ManagerType not supported: %s",
				i, m.RedfishSubtype)
		} else if m.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, m.LastStatus)
			errlog.Printf("Managers: discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// This is the second discovery phase, after all information from
// the parent endpoint has been gathered.  This is not really intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the endpoint first, so that it is available during later steps.
func (m *EpManager) discoverLocalPhase2() {
	// Should never happen
	if m.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for system odataID: %s\n",
			m.OdataID)
		m.LastStatus = EndpointInvalid
		return
	}
	if m.LastStatus != VerifyingData {
		return
	}
	m.Ordinal = m.epRF.getManagerOrdinal(m)
	m.Type = m.epRF.getManagerHMSType(m)
	if m.Type == base.HMSTypeInvalid.String() {
		m.LastStatus = RedfishSubtypeNoSupport
		return
	}
	m.ID = m.epRF.getManagerHMSID(m, m.Type, m.Ordinal)
	if m.ID == "" {
		m.LastStatus = RedfishSubtypeNoSupport
		return
	}
	generatedFRUID, err := GetManagerFRUID(m)
	if err != nil {
		errlog.Printf("FRUID Error: %s\n", err.Error())
		errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
	}
	m.FRUID = generatedFRUID
	m.Name = m.ManagerRF.Name

	// Sets Manager ComponentEndpoint MACAddress and EthernetNICInfo entries.
	m.discoverComponentEPEthInterfaces()

	// Sets up HMS state fields using Status/State/Health info from Redfish
	m.discoverComponentState()

	// TODO: actually discover these
	m.Arch = base.ArchX86.String()
	m.NetType = base.NetSling.String()

	// Check if we have something valid to insert into the data store
	hmsType := base.GetHMSType(m.ID)
	if !base.IsHMSTypeController(hmsType) || m.Type != hmsType.String() {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s') for %s\n",
			m.ID, m.Type, m.ManagerURL)
		m.LastStatus = VerificationFailed
		return
	}
	m.LastStatus = DiscoverOK

}

// Sets Manager ComponentEndpoint MACAddress and EthernetNICInfo entries.
func (m *EpManager) discoverComponentEPEthInterfaces() {
	// Provide a brief summary of all attached ethernet interfaces
	m.EthNICInfo = make([]*EthernetNICInfo, 0, 1)
	enabledIfaceMAC := ""
	enabledStateMAC := ""
	enabledOKStateMAC := ""
	for _, e := range m.ENetInterfaces.OIDs {
		ethIDAddr := new(EthernetNICInfo)

		// Create the summary for this interface.
		ethIDAddr.RedfishId = e.BaseOdataID
		ethIDAddr.Oid = e.EtherIfaceRF.Oid
		ethIDAddr.Description = e.EtherIfaceRF.Description
		ethIDAddr.FQDN = e.EtherIfaceRF.FQDN
		ethIDAddr.Hostname = e.EtherIfaceRF.Hostname
		ethIDAddr.InterfaceEnabled = e.EtherIfaceRF.InterfaceEnabled
		ethIDAddr.MACAddress = NormalizeMAC(e.EtherIfaceRF.MACAddress)
		ethIDAddr.PermanentMACAddress = NormalizeMAC(
			e.EtherIfaceRF.PermanentMACAddress,
		)

		m.EthNICInfo = append(m.EthNICInfo, ethIDAddr)

		// Find MAC.  Should be enabled.
		if ethIDAddr.InterfaceEnabled != nil {
			if *ethIDAddr.InterfaceEnabled == true {
				enabledIfaceMAC = ethIDAddr.MACAddress
			}
		}
		if e.EtherIfaceRF.Status.State == "Enabled" {
			enabledStateMAC = ethIDAddr.MACAddress
			if e.EtherIfaceRF.Status.Health == "OK" {
				enabledOKStateMAC = ethIDAddr.MACAddress
			}
		}
		// If interfaceEnabled flag is set and state is enabled then that's
		// the best choice.  If not relax on the health == OK part.
		if enabledIfaceMAC == enabledOKStateMAC {
			m.MACAddr = enabledIfaceMAC
		} else if m.MACAddr == "" {
			if enabledIfaceMAC == enabledStateMAC {
				m.MACAddr = enabledIfaceMAC
			}
		}
	}
	// Fallback find MAC if we didn't find an enabled interface with matching
	// state/health.
	if m.MACAddr == "" {
		if len(m.ENetInterfaces.OIDs) == 1 {
			// if only one NIC, obviously that's it.
			m.MACAddr = NormalizeMAC(m.EthNICInfo[0].MACAddress)
		} else if enabledIfaceMAC != "" {
			// Otherwise, take the last interfaceEnabled one we saw
			m.MACAddr = enabledIfaceMAC
		} else if enabledOKStateMAC != "" {
			// If none enabled, take the last one with state = enabled/OK
			m.MACAddr = enabledOKStateMAC
		} else if enabledStateMAC != "" {
			// Otherwise health != OK but state == enabled.
			m.MACAddr = enabledStateMAC
		}
	}
}

// Sets up HMS state fields using Status/State/Health info from Redfish
func (m *EpManager) discoverComponentState() {
	if m.ManagerRF.Status.State != "Absent" {
		m.Status = "Populated"
		m.State = base.StatePopulated.String()
		m.Flag = base.FlagOK.String()
		// Check power state - if no info leave as populated unless it's us.
		if m.ManagerRF.PowerState != "" {
			if m.ManagerRF.PowerState == POWER_STATE_OFF ||
				m.ManagerRF.PowerState == POWER_STATE_POWERING_ON {
				m.State = base.StateOff.String()
			} else if m.ManagerRF.PowerState == POWER_STATE_ON ||
				m.ManagerRF.PowerState == POWER_STATE_POWERING_OFF {
				m.State = base.StateOn.String()
			}
		}
		if m.ID == m.epRF.ID {
			// Even if info is missing, if we are the manager for the
			// RedfishEndpoint, we must be working.
			m.State = base.StateReady.String()
		}
		// Set Flags. Don't set flags on components without tracked state
		if m.State != base.StatePopulated.String() {
			if m.ManagerRF.Status.State != "" {
				if m.ManagerRF.Status.State == "Enabled" {
					if m.State == base.StateOn.String() &&
						(m.ManagerRF.Status.Health == "OK" ||
							m.ManagerRF.Status.Health == "") {
						// If not the manager we're talking to, but powered
						// and enabled and OK, then set ready also, since we
						// can also manage this manager via this endpoint.
						m.State = base.StateReady.String()
					}
				} else {
					// Non-enabled RF state, but power state is valid, set warn
					m.Flag = base.FlagWarning.String()
				}
			}
			// Also pass along Redfish health as a flag.  Even if we
			// know the controller is alive something may be up.
			if m.ManagerRF.Status.Health == "Warning" {
				m.Flag = base.FlagWarning.String()
			} else if m.ManagerRF.Status.Health == "Critical" {
				m.Flag = base.FlagAlert.String()
			}
		}
		// Non-empty, set FRU ID
		generatedFRUID, err := GetManagerFRUID(m)
		if err != nil {
			errlog.Printf("FRUID Error: %s\n", err.Error())
			errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
		}
		m.FRUID = generatedFRUID
	} else {
		// Empty - No FRU ID
		m.Status = "Empty"
		m.State = base.StateEmpty.String()
		//the state of the component is known (empty), it is not locked, does not have an alert or warning, so therefore Flag defaults to OK.
		m.Flag = base.FlagOK.String()
	}
}

/////////////////////////////////////////////////////////////////////////////
//
// Redfish ComputerSystem (Node) component discovery
//
/////////////////////////////////////////////////////////////////////////////

// These are the discovered attributes for a particular entry in the Systems
// URL for a given RF endpoint.  They will not be usable by themselves unless
// the hostname matching the host IP is available via Redfish.
// In the future, there will be a way to map "canonical" hostnames to each
// system under an endpoint via a file, and/or generate them programatically
// according to some method.
type EpSystem struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	// Embedded struct
	ComponentSystemInfo

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	SystemURL   string `json:"systemURL"` // Full URL to this ComputerSystem
	BaseOdataID string `json:"BaseOdataID"`

	PChassisOID  string `json:"ParentOID"`    // odata.id for parent chassis
	PChassisType string `json:"PChassisType"` // ChassisType enum from parent
	ManagerOID   string `json:"ManagerOID"`   // Most immediate manager if mult.
	ManagerType  string `json:"ManagerType"`  // Almost always BMC or equiv.

	LastStatus string `json:"lastStatus"`

	// These values are used as keys to locate EpManager and EpChassis
	// components via the top-level maps for each RF endpoint, via the epRF
	// endpoint, since these a Chassisre top-level groups just like "systems" are.
	//
	// Note that for simple systems, we are likely to only have the
	// containing (e.g. server) chassis and BMC listed here, which we will
	// use to populate the PChassisOID and ManagerOID fields above, but we
	// save the whole list for completeness (as we create a struct for
	// every chassis listed at the root) as we may need to support more
	// complicated structures later on.
	ManagedBy     []ResourceID `json:"managedBy"`
	ChassisForSys []ResourceID `json:"chassisForSys"`

	SystemRF  ComputerSystem   `json:"ComputerSystemRF"`
	sysURLRaw *json.RawMessage //`json:"sysURLRaw"`

	// Ethernet Interfaces are children of Managers or Systems, not top
	// level components like Chassis or Managers.  Therefore we put the
	// local collection here.  Chassis or Managers are consolodated at the
	// EpEndpoint level in maps and we will use the odataID recorded here to
	// reference these via the epRF pointer.
	ENetInterfaces EpEthInterfaces `json:"enetInterfaces"`

	// Assembly and NodeAccelRiser info comes from the Chassis level but we
	// associate it with nodes (systems) so we record it here.
	Assembly        *EpAssembly       `json:"Assembly"`
	NodeAccelRisers EpNodeAccelRisers `json:"NodeAccelRisers"`
	
	// HpeDevice info comes from the Chassis level HPE OEM Links but we
	// associate it with nodes (systems) so we record it here. We discover
	// GPUs on HPE hardware as an HpeDevice.
	HpeDevices EpHpeDevices `json:"HpeDevices"`

	// NetworkAdapter (HSN NIC) info comes from the Chassis level but we
	// associate it with nodes (systems) so we record it here.
	NetworkAdapters EpNetworkAdapters `json:"NetworkAdapters"`

	// Power info comes from the chassis level but we associate it with
	// nodes (systems) so we record it here.
	PowerInfo PowerInfo `json:"powerInfo"`

	// Processors and Memory are similar to EthernetInterfaces, children
	// only of a particular ComputerSystem
	Processors EpProcessors `json:"processors"`
	MemoryMods EpMemoryMods `json:"memoryMods"`

	cpuCount   int
	accelCount int

	StorageGroups EpStorageCollections `json:"storageGroups"`
	Drives        EpDrives             `json:"drives"`

	epRF *RedfishEP // Backpointer, for connection details, Chassis maps, etc.
}

// Set of EpSystem, representing a Redfish "ComputerSystem" listed in some RF
// endpoint's "Systems" collection.
type EpSystems struct {
	Num  int                  `json:"num"`
	OIDs map[string]*EpSystem `json:"ids"`
}

// Initializes EpSystem struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the system to look at.
// This should be the only way this struct is created.
func NewEpSystem(epRF *RedfishEP, odataID ResourceID, rawOrdinal int) *EpSystem {
	s := new(EpSystem)
	s.Type = base.Node.String()
	s.OdataID = odataID.Oid
	s.BaseOdataID = odataID.Basename()
	s.RedfishType = ComputerSystemType
	s.RfEndpointID = epRF.ID
	s.LastStatus = NotYetQueried
	s.Ordinal = -1
	s.RawOrdinal = rawOrdinal
	s.epRF = epRF
	s.cpuCount = 0
	s.accelCount = 0
	return s
}

// Makes contact with remote endpoint to discover basic information about
// this EpSystem (i.e. a particular Redfish "System", i.e. node).  This is
// the first step once the struct is initialized.  After this discovery
// is complete for all components, the second phase is done.  The second phase
// doesn't obtain new data but rather explores data from phase1 further, and
// should be done after all components under a particular RF entry point
// have completed Phase 1.
func (ss *EpSystems) discoverRemotePhase1() {
	for _, s := range ss.OIDs {
		s.discoverRemotePhase1()
	}
}

// Makes contact with remote endpoint to discover information about
// the calling struct (a Redfish ComputerSystem i.e. node)
func (s *EpSystem) discoverRemotePhase1() {
	// Should never happen
	if s.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for system odataID: %s\n",
			s.OdataID)
		s.LastStatus = EndpointInvalid
		return
	}
	s.SystemURL = s.epRF.FQDN + s.OdataID

	path := s.OdataID
	url := s.epRF.FQDN + path
	topURL := url

	// If the RF health status is 'Starting' then the hwinv info might
	// still be loading from BIOS. Keep checking. This could take ~45
	// seconds. We'll stop waiting after 60 seconds.
	for retry := 0; retry <= 6; retry++ {
		if retry != 0 {
			time.Sleep(10 * time.Second)
		}
		systemURLJSON, err := s.epRF.GETRelative(path)
		if err != nil || systemURLJSON == nil {
			s.LastStatus = HTTPsGetFailed
			return
		}
		s.sysURLRaw = &systemURLJSON
		s.LastStatus = HTTPsGetOk

		// Decode JSON into ComputerSystem structure
		if err := json.Unmarshal(*s.sysURLRaw, &s.SystemRF); err != nil {
			if IsUnmarshalTypeError(err) {
				errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
			} else {
				errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
				s.LastStatus = EPResponseFailedDecode
				return
			}
		}

		// Keep checking for info to be done loading from BIOS
		if s.SystemRF.Status.State == "Starting" {
			errlog.Printf("%s: System status is 'Starting'. Retrying... \n", url)
			continue
		}
		if len(s.SystemRF.ProcessorSummary.Count) == 0 ||
			s.SystemRF.ProcessorSummary.Count == "0" {
			errlog.Printf("%s: System ProcessorSummary is not populated. Retrying... \n", url)
			continue
		}
		if len(s.SystemRF.MemorySummary.TotalSystemMemoryGiB) == 0 ||
			s.SystemRF.MemorySummary.TotalSystemMemoryGiB == "0" {
			errlog.Printf("%s: System MemorySummary is not populated. Retrying... \n", url)
			continue
		}
		break
	}
	s.RedfishSubtype = s.SystemRF.SystemType
	s.UUID = s.SystemRF.UUID
	s.ManagedBy = s.SystemRF.Links.ManagedBy
	s.ChassisForSys = s.SystemRF.Links.Chassis
	// The format of the Actions field of the ComputerSystem Redfish response
	// has changed in the AMI Redfish implementation. Both the Mountain and
	// Gigabyte nodes use this new Action field.
	// Old field:
	//		"Actions" : {
	//			"#ComputerSystem.Reset" : {
	//				"target" : "/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset",
	//				"ResetType@Redfish.AllowableValues" : [
	//					"On",
	//					"ForceOff",
	//					"Off"
	//					]
	//			}
	//		},
	// New fields:
	//		"Actions" : {
	//			"#ComputerSystem.Reset" : {
	//				"target" : "/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset",
	//				"@Redfish.ActionInfo" : "/redfish/v1/Systems/Self/ResetActionInfo"
	//			}
	//		}
	//		/redfish/v1/Systems/Node0/ResetActionInfo:
	//		{
	//			"Id" : "ResetAction",
	//			"Parameters" : [
	//				{
	//					"Name" : "ResetType",
	//					"AllowableValues" : [
	//						"On",
	//						"ForceOff",
	//						"Off"
	//					],
	//					"DataType" : "String",
	//					"Required" : true
	//				}
	//			],
	//			"@odata.id" : "/redfish/v1/Systems/Self/ResetActionInfo",
	//			"Name" : "ResetAction",
	//			"@odata.type" : "#ActionInfo.v1_0_3.ActionInfo",
	//			"Description" : "This action is used to reset the Systems",
	//			"@odata.etag" : "W/\"1528\"",
	//			"@odata.context" : "/redfish/v1/$metadata#ActionInfo.ActionInfo"
	//		}
	// Another level of indirection is needed to gather the AllowableValues for
	// the RestType. This is done by gathering the information at
	// @Redfish.ActionInfo from the Redfish endpoint and then putting it in the
	// existing structure layout.
	if s.SystemRF.Actions != nil {
		s.Actions = s.SystemRF.Actions
		csr := s.Actions.ComputerSystemReset
		if csr.RFActionInfo != "" {
			actionInfoJSON, err := s.epRF.GETRelative(csr.RFActionInfo)
			if err != nil || actionInfoJSON == nil {
				s.LastStatus = HTTPsGetFailed
				return
			}
			var actionInfo ResetActionInfo
			err = json.Unmarshal(actionInfoJSON, &actionInfo)
			if err != nil {
				errlog.Printf("Failed to decode %s: %s\n", url, err)
				s.LastStatus = EPResponseFailedDecode
			}
			for _, p := range actionInfo.RAParameters {
				if p.Name == "ResetType" {
					s.Actions.ComputerSystemReset.AllowableValues = p.AllowableValues
				}
			}
		}
	}

	//
	// Get Chassis-level info associated with the system (node)
	//
	// Some info (Power, NodeAccelRiser, HSN NIC, etc) is at the chassis level
	// but we associate it with nodes (systems). There will be a chassis URL
	// with our system's id if there is info to get.
	nodeChassis, ok := s.epRF.Chassis.OIDs[s.SystemRF.Id]
	if ok {

		//
		// Get PowerControl Info if it exists
		//
		if nodeChassis.ChassisRF.Controls.Oid != "" {
			path = nodeChassis.ChassisRF.Controls.Oid
			ctlURLJSON, err := s.epRF.GETRelative(path)
			if err != nil || ctlURLJSON == nil {
				s.LastStatus = HTTPsGetFailed
				return
			}
			s.LastStatus = HTTPsGetOk

			// Decode JSON into PowerControl structure
			var controlCollection ControlCollection
			if err := json.Unmarshal(ctlURLJSON, &controlCollection); err != nil {
				if IsUnmarshalTypeError(err) {
					errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
				} else {
					errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
					s.LastStatus = EPResponseFailedDecode
					return
				}
			}
			for _, url := range controlCollection.Members {
				if url.Oid == "" {
					continue
				}
				controlJSON, err := s.epRF.GETRelative(url.Oid)
				if err != nil || controlJSON == nil {
					break
				}
				// Decode JSON into PowerControl structure
				var rfControl RFControl
				if err := json.Unmarshal(controlJSON, &rfControl); err != nil {
					if IsUnmarshalTypeError(err) {
						errlog.Printf("bad field(s) skipped: %s: %s\n", url.Oid, err)
					} else {
						errlog.Printf("ERROR: json decode failed: %s: %s\n", url.Oid, err)
						break
					}
				}
				control := Control{
					URL: url.Oid,
					Control: rfControl,
				}
				s.Controls = append(s.Controls, &control)
			}
		}
		if nodeChassis.ChassisRF.Power.Oid != "" {
			path = nodeChassis.ChassisRF.Power.Oid
			pwrCtlURLJSON, err := s.epRF.GETRelative(path)
			if err != nil || pwrCtlURLJSON == nil {
				s.LastStatus = HTTPsGetFailed
				return
			}
			s.PowerURL = path
			s.LastStatus = HTTPsGetOk

			// Decode JSON into PowerControl structure
			if err := json.Unmarshal(pwrCtlURLJSON, &s.PowerInfo); err != nil {
				if IsUnmarshalTypeError(err) {
					errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
				} else {
					errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
					s.LastStatus = EPResponseFailedDecode
					return
				}
			}
			if s.PowerInfo.OEM != nil && s.PowerInfo.OEM.HPE != nil && len(s.PowerInfo.PowerControl) > 0 {
				oemPwr := PwrCtlOEM{HPE: &PwrCtlOEMHPE{
					Status: "Empty",
				}}
				for {
					if s.PowerInfo.OEM.HPE.Links.AccPowerService.Oid == "" {
						break
					}
					
					path = s.PowerInfo.OEM.HPE.Links.AccPowerService.Oid
					hpeAccPowerServiceJSON, err := s.epRF.GETRelative(path)
					if err != nil || hpeAccPowerServiceJSON == nil {
						if err == ErrRFDiscILOLicenseReq {
							oemPwr.HPE.Status = "LicenseNeeded"
						}
						break
					}
					// Decode JSON into PowerControl structure
					var hpeAccPowerService HPEAccPowerService
					if err := json.Unmarshal(hpeAccPowerServiceJSON, &hpeAccPowerService); err != nil {
						if IsUnmarshalTypeError(err) {
							errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
						} else {
							errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
							break
						}
					}
					if hpeAccPowerService.Links.PowerLimit.Oid == "" {
						break
					}
					path = hpeAccPowerService.Links.PowerLimit.Oid
					hpePowerLimitJSON, err := s.epRF.GETRelative(path)
					if err != nil || hpePowerLimitJSON == nil {
						if err == ErrRFDiscILOLicenseReq {
							oemPwr.HPE.Status = "LicenseNeeded"
						}
						break
					}
					// Decode JSON into PowerControl structure
					var hpePowerLimit HPEPowerLimit
					if err := json.Unmarshal(hpePowerLimitJSON, &hpePowerLimit); err != nil {
						if IsUnmarshalTypeError(err) {
							errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
						} else {
							errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
							break
						}
					}
					oemPwr.HPE.PowerLimit.Min = hpePowerLimit.PowerLimitRanges[0].MinimumPowerLimit
					oemPwr.HPE.PowerLimit.Max = hpePowerLimit.PowerLimitRanges[0].MaximumPowerLimit
					oemPwr.HPE.Target = hpePowerLimit.Actions.ConfigurePowerLimit.Target
					oemPwr.HPE.Status = "OK"
					oemPwr.HPE.PowerRegulationEnabled = hpeAccPowerService.PowerRegulationEnabled
					s.PowerURL = hpeAccPowerService.Links.PowerLimit.Oid
					s.PowerInfo.PowerControl[0].Name = hpePowerLimit.Name
					break
				}
				s.PowerInfo.PowerControl[0].OEM = &oemPwr
			}
			s.PowerCtl = s.PowerInfo.PowerControl
		}

		//
		// Get Chassis assembly (NodeAccelRiser) info if it exists
		//
		if nodeChassis.ChassisRF.Assembly.Oid == "" {
			//errlog.Printf("%s: No assembly obj found.\n", topURL)
			s.NodeAccelRisers.Num = 0
			s.NodeAccelRisers.OIDs = make(map[string]*EpNodeAccelRiser)
		} else {
			//create a new EpAssembly object using chassis and Assembly.OID
			s.Assembly = NewEpAssembly(s, nodeChassis.ChassisRF.Assembly, nodeChassis.OdataID, nodeChassis.RedfishType)

			//retrieve the Assembly RF
			s.Assembly.discoverRemotePhase1()

			//discover any NodeAccelRiser cards
			if len(s.Assembly.AssemblyRF.Assemblies) > 0 {
				s.NodeAccelRisers.OIDs = make(map[string]*EpNodeAccelRiser)
				indexNodeAccelRisersOIDs := 0
				for i, assembly := range s.Assembly.AssemblyRF.Assemblies {
					//Need to ignore any assembly that is not a GPUSubsystem
					if assembly.PhysicalContext == NodeAccelRiserType {
						rID := assembly.Oid
						s.NodeAccelRisers.OIDs[rID] = NewEpNodeAccelRiser(s.Assembly, ResourceID{rID}, i)
						indexNodeAccelRisersOIDs++
					}
				}
				//have to set the Num of NodeAccelRisers only after iterating across the Assemblies array
				//and identifying each individual NodeAccelRiser
				s.NodeAccelRisers.Num = len(s.NodeAccelRisers.OIDs)
				//invoke the series of discoverRemotePhase1 calls for each NodeAccelRiser
				s.NodeAccelRisers.discoverRemotePhase1()
			}
		}

		//
		// Get Chassis NetworkAdapter (HSN NIC) info if it exists
		//
		if nodeChassis.ChassisRF.NetworkAdapters.Oid == "" {
			//errlog.Printf("%s: No assembly obj found.\n", topURL)
			s.NetworkAdapters.Num = 0
			s.NetworkAdapters.OIDs = make(map[string]*EpNetworkAdapter)
		} else {
			path = nodeChassis.ChassisRF.NetworkAdapters.Oid
			url = nodeChassis.epRF.FQDN + path
			naJSON, err := s.epRF.GETRelative(path)
			if err != nil || naJSON == nil {
				s.LastStatus = HTTPsGetFailed
				return
			}
			if rfDebug > 0 {
				errlog.Printf("%s: %s\n", url, naJSON)
			}
			s.LastStatus = HTTPsGetOk

			var naInfo NetworkAdapterCollection
			if err := json.Unmarshal(naJSON, &naInfo); err != nil {
				errlog.Printf("Failed to decode %s: %s\n", url, err)
				s.LastStatus = EPResponseFailedDecode
			}

			s.NetworkAdapters.Num = len(naInfo.Members)
			s.NetworkAdapters.OIDs = make(map[string]*EpNetworkAdapter)

			sort.Sort(ResourceIDSlice(naInfo.Members))
			for i, naoid := range naInfo.Members {
				naid := naoid.Basename()
				s.NetworkAdapters.OIDs[naid] = NewEpNetworkAdapter(s, s.OdataID, s.RedfishType, naoid, i)
			}
			s.NetworkAdapters.discoverRemotePhase1()
		}

		// Discover HPE devices to find GPUs
		if strings.ToLower(s.SystemRF.Manufacturer) == "hpe" &&
			nodeChassis.ChassisRF.OEM != nil &&
			nodeChassis.ChassisRF.OEM.Hpe != nil &&
			nodeChassis.ChassisRF.OEM.Hpe.Links.Devices.Oid != "" {
			path = nodeChassis.ChassisRF.OEM.Hpe.Links.Devices.Oid
			url = s.epRF.FQDN + path
			devicesJSON, err := s.epRF.GETRelative(path)
			if err != nil || devicesJSON == nil {
				s.LastStatus = HTTPsGetFailed
				return
			}
			if rfDebug > 0 {
				errlog.Printf("%s: %s\n", url, devicesJSON)
			}
			s.LastStatus = HTTPsGetOk

			var deviceInfo HpeDeviceCollection
			if err := json.Unmarshal(devicesJSON, &deviceInfo); err != nil {
				errlog.Printf("Failed to decode %s: %s\n", url, err)
				s.LastStatus = EPResponseFailedDecode
			}

			s.HpeDevices.Num = len(deviceInfo.Members)
			s.HpeDevices.OIDs = make(map[string]*EpHpeDevice)

			sort.Sort(ResourceIDSlice(deviceInfo.Members))
			for deviceOrd, dOID := range deviceInfo.Members {
				dID := dOID.Basename()
				s.HpeDevices.OIDs[dID] = NewEpHpeDevice(s, dOID, nodeChassis.OdataID, nodeChassis.RedfishType, deviceOrd)
			}
			s.HpeDevices.discoverRemotePhase1()
		} else {
			s.HpeDevices.Num = 0
			s.HpeDevices.OIDs = make(map[string]*EpHpeDevice)
		}
	}

	//
	// Get link to systems's ethernet interfaces
	//

	if s.SystemRF.EthernetInterfaces.Oid == "" {
		// TODO: Just try default path?
		errlog.Printf("%s: No EthernetInterfaces found.\n", url)
		s.ENetInterfaces.Num = 0
		s.ENetInterfaces.OIDs = make(map[string]*EpEthInterface)
	} else {
		path = s.SystemRF.EthernetInterfaces.Oid
		url = s.epRF.FQDN + path
		ethIfacesJSON, err := s.epRF.GETRelative(path)
		if err != nil || ethIfacesJSON == nil {
			s.LastStatus = HTTPsGetFailed
			return
		}
		if rfDebug > 0 {
			errlog.Printf("%s: %s\n", url, ethIfacesJSON)
		}
		s.LastStatus = HTTPsGetOk

		var ethInfo EthernetInterfaceCollection
		if err := json.Unmarshal(ethIfacesJSON, &ethInfo); err != nil {
			errlog.Printf("Failed to decode %s: %s\n", url, err)
			s.LastStatus = EPResponseFailedDecode
		}

		// The count is typically given as "Members@odata.count", but
		// older versions drop the "Members" identifier
		if ethInfo.MembersOCount > 0 && ethInfo.MembersOCount != len(ethInfo.Members) {
			errlog.Printf("%s: Member@odata.count != Member array len\n", url)
		} else if ethInfo.OCount > 0 && ethInfo.OCount != len(ethInfo.Members) {
			errlog.Printf("%s: odata.count != Member array len\n", url)
		}
		s.ENetInterfaces.Num = len(ethInfo.Members)
		s.ENetInterfaces.OIDs = make(map[string]*EpEthInterface)

		sort.Sort(ResourceIDSlice(ethInfo.Members))
		for i, eoid := range ethInfo.Members {
			eid := eoid.Basename()
			s.ENetInterfaces.OIDs[eid] = NewEpEthInterface(s.epRF, s.OdataID, s.RedfishType, eoid, i)
		}
		s.ENetInterfaces.discoverRemotePhase1()
	}

	//
	// Get link to systems's ProcessorCollection
	//

	if s.SystemRF.Processors.Oid == "" {
		errlog.Printf("%s: No ProcessorCollection found.\n", topURL)
		s.Processors.Num = 0
		s.Processors.OIDs = make(map[string]*EpProcessor)
	} else {
		path = s.SystemRF.Processors.Oid
		url = s.epRF.FQDN + path
		processorsJSON, err := s.epRF.GETRelative(path)
		if err != nil || processorsJSON == nil {
			s.LastStatus = HTTPsGetFailed
			return
		}
		if rfDebug > 0 {
			errlog.Printf("%s: %s\n", url, processorsJSON)
		}
		s.LastStatus = HTTPsGetOk

		var procInfo ProcessorCollection
		if err := json.Unmarshal(processorsJSON, &procInfo); err != nil {
			errlog.Printf("Failed to decode %s: %s\n", url, err)
			s.LastStatus = EPResponseFailedDecode
		}

		// The count is typically given as "Members@odata.count", but
		// older versions drop the "Members" identifier
		if procInfo.MembersOCount > 0 && procInfo.MembersOCount != len(procInfo.Members) {
			errlog.Printf("%s: Member@odata.count != Member array len\n", url)
		} else if procInfo.OCount > 0 && procInfo.OCount != len(procInfo.Members) {
			errlog.Printf("%s: odata.count != Member array len\n", url)
		}
		s.Processors.Num = len(procInfo.Members)
		s.Processors.OIDs = make(map[string]*EpProcessor)

		sort.Sort(ResourceIDSlice(procInfo.Members))
		for procOrd, pOID := range procInfo.Members {
			pID := pOID.Basename()
			// Both CPUs and GPUs show up under /redfish/v1/Systems/{systemID}/Processors
			// Need to update ordinal value of each processor based on value of its ProcessorType field (CPU or GPU)
			// in EpProcessor.discoverPhase2
			s.Processors.OIDs[pID] = NewEpProcessor(s, pOID, procOrd)
		}
		s.Processors.discoverRemotePhase1()
	}

	//
	// Get link to systems's MemoryCollection
	//

	if s.SystemRF.Memory.Oid == "" {
		errlog.Printf("%s: No MemoryCollection found.\n", topURL)
		s.MemoryMods.Num = 0
		s.MemoryMods.OIDs = make(map[string]*EpMemory)
	} else {
		path = s.SystemRF.Memory.Oid
		url = s.epRF.FQDN + path
		memoryModsJSON, err := s.epRF.GETRelative(path)
		if err != nil || memoryModsJSON == nil {
			s.LastStatus = HTTPsGetFailed
			return
		}
		if rfDebug > 0 {
			errlog.Printf("%s: %s\n", url, memoryModsJSON)
		}
		s.LastStatus = HTTPsGetOk

		var memInfo MemoryCollection
		if err := json.Unmarshal(memoryModsJSON, &memInfo); err != nil {
			errlog.Printf("Failed to decode %s: %s\n", url, err)
			s.LastStatus = EPResponseFailedDecode
		}

		// The count is typically given as "Members@odata.count", but
		// older versions drop the "Members" identifier
		if memInfo.MembersOCount > 0 && memInfo.MembersOCount != len(memInfo.Members) {
			errlog.Printf("%s: Member@odata.count != Member array len\n", url)
		} else if memInfo.OCount > 0 && memInfo.OCount != len(memInfo.Members) {
			errlog.Printf("%s: odata.count != Member array len\n", url)
		}
		s.MemoryMods.Num = len(memInfo.Members)
		s.MemoryMods.OIDs = make(map[string]*EpMemory)

		sort.Sort(ResourceIDSlice(memInfo.Members))
		for i, mOID := range memInfo.Members {
			mID := mOID.Basename()
			s.MemoryMods.OIDs[mID] = NewEpMemory(s, mOID, i)
		}
		s.MemoryMods.discoverRemotePhase1()
	}

	//
	// Get link to systems's StorageCollection
	//

	if s.SystemRF.Storage.Oid == "" {
		errlog.Printf("%s: No StorageCollection found.\n", topURL)
		s.Drives.Num = 0
		s.Drives.OIDs = make(map[string]*EpDrive)
	} else {
		path = s.SystemRF.Storage.Oid
		url = s.epRF.FQDN + path
		storageJSON, err := s.epRF.GETRelative(path)
		if err != nil || storageJSON == nil {
			s.LastStatus = HTTPsGetFailed
			return
		}
		if rfDebug > 0 {
			errlog.Printf("%s: %s\n", url, storageJSON)
		}
		s.LastStatus = HTTPsGetOk

		var storageInfo Storage
		if err := json.Unmarshal(storageJSON, &storageInfo); err != nil {
			errlog.Printf("Failed to decode %s: %s\n", url, err)
			s.LastStatus = EPResponseFailedDecode
		}

		// The count is typically given as "Members@odata.count", but
		// older versions drop the "Members" identifier
		if storageInfo.MembersOCount > 0 && storageInfo.MembersOCount != len(storageInfo.Members) {
			errlog.Printf("%s: Member@odata.count != Member array len\n", url)
		} else if storageInfo.OCount > 0 && storageInfo.OCount != len(storageInfo.Members) {
			errlog.Printf("%s: odata.count != Member array len\n", url)
		}
		//iterate across storageInfo.members and create a new EpStorageCollection
		//for each, and push it into the StorageGroups OIDs
		var epStorage = NewEpStorage(s, ResourceID{storageInfo.Oid})
		s.StorageGroups.Num = len(storageInfo.Members)
		s.StorageGroups.OIDs = make(map[string]*EpStorageCollection)
		sort.Sort(ResourceIDSlice(storageInfo.Members))
		for i, gOID := range storageInfo.Members {
			gID := gOID.Basename()
			s.StorageGroups.OIDs[gID] = NewEpStorageCollection(epStorage, gOID, i)
		}
		s.Drives.Num = 0
		s.Drives.OIDs = make(map[string]*EpDrive)
		//the s.Drives collection will also be populated after this call
		s.StorageGroups.discoverRemotePhase1()
		s.Drives.discoverRemotePhase1()
	}

	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(s, "", "   ")
		errlog.Printf("%s: %s\n", topURL, jout)
	}
	s.LastStatus = VerifyingData

}

// This is the second discovery phase, after all information from
// the parent endpoint has been gathered.  This is not really intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the endpoint first, so that it is available during later steps.
func (ss *EpSystems) discoverLocalPhase2() error {
	var savedError error
	for i, s := range ss.OIDs {
		s.discoverLocalPhase2()
		if s.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF SystemType not supported: %s",
				i, s.RedfishSubtype)
		} else if s.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, s.LastStatus)
			errlog.Printf("Systems discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual system.  Now that all information
// has been gathered, we can set the remaining fields needed to provide
// HMS with information about where the system is located.
func (s *EpSystem) discoverLocalPhase2() {
	// Should never happen
	if s.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for system odataID: %s\n",
			s.OdataID)
		s.LastStatus = EndpointInvalid
		return
	}
	if s.LastStatus != VerifyingData {
		return
	}
	// Get ordinal of Node if a physical system and not one of the
	// logical types (only support these, they are the only ones
	// that seem consistent with the HMS notion of a node, and
	// we don't know how or if the other types will be used at higher
	// levels.)
	s.Ordinal, s.Type = s.epRF.getSystemOrdinalAndType(s)
	if s.Ordinal == -1 || s.Type == "" {
		errlog.Printf("%s: Unsupported RF type '%s'",
			s.epRF.ID, s.SystemRF.SystemType)
		s.LastStatus = RedfishSubtypeNoSupport
		return
	}
	// If we got Type and Ordinal without errors, we don't expect any
	// failure due to unsupported HW here.
	s.ID = s.epRF.getSystemHMSID(s, s.Type, s.Ordinal)
	if s.ID == "" {
		errlog.Printf("%s: ERROR: Got no HMS ID string (empty)",
			s.epRF.ID)
		s.LastStatus = VerificationFailed
		return
	}
	s.Domain = s.epRF.getNodeSvcNetDomain(s)
	s.Name = s.SystemRF.Name

	s.discoverComponentEPEthInterfaces()

	s.discoverComponentState()

	// Check if we have something valid to insert into the data store
	if base.GetHMSType(s.ID) != base.Node || s.Type != base.Node.String() {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s') for %s\n",
			s.ID, s.Type, s.SystemURL)
		s.LastStatus = VerificationFailed
		return
	}
	// TODO: actually discover these
	s.Arch = base.ArchX86.String()
	s.NetType = base.NetSling.String()
	s.DefaultRole = base.RoleCompute.String()
	s.DefaultSubRole = ""
	s.DefaultClass = ""

	// Complete discovery and verify subcomponents
	var childStatus string = DiscoverOK
	if err := s.Processors.discoverLocalPhase2(); err != nil {
		childStatus = ChildVerificationFailed
	}
	if err := s.MemoryMods.discoverLocalPhase2(); err != nil {
		childStatus = ChildVerificationFailed
	}
	if err := s.Drives.discoverLocalPhase2(); err != nil {
		fmt.Printf("s.Drives.discoverLocalPhase2(): returned err %v", err)
		childStatus = ChildVerificationFailed
	}
	if err := s.NodeAccelRisers.discoverLocalPhase2(); err != nil {
		fmt.Printf("s.NodeAccelRisers.discoverLocalPhase2(): returned err %v", err)
		childStatus = ChildVerificationFailed
	}
	if err := s.NetworkAdapters.discoverLocalPhase2(); err != nil {
		fmt.Printf("s.NetworkAdapters.discoverLocalPhase2(): returned err %v", err)
		childStatus = ChildVerificationFailed
	}
	if err := s.HpeDevices.discoverLocalPhase2(); err != nil {
		fmt.Printf("s.HpeDevices.discoverLocalPhase2(): returned err %v", err)
		childStatus = ChildVerificationFailed
	}

	s.LastStatus = childStatus
}

// Sets System ComponentEndpoint MACAddress and EthernetNICInfo entries.
func (s *EpSystem) discoverComponentEPEthInterfaces() {
	// Select default interface to use as main MAC address
	ethID := s.epRF.getNodeSvcNetEthIfaceId(s)

	// Provide a brief summary of all attached ethernet interfaces
	// Also, try to chose the main node MAC address.
	s.EthNICInfo = make([]*EthernetNICInfo, 0, 1)
	for _, e := range s.ENetInterfaces.OIDs {
		ethIDAddr := new(EthernetNICInfo)
		// Create the summary for this interface.
		ethIDAddr.RedfishId = e.BaseOdataID
		ethIDAddr.Oid = e.EtherIfaceRF.Oid
		ethIDAddr.Description = e.EtherIfaceRF.Description
		ethIDAddr.FQDN = e.EtherIfaceRF.FQDN
		ethIDAddr.Hostname = e.EtherIfaceRF.Hostname
		ethIDAddr.InterfaceEnabled = e.EtherIfaceRF.InterfaceEnabled
		ethIDAddr.MACAddress = NormalizeMAC(e.EtherIfaceRF.MACAddress)
		ethIDAddr.PermanentMACAddress = NormalizeMAC(
			e.EtherIfaceRF.PermanentMACAddress,
		)
		if len(s.ENetInterfaces.OIDs) == 1 || e.BaseOdataID == ethID {
			// Assign this MAC as the main address, matches default interface
			// or is the only one.
			if ethIDAddr.PermanentMACAddress != "" {
				s.MACAddr = ethIDAddr.PermanentMACAddress
				if ethIDAddr.PermanentMACAddress != ethIDAddr.MACAddress {
					errlog.Printf("%s: %s PermanentMAC and MAC don't match.",
						s.ID, ethID)
				}
			} else {
				s.MACAddr = ethIDAddr.MACAddress
			}
		}
		s.EthNICInfo = append(s.EthNICInfo, ethIDAddr)
	}
	// No EthernetInterface objects found.  Uh oh.
	// See if we can apply a workaround for Intel s2600 boards
	// or gigabyte nodes with R05 bios.
	intelMACWorkaround := false
	gigayteMACWorkaround := false
	if len(s.ENetInterfaces.OIDs) == 0 {
		// Intel Buchanan Pass, Wolf Pass, etc.
		if strings.Contains(strings.ToLower(s.SystemRF.Model), "s2600") ||
			strings.Contains(strings.ToLower(s.SystemRF.Name), "s2600") ||
			strings.HasPrefix(strings.ToLower(s.SystemRF.Id), "qs") {

			intelMACWorkaround = true
		} else if strings.Contains(strings.ToLower(s.SystemRF.Model), "r272-z30-00") {
			// Gigabyte nodes
			gigayteMACWorkaround = true
		}
	}
	// Use s2600 workaround as we seem to have this type of board and got
	// zero node/system ethernet interfaces.  We need to get the lowest
	// MAC address amongst the BMC MAC addresses and compute an offset.
	if intelMACWorkaround || gigayteMACWorkaround {
		var mgr *EpManager = nil
		var ok bool = false
		// Get the first manager linked to in the system object
		for _, oid := range s.ManagedBy {
			mgr, ok = s.epRF.Managers.OIDs[oid.Basename()]
			if ok {
				break
			}
		}
		// If no link to ManagedBy Manager in system object, just pick
		// the first manager (there is likely only one)
		if !ok {
			for _, m := range s.epRF.Managers.OIDs {
				mgr = m
				break
			}
		}
		// Found a manager, now look at its interfaces
		if mgr != nil {
			// Don't apply workaround on gigabyte nodes if they
			// aren't the known problem firmware
			if intelMACWorkaround ||
				strings.HasPrefix(strings.ToLower(mgr.ManagerRF.FirmwareVersion), "12") {
				minMAC := ""
				for _, eth := range mgr.ENetInterfaces.OIDs {
					thisMAC := NormalizeMAC(eth.EtherIfaceRF.MACAddress)
					if minMAC == "" {
						minMAC = thisMAC
					} else {
						cmp, err := MACCompare(minMAC, thisMAC)
						if err != nil {
							errlog.Printf("MACCompare: %s <> %s: %s",
								minMAC, thisMAC, err)
						} else if cmp > 0 {
							// This MAC is numbered lower than the last min.
							minMAC = thisMAC
						}
					}
				}
				// Create two placeholder interfaces with MACs that are 2 and 1
				// less than the lowest BMC NIC's MAC
				if minMAC != "" {
					idx := 1
					for offset := -2; offset < 0; offset++ {
						adjMAC, err := GetOffsetMACString(minMAC, int64(offset))
						if err != nil {
							errlog.Printf("GetOffsetMACCompare(%s): %s",
								minMAC, err)
							continue
						}
						// Use lowest one for management interface.
						if s.MACAddr == "" {
							s.MACAddr = adjMAC
						}
						ethIDAddr := new(EthernetNICInfo)
						ethIDAddr.MACAddress = adjMAC
						ethIDAddr.Description = fmt.Sprintf(
							"Missing interface %d, MAC computed via workaround",
							idx,
						)
						s.EthNICInfo = append(s.EthNICInfo, ethIDAddr)
						idx++
					}
				}
			}
		}
	}
}

// Sets up HMS state fields using Status/State/Health info from Redfish
func (s *EpSystem) discoverComponentState() {
	if s.SystemRF.Status.State != "Absent" {
		s.Status = "Populated"
		s.State = base.StatePopulated.String()
		s.Flag = base.FlagOK.String()
		if s.SystemRF.PowerState != "" {
			if s.SystemRF.PowerState == POWER_STATE_OFF ||
				s.SystemRF.PowerState == POWER_STATE_POWERING_ON {
				s.State = base.StateOff.String()
			} else if s.SystemRF.PowerState == POWER_STATE_ON ||
				s.SystemRF.PowerState == POWER_STATE_POWERING_OFF {
				s.State = base.StateOn.String()
			}
		} else {
			if s.SystemRF.Status.Health == "OK" {
				if s.SystemRF.Status.State == "Enabled" {
					s.State = base.StateOn.String()
				}
			}
		}
		if s.SystemRF.Status.Health == "Warning" {
			s.Flag = base.FlagWarning.String()
		} else if s.SystemRF.Status.Health == "Critical" {
			s.Flag = base.FlagAlert.String()
		}
		generatedFRUID, err := GetSystemFRUID(s)
		if err != nil {
			errlog.Printf("FRUID Error: %s\n", err.Error())
			errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
		}
		s.FRUID = generatedFRUID
	} else {
		s.Status = "Empty"
		s.State = base.StateEmpty.String()
		//the state of the component is known (empty), it is not locked, does not have an alert or warning, so therefore Flag defaults to OK.
		s.Flag = base.FlagOK.String()
	}
}

/////////////////////////////////////////////////////////////////////////////
// ComputerSystem - Ethernet Interfaces
/////////////////////////////////////////////////////////////////////////////

// This is one of possibly several ethernet interfaces for a a particular
// EpSystem(Redfish "ComputerSystem" or just "System"), or for a particular
// manager, i.e. BMC.
type EpEthInterface struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	BaseOdataID string `json:"BaseOdataID"`

	EtherURL   string `json:"etherURL"`   // Full URL to this  EthernetInterface
	ParentOID  string `json:"parentOID"`  // odata.id for parent
	ParentType string `json:"parentType"` // ComputerSystem or Manager

	Hostname string `json:"hostname"`
	Domain   string `json:"domain"`
	MACAddr  string `json:"MACAddr"`

	LastStatus string `json:"LastStatus"`

	EtherIfaceRF  EthernetInterface `json:"EtherIfaceRF"`
	etherIfaceRaw *json.RawMessage  //`json:"etherIfaceRaw"`

	epRF *RedfishEP // Backpointer, for connection details, etc.
}

// Set of EpEthInterface, each representing a Redfish "EthernetInterface"
// listed under some Redfish "System" or "Manager".  Unlike Systems or Chassis,
// there is no top-level listing of all interfaces under an RF endpoint, i.e.
// each set is specific to a single parent System or Manager.
type EpEthInterfaces struct {
	Num  int                        `json:"num"`
	OIDs map[string]*EpEthInterface `json:"oids"`
}

// Initializes EpEthInterface struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the EthernetInterface to
// look at.  This should be the only way this struct is created for
// interfaces under a system.  TODO: constructor for manager if needed.
func NewEpEthInterface(e *RedfishEP, pOID, pType string, odataID ResourceID, rawOrdinal int) *EpEthInterface {
	ei := new(EpEthInterface)
	ei.OdataID = odataID.Oid
	ei.Type = base.HMSTypeInvalid.String() // Not used in inventory/state-tracking
	ei.BaseOdataID = odataID.Basename()
	ei.RedfishType = EthernetInterfaceType
	ei.RfEndpointID = e.ID
	ei.EtherURL = e.FQDN + odataID.Oid
	ei.ParentOID = pOID
	ei.ParentType = pType
	ei.LastStatus = NotYetQueried
	ei.Ordinal = -1
	ei.RawOrdinal = rawOrdinal
	ei.epRF = e

	return ei
}

// Makes contact with redfish endpoint to discover information about
// all EthernetInterfaces for a system or manager.  EpEthInterface entries
// should be created with the appropriate constructor first.
func (es *EpEthInterfaces) discoverRemotePhase1() {
	for _, ei := range es.OIDs {
		ei.discoverRemotePhase1()
	}
}

// Makes contact with redfish endpoint to discover information about
// a particular EthernetInterface for a system or manager.   Note that the
// EpEthInterface should be created with the appropriate constructor first.
func (ei *EpEthInterface) discoverRemotePhase1() {
	rpath := ei.OdataID
	url := ei.epRF.FQDN + rpath
	etherURLJSON, err := ei.epRF.GETRelative(rpath)
	if err != nil || etherURLJSON == nil {
		ei.LastStatus = HTTPsGetFailed
		return
	}
	ei.etherIfaceRaw = &etherURLJSON
	ei.LastStatus = HTTPsGetOk

	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, etherURLJSON)
	}
	// Decode JSON into EthernetInterface structure
	if err := json.Unmarshal(etherURLJSON, &ei.EtherIfaceRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			ei.LastStatus = EPResponseFailedDecode
			return
		}
	}
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(ei, "", "   ")
		errlog.Printf("%s: %s\n", url, jout)
	}
	ei.LastStatus = VerifyingData
}

/////////////////////////////////////////////////////////////////////////////
// ComputerSystem - Processors
/////////////////////////////////////////////////////////////////////////////

// This is one of possibly several processors for a a particular
// EpSystem(Redfish "ComputerSystem" or just "System").
type EpProcessor struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	ProcessorURL string `json:"processorURL"` // Full URL to this RF Processor obj
	ParentOID    string `json:"parentOID"`    // odata.id for parent
	ParentType   string `json:"parentType"`   // ComputerSystem or Manager
	LastStatus   string `json:"LastStatus"`

	ProcessorRF  Processor        `json:"ProcessorRF"`
	processorRaw *json.RawMessage //`json:"processorRaw"`

	epRF  *RedfishEP // Backpointer to RF EP, for connection details, etc.
	sysRF *EpSystem  // Backpointer to parent system.
}

// Set of EpProcessor, each representing a Redfish "Processor"
// listed under a Redfish ComputerSystem aka System.  Unlike Systems/Chassis,
// there is no top-level listing of all processors under an RF endpoint, i.e.
// each one is specific to a collection belonging to a single parent System.
type EpProcessors struct {
	Num  int                     `json:"num"`
	OIDs map[string]*EpProcessor `json:"oids"`
}

// Initializes EpProcesor struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the Processor to
// look at.  This should be the only way this struct is created for
// processors under a system.
func NewEpProcessor(s *EpSystem, odataID ResourceID, rawOrdinal int) *EpProcessor {
	p := new(EpProcessor)
	p.OdataID = odataID.Oid
	p.Type = base.Processor.String()
	p.BaseOdataID = odataID.Basename()
	p.RedfishType = ProcessorType
	p.RfEndpointID = s.epRF.ID

	p.ProcessorURL = s.epRF.FQDN + odataID.Oid

	p.ParentOID = s.OdataID
	p.ParentType = ComputerSystemType

	p.Ordinal = -1
	p.RawOrdinal = rawOrdinal

	p.LastStatus = NotYetQueried
	p.epRF = s.epRF
	p.sysRF = s

	return p
}

// Makes contact with redfish endpoint to discover information about
// all Processors for a given Redfish System.  EpProcessor entries
// should be created with the appropriate constructor first.
func (ps *EpProcessors) discoverRemotePhase1() {
	for _, p := range ps.OIDs {
		p.discoverRemotePhase1()
	}
}

// Makes contact with redfish endpoint to discover information about
// a particular processor under a ComputerSystem aka System.   Note that the
// EpProcessor should be created with the appropriate constructor first.
func (p *EpProcessor) discoverRemotePhase1() {
	rpath := p.OdataID
	url := p.epRF.FQDN + rpath
	urlJSON, err := p.epRF.GETRelative(rpath)
	if err != nil || urlJSON == nil {
		if err == ErrRFDiscURLNotFound {
			errlog.Printf("%s: Redfish bug! Link %s was dead (404).  "+
				"Will try to continue.  No component will be created.",
				p.epRF.ID, rpath)
			p.LastStatus = RedfishSubtypeNoSupport
			p.RedfishSubtype = RFSubtypeUnknown
		} else {
			p.LastStatus = HTTPsGetFailed
		}
		return
	}
	p.processorRaw = &urlJSON
	p.LastStatus = HTTPsGetOk

	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, urlJSON)
	}
	// Decode JSON into Processor structure containing Redfish data
	if err := json.Unmarshal(urlJSON, &p.ProcessorRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			p.LastStatus = EPResponseFailedDecode
			return
		}
	}
	p.RedfishSubtype = p.ProcessorRF.ProcessorType

	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(p, "", "   ")
		errlog.Printf("%s: %s\n", url, jout)
	}

	p.LastStatus = VerifyingData
}

// This is the second discovery phase, after all information from
// the parent system has been gathered.  This is not intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the system first, so that it is available during later steps.
func (ps *EpProcessors) discoverLocalPhase2() error {
	var savedError error
	for i, p := range ps.OIDs {
		p.discoverLocalPhase2()
		if p.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF Processor type not supported: %s",
				i, p.RedfishSubtype)
		} else if p.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, p.LastStatus)
			errlog.Printf("Proccesors discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual processor.  Now that all information
// has been gathered, we can set the remaining fields needed to provide
// HMS with information about where the processor is located
func (p *EpProcessor) discoverLocalPhase2() {
	// Should never happen
	if p.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for odataID: %s\n",
			p.OdataID)
		p.LastStatus = EndpointInvalid
		return
	}
	if p.LastStatus != VerifyingData {
		return
	}

	// CPUs and GPUs are both under processors
	p.Ordinal = p.epRF.getProcessorOrdinal(p)
	if strings.ToLower(p.RedfishSubtype) == "gpu" {
		p.ID = p.sysRF.ID + "a" + strconv.Itoa(p.Ordinal)
		p.Type = base.NodeAccel.String()
	} else {
		p.ID = p.sysRF.ID + "p" + strconv.Itoa(p.Ordinal)
	}
	if p.ProcessorRF.Status.State != "Absent" {
		p.Status = "Populated"
		p.State = base.StatePopulated.String()
		p.Flag = base.FlagOK.String()
		if p.ProcessorRF.SerialNumber == "" {
			//look for special case GBTProcessorOemProperty
			if p.ProcessorRF.Oem != nil {
				if p.ProcessorRF.Oem.GBTProcessorOemProperty != nil {
					if p.ProcessorRF.Oem.GBTProcessorOemProperty.ProcessorSerialNumber != "" {
						p.ProcessorRF.SerialNumber = p.ProcessorRF.Oem.GBTProcessorOemProperty.ProcessorSerialNumber
					}
				}
			}
		}
		generatedFRUID, err := GetProcessorFRUID(p)
		if err != nil {
			errlog.Printf("FRUID Error: %s\n", err.Error())
			errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
		}
		p.FRUID = generatedFRUID
	} else {
		p.Status = "Empty"
		p.State = base.StateEmpty.String()
		//the state of the component is known (empty), it is not locked, does not have an alert or warning, so therefore Flag defaults to OK.
		p.Flag = base.FlagOK.String()
	}
	// Check if we have something valid to insert into the data store
	if (base.GetHMSType(p.ID) != base.Processor ||
		p.Type != base.Processor.String()) &&
		(base.GetHMSType(p.ID) != base.NodeAccel ||
			p.Type != base.NodeAccel.String()) {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s') for: %s\n",
			p.ID, p.Type, p.ProcessorURL)
		p.LastStatus = VerificationFailed
		return
	}

	errlog.Printf("Processor xname ID ('%s') and Type ('%s') for: %s\n", p.ID, p.Type, p.ProcessorURL)
	p.LastStatus = DiscoverOK
}

/////////////////////////////////////////////////////////////////////////////
// ComputerSystem - Memory Modules
/////////////////////////////////////////////////////////////////////////////

// This is one of possibly several memory modules, e.g. DIMMS for a particular
// EpSystem(Redfish "ComputerSystem" or just "System").
type EpMemory struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	BaseOdataID string `json:"BaseOdataID"`

	MemoryURL  string `json:"memoryURL"`  // Full URL to this RF Memory obj
	ParentOID  string `json:"parentOID"`  // odata.id for parent
	ParentType string `json:"parentType"` // ComputerSystem or Manager
	LastStatus string `json:"LastStatus"`

	MemoryRF  Memory           `json:"MemoryRF"`
	memoryRaw *json.RawMessage //`json:"memoryRaw"`

	epRF  *RedfishEP // Backpointer to RF EP, for connection details, etc.
	sysRF *EpSystem  // Backpointer to parent system.
}

// Set of EpMemory, each representing a Redfish "Memory" object, e.g. DIMM
// listed under a Redfish ComputerSystem aka System.  Unlike Systems/Chassis,
// there is no top-level listing of all memory under an RF endpoint, i.e.
// each one is specific to a collection belonging to a single parent System.
type EpMemoryMods struct {
	Num  int                  `json:"num"`
	OIDs map[string]*EpMemory `json:"oids"`
}

// Initializes EpMemory struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the Memory module to
// look at.  This should be the only way this struct is created for
// Redfish Memory objects under a system.
func NewEpMemory(s *EpSystem, odataID ResourceID, rawOrdinal int) *EpMemory {
	m := new(EpMemory)
	m.OdataID = odataID.Oid
	m.Type = base.Memory.String()
	m.BaseOdataID = odataID.Basename()
	m.RedfishType = MemoryType
	m.RfEndpointID = s.epRF.ID

	m.MemoryURL = s.epRF.FQDN + odataID.Oid

	m.ParentOID = s.OdataID
	m.ParentType = ComputerSystemType

	m.Ordinal = -1
	m.RawOrdinal = rawOrdinal

	m.LastStatus = NotYetQueried
	m.epRF = s.epRF
	m.sysRF = s

	return m
}

// Makes contact with redfish endpoint to discover information about
// all memory modules for a system or manager.  EpMemory entries
// should be created with the appropriate constructor first.
func (ms *EpMemoryMods) discoverRemotePhase1() {
	for _, m := range ms.OIDs {
		m.discoverRemotePhase1()
	}
}

// Makes contact with redfish endpoint to discover information about
// a particular Memory module for a ComputerSystem/System.   Note that the
// EpMemory struct should be created with the appropriate constructor first.
func (m *EpMemory) discoverRemotePhase1() {
	rpath := m.OdataID
	url := m.epRF.FQDN + rpath
	urlJSON, err := m.epRF.GETRelative(rpath)
	if err != nil || urlJSON == nil {
		if err == ErrRFDiscURLNotFound {
			errlog.Printf("%s: Redfish bug! Link %s was dead (404).  "+
				"Will try to continue.  No component will be created.",
				m.epRF.ID, rpath)
			m.LastStatus = RedfishSubtypeNoSupport
			m.RedfishSubtype = RFSubtypeUnknown
		} else {
			m.LastStatus = HTTPsGetFailed
		}
		return
	}
	m.memoryRaw = &urlJSON
	m.LastStatus = HTTPsGetOk

	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, urlJSON)
	}
	// Decode JSON into Memory structure containing the Redfish data.
	if err := json.Unmarshal(urlJSON, &m.MemoryRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			m.LastStatus = EPResponseFailedDecode
			return
		}
	}
	m.RedfishSubtype = m.MemoryRF.MemoryType

	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(m, "", "   ")
		errlog.Printf("%s: %s\n", url, jout)
	}
	m.LastStatus = VerifyingData
}

// This is the second discovery phase, after all information from
// the parent system has been gathered.  This is not intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the system first, so that it is available during later steps.
func (ms *EpMemoryMods) discoverLocalPhase2() error {
	var savedError error
	for i, m := range ms.OIDs {
		m.discoverLocalPhase2()
		if m.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF Memory type not supported: %s",
				i, m.RedfishSubtype)
		} else if m.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, m.LastStatus)
			errlog.Printf("MMods discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual memory module.  Now that all information
// has been gathered, we can set the remaining fields needed to provide
// HMS with information about where the memory module is located
func (m *EpMemory) discoverLocalPhase2() {
	// Should never happen
	if m.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for odataID: %s\n",
			m.OdataID)
		m.LastStatus = EndpointInvalid
		return
	}
	if m.LastStatus != VerifyingData {
		return
	}

	m.Ordinal = m.epRF.getMemoryOrdinal(m)
	m.ID = m.sysRF.ID + "d" + strconv.Itoa(m.Ordinal)
	if m.MemoryRF.Status.State != "Absent" {
		m.Status = "Populated"
		m.State = base.StatePopulated.String()
		m.Flag = base.FlagOK.String()
		generatedFRUID, err := GetMemoryFRUID(m)
		if err != nil {
			errlog.Printf("FRUID Error: %s\n", err.Error())
			errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
		}
		m.FRUID = generatedFRUID
	} else {
		m.Status = "Empty"
		m.State = base.StateEmpty.String()
		//the state of the component is known (empty), it is not locked, does not have an alert or warning, so therefore Flag defaults to OK.
		m.Flag = base.FlagOK.String()
	}
	// Check if we have something valid to insert into the data store
	if base.GetHMSType(m.ID) != base.Memory || m.Type != base.Memory.String() {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s') for %s\n",
			m.ID, m.Type, m.MemoryURL)
		m.LastStatus = VerificationFailed
		return
	}
	m.LastStatus = DiscoverOK
}
