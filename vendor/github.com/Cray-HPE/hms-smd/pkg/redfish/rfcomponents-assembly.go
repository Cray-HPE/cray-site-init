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

	base "github.com/Cray-HPE/hms-base"
)

/////////////////////////////////////////////////////////////////////////////
// Chassis - Assembly
/////////////////////////////////////////////////////////////////////////////

// This is the top-level Assembly object for a particular Chassis.
type EpAssembly struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	InventoryData

	AssemblyURL string `json:"assemblyURL"` // Full URL to this RF Assembly obj
	ParentOID   string `json:"parentOID"`   // odata.id for parent
	ParentType  string `json:"parentType"`  // Chassis
	LastStatus  string `json:"LastStatus"`

	AssemblyRF  Assembly `json:"AssemblyRF"`
	AssemblyRaw *json.RawMessage

	systemRF *EpSystem  // Backpointer to associated system.
	epRF     *RedfishEP // Backpointer to RF EP, for connection details, etc.
}

// Initializes EpAssembly struct with minimal information needed to
// pass along to its children.
func NewEpAssembly(s *EpSystem, odataID ResourceID, pOID, pType string) *EpAssembly {
	a := new(EpAssembly)
	a.OdataID = odataID.Oid
	a.Type = AssemblyType
	a.BaseOdataID = odataID.Basename()
	a.RedfishType = AssemblyType
	a.RfEndpointID = s.epRF.ID

	a.AssemblyURL = s.epRF.FQDN + odataID.Oid
	a.ParentOID = pOID
	a.ParentType = pType

	a.LastStatus = NotYetQueried
	a.systemRF = s
	a.epRF = s.epRF

	return a
}

// Makes contact with redfish endpoint to discover information about
// the Assembly object under a Chassis. Note that the
// EpAssembly should be created with the appropriate constructor first.
func (a *EpAssembly) discoverRemotePhase1() {
	path := a.OdataID
	url := a.epRF.FQDN + path
	assemblyJSON, err := a.epRF.GETRelative(path)
	if err != nil || assemblyJSON == nil {
		a.LastStatus = HTTPsGetFailed
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, assemblyJSON)
	}
	a.AssemblyRaw = &assemblyJSON
	a.LastStatus = HTTPsGetOk

	if err := json.Unmarshal(assemblyJSON, &a.AssemblyRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			a.LastStatus = EPResponseFailedDecode
			return
		}
	}
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(a, "", "   ")
		errlog.Printf("%s: %s\n", url, jout)
	}
	a.LastStatus = VerifyingData

}

/////////////////////////////////////////////////////////////////////////////
// Chassis - NodeAccelRisers
/////////////////////////////////////////////////////////////////////////////

// Set of EpNodeAccelRiser, each representing a GPUSubsystem baseboard "NodeAccelRiser"
// listed under a Redfish Chassis.
type EpNodeAccelRisers struct {
	Num  int                          `json:"num"`
	OIDs map[string]*EpNodeAccelRiser `json:"oids"`
}

// This is one of possibly several NodeAccelRiser cards for a particular
// EpChassis (Redfish "Chassis").
type EpNodeAccelRiser struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	NodeAccelRiserURL string `json:"nodeAccelRiserURL"` // Full URL to this RF PowerSupply obj
	ParentOID         string `json:"parentOID"`         // odata.id for parent
	ParentType        string `json:"parentType"`        // Chassis
	LastStatus        string `json:"LastStatus"`

	NodeAccelRiserRF  *NodeAccelRiser `json:"NodeAccelRiserRF"`
	nodeAccelRiserRaw *json.RawMessage

	epRF       *RedfishEP  // Backpointer to RF EP, for connection details, etc.
	systemRF   *EpSystem   // Backpointer to associated system.
	assemblyRF *EpAssembly // Backpointer to parent Assembly obj.
}

// Initializes EpNodeAccelRiser struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the NodeAccelRiser to
// look at.  This should be the only way this struct is created for
// NodeAccelRiser cards under a chassis.
func NewEpNodeAccelRiser(assembly *EpAssembly, odataID ResourceID, rawOrdinal int) *EpNodeAccelRiser {
	ar := new(EpNodeAccelRiser)
	ar.OdataID = odataID.Oid

	ar.Type = NodeAccelRiserType
	ar.BaseOdataID = odataID.Basename()
	ar.RedfishType = NodeAccelRiserType
	ar.RfEndpointID = assembly.epRF.ID

	ar.NodeAccelRiserURL = assembly.epRF.FQDN + odataID.Oid
	ar.ParentOID = assembly.OdataID
	ar.ParentType = assembly.Type

	ar.Ordinal = -1
	ar.RawOrdinal = rawOrdinal

	ar.LastStatus = NotYetQueried
	ar.epRF = assembly.epRF
	ar.systemRF = assembly.systemRF
	ar.assemblyRF = assembly

	return ar
}

// Examines parent Assembly redfish endpoint to discover information about
// all NodeAccelRisers for a given Redfish Chassis.  EpNodeAccel entries
// should be created with the appropriate constructor first.
func (rs *EpNodeAccelRisers) discoverRemotePhase1() {
	for _, r := range rs.OIDs {
		r.discoverRemotePhase1()
	}
}

// Retrieves the NodeAccelRF from the parent Assembly.Assemblies array.
func (r *EpNodeAccelRiser) discoverRemotePhase1() {

	//Since the parent Assembly obj already has the array of NodeAccelRiser objects
	//the NodeAccelRiserRF field for this NodeAccelRiser just needs to be pulled from there
	//instead of retrieving it using an HTTP call
	if r.assemblyRF.AssemblyRF.Assemblies == nil {
		//this is a lookup error
		errlog.Printf("%s: No Assemblies array found in Parent.\n", r.OdataID)
		r.LastStatus = HTTPsGetFailed
		return
	}
	//If we got this far, then the EpAssembly call to discoverRemotePhase1 was successful
	r.LastStatus = HTTPsGetOk

	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", r.ParentOID, r.assemblyRF.AssemblyRaw)
	}

	//use r.RawOrdinal as the index to retrieve the NodeAccelRiser entry from the parent Assembly.Assemblies array,
	//and assign it to r.NodeAccelRiserRF
	if (len(r.assemblyRF.AssemblyRF.Assemblies) > r.RawOrdinal) && (r.assemblyRF.AssemblyRF.Assemblies[r.RawOrdinal] != nil) {
		r.NodeAccelRiserRF = r.assemblyRF.AssemblyRF.Assemblies[r.RawOrdinal]
	} else {
		//this is a lookup error
		errlog.Printf("%s: failure retrieving NodeAccelRiser from Assembly.Assemblies[%d].\n", r.OdataID, r.RawOrdinal)
		r.LastStatus = HTTPsGetFailed
		return
	}
	r.RedfishSubtype = NodeAccelRiserType

	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(r, "", "   ")
		errlog.Printf("%s: %s\n", r.NodeAccelRiserURL, jout)
	}

	r.LastStatus = VerifyingData
}

// This is the second discovery phase, after all information from
// the parent system has been gathered.  This is not intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the chassis first, so that it is available during later steps.
func (rs *EpNodeAccelRisers) discoverLocalPhase2() error {
	var savedError error
	for i, r := range rs.OIDs {
		r.discoverLocalPhase2()
		if r.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF NodeAccelRiser type not supported: %s",
				i, r.RedfishSubtype)
		} else if r.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, r.LastStatus)
			errlog.Printf("NodeAccelRisers discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual NodeAccelRiser.  Now that all information
// has been gathered, we can set the remaining fields needed to provide
// HMS with information about where the NodeAccelRiser is located
func (r *EpNodeAccelRiser) discoverLocalPhase2() {
	// Should never happen
	if r.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for odataID: %s\n",
			r.OdataID)
		r.LastStatus = EndpointInvalid
		return
	}
	if r.LastStatus != VerifyingData {
		return
	}

	r.Ordinal = r.epRF.getNodeAccelRiserOrdinal(r)
	r.Type = r.epRF.getNodeAccelRiserHMSType(r)
	r.ID = r.epRF.getNodeAccelRiserHMSID(r, r.Type, r.Ordinal)
	if r.NodeAccelRiserRF.Status.State != "Absent" {
		r.Status = "Populated"
		r.State = base.StatePopulated.String()
		r.Flag = base.FlagOK.String()
		generatedFRUID, err := GetNodeAccelRiserFRUID(r)
		if err != nil {
			errlog.Printf("FRUID Error: %s\n", err.Error())
			errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
		}
		r.FRUID = generatedFRUID
	} else {
		r.Status = "Empty"
		r.State = base.StateEmpty.String()
		//the state of the component is known (empty), it is not locked, does not have an alert or warning, so therefore Flag defaults to OK.
		r.Flag = base.FlagOK.String()
	}
	// Check if we have something valid to insert into the data store
	if (base.GetHMSType(r.ID) == base.NodeAccelRiser) && (r.Type == base.NodeAccelRiser.String()) {
		errlog.Printf("NodeAccelRiser discoverLocalPhase2: VALID xname ID ('%s') and Type ('%s') for: %s\n",
			r.ID, r.Type, r.NodeAccelRiserURL)
	} else {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s') for: %s\n",
			r.ID, r.Type, r.NodeAccelRiserURL)
		r.LastStatus = VerificationFailed
		return
	}
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(r, "", "   ")
		errlog.Printf("%s\n", jout)
		errlog.Printf("NodeAccelRiser ID: %s\n", r.ID)
		errlog.Printf("NodeAccelRiser FRUID: %s\n", r.FRUID)
	}
	r.LastStatus = DiscoverOK
}
