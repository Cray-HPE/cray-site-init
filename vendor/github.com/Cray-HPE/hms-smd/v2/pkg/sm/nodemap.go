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
	"encoding/json"
	"fmt"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

// An entry mapping a node xname to a NID
type NodeMap struct {
	ID       string           `json:"ID"`
	NID      int              `json:"NID"`
	Role     string           `json:"Role,omitempty"`
	SubRole  string           `json:"SubRole,omitempty"`
	NodeInfo *json.RawMessage `json:"NodeInfo,omitempty"`
}

// Named array of NodeMap entries, for representing a collection of
// them.
type NodeMapArray struct {
	NodeMaps []*NodeMap `json:"NodeMaps"`
}

// This wraps basic RedfishEndpointDescription data with the structure
// used for query responses.
func NewNodeMap(id, role, subRole string, nid int, nodeInfo *json.RawMessage) (*NodeMap, error) {
	m := new(NodeMap)
	idNorm := xnametypes.NormalizeHMSCompID(id)
	if !(xnametypes.GetHMSType(idNorm) == xnametypes.Node || xnametypes.GetHMSType(idNorm) == xnametypes.VirtualNode) {
		err := fmt.Errorf("xname ID '%s' is invalid or not a node", id)
		return nil, err
	}
	m.ID = idNorm
	if nid > 0 {
		m.NID = nid
	} else {
		err := fmt.Errorf("NID '%d' is out of range", nid)
		return nil, err
	}
	if role != "" {
		normRole := base.VerifyNormalizeRole(role)
		if normRole != "" {
			m.Role = normRole
		} else {
			err := fmt.Errorf("Role '%s' is not valid.", role)
			return nil, err
		}
	}
	if subRole != "" {
		normSubRole := base.VerifyNormalizeSubRole(subRole)
		if normSubRole != "" {
			m.SubRole = normSubRole
		} else {
			err := fmt.Errorf("SubRole '%s' is not valid.", subRole)
			return nil, err
		}
	}
	if nodeInfo != nil {
		m.NodeInfo = nodeInfo
	}
	return m, nil
}
