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

// This file is contains struct defines for CompLocks

package sm

// This package defines structures for component locks

import (
	"math"
	"strings"

	base "github.com/Cray-HPE/hms-base"
)

//
// Format checking for database keys and query parameters.
//

var ErrCompLockBadLifetime = base.NewHMSError("sm",
	"Invalid CompLock lifetime")

///////////////////////////////////////////////////////////////////////////
//
// CompLock
//
///////////////////////////////////////////////////////////////////////////

// A component lock is a formal, non-overlapping group of components that are
// reserved by a service.
type CompLock struct {
	ID       string   `json:"id"`
	Created  string   `json:"created,omitempty"`
	Reason   string   `json:"reason"`
	Owner    string   `json:"owner"`
	Lifetime int      `json:"lifetime"`
	Xnames   []string `json:"xnames"` // List of xname ids, required.

	// Private
	normalized bool
	verified   bool
}

// Allocate and initialize new CompLock struct, validating it.
// If you already have a created CompLock, you can check the inputs with
// CompLock.Verify()
func NewCompLock(reason, owner string, lifetime int, xnames []string) (*CompLock, error) {
	cl := new(CompLock)
	cl.Reason = reason
	cl.Owner = owner
	// Round lifetime up to the closest minute to be compatible with v2 reservations
	cl.Lifetime = int(math.Ceil(float64(lifetime) / 60) * 60)
	cl.Xnames = append([]string(nil), xnames...)
	cl.Normalize()
	return cl, cl.Verify()
}

// Normalize xnames in CompLockMembers.
func (cl *CompLock) Normalize() {
	if cl.normalized == true {
		return
	}
	cl.normalized = true

	cl.Lifetime = int(math.Ceil(float64(cl.Lifetime) / 60) * 60)
	for i, xname := range cl.Xnames {
		cl.Xnames[i] = base.NormalizeHMSCompID(xname)
	}
}

// Check input fields of a group.  If no error is returned, the result should
// be ok to put into the database.
func (cl *CompLock) Verify() error {
	if cl.verified == true {
		return nil
	}
	cl.verified = true

    // Can't be greater than 15 mins
	if cl.Lifetime <= 0 || cl.Lifetime > 900 {
		return ErrCompLockBadLifetime
	}

	for _, xname := range cl.Xnames {
		if ok := base.IsHMSCompIDValid(xname); ok == false {
			return base.ErrHMSTypeInvalid
		}
	}
	return nil
}

// Patchable fields if included in payload.
type CompLockPatch struct {
	Reason   *string
	Owner    *string
	Lifetime *int
}

// Normalize CompLockPatch (just lower case tags, basically, but keeping same
// interface as others.
func (clp *CompLockPatch) Normalize() {
	if clp.Lifetime != nil {
		// Round lifetime up to the closest minute to be compatible with v2 reservations
		*clp.Lifetime = int(math.Ceil(float64(*clp.Lifetime) / 60) * 60)
	}
	return
}

// Analgous Verify call for CompLockPatch objects.
func (clp *CompLockPatch) Verify() error {
	if clp.Lifetime != nil && *clp.Lifetime <= 0 {
		return ErrCompLockBadLifetime
	}
	return nil
}

//////////////////////////////
// Locks V2
//////////////////////////////

var ErrCompLockV2BadProcessingModel = base.NewHMSError("sm",
	"Invalid Processing Model")
var ErrCompLockV2BadDuration = base.NewHMSError("sm",
	"Invalid Reservation Duration")
var ErrCompLockV2CompLocked = base.NewHMSError("sm",
	"Component is Locked")
var ErrCompLockV2CompUnlocked = base.NewHMSError("sm",
	"Component is Unlocked")
var ErrCompLockV2CompDisabled = base.NewHMSError("sm",
	"Component reservations are disabled")
var ErrCompLockV2CompReserved = base.NewHMSError("sm",
	"Component is Reserved")
var ErrCompLockV2DupLock = base.NewHMSError("sm",
	"Component is Locked")
var ErrCompLockV2NotFound = base.NewHMSError("sm",
	"Component not found")
var ErrCompLockV2Unknown = base.NewHMSError("sm",
	"Unknown locking error")
var ErrCompLockV2RKey = base.NewHMSError("sm",
	"Reservation Key required for operation")
var ErrCompLockV2DKey = base.NewHMSError("sm",
	"Deputy Key required for operation")

const (
	CLProcessingModelRigid = "rigid"
	CLProcessingModelFlex  = "flexible"
)

var processingModelMap = map[string]bool{
	CLProcessingModelRigid: true,
	CLProcessingModelFlex: true,
}

func VerifyNormalizeProcessingModel(pm string) string {
	if pm == "" {
		return CLProcessingModelRigid
	}
	pmLower := strings.ToLower(pm)
	_, ok := processingModelMap[pmLower]
	if ok != true {
		return ""
	} else {
		return pmLower
	}
}

const (
	CLResultSuccess     = "Success"
	CLResultNotFound    = "NotFound"
	CLResultLocked      = "Locked"
	CLResultUnlocked    = "Unlocked"
	CLResultDisabled    = "Disabled"
	CLResultReserved    = "Reserved"
	CLResultServerError = "ServerError"
)

//////////////////////////////////////////////
// Responses
//////////////////////////////////////////////

// Create/Check (Serv)Res
type CompLockV2Success struct {
	ID             string `json:"ID"`
	DeputyKey      string `json:"DeputyKey"`
	ReservationKey string `json:"ReservationKey,omitempty"`
	CreationTime   string `json:"CreationTime,omitempty"`
	ExpirationTime string `json:"ExpirationTime,omitempty"`
}
type CompLockV2Failure struct {
	ID     string `json:"ID"`
	Reason string `json:"Reason"`
}
type CompLockV2ReservationResult struct {
	Success []CompLockV2Success `json:"Success"`
	Failure []CompLockV2Failure `json:"Failure"`
}

// Renew/Release ServRes, Release/Remove Res, Create/Unlock/Repair/Disable locks
type CompLockV2Count struct {
	Total   int `json:"Total"`
	Success int `json:"Success"`
	Failure int `json:"Failure"`
}
type CompLockV2SuccessArray struct {
	ComponentIDs []string `json:"ComponentIDs"`
}
type CompLockV2UpdateResult struct {
	Counts  CompLockV2Count        `json:"Counts"`
	Success CompLockV2SuccessArray `json:"Success"`
	Failure []CompLockV2Failure    `json:"Failure"`
}

// Lock Status
type CompLockV2 struct {
	ID                  string `json:"ID"`
	Locked              bool   `json:"Locked"`
	Reserved            bool   `json:"Reserved"`
	CreationTime        string `json:"CreationTime,omitempty"`
	ExpirationTime      string `json:"ExpirationTime,omitempty"`
	ReservationDisabled bool   `json:"ReservationDisabled"`
}
type CompLockV2Status struct {
	Components []CompLockV2 `json:"Components"`
	NotFound   []string     `json:"NotFound"`
}

//////////////////////////////////////////////
// Payloads
//////////////////////////////////////////////

// Create/Remove Res, Create ServRes, Check/Lock/Unlock/Repair/Disable Lock
type CompLockV2Filter struct {
	ID                  []string `json:"ComponentIDs"`
	NID                 []string `json:"NID"`
	Type                []string `json:"Sype"`
	State               []string `json:"State"`
	Flag                []string `json:"Flag"`
	Enabled             []string `json:"Enabled"`
	SwStatus            []string `json:"SoftwareStatus"`
	Role                []string `json:"Role"`
	SubRole             []string `json:"Subrole"`
	Subtype             []string `json:"Subtype"`
	Arch                []string `json:"Arch"`
	Class               []string `json:"Class"`
	Group               []string `json:"Group"`
	Partition           []string `json:"Partition"`
	ProcessingModel     string   `json:"ProcessingModel"`
	ReservationDuration int      `json:"ReservationDuration"`
}

// Release Res, Release/Renew ServRes
type CompLockV2Key struct {
	ID  string `json:"ID"`
	Key string `json:"Key"`
}
type CompLockV2ReservationFilter struct {
	ReservationKeys     []CompLockV2Key `json:"ReservationKeys"`
	ProcessingModel     string          `json:"ProcessingModel"`
	ReservationDuration int             `json:"ReservationDuration"`
}

// Check ServRes
type CompLockV2DeputyKeyArray struct {
	DeputyKeys []CompLockV2Key `json:"DeputyKeys"`
}

func (cl *CompLockV2Filter) VerifyNormalize() error {
	cl.ProcessingModel = VerifyNormalizeProcessingModel(cl.ProcessingModel)
	if cl.ProcessingModel == "" {
		return ErrCompLockV2BadProcessingModel
	}
	if cl.ReservationDuration > 15 {
		return ErrCompLockV2BadDuration
	}
	return nil
}

func (clk *CompLockV2Key) VerifyNormalize() error {
	clk.ID = base.VerifyNormalizeCompID(clk.ID)
	if clk.ID == "" {
		return base.ErrHMSTypeInvalid
	}
	clk.Key = strings.ToLower(clk.Key)
	return nil
}

func (clr *CompLockV2ReservationFilter) VerifyNormalize() error {
	clr.ProcessingModel = VerifyNormalizeProcessingModel(clr.ProcessingModel)
	if clr.ProcessingModel == "" {
		return ErrCompLockV2BadProcessingModel
	}
	if clr.ReservationDuration > 15 {
		return ErrCompLockV2BadDuration
	}
	for i, key := range clr.ReservationKeys {
		err := key.VerifyNormalize()
		if err != nil {
			return err
		}
		clr.ReservationKeys[i] = key
	}
	return nil
}

func (cldk *CompLockV2DeputyKeyArray) VerifyNormalize() error {
	for i, key := range cldk.DeputyKeys {
		err := key.VerifyNormalize()
		if err != nil {
			return err
		}
		cldk.DeputyKeys[i] = key
	}
	return nil
}
