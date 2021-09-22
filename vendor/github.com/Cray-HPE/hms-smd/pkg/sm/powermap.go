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
	"fmt"
	base "github.com/Cray-HPE/hms-base"
)

// An entry mapping an xname to a power supplies
type PowerMap struct {
	ID        string   `json:"id"`
	PoweredBy []string `json:"poweredBy,omitempty"`
}

// This wraps basic PowerMap data with the structure used for query responses.
func NewPowerMap(id string, poweredBy []string) (*PowerMap, error) {
	m := new(PowerMap)
	idNorm := base.VerifyNormalizeCompID(id)
	if idNorm == "" {
		err := fmt.Errorf("xname ID '%s' is invalid", id)
		return nil, err
	}
	m.ID = idNorm
	if len(poweredBy) > 0 {
		for _, pwrId := range poweredBy {
			normPwrID := base.VerifyNormalizeCompID(pwrId)
			if normPwrID == "" {
				err := fmt.Errorf("Power supply xname ID '%s' is invalid", pwrId)
				return nil, err
			} else {
				m.PoweredBy = append(m.PoweredBy, normPwrID)
			}
		}
	}
	return m, nil
}
