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
// Chassis - NetworkAdapters
/////////////////////////////////////////////////////////////////////////////

// Set of EpNetworkAdapters, each representing a NetworkAdapter (HSN NIC, etc)
// listed under a Redfish Chassis.
type EpNetworkAdapters struct {
	Num  int                          `json:"num"`
	OIDs map[string]*EpNetworkAdapter `json:"oids"`
}

// This is one of possibly several NetworkAdapters for a particular
// EpChassis (Redfish "Chassis").
type EpNetworkAdapter struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	NetworkAdapterURL string `json:"nodeAccelRiserURL"` // Full URL to this RF NetworkAdapter obj
	ParentOID         string `json:"parentOID"`         // odata.id for parent
	ParentType        string `json:"parentType"`        // Chassis
	LastStatus        string `json:"LastStatus"`

	NetworkAdapterRF  *NetworkAdapter `json:"NetworkAdapterRF"`
	networkAdapterRaw *json.RawMessage

	epRF     *RedfishEP // Backpointer to RF EP, for connection details, etc.
	systemRF *EpSystem  // Backpointer to the associated system.
}

// Initializes EpNetworkAdapter struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the NetworkAdapter to
// look at.  This should be the only way this struct is created for
// NetworkAdapters under a chassis.
func NewEpNetworkAdapter(s *EpSystem, poid, pType string, odataID ResourceID, rawOrdinal int) *EpNetworkAdapter {
	na := new(EpNetworkAdapter)
	na.OdataID = odataID.Oid

	na.Type = NetworkAdapterType
	na.BaseOdataID = odataID.Basename()
	na.RedfishType = NetworkAdapterType
	na.RfEndpointID = s.epRF.ID

	na.NetworkAdapterURL = s.epRF.FQDN + odataID.Oid
	na.ParentOID = poid
	na.ParentType = pType

	na.Ordinal = -1
	na.RawOrdinal = rawOrdinal

	na.LastStatus = NotYetQueried
	na.epRF = s.epRF
	na.systemRF = s

	return na
}

// Makes contact with redfish endpoint to discover information about
// all NetworkAdapters for a given Redfish Chassis.  EpNetworkAdapter
// entries should be created with the appropriate constructor first.
func (nas *EpNetworkAdapters) discoverRemotePhase1() {
	for _, na := range nas.OIDs {
		na.discoverRemotePhase1()
	}
}

// Makes contact with redfish endpoint to discover information about
// a particular NetworkAdapter under a redfish Chassis.  Note that the
// EpNetworkAdapter should be created with the appropriate constructor
// first.
func (na *EpNetworkAdapter) discoverRemotePhase1() {
	rpath := na.OdataID
	url := na.epRF.FQDN + rpath
	urlJSON, err := na.epRF.GETRelative(rpath)
	if err != nil || urlJSON == nil {
		na.LastStatus = HTTPsGetFailed
		return
	}
	na.networkAdapterRaw = &urlJSON
	na.LastStatus = HTTPsGetOk

	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, urlJSON)
	}
	// Decode JSON into Drive structure containing Redfish data
	if err := json.Unmarshal(urlJSON, &na.NetworkAdapterRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			na.LastStatus = EPResponseFailedDecode
			return
		}
	}
	na.RedfishSubtype = NetworkAdapterType

	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(na, "", "   ")
		errlog.Printf("%s: %s\n", url, jout)
	}

	na.LastStatus = VerifyingData
}

// This is the second discovery phase, after all information from
// the parent system has been gathered.  This is not intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the chassis first, so that it is available during later steps.
func (nas *EpNetworkAdapters) discoverLocalPhase2() error {
	var savedError error
	for i, na := range nas.OIDs {
		na.discoverLocalPhase2()
		if na.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF NetworkAdapter type not supported: %s",
				i, na.RedfishSubtype)
		} else if na.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, na.LastStatus)
			errlog.Printf("NetworkAdapter discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual NetworkAdapter.  Now that all information
// has been gathered, we can set the remaining fields needed to provide
// HMS with information about where the NetworkAdapter is located
func (na *EpNetworkAdapter) discoverLocalPhase2() {
	// Should never happen
	if na.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for odataID: %s\n",
			na.OdataID)
		na.LastStatus = EndpointInvalid
		return
	}
	if na.LastStatus != VerifyingData {
		return
	}

	na.Ordinal = na.epRF.getNetworkAdapterOrdinal(na)
	na.Type = na.epRF.getNetworkAdapterHMSType(na)
	na.ID = na.epRF.getNetworkAdapterHMSID(na, na.Type, na.Ordinal)
	na.Status = "Populated"
	na.State = base.StatePopulated.String()
	na.Flag = base.FlagOK.String()
	generatedFRUID, err := GetNetworkAdapterFRUID(na)
	if err != nil {
		errlog.Printf("FRUID Error: %s\n", err.Error())
		errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
	}
	na.FRUID = generatedFRUID

	// Check if we have something valid to insert into the data store
	if (base.GetHMSType(na.ID) == base.NodeHsnNic) && (na.Type == base.NodeHsnNic.String()) {
		errlog.Printf("NetworkAdapter discoverLocalPhase2: VALID xname ID ('%s') and Type ('%s') for: %s\n",
			na.ID, na.Type, na.NetworkAdapterURL)
	} else {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s') for: %s\n",
			na.ID, na.Type, na.NetworkAdapterURL)
		na.LastStatus = VerificationFailed
		return
	}
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(na, "", "   ")
		errlog.Printf("%s\n", jout)
		errlog.Printf("NodeHsnNic ID: %s\n", na.ID)
		errlog.Printf("NodeHsnNic FRUID: %s\n", na.FRUID)
	}
	na.LastStatus = DiscoverOK
}
