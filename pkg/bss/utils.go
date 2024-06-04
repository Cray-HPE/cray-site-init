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

package bss

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
)

// UtilsClient - Structure for BSS client.
type UtilsClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewBSSClient - Creates a new BSS client.
func NewBSSClient(baseURL string, httpClient *http.Client, token string) *UtilsClient {
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

// UploadEntryToBSS - Uploads an entry to BSS.
func (utilsClient *UtilsClient) UploadEntryToBSS(bssEntry bssTypes.BootParams, method string) (string, error) {
	url := fmt.Sprintf(
		"%s/boot/v1/bootparameters",
		utilsClient.baseURL,
	)

	jsonBytes, err := json.Marshal(bssEntry)
	if err != nil {
		return "", fmt.Errorf(
			"failed to marshal BSS entry: %w",
			err,
		)
	}

	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return "", fmt.Errorf(
			"failed to create new request: %w",
			err,
		)
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
	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := utilsClient.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf(
			"failed to %s BSS entry: %w",
			method,
			err,
		)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"failed to %s BSS entry: %s",
			method,
			string(bodyBytes),
		)
	}

	jsonPrettyBytes, _ := json.MarshalIndent(
		bssEntry,
		"",
		"  ",
	)

	return string(jsonPrettyBytes), nil
}

// GetBSSBootparametersForXname - Gets the BSS boot parameters for a given xname.
func (utilsClient *UtilsClient) GetBSSBootparametersForXname(xname string) (*bssTypes.BootParams, error) {
	url := fmt.Sprintf(
		"%s/boot/v1/bootparameters?name=%s",
		utilsClient.baseURL,
		xname,
	)

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create new request: %s",
			err,
		)
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
		return nil, fmt.Errorf(
			"failed to get BSS entry: %s",
			err,
		)
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to get BSS entry: %s",
			string(bodyBytes),
		)
	}

	// BSS gives back an array.
	var bssEntries []bssTypes.BootParams
	err = json.Unmarshal(
		bodyBytes,
		&bssEntries,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal BSS entries: %s",
			err,
		)
	}

	// We should only ever get one entry for a given xname.
	if len(bssEntries) != 1 {
		return nil, fmt.Errorf(
			"unexpected number of BSS entries: %+v",
			bssEntries,
		)
	}

	return &bssEntries[0], nil
}
