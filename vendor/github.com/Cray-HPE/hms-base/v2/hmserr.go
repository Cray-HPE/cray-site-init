// MIT License
//
// (C) Copyright [2018, 2021] Hewlett Packard Enterprise Development LP
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
	"fmt"
	"net/http"
	"strings"
)

////////////////////////////////////////////////////////////////////////////
//
// RFC 7807-compliant Problem Details struct
//
////////////////////////////////////////////////////////////////////////////

const ProblemDetailsHTTPStatusType = "about:blank"
const ProblemDetailContentType = "application/problem+json"

// RFC 7807 Problem Details
//
// These are the officially-specified fields.  The implementation
// is allowed to add new ones, but we'll stick with these for now.
// Almost all are optional, however, and blank fields are not encoded
// if they are the empty string (which is fine
//
// The only required field is Type and is expected to be a URL that
// should describe the problem type when dereferenced.  It's not intended
// to be a dead link, but I don't think clients are allowed to assume it's
// not, either.  We don't implement a way of serving these URLs here,
// in any case. It's not obligatory to have a lot of different problem
// types, or even desirable if the client can just andling them all the
// same basic way.
//
// The only non-URL that is allowed for Type is "about:blank", but only in
// the situation where the semantics of the problem can be understood okay
// as a basic HTTP error.  In this case Title should be the official HTTP
// status code string, not something custom.  You are free to put anything
// you want in Detail however, so if you already have an error function
// that returns an error string, you can easily convert it into one of these
// generic HTTP errors.  In fact, we offer a shortcut to doing this.
//
// If Type is a URL, Title should describe the problem type.  If you are
// doing this, and not just using type:"about:blank" and the HTTP error,
// for the title, it should be because you have some kind of problem that
// needs, or could benefit from, special or more involved treatment that
// the URL describes when dereferenced.
//
// Detail explains in a human readible way what happened when a specific
// problem occurred.
//
// Instance is another URI describing a specific occurrence of a problem. So
// within a particular Type and Title (a general problem type) the Detail
// and Instance document a specific incident in which that problem occurred.
// Or at least that's the general idea.
//
// Status is the numerical HTTP status code.  It is optional, and if present,
// should match at least the intended HTTP code in the header, though this
// is not strictly required.
//
// For more info, reading RFC 7807 would obviously be the authoritative source.
type ProblemDetails struct {
	Type     string `json:"type"` // either url or "about:blank"
	Title    string `json:"title,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Status   int    `json:"status,omitempty"`
}

// New full ProblemDetails, will all fields specified.
// Type is the only required field, and if any of the other args are left
// as the empty string, the fields won't appear in the JSON output at all.
//
//  p := base.NewProblemDetails(
//      "https://example.com/probs/MyProblem",
//      "My Problem's Title",
//      "Details for this problem",
//      "/instances/myprob/1",
//      http.StatusBadRequest,
//  )
func NewProblemDetails(ptype, title, detail, instance string, status int) *ProblemDetails {
	p := new(ProblemDetails)
	p.Type = ptype
	p.Title = title
	p.Detail = detail
	p.Instance = instance
	p.Status = status

	return p
}

// New generic ProblemDetails for errors that are just based on the
// HTTP status code.   Type and Title are filled in based on the
// status code.
//
// There is no need for a URL in this case, as the Type is allowed to be
// "about:blank" if title is just the Status code text and there is no special
// handling needed beside the usual for that HTTP (error) StatusCode.
//
// NOTE: Status should only be filled in using http.Status* instead of a
// literal number.
//
// 'status' will be treated as http.StatusBadRequest (400) if it does
// not match a valid http status code.
//
// Example:
//  p := base.NewProblemDetailsStatus(http.StatusNotFound, "No such component")
// Produces a struct p with the following fields
//  &ProblemDetails{
//     Type: "about:blank",
//     Title: "Not Found",
//     Detail: "No such component",
//     Status: 404,
//  }
func NewProblemDetailsStatus(detail string, status int) *ProblemDetails {
	p := new(ProblemDetails)
	p.Type = ProblemDetailsHTTPStatusType
	p.Title = http.StatusText(status)
	// Got bad status type, default to 400
	if p.Title == "" {
		p.Status = http.StatusBadRequest
		p.Title = http.StatusText(p.Status)
	} else {
		p.Status = status
	}
	p.Detail = detail

	return p
}

// Create a new ProblemDetails struct copied from p with only detail and
// instance (optionally) updated.  If either field is the empty string, it
// will not be updated and the parent values will be used.
//
// The basic idea here is to be able to define a few prototype ProblemDetails
// with the Type and Title filled in (along with a default Detail and HTTP
// Status code, if desired), and then create a child copy when a problem
// actually occurs with the specific Detail and (optionally) instance
// URI filled in.
//
// Example:
//  p := base.NewProblemDetails(
//      "https://example.com/probs/MyProblem",
//      "My Problem's Title",
//      "Generic details for this problem type",
//      "/instances/myprobs/",
//      http.StatusBadRequest,
//  )
//
//  // Copy updates Detail and instance
//  pChild := p.NewChild("Specific details for this problem", "/instances/myprobs/1")
//
//  // Copy has updated Detail field only
//  // i.e. p.Instance == pChildDetailOnly.Instance
//  pChildDetailOnly := p.NewChild("Specific details for this problem", "")
//
//  // Strict copy only, new ProblemDetails has identical fields.
//  pCopy := p.NewChild("", "")
func (p *ProblemDetails) NewChild(detail, instance string) *ProblemDetails {
	newProb := new(ProblemDetails)
	newProb.Type = p.Type
	newProb.Title = p.Title
	newProb.Status = p.Status
	if detail != "" {
		newProb.Detail = detail
	} else {
		newProb.Detail = p.Detail
	}
	if instance != "" {
		newProb.Instance = instance
	} else {
		newProb.Instance = p.Instance
	}

	return newProb
}

///////////////////////////////////////////////////////////////////////////
// http.ResponseWriter response formatting for ProblemDetails
///////////////////////////////////////////////////////////////////////////

// Send specially-coded RFC7807 Problem Details response.  A set ProblemDetails
// struct is expected ('p') and should always contain at least Type to conform
// to the spec (we will always print the Type field, but it would contain the
// empty string if unset which isn't technically valid for the spec.
//
// The HTTP status code in the header is set to the 'status' arg if it is
// NON-zero.  If 'status' = 0, then 'p'.Status will be used.  If that is
// unset (i.e. 0) also, we will use a default error code of 400.
//
// For example:
//  p := base.NewProblemDetails(
//      "https://example.com/probs/MyProblem",
//      "My Problem's Title",
//      "Detail for this problem",
//      "/instances/myprob/1",
//      http.StatusConflict)  // 409
//  ...
//  if err != nil {
//     base.SendProblemDetails(w, p, 0)
//     return
//  }
//
// ...will include the following in the response header:
//  HTTP/1.1 409 Conflict
//  Content-Type: application/problem+json
// With the following body:
//  {
//     "Type": "https://example.com/probs/MyProblem",
//     "Title": "My Problem's Title",
//     "Detail": "Detail for this problem",
//     "Instance": "/instances/myprob/1",
//     "Status": 409
//  }
func SendProblemDetails(w http.ResponseWriter, p *ProblemDetails, status int) error {
	// Set proper content type for RFC 7807 Problem details
	w.Header().Set("Content-Type", ProblemDetailContentType)
	// Variable 'status' is preferred over p.Status, which might not be set,
	// being an optional field.
	realStatus := status
	if status == 0 {
		realStatus = p.Status
	}
	// We shouldn't be returning an actual 0.  This is the fall back.
	if realStatus == 0 {
		realStatus = http.StatusBadRequest
	}
	w.WriteHeader(realStatus)
	err := json.NewEncoder(w).Encode(p)
	if err != nil {
		retErr := fmt.Errorf("couldn't encode a JSON problem response: %s\n", err)
		return retErr
	}
	return nil
}

// Generate and then send a generic RFC 7807 problem response based on a
// HTTP status code ('status')  and a message string ('msg').  This
// generates a generic title and type based on the HTTP 'status' code.  The
// ProblemDetails "Detail" field is filled in using 'msg'.
//
// If you are already returning just a simple error message and want an
// easy way to convert this into a basic (but valid) RFC7807 ProblemDetails,
// this is almost a drop-in replacement.
//
//  // If request was for a component that is not in DB, send RFC 7807
//  // ProblemDetails and exit HTTP handler
//  err := GetComponentFromDB(component_from_query)
//  if err != nil {
//     base.SendProblemDetailsGeneric(w, http.StatusNotFound, "No such component")
//     return
//  }
//
// The above will include the following in the response header:
//  HTTP/1.1 404 Not Found
//  Content-Type: application/problem+json
// With the following body:
//  {
//     "Type": "about:blank",
//     "Title": "Not Found",
//     "Detail": "No such component",
//     "Status": 404
//  }
func SendProblemDetailsGeneric(w http.ResponseWriter, status int, msg string) error {
	problem := NewProblemDetailsStatus(msg, status)
	err := SendProblemDetails(w, problem, problem.Status)
	return err
}

////////////////////////////////////////////////////////////////////////////
//
//
// HMSError - Custom Go 'error' type for HMS.
//
//
////////////////////////////////////////////////////////////////////////////

const HMSErrorUnsetDefault = "no error message or class set"

// Custom Go 'error' type for HMS - Implements Error()
//
// Works like a standard Go 'error', but can be distinguished
// as an HMS-specific error with an optional class for better
// handling upstream.  An RFC7807 error can optionally be added
// in case we need to pass those through multiple layers- but
// without forcing us to (since they still look like regular
// Go errors).  Note that this is just a starting point.  We can
// attach other types of info later if we want.
//
// Part of the motivation is so we can determine which errors
// are safe to return to users, i.e. we don't want to give
// them things that expose details about the database structure.
//
// Can attach custom RFC7807 response if needed.
type HMSError struct {
	// Used to identify similar groups of errors to higher-level software.
	Class string `json:"class"`

	// Message, used as error string for Error()
	Message string `json:"message"`

	// Optional full ProblemDetails
	Problem *ProblemDetails `json:"problem,omitempty"`
}

// New HMSError, with message string and optional class.
//
// A subsequent call to AddProblem() can associate a ProblemDetails with it.
// By default Problem pointer is nil (again, this is optional functionality)
func NewHMSError(class, msg string) *HMSError {
	newErr := new(HMSError)
	newErr.Class = class
	newErr.Message = msg

	return newErr
}

// Implement Error() interface - This makes it so an HMSError can be
// returned anywhere an ordinary Go 'error' is returned.
//
// The receiver of an 'error' can then use one of the IsHMSError* or HasHMSError
// functions to test if it is an HMSError, and if so, to obtain the additional
// fields.
//
func (e *HMSError) Error() string {
	if e.Message != "" {
		return e.Message
	} else if e.Class != "" {
		return e.Class
	} else {
		return HMSErrorUnsetDefault
	}
}

// See if an error (some thing implementing Error() interface is an HMSError
//
// Returns 'true' if err is HMSError
// Returns 'false' if err is something else that implements Error()
func IsHMSError(err error) bool {
	_, ok := err.(*HMSError)
	if !ok {
		return false
	} else {
		return true
	}
}

// Test and retrieve HMSError info, if error is in fact of that type.
//
// If bool is 'true', HMSError will be a non-nil HMSError with the expected
// extended HMSError fields.
//
// If bool is 'false', err is not an HMSError and *HMSError will be nil
func GetHMSError(err error) (*HMSError, bool) {
	hmserr, ok := err.(*HMSError)
	if !ok {
		return nil, false
	} else {
		return hmserr, true
	}
}

// Return true if 'class' exactly matches Class field for HMSError
func (e *HMSError) IsClass(class string) bool {
	if class == e.Class {
		return true
	}
	return false
}

// Return true if 'class' matches Class field for HMSError (case insensitive)
func (e *HMSError) IsClassIgnCase(class string) bool {
	if strings.ToLower(class) == strings.ToLower(e.Class) {
		return true
	}
	return false
}

// Returns false if 'err' is not an HMSError, or, if it is, if 'class' doesn't
// match the HMSError's Class field.
func IsHMSErrorClass(err error, class string) bool {
	hmserr, ok := GetHMSError(err)
	if !ok {
		return false
	}
	return hmserr.IsClass(class)
}

// Returns false if 'err' is not an HMSError, or, if it is, if 'class' doesn't
// match the HMSError's Class field (case insensitive).
func IsHMSErrorClassIgnCase(err error, class string) bool {
	hmserr, ok := GetHMSError(err)
	if !ok {
		return false
	}
	return hmserr.IsClassIgnCase(class)
}

// Add an ProblemDetails to be associated with e
//
// Note: This simply associates the pointer to p with e.  There is no
// new ProblemDetails created and no deep copy is performed.
func (e *HMSError) AddProblem(p *ProblemDetails) {
	e.Problem = p
}

// If e has a ProblemDetails associated with it, return a pointer to it
// If not, return nil.
func (e *HMSError) GetProblem() *ProblemDetails {
	return e.Problem
}

// Create a new HMSError that is a copy of e, except with an updated "Message"
// field if 'msg' is not the empty string.  Conversely, if 'msg' IS
// the empty string, i.e. "", the operation is equivalent to a copy.
//
// Note: This DOES NOT COPY ProblemDetails if added with AddProblem.
// Instead, use NewChildWithProblem() for cases like this (or if you're
// unsure).
//
// The (newerr).Problem pointer is always set to nil with this function.
func (e *HMSError) NewChild(msg string) *HMSError {
	newErr := new(HMSError)
	newErr.Class = e.Class
	if msg != "" {
		newErr.Message = msg
	} else {
		newErr.Message = e.Message
	}
	return newErr
}

// Create a newly-allocated child HMSError, deep-copying the parent, including
// any ProblemDetails that (may) have been associated using AddProblem().
//
// In addition to copying, the 'msg' arg (if non-empty) is used to overwrite
// HMSError's "Message" and (if ProblemDetails are present) the ProblemDetails
// "Detail" field is also set to the same string.
//
// The ProblemDetails 'Instance' is also updated if the 'instance' arg is a
// non-empty string.
//
// If both 'msg' and 'instance' are unset the result is effectively a simple
// (deep) copy of e (including a deep copy of e.Problem).
//
// This method basically combines the functionality of both the HMSError
// and ProblemDetails NewChild() functions.
func (e *HMSError) NewChildWithProblem(msg, instance string) *HMSError {
	newErr := new(HMSError)
	newErr.Class = e.Class
	if msg != "" {
		newErr.Message = msg
	} else {
		newErr.Message = e.Message
	}
	if e.Problem != nil {
		newErr.Problem = e.Problem.NewChild(msg, instance)
	}
	return newErr
}
