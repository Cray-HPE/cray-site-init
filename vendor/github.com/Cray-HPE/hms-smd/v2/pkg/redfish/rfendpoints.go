// MIT License
//
// (C) Copyright [2019-2024-2025] Hewlett Packard Enterprise Development LP
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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-certs/pkg/hms_certs"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

const PKG_VERSION = "0.2"

// Error codes for problems obtaining data from remote RF endpoints.
const (
	DiscoveryStarted        = "DiscoveryStarted"
	EndpointInvalid         = "EndpointInvalid"
	EPResponseFailedDecode  = "EPResponseFailedDecode"
	HTTPsGetFailed          = "HTTPsGetFailed"
	HTTPsGetOk              = "HTTPsGetOk"
	NoEthIfacesFound        = "NoEthIfacesFound"
	NotYetQueried           = "NotYetQueried"
	VerifyingData           = "VerifyingData"
	VerificationFailed      = "VerificationFailed"
	ChildVerificationFailed = "ChildVerificationFailed"

	RedfishSubtypeNoSupport  = "RedfishSubtypeNoSupport"
	EndpointTypeNotSupported = "EndpointTypeNotSupported"
	EndpointNotEnabled       = "EndpointNotEnabled"
	DiscoverOK               = "DiscoverOK"

	StoreFailed             = "StoreFailed"
	UnexpectedErrorPreStore = "UnexpectedErrorPreStore"
)

// These are types of structures in rfendpoints that are built upon
// the underlying Redfish type of the same name.
const (
	ServiceRootType       = "ServiceRoot"
	ChassisType           = "Chassis"
	ComputerSystemType    = "ComputerSystem"
	EthernetInterfaceType = "EthernetInterface"
	ManagerType           = "Manager"
	MemoryType            = "Memory"
	ProcessorType         = "Processor"
	DriveType             = "Drive"
	StorageGroupType      = "StorageGroup"
	PowerSupplyType       = "PowerSupply"
	PowerType             = "Power"
	NodeAccelRiserType    = "GPUSubsystem"
	AssemblyType          = "Assembly"
	HpeDeviceType         = "HpeDevice"
	OutletType            = "Outlet"
	PDUType               = "PowerDistribution"
	NetworkAdapterType    = "NetworkAdapter"
	AccountServiceType    = "AccountService"
	EventServiceType      = "EventService"
	LogServiceType        = "LogService"
	SessionServiceType    = "SessionService"
	TaskServiceType       = "TaskService"
	UpdateServiceType     = "UpdateService"
)

// Redfish object subtypes, i.e. {type-name}Type,
// For example, ChassisType, ManagerType, etc.
const (
	RFSubtypeRack       = "Rack"
	RFSubtypeBlade      = "Blade"
	RFSubtypeEnclosure  = "Enclosure"
	RFSubtypeStandAlone = "StandAlone"
	RFSubtypeRackMount  = "RackMount"
	RFSubtypeCard       = "Card"
	RFSubtypeCartridge  = "Cartridge"
	RFSubtypeRow        = "Row"
	RFSubtypePod        = "Pod"
	RFSubtypeExpansion  = "Expansion"
	RFSubtypeSidecar    = "Sidecar"
	RFSubtypeZone       = "Zone"
	RFSubtypeSled       = "Sled"
	RFSubtypeShelf      = "Shelf"
	RFSubtypeDrawer     = "Drawer"
	RFSubtypeModule     = "Module"
	RFSubtypeComponent  = "Component"
	RFSubtypeOther      = "Other"

	RFSubtypePhysical              = "Physical"
	RFSubtypeVirtual               = "Virtual"
	RFSubtypeOS                    = "OS"
	RFSubtypePhysicallyPartitioned = "PhysicallyPartitioned"
	RFSubtypeVirtuallyPartitioned  = "VirtuallyPartitioned"

	RFSubtypeManagementController = "ManagementController"
	RFSubtypeEnclosureManager     = "EnclosureManager"
	RFSubtypeBMC                  = "BMC"
	RFSubtypeRackManager          = "RackManager"
	RFSubtypeAuxiliaryController  = "AuxiliaryController"

	// PDU Types
	RFSubtypeRackPDU                 = "RackPDU"
	RFSubtypeFloorPDU                = "FloorPDU"
	RFSubtypeManualTransferSwitch    = "ManualTransferSwitch"
	RFSubtypeAutomaticTransferSwitch = "AutomaticTransferSwitch"

	// Outlet types
	RFSubtypeOutNEMA_5_15R       = "NEMA_5_15R"
	RFSubtypeOutNEMA_5_20R       = "NEMA_5_20R"
	RFSubtypeOutNEMA_L5_20R      = "NEMA_L5_20R"
	RFSubtypeOutNEMA_L5_30R      = "NEMA_L5_30R"
	RFSubtypeOutNEMA_L6_20R      = "NEMA_L6_20R"
	RFSubtypeOutNEMA_L6_30R      = "NEMA_L6_30R"
	RFSubtypeOutC13              = "C13"
	RFSubtypeOutC15              = "C15"
	RFSubtypeOutC19              = "C19"
	RFSubtypeOutCEE_7_Type_E     = "CEE_7_Type_E"
	RFSubtypeOutCEE_7_Type_F     = "CEE_7_Type_F"
	RFSubtypeOutSEV_1011_TYPE_12 = "SEV_1011_TYPE_12"
	RFSubtypeOutSEV_1011_TYPE_23 = "SEV_1011_TYPE_23"
	RFSubtypeOutBS_1363_Type_G   = "BS_1363_Type_G"

	RFSubtypeUnknown = "Unknown" // Not found/error
)

const MaxFanout int = 1000

var ErrRFDiscFQDNMissing = errors.New("FQDN unexpectedly empty string")
var ErrRFDiscURLNotFound = errors.New("URL request returned 404: Not Found")
var ErrRFDiscILOLicenseReq = errors.New("iLO License Required")

/////////////////////////////////////////////////////////////////////////////
//
// RedfishEndpoint creation
//
/////////////////////////////////////////////////////////////////////////////

//
// This is a scan-friendly version of RedfishEndpoint that can be used to
// create or update an entry (or at least the writable fields set by the user).
// The redfish endpoint addresses can be given variously as hostname + domain,
// FQDN, or IP addresses.  They are "preferred" in this order.  If the hostname
// and domain are not provided, they will be obtained from the FQDN, however
// if the hostname is provided, the FQDN must match.  If both hostname and
// domain are provided, they override any FQDN values or an error occurs.
//
// TODO: In the future there will be some process to autmatically add these
//       for endpoints that advertise their presence, e.g. via SSDP, but we
//       will likely always need the ability to add manual endpoints that do
//       not.  Those that do advertise will likely just need a generic
//       identifier, e.g. a domain with no specific host info, or perhaps a
//       subnet.
type RawRedfishEP struct {
	ID             string `json:"ID"`
	Type           string `json:"Type"`
	Name           string `json:"Name"` // user supplied descriptive name
	Hostname       string `json:"Hostname"`
	Domain         string `json:"Domain"`
	FQDN           string `json:"FQDN"`
	Enabled        *bool  `json:"Enabled"`
	UUID           string `json:"UUID"`
	User           string `json:"User"`
	Password       string `json:"Password"`
	UseSSDP        *bool  `json:"UseSSDP"`
	MACRequired    *bool  `json:"MACRequired"`
	MACAddr        string `json:"MACAddr"`
	IPAddr         string `json:"IPAddress"`
	RediscOnUpdate *bool  `json:"RediscoverOnUpdate"`
	TemplateID     string `json:"TemplateID"`
}

// String function to redact passwords from any kind of output
func (rrep RawRedfishEP) String() string {
	// NOTE: the value form is slightly less efficient since it involves a
	//  copy, but it will work for both pass by value and pass by pointer.
	buf := bytes.NewBufferString("{")
	fmt.Fprintf(buf, "ID: %s, ", rrep.ID)
	fmt.Fprintf(buf, "Type: %s, ", rrep.Type)
	fmt.Fprintf(buf, "Name: %s, ", rrep.Name)
	fmt.Fprintf(buf, "Hostname: %s, ", rrep.Hostname)
	fmt.Fprintf(buf, "Domain: %s, ", rrep.Domain)
	fmt.Fprintf(buf, "FQDN: %s, ", rrep.FQDN)
	if rrep.Enabled == nil {
		fmt.Fprintf(buf, "Enabled: nil, ")
	} else {
		fmt.Fprintf(buf, "Enabled: %t, ", *rrep.Enabled)
	}
	fmt.Fprintf(buf, "UUID: %s, ", rrep.UUID)
	fmt.Fprintf(buf, "User: %s, ", rrep.User)
	fmt.Fprintf(buf, "Password: <REDACTED>, ")
	if rrep.UseSSDP == nil {
		fmt.Fprintf(buf, "UseSSDP: nil, ")
	} else {
		fmt.Fprintf(buf, "UseSSDP: %t, ", *rrep.UseSSDP)
	}
	if rrep.MACRequired == nil {
		fmt.Fprintf(buf, "MACRequired: nil, ")
	} else {
		fmt.Fprintf(buf, "MACRequired: %t, ", *rrep.MACRequired)
	}
	fmt.Fprintf(buf, "MACAddr: %s, ", rrep.MACAddr)
	fmt.Fprintf(buf, "IPAddress: %s, ", rrep.IPAddr)
	if rrep.RediscOnUpdate == nil {
		fmt.Fprintf(buf, "RediscOnUpdate: nil, ")
	} else {
		fmt.Fprintf(buf, "RediscOnUpdate: %t, ", *rrep.RediscOnUpdate)
	}
	fmt.Fprintf(buf, "TemplateID: %s, ", rrep.TemplateID)
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// Defaults for RedfishEndpoint properties
const (
	EnabledDefault        = true
	UseSSDPDefault        = false
	MACRequiredDefault    = false
	RediscOnUpdateDefault = false
)

// JSON-friendly array of RawRedfishEP entries
type RawRedfishEPs struct {
	RedfishEndpoints []RawRedfishEP `json:"RedfishEndpoints"`
}

// Create RedfishEPDescription from unvalidated input, e.g. provided by
// the user.  Fields that were omitted can be populated with default values.
func NewRedfishEPDescription(rep *RawRedfishEP) (*RedfishEPDescription, error) {
	if rep == nil {
		err := fmt.Errorf("got nil pointer")
		return nil, err
	}
	ep := new(RedfishEPDescription)
	//
	// Figure out the FQDN to contact the endpoint at.
	// Use Hostname and Domain fields first.
	//
	if rep.Domain != "" {
		ep.Domain = strings.Trim(rep.Domain, "./ ")
	}
	hostIsIP := false
	if rep.Hostname != "" {
		ep.Hostname = strings.Trim(rep.Hostname, "./ ")
		ipHost := GetIPAddressString(ep.Hostname)
		if ipHost != "" {
			hostIsIP = true
			ep.Hostname = ipHost
		} else {
			splitHost := strings.SplitN(ep.Hostname, ".", 2)
			if len(splitHost) > 1 {
				err := fmt.Errorf("Hostname is not a basename or IP address.")
				return nil, err
			}
		}
	}
	// If a FQDN was given instead, use those to fill in the hostname
	// or domain if either was missing.  But we must not have a mismatch.
	if rep.FQDN != "" {
		// If the FQDN is an IP address, don't treat it as host.domain.
		// Just treat it as a single hostname.  We don't really want to
		// use IP addresses, but if that's all we have, it will work as
		// long as ID is set to an xname.
		fqdnIP := GetIPAddressString(rep.FQDN)
		if fqdnIP != "" && ep.Hostname == "" {
			// Just use the IP as the Hostname/FQDN with no domain.
			ep.Hostname = fqdnIP
			hostIsIP = true
		}
		// Assume FQDN is in normal host.domain1.domain2.etc format.
		fqdnHost := ""
		fqdnDomain := ""
		splitFQDN := strings.SplitN(rep.FQDN, ".", 2)
		fqdnHost = strings.Trim(splitFQDN[0], "./ ")
		if len(splitFQDN) > 1 {
			fqdnDomain = strings.Trim(splitFQDN[1], "./ ")
		}
		if ep.Hostname == "" {
			ep.Hostname = fqdnHost
		}
		if ep.Domain == "" {
			if ep.Hostname == fqdnHost {
				ep.Domain = fqdnDomain
			}
		}
	}

	//
	// Set fields in RedfishEndpointDescription from its raw equivalent.
	//

	// ID must be given, or the Hostname will be used instead.
	// The difference is the ID must be an xname, the hostname
	// need not be (unless the ID is not given and the hostname is used).
	// Generally the hostname should always match, unless we are stuck
	// with a hostname that is not an xname but still need to contact the
	// endpoint.
	if rep.ID == "" {
		ep.ID = ep.Hostname
	} else {
		ep.ID = strings.Trim(rep.ID, "./ ")
	}
	// ID should be set by now unless we got no ID, hostname or FQDN.
	if ep.ID == "" {
		err := fmt.Errorf("No xname ID found")
		return nil, err
	}
	ep.ID = xnametypes.NormalizeHMSCompID(ep.ID)

	// Get type from ID (or hostname if ID not given).  It should be a
	// valid controller type.
	hmsType := xnametypes.GetHMSType(ep.ID)
	if xnametypes.IsHMSTypeController(hmsType) ||
		hmsType == xnametypes.MgmtSwitch ||
		hmsType == xnametypes.MgmtHLSwitch ||
		hmsType == xnametypes.CDUMgmtSwitch {
		ep.Type = hmsType.String()
	} else if hmsType == xnametypes.HMSTypeInvalid {
		// No type found.  Not a valid xname
		err := fmt.Errorf("%s is not a valid locational xname ID", ep.ID)
		return nil, err
	} else {
		// Found a type but it is not a controller.  Other component
		// types should not be Redfish endpoints.
		err := fmt.Errorf("xname ID %s has wrong type for RF endpoint: %s",
			ep.ID, hmsType.String())
		return nil, err
	}
	ep.Name = rep.Name

	// Hostname may not be set if just the ID field is given.  This is
	// used for the FQDN instead.
	hostname := ep.Hostname
	if hostname == "" {
		hostname = ep.ID
	}

	// Set FQDN.  This is not actually a stored field, as these must match
	// the hostname + domain at this point, either because both were given,
	// or because we got them from the rep.FQDN.
	if ep.Domain != "" {
		if hostIsIP {
			ep.FQDN = ep.Hostname
		} else {
			ep.FQDN = hostname + "." + ep.Domain
		}
	} else {
		ep.FQDN = hostname
	}
	// If these don't match, host/id + domain given but it doesn't agree
	// with non-empty FQDN
	repFQDN := GetIPAddressString(rep.FQDN)
	if repFQDN == "" {
		repFQDN = rep.FQDN
	}
	if rep.FQDN != "" && ep.FQDN != repFQDN {
		err := fmt.Errorf("host/domain conflicts with FQDN: '%s' != '%s'",
			ep.FQDN, repFQDN)
		return nil, err
	}
	// Validate given IP address that is not in the 'Hostname' or 'FQDN' fields
	if rep.IPAddr != "" {
		repIP := GetIPAddressString(rep.IPAddr)
		if repIP == "" {
			err := fmt.Errorf("IPAddress is not a valid IP address: '%s'", rep.IPAddr)
			return nil, err
		}
		// If the hostname is an IP address, the given IP address should match
		if hostIsIP && repIP != ep.Hostname {
			err := fmt.Errorf("hostname IP address conflicts with given IP address: '%s' != '%s'",
				ep.Hostname, repIP)
			return nil, err
		}
		ep.IPAddr = repIP
	} else if hostIsIP {
		// Use the IP address hostname as the IP address.
		ep.IPAddr = ep.Hostname
	}
	if rep.Enabled != nil {
		ep.Enabled = *rep.Enabled
	} else {
		ep.Enabled = EnabledDefault
	}
	ep.UUID = rep.UUID
	ep.User = rep.User
	ep.Password = rep.Password
	if rep.UseSSDP != nil {
		ep.UseSSDP = *rep.UseSSDP
	} else {
		ep.UseSSDP = UseSSDPDefault
	}
	if rep.MACRequired != nil {
		ep.MACRequired = *rep.MACRequired
	} else {
		ep.MACRequired = MACRequiredDefault
	}
	ep.MACAddr = rep.MACAddr
	if rep.RediscOnUpdate != nil {
		ep.RediscOnUpdate = *rep.RediscOnUpdate
	} else {
		ep.RediscOnUpdate = RediscOnUpdateDefault
	}
	ep.DiscInfo.LastStatus = NotYetQueried
	return ep, nil
}

/////////////////////////////////////////////////////////////////////////////
//
// RedfishEndpoint discovery
//
/////////////////////////////////////////////////////////////////////////////

type RedfishEPDescription struct {
	ID             string        `json:"ID"`
	Type           string        `json:"Type"`
	Name           string        `json:"Name,omitempty"` // user supplied descriptive name
	Hostname       string        `json:"Hostname"`
	Domain         string        `json:"Domain"`
	FQDN           string        `json:"FQDN"`
	Enabled        bool          `json:"Enabled"`
	UUID           string        `json:"UUID,omitempty"`
	User           string        `json:"User"`
	Password       string        `json:"Password"` // Temporary until more secure method
	UseSSDP        bool          `json:"UseSSDP,omitempty"`
	MACRequired    bool          `json:"MACRequired,omitempty"`
	MACAddr        string        `json:"MACAddr,omitempty"`
	IPAddr         string        `json:"IPAddress,omitempty"`
	RediscOnUpdate bool          `json:"RediscoverOnUpdate"`
	TemplateID     string        `json:"TemplateID,omitempty"`
	DiscInfo       DiscoveryInfo `json:"DiscoveryInfo"`
}

// String function to redact passwords from any kind of output
func (red RedfishEPDescription) String() string {
	// NOTE: the value form is slightly less efficient since it involves a
	//  copy, but it will work for both pass by value and pass by pointer.
	buf := bytes.NewBufferString("{")
	fmt.Fprintf(buf, "ID: %s, ", red.ID)
	fmt.Fprintf(buf, "Type: %s, ", red.Type)
	fmt.Fprintf(buf, "Name: %s, ", red.Name)
	fmt.Fprintf(buf, "Hostname: %s, ", red.Hostname)
	fmt.Fprintf(buf, "Domain: %s, ", red.Domain)
	fmt.Fprintf(buf, "FQDN: %s, ", red.FQDN)
	fmt.Fprintf(buf, "Enabled: %t, ", red.Enabled)
	fmt.Fprintf(buf, "UUID: %s, ", red.UUID)
	fmt.Fprintf(buf, "User: %s, ", red.User)
	fmt.Fprintf(buf, "Password: <REDACTED>, ")
	fmt.Fprintf(buf, "UseSSDP: %t, ", red.UseSSDP)
	fmt.Fprintf(buf, "MACRequired: %t, ", red.MACRequired)
	fmt.Fprintf(buf, "MACAddr: %s, ", red.MACAddr)
	fmt.Fprintf(buf, "IPAddress: %s, ", red.IPAddr)
	fmt.Fprintf(buf, "RediscOnUpdate: %t, ", red.RediscOnUpdate)
	fmt.Fprintf(buf, "TemplateID: %s, ", red.TemplateID)
	fmt.Fprintf(buf, "DiscInfo: %+v", red.DiscInfo)
	fmt.Fprintf(buf, "}")
	return buf.String()
}

type DiscoveryInfo struct {
	LastAttempt    string `json:"LastDiscoveryAttempt,omitempty"`
	LastStatus     string `json:"LastDiscoveryStatus"`
	RedfishVersion string `json:"RedfishVersion,omitempty"`
}

// Update Status and set timestamp to now.
func (d *DiscoveryInfo) UpdateLastStatusWithTS(status string) {
	d.LastAttempt = time.Now().UTC().Format("2006-01-02T15:04:05.000000Z07:00")
	d.LastStatus = status
}

// Set timestamp to now.
func (d *DiscoveryInfo) TSNow() {
	d.LastAttempt = time.Now().UTC().Format("2006-01-02T15:04:05.000000Z07:00")
}

type RedfishEPDescriptions struct {
	RfEPDescriptions []RedfishEPDescription
}

// Create RedfishEPDescriptions struct from raw input from decoded RawEndpoints
// struct, e.g. from a JSON file or POST body.
func NewRedfishEPDescriptions(reps *RawRedfishEPs) (*RedfishEPDescriptions, error) {
	// We create one endpoint per desc
	epds := new(RedfishEPDescriptions)
	if reps == nil {
		err := fmt.Errorf("got nil pointer")
		return epds, err
	}

	var savedError error
	for _, rep := range reps.RedfishEndpoints {
		epd, err := NewRedfishEPDescription(&rep)
		if err == nil {
			epds.RfEPDescriptions = append(epds.RfEPDescriptions, *epd)
		} else {
			errlog.Printf("NewRedfishEPDescriptions: %s", err)
			savedError = err
		}
	}
	return epds, savedError
}

// This is the endpoint structure generated from a base RedfishEPDescription.
// It is used to facilitate connections to the endpoints and is
// is used to discover top-level resources (e.g. systems) so that
// they can be discovered in more details with routines appropriate
// to each type.
type RedfishEP struct {
	// Embedded struct
	RedfishEPDescription

	ServiceRootURL string `json:"serviceRootURL"` // URL of root service
	RedfishType    string `json:"redfishType"`    // i.e. ServiceRoot
	IPaddr         string `json:"ipaddr"`
	OdataID        string `json:"odataID"` // i.e. /redfish/v1

	ServiceRootRF  ServiceRoot       `json:"serviceRootRF"`
	NumChassis     int               `json:"numChassis"`
	NumManagers    int               `json:"numManagers"`
	NumSystems     int               `json:"numSystems"`
	NumRackPDUs    int               `json:"numRackPDUs"`
	AccountService *EpAccountService `json:"accountService"`
	SessionService *EpSessionService `json:"sessionService"`
	EventService   *EpEventService   `json:"eventService"`
	TaskService    *EpTaskService    `json:"taskService"`
	UpdateService  *EpUpdateService  `json:"updateService"`
	Chassis        EpChassisSet      `json:"chassis"`
	Managers       EpManagers        `json:"managers"`
	Systems        EpSystems         `json:"systems"`
	RackPDUs       EpPDUs            `json:"rackpdus"`

	rootSvcRaw  *json.RawMessage //`json:"rootSvcRaw"`
	chassisRaw  *json.RawMessage //`json:"chassisRaw"`
	managersRaw *json.RawMessage //`json:"managersRaw"`
	systemsRaw  *json.RawMessage //`json:"systemsRaw"`

	// Contains various PowerEquipment links; we only care about PDUs for now
	powerEquipment *PowerEquipment

	client *hms_certs.HTTPClientPair
}

// Create RedfishEP struct from a validated RedfishEndpointDescription.
// The description would be generated from user-supplied RawEndpoints and/or
// retrieved from the database.
// The RedfishEP struct is set up based on the description to conduct queries
// of the remote endpoint and to store the raw data retrieved from it.
func NewRedfishEp(rep *RedfishEPDescription) (*RedfishEP, error) {
	ep := new(RedfishEP)

	if rep == nil {
		err := fmt.Errorf("got nil pointer!")
		ep.DiscInfo.UpdateLastStatusWithTS(EndpointInvalid)
		return ep, err
	}
	ep.RedfishEPDescription = *rep

	ep.ServiceRootURL = ep.FQDN + "/redfish/v1"
	ep.OdataID = "/redfish/v1"
	ep.NumSystems = 0
	// Add client handle.  Allow for proxy if configured.
	/*
		if httpClientProxyURL != "" {
			ep.client = RfProxyClient(httpClientProxyURL)
		} else {
			ep.client = RfDefaultClient()
		}
	*/
	ep.client = RfDefaultClient()
	err := ep.CheckPrePhase1()
	if err != nil {
		errlog.Printf("NewRedfishEp failed: %s", err)
		ep.DiscInfo.UpdateLastStatusWithTS(EndpointInvalid)
		return ep, err
	}
	return ep, nil
}

// Set of RedfishEP, struct representing a root-level RF endpoint in system.
type RedfishEPs struct {
	Num       int                   `json:"num"`
	IDs       map[string]*RedfishEP `json:"ids"`
	waitGroup sync.WaitGroup        // For synchronizing threads
}

// Create RedfishEPs structs from a set of validated RedfishEndpointDescriptions.
func NewRedfishEps(epds *RedfishEPDescriptions) (*RedfishEPs, error) {
	// We create one endpoint per raw entry under a RedfishEPs struct.
	eps := new(RedfishEPs)
	eps.IDs = make(map[string]*RedfishEP)
	eps.Num = 0
	if epds == nil {
		err := fmt.Errorf("got nil pointer")
		return eps, err
	}
	var savedError error
	for i, epd := range epds.RfEPDescriptions {
		ep, err := NewRedfishEp(&epd)
		if err != nil {
			errlog.Println("Endpoint ", i, " was invalid: ")
			savedError = err
		} else {
			eps.IDs[ep.ID] = ep
			eps.Num = eps.Num + 1
		}
	}
	return eps, savedError
}

// GET the page at the given rpath relative to the redfish hostname of
// the given endpoint, e.g. /redfish/v1/Systems/System.Embedded.1.  Keeping
// with Redfish style there should always be a leading slash and the
// "/redfish/v1" part should presumably always be present in the rpath (as the
// odata.id always includes this).
//
// There is an optional argument to provide the retry count.  If not given,
// the default is 3.  This is the number of times to retry the GET if it fails.
//
// If no error results, result should be the raw body (i.e. Redfish JSON).
// returned.
// This is the starting point for decoding the JSON into a particular
// structure (i.e. given the resource's schema, or into a generic
// interface{} map.
func (ep *RedfishEP) GETRelative(rpath string, optionalArgs ...int) (json.RawMessage, error) {
	var rsp *http.Response
	var path string = "https://" + ep.FQDN + strings.Replace(rpath, "#", "%23", -1)
	var body []byte

	// Process optional timeout argument
	retryCount := 3
	if len(optionalArgs) > 0 {
		retryCount = optionalArgs[0]
	}

	// In case we don't catch this...
	if ep.FQDN == "" {
		errlog.Printf("Can't HTTP GET (%s): FQDN is empty", path)
		return nil, ErrRFDiscFQDNMissing
	}
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		errlog.Printf("Error forming new request for (%s) %s", path, err)
		return nil, err
	}
	req.SetBasicAuth(ep.User, ep.Password)
	req.Header.Set("Accept", "*/*")
	req.Close = true

	//TODO: Future enhancement for unsupported River BMCs to reduce RF failovers
	//and log clutter:
	//
	// Check the ID (xname) and:
	// o If ep.client.SecureClient != nil && != InsecureClient:
	//   o If ID shows 'c0', then scrutinize:
	//     o If there have been > 1 failovers with successful Do() calls, then set
	//       ep.client.SecureClient = InsecureClient

	// Do retries on errors. They could be temporary interuptions in service.
	sleepTime := 1
	for retry := 0; retry <= retryCount; retry++ {
		rsp, err = ep.client.Do(req)
		if err != nil {
			base.DrainAndCloseResponseBody(rsp)
			if retry == retryCount {
				errlog.Printf("GETRelative (%s) ERROR: %s, Failing after %d retries", path, err, retry)
				return nil, err
			} else {
				errlog.Printf("GETRelative (%s) ERROR: %s, Retry %d after %d seconds...", path, err, retry + 1, sleepTime)
				time.Sleep(time.Duration(sleepTime) * time.Second)
				sleepTime += (retry + 1) * 10
				continue
			}
		}
		break
	}

	if rsp.Body != nil {
		body, _ = ioutil.ReadAll(rsp.Body)
	}
	base.DrainAndCloseResponseBody(rsp)

	if rsp.StatusCode != http.StatusOK {
		rerr := fmt.Errorf("%s", http.StatusText(rsp.StatusCode))
		errlog.Printf("GETRelative (%s) Bad rsp: %s", path, rerr)
		if rsp.StatusCode == http.StatusNotFound {
			// Return a named error so we can take special action
			return nil, ErrRFDiscURLNotFound
		} else {
			var compErr RedfishError
			if err := json.Unmarshal(json.RawMessage(body), &compErr); err != nil {
				if IsUnmarshalTypeError(err) {
					errlog.Printf("bad field(s) skipped: %s: %s\n", path, err)
				} else {
					errlog.Printf("ERROR: json decode failed: %s: %s\n", path, err)
					return nil, rerr
				}
			}
			if len(compErr.Error.ExtendedInfo) > 0 && strings.Contains(compErr.Error.ExtendedInfo[0].MessageId, "LicenseKeyRequired") {
				return nil, ErrRFDiscILOLicenseReq
			}
		}
		return nil, rerr
	}

	// We want to return the raw JSON output.  It unmarshals just as
	// well if it's indented, so we do that here to verify that it is
	// valid JSON. This lets us defer to the caller how to unmarshall it,
	// and puts it in a more (human) readable format.
	var out bytes.Buffer
	err = json.Indent(&out, body, "", "\t")
	if err != nil {
		errlog.Printf("Error decoding %s: %s", path, err)
		return nil, err
	}
	// Dump response and path in a unit-test friendly format for regression
	// test auto-generation.
	if genTestingPayloadsTitle != "" {
		if genTestingPayloadsDumpEpID == ep.ID {
			GenTestingPayloads(genTestingPayloadsOutfile,
				genTestingPayloadsTitle,
				rpath,
				out.Bytes())
		}
	}
	jsonBody := json.RawMessage(out.Bytes())
	return jsonBody, nil
}

// Loop through all endpoints to get top-level information, i.e.
// how many systems, etc.  and initalize these structures so they
// can be discovered in more detail.
//
// Note this is done in parallel, each RedfishEP in a separate thread.
// Each RedfishEP structure should not share any data with the others so
// little synchronization is needed.  We could conceivably speed things up
// more by parallelizing the individual page gets, but that would have to
// be done a bit more carefully and for now this should be fast enough.
func (eps *RedfishEPs) GetAllRootInfo() {
	var numWait int = 0
	//	for _, ep := range eps.Endpoints {
	for _, ep := range eps.IDs {
		// Wait for this go routine to finish
		eps.waitGroup.Add(1)
		// Start each endpoint as a separate thread
		go func(e *RedfishEP) {
			defer eps.waitGroup.Done()
			e.GetRootInfo()
		}(ep)
		numWait++
		if numWait > MaxFanout {
			errlog.Printf("GetAllRootInfo() Max fanout of %d reached.  "+
				"Wait for completion.", MaxFanout)
			eps.waitGroup.Wait()
			numWait = 0
		}
	}
	// Wait for all goroutines to finish
	eps.waitGroup.Wait()
}

// For a given Redfish endpoint, get top-level information, i.e.
// how many systems, chassis, managers, etc. and initalize these structures
// so can be discovered in more detail.
func (ep *RedfishEP) GetRootInfo() {
	ep.DiscInfo.TSNow()
	err := ep.CheckPrePhase1()
	if err != nil {
		errlog.Printf("Discover failed: %s", err)
		ep.DiscInfo.UpdateLastStatusWithTS(EndpointInvalid)
		return
	}
	// Skip inventory discovery if not enabled.
	if ep.Enabled == false {
		errlog.Printf("Discover skipped for %s: '%s'",
			ep.ID, EndpointNotEnabled)
		ep.DiscInfo.UpdateLastStatusWithTS(EndpointNotEnabled)
		return
	}
	// Get ServiceRoot for endpoint
	path := ep.OdataID
	rootSvcJSON, err := ep.GETRelative(path)
	if err != nil || rootSvcJSON == nil {
		ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetFailed)
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", ep.FQDN+path, rootSvcJSON)
	}
	ep.rootSvcRaw = &rootSvcJSON
	ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetOk)

	// Decode ServiceRoot JSON into matching Go struct
	err = json.Unmarshal(rootSvcJSON, &ep.ServiceRootRF)
	if err != nil {
		errlog.Printf("Failed to decode %s: %s\n", path, err)
		ep.DiscInfo.UpdateLastStatusWithTS(EPResponseFailedDecode)
	}
	ep.RedfishType = ServiceRootType
	ep.DiscInfo.RedfishVersion = ep.ServiceRootRF.RedfishVersion
	ep.UUID = ep.ServiceRootRF.UUID

	//
	// Now create structs for each of the services in the
	// SystemRoot, then discover them, so that we can interact
	// with the services they provide.
	//
	if ep.ServiceRootRF.AccountService.Oid != "" {
		oid := ep.ServiceRootRF.AccountService.Oid
		ep.AccountService = NewEpAccountService(ep, oid)
		ep.AccountService.discoverRemotePhase1()
	} else {
		errlog.Printf("%s: No AccountService entry found!\n", ep.FQDN)
	}
	if ep.ServiceRootRF.SessionService.Oid != "" {
		oid := ep.ServiceRootRF.SessionService.Oid
		ep.SessionService = NewEpSessionService(ep, oid)
		ep.SessionService.discoverRemotePhase1()
	} else {
		errlog.Printf("%s: No SessionService entry found!\n", ep.FQDN)
	}
	if ep.ServiceRootRF.EventService.Oid != "" {
		oid := ep.ServiceRootRF.EventService.Oid
		ep.EventService = NewEpEventService(ep, oid)
		ep.EventService.discoverRemotePhase1()
	} else {
		errlog.Printf("%s: No EventService entry found!\n", ep.FQDN)
	}
	// Note: The service root property is called "Tasks" but should point to
	// /redfish/v1/TaskService.  We use the latter for consistency
	// in the structs created here.
	if ep.ServiceRootRF.Tasks.Oid != "" {
		oid := ep.ServiceRootRF.Tasks.Oid
		ep.TaskService = NewEpTaskService(ep, oid)
		ep.TaskService.discoverRemotePhase1()
	} else {
		errlog.Printf("%s: No TaskService entry found!\n", ep.FQDN)
	}
	if ep.ServiceRootRF.UpdateService.Oid != "" {
		oid := ep.ServiceRootRF.UpdateService.Oid
		ep.UpdateService = NewEpUpdateService(ep, oid)
		ep.UpdateService.discoverRemotePhase1()
	} else {
		errlog.Printf("%s: No UpdateService entry found!\n", ep.FQDN)
	}
	//
	// We now take each set of root level Redfish component objects in
	// turn so we can dive deeper and collect info on those we need for
	// futher system discovery.
	//
	// First, the set of Redfish Chassis objects for the endpoint.
	// Start by fetching the Chassis/ set from the root.
	//
	if ep.ServiceRootRF.Chassis.Oid != "" {
		path = ep.ServiceRootRF.Chassis.Oid
	} else {
		path = ep.OdataID + "/Chassis"
	}
	chassisJSON, err := ep.GETRelative(path)
	if err != nil && !xnametypes.ControllerHasChassisStr(ep.Type) {
		// Don't expect any Chassis here, so if no collection, no problem.
		// Just create an empty collection so we don't choke later.
		ep.NumChassis = 0
		ep.Chassis.OIDs = make(map[string]*EpChassis)
	} else if err != nil || chassisJSON == nil {
		// Expected Chassis collection but didn't get one or it was corrupt.
		ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetFailed)
		return
	} else {
		// Found Chassis collection
		if rfDebug > 0 {
			errlog.Printf("%s: %s\n", ep.FQDN+path, chassisJSON)
		}
		ep.chassisRaw = &chassisJSON
		ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetOk)

		// Decode chassis list for endpoint, create an EpChassis for each
		// for subsequent discovery.
		var chInfo ChassisCollection
		err = json.Unmarshal(chassisJSON, &chInfo)
		if err != nil {
			errlog.Printf("Failed to decode %s: %s\n", path, err)
			ep.DiscInfo.UpdateLastStatusWithTS(EPResponseFailedDecode)
			return
		}

		// The count is typically given as "Members@odata.count", but
		// older versions drop the "Members" identifier
		ep.NumChassis = len(chInfo.Members)
		if chInfo.MembersOCount > 0 && chInfo.MembersOCount != ep.NumChassis {
			errlog.Printf("%s: Member@odata.count != Member array len\n", ep.FQDN+path)
		} else if chInfo.OCount > 0 && chInfo.OCount != ep.NumChassis {
			errlog.Printf("%s: odata.count != Member array len\n", ep.FQDN+path)
		}
		ep.Chassis.OIDs = make(map[string]*EpChassis)
		ep.Chassis.Num = ep.NumChassis
		sort.Sort(ResourceIDSlice(chInfo.Members))
		for i, chOID := range chInfo.Members {
			chID := chOID.Basename()
			ep.Chassis.OIDs[chID] = NewEpChassis(ep, chOID, i)
		}
		// Fetch info for each chassis in  list and populate new structs.
		ep.Chassis.discoverRemotePhase1()
	}

	//
	// Next,  the set of Managers for the endpoint.
	// Get Managers/ root listing of all Managers (BMCs, etc.) under endpoint.
	//
	if ep.ServiceRootRF.Managers.Oid != "" {
		path = ep.ServiceRootRF.Managers.Oid
	} else {
		path = ep.OdataID + "/Managers"
	}
	managersJSON, err := ep.GETRelative(path)
	if err != nil || managersJSON == nil {
		ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetFailed)
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", ep.FQDN+path, managersJSON)
	}
	ep.managersRaw = &managersJSON
	ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetOk)

	// Decode Managers list for endpoint, create an EpManager for each
	// for subsequent discovery.
	var manInfo ManagerCollection
	err = json.Unmarshal(managersJSON, &manInfo)
	if err != nil {
		errlog.Printf("Failed to decode %s: %s\n", path, err)
		ep.DiscInfo.UpdateLastStatusWithTS(EPResponseFailedDecode)
		return
	}
	ep.NumManagers = len(manInfo.Members)
	if manInfo.MembersOCount > 0 && manInfo.MembersOCount != ep.NumManagers {
		errlog.Printf("%s: Member@odata.count != Member array len\n", ep.FQDN+path)
	} else if manInfo.OCount > 0 && manInfo.OCount != ep.NumManagers {
		errlog.Printf("%s: odata.count != Member array len\n", ep.FQDN+path)
	}
	ep.Managers.OIDs = make(map[string]*EpManager)
	ep.Managers.Num = ep.NumManagers
	sort.Sort(ResourceIDSlice(manInfo.Members))
	for i, mOID := range manInfo.Members {
		mID := mOID.Basename()
		ep.Managers.OIDs[mID] = NewEpManager(ep, mOID, i)
	}
	ep.Managers.discoverRemotePhase1()

	//
	// Next, the set of ComputerSystems for the endpoint.
	// Get Systems/ root listing of all Systems under endpoint.
	//
	status := ep.GetSystems()
	if status != HTTPsGetOk {
		return
	}

	//
	// Next, the PowerEquipment for the endpoint, if it exits.  For now,
	// we just get the RackPDUs collection under it.
	//
	// HPE PDUs use PowerDistribution, so setup PowerEquipment path
	if ep.ServiceRootRF.PowerDistribution.Oid != "" {
		ep.ServiceRootRF.PowerEquipment.Oid = "/redfish/v1/PowerEquipment"
	}

	if ep.ServiceRootRF.PowerEquipment.Oid != "" {
		path = ep.ServiceRootRF.PowerEquipment.Oid
		powerJSON, err := ep.GETRelative(path)
		if err != nil || powerJSON == nil {
			ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetFailed)
			return
		}
		if rfDebug > 0 {
			errlog.Printf("%s: %s\n", ep.FQDN+path, powerJSON)
		}
		ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetOk)

		// Decode PowerEquipment object
		var powerInfo PowerEquipment
		err = json.Unmarshal(powerJSON, &powerInfo)
		if err != nil {
			errlog.Printf("Failed to decode %s: %s\n", path, err)
			ep.DiscInfo.UpdateLastStatusWithTS(EPResponseFailedDecode)
			return
		}
		ep.powerEquipment = &powerInfo

		// Get RackPDU collection, if it exists
		if powerInfo.RackPDUs.Oid != "" {
			path = powerInfo.RackPDUs.Oid
			pduJSON, err := ep.GETRelative(path)
			if err != nil || pduJSON == nil {
				ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetFailed)
				return
			}
			if rfDebug > 0 {
				errlog.Printf("%s: %s\n", ep.FQDN+path, pduJSON)
			}
			ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetOk)

			var pduInfo PowerDistributionCollection
			err = json.Unmarshal(pduJSON, &pduInfo)
			if err != nil {
				errlog.Printf("Failed to decode %s: %s\n", path, err)
				ep.DiscInfo.UpdateLastStatusWithTS(EPResponseFailedDecode)
				return
			}
			ep.NumRackPDUs = len(pduInfo.Members)
			ep.RackPDUs.Num = ep.NumRackPDUs

			ep.RackPDUs.OIDs = make(map[string]*EpPDU)
			sort.Sort(ResourceIDSlice(pduInfo.Members))
			for i, pduOID := range pduInfo.Members {
				pduID := pduOID.Basename()
				ep.RackPDUs.OIDs[pduID] = NewEpPDU(ep, pduOID, i)
			}
			ep.RackPDUs.discoverRemotePhase1()
		}
	}

	//
	// Phase 2 - remote queries are done for entire root.  Now use this
	// info to tie the Redfish properties to HMS ones, like HMS Type and
	// location, so they can be organized into a larger system that contains
	// the discovered hardware for all of the system's endpoints.
	//
	ep.DiscInfo.UpdateLastStatusWithTS(VerifyingData)

	var childStatus string = DiscoverOK
	if err := ep.Chassis.discoverLocalPhase2(); err != nil {
		errlog.Printf("ERROR: Chassis verification failed: %s", err)
		childStatus = ChildVerificationFailed
	}
	if err := ep.Managers.discoverLocalPhase2(); err != nil {
		errlog.Printf("ERROR: Managers verification failed: %s", err)
		childStatus = ChildVerificationFailed
	}
	if err := ep.RackPDUs.discoverLocalPhase2(); err != nil {
		errlog.Printf("ERROR: RackPDUs verification failed: %s", err)
		childStatus = ChildVerificationFailed
	}
	// Note we need to do systems last because they are the most likely
	// to need info from the other objects.
	if err := ep.VerifySystems(); err != nil {
		errlog.Printf("ERROR: Systems verification failed: %s", err)
		childStatus = ChildVerificationFailed
	}
	ep.DiscInfo.UpdateLastStatusWithTS(childStatus)
}

func (ep *RedfishEP) GetSystems() string {
	var path string

	// This is the CMC special name. Skip discovering this node
	// that shouldn't exist.
	if xnametypes.GetHMSType(ep.ID) == xnametypes.NodeBMC &&
		strings.HasSuffix(ep.ID, "b999") {
		return HTTPsGetOk
	}

	if ep.ServiceRootRF.Systems.Oid != "" {
		path = ep.ServiceRootRF.Systems.Oid
	} else {
		path = ep.OdataID + "/Systems"
	}
	systemsJSON, err := ep.GETRelative(path)
	if err != nil && !xnametypes.ControllerHasSystemsStr(ep.Type) {
		// Don't expect systems, so if the collection is missing, just
		// mark there as being zero move on.
		ep.NumSystems = 0
		ep.Systems.OIDs = make(map[string]*EpSystem)
	} else if err != nil || systemsJSON == nil {
		// Really expected a collection here, but it was missing or corrupt.
		ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetFailed)
		return HTTPsGetFailed
	} else {
		if rfDebug > 0 {
			errlog.Printf("%s: %s\n", ep.FQDN+path, systemsJSON)
		}
		ep.systemsRaw = &systemsJSON
		ep.DiscInfo.UpdateLastStatusWithTS(HTTPsGetOk)

		// Decode Systems list for endpoint, create an EpSystem for each
		// for subsequent discovery.
		var sysInfo SystemCollection
		err = json.Unmarshal(systemsJSON, &sysInfo)
		if err != nil {
			errlog.Printf("Failed to decode %s: %s\n", path, err)
			ep.DiscInfo.UpdateLastStatusWithTS(EPResponseFailedDecode)
			return EPResponseFailedDecode
		}
		ep.NumSystems = len(sysInfo.Members)
		if sysInfo.MembersOCount > 0 && sysInfo.MembersOCount != ep.NumSystems {
			errlog.Printf("%s: Member@odata.count != Member array len\n", ep.FQDN+path)
		} else if sysInfo.OCount > 0 && sysInfo.OCount != ep.NumSystems {
			errlog.Printf("%s: odata.count != Member array len\n", ep.FQDN+path)
		}
		ep.Systems.OIDs = make(map[string]*EpSystem)
		ep.Systems.Num = ep.NumSystems
		sort.Sort(ResourceIDSlice(sysInfo.Members))
		for i, sysOID := range sysInfo.Members {
			sID := sysOID.Basename()
			ep.Systems.OIDs[sID] = NewEpSystem(ep, sysOID, i)
		}
		ep.Systems.discoverRemotePhase1()
	}
	return HTTPsGetOk
}

func (ep *RedfishEP) VerifySystems() error {
	return ep.Systems.discoverLocalPhase2()
}

// Checks any fields that must be set in order to properly query the endpoint
// and organize and name subcomponents afterwards.
func (ep *RedfishEP) CheckPrePhase1() error {
	// These are all things that should always be set properly when the
	// RedfishEndpoint is first created.  But if they are unset odd
	// or incorrect things will happen so we don't want to go further.
	if ep.FQDN == "" {
		err := fmt.Errorf("no FQDN set for ID='%s', Hostname='%s', Domain='%s'",
			ep.ID, ep.Hostname, ep.Domain)
		return err
	}
	hmsType := xnametypes.GetHMSType(ep.ID)
	if (!xnametypes.IsHMSTypeController(hmsType) &&
		hmsType != xnametypes.MgmtSwitch &&
		hmsType != xnametypes.MgmtHLSwitch &&
		hmsType != xnametypes.CDUMgmtSwitch) ||
		ep.Type != hmsType.String() {
		err := fmt.Errorf("bad xname ID ('%s') or Type ('%s') for %s\n",
			ep.ID, ep.Type, ep.FQDN)
		return err
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////
//
//              Helper functions for discovery
//
// These are all linked to the RedfishEndpoint as it knows about all of
// the discovered hardware once the initial remote queries are done.  In
// addition, if there are settings that cannot be auto-discovered or should
// be set to non-default values, the RedfishEndpoint must manage this
// configuration info, as this makes the most sense scope-wise, and also
// users must already be able to edit info associated with the Redfish
// Endpoint for those endpoints that do not use SSDP.
//
// TODO: These are placeholders with very simples rule tuned to the
//       hardware we have right now.  They will need to choose values
//       based on the initial Redfish discovery info in the future to
//       determine the right values.
//
///////////////////////////////////////////////////////////////////////////

//
// RedfishEndpoint field discovery - Post phase 1 discovery.
//

// Note these are all sort of inter-related, but don't call them from each
// other.  If the code needs a value set elsewhere, make it a parameter.
// Also, don't modify the structure directly, and don't assume any non-Redfish
// fieldbeyond phase 1 (except those explicitly given as parameters) has been
// set by a previous call.

//////////////////////////////////////////////////////////////////////////
// Chassis parameter discovery
//////////////////////////////////////////////////////////////////////////

// Determines based on discovered info the xname of the Chassis.
// Need to know real (not raw ordinal) and HMS type first.
// Post phase 1 discovery.
func (ep *RedfishEP) getChassisHMSID(c *EpChassis, hmsType string, ordinal int) string {
	hmsTypeStr := xnametypes.VerifyNormalizeType(hmsType)
	if hmsTypeStr == "" {
		// This is an error or a skipped type.
		return ""
	}
	if ordinal < 0 || ordinal >= len(ep.Chassis.OIDs) {
		// Invalid ordinal or initial -1 value.
		return ""
	}
	if hmsTypeStr == xnametypes.MgmtSwitch.String() ||
		hmsTypeStr == xnametypes.MgmtHLSwitch.String() ||
		hmsTypeStr == xnametypes.CDUMgmtSwitch.String() {
		return ep.ID
	}
	// If the RedfishEndpoint ID is valid, there will be a b in the xname.
	epIDSplit := strings.SplitN(ep.ID, "b", 2)
	if len(epIDSplit) == 1 {
		// Bad RFEndpoint xname - not a BMC - may match other component
		return ""
	}
	if hmsTypeStr == xnametypes.HSNBoard.String() {
		// If RouterBMC is x0c0r0b[0-9], enclosure is  x0c0r0e[0-9]
		return epIDSplit[0] + "e" + epIDSplit[1]
	}
	if hmsTypeStr == xnametypes.NodeEnclosure.String() {
		// If NodeBMC is x0c0s0b[0-9]+, enclosure is  x0c0s0e[0-9]+
		return epIDSplit[0] + "e" + epIDSplit[1]
	}
	if hmsTypeStr == xnametypes.Chassis.String() {
		// Top-level Chassis for x0c0b0 is x0c0
		return epIDSplit[0]
	}
	if hmsTypeStr == xnametypes.ComputeModule.String() {
		if ep.Type == xnametypes.ChassisBMC.String() {
			// Chassis + s + ordinal
			return epIDSplit[0] + "s" + strconv.FormatInt(int64(ordinal), 10)
		}
		// Shouldn't happen.  This should be the only possibility for now.
		return ""
	}
	if hmsTypeStr == xnametypes.RouterModule.String() {
		if ep.Type == xnametypes.ChassisBMC.String() {
			// Chassis + r + ordinal
			return epIDSplit[0] + "r" + strconv.FormatInt(int64(ordinal), 10)
		}
		if ep.Type == xnametypes.RouterBMC.String() {
			// Already have Chassis +r in the path, remove bmc portion
			return epIDSplit[0]
		}
		// Shouldn't happen.  These should be the only possibilities
		return ""
	}
	return ""
}

// Gets the HMS type of the Chassis - Note Invalid means something
// special here, namely, "skip it"/"not supported".  Generally we do our
// best with Chassis components since there may be many that don't represent
// what we actually track.
// Post phase 1 discovery.
func (ep *RedfishEP) getChassisHMSType(c *EpChassis) string {
	switch c.RedfishSubtype {
	case RFSubtypeEnclosure:
		if ep.Type == xnametypes.ChassisBMC.String() &&
			IsManufacturer(c.ChassisRF.Manufacturer, CrayMfr) != 0 {
			// ChassisBMC and is not non-Cray, must be the Chassis itself
			// by convention.
			return xnametypes.Chassis.String()
		}
		if ep.Type == xnametypes.RouterBMC.String() {
			// RouterBMC, must be the Chassis itself (Router card or TOR
			// enclosure) by convention.  We only have slingshot at the
			// moment so no need to guess.
			return xnametypes.HSNBoard.String()
		}
		// NodeEnclosures may be RackMount, Enclosure.
		fallthrough
	case RFSubtypeRackMount:
		if isFoxconnChassis(c) {
			// Foxconn Paradise has a bunch of RackMount chassis we can ignore
			return xnametypes.HMSTypeInvalid.String()
		}
		if ep.NumSystems > 0 {
			// Does the endpoint contain nodes?
			// For now assume NodeEnclosure.
			return xnametypes.NodeEnclosure.String()
		} else {
			return xnametypes.HMSTypeInvalid.String()
		}
	case RFSubtypeStandAlone:
		if IsManufacturer(c.ChassisRF.Manufacturer, GigabyteMfr) != 0 &&
			ep.NumSystems > 0 {
			// Is gigabyte ChassisBMC and has nodes, it is the node enclosure.
			return xnametypes.NodeEnclosure.String()
		} else {
			return xnametypes.HMSTypeInvalid.String()
		}
	case RFSubtypeBlade:
		if ep.Type == xnametypes.ChassisBMC.String() {
			// If is not non-Cray and Chassis BMC, but be compute or router
			// blade.
			// TODO: When there is something more reliable than the name of
			// the blade to use, use that instead.
			if IsManufacturer(c.ChassisRF.Manufacturer, CrayMfr) != 0 {
				if strings.HasPrefix(strings.ToLower(c.ChassisRF.Id), "blade") {
					return xnametypes.ComputeModule.String()
				}
				if strings.HasPrefix(strings.ToLower(c.ChassisRF.Id), "perif") {
					return xnametypes.RouterModule.String()
				}
			}
		}
		return xnametypes.HMSTypeInvalid.String()
	case RFSubtypeDrawer:
		if ep.Type == xnametypes.MgmtSwitch.String() ||
			ep.Type == xnametypes.MgmtHLSwitch.String() ||
			ep.Type == xnametypes.CDUMgmtSwitch.String() {
			return ep.Type
		}
		return xnametypes.HMSTypeInvalid.String()
	case RFSubtypeZone:
		if isFoxconnChassis(c) {
			// Foxconn Paradise uses the Baseboard_0 chassis as the primary node enclosure
			return xnametypes.NodeEnclosure.String()
		}
		return xnametypes.HMSTypeInvalid.String()
	default:
		// Other types are usually subcomponents we don't track and are
		// often not represented very consistently by different manufacturers.
		errlog.Printf("getChassisHMSType default case: c.RedfishSubtype: %s", c.RedfishSubtype)
		return xnametypes.HMSTypeInvalid.String()
	}
}

// Get type-specific ordinal of chassis, starting from zero.
// We take the Chassis set and find the longest non-numerical prefix
// in the current chassis.  We then count the number of items with such
// prefixes that have RawOrdinal values lower than c's, and whose types match.
//
// Note that the RawOrdinal values are assigned in lexographic order so we can
// retain that ordering even in an unsorted map.
//
// Also note that we do the prefix stuff because we can have the same Redfish
// type map to more than one HMS type (compute vs. router blades).
func (ep *RedfishEP) getChassisOrdinal(c *EpChassis) int {
	ordinal := 0
	prefix := ""

	// Get the prefix of the Redfish Id of C up to the first number.
	f := func(c rune) bool { return !unicode.IsLetter(c) && !unicode.IsPunct(c) }
	split := strings.FieldsFunc(c.BaseOdataID, f)
	if len(split) > 0 {
		prefix = split[0]
	}
	prefixLower := strings.ToLower(prefix)
	// Count the lower ordered chassis that are the same type as c and
	// have the same prefix if one exists.
	for _, ch := range ep.Chassis.OIDs {
		if c.RawOrdinal > ch.RawOrdinal {
			oidLower := strings.ToLower(ch.BaseOdataID)
			if prefixLower != "" && strings.HasPrefix(oidLower, prefixLower) {
				if c.ChassisRF.ChassisType == ch.ChassisRF.ChassisType {
					ordinal += 1
				}
			}
		}
	}
	// Something went wrong.
	if ordinal > c.RawOrdinal {
		errlog.Printf("BUG: Bad ordinal.")
		return -1
	}
	return ordinal
}

//////////////////////////////////////////////////////////////////////////
// System i.e. node field discovery - Post Phase 1
//////////////////////////////////////////////////////////////////////////

// Given ordinal and type of a system, return the xname of a
// ComputerSystem.
func (ep *RedfishEP) getSystemHMSID(s *EpSystem, hmsType string, ordinal int) string {
	hmsTypeStr := xnametypes.VerifyNormalizeType(hmsType)
	if hmsTypeStr == "" {
		// This is an error or a skipped type.
		return ""
	}
	if ordinal < 0 {
		// Ordinal was never set.
		errlog.Printf("BUG: Bad ordinal.")
		return ""
	}
	if hmsTypeStr == xnametypes.Node.String() {
		// Only one type to support at the moment, node.
		xnameID := ep.ID + "n" + strconv.Itoa(ordinal)
		return xnameID
	}
	return ""
}

// Determines based on discovered info and original list order what the
// node ordinal is, i.e. the n[0-n] in the xname, along with the HMS Type.
// Note: Only use physical systems for now (or systems with no type, provided
// there are no physical systems)
// Return -1 if invalid (bad input or unsupported RF SystemType).
func (ep *RedfishEP) getSystemOrdinalAndType(s *EpSystem) (int, string) {
	// Always use the order in the System collection for now.
	ordinal := 0
	hmsType := ""

	// Skip logical system types.
	if s.SystemRF.SystemType != RFSubtypePhysical &&
		s.SystemRF.SystemType != "" {

		return -1, hmsType
	} else {
		hmsType = xnametypes.Node.String()
	}
	// Count the lower ordered chassis that are the same type as m
	for _, sys := range ep.Systems.OIDs {
		if sys.SystemRF.SystemType == RFSubtypePhysical ||
			sys.SystemRF.SystemType == "" {

			if sys.SystemRF.SystemType != s.SystemRF.SystemType {
				if sys.SystemRF.SystemType == RFSubtypePhysical {
					// No SystemType, but at least one other system has
					// SystemType physical, so don't count this one.
					return -1, ""
				}
			}
			if s.RawOrdinal > sys.RawOrdinal {
				ordinal += 1
			}
		}
	}
	// Something went wrong or bad input.
	if ordinal > s.RawOrdinal {
		errlog.Printf("BUG: Bad ordinal.")
		return -1, ""
	}
	return ordinal, hmsType
}

// Fills in Domain part of Node component.  Node xname/ID + Domain = FQDN.
func (ep *RedfishEP) getNodeSvcNetDomain(s *EpSystem) string {
	// Default is to just use same domain as parent.
	// TODO: Find out other options are.  System-level default?  Per-RedfishEP?
	return ep.Domain
}

// For nodes with multiple ethernet interfaces, return the Redfish ID
// of the one that will be plugged into the management network.
func (ep *RedfishEP) getNodeSvcNetEthIfaceId(s *EpSystem) string {
	// TODO: Next discovery phase.  Hardcoded for now, only valid for
	//       the Dell boxes we currently have, but that is the only
	//       multi-interface scenario we don't handle another way.
	return "NIC.Integrated.1-3-1"
}

//
// System subcomponents
//

// Determines based on discovered info and original list order what the
// processor ordinal is, i.e. the n0p[0-n] in the xname.
func (ep *RedfishEP) getProcessorOrdinal(p *EpProcessor) int {
	// Always use the order in the System's ProcessorCollection for now.
	//look at the EpProcessor's ProcessorRF.ProcessorType field
	//to determine processor type
	ordinal := 0
	if p.ProcessorRF.ProcessorType == "GPU" || p.ProcessorRF.ProcessorType == "Accelerator" {
		ordinal = p.sysRF.accelCount
		p.sysRF.accelCount = p.sysRF.accelCount + 1
	} else {
		ordinal = p.sysRF.cpuCount
		p.sysRF.cpuCount = p.sysRF.cpuCount + 1
	}
	return ordinal
}

// Determines based on discovered info and original list order what the
// Memory module ordinal is, i.e. the n0d[0-n] in the xname.
func (ep *RedfishEP) getMemoryOrdinal(m *EpMemory) int {
	// Always use the order in the System's MemoryCollection for now.
	return m.RawOrdinal
}

// Determined based on discovered info and original list order what the
// drive ordinal is, i.e. the n0g[0-n]k[0-n] in the xname.
func (ep *RedfishEP) getDriveOrdinal(d *EpDrive) int {
	ordinal, err := strconv.Atoi(path.Base(d.OdataID))
	if err != nil {
		errlog.Printf("unable to convert Drive OdataID to ordinal: %s ", d.OdataID)
		errlog.Printf("using raw ordinal %s ", strconv.Itoa(d.RawOrdinal))
		ordinal = d.RawOrdinal
	}
	return ordinal
}

// Determined based on discovered info and original list order what the
// drive ordinal is, i.e. the n0g[0-n]k[0-n] in the xname.
func (ep *RedfishEP) getStorageCollectionOrdinal(sc *EpStorageCollection) int {
	ordinal, err := strconv.Atoi(path.Base(sc.OdataID))
	if err != nil {
		errlog.Printf("unable to convert StorageCollection OdataID to ordinal, using raw ordinal ")
		ordinal = sc.RawOrdinal
	}
	return ordinal
}

// Determines based on discovered info the xname of the Chassis.
// Need to know real (not raw ordinal) and HMS type first.
// Post phase 1 discovery.
func (ep *RedfishEP) getPowerSupplyHMSID(p *EpPowerSupply, hmsType string, ordinal int) string {
	hmsTypeStr := xnametypes.VerifyNormalizeType(hmsType)
	if hmsTypeStr == "" {
		// This is an error or a skipped type.
		return ""
	}
	//get the parent Power.PowerSupplies array
	if ordinal < 0 || ordinal >= len(p.powerRF.PowerRF.PowerSupplies) {
		// Invalid ordinal or initial -1 value.
		return ""
	}
	return p.chassisRF.ID + "t" + strconv.Itoa(p.Ordinal)
}

// Gets the HMS type of the PowerSupply - Note Invalid means something
// special here, namely, "skip it"/"not supported".
// Post phase 1 discovery.
func (ep *RedfishEP) getPowerSupplyHMSType(p *EpPowerSupply) string {
	parentChassisType := ep.getChassisHMSType(p.chassisRF)
	if parentChassisType == xnametypes.NodeEnclosure.String() {
		return xnametypes.NodeEnclosurePowerSupply.String()
	}
	if parentChassisType == xnametypes.Chassis.String() {
		return xnametypes.CMMRectifier.String()
	}
	return ""
}

// Determined based on discovered info and original list order what the
// PowerSupply ordinal is, i.e. the x0c[0-n]t[0-n] in the xname.
func (ep *RedfishEP) getPowerSupplyOrdinal(p *EpPowerSupply) int {
	//The position of any power supply in relation to its siblings is indicated
	//by the basename of its OdataID, so it is possible to retrieve and sort the keys of the
	//chassis PowerSupplies OIDS map to determine the proper Ordinal of any particular PowerSupply
	var ordinal = p.RawOrdinal
	if len(p.chassisRF.PowerSupplies.OIDs) > 0 {
		psOIDs := make([]string, 0, len(p.chassisRF.PowerSupplies.OIDs))
		for oid := range p.chassisRF.PowerSupplies.OIDs {
			psOIDs = append(psOIDs, oid)
		}
		//sort the OIDs in PowerSupplies.OIDs map
		sort.Strings(psOIDs)
		//the proper ordinal for this power supply is now the position of its OdataID in the psOIDs slice
		for i, psOID := range psOIDs {
			if psOID == p.OdataID {
				ordinal = i
				break
			}
		}
	}
	return ordinal
}

// Determined based on discovered info the xname of the Chassis.
// Need to know real (not raw ordinal) and HMS type first.
// Post phase 1 discovery.
func (ep *RedfishEP) getNodeAccelRiserHMSID(r *EpNodeAccelRiser, hmsType string, ordinal int) string {
	hmsTypeStr := xnametypes.VerifyNormalizeType(hmsType)
	if hmsTypeStr == "" {
		// This is an error or a skipped type.
		return ""
	}
	//get the parent Assembly.Assemblies array
	if ordinal < 0 || ordinal >= len(r.assemblyRF.AssemblyRF.Assemblies) {
		// Invalid ordinal or initial -1 value.
		return ""
	}
	return r.systemRF.ID + "r" + strconv.Itoa(r.Ordinal)
}

// Gets the HMS type of the NodeAccelRiser - Note Invalid means something
// special here, namely, "skip it"/"not supported".
// Post phase 1 discovery.
func (ep *RedfishEP) getNodeAccelRiserHMSType(r *EpNodeAccelRiser) string {
	return xnametypes.NodeAccelRiser.String()
}

// Determined based on discovered info and original list order that the
// NodeAccelRiser ordinal is.
func (ep *RedfishEP) getNodeAccelRiserOrdinal(r *EpNodeAccelRiser) int {
	//The position of any node accel riser in relation to its siblings is indicated
	//by the basename of its OdataID, so it is possible to retrieve and sort the keys of the
	//chassis NodeAccelRisers OIDS map to determine the proper Ordinal of any particular NodeAccelRiser
	var ordinal = r.RawOrdinal
	if len(r.systemRF.NodeAccelRisers.OIDs) > 0 {
		rsOIDs := make([]string, 0, len(r.systemRF.NodeAccelRisers.OIDs))
		for oid := range r.systemRF.NodeAccelRisers.OIDs {
			rsOIDs = append(rsOIDs, oid)
		}
		//sort the OIDs in NodeAccelRisers.OIDs map
		sort.Strings(rsOIDs)
		//the proper ordinal for this Node Accel Riser is now the position of its OdataID in the rsOIDs slice
		for i, rsOID := range rsOIDs {
			if rsOID == r.OdataID {
				ordinal = i
				break
			}
		}
	}
	return ordinal
}

// Determined based on discovered info the xname of the Chassis.
// Need to know real (not raw ordinal) and HMS type first.
// Post phase 1 discovery.
func (ep *RedfishEP) getNetworkAdapterHMSID(na *EpNetworkAdapter, hmsType string, ordinal int) string {
	hmsTypeStr := xnametypes.VerifyNormalizeType(hmsType)
	if hmsTypeStr == "" {
		// This is an error or a skipped type.
		return ""
	}
	//get the parent Assembly.Assemblies array
	if ordinal < 0 || ordinal >= len(na.systemRF.NetworkAdapters.OIDs) {
		// Invalid ordinal or initial -1 value.
		return ""
	}
	//unsupported parent chassis type
	return na.systemRF.ID + "h" + strconv.Itoa(na.Ordinal)
}

// Gets the HMS type of the NetworkAdapter - Note Invalid means something
// special here, namely, "skip it"/"not supported".
// Post phase 1 discovery.
func (ep *RedfishEP) getNetworkAdapterHMSType(na *EpNetworkAdapter) string {
	return xnametypes.NodeHsnNic.String()
}

// Determined based on discovered info and original list order that the
// NetworkAdapter ordinal is.
func (ep *RedfishEP) getNetworkAdapterOrdinal(na *EpNetworkAdapter) int {
	//The position of any NetworkAdapter in relation to its siblings is indicated
	//by the basename of its OdataID. This ordering is already reflected in the RawOrdinal

	return na.RawOrdinal
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetNetworkAdapterFRUID(na *EpNetworkAdapter) (fruid string, err error) {
	return getStandardFRUID(na.Type, na.ID, na.NetworkAdapterRF.Manufacturer, na.NetworkAdapterRF.PartNumber, na.NetworkAdapterRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetNodeAccelRiserFRUID(r *EpNodeAccelRiser) (fruid string, err error) {
	return getStandardFRUID(r.Type, r.ID, r.NodeAccelRiserRF.Producer, r.NodeAccelRiserRF.PartNumber, r.NodeAccelRiserRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetPowerSupplyFRUID(p *EpPowerSupply) (fruid string, err error) {
	//PowerSupplies do not currently include PartNumbers
	return getStandardFRUID(p.Type, p.ID, p.PowerSupplyRF.Manufacturer, "", p.PowerSupplyRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetDriveFRUID(d *EpDrive) (fruid string, err error) {
	return getStandardFRUID(d.Type, d.ID, d.DriveRF.Manufacturer, d.DriveRF.PartNumber, d.DriveRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetMemoryFRUID(m *EpMemory) (fruid string, err error) {
	return getStandardFRUID(m.Type, m.ID, m.MemoryRF.Manufacturer, m.MemoryRF.PartNumber, m.MemoryRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetChassisFRUID(c *EpChassis) (fruid string, err error) {
	return getStandardFRUID(c.Type, c.ID, c.ChassisRF.Manufacturer, c.ChassisRF.PartNumber, c.ChassisRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetSystemFRUID(s *EpSystem) (fruid string, err error) {
	return getStandardFRUID(s.Type, s.ID, s.SystemRF.Manufacturer, s.SystemRF.PartNumber, s.SystemRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetManagerFRUID(m *EpManager) (fruid string, err error) {
	return getStandardFRUID(m.Type, m.ID, m.ManagerRF.Manufacturer, m.ManagerRF.PartNumber, m.ManagerRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetPDUFRUID(p *EpPDU) (fruid string, err error) {
	return getStandardFRUID(p.Type, p.ID, p.PowerDistributionRF.Manufacturer, p.PowerDistributionRF.PartNumber, p.PowerDistributionRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetProcessorFRUID(p *EpProcessor) (fruid string, err error) {
	return getStandardFRUID(p.Type, p.ID, p.ProcessorRF.Manufacturer, p.ProcessorRF.PartNumber, p.ProcessorRF.SerialNumber)
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func GetHSNNICFRUID(hmstype, id, manufacturer, partNum, serialNum string) string {
	fruID, err := getStandardFRUID(hmstype, id, manufacturer, partNum, serialNum)
	if err != nil {
		errlog.Printf("FRUID Error: %s\n", err.Error())
		errlog.Printf("Using untrackable FRUID: %s\n", fruID)
	}
	return fruID
}

// Build FRUID using standard fields: <Type>.<Manufacturer>.<PartNumber>.<SerialNumber>
// else return an error.
func getStandardFRUID(hmstype, id, manufacturer, partnumber, serialnumber string) (fruid string, err error) {
	isFRUTrackable := true
	var fruidBuilder strings.Builder
	var errorBuilder strings.Builder
	reg, _ := regexp.Compile("[^a-zA-Z0-9]*")
	errorBuilder.WriteString("Trackable " + hmstype + " FRUID can't be created. ")

	// Clean all fields
	manufacturerClean := reg.ReplaceAllString(manufacturer, "")
	partnumberClean := reg.ReplaceAllString(partnumber, "")
	serialnumberClean := reg.ReplaceAllString(serialnumber, "")

	//Need either Manufacturer or PartNumber and SerialNumber in order to build a unique FRUID
	if manufacturerClean != "" || partnumberClean != "" {
		periodDelimiter := ""
		if hmstype != "" {
			fruidBuilder.WriteString(hmstype)
			periodDelimiter = "."
		}
		if manufacturerClean != "" {
			fruidBuilder.WriteString(periodDelimiter + manufacturerClean)
			periodDelimiter = "."
		}
		if partnumberClean != "" {
			fruidBuilder.WriteString(periodDelimiter + partnumberClean)
			periodDelimiter = "."
		}
		if serialnumberClean != "" {
			fruidBuilder.WriteString(periodDelimiter + serialnumberClean)
		} else {
			isFRUTrackable = false
			errorBuilder.WriteString("Missing required fields: SerialNumber")
		}
	} else {
		isFRUTrackable = false
		errorBuilder.WriteString("Missing required fields: Manufacturer and/or PartNumber")
	}

	// The length limit here is to prevent FRUID generation from
	// causing database errors by storing FRUIDs that are too long.
	if isFRUTrackable && fruidBuilder.Len() > 255 {
		isFRUTrackable = false
		errorBuilder.WriteString("FRUID is too long, '" + fruidBuilder.String() + "'")
	}

	if isFRUTrackable {
		fruid = fruidBuilder.String()
	} else {
		fruid = "FRUIDfor" + id
		err = errors.New(errorBuilder.String())
	}

	return fruid, err
}

//////////////////////////////////////////////////////////////////////////
// Manager field discovery
//////////////////////////////////////////////////////////////////////////

// Determines based on discovered info the xname of the manager.
// If there is only one and has the same type as the manager, it must be
// the same as the parent RedfishEndpoint's xname ID.
func (ep *RedfishEP) getManagerHMSID(m *EpManager, hmsType string, ordinal int) string {
	// Note every hmsType and ordinal pair must get a unique xname ID
	hmsTypeStr := xnametypes.VerifyNormalizeType(hmsType)
	if hmsTypeStr == "" {
		// This is an error or a skipped type.
		return ""
	}
	if ordinal < 0 {
		// Ordinal was never set.
		return ""
	}
	if hmsTypeStr == ep.Type && ordinal == 0 {
		return ep.ID
	}
	// No idea what this manages.
	return ""
}

// Get the HMS type of the manager. int64( In most cases there is only one,
// the endpoint we're talking to, and so this is easy.
// In other cases, we should always produce exactly one
// xname (once/if the ordinal is added) per retur	"unicode"ned type, given a
// particular Redfish endpoint xname and type.
func (ep *RedfishEP) getManagerHMSType(m *EpManager) string {
	// Don't discover Management switch BMCs.
	if ep.Type == xnametypes.MgmtSwitch.String() ||
		ep.Type == xnametypes.MgmtHLSwitch.String() ||
		ep.Type == xnametypes.CDUMgmtSwitch.String() {
		return xnametypes.HMSTypeInvalid.String()
	}
	// Just one?  That's this endpoint's type.
	// example: RouterBMC
	if ep.Managers.Num == 1 {
		return ep.Type
	}
	// See if manager provides an entry point to the root service
	if ep.UUID != "" && m.ManagerRF.ServiceEntryPointUUID == ep.UUID {
		return ep.Type
	}
	if len(m.ManagedSystems) > 0 {
		// Does it manage any systems (i.e. nodes)?   NodeBMC.
		return xnametypes.NodeBMC.String()
	}
	if m.ManagerRF.ManagerType == RFSubtypeEnclosureManager {
		// Cassini NodeBMCs look like ChassisBMCs because they're missing the
		// Links.ManagerForServers field. If the managerType is "EnclosureManager",
		// check to see if there are any Systems at all.
		if ep.NumSystems > 0 {
			return xnametypes.NodeBMC.String()
		}
		return xnametypes.ChassisBMC.String()
	}
	// TODO: Multiple managers.  We don't roll up managers
	// so there should only be one per redfish endpoint.  If
	// that ever changes we will need to do more work.
	return xnametypes.HMSTypeInvalid.String()
}

// Determines based on discovered info and original list order what the
// Manager ordinal is, i.e. the b[0-n] in the xname.
func (ep *RedfishEP) getManagerOrdinal(m *EpManager) int {
	ordinal := 0

	// Count the lower ordered chassis that are the same type as m
	for _, mgr := range ep.Managers.OIDs {
		if m.RawOrdinal > mgr.RawOrdinal {
			if m.ManagerRF.ManagerType == mgr.ManagerRF.ManagerType {
				ordinal += 1
			}
		}
	}
	// Something went wrong or bad RawOrdinal value.
	if ordinal > m.RawOrdinal {
		return -1
	}
	return ordinal
}

////////////////////////////////////////////////////////////////////////////
// Component type/manufacturer detection
////////////////////////////////////////////////////////////////////////////

// Parsing manufacturer string
const (
	CrayMfr     = "Cray"
	IntelMfr    = "Intel"
	DellMfr     = "Dell"
	GigabyteMfr = "Gigabyte"
	FoxconnMfr  = "Foxconn"
)

// This should only return 1 if the RF manufacturer string (mfrCheckStr) is mfr
// (see above), 0 for not and -1 if mfrCheckStr is blank or non-alpha-numeric.
// This should be used in combination with other checks ideally.
func IsManufacturer(mfrCheckStr, mfr string) int {
	if strings.IndexFunc(mfrCheckStr, func(c rune) bool {
		return unicode.IsLetter(c) || unicode.IsNumber(c)
	}) != -1 {
		lower := strings.ToLower(mfrCheckStr)
		// Split into chunks containing only sequences of letters
		// so we will find cray unless it's a substring of another word.
		f := func(c rune) bool { return !unicode.IsLetter(c) }
		split := strings.FieldsFunc(lower, f)
		for _, s := range split {
			switch mfr {
			case CrayMfr:
				if s == "cray" || s == "crayinc" || s == "crayincorporated" || s == "hpe" {
					return 1
				}
			case IntelMfr:
				if s == "intel" || s == "intelinc" ||
					s == "intelincorporated" || s == "intelcorp" ||
					s == "intelcorporation" {
					return 1
				}
			case DellMfr:
				if s == "dell" || s == "dellinc" || s == "dellcorporation" {
					return 1
				}
			case GigabyteMfr:
				if s == "gigabyte" {
					return 1
				}
			case FoxconnMfr:
				if s == "foxconn" {
					return 1
				}
			}
		}
		return 0
	}
	return -1
}

////////////////////////////////////////////////////////////////////////////
// Processor architecture detection
////////////////////////////////////////////////////////////////////////////

// ProcessorArchitecture Enum
const (
	ProcessorArchARM   string = "ARM"   // ARM
	ProcessorArchIA64  string = "IA-64" // Intel Itanium
	ProcessorArchMIPS  string = "MIPS"  // MIPS
	ProcessorArchOEM   string = "OEM"   // OEM-defined
	ProcessorArchPower string = "Power" // Power
	ProcessorArchX86   string = "x86"   // x86 or x86-64
)

// Processor InstructionSet Enum
const (
	ProcessorInstructionSetARMA32   string = "ARM-A32"  // ARM 32-bit
	ProcessorInstructionSetARMA64   string = "ARM-A64"  // ARM 64-bit
	ProcessorInstructionSetIA64     string = "IA-64"    // Intel IA-64
	ProcessorInstructionSetMIPS32   string = "MIPS32"   // MIPS 32-bit
	ProcessorInstructionSetMIPS64   string = "MIPS64"   // MIPS 64-bit
	ProcessorInstructionSetOEM      string = "OEM"      // OEM-defined
	ProcessorInstructionSetPowerISA string = "PowerISA" // PowerISA-64 or PowerISA-32
	ProcessorInstructionSetX86      string = "x86"      // x86
	ProcessorInstructionSetX86_64   string = "x86-64"   // x86-64
)

// Check the processor's ProcessorArchitecture and InstructionSet fields
// to determine the architecture.
func GetProcessorArch(p *EpProcessor) (procArch string) {
	rfArch := p.ProcessorRF.ProcessorArchitecture
	if rfArch == "" {
		rfArch = p.ProcessorRF.InstructionSet
	}
	switch rfArch {
	case "":
		procArch = base.ArchUnknown.String()
	case ProcessorArchX86:
		fallthrough
	case ProcessorInstructionSetX86_64:
		procArch = base.ArchX86.String()
	case ProcessorArchARM:
		fallthrough
	case ProcessorInstructionSetARMA32:
		fallthrough
	case ProcessorInstructionSetARMA64:
		procArch = base.ArchARM.String()
	default:
		procArch = base.ArchOther.String()
	}
	return
}

////////////////////////////////////////////////////////////////////////////
// System architecture detection
////////////////////////////////////////////////////////////////////////////

// These *ArchMaps are used for working around redfish on hardware not
// supplying the 'ProcessorArchitecture' or 'InstructionSet' fields for the
// processors. They are used for matching known hardware types to a
// processor architecture.
//
// These lists are needed for at least early bring-up of hardware as the Redfish
// does not always provide the needed information in the early firmware.

// Model matching strings for Cray EX hardware.
var CrayEXModelArchMap = map[string]string{
	"ex235":  base.ArchX86.String(), // Grizzly Peak (ex235n), Bard Peak (ex235a)
	"ex420":  base.ArchX86.String(), // Castle (ex420)
	"ex425":  base.ArchX86.String(), // Windom (ex425)
	"ex254":  base.ArchARM.String(), // Blanca Peak (ex254n)
	"ex255":  base.ArchX86.String(), // Parry Peak (ex255a)
	"ex4252": base.ArchX86.String(), // Antero (ex4252)
}

// Drescription matching strings for Cray EX hardware
var CrayEXDescrArchMap = map[string]string{
	"windomnodecard":    base.ArchX86.String(),
	"wnc":               base.ArchX86.String(),
	"cnc":               base.ArchX86.String(),
	"bardpeaknc":        base.ArchX86.String(),
	"grizzlypknodecard": base.ArchX86.String(),
	"antero":            base.ArchX86.String(),
	"blancapeaknc":      base.ArchARM.String(),
	"parrypeaknc":       base.ArchX86.String(),
}

var FoxconnModelArchMap = map[string]string{
	"hpe cray supercomputing xd224": base.ArchARM.String(),	// Paradise (official)
	"1a62wcb00-600-g": base.ArchARM.String(),				// Paradise (some systems slipped to field with this)
}

func GetSystemArch(s *EpSystem) string {
	sysArch := base.ArchUnknown.String()
	// Search the processor collection for the architecture.
	for _, proc := range s.Processors.OIDs {
		if proc.Type != xnametypes.Processor.String() {
			// Skip GPUs
			continue
		}
		if sysArch == base.ArchUnknown.String() ||
			(sysArch == base.ArchOther.String() &&
				proc.Arch != base.ArchUnknown.String()) {
			// Try for the best identification (X86/ARM > Other > UNKNOWN).
			sysArch = proc.Arch
		}
		if sysArch != base.ArchUnknown.String() &&
			sysArch != base.ArchOther.String() {
			// Found x86 or ARM
			break
		}
	}

	// If the Arch is still unknown it might be because Cray-HPE EX* hardware
	// is not supplying the 'ProcessorArchitecture' or 'InstructionSet' fields
	// for the processors or the processor collection wasn't present. Try to
	// determine the processor architecture based on the node's model.
	if sysArch == base.ArchUnknown.String() {
		if IsManufacturer(s.SystemRF.Manufacturer, CrayMfr) == 1 {
			if len(s.SystemRF.Model) > 0 {
				rfModel := strings.ToLower(s.SystemRF.Model)
				for matchStr, arch := range CrayEXModelArchMap {
					if strings.Contains(rfModel, matchStr) {
						return arch
					}
				}
			}
			if len(s.SystemRF.Description) > 0 {
				rfDescr := strings.ToLower(s.SystemRF.Description)
				for matchStr, arch := range CrayEXDescrArchMap {
					if strings.Contains(rfDescr, matchStr) {
						return arch
					}
				}
			}
		}
		if IsManufacturer(s.SystemRF.Manufacturer, FoxconnMfr) == 1 {
			if len(s.SystemRF.Model) > 0 {
				rfModel := strings.ToLower(s.SystemRF.Model)
				for matchStr, arch := range FoxconnModelArchMap {
					if strings.Contains(rfModel, matchStr) {
						return arch
					}
				}
			}
		}
	}
	return sysArch
}
