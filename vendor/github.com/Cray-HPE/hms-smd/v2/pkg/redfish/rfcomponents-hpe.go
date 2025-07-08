// MIT License
//
// (C) Copyright [2022-2023] Hewlett Packard Enterprise Development LP
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

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

/////////////////////////////////////////////////////////////////////////////
// Chassis - HpeDevice
/////////////////////////////////////////////////////////////////////////////

// This is the top-level Oem HpeDevice object for a particular Chassis.
type EpHpeDevice struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ComponentDescription

	BaseOdataID string `json:"BaseOdataID"`

	InventoryData

	DeviceURL   string `json:"deviceURL"` // Full URL to this RF Assembly obj
	ParentOID   string `json:"parentOID"`   // odata.id for parent
	ParentType  string `json:"parentType"`  // Chassis
	LastStatus  string `json:"LastStatus"`

	DeviceRF  HpeDevice `json:"DeviceRF"`
	deviceRaw *json.RawMessage

	systemRF *EpSystem  // Backpointer to associated system.
	epRF     *RedfishEP // Backpointer to RF EP, for connection details, etc.
}

// Set of EpHpeDevices, each representing a Redfish HPE OEM device (possibly a GPU)
// listed under a Redfish Chassis.
type EpHpeDevices struct {
	Num  int                     `json:"num"`
	OIDs map[string]*EpHpeDevice `json:"oids"`
}

// Initializes EpHpeDevice struct with minimal information needed to
// pass along to its children.
func NewEpHpeDevice(s *EpSystem, odataID ResourceID, pOID, pType string, rawOrdinal int) *EpHpeDevice {
	d := new(EpHpeDevice)
	d.OdataID = odataID.Oid
	d.Type = HpeDeviceType
	d.BaseOdataID = odataID.Basename()
	d.RedfishType = HpeDeviceType
	d.RfEndpointID = s.epRF.ID

	d.DeviceURL = s.epRF.FQDN + odataID.Oid
	d.ParentOID = pOID
	d.ParentType = pType

	d.Ordinal = -1
	d.RawOrdinal = rawOrdinal

	d.LastStatus = NotYetQueried
	d.systemRF = s
	d.epRF = s.epRF

	return d
}

// Makes contact with redfish endpoint to discover information about
// all HPE Devices for a given Redfish Chassis. EpHpeDevice entries
// should be created with the appropriate constructor first.
func (ds *EpHpeDevices) discoverRemotePhase1() {
	for _, d := range ds.OIDs {
		d.discoverRemotePhase1()
	}
}

// Makes contact with redfish endpoint to discover information about
// a particular HPE Device under a Chassis.  Note that the
// EpHpeDevice should be created with the appropriate constructor first.
func (d *EpHpeDevice) discoverRemotePhase1() {
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
	d.deviceRaw = &urlJSON
	d.LastStatus = HTTPsGetOk

	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", url, urlJSON)
	}
	// Decode JSON into Processor structure containing Redfish data
	if err := json.Unmarshal(urlJSON, &d.DeviceRF); err != nil {
		if IsUnmarshalTypeError(err) {
			errlog.Printf("bad field(s) skipped: %s: %s\n", url, err)
		} else {
			errlog.Printf("ERROR: json decode failed: %s: %s\n", url, err)
			d.LastStatus = EPResponseFailedDecode
			return
		}
	}
	d.RedfishSubtype = d.DeviceRF.DeviceType

	// Workaround CASMHMS-4951 GPUs missing Model and Partnumber.
	// Use Name in place of Model and ProductPartNumber in place
	// of PartNumber.
	if d.DeviceRF.PartNumber == "" {
		d.DeviceRF.PartNumber = d.DeviceRF.ProductPartNumber
	}
	if d.DeviceRF.Model == "" {
		d.DeviceRF.Model = d.DeviceRF.Name
	}

	if rfVerbose > 0 {
		jout, _ := json.MarshalIndent(d, "", "   ")
		errlog.Printf("%s: %s\n", url, jout)
	}

	d.LastStatus = VerifyingData
}

// This is the second discovery phase, after all information from
// the parent chassis has been gathered. This is not intended to
// be run as a separate step; it is separate because certain discovery
// activities may require that information is gathered for all components
// under the chassis first, so that it is available during later steps.
func (ds *EpHpeDevices) discoverLocalPhase2() error {
	var savedError error
	for i, d := range ds.OIDs {
		d.discoverLocalPhase2()
		if d.LastStatus == RedfishSubtypeNoSupport {
			errlog.Printf("Key %s: RF Processor type not supported: %s",
				i, d.RedfishSubtype)
		} else if d.LastStatus != DiscoverOK {
			err := fmt.Errorf("Key %s: %s", i, d.LastStatus)
			errlog.Printf("Proccesors discoverLocalPhase2: saw error: %s", err)
			savedError = err
		}
	}
	return savedError
}

// Phase2 discovery for an individual HPE device. Now that all information
// has been gathered, we can set the remaining fields needed to provide
// HMS with information about where the HPE device is located
func (d *EpHpeDevice) discoverLocalPhase2() {
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

	// GPUs are under HPE devices on HPE hardware
	d.Ordinal = d.epRF.getHpeDeviceOrdinal(d)
	if strings.ToLower(d.RedfishSubtype) == "gpu" &&
	   !strings.Contains(strings.ToLower(d.DeviceRF.Name), "switch") &&
	   !strings.Contains(strings.ToLower(d.DeviceRF.Location), "baseboard") {
		d.ID = d.systemRF.ID + "a" + strconv.Itoa(d.Ordinal)
		d.Type = xnametypes.NodeAccel.String()
	} else if strings.Contains(strings.ToLower(d.RedfishSubtype), "nic") &&
	          (strings.Contains(strings.ToLower(d.DeviceRF.Manufacturer), "mellanox") ||
	           strings.Contains(strings.ToLower(d.DeviceRF.Manufacturer), "hpe") ||
	           strings.Contains(strings.ToLower(d.DeviceRF.Manufacturer), "bei")) {
		// Accept Mellanox or Cassini HSN NICs so we ignore non-HSN NICs.
		// Cassini shows as HPE instead of BEI in Proliant iLO redfish
		// implementations so we check for both just incase this changes
		// in the future.
		d.ID = d.systemRF.ID + "h" + strconv.Itoa(d.Ordinal)
		d.Type = xnametypes.NodeHsnNic.String()
	} else {
		// What to do with non-GPUs, trash for now?
		d.Type = xnametypes.HMSTypeInvalid.String()
		d.LastStatus = RedfishSubtypeNoSupport
		return
	}
	if d.DeviceRF.Status.State != "Absent" {
		d.Status = "Populated"
		d.State = base.StatePopulated.String()
		d.Flag = base.FlagOK.String()
		generatedFRUID, err := GetHpeDeviceFRUID(d)
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
	if (xnametypes.GetHMSType(d.ID) != xnametypes.NodeAccel || d.Type != xnametypes.NodeAccel.String()) &&
	   (xnametypes.GetHMSType(d.ID) != xnametypes.NodeHsnNic || d.Type != xnametypes.NodeHsnNic.String()) {
		errlog.Printf("Error: Bad xname ID ('%s') or Type ('%s') for: %s\n", d.ID, d.Type, d.DeviceURL)
		d.LastStatus = VerificationFailed
		return
	}

	errlog.Printf("HPE Device xname ID ('%s') and Type ('%s') for: %s\n", d.ID, d.Type, d.DeviceURL)
	d.LastStatus = DiscoverOK
}

// Determined based on discovered info and original list order that the
// HpeDevice ordinal is.
func (ep *RedfishEP) getHpeDeviceOrdinal(d *EpHpeDevice) int {
	//The position of any HPE device in relation to its siblings is indicated
	//by the basename of its OdataID, so it is possible to retrieve and sort the keys of the
	//chassis HPE device OIDS map to determine the proper Ordinal of any particular HPE device
	var ordinal = d.RawOrdinal
	if len(d.systemRF.HpeDevices.OIDs) > 0 {
		dsOIDs := make([]string, 0, len(d.systemRF.HpeDevices.OIDs))
		for oid, device := range d.systemRF.HpeDevices.OIDs {
			// Get only devices of the same type
			if strings.ToLower(device.RedfishSubtype) == strings.ToLower(d.RedfishSubtype) {
				if strings.Contains(strings.ToLower(d.RedfishSubtype), "nic") {
					// Accept Mellanox or Cassini HSN NICs so we ignore non-HSN NICs.
					// Cassini shows as HPE instead of BEI in Proliant iLO redfish
					// implementations so we check for both just incase this changes
					// in the future. 
					if strings.Contains(strings.ToLower(device.DeviceRF.Manufacturer), "mellanox") ||
					   strings.Contains(strings.ToLower(device.DeviceRF.Manufacturer), "hpe") ||
					   strings.Contains(strings.ToLower(device.DeviceRF.Manufacturer), "bei") {
						dsOIDs = append(dsOIDs, oid)
					}
				} else {
					dsOIDs = append(dsOIDs, oid)
				}
			}
		}
		//sort the OIDs
		sort.Strings(dsOIDs)
		//the proper ordinal for this Node Accel Riser is now the position of its OdataID in the rsOIDs slice
		for i, dsOID := range dsOIDs {
			if dsOID == d.BaseOdataID {
				ordinal = i
				break
			}
		}
	}
	return ordinal
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetHpeDeviceFRUID(d *EpHpeDevice) (fruid string, err error) {
	return getStandardFRUID(d.Type, d.ID, d.DeviceRF.Manufacturer, d.DeviceRF.PartNumber, d.DeviceRF.SerialNumber)
}
