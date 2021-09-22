// MIT License
//
// (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
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
	base "github.com/Cray-HPE/hms-base"
)

var ErrHWHistEventTypeInvalid = base.NewHMSError("sm", "Invalid hardware inventory history event type")
var ErrHWInvHistFmtInvalid = base.NewHMSError("sm", "Invalid HW Inventory History format")

// Valid values for event types
const (
	HWInvHistEventTypeAdded    = "Added"
	HWInvHistEventTypeRemoved  = "Removed"
	HWInvHistEventTypeScanned  = "Scanned"
	HWInvHistEventTypeDetected = "Detected"
)

// For case-insensitive verification and normalization of state strings
var hwInvHistEventTypeMap = map[string]string{
	"added":    HWInvHistEventTypeAdded,
	"removed":  HWInvHistEventTypeRemoved,
	"scanned":  HWInvHistEventTypeScanned,
	"detected": HWInvHistEventTypeDetected,
}

type HWInvHistFmt int

const (
	HWInvHistFmtByLoc HWInvHistFmt = iota
	HWInvHistFmtByFRU
)

type HWInvHist struct {
	ID        string `json:"ID"`        // xname location where the event happened
	FruId     string `json:"FRUID"`     // FRU ID of the affected FRU
	Timestamp string `json:"Timestamp"` // Timestamp of the event
	EventType string `json:"EventType"` // (i.e. Added, Removed, Scanned)
}

type HWInvHistArray struct {
	ID      string       `json:"ID"`      // xname or FruId (if ByFRU)
	History []*HWInvHist `json:"History"`
}

type HWInvHistResp struct {
	Components []HWInvHistArray `json:"Components"`
}

// Create formatted HWInvHistResp from a random array of HWInvHist entries.
// No sorting is done (with components of the same type), so pre/post-sort if
// needed.
func NewHWInvHistResp(hwHists []*HWInvHist, format HWInvHistFmt) (*HWInvHistResp, error) {
	compHist := new(HWInvHistResp)
	compHistMap := make(map[string]int)
	if !(format == HWInvHistFmtByLoc || format == HWInvHistFmtByFRU) {
		return nil, ErrHWInvHistFmtInvalid
	}
	var idx int
	for _, hwHist := range hwHists {
		id := hwHist.ID
		if format == HWInvHistFmtByFRU {
			id = hwHist.FruId
		}
		index, ok := compHistMap[id]
		if !ok {
			compHistMap[id] = idx
			idx++
			compHistArray := HWInvHistArray{
				ID: id,
				History: []*HWInvHist{hwHist},
			}
			compHist.Components = append(compHist.Components, compHistArray)
		} else {
			compHist.Components[index].History = append(compHist.Components[index].History, hwHist)
		}
	}
	return compHist, nil
}

// Validate and Normalize event types used in queries
func VerifyNormalizeHWInvHistEventType(eventType string) string {
	evtLower := strings.ToLower(eventType)
	value, ok := hwInvHistEventTypeMap[evtLower]
	if !ok {
		return ""
	} else {
		return value
	}
}
