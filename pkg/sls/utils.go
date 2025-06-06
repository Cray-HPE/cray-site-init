/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
*/

package sls

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
)

// UtilsClient - Structure for SLS client.
type UtilsClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewSLSClient - Creates a new SLS client.
func NewSLSClient(baseURL string, httpClient *http.Client, token string) *UtilsClient {
	if httpClient == nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		httpClient = &http.Client{Transport: transport}
	}

	return &UtilsClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		token:      token,
	}
}

// GetManagementNCNs - Returns all the management NCNs from SLS.
func (utilsClient *UtilsClient) GetManagementNCNs() (managementNCNs []slsCommon.GenericHardware, err error) {
	url := fmt.Sprintf(
		"%s/v1/search/hardware?extra_properties.Role=Management",
		utilsClient.baseURL,
	)
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)

	// Indicates whether to close the connection after sending the request
	req.Close = true

	if err != nil {
		err = fmt.Errorf(
			"failed to create new request: %w",
			err,
		)
		return
	}
	if utilsClient.token != "" {
		req.Header.Add(
			"Authorization",
			fmt.Sprintf(
				"Bearer %s",
				utilsClient.token,
			),
		)
	}

	resp, err := utilsClient.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf(
			"failed to do request: %w",
			err,
		)
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(
		body,
		&managementNCNs,
	)
	if err != nil {
		err = fmt.Errorf(
			"failed to unmarshal body: %w",
			err,
		)
	}

	return
}

// GetNetworks - Returns all the networks from SLS.
func (utilsClient *UtilsClient) GetNetworks() (networks slsCommon.NetworkArray, err error) {
	url := fmt.Sprintf(
		"%s/v1/networks",
		utilsClient.baseURL,
	)
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)

	// Indicates whether to close the connection after sending the request
	req.Close = true

	if err != nil {
		err = fmt.Errorf(
			"failed to create new request: %w",
			err,
		)
		return
	}
	if utilsClient.token != "" {
		req.Header.Add(
			"Authorization",
			fmt.Sprintf(
				"Bearer %s",
				utilsClient.token,
			),
		)
	}

	resp, err := utilsClient.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf(
			"failed to do request: %w",
			err,
		)
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(
		body,
		&networks,
	)
	if err != nil {
		err = fmt.Errorf(
			"failed to unmarshal body: %w",
			err,
		)
	}

	return
}
