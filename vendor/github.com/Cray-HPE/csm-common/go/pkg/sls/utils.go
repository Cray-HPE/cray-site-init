package sls

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
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
func (utilsClient *UtilsClient) GetManagementNCNs() (managementNCNs []sls_common.GenericHardware, err error) {
	url := fmt.Sprintf("%s/v1/search/hardware?extra_properties.Role=Management",
		utilsClient.baseURL)
	req, err := http.NewRequest("GET", url, nil)

	// Indicates whether to close the connection after sending the request
	req.Close = true

	if err != nil {
		err = fmt.Errorf("failed to create new request: %w", err)
		return
	}
	if utilsClient.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", utilsClient.token))
	}

	resp, err := utilsClient.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to do request: %w", err)
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &managementNCNs)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal body: %w", err)
	}

	return
}

// GetNetworks - Returns all the networks from SLS.
func (utilsClient *UtilsClient) GetNetworks() (networks sls_common.NetworkArray, err error) {
	url := fmt.Sprintf("%s/v1/networks", utilsClient.baseURL)
	req, err := http.NewRequest("GET", url, nil)

	// Indicates whether to close the connection after sending the request
	req.Close = true

	if err != nil {
		err = fmt.Errorf("failed to create new request: %w", err)
		return
	}
	if utilsClient.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", utilsClient.token))
	}

	resp, err := utilsClient.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to do request: %w", err)
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &networks)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal body: %w", err)
	}

	return
}
