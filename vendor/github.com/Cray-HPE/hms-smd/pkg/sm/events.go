// MIT License
//
// (C) Copyright [2018-2021] Hewlett Packard Enterprise Development LP
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
	base "github.com/Cray-HPE/hms-base"
)

type SMEventType string

const (
	NodeStateChange       SMEventType = "NodeStateChange"
	StateChange           SMEventType = "StateChange"
	RedfishEndpointChange SMEventType = "RedfishEndpointChange"
	HWInventoryChange     SMEventType = "HWInventoryChange"
)

type SMEventSubtype string

const (
	StateTransitionOK       SMEventSubtype = "StateTransitionOK"       // StateChange - Successful state change
	StateTransitionAbnormal SMEventSubtype = "StateTransitionAbnormal" // StateChange - Change due to problem, warn/alert
	StateTransitionDisable  SMEventSubtype = "StateTransitionDisable"  // StateChange
	StateTransitionEnable   SMEventSubtype = "StateTransitionEnable"   // StateChange
	NodeAvailable           SMEventSubtype = "NodeAvailable"           // NodeStateChange
	NodeUnavailable         SMEventSubtype = "NodeUnavailable"         // NodeStateChange
	NodeFailed              SMEventSubtype = "NodeFailed"              // NodeStateChange
	NodeStandby             SMEventSubtype = "NodeStandby"             // NodeStateChange
	NodeRoleChanged         SMEventSubtype = "NodeRoleChanged"         // NodeStateChange
	NodeSubRoleChanged      SMEventSubtype = "NodeSubRoleChanged"      // NodeStateChange
	NodeNIDChanged          SMEventSubtype = "NodeNIDChanged"          // NodeStateChange
	RedfishEndpointAdded    SMEventSubtype = "RedfishEndpointAdded"    // RedfishEndpointChange
	RedfishEndpointModified SMEventSubtype = "RedfishEndpointModified" // RedfishEndpointChange
	RedfishEndpointEnabled  SMEventSubtype = "RedfishEndpointEnabled"  // RedfishEndpointChange
	RedfishEndpointDisabled SMEventSubtype = "RedfishEndpointDisabled" // RedfishEndpointChange
	RedfishEndpointRemoved  SMEventSubtype = "RedfishEndpointRemoved"  // RedfishEndpointChange
	HWInventoryAdded        SMEventSubtype = "HWInventoryAdded"        // HWInventoryChange
	HWInventoryModifed      SMEventSubtype = "HWInventoryModified"     // HWInventoryChange
	HWInventoryRemoved      SMEventSubtype = "HWInventoryRemoved"      // HWInventoryChange
)

type SMEvent struct {
	EventType    string `json:"EventType"`
	EventSubtype string `json:"EventSubtype"`

	// At least one of, as per event type:
	ComponentArray       *base.ComponentArray  `json:"ComponentArray,omitempty"`
	HWInventory          *SystemHWInventory    `json:"HWInventory,omitempty"`
	RedfishEndpointArray *RedfishEndpointArray `json:"RedfishEndpointArray,omitempty"`
}

type SMEventArray struct {
	Name      string     `json:"Name,omitempty"`
	Version   string     `json:"Version"`
	Timestamp string     `json:"Timestamp"`
	Events    []*SMEvent `json:"Events"`
}
