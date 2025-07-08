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
	"fmt"
	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

// The payload for a Components POST
type ComponentsPost struct {
	base.ComponentArray
	Force bool `json:"force"`
}

// The payload for a Component PUT
type ComponentPut struct {
	Component base.Component `json:"component"`
	Force     bool           `json:"force"`
}

// This creates a ComponentsPost payload and verifies that the components are
// valid. At the very least ID and State for each component are required.
func NewCompPost(comps []base.Component, force bool) (*ComponentsPost, error) {
	cp := new(ComponentsPost)
	for _, comp := range comps {
		c := new(base.Component)
		c.ID = xnametypes.VerifyNormalizeCompID(comp.ID)
		if len(c.ID) == 0 {
			err := fmt.Errorf("xname ID '%s' is invalid", comp.ID)
			return nil, err
		}
		c.Type = xnametypes.GetHMSTypeString(c.ID)
		c.State = base.VerifyNormalizeState(comp.State)
		if len(c.State) == 0 {
			err := fmt.Errorf("state '%s' is invalid", comp.State)
			return nil, err
		}
		c.Flag = base.VerifyNormalizeFlagOK(comp.Flag)
		if len(c.Flag) == 0 {
			err := fmt.Errorf("flag '%s' is invalid", comp.Flag)
			return nil, err
		}
		c.Enabled = comp.Enabled
		c.SwStatus = comp.SwStatus
		if len(comp.Role) != 0 {
			c.Role = base.VerifyNormalizeRole(comp.Role)
			if len(c.Role) == 0 {
				err := fmt.Errorf("role '%s' is invalid", comp.Role)
				return nil, err
			}
		}
		if len(comp.SubRole) != 0 {
			c.SubRole = base.VerifyNormalizeSubRole(comp.SubRole)
			if len(c.SubRole) == 0 {
				err := fmt.Errorf("subRole '%s' is invalid", comp.SubRole)
				return nil, err
			}
		}
		c.NID = comp.NID
		c.Subtype = comp.Subtype
		if len(comp.NetType) != 0 {
			c.NetType = base.VerifyNormalizeNetType(comp.NetType)
			if len(c.NetType) == 0 {
				err := fmt.Errorf("netType '%s' is invalid", comp.NetType)
				return nil, err
			}
		}
		if len(comp.Arch) != 0 {
			c.Arch = base.VerifyNormalizeArch(comp.Arch)
			if len(c.Arch) == 0 {
				err := fmt.Errorf("arch '%s' is invalid", comp.Arch)
				return nil, err
			}
		}
		if len(comp.Class) != 0 {
			c.Class = base.VerifyNormalizeClass(comp.Class)
			if len(c.Class) == 0 {
				err := fmt.Errorf("class '%s' is invalid", comp.Class)
				return nil, err
			}
		}
		cp.Components = append(cp.Components, c)
	}
	cp.Force = force
	return cp, nil
}

func (cp *ComponentsPost) VerifyNormalize() error {
	for _, comp := range cp.Components {
		normID := xnametypes.VerifyNormalizeCompID(comp.ID)
		if len(normID) == 0 {
			err := fmt.Errorf("xname ID '%s' is invalid", comp.ID)
			return err
		} else {
			comp.ID = normID
		}
		comp.Type = xnametypes.GetHMSTypeString(comp.ID)
		normState := base.VerifyNormalizeState(comp.State)
		if len(normState) == 0 {
			err := fmt.Errorf("state '%s' is invalid", comp.State)
			return err
		} else {
			comp.State = normState
		}
		normFlag := base.VerifyNormalizeFlagOK(comp.Flag)
		if len(normFlag) == 0 {
			err := fmt.Errorf("flag '%s' is invalid", comp.Flag)
			return err
		} else {
			comp.Flag = normFlag
		}
		if len(comp.Role) != 0 {
			normRole := base.VerifyNormalizeRole(comp.Role)
			if len(normRole) == 0 {
				err := fmt.Errorf("role '%s' is invalid", comp.Role)
				return err
			} else {
				comp.Role = normRole
			}
		}
		if len(comp.SubRole) != 0 {
			normSubRole := base.VerifyNormalizeSubRole(comp.SubRole)
			if len(normSubRole) == 0 {
				err := fmt.Errorf("subRole '%s' is invalid", comp.SubRole)
				return err
			} else {
				comp.SubRole = normSubRole
			}
		}
		if len(comp.NetType) != 0 {
			normNetType := base.VerifyNormalizeNetType(comp.NetType)
			if len(normNetType) == 0 {
				err := fmt.Errorf("netType '%s' is invalid", comp.NetType)
				return err
			} else {
				comp.NetType = normNetType
			}
		}
		if len(comp.Arch) != 0 {
			normArch := base.VerifyNormalizeArch(comp.Arch)
			if len(normArch) == 0 {
				err := fmt.Errorf("arch '%s' is invalid", comp.Arch)
				return err
			} else {
				comp.Arch = normArch
			}
		}
		if len(comp.Class) != 0 {
			normClass := base.VerifyNormalizeClass(comp.Class)
			if len(normClass) == 0 {
				err := fmt.Errorf("class '%s' is invalid", comp.Class)
				return err
			} else {
				comp.Class = normClass
			}
		}
	}
	return nil
}

// This creates a ComponentPut payload and verifies that the component is
// valid. At the very least ID and State for the component are required.
func NewCompPut(comp base.Component, force bool) (*ComponentPut, error) {
	cp := new(ComponentPut)
	c := &cp.Component
	c.ID = xnametypes.VerifyNormalizeCompID(comp.ID)
	if len(c.ID) == 0 {
		err := fmt.Errorf("xname ID '%s' is invalid", comp.ID)
		return nil, err
	}
	c.Type = xnametypes.GetHMSTypeString(c.ID)
	c.State = base.VerifyNormalizeState(comp.State)
	if len(c.State) == 0 {
		err := fmt.Errorf("state '%s' is invalid", comp.State)
		return nil, err
	}
	c.Flag = base.VerifyNormalizeFlagOK(comp.Flag)
	if len(c.Flag) == 0 {
		err := fmt.Errorf("flag '%s' is invalid", comp.Flag)
		return nil, err
	}
	c.Enabled = comp.Enabled
	c.SwStatus = comp.SwStatus
	if len(comp.Role) != 0 {
		c.Role = base.VerifyNormalizeRole(comp.Role)
		if len(c.Role) == 0 {
			err := fmt.Errorf("role '%s' is invalid", comp.Role)
			return nil, err
		}
	}
	if len(comp.SubRole) != 0 {
		c.Role = base.VerifyNormalizeSubRole(comp.SubRole)
		if len(c.SubRole) == 0 {
			err := fmt.Errorf("subRole '%s' is invalid", comp.SubRole)
			return nil, err
		}
	}
	c.NID = comp.NID
	c.Subtype = comp.Subtype
	if len(comp.NetType) != 0 {
		c.NetType = base.VerifyNormalizeNetType(comp.NetType)
		if len(c.NetType) == 0 {
			err := fmt.Errorf("netType '%s' is invalid", comp.NetType)
			return nil, err
		}
	}
	if len(comp.Arch) != 0 {
		c.Arch = base.VerifyNormalizeArch(comp.Arch)
		if len(c.Arch) == 0 {
			err := fmt.Errorf("arch '%s' is invalid", comp.Arch)
			return nil, err
		}
	}
	if len(comp.Class) != 0 {
		c.Class = base.VerifyNormalizeClass(comp.Class)
		if len(c.Class) == 0 {
			err := fmt.Errorf("class '%s' is invalid", comp.Class)
			return nil, err
		}
	}
	cp.Force = force
	return cp, nil
}

func (cp *ComponentPut) VerifyNormalize() error {
	c := &cp.Component
	normID := xnametypes.VerifyNormalizeCompID(c.ID)
	if len(normID) == 0 {
		err := fmt.Errorf("xname ID '%s' is invalid", c.ID)
		return err
	} else {
		c.ID = normID
	}
	c.Type = xnametypes.GetHMSTypeString(c.ID)
	normState := base.VerifyNormalizeState(c.State)
	if len(normState) == 0 {
		err := fmt.Errorf("state '%s' is invalid", c.State)
		return err
	} else {
		c.State = normState
	}
	normFlag := base.VerifyNormalizeFlagOK(c.Flag)
	if len(normFlag) == 0 {
		err := fmt.Errorf("flag '%s' is invalid", c.Flag)
		return err
	} else {
		c.Flag = normFlag
	}
	if len(c.Role) != 0 {
		normRole := base.VerifyNormalizeRole(c.Role)
		if len(normRole) == 0 {
			err := fmt.Errorf("role '%s' is invalid", c.Role)
			return err
		} else {
			c.Role = normRole
		}
	}
	if len(c.SubRole) != 0 {
		normSubRole := base.VerifyNormalizeSubRole(c.SubRole)
		if len(normSubRole) == 0 {
			err := fmt.Errorf("subRole '%s' is invalid", c.SubRole)
			return err
		} else {
			c.SubRole = normSubRole
		}
	}
	if len(c.NetType) != 0 {
		normNetType := base.VerifyNormalizeNetType(c.NetType)
		if len(normNetType) == 0 {
			err := fmt.Errorf("netType '%s' is invalid", c.NetType)
			return err
		} else {
			c.NetType = normNetType
		}
	}
	if len(c.Arch) != 0 {
		normArch := base.VerifyNormalizeArch(c.Arch)
		if len(normArch) == 0 {
			err := fmt.Errorf("arch '%s' is invalid", c.Arch)
			return err
		} else {
			c.Arch = normArch
		}
	}
	if len(c.Class) != 0 {
		normClass := base.VerifyNormalizeClass(c.Class)
		if len(normClass) == 0 {
			err := fmt.Errorf("class '%s' is invalid", c.Class)
			return err
		} else {
			c.Class = normClass
		}
	}
	return nil
}
