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
	//"bytes"
	"encoding/json"
	//base "github.com/Cray-HPE/hms-base"
	//"io/ioutil"
	//"path"
	//"strings"
	//"time"
)

/////////////////////////////////////////////////////////////////////////////
//
// Redfish service discovery
//
/////////////////////////////////////////////////////////////////////////////

// Struct to be embedded in Redfish service objects.  These are the
// key fields in the storage schema.
type ServiceDescription struct {
	RfEndpointID   string `json:"RedfishEndpointID"` // Key
	RedfishType    string `json:"RedfishType"`       // Key
	RedfishSubtype string `json:"RedfishSubtype,omitempty"`
	UUID           string `json:"UUID"`
	OdataID        string `json:"OdataID"`
}

// This is the AccountService for the corresponding RedfishEP
type EpAccountService struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ServiceDescription

	AccountServiceURL string `json:"accountServiceURL"` // Full URL to this svc
	RootFQDN          string `json:"rootFQDN"`          // i.e. for epRF
	RootHostname      string `json:"rootHostname"`
	RootDomain        string `json:"rootDomain"`

	LastStatus string `json:"lastStatus"`

	AccountServiceRF     AccountService   `json:"accountServiceRF"`
	accountServiceURLRaw *json.RawMessage // `json:"AccountServiceURLRaw"`

	epRF *RedfishEP // Backpointer, for connection details, etc.
}

// Create new struct to discover the AccountService for this RedfishEP
func NewEpAccountService(epRF *RedfishEP, odataID string) *EpAccountService {
	s := new(EpAccountService)
	s.OdataID = odataID
	s.RfEndpointID = epRF.ID
	s.RedfishType = AccountServiceType
	s.LastStatus = NotYetQueried
	s.epRF = epRF
	return s
}

// Contact RedfishEP and discover properties of the AccountService
func (s *EpAccountService) discoverRemotePhase1() {
	// Should never happen
	if s.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for AccountService odataID: %s\n",
			s.OdataID)
		s.LastStatus = EndpointInvalid
		return
	}
	s.AccountServiceURL = s.epRF.FQDN + s.OdataID
	s.RootFQDN = s.epRF.FQDN
	s.RootHostname = s.epRF.Hostname
	s.RootDomain = s.epRF.Domain

	path := s.OdataID
	svcURLJSON, err := s.epRF.GETRelative(path)
	if err != nil || svcURLJSON == nil {
		errlog.Println(err)
		s.LastStatus = HTTPsGetFailed
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", s.epRF.FQDN+path, svcURLJSON)
	}
	s.accountServiceURLRaw = &svcURLJSON
	s.LastStatus = HTTPsGetOk

	// Decode Raw JSON into AccountService Go struct
	if err := json.Unmarshal(svcURLJSON, &s.AccountServiceRF); err != nil {
		errlog.Printf("Bad Decode: %s: %s\n", s.RootFQDN+path, err)
		s.LastStatus = EPResponseFailedDecode
		return
	}
}

// This is the SessionService for the corresponding RedfishEP
type EpSessionService struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ServiceDescription

	SessionServiceURL string `json:"sessionServiceURL"` // Full URL to this svc
	RootFQDN          string `json:"rootFQDN"`          // i.e. for epRF
	RootHostname      string `json:"rootHostname"`
	RootDomain        string `json:"rootDomain"`

	LastStatus string `json:"lastStatus"`

	SessionServiceRF     SessionService   `json:"sessionServiceRF"`
	sessionServiceURLRaw *json.RawMessage // `json:"sessionServiceURLRaw"`

	epRF *RedfishEP // Backpointer, for connection details, etc.
}

// Create new struct to discover the SessionService for this RedfishEP
func NewEpSessionService(epRF *RedfishEP, odataID string) *EpSessionService {
	s := new(EpSessionService)
	s.OdataID = odataID
	s.RfEndpointID = epRF.ID
	s.RedfishType = SessionServiceType
	s.LastStatus = NotYetQueried
	s.epRF = epRF
	return s
}

// Contact RedfishEP and discover properties of the SessionService
func (s *EpSessionService) discoverRemotePhase1() {
	// Should never happen
	if s.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for SessionService odataID: %s\n",
			s.OdataID)
		s.LastStatus = EndpointInvalid
		return
	}
	s.SessionServiceURL = s.epRF.FQDN + s.OdataID
	s.RootFQDN = s.epRF.FQDN
	s.RootHostname = s.epRF.Hostname
	s.RootDomain = s.epRF.Domain

	path := s.OdataID
	svcURLJSON, err := s.epRF.GETRelative(path)
	if err != nil || svcURLJSON == nil {
		errlog.Println(err)
		s.LastStatus = HTTPsGetFailed
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", s.epRF.FQDN+path, svcURLJSON)
	}
	s.sessionServiceURLRaw = &svcURLJSON
	s.LastStatus = HTTPsGetOk

	// Decode Raw JSON into AccountService Go struct
	if err := json.Unmarshal(svcURLJSON, &s.SessionServiceRF); err != nil {
		errlog.Printf("Bad Decode: %s: %s\n", s.RootFQDN+path, err)
		s.LastStatus = EPResponseFailedDecode
		return
	}
}

// This is the EventService for the corresponding RedfishEP
type EpEventService struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ServiceDescription

	EventServiceURL string `json:"eventServiceURL"` // Full URL to this svc
	RootFQDN        string `json:"rootFQDN"`        // i.e. for epRF
	RootHostname    string `json:"rootHostname"`
	RootDomain      string `json:"rootDomain"`

	LastStatus string `json:"lastStatus"`

	EventServiceRF     EventService     `json:"eventServiceRF"`
	eventServiceURLRaw *json.RawMessage // `json:"eventServiceURLRaw"`

	epRF *RedfishEP // Backpointer, for connection details, etc.
}

// Create new struct to discover the EventService for this RedfishEP
func NewEpEventService(epRF *RedfishEP, odataID string) *EpEventService {
	s := new(EpEventService)
	s.OdataID = odataID
	s.RfEndpointID = epRF.ID
	s.RedfishType = EventServiceType
	s.LastStatus = NotYetQueried
	s.epRF = epRF
	return s
}

// Contact RedfishEP and discover properties of the EventService
func (s *EpEventService) discoverRemotePhase1() {
	// Should never happen
	if s.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for EventService odataID: %s\n",
			s.OdataID)
		s.LastStatus = EndpointInvalid
		return
	}
	s.EventServiceURL = s.epRF.FQDN + s.OdataID
	s.RootFQDN = s.epRF.FQDN
	s.RootHostname = s.epRF.Hostname
	s.RootDomain = s.epRF.Domain

	path := s.OdataID
	svcURLJSON, err := s.epRF.GETRelative(path)
	if err != nil || svcURLJSON == nil {
		errlog.Println(err)
		s.LastStatus = HTTPsGetFailed
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", s.epRF.FQDN+path, svcURLJSON)
	}
	s.eventServiceURLRaw = &svcURLJSON
	s.LastStatus = HTTPsGetOk

	// Decode Raw JSON into AccountService Go struct
	if err := json.Unmarshal(svcURLJSON, &s.EventServiceRF); err != nil {
		errlog.Printf("Bad Decode: %s: %s\n", s.RootFQDN+path, err)
		s.LastStatus = EPResponseFailedDecode
		return
	}
}

// This is the TaskService for the corresponding RedfishEP
type EpTaskService struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ServiceDescription

	TaskServiceURL string `json:"taskServiceURL"` // Full URL to this svc
	RootFQDN       string `json:"rootFQDN"`       // i.e. for epRF
	RootHostname   string `json:"rootHostname"`
	RootDomain     string `json:"rootDomain"`

	LastStatus string `json:"lastStatus"`

	TaskServiceRF     TaskService      `json:"taskServiceRF"`
	taskServiceURLRaw *json.RawMessage // `json:"taskServiceURLRaw"`

	epRF *RedfishEP // Backpointer, for connection details, etc.
}

// Create new struct to discover the TaskService for this RedfishEP
func NewEpTaskService(epRF *RedfishEP, odataID string) *EpTaskService {
	s := new(EpTaskService)
	s.OdataID = odataID
	s.RfEndpointID = epRF.ID
	s.RedfishType = TaskServiceType
	s.LastStatus = NotYetQueried
	s.epRF = epRF
	return s
}

// Contact RedfishEP and discover properties of the TaskService
func (s *EpTaskService) discoverRemotePhase1() {
	// Should never happen
	if s.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for TaskService odataID: %s\n",
			s.OdataID)
		s.LastStatus = EndpointInvalid
		return
	}
	s.TaskServiceURL = s.epRF.FQDN + s.OdataID
	s.RootFQDN = s.epRF.FQDN
	s.RootHostname = s.epRF.Hostname
	s.RootDomain = s.epRF.Domain

	path := s.OdataID
	svcURLJSON, err := s.epRF.GETRelative(path)
	if err != nil || svcURLJSON == nil {
		errlog.Println(err)
		s.LastStatus = HTTPsGetFailed
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", s.epRF.FQDN+path, svcURLJSON)
	}
	s.taskServiceURLRaw = &svcURLJSON
	s.LastStatus = HTTPsGetOk

	// Decode Raw JSON into AccountService Go struct
	if err := json.Unmarshal(svcURLJSON, &s.TaskServiceRF); err != nil {
		errlog.Printf("Bad Decode: %s: %s\n", s.RootFQDN+path, err)
		s.LastStatus = EPResponseFailedDecode
		return
	}
}

// This is the UpdateService for the corresponding RedfishEP
type EpUpdateService struct {
	// Embedded struct: id, type, odataID and associated RfEndpointID.
	ServiceDescription

	UpdateServiceURL string `json:"updateServiceURL"` // Full URL to this svc
	RootFQDN         string `json:"rootFQDN"`         // i.e. for epRF
	RootHostname     string `json:"rootHostname"`
	RootDomain       string `json:"rootDomain"`

	LastStatus string `json:"lastStatus"`

	UpdateServiceRF     UpdateService    `json:"updateServiceRF"`
	updateServiceURLRaw *json.RawMessage // `json:"eventServiceURLRaw"`

	epRF *RedfishEP // Backpointer, for connection details, etc.
}

// Create new struct to discover the UpdateService for this RedfishEP
func NewEpUpdateService(epRF *RedfishEP, odataID string) *EpUpdateService {
	s := new(EpUpdateService)
	s.OdataID = odataID
	s.RfEndpointID = epRF.ID
	s.RedfishType = UpdateServiceType
	s.LastStatus = NotYetQueried
	s.epRF = epRF
	return s
}

// Contact RedfishEP and discover properties of the UpdateService
func (s *EpUpdateService) discoverRemotePhase1() {
	// Should never happen
	if s.epRF == nil {
		errlog.Printf("Error: RedfishEP == nil for UpdateService odataID: %s\n",
			s.OdataID)
		s.LastStatus = EndpointInvalid
		return
	}
	s.UpdateServiceURL = s.epRF.FQDN + s.OdataID
	s.RootFQDN = s.epRF.FQDN
	s.RootHostname = s.epRF.Hostname
	s.RootDomain = s.epRF.Domain

	path := s.OdataID
	svcURLJSON, err := s.epRF.GETRelative(path)
	if err != nil || svcURLJSON == nil {
		errlog.Println(err)
		s.LastStatus = HTTPsGetFailed
		return
	}
	if rfDebug > 0 {
		errlog.Printf("%s: %s\n", s.epRF.FQDN+path, svcURLJSON)
	}
	s.updateServiceURLRaw = &svcURLJSON
	s.LastStatus = HTTPsGetOk

	// Decode Raw JSON into UpdateService Go struct
	if err := json.Unmarshal(svcURLJSON, &s.UpdateServiceRF); err != nil {
		errlog.Printf("Bad Decode: %s: %s\n", s.RootFQDN+path, err)
		s.LastStatus = EPResponseFailedDecode
		return
	}
}
