package sls

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/mitchellh/mapstructure"
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

func GetIPReservation(networks sls_common.NetworkArray, networkName, subnetName, ipReservationName string) *IPReservation {
	// Search SLS networks for this network.
	var targetSLSNetwork *sls_common.Network
	for _, slsNetwork := range networks {
		if strings.ToLower(slsNetwork.Name) == strings.ToLower(networkName) {
			targetSLSNetwork = &slsNetwork
			break
		}
	}

	if targetSLSNetwork == nil {
		log.Fatalf("Failed to find required IPAM network %s in SLS networks!", networkName)
	}

	// Map this network to a usable structure.
	var networkExtraProperties NetworkExtraProperties
	err := mapstructure.Decode(targetSLSNetwork.ExtraPropertiesRaw, &networkExtraProperties)
	if err != nil {
		log.Fatalf("Failed to decode raw network extra properties to correct structure: %s", err)
	}

	// Find the subnet of intrest within SLS network
	var targetSubnet *IPV4Subnet
	for _, subnet := range networkExtraProperties.Subnets {
		if strings.ToLower(subnet.Name) == strings.ToLower(subnetName) {
			targetSubnet = &subnet
			break
		}
	}

	// Find the IP Reservation within the subnet
	var targetReservation *IPReservation
	for _, reservation := range targetSubnet.IPReservations {
		// Yeah, this is as strange as it looks...convention is to put the xname in the comment
		// field. ¯\_(ツ)_/¯
		if reservation.Name == ipReservationName {
			targetReservation = &reservation
			break
		}
	}

	if targetSubnet == nil || targetReservation == nil {
		log.Fatalf("Failed to find subnet/reservation (%s) in subnet (%s) in the SLS Network (%s)!",
			networkName, subnetName, networkName)
	}

	return targetReservation
}