// MIT License
//
// (C) Copyright [2018-2022] Hewlett Packard Enterprise Development LP
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

package base

import (
	"encoding/json"
	"strings"
)

// Use HMS-wrapped errors.  Subsequent errors will be children of this one.
var e = NewHMSError("hms", "GenericError")

//
// State field used in component, set in response to events by state manager.
// 1.0.0
//
type HMSState string

// Valid state values for components - should refect hardware state
// Enabled/Disabled is a separate boolean field, as the component should
// still have it's actual physical state known and tracked at all times, so
// we know what it is when it is enabled.  It also avoids the primary case
// where admins need to modify the state field manually.
//
// NOTE: there will be no state between on and ready.  If the managed plane
// software does not have heartbeats, On is as high as it will ever get.
// So "active" is not useful.   'Paused' is not in scope now that the software
// status field exists.
const (
	StateUnknown   HMSState = "Unknown"   // The State is unknown.  Appears missing but has not been confirmed as empty.
	StateEmpty     HMSState = "Empty"     // The location is not populated with a component
	StatePopulated HMSState = "Populated" // Present (not empty), but no further track can or is being done.
	StateOff       HMSState = "Off"       // Present but powered off
	StateOn        HMSState = "On"        // Powered on.  If no heartbeat mechanism is available, it's software state may be unknown.

	StateStandby HMSState = "Standby" // No longer Ready and presumed dead.  It typically means HB has been lost (w/alert).
	StateHalt    HMSState = "Halt"    // No longer Ready and halted.  OS has been gracefully shutdown or panicked (w/ alert).
	StateReady   HMSState = "Ready"   // Both On and Ready to provide its expected services, i.e. used for jobs.

	//  Retired (actually never used) states:
	// StateActive    HMSState = "Active"    // If level-2 systems without hb monitoring can make a distinction between on and booting/booted.
	// StatePaused    HMSState = "Paused"    // Was in a Ready state, but is temporarily unavailable due to admin action or a transient issue.
)

var ErrHMSStateInvalid = e.NewChild("was not a valid HMS state")
var ErrHMSStateUnsupported = e.NewChild("HMSState value not supported for this operation")
var ErrHMSNeedForce = e.NewChild("operation not allowed and not forced.")
var ErrHMSTypeInvalid = e.NewChild("got HMSTypeInvalid instead of valid type")
var ErrHMSTypeUnsupported = e.NewChild("HMSType value not supported for this operation") // TODO should this be in base?

// For case-insensitive verification and normalization of state strings
var hmsStateMap = map[string]HMSState{
	"unknown":   StateUnknown,
	"empty":     StateEmpty,
	"populated": StatePopulated,
	"off":       StateOff,
	"on":        StateOn,
	"standby":   StateStandby,
	"halt":      StateHalt,
	"ready":     StateReady,
}

func GetHMSStateList() []string {
	hmsStateList := []string{}
	for _, state := range hmsStateMap {
		hmsStateList = append(hmsStateList, state.String())
	}
	return hmsStateList
}

// Returns the given state string (adjusting any capitalization differences),
// if a valid state is given.  Else, return the empty string.
func VerifyNormalizeState(stateStr string) string {
	stateLower := strings.ToLower(stateStr)
	value, ok := hmsStateMap[stateLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// Specifies valid STARTING states before changing to the indicated state,
// at least without forcing the change, which would normally be a bad idea.
// An empty array means "None without forcing.
var hmsValidStartStatesMap = map[string][]string{
	"unknown":   []string{}, // Force/HSM only
	"empty":     []string{}, // Force/HSM only
	"populated": []string{}, // Force/HSM only
	"off":       []string{string(StateOff), string(StateOn), string(StateStandby), string(StateHalt), string(StateReady)},
	"on":        []string{string(StateOn), string(StateOff), string(StateStandby), string(StateHalt)},
	"standby":   []string{string(StateStandby), string(StateReady)},
	"halt":      []string{string(StateHalt), string(StateReady)},
	"ready":     []string{string(StateReady), string(StateOn), string(StateOff), string(StateStandby), string(StateHalt)}, // Last three are needed (for now) if RF events break.
}

// If ok == true, beforeStates contain valid current states a
// component can be in if it is being transitioned to afterState without
// being forced (either because it is a bad idea, or the state should
// only be set by HSM and not by other software).  An empty array means 'None
// without force=true
//
// If ok == false, afterState matched no valid HMS State (case insensitive)
func GetValidStartStates(afterState string) (beforeStates []string, ok bool) {
	stateLower := strings.ToLower(afterState)
	beforeStates, ok = hmsValidStartStatesMap[stateLower]
	return
}

// Same as above, but with force flag.  If not found, returns
// ErrHMSStateInvalid.  If can only be forced, and force = false,
// error will be ErrHMSNeedForce.   Otherwise list of starting states.
// If force = true and no errors, an empty array means no restrictions.
func GetValidStartStateWForce(
	afterState string,
	force bool,
) (beforeStates []string, err error) {

	beforeStates = []string{}
	// See if transition is valid.
	if force == false {
		var ok bool
		beforeStates, ok = GetValidStartStates(afterState)
		if !ok {
			err = ErrHMSStateInvalid
		} else if len(beforeStates) == 0 {
			err = ErrHMSNeedForce
		}
	}
	return
}

// Check to see if the state is above on (on is the highest we will get
// from Redfish, so these are state set by higher software layers)
func IsPostBootState(stateStr string) bool {
	stateLower := strings.ToLower(stateStr)
	value, ok := hmsStateMap[stateLower]
	if ok != true {
		return false
	} else {
		switch value {
		//case StateActive:
		//	fallthrough
		case StateStandby:
			fallthrough
		case StateHalt:
			fallthrough
		case StateReady:
			return true
		//case StatePaused:
		//	return true
		default:
			return false
		}
	}
}

// Allow HMSState to be treated as a standard string type.
func (s HMSState) String() string { return string(s) }

//
// Flag field used in component, set in response to events by state manager.
// 1.0.0
//

type HMSFlag string

// Valid flag values.
const (
	FlagUnknown HMSFlag = "Unknown"
	FlagOK      HMSFlag = "OK"      // Functioning properly
	FlagWarning HMSFlag = "Warning" // Continues to operate, but has an issue that may require attention.
	FlagAlert   HMSFlag = "Alert"   // No longer operating as expected.  The state may also have changed due to error.
	FlagLocked  HMSFlag = "Locked"  // Another service has reserved this component.
)

// For case-insensitive verification and normalization of flag strings
var hmsFlagMap = map[string]HMSFlag{
	"unknown": FlagUnknown,
	"ok":      FlagOK,
	"warning": FlagWarning,
	"warn":    FlagWarning,
	"alert":   FlagAlert,
	"locked":  FlagLocked,
}

// Get a list of all valid HMS flags
func GetHMSFlagList() []string {
	hmsFlagList := []string{}
	for _, flag := range hmsFlagMap {
		hmsFlagList = append(hmsFlagList, flag.String())
	}
	return hmsFlagList
}

// Returns the given flag string (adjusting any capitalization differences),
// if a valid flag was given.  Else, return the empty string.
func VerifyNormalizeFlag(flagStr string) string {
	flagLower := strings.ToLower(flagStr)
	value, ok := hmsFlagMap[flagLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// As above, but if flag is the empty string, then return FlagOK.
// If non-empty and invalid, return the empty string.
func VerifyNormalizeFlagOK(flag string) string {
	if flag == "" {
		return FlagOK.String()
	}
	return VerifyNormalizeFlag(flag)
}

// Allow HMSFlag to be treated as a standard string type.
func (f HMSFlag) String() string { return string(f) }

//
// Role of component
// 1.0.0
//

type HMSRole string

// Valid role values.
const (
	RoleCompute     HMSRole = "Compute"
	RoleService     HMSRole = "Service"
	RoleSystem      HMSRole = "System"
	RoleApplication HMSRole = "Application"
	RoleStorage     HMSRole = "Storage"
	RoleManagement  HMSRole = "Management"
)

// For case-insensitive verification and normalization of role strings
var defaultHMSRoleMap = map[string]string{
	"compute":     RoleCompute.String(),
	"service":     RoleService.String(),
	"system":      RoleSystem.String(),
	"application": RoleApplication.String(),
	"storage":     RoleStorage.String(),
	"management":  RoleManagement.String(),
}

var hmsRoleMap = defaultHMSRoleMap

// Get a list of all valid HMS roles
func GetHMSRoleList() []string {
	hmsRoleList := []string{}
	for _, role := range hmsRoleMap {
		hmsRoleList = append(hmsRoleList, role)
	}
	return hmsRoleList
}

// Returns the given role string (adjusting any capitalization differences),
// if a valid role was given.  Else, return the empty string.
func VerifyNormalizeRole(roleStr string) string {
	roleLower := strings.ToLower(roleStr)
	value, ok := hmsRoleMap[roleLower]
	if ok != true {
		return ""
	} else {
		return value
	}
}

// Allow HMSRole to be treated as a standard string type.
func (r HMSRole) String() string { return string(r) }

//
// SubRole of component
// 1.0.0
//

type HMSSubRole string

// Valid SubRole values.
const (
	SubRoleMaster  HMSSubRole = "Master"
	SubRoleWorker  HMSSubRole = "Worker"
	SubRoleStorage HMSSubRole = "Storage"
)

// For case-insensitive verification and normalization of SubRole strings
var defaultHMSSubRoleMap = map[string]string{
	"master":  SubRoleMaster.String(),
	"worker":  SubRoleWorker.String(),
	"storage": SubRoleStorage.String(),
}

var hmsSubRoleMap = defaultHMSSubRoleMap

// Get a list of all valid HMS subroles
func GetHMSSubRoleList() []string {
	hmsSubRoleList := []string{}
	for _, subrole := range hmsSubRoleMap {
		hmsSubRoleList = append(hmsSubRoleList, subrole)
	}
	return hmsSubRoleList
}

// Returns the given SubRole string (adjusting any capitalization differences),
// if a valid SubRole was given.  Else, return the empty string.
func VerifyNormalizeSubRole(subRoleStr string) string {
	subRoleLower := strings.ToLower(subRoleStr)
	value, ok := hmsSubRoleMap[subRoleLower]
	if ok != true {
		return ""
	} else {
		return value
	}
}

// Allow HMSSubRole to be treated as a standard string type.
func (r HMSSubRole) String() string { return string(r) }

//
// HMSNetType - type of high speed network
// 1.0.0
//

type HMSNetType string

const (
	NetSling      HMSNetType = "Sling"
	NetInfiniband HMSNetType = "Infiniband"
	NetEthernet   HMSNetType = "Ethernet"
	NetOEM        HMSNetType = "OEM" // Placeholder for non-slingshot
	NetNone       HMSNetType = "None"
)

// For case-insensitive verification and normalization of HSN network types
var hmsNetTypeMap = map[string]HMSNetType{
	"sling":      NetSling,
	"infiniband": NetInfiniband,
	"ethernet":   NetEthernet,
	"oem":        NetOEM,
	"none":       NetNone,
}

// Get a list of all valid HMS NetTypes
func GetHMSNetTypeList() []string {
	hmsNetTypeList := []string{}
	for _, netType := range hmsNetTypeMap {
		hmsNetTypeList = append(hmsNetTypeList, netType.String())
	}
	return hmsNetTypeList
}

// Returns the given net type string (adjusting any capitalization differences),
// if a valid netType was given.  Else, return the empty string.
func VerifyNormalizeNetType(netTypeStr string) string {
	netTypeLower := strings.ToLower(netTypeStr)
	value, ok := hmsNetTypeMap[netTypeLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// Allow HMSNetType to be treated as a standard string type.
func (r HMSNetType) String() string { return string(r) }

//
// HMSArch - binary type needed for component
// 1.0.0
//

type HMSArch string

const (
	ArchX86     HMSArch = "X86"
	ArchARM     HMSArch = "ARM"
	ArchUnknown HMSArch = "UNKNOWN"
	ArchOther   HMSArch = "Other"
)

// For case-insensitive verification and normalization of HSN network types
var hmsArchMap = map[string]HMSArch{
	"x86":     ArchX86,
	"arm":     ArchARM,
	"unknown": ArchUnknown,
	"other":   ArchOther,
}

// Get a list of all valid HMS Arch
func GetHMSArchList() []string {
	hmsArchList := []string{}
	for _, arch := range hmsArchMap {
		hmsArchList = append(hmsArchList, arch.String())
	}
	return hmsArchList
}

// Returns the given arch string (adjusting any capitalization differences),
// if a valid arch was given.  Else, return the empty string.
func VerifyNormalizeArch(archStr string) string {
	archLower := strings.ToLower(archStr)
	value, ok := hmsArchMap[archLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// Allow HMSArch to be treated as a standard string type.
func (r HMSArch) String() string { return string(r) }

//
// HMSClass - Physical hardware profile
// 1.0.0
//

type HMSClass string

const (
	ClassRiver    HMSClass = "River"
	ClassMountain HMSClass = "Mountain"
	ClassHill     HMSClass = "Hill"
)

// For case-insensitive verification and normalization of HMS Class
var hmsClassMap = map[string]HMSClass{
	"river":    ClassRiver,
	"mountain": ClassMountain,
	"hill":     ClassHill,
}

// Get a list of all valid HMS Class
func GetHMSClassList() []string {
	hmsClassList := []string{}
	for _, class := range hmsClassMap {
		hmsClassList = append(hmsClassList, class.String())
	}
	return hmsClassList
}

// Returns the given class string (adjusting any capitalization differences),
// if a valid class was given.  Else, return the empty string.
func VerifyNormalizeClass(classStr string) string {
	classLower := strings.ToLower(classStr)
	value, ok := hmsClassMap[classLower]
	if ok != true {
		return ""
	} else {
		return value.String()
	}
}

// Allow HMSClass to be treated as a standard string type.
func (r HMSClass) String() string { return string(r) }

//
// This is the equivalent to rs_node_t in Cascade.  It is the minimal
// amount of of information for tracking component state and other vital
// info at an abstract level.  The hwinv is for component-type specific
// fields and detailed HW attributes, i.e. just like XC.
//
// For most HMS operations (and non-inventory ones in the managed plane)
// this info should be sufficient.  We want to keep it minimal for speed.
// Those fields that are not fixed at discovery should be those that can
// change outside of discovery in response to system activity, i.e. hwinv
// should contain only fields that are basically static between discoveries
// of the endpoint.   Things like firmware versions might be an exception,
// but that would be a separate process SM would
//
// 1.0.0
//
type Component struct {
	ID                  string      `json:"ID"`
	Type                string      `json:"Type"`
	State               string      `json:"State,omitempty"`
	Flag                string      `json:"Flag,omitempty"`
	Enabled             *bool       `json:"Enabled,omitempty"`
	SwStatus            string      `json:"SoftwareStatus,omitempty"`
	Role                string      `json:"Role,omitempty"`
	SubRole             string      `json:"SubRole,omitempty"`
	NID                 json.Number `json:"NID,omitempty"`
	Subtype             string      `json:"Subtype,omitempty"`
	NetType             string      `json:"NetType,omitempty"`
	Arch                string      `json:"Arch,omitempty"`
	Class               string      `json:"Class,omitempty"`
	ReservationDisabled bool        `json:"ReservationDisabled,omitempty"`
	Locked              bool        `json:"Locked,omitempty"`
}

// A collection of 0-n Components.  It could just be an ordinary
// array but we want to save the option to have indentifying info, etc.
// packaged with it, e.g. the query parameters or options that produced it,
// especially if there are fewer fields than normal being included.
type ComponentArray struct {
	Components []*Component `json:"Components"`
}
