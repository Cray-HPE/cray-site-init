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

// This package defines structures for groups and partitions

import (
	"regexp"
	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"strings"
)

//
// Format checking for database keys and query parameters.
//

var e = base.NewHMSError("sm", "GenericError")

var ErrGroupBadField = base.NewHMSError("sm",
	"group or partition field has invalid characters")
var ErrPartBadName = base.NewHMSError("sm",
	"Bad partition name. Must be p# or p#.#")

// Normalize group field by lowercasing
func NormalizeGroupField(f string) string {
	return strings.ToLower(f)
}

// Verify group field.
func VerifyGroupField(f string) error {
	re := regexp.MustCompile(`^[-a-z0-9:._]+$`)
	if re.MatchString(f) {
		return nil
	}
	return ErrGroupBadField
}

///////////////////////////////////////////////////////////////////////////
//
// Members
//
///////////////////////////////////////////////////////////////////////////

// This just stores a list of component xname ids for now, but could grow if
// we need it to.
type Members struct {
	IDs []string `json:"ids"` // xname array

	// Private
	normalized bool
	verified   bool
}

// Create new empty members
func NewMembers() *Members {
	ms := new(Members)
	ms.IDs = make([]string, 0, 1)
	return ms
}

// For POST to members endpoint to create new member
type MemberAddBody struct {
	ID string `json:"id"` // xname
}

// Check ids array for xname fitneess.  If no error is returned,
// the ids are valid.
func (ms *Members) Verify() error {
	if ms.verified == true {
		return nil
	}
	ms.verified = true
	for _, id := range ms.IDs {
		if ok := xnametypes.IsHMSCompIDValid(id); ok == false {
			return base.ErrHMSTypeInvalid
		}
	}
	return nil
}

// Normalize xnames in Members
func (ms *Members) Normalize() {
	if ms.normalized == true {
		return
	}
	ms.normalized = true
	for i, id := range ms.IDs {
		ms.IDs[i] = xnametypes.NormalizeHMSCompID(id)
	}
}

///////////////////////////////////////////////////////////////////////////
//
// Groups
//
///////////////////////////////////////////////////////////////////////////

// Component Group, typically nodes.   Like a partition but just a free
// form collection, not necessarily non-overlapping, and with no predetermined
// purpose.
type Group struct {
	Label          string   `json:"label"`
	Description    string   `json:"description"`
	ExclusiveGroup string   `json:"exclusiveGroup,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Members        Members  `json:"members"` // List of xnames, required.

	// Private
	normalized bool
	verified   bool
}

// Allocate and initialize new Group struct, validating it.
// This copies tags and members vs. just the pointer to the slice.
// If you already have a created group, you can check the inputs with
// group.Verify()
func NewGroup(label, desc, exclGrp string, tags, member_ids []string) (*Group, error) {
	g := new(Group)
	g.Label = label
	g.Description = desc
	g.ExclusiveGroup = exclGrp
	g.Tags = append([]string(nil), tags...)
	g.Members.IDs = append([]string(nil), member_ids...)
	g.Normalize()
	return g, g.Verify()
}

// Lowercase field names and normalize xnames in Members.
func (g *Group) Normalize() {
	if g.normalized == true {
		return
	}
	g.normalized = true

	g.Label = strings.ToLower(g.Label)
	g.ExclusiveGroup = strings.ToLower(g.ExclusiveGroup)
	for i, f := range g.Tags {
		g.Tags[i] = strings.ToLower(f)
	}
	g.Members.Normalize()
}

// Check input fields of a group.  If no error is returned, the result should
// be ok to put into the database.
func (g *Group) Verify() error {
	if g.verified == true {
		return nil
	}
	g.verified = true

	if err := VerifyGroupField(g.Label); err != nil {
		return err
	}
	if g.ExclusiveGroup != "" {
		if err := VerifyGroupField(g.ExclusiveGroup); err != nil {
			return err
		}
	}
	for _, f := range g.Tags {
		if err := VerifyGroupField(f); err != nil {
			return err
		}
	}
	if err := g.Members.Verify(); err != nil {
		return err
	}
	return nil
}

// Patchable fields if included in payload.
type GroupPatch struct {
	Description *string   `json:"description"`
	Tags        *[]string `json:"tags"`
}

// Normalize groupPatch (just lower case tags, basically, but keeping same
// interface as others.
func (gp *GroupPatch) Normalize() {
	if gp.Tags == nil {
		return
	}
	for i, f := range *gp.Tags {
		(*gp.Tags)[i] = strings.ToLower(f)
	}
}

// Analgous Verify call for GroupPatch objects.
func (gp *GroupPatch) Verify() error {
	if gp.Tags == nil {
		return nil
	}
	for _, f := range *gp.Tags {
		if err := VerifyGroupField(f); err != nil {
			return err
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////
//
// Partitions
//
///////////////////////////////////////////////////////////////////////////

// A partition is a formal, non-overlapping division of the system that forms
// an administratively distinct sub-system e.g. for implementing multi-tenancy.
type Partition struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Members     Members  `json:"members"` // List of xname ids, required.

	// Private
	normalized bool
	verified   bool
}

// Allocate and initialize new Group struct, validating it.
// This copies tags and members vs. just the pointer to the slice.
// If you already have a created group, you can check the inputs with
// group.Verify()
func NewPartition(name, desc string, tags, member_ids []string) (*Partition, error) {
	p := new(Partition)
	p.Name = name
	p.Description = desc
	p.Tags = append([]string(nil), tags...)
	p.Members.IDs = append([]string(nil), member_ids...)
	p.Normalize()
	return p, p.Verify()
}

// Lowercase field names and normalize xnames in Members.
func (p *Partition) Normalize() {
	if p.normalized == true {
		return
	}
	p.normalized = true

	p.Name = strings.ToLower(p.Name)
	for i, f := range p.Tags {
		p.Tags[i] = strings.ToLower(f)
	}
	p.Members.Normalize()
}

// Check input fields of a group.  If no error is returned, the result should
// be ok to put into the database.
func (p *Partition) Verify() error {
	if p.verified == true {
		return nil
	}
	p.verified = true

	if xnametypes.GetHMSType(p.Name) != xnametypes.Partition {
		return ErrPartBadName
	}
	for _, f := range p.Tags {
		if err := VerifyGroupField(f); err != nil {
			return err
		}
	}
	if err := p.Members.Verify(); err != nil {
		return err
	}
	return nil
}

// Patchable fields if included in payload.
type PartitionPatch struct {
	Description *string   `json:"description"`
	Tags        *[]string `json:"tags"`
}

// Normalize PartitionPatch (just lower case tags, basically, but keeping same
// interface as others.
func (pp *PartitionPatch) Normalize() {
	if pp.Tags == nil {
		return
	}
	for i, f := range *pp.Tags {
		(*pp.Tags)[i] = strings.ToLower(f)
	}
}

// Analgous Verify call for PartitionPatch objects.
func (pp *PartitionPatch) Verify() error {
	if pp.Tags == nil {
		return nil
	}
	for _, f := range *pp.Tags {
		if err := VerifyGroupField(f); err != nil {
			return err
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////
//
// Membership - Reverse lookup of group and partition info by component id
//
///////////////////////////////////////////////////////////////////////////

type Membership struct {
	ID            string   `json:"id"`
	GroupLabels   []string `json:"groupLabels"`
	PartitionName string   `json:"partitionName"`
}
