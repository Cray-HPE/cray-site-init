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

package sm

import (
	"strings"
)

type SMPatchOp int

const (
	PatchOpInvalid SMPatchOp = 0
	PatchOpAdd     SMPatchOp = 1
	PatchOpRemove  SMPatchOp = 2
	PatchOpReplace SMPatchOp = 3
)

var smPatchOpMap = map[string]SMPatchOp{
	"add":     PatchOpAdd,
	"remove":  PatchOpRemove,
	"replace": PatchOpReplace,
}

type SCNPostSubscription struct {
	Subscriber     string   `json:"Subscriber"`
	Enabled        *bool    `json:"Enabled,omitempty"`
	Roles          []string `json:"Roles,omitempty"`
	SubRoles       []string `json:"SubRoles,omitempty"`
	SoftwareStatus []string `json:"SoftwareStatus,omitempty"`
	States         []string `json:"States,omitempty"`
	Url            string   `json:"Url"`
}

type SCNSubscription struct {
	ID             int64    `json:"ID"`
	Subscriber     string   `json:"Subscriber"`
	Enabled        *bool    `json:"Enabled,omitempty"`
	Roles          []string `json:"Roles,omitempty"`
	SubRoles       []string `json:"SubRoles,omitempty"`
	SoftwareStatus []string `json:"SoftwareStatus,omitempty"`
	States         []string `json:"States,omitempty"`
	Url            string   `json:"Url"`
}

type SCNPatchSubscription struct {
	Op             string   `json:"Op"`
	Enabled        *bool    `json:"Enabled,omitempty"`
	Roles          []string `json:"Roles,omitempty"`
	SubRoles       []string `json:"SubRoles,omitempty"`
	SoftwareStatus []string `json:"SoftwareStatus,omitempty"`
	States         []string `json:"States,omitempty"`
}

type SCNSubscriptionArray struct {
	SubscriptionList []SCNSubscription `json:"SubscriptionList"`
}

type SCNPayload struct {
	Components     []string `json:"Components"`
	Enabled        *bool    `json:"Enabled,omitempty"`
	Flag           string   `json:"Flag,omitempty"`
	Role           string   `json:"Role,omitempty"`
	SubRole        string   `json:"SubRole,omitempty"`
	SoftwareStatus string   `json:"SoftwareStatus,omitempty"`
	State          string   `json:"State,omitempty"`
}

func GetPatchOp(op string) SMPatchOp {
	opInt, ok := smPatchOpMap[strings.ToLower(op)]
	if !ok {
		return PatchOpInvalid
	}
	return opInt
}
