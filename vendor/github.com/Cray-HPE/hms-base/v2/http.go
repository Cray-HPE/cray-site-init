// MIT License
//
// (C) Copyright [2019-2021,2025] Hewlett Packard Enterprise Development LP
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
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Package to slightly abstract some of the most mundane of HTTP interactions. Primary intention is as a JSON
// getter and parser, with the latter being a generic interface that can be converted to a custom structure.
type HTTPRequest struct {
	Context            context.Context // Context to pass to the underlying HTTP client.
	FullURL            string          // The full URL to pass to the HTTP client.
	Method             string          // HTTP method to use.
	Payload            []byte          // Bytes payload to pass if desired of ContentType.
	Auth               *Auth           // Basic authentication if necessary using Auth struct.
	Timeout            time.Duration   // Timeout for entire transaction.
	SkipTLSVerify      bool            // Ignore TLS verification errors?
	ExpectedStatusCode int             // Expected HTTP status return code.
	ContentType        string          // HTTP content type of Payload.
}

// These are used to reduce duplication when adding User-Agent headers to requests.

const USERAGENT = "User-Agent"

func GetServiceInstanceName() (string, error) {
	return os.Hostname()
}

func SetHTTPUserAgent(req *http.Request, instName string) {
	if req == nil {
		return
	}

	//See if this User Agent is already in place

	found := false
	_, ok := req.Header[USERAGENT]

	if ok {
		for _, v := range req.Header[USERAGENT] {
			if v == instName {
				found = true
				break
			}
		}
	}

	if !found {
		req.Header.Add(USERAGENT, instName)
	}
}

// NewHTTPRequest creates a new HTTPRequest with default settings.
func NewHTTPRequest(fullURL string) *HTTPRequest {
	return &HTTPRequest{
		Context:            context.Background(),
		FullURL:            fullURL,
		Method:             "GET",
		Payload:            nil,
		Auth:               nil,
		Timeout:            30 * time.Second,
		SkipTLSVerify:      false,
		ExpectedStatusCode: http.StatusOK,
		ContentType:        "application/json",
	}
}

func (request *HTTPRequest) String() string {
	return fmt.Sprintf(
		"Context: %s, "+
			"Method: %s, "+
			"Full URL: %s, "+
			"Payload: %s, "+
			"Auth: (%s), "+
			"Timeout: %d, "+
			"SkipTLSVerify: %t, "+
			"ExpectedStatusCode: %d, "+
			"ContentType: %s",
		request.Context,
		request.Method,
		request.FullURL,
		string(request.Payload),
		request.Auth,
		request.Timeout,
		request.SkipTLSVerify,
		request.ExpectedStatusCode,
		request.ContentType)
}

// HTTP basic authentication structure.
type Auth struct {
	Username string
	Password string
}

// Custom String function to prevent passwords from being printed directly (accidentally) to output.
func (auth Auth) String() string {
	return fmt.Sprintf("Username: %s, Password: <REDACTED>", auth.Username)
}

// Functions

// Given a HTTPRequest this function will facilitate the desired operation using the retryablehttp package to gracefully
// retry should the connection fail.
func (request *HTTPRequest) DoHTTPAction() (payloadBytes []byte, err error) {
	// Sanity check
	if request.FullURL == "" {
		err = fmt.Errorf("URL can not be empty")
		return
	}

	// Setup the common HTTP request stuff.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: request.SkipTLSVerify},
	}
	client := retryablehttp.NewClient()
	client.HTTPClient.Timeout = request.Timeout
	client.HTTPClient.Transport = transport

	var req *retryablehttp.Request

	// If there's a payload, make sure to include it.
	if request.Payload == nil {
		req, _ = retryablehttp.NewRequest(request.Method, request.FullURL, nil)
	} else {
		req, _ = retryablehttp.NewRequest(request.Method, request.FullURL, bytes.NewBuffer(request.Payload))
	}

	// Set the context to the same we were given on the way in.
	req = req.WithContext(request.Context)

	req.Header.Set("Content-Type", request.ContentType)

	if request.Auth != nil {
		req.SetBasicAuth(request.Auth.Username, request.Auth.Password)
	}

	resp, doErr := client.Do(req)
	defer DrainAndCloseResponseBody(resp)
	if doErr != nil {
		err = fmt.Errorf("unable to do request: %s", doErr)
		return
	}

	// Make sure we get the status code we expect.
	if resp.StatusCode != request.ExpectedStatusCode {
		err = fmt.Errorf("received unexpected status code: %d", resp.StatusCode)
		return
	}

	// Get the payload.
	payloadBytes, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		err = fmt.Errorf("unable to read response body: %s", readErr)
		return
	}

	return
}

// Returns an interface for the response body for a given request by calling DoHTTPAction and unmarshaling.
// As such, do NOT call this method unless you expect a JSON body in return!
//
// A powerful way to use this function is by feeding its result to the mapstructure package's Decode method:
// 		v := request.GetBodyForHTTPRequest()
// 		myTypeInterface := v.(map[string]interface{})
// 		var myPopulatedStruct MyType
// 		mapstructure.Decode(myTypeInterface, &myPopulatedStruct)
// In this way you can generically make all your HTTP requests and essentially "cast" the resulting interface to a
// structure of your choosing using it as normal after that point. Just make sure to infer the correct type for `v`.
func (request *HTTPRequest) GetBodyForHTTPRequest() (v interface{}, err error) {
	payloadBytes, err := request.DoHTTPAction()
	if err != nil {
		return
	}

	stringPayloadBytes := string(payloadBytes)
	if stringPayloadBytes != "" {
		// If we've made it to here we have all we need, unmarshal.
		jsonErr := json.Unmarshal(payloadBytes, &v)
		if jsonErr != nil {
			err = fmt.Errorf("unable to unmarshal payload: %s", jsonErr)
			return
		}
	}

	return
}

// Response bodies should always be drained and closed, else we leak resources
// and fail to reuse network connections.

func DrainAndCloseResponseBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body) // ok even if already drained
			resp.Body.Close()                     // ok even if already closed
	}
}

// While it is generally not a requirement to close request bodies in server
// handlers, it is good practice.  If a body is only partially read, there can
// be a resource leak.  Additionally, if the body is not read at all, the
// network connection will be closed and will not be reused even though the
// http server will properly drain and close the request body.

func DrainAndCloseRequestBody(req *http.Request) {
	if req != nil && req.Body != nil {
			_, _ = io.Copy(io.Discard, req.Body) // ok even if already drained
			req.Body.Close()                     // ok even if already closed
	}
}
