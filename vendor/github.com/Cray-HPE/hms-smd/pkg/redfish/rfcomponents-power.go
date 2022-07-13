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

	base "github.com/Cray-HPE/hms-base"
	//"strings"
)

/////////////////////////////////////////////////////////////////////////////
//
// Redfish PowerDistribution (i.e. PDU) Discovery, plus child Outlets, etc.
//
/////////////////////////////////////////////////////////////////////////////

// This represents a PDU.  Most of the behavior is defined on the child
// outlets that we will discover as children, (e.g. like processors under
// Systems), a relationship that will be hold in the detailed hardware
// inventory.
type EpPDU struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	// Embedded struct - PDU specific info
	ComponentPDUInfo

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	BaseOdataID             string            `json:"BaseOdataID"`
	PDUURL                  string            `json:"pduURL"` // Full URL to me
	LastStatus              string            `json:"lastStatus"`
	PowerDistributionRF     PowerDistribution `json:"powerDistributionRF"`
	powerDistributionURLRaw *json.RawMessage

	// Child/linked components
	Outlets EpOutlets `json:"outlets"`
	//Circuits EpCircuits `json:"circuits"`

	epRF *RedfishEP // Backpointer, for connection details, etc.
}

// Set of EpPDU, representing a Redfish "PowerDistribution" object, e.g. RackPDU
type EpPDUs struct {
	Num  int               `json:"num"`
	OIDs map[string]*EpPDU `json:"oids"`
}

// Initializes EpSystem struct with minimal information needed to
// discover it, i.e. endpoint info and the OdataID of the system to look at.
// This should be the only way this struct is created.
func NewEpPDU(epRF *RedfishEP, odataID ResourceID, rawOrdinal int) *EpPDU {
	pdu := new(EpPDU)
	pdu.Type = base.HMSTypeInvalid.String() // Must be updated later
	pdu.OdataID = odataID.Oid
	pdu.BaseOdataID = odataID.Basename()
	pdu.RedfishType = PDUType
	pdu.RfEndpointID = epRF.ID
	pdu.LastStatus = NotYetQueried
	pdu.Ordinal = -1
	pdu.RawOrdinal = rawOrdinal
	pdu.epRF = epRF
	return pdu
}

// Makes contact with the remote endpoint to discover basic information
// abou all Redfish PowerDistribution objects in EpPDUs.  Phase1 discovery
// fetches all relevant data from the RF entry point but does not fully
// discover all info.  This is left for Phase2, which is intended to be
// run only after ALL components (managers, chassis, systems, etc.) have
// completed phase 1 under a particular endpoint.   Note that we also
// gather Outlet and (if needed later) Circuit info.
func (pdus *EpPDUs) discoverRemotePhase1() {
	for _, pdu := range pdus.OIDs {
		pdu.discoverRemotePhase1()
	}
}

// Makes contact with remote endpoint to discover information about
// the given PowerDistribution object, e.g. RackPDU
func (pdu *EpPDU) discoverRemotePhase1() {
	// Should never happen
	if pdu.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for system odataID: %s\n",
			pdu.OdataID)
		pdu.LastStatus = EndpointInvalid
		return
	}
	pdu.PDUURL = pdu.epRF.FQDN + pdu.OdataID

	path := pdu.OdataID
	url := pdu.epRF.FQDN + path
	topURL := url
	pduURLJSON, err := pdu.epRF.GETRelative(path)
	if err != nil || pduURLJSON == nil {
		pdu.LastStatus = HTTPsGetFailed
		return
	}
	pdu.powerDistributionURLRaw = &pduURLJSON
	pdu.LastStatus = HTTPsGetOk

	// Decode JSON into ComputerSystem structure
	if err := json.Unmarshal(pduURLJSON, &pdu.PowerDistributionRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			pdu.LastStatus = EPResponseFailedDecode
			return
		}
	}
	pdu.RedfishSubtype = pdu.PowerDistributionRF.EquipmentType
	pdu.UUID = pdu.PowerDistributionRF.UUID
	if pdu.PowerDistributionRF.Actions != nil {
		pdu.Actions = pdu.PowerDistributionRF.Actions
	}
	pdu.Name = pdu.PowerDistributionRF.Name
	//
	// Get link to PDU OutletCollection
	//

	if pdu.PowerDistributionRF.Outlets.Oid == "" {
		errlog.Printf("%s: No OutletsCollection found.\n", topURL)
		pdu.Outlets.Num = 0
		pdu.Outlets.OIDs = make(map[string]*EpOutlet)
	} else {
		path = pdu.PowerDistributionRF.Outlets.Oid
		url = pdu.epRF.FQDN + path
		outsJSON, err := pdu.epRF.GETRelative(path)
		if err != nil || outsJSON == nil {
			pdu.LastStatus = HTTPsGetFailed
			return
		}
		if rfDebug > 0 {
			errlog.Printf("%s: %s\n", url, outsJSON)
		}
		pdu.LastStatus = HTTPsGetOk

		var outInfo OutletCollection
		if err := json.Unmarshal(outsJSON, &outInfo); err != nil {
			errlog.Printf("Failed to decode %s: %s\n", url, err)
			pdu.LastStatus = EPResponseFailedDecode
		}

		// HPE PDUs use Outlets instead of Members, so copy to Members
		if len(outInfo.Outlets) > 0 {
			outInfo.Members = outInfo.Outlets
		}

		// The count is typically given as "Members@odata.count", but
		// older versions drop the "Members" identifier
		if outInfo.MembersOCount > 0 && outInfo.MembersOCount != len(outInfo.Members) {
			errlog.Printf("%s: Member@odata.count != Member array len\n", url)
		} else if outInfo.OCount > 0 && outInfo.OCount != len(outInfo.Members) {
			errlog.Printf("%s: odata.count != Member array len\n", url)
		}
		pdu.Outlets.Num = len(outInfo.Members)
		pdu.Outlets.OIDs = make(map[string]*EpOutlet)

		// Sort in lexical order, so the ordinal values will keep the same
		// ordering.
		sort.Sort(ResourceIDSlice(outInfo.Members))
		for i, outOID := range outInfo.Members {
			outID := outOID.Basename()
			pdu.Outlets.OIDs[outID] = NewEpOutlet(pdu, outOID, i)
		}
		pdu.Outlets.discoverRemotePhase1()
	}

	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(pdu, "", "   ")
		errlog.Printf("%s: %s\n", topURL, jout)
	}
	pdu.LastStatus = VerifyingData

}

// PDUs: This is the second discovery phase, after all information from
// the parent endpoint has been gathered.  This is not really intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the endpoint first, so that it is available during phase 2 steps.
func (pdus *EpPDUs) discoverLocalPhase2() error {
	var savedError error
	for i, pdu := range pdus.OIDs {
		pdu.discoverLocalPhase2()
		if pdu.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF PDUType not supported: %s",
				i, pdu.RedfishSubtype)
		} else if pdu.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, pdu.LastStatus)
			errlog.Printf("PDUs discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual PDU.  Now that all information
// has been gathered, we can set the remaining fields to set the
// corresponding xname, state and so on.
func (pdu *EpPDU) discoverLocalPhase2() {
	// Should never happen
	if pdu.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for odataID: %s\n",
			pdu.OdataID)
		pdu.LastStatus = EndpointInvalid
		return
	}
	if pdu.LastStatus != VerifyingData {
		return
	}
	// Set Outlet ordinal, type and xname
	// There may be outlet types that are not supported.
	pdu.Ordinal = pdu.epRF.getPDUOrdinal(pdu)
	pdu.Type = pdu.epRF.getPDUHMSType(pdu, pdu.Ordinal)
	if pdu.Type == base.HMSTypeInvalid.String() {
		pdu.LastStatus = RedfishSubtypeNoSupport
		return
	}
	pdu.ID = pdu.epRF.getPDUHMSID(pdu, pdu.Type, pdu.Ordinal)

	// Set HMS State and Flag and set PDU (if non-empty)
	pdu.discoverComponentState()

	// Check if we have something valid to insert into the data store
	if base.GetHMSTypeString(pdu.ID) != pdu.Type ||
		pdu.Type == base.HMSTypeInvalid.String() || pdu.Type == "" {

		errlog.Printf("PDU: Error: Bad xname ID ('%s') or Type ('%s') for %s\n",
			pdu.ID, pdu.Type, pdu.PDUURL)
		pdu.LastStatus = VerificationFailed
		return
	}
	// Complete discovery and verify subcomponents
	var childStatus string = DiscoverOK
	if err := pdu.Outlets.discoverLocalPhase2(); err != nil {
		childStatus = ChildVerificationFailed
	}
	pdu.LastStatus = childStatus
}

// Sets up HMS state fields for PDUs using Status/State/Health info
// from Redfish
func (pdu *EpPDU) discoverComponentState() {
	if pdu.PowerDistributionRF.Status.State != "Absent" {
		pdu.Status = "Populated"
		pdu.State = base.StatePopulated.String()
		pdu.Flag = base.FlagOK.String()
		if pdu.PowerDistributionRF.PowerState != "" {
			if pdu.PowerDistributionRF.PowerState == POWER_STATE_OFF ||
				pdu.PowerDistributionRF.PowerState == POWER_STATE_POWERING_ON {
				pdu.State = base.StateOff.String()
			} else if pdu.PowerDistributionRF.PowerState == POWER_STATE_ON ||
				pdu.PowerDistributionRF.PowerState == POWER_STATE_POWERING_OFF {
				pdu.State = base.StateOn.String()
			}
		} else {
			if pdu.PowerDistributionRF.Status.Health == "OK" {
				if pdu.PowerDistributionRF.Status.State == "Enabled" {
					pdu.State = base.StateOn.String()
				}
			}
		}
		if pdu.PowerDistributionRF.Status.Health == "Warning" {
			pdu.Flag = base.FlagWarning.String()
		} else if pdu.PowerDistributionRF.Status.Health == "Critical" {
			pdu.Flag = base.FlagAlert.String()
		}
		generatedFRUID, err := GetPDUFRUID(pdu)
		if err != nil {
			errlog.Printf("FRUID Error: %s\n", err.Error())
			errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
		}
		pdu.FRUID = generatedFRUID
	} else {
		pdu.Status = "Empty"
		pdu.State = base.StateEmpty.String()
		//the state of the component is known (empty), it is not locked, does not have an alert or warning, so therefore Flag defaults to OK.
		pdu.Flag = base.FlagOK.String()
	}
}

/////////////////////////////////////////////////////////////////////////////
// Redfish Outlets - Children of PDUs (and related power components)
/////////////////////////////////////////////////////////////////////////////

// This represents an outlet under a PDU parent.  It is the individual outlets
// that will be interacted with (e.g. turned on or off) in most cases.
type EpOutlet struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	// Embedded struct - PDU specific info
	ComponentOutletInfo

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	BaseOdataID  string `json:"BaseOdataID"`
	OutletURL    string `json:"outletURL"` // Full URL to this Chassis
	LastStatus   string `json:"lastStatus"`
	OutletRF     Outlet `json:"outletRF"`
	outletURLRaw *json.RawMessage

	epPDU *EpPDU     // Backpointer to parent PDU
	epRF  *RedfishEP // Backpointer, for connection details, etc.
}

// Set of EpOutlet, representing a individual outlets under a Redfish
// "PowerDistribution" object, e.g. RackPDU.
type EpOutlets struct {
	Num  int                  `json:"num"`
	OIDs map[string]*EpOutlet `json:"oids"`
}

// Initializes EpSystem struct with minimal information needed to
// discover it, i.e. endpoint info and the OdataID of the system to look at.
// This should be the only way this struct is created.
func NewEpOutlet(pdu *EpPDU, odataID ResourceID, rawOrdinal int) *EpOutlet {
	out := new(EpOutlet)
	out.Type = base.HMSTypeInvalid.String() // Must be updated later
	out.OdataID = odataID.Oid
	out.BaseOdataID = odataID.Basename()
	out.RedfishType = OutletType
	out.RfEndpointID = pdu.epRF.ID
	out.LastStatus = NotYetQueried
	out.Ordinal = -1
	out.RawOrdinal = rawOrdinal
	out.epPDU = pdu
	out.epRF = pdu.epRF
	return out
}

// Makes contact with the remote endpoint to discover basic information
// abou all Redfish PowerDistribution objects in EpPDUs.  Phase1 discovery
// fetches all relevant data from the RF entry point but does not fully
// discover all info.  This is left for Phase2, which is intended to be
// run only after ALL components (managers, chassis, systems, etc.) have
// completed phase 1 under a particular endpoint.   Note that we also
// gather Outlet and (if needed later) Circuit info.
func (outs *EpOutlets) discoverRemotePhase1() {
	for _, out := range outs.OIDs {
		out.discoverRemotePhase1()
	}
}

// Makes contact with remote endpoint to discover information about
// the given outlet under a PowerDistribution object
func (out *EpOutlet) discoverRemotePhase1() {
	// Should never happen
	if out.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for system odataID: %s\n",
			out.OdataID)
		out.LastStatus = EndpointInvalid
		return
	}
	out.OutletURL = out.epRF.FQDN + out.OdataID

	path := out.OdataID
	url := out.epRF.FQDN + path
	topURL := url
	outURLJSON, err := out.epRF.GETRelative(path)
	if err != nil || outURLJSON == nil {
		out.LastStatus = HTTPsGetFailed
		return
	}
	out.outletURLRaw = &outURLJSON
	out.LastStatus = HTTPsGetOk

	// Decode JSON into ComputerSystem structure
	if err := json.Unmarshal(outURLJSON, &out.OutletRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			out.LastStatus = EPResponseFailedDecode
			return
		}
	}
	out.RedfishSubtype = out.OutletRF.OutletType
	if out.OutletRF.Actions != nil {
		out.Actions = out.OutletRF.Actions
		// HPE PDUs do not supply Allowable Values, so add them
		if out.Actions.PowerControl != nil && len(out.Actions.PowerControl.AllowableValues) == 0 {
			out.Actions.PowerControl.AllowableValues = append(out.Actions.PowerControl.AllowableValues, "On", "Off")
		}
	}
	out.Name = out.OutletRF.Name
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(out, "", "   ")
		errlog.Printf("%s: %s\n", topURL, jout)
	}
	out.LastStatus = VerifyingData
}

// Outlets: This is the second discovery phase, after all information from
// the parent endpoint has been gathered.  This is not really intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the endpoint first, so that it is available during phase 2 steps.
func (outs *EpOutlets) discoverLocalPhase2() error {
	var savedError error
	for i, out := range outs.OIDs {
		out.discoverLocalPhase2()
		if out.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF OutletType not supported: %s",
				i, out.RedfishSubtype)
		} else if out.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, out.LastStatus)
			errlog.Printf("Outlets discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual Outlet of a PDU.  Now that all
// information has been gathered, we can set the remaining fields to
// set the corresponding xnames and so on.
func (out *EpOutlet) discoverLocalPhase2() {
	// Should never happen
	if out.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for odataID: %s\n",
			out.OdataID)
		out.LastStatus = EndpointInvalid
		return
	}
	if out.LastStatus != VerifyingData {
		return
	}
	// Set Outlet ordinal, type and xname
	// There may be outlet types that are not supported.
	out.Ordinal = out.epRF.getOutletOrdinal(out)
	out.Type = out.epRF.getOutletHMSType(out)
	if out.Type == base.HMSTypeInvalid.String() {
		out.LastStatus = RedfishSubtypeNoSupport
		return
	}
	out.ID = out.epRF.getOutletHMSID(out, out.Type, out.Ordinal)

	// Set HMS State and Flag
	out.discoverComponentState()

	// Check if we have something valid to insert into the data store
	if base.GetHMSTypeString(out.ID) != out.Type ||
		out.Type == base.HMSTypeInvalid.String() || out.Type == "" {

		errlog.Printf("Outlet: Error: Bad ID ('%s') or Type ('%s') for %s\n",
			out.ID, out.Type, out.OutletURL)
		out.LastStatus = VerificationFailed
		return
	}
	out.LastStatus = DiscoverOK
}

// Sets up HMS state fields for Outlets using Status/State/Health info
// from Redfish
func (out *EpOutlet) discoverComponentState() {
	if out.OutletRF.Status.State != "Absent" {
		out.Status = "Populated"
		out.State = base.StatePopulated.String()
		out.Flag = base.FlagOK.String()
		if out.OutletRF.PowerState != "" {
			if out.OutletRF.PowerState == POWER_STATE_OFF ||
				out.OutletRF.PowerState == POWER_STATE_POWERING_ON {
				out.State = base.StateOff.String()
			} else if out.OutletRF.PowerState == POWER_STATE_ON ||
				out.OutletRF.PowerState == POWER_STATE_POWERING_OFF {
				out.State = base.StateOn.String()
			}
		} else {
			if out.OutletRF.Status.Health == "OK" {
				if out.OutletRF.Status.State == "Enabled" {
					out.State = base.StateOn.String()
				}
			}
		}
		if out.OutletRF.Status.Health == "Warning" {
			out.Flag = base.FlagWarning.String()
		} else if out.OutletRF.Status.Health == "Critical" {
			out.Flag = base.FlagAlert.String()
		}
		if out.epPDU.FRUID != "" {
			out.FRUID = out.Type + "." + strconv.Itoa(out.Ordinal) +
				"." + out.epPDU.FRUID
		} else {
			out.FRUID = "FRUIDfor" + out.ID
		}
	} else {
		out.Status = "Empty"
		out.State = base.StateEmpty.String()
		//the state of the component is known (empty), it is not locked, does not have an alert or warning, so therefore Flag defaults to OK.
		out.Flag = base.FlagOK.String()
	}
}

//////////////////////////////////////////////////////////////////////////
//
// Power field discovery
//
//////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////
// PDUs (PowerDistribution)
//////////////////////////////////////////////////////////////////////////

// Determines based on discovered info the xname of the PDU.
// If there is only expected to be one for the endpoint, it will have
// the same xname.
func (ep *RedfishEP) getPDUHMSID(pdu *EpPDU, hmsType string, ordinal int) string {
	// Note every hmsType and ordinal pair must get a unique xname ID
	hmsTypeStr := base.VerifyNormalizeType(hmsType)
	if hmsTypeStr == "" {
		// This is an error or a skipped type.
		return ""
	}
	if ordinal < 0 {
		// Ordinal was never set.
		return ""
	}
	if hmsType == base.CabinetPDU.String() {
		return pdu.epRF.ID + "p" + strconv.Itoa(ordinal)
	}
	// Something went wrong
	return ""
}

// Get the HMS type of the Cabinet PDU
func (ep *RedfishEP) getPDUHMSType(pdu *EpPDU, ordinal int) string {
	// There can 1 or more Cabinet PDUs under a Cabinet PDU controller
	if ep.Type == base.CabinetPDUController.String() && ordinal >= 0 {
		return base.CabinetPDU.String()
	}
	// Shouldn't happen
	return base.HMSTypeInvalid.String()
}

// Determines based on discovered info and original list order what the
// Manager ordinal is, i.e. the b[0-n] in the xname.
func (ep *RedfishEP) getPDUOrdinal(pdu *EpPDU) int {
	if pdu.RawOrdinal < 0 {
		return -1
	}
	return pdu.RawOrdinal
}

//////////////////////////////////////////////////////////////////////////
// Outlets
//////////////////////////////////////////////////////////////////////////

// Determines based on discovered info the xname of the Outlet.
// If there is only one and has the same type as the manager, it must be
// the same as the parent RedfishEndpoint's xname ID.
func (ep *RedfishEP) getOutletHMSID(out *EpOutlet, hmsType string, ordinal int) string {
	// Note every hmsType and ordinal pair must get a unique xname ID
	hmsTypeStr := base.VerifyNormalizeType(hmsType)
	if hmsTypeStr == "" {
		// This is an error or a skipped type.
		return ""
	}
	if ordinal < 0 {
		// Ordinal was never set (properly)
		return ""
	}
	if hmsType == base.CabinetPDUOutlet.String() {
		// Do not start at zero for the jJ portion of the xname,
		// start at one.  We keep the ordinal at the original value for
		// consistency in hwinv so ordinal 0 => xXmMpPj1
		return out.epPDU.ID + "j" + strconv.Itoa(ordinal+1)
	}
	if hmsType == base.CabinetPDUPowerConnector.String() {
		// Do not start at zero for the jJ portion of the xname,
		// start at one.  We keep the ordinal at the original value for
		// consistency in hwinv so ordinal 0 => xXmMpPj1
		return out.epPDU.ID + "v" + strconv.Itoa(ordinal+1)
	}
	// Something went wrong
	return ""
}

// Get the HMS type of the Outlet
func (ep *RedfishEP) getOutletHMSType(out *EpOutlet) string {
	// Just one?  That's this endpoint's type.
	if out.epPDU.Type == base.CabinetPDU.String() {
		return base.CabinetPDUPowerConnector.String()
	}
	// Shouldn't happen
	return base.HMSTypeInvalid.String()
}

// Determines based on discovered info and original list order what the
// Outlet ordinal is, i.e. the j[1-n] in the xname.
// NOTE: Keep ordinal at zero, but xname will start from 1 not 0,
// like all outlets.
func (ep *RedfishEP) getOutletOrdinal(out *EpOutlet) int {
	return out.RawOrdinal
}

/////////////////////////////////////////////////////////////////////////////
// Chassis - Power
/////////////////////////////////////////////////////////////////////////////

// This is the top-level Power object for a particular Chassis.
type EpPower struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	InventoryData

	PowerURL   string `json:"powerURL"`   // Full URL to this RF Power obj
	ParentOID  string `json:"parentOID"`  // odata.id for parent
	ParentType string `json:"parentType"` // Chassis
	LastStatus string `json:"LastStatus"`

	PowerRF  Power `json:"PowerRF"`
	PowerRaw *json.RawMessage

	epRF      *RedfishEP // Backpointer to RF EP, for connection details, etc.
	chassisRF *EpChassis // Backpointer to parent chassis.
}

// Initializes EpPower struct with minimal information needed to
// pass along to its children.
func NewEpPower(chassis *EpChassis, odataID ResourceID) *EpPower {
	p := new(EpPower)
	p.OdataID = odataID.Oid
	p.Type = PowerType
	p.BaseOdataID = odataID.Basename()
	p.RedfishType = PowerType
	p.RfEndpointID = chassis.epRF.ID

	p.PowerURL = chassis.epRF.FQDN + odataID.Oid
	p.ParentOID = chassis.OdataID
	p.ParentType = ChassisType

	p.LastStatus = NotYetQueried
	p.epRF = chassis.epRF
	p.chassisRF = chassis

	return p
}

// Makes contact with redfish endpoint to discover information about
// the Power object under a Chassis. Note that the
// EpPower should be created with the appropriate constructor first.
func (p *EpPower) discoverRemotePhase1() {
	path := p.chassisRF.ChassisRF.Power.Oid
	url := p.epRF.FQDN + path
	powerJSON, err := p.epRF.GETRelative(path)
	if err != nil || powerJSON == nil {
		p.LastStatus = HTTPsGetFailed
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, powerJSON)
	}
	p.PowerRaw = &powerJSON
	p.LastStatus = HTTPsGetOk

	if err := json.Unmarshal(powerJSON, &p.PowerRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			p.LastStatus = EPResponseFailedDecode
			return
		}
	}
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(p, "", "   ")
		errlog.Printf("%s: %s\n", url, jout)
	}
	p.LastStatus = VerifyingData

}

/////////////////////////////////////////////////////////////////////////////
// Chassis - PowerSupplies
/////////////////////////////////////////////////////////////////////////////

// Set of EpPowerSupply, each representing a Redfish "PowerSupply"
// listed under a Redfish Chassis.
type EpPowerSupplies struct {
	Num  int                       `json:"num"`
	OIDs map[string]*EpPowerSupply `json:"oids"`
}

// This is one of possibly several PowerSupplies for a particular
// EpChassis (Redfish "Chassis").
type EpPowerSupply struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	PowerSupplyURL string `json:"powerSupplyURL"` // Full URL to this RF PowerSupply obj
	ParentOID      string `json:"parentOID"`      // odata.id for parent
	ParentType     string `json:"parentType"`     // Chassis
	LastStatus     string `json:"LastStatus"`

	PowerSupplyRF  *PowerSupply `json:"PowerSupplyRF"`
	powerSupplyRaw *json.RawMessage

	epRF      *RedfishEP // Backpointer to RF EP, for connection details, etc.
	chassisRF *EpChassis // Backpointer to parent chassis.
	powerRF   *EpPower   // Backpointer to parent Power obj.
}

// Initializes EpPowerSupply struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the PowerSupply to
// look at.  This should be the only way this struct is created for
// PowerSupplies under a chassis.
func NewEpPowerSupply(power *EpPower, odataID ResourceID, rawOrdinal int) *EpPowerSupply {
	ps := new(EpPowerSupply)
	ps.OdataID = odataID.Oid

	ps.Type = PowerSupplyType
	ps.BaseOdataID = odataID.Basename()
	ps.RedfishType = PowerSupplyType
	ps.RfEndpointID = power.epRF.ID

	ps.PowerSupplyURL = power.epRF.FQDN + odataID.Oid
	ps.ParentOID = power.OdataID
	ps.ParentType = PowerType

	ps.Ordinal = -1
	ps.RawOrdinal = rawOrdinal

	ps.LastStatus = NotYetQueried
	ps.epRF = power.epRF

	ps.chassisRF = power.chassisRF
	ps.powerRF = power

	return ps
}

// Examines parent Power redfish endpoint to discover information about
// all PowerSupplies for a given Redfish Chassis.  EpPowerSupply entries
// should be created with the appropriate constructor first.
func (ps *EpPowerSupplies) discoverRemotePhase1() {
	for _, p := range ps.OIDs {
		p.discoverRemotePhase1()
	}
}

// Retrieves the PowerSupplyRF from the parent Power.PowerSupplies array.
func (p *EpPowerSupply) discoverRemotePhase1() {

	//Since the parent Power obj already has the array of PowerSupply objects
	//the PowerSupplyRF field for this PowerSupply just needs to be pulled from there
	//instead of retrieving it using an HTTP call
	if p.powerRF.PowerRF.PowerSupplies == nil {
		//this is a lookup error
		errlog.Printf("%s: No PowerSupplies array found in Parent.\n", p.OdataID)
		p.LastStatus = HTTPsGetFailed
		return
	}
	//If we got this far, then the EpPower call to discoverRemotePhase1 was successful
	p.LastStatus = HTTPsGetOk

	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", p.ParentOID, p.powerRF.PowerRaw)
	}

	//use p.RawOrdinal as index to retrieve the PowerSupply entry from the parent Power.PowerSupplies array,
	//and assign it to p.PowerSupplyRF
	if len(p.powerRF.PowerRF.PowerSupplies) > p.RawOrdinal && p.powerRF.PowerRF.PowerSupplies[p.RawOrdinal] != nil {
		p.PowerSupplyRF = p.powerRF.PowerRF.PowerSupplies[p.RawOrdinal]
	} else {
		//this is a lookup error
		errlog.Printf("%s: failure retrieving PowerSupply from Power.PowerSupplies[%d].\n", p.OdataID, p.RawOrdinal)
		p.LastStatus = HTTPsGetFailed
		return
	}
	p.RedfishSubtype = PowerSupplyType //TODO: determine if there is a better value for this

	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(p, "", "   ")
		errlog.Printf("%s: %s\n", p.PowerSupplyURL, jout)
	}

	p.LastStatus = VerifyingData
}

// This is the second discovery phase, after all information from
// the parent system has been gathered.  This is not intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the chassis first, so that it is available during later steps.
func (ps *EpPowerSupplies) discoverLocalPhase2() error {
	var savedError error
	for i, p := range ps.OIDs {
		p.discoverLocalPhase2()
		if p.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF PowerSupply type not supported: %s",
				i, p.RedfishSubtype)
		} else if p.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, p.LastStatus)
			errlog.Printf("PowerSupplies discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual PowerSupply.  Now that all information
// has been gathered, we can set the remaining fields needed to provide
// HMS with information about where the PowerSupply is located
func (p *EpPowerSupply) discoverLocalPhase2() {
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

	p.Ordinal = p.epRF.getPowerSupplyOrdinal(p)
	p.Type = p.epRF.getPowerSupplyHMSType(p)
	p.ID = p.epRF.getPowerSupplyHMSID(p, p.Type, p.Ordinal)
	if p.PowerSupplyRF.Status.State != "Absent" {
		p.Status = "Populated"
		p.State = base.StatePopulated.String()
		p.Flag = base.FlagOK.String()
		generatedFRUID, err := GetPowerSupplyFRUID(p)
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
	if ((base.GetHMSType(p.ID) == base.CMMRectifier) || (base.GetHMSType(p.ID) == base.NodeEnclosurePowerSupply)) && (p.Type == base.CMMRectifier.String() || p.Type == base.NodeEnclosurePowerSupply.String()) {
		if rfVerbose > 0 {
			errlog.Printf("PowerSupply discoverLocalPhase2: VALID xname ID ('%s') and Type ('%s') for: %s\n",
				p.ID, p.Type, p.PowerSupplyURL)
		}
	} else {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s') for: %s\n",
			p.ID, p.Type, p.PowerSupplyURL)
		p.LastStatus = VerificationFailed
		return
	}
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(p, "", "   ")
		errlog.Printf("%s\n", jout)
		errlog.Printf("PowerSupply ID: %s\n", p.ID)
		errlog.Printf("PowerSupply FRUID: %s\n", p.FRUID)
	}
	p.LastStatus = DiscoverOK
}
