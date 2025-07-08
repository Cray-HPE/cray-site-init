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

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

/////////////////////////////////////////////////////////////////////////////
// ComputerSystem - Storage
/////////////////////////////////////////////////////////////////////////////

// This is the top-level Storage object for a particular
// ComputerSystem.
type EpStorage struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	StorageURL string `json:"storageCollectionURL"` // Full URL to this RF Storage obj
	ParentOID  string `json:"parentOID"`            // odata.id for parent
	ParentType string `json:"parentType"`           // ComputerSystem or Manager
	LastStatus string `json:"LastStatus"`

	StorageRF  Storage `json:"StorageRF"`
	StorageRaw *json.RawMessage

	epRF  *RedfishEP // Backpointer to RF EP, for connection details, etc.
	sysRF *EpSystem  // Backpointer to parent system.
}

// Initializes EpStorage struct with minimal information needed to
// pass along to its children.
func NewEpStorage(sys *EpSystem, odataID ResourceID) *EpStorage {
	s := new(EpStorage)
	s.OdataID = odataID.Oid
	s.Type = "" //should be base.Storage.String()
	s.BaseOdataID = odataID.Basename()
	s.RedfishType = "" //should be StorageType
	s.RfEndpointID = sys.epRF.ID

	s.StorageURL = sys.epRF.FQDN + odataID.Oid
	s.ParentOID = sys.OdataID
	s.ParentType = ComputerSystemType

	s.LastStatus = NotYetQueried
	s.epRF = sys.epRF
	s.sysRF = sys

	return s
}

// This is one of possibly several Storage Collections for a particular
// ComputerSystem's Storage object.
type EpStorageCollection struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	Ordinal    int `json:"Ordinal"`
	RawOrdinal int `json:"-"`

	StorageCollectionURL string `json:"storageCollectionURL"` // Full URL to this RF StorageCollection obj
	ParentOID            string `json:"parentOID"`            // odata.id for parent
	ParentType           string `json:"parentType"`           // ComputerSystem or Manager
	LastStatus           string `json:"LastStatus"`

	StorageCollectionRF  StorageCollection `json:"StorageCollectionRF"`
	StorageCollectionRaw *json.RawMessage

	epRF      *RedfishEP // Backpointer to RF EP, for connection details, etc.
	sysRF     *EpSystem  // Backpointer to parent system.
	storageRF *EpStorage // Backpointer to parent storage object
}

// Set of EpStorageCollection objects, each representing a Redfish "StorageCollection"
// listed under a Redfish ComputerSystem's Storage object.
type EpStorageCollections struct {
	Num  int                             `json:"num"`
	OIDs map[string]*EpStorageCollection `json:"oids"`
}

// Initializes EpStorageCollection struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the StorageCollection to
// look at.
func NewEpStorageCollection(s *EpStorage, odataID ResourceID, rawOrdinal int) *EpStorageCollection {
	sc := new(EpStorageCollection)
	sc.OdataID = odataID.Oid
	sc.Type = xnametypes.StorageGroup.String()
	sc.BaseOdataID = odataID.Basename()
	sc.RedfishType = StorageGroupType
	sc.RfEndpointID = s.epRF.ID

	sc.StorageCollectionURL = s.epRF.FQDN + odataID.Oid
	sc.ParentOID = s.OdataID
	sc.ParentType = ComputerSystemType

	sc.Ordinal = -1
	sc.RawOrdinal = rawOrdinal

	sc.LastStatus = NotYetQueried

	sc.epRF = s.epRF
	sc.sysRF = s.sysRF
	sc.storageRF = s

	return sc
}

// Makes contact with redfish endpoint to discover information about
// all Drives for a given Redfish System.  EpDrive entries
// should be created with the appropriate constructor first.
func (cs *EpStorageCollections) discoverRemotePhase1() {
	for _, c := range cs.OIDs {
		c.discoverRemotePhase1()
	}
}

// Makes contact with redfish endpoint to discover information about
// a particular drive under a ComputerSystem aka System.   Note that the
// EpDrive should be created with the appropriate constructor first.
func (c *EpStorageCollection) discoverRemotePhase1() {
	//the job of this function is to retrieve the list of Drives
	//for this particular storage collection
	//and populate the EpDrives collection
	rpath := c.OdataID
	url := c.epRF.FQDN + rpath
	urlJSON, err := c.epRF.GETRelative(rpath)
	if err != nil || urlJSON == nil {
		if err == ErrRFDiscURLNotFound {
			errlog.Printf("%s: Redfish bug! Link %s was dead (404).  "+
				"Will try to continue.  No component will be created.",
				c.epRF.ID, rpath)
			c.LastStatus = RedfishSubtypeNoSupport
			c.RedfishSubtype = RFSubtypeUnknown
		} else {
			c.LastStatus = HTTPsGetFailed
		}
		return
	}
	c.StorageCollectionRaw = &urlJSON
	c.LastStatus = HTTPsGetOk

	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, urlJSON)
	}
	// Decode JSON into Drive structure containing Redfish data
	if err := json.Unmarshal(urlJSON, &c.StorageCollectionRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			c.LastStatus = EPResponseFailedDecode
			return
		}
	}
	//examine the Drives array, and push each Drive
	//onto the parent system's Drives collection (c.sysRF.Drives.OIDs)
	sort.Sort(ResourceIDSlice(c.StorageCollectionRF.Drives))
	for i, dOID := range c.StorageCollectionRF.Drives {
		c.sysRF.Drives.OIDs[dOID.Oid] = NewEpDrive(c, dOID, i)
		c.sysRF.Drives.Num = c.sysRF.Drives.Num + 1
	}
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(c, "", "   ")
		errlog.Printf("%s: %s\n", url, jout)
	}

	c.LastStatus = VerifyingData
}

/////////////////////////////////////////////////////////////////////////////
// ComputerSystem - Drives
/////////////////////////////////////////////////////////////////////////////

// This is one of possibly several drives for a particular
// EpSystem(Redfish "ComputerSystem" or just "System").
type EpDrive struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	// Embedded struct - Locational/FRU, state, and status info
	InventoryData

	DriveURL   string `json:"driveURL"`   // Full URL to this RF Drive obj
	ParentOID  string `json:"parentOID"`  // odata.id for parent
	ParentType string `json:"parentType"` // ComputerSystem or Manager
	LastStatus string `json:"LastStatus"`

	DriveRF  Drive `json:"DriveRF"`
	driveRaw *json.RawMessage

	epRF                *RedfishEP           // Backpointer to RF EP, for connection details, etc.
	sysRF               *EpSystem            // Backpointer to parent system.
	storageCollectionRF *EpStorageCollection // Backpointer to parent StorageCollection.
}

// Set of EpDrive, each representing a Redfish "Drive"
// listed under a Redfish ComputerSystem aka System.  Unlike Systems/Chassis,
// there is no top-level listing of all drives under an RF endpoint, i.e.
// each one is specific to a collection belonging to a single parent System.
type EpDrives struct {
	Num  int                 `json:"num"`
	OIDs map[string]*EpDrive `json:"oids"`
}

// Initializes EpDrive struct with minimal information needed to
// discover it, i.e. endpoint info and the odataID of the Drive to
// look at.  This should be the only way this struct is created for
// Drives under a system.
func NewEpDrive(s *EpStorageCollection, odataID ResourceID, rawOrdinal int) *EpDrive {
	d := new(EpDrive)
	d.OdataID = odataID.Oid
	d.Type = xnametypes.Drive.String()
	d.BaseOdataID = odataID.Basename()
	d.RedfishType = DriveType
	d.RfEndpointID = s.epRF.ID

	d.DriveURL = s.epRF.FQDN + odataID.Oid
	//TODO: review this
	d.ParentOID = s.OdataID
	d.ParentType = ComputerSystemType

	d.Ordinal = -1
	d.RawOrdinal = rawOrdinal

	d.LastStatus = NotYetQueried
	d.epRF = s.epRF
	d.sysRF = s.sysRF
	d.storageCollectionRF = s

	return d
}

// Makes contact with redfish endpoint to discover information about
// all Drives for a given Redfish System.  EpDrive entries
// should be created with the appropriate constructor first.
func (ds *EpDrives) discoverRemotePhase1() {
	for _, d := range ds.OIDs {
		d.discoverRemotePhase1()
	}
}

// Makes contact with redfish endpoint to discover information about
// a particular drive under a ComputerSystem aka System.   Note that the
// EpDrive should be created with the appropriate constructor first.
func (d *EpDrive) discoverRemotePhase1() {
	rpath := d.OdataID
	url := d.epRF.FQDN + rpath
	urlJSON, err := d.epRF.GETRelative(rpath)
	if err != nil || urlJSON == nil {
		if err == ErrRFDiscURLNotFound {
			errlog.Printf("%s: Redfish bug! Link %s was dead (404).  "+
				"Will try to continue.  No component will be created.",
				d.epRF.ID, rpath)
			d.LastStatus = RedfishSubtypeNoSupport
			d.RedfishSubtype = RFSubtypeUnknown
		} else {
			d.LastStatus = HTTPsGetFailed
		}
		return
	}
	d.driveRaw = &urlJSON
	d.LastStatus = HTTPsGetOk

	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, urlJSON)
	}
	// Decode JSON into Drive structure containing Redfish data
	if err := json.Unmarshal(urlJSON, &d.DriveRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			d.LastStatus = EPResponseFailedDecode
			return
		}
	}
	d.RedfishSubtype = d.DriveRF.MediaType

	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(d, "", "   ")
		errlog.Printf("%s: %s\n", url, jout)
	}

	d.LastStatus = VerifyingData
}

// This is the second discovery phase, after all information from
// the parent system has been gathered.  This is not intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the system first, so that it is available during later steps.
func (ds *EpDrives) discoverLocalPhase2() error {
	var savedError error
	for i, d := range ds.OIDs {
		d.discoverLocalPhase2()
		if d.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF Drive type not supported: %s",
				i, d.RedfishSubtype)
		} else if d.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, d.LastStatus)
			errlog.Printf("Drives discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual drive.  Now that all information
// has been gathered, we can set the remaining fields needed to provide
// HMS with information about where the drive is located
func (d *EpDrive) discoverLocalPhase2() {
	// Should never happen
	if d.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for odataID: %s\n",
			d.OdataID)
		d.LastStatus = EndpointInvalid
		return
	}
	if d.LastStatus != VerifyingData {
		return
	}

	d.Ordinal = d.epRF.getDriveOrdinal(d)
	//this is the xname
	//TODO: get StorageGroup ordinal from ParentOID
	d.ID = d.sysRF.ID + "g" + strconv.Itoa(d.epRF.getStorageCollectionOrdinal(d.storageCollectionRF)) + "k" + strconv.Itoa(d.Ordinal)
	if d.DriveRF.Status.State != "Absent" {
		d.Status = "Populated"
		d.State = base.StatePopulated.String()
		d.Flag = base.FlagOK.String()
		generatedFRUID, err := GetDriveFRUID(d)
		if err != nil {
			errlog.Printf("FRUID Error: %s\n", err.Error())
			errlog.Printf("Using untrackable FRUID: %s\n", generatedFRUID)
		}
		d.FRUID = generatedFRUID
	} else {
		d.Status = "Empty"
		d.State = base.StateEmpty.String()
		//the state of the component is known (empty), it is not locked, does not have an alert or warning, so therefore Flag defaults to OK.
		d.Flag = base.FlagOK.String()
	}
	// Check if we have something valid to insert into the data store
	if xnametypes.GetHMSType(d.ID) != xnametypes.Drive ||
		d.Type != xnametypes.Drive.String() {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s') for: %s\n",
			d.ID, d.Type, d.DriveURL)
		d.LastStatus = VerificationFailed
		return
	}
	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(d, "", "   ")
		errlog.Printf("%s\n", jout)
		errlog.Printf("Drive ID: %s\n", d.ID)
		errlog.Printf("Drive FRUID: %s\n", d.FRUID)
	}
	d.LastStatus = DiscoverOK
}
