// MIT License
//
// (C) Copyright 2022,2025 Hewlett Packard Enterprise Development LP
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

package sls_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	base "github.com/Cray-HPE/hms-base/v2"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

type SLSClient struct {
	baseURL      string
	instanceName string
	client       *http.Client
	apiToken     string
}

func NewSLSClient(baseURL string, client *http.Client, instanceName string) *SLSClient {
	return &SLSClient{
		baseURL:      baseURL,
		client:       client,
		instanceName: instanceName,
	}
}

func (sc *SLSClient) WithAPIToken(apiToken string) *SLSClient {
	sc.apiToken = apiToken
	return sc
}

func (sc *SLSClient) addAPITokenHeader(request *http.Request) {
	if sc.apiToken != "" {
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sc.apiToken))
	}
}

func (sc *SLSClient) GetDumpState(ctx context.Context) (sls_common.SLSState, error) {
	// Build up the request
	request, err := http.NewRequestWithContext(ctx, "GET", sc.baseURL+"/v1/dumpstate", nil)
	if err != nil {
		return sls_common.SLSState{}, err
	}
	base.SetHTTPUserAgent(request, sc.instanceName)
	sc.addAPITokenHeader(request)

	// Perform the request!
	response, err := sc.client.Do(request)
	defer base.DrainAndCloseResponseBody(response)
	if err != nil {
		return sls_common.SLSState{}, err
	}

	if response.StatusCode != http.StatusOK {
		return sls_common.SLSState{}, fmt.Errorf("unexpected status code %d expected 200", response.StatusCode)
	}

	var dumpState sls_common.SLSState
	if err := json.NewDecoder(response.Body).Decode(&dumpState); err != nil {
		return sls_common.SLSState{}, err
	}

	return dumpState, nil
}

func (sc *SLSClient) GetAllHardware(ctx context.Context) ([]sls_common.GenericHardware, error) {
	// Build up the request
	request, err := http.NewRequestWithContext(ctx, "GET", sc.baseURL+"/v1/hardware", nil)
	if err != nil {
		return nil, err
	}
	base.SetHTTPUserAgent(request, sc.instanceName)
	sc.addAPITokenHeader(request)

	// Perform the request!
	response, err := sc.client.Do(request)
	defer base.DrainAndCloseResponseBody(response)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d expected 200", response.StatusCode)
	}

	var allHardware []sls_common.GenericHardware
	if err := json.NewDecoder(response.Body).Decode(&allHardware); err != nil {
		return nil, err
	}

	return allHardware, nil
}

func (sc *SLSClient) PutHardware(ctx context.Context, hardware sls_common.GenericHardware) error {
	if !xnametypes.IsHMSCompIDValid(hardware.Xname) {
		return fmt.Errorf("hardware has invalid xname %s", hardware.Xname)
	}

	rawRequestBody, err := json.Marshal(hardware)
	if err != nil {
		return fmt.Errorf("failed to marshal hardware object to json - %w", err)
	}

	// Build up the request!
	request, err := http.NewRequestWithContext(ctx, "PUT", sc.baseURL+"/v1/hardware/"+hardware.Xname, bytes.NewBuffer(rawRequestBody))
	if err != nil {
		return err
	}
	base.SetHTTPUserAgent(request, sc.instanceName)
	sc.addAPITokenHeader(request)

	// Perform the request!
	response, err := sc.client.Do(request)
	defer base.DrainAndCloseResponseBody(response)
	if err != nil {
		return err
	}

	// PUT can either create or update objects.
	if !(response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated) {
		return fmt.Errorf("unexpected status code %d expected 200 or 201", response.StatusCode)
	}

	return nil
}

func (sc *SLSClient) PutNetwork(ctx context.Context, network sls_common.Network) error {
	if len(network.Name) == 0 {
		return fmt.Errorf("network has empty network name")
	}

	if strings.Contains(network.Name, " ") {
		return fmt.Errorf("network name contains spaces (%s)", network.Name)
	}

	rawRequestBody, err := json.Marshal(network)
	if err != nil {
		return fmt.Errorf("failed to marshal hardware object to json - %w", err)
	}

	// Build up the request!
	request, err := http.NewRequestWithContext(ctx, "PUT", sc.baseURL+"/v1/networks/"+network.Name, bytes.NewBuffer(rawRequestBody))
	if err != nil {
		return err
	}
	base.SetHTTPUserAgent(request, sc.instanceName)
	sc.addAPITokenHeader(request)

	// Perform the request!
	response, err := sc.client.Do(request)
	defer base.DrainAndCloseResponseBody(response)
	if err != nil {
		return err
	}

	// PUT can either create or update objects.
	if !(response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated) {
		return fmt.Errorf("unexpected status code %d expected 200 or 201", response.StatusCode)
	}

	return nil
}

func (sc *SLSClient) GetNetworks(ctx context.Context) ([]sls_common.Network, error) {
	// Build up the request
	request, err := http.NewRequestWithContext(ctx, "GET", sc.baseURL+"/v1/networks", nil)
	if err != nil {
		return nil, err
	}
	base.SetHTTPUserAgent(request, sc.instanceName)
	sc.addAPITokenHeader(request)

	// Perform the request!
	response, err := sc.client.Do(request)
	defer base.DrainAndCloseResponseBody(response)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d expected 200", response.StatusCode)
	}

	var networks []sls_common.Network
	if err := json.NewDecoder(response.Body).Decode(&networks); err != nil {
		return nil, err
	}

	return networks, nil
}

func (sc *SLSClient) GetNetwork(ctx context.Context, networkName string) (sls_common.Network, error) {
	// Build up the request
	request, err := http.NewRequestWithContext(ctx, "GET", sc.baseURL+"/v1/networks/"+networkName, nil)
	if err != nil {
		return sls_common.Network{}, err
	}
	base.SetHTTPUserAgent(request, sc.instanceName)
	sc.addAPITokenHeader(request)

	// Perform the request!
	response, err := sc.client.Do(request)
	defer base.DrainAndCloseResponseBody(response)
	if err != nil {
		return sls_common.Network{}, err
	}

	if response.StatusCode != http.StatusOK {
		return sls_common.Network{}, fmt.Errorf("unexpected status code %d expected 200", response.StatusCode)
	}

	var network sls_common.Network
	if err := json.NewDecoder(response.Body).Decode(&network); err != nil {
		return sls_common.Network{}, err
	}

	return network, nil
}
