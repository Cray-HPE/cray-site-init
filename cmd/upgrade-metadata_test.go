/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

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

package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/Cray-HPE/cray-site-init/pkg/bss"
	"github.com/Cray-HPE/cray-site-init/pkg/sls"
	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/suite"
)

type UpgradeBSSMetadataSuite struct {
	suite.Suite

	slsNetworks sls_common.NetworkArray

	expectedGlobalBootparameters bssTypes.BootParams
	expectedGlobalHostRecords    bss.HostRecords
}

func (suite *UpgradeBSSMetadataSuite) getHostRecords(globalBootParameters bssTypes.BootParams) bss.HostRecords {
	var globalHostRecords bss.HostRecords
	err := mapstructure.Decode(globalBootParameters.CloudInit.MetaData["host_records"], &globalHostRecords)
	suite.NoError(err)

	return globalHostRecords
}

func (suite *UpgradeBSSMetadataSuite) sortHostRecords(hostRecords []bss.HostRecord) {
	sort.SliceStable(hostRecords, func(i, j int) bool {
		return hostRecords[i].IP < hostRecords[j].IP
	})
}

func (suite *UpgradeBSSMetadataSuite) SetupTest() {
	// Load in management NCNs from SLS, into the global variable managementNCNs
	ncnsRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/sls_management_ncns.json")
	suite.NoError(err)
	err = json.Unmarshal(ncnsRaw, &managementNCNs)
	suite.NoError(err)
	suite.NotEmpty(managementNCNs)

}

func (suite *UpgradeBSSMetadataSuite) TestUpdateBSS_oneToOneTwo_CANOnly() {
	//
	// Load in BSS test data
	//
	allBootParametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/csm1.0_bss_bootparameters.json")
	suite.NoError(err)

	var allBootParametersArray []bssTypes.BootParams
	err = json.Unmarshal(allBootParametersRaw, &allBootParametersArray)
	suite.NoError(err)

	allBootParameters := map[string]bssTypes.BootParams{}
	for _, bootParameters := range allBootParametersArray {
		if len(bootParameters.Hosts) != 1 {
			continue
		}
		allBootParameters[bootParameters.Hosts[0]] = bootParameters
	}

	// Verify the CSM 1.0 Global boot parameters contain the can-if and can-gw keys
	suite.Contains(allBootParameters["Global"].CloudInit.MetaData, "can-if")
	suite.Contains(allBootParameters["Global"].CloudInit.MetaData, "can-gw")

	// The SLS Network data is expected to be already configured for CSM 1.2
	networksRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/can-only/csm1.2_sls_networks.json")
	suite.NoError(err)
	err = json.Unmarshal(networksRaw, &suite.slsNetworks)
	suite.NoError(err)
	suite.NotEmpty(suite.slsNetworks)

	// Load in the expected global BSS boot parameters for CSM 1.2
	expectedGlobalBootparametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/can-only/csm1.2_expected_global_bootparameters.json")
	suite.NoError(err)
	err = json.Unmarshal(expectedGlobalBootparametersRaw, &suite.expectedGlobalBootparameters)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalBootparameters)

	// Extract the host_records from the global BSS boot parameters
	suite.expectedGlobalHostRecords = bss.HostRecords{}
	err = mapstructure.Decode(suite.expectedGlobalBootparameters.CloudInit.MetaData["host_records"], &suite.expectedGlobalHostRecords)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalHostRecords)

	//
	// Setup test HTTP servers for BSS and SLS
	//
	bssTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.T().Logf("Received BSS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/boot/v1/bootparameters" {
			host := r.URL.Query().Get("name")
			bootParameters, ok := allBootParameters[host]
			if !ok {
				suite.T().Logf("Unable to find host %s in BSS Bootparameters", host)
				suite.T().Fail()
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]bssTypes.BootParams{bootParameters})
			return
		} else if r.Method == http.MethodPatch && r.URL.Path == "/boot/v1/bootparameters" {
			var bootParameters bssTypes.BootParams
			err = json.NewDecoder(r.Body).Decode(&bootParameters)
			suite.NoError(err)

			if len(bootParameters.Hosts) != 1 {
				// For this test case we expect for 1 host to be set.
				suite.T().Logf("Received unexpected BSS Put request with %d hosts specified", len(bootParameters.Hosts))
				suite.T().Fail()
				w.WriteHeader(http.StatusBadRequest)
			}

			host := bootParameters.Hosts[0]
			if _, present := allBootParameters[host]; !present {
				suite.T().Logf("Received unexpected BSS Put request for nonexistent host %s", host)
				suite.T().Fail()
			}

			allBootParameters[host] = bootParameters
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bssTS.Close()

	slsTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.T().Logf("Received SLS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/v1/networks" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(suite.slsNetworks)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer slsTS.Close()

	// Configure BSS and SLS clients to use our test HTTP Servers
	bssClient = bss.NewBSSClient(bssTS.URL, bssTS.Client(), "")
	slsClient = sls.NewSLSClient(slsTS.URL, slsTS.Client(), "")

	// Function under test
	updateBSS10to12()

	// The global cloud-init meta-data should no longer contain the keys can-if and can-gw
	globalBootParameters := allBootParameters["Global"]
	suite.NotContains(globalBootParameters.CloudInit.MetaData, "can-if")
	suite.NotContains(globalBootParameters.CloudInit.MetaData, "can-gw")

	// Verify Global cloud-init data has updated host_records entries
	// When comparing the host_records the ordering could be different, but are functionally equivalent.
	globalHostRecords := suite.getHostRecords(globalBootParameters)
	suite.sortHostRecords(globalHostRecords)
	suite.sortHostRecords(suite.expectedGlobalHostRecords)
	suite.Equal(suite.expectedGlobalHostRecords, globalHostRecords)

	// Verify IPAM data for a single node:
	worker3BootParameters := allBootParameters["x3000c0s11b0n0"]
	var worker3IPAM bss.CloudInitIPAM
	err = mapstructure.Decode(worker3BootParameters.CloudInit.MetaData["ipam"], &worker3IPAM)
	suite.NoError(err)

	suite.Contains(worker3IPAM, "can")
	suite.Contains(worker3IPAM, "cmn")
	suite.Contains(worker3IPAM, "hmn")
	suite.Contains(worker3IPAM, "mtl")
	suite.Contains(worker3IPAM, "nmn")

	suite.Equal(bss.IPAMNetwork{Gateway: "10.102.4.129", CIDR: "10.102.4.141/26", ParentDevice: "bond0", VlanID: 6}, worker3IPAM["can"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.102.4.1", CIDR: "10.102.4.12/25", ParentDevice: "bond0", VlanID: 7}, worker3IPAM["cmn"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.254.0.1", CIDR: "10.254.1.20/17", ParentDevice: "bond0", VlanID: 4}, worker3IPAM["hmn"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.1.0.1", CIDR: "10.1.1.10/16", ParentDevice: "bond0", VlanID: 0}, worker3IPAM["mtl"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.252.0.1", CIDR: "10.252.1.12/17", ParentDevice: "bond0", VlanID: 2}, worker3IPAM["nmn"])

	// The CHN should not be present in the BSS IPAM structure, as that network is managed differently
	suite.NotContains(worker3IPAM, "chn")

	// Kernel parameters are the same regardless of CAN/CHN so they will only be tested in this test
	// Verify kernel params for worker node
	worker3Params := allBootParameters["x3000c0s11b0n0"].Params
	suite.Contains(worker3Params, "ifname=mgmt0:50:6b:4b:08:d0:4a")
	suite.Contains(worker3Params, "ifname=mgmt1:50:6b:4b:08:d0:4b")
	suite.Contains(worker3Params, "ifname=hsn")
	suite.NotContains(worker3Params, "ifname=sun")
	suite.NotContains(worker3Params, "bond0")
	suite.Contains(worker3Params, "rd.net.dhcp.retry=5")
	suite.Contains(worker3Params, "rd.peerdns=0")
	suite.Contains(worker3Params, "ip=mgmt0:dhcp")
	suite.NotContains(worker3Params, "auto6")
	suite.NotContains(worker3Params, "vlan")
	suite.NotContains(worker3Params, "bootdev")
	suite.NotContains(worker3Params, "hwprobe")

	// Verify kernel params for master node with 1 NIC
	master2Params := allBootParameters["x3000c0s3b0n0"].Params
	suite.Contains(master2Params, "ifname=mgmt0:b8:59:9f:2b:31:02")
	suite.Contains(master2Params, "ifname=mgmt1:b8:59:9f:2b:31:03")
	suite.NotContains(master2Params, "ifname=sun")
	suite.NotContains(master2Params, "bond0")
	suite.Contains(master2Params, "rd.net.dhcp.retry=5")
	suite.Contains(master2Params, "rd.peerdns=0")
	suite.Contains(master2Params, "ip=mgmt0:dhcp")
	suite.NotContains(master2Params, "auto6")
	suite.NotContains(master2Params, "vlan")
	suite.NotContains(master2Params, "bootdev")
	suite.NotContains(master2Params, "hwprobe")

	// Verify kernel params for master node with 2 NICs
	master3Params := allBootParameters["x3000c0s5b0n0"].Params
	suite.Contains(master3Params, "ifname=mgmt0:14:02:ec:d5:fa:38")
	suite.Contains(master3Params, "ifname=sun0:14:02:ec:d5:fa:39")
	suite.Contains(master3Params, "ifname=mgmt1:94:40:c9:5c:86:86")
	suite.Contains(master3Params, "ifname=sun1:94:40:c9:5c:86:87")
	suite.NotContains(master3Params, "bond0")
	suite.Contains(master3Params, "rd.net.dhcp.retry=5")
	suite.Contains(master3Params, "rd.peerdns=0")
	suite.Contains(master3Params, "ip=mgmt0:dhcp")
	suite.NotContains(master3Params, "auto6")
	suite.NotContains(master3Params, "vlan")
	suite.NotContains(master3Params, "bootdev")
	suite.NotContains(master3Params, "hwprobe")

	// Verify kernel params for storage node with 1 NIC
	storage2Params := allBootParameters["x3000c0s15b0n0"].Params
	suite.Contains(storage2Params, "ifname=mgmt0:b8:59:9f:34:88:9e")
	suite.Contains(storage2Params, "ifname=mgmt1:b8:59:9f:34:88:9f")
	suite.NotContains(storage2Params, "ifname=sun")
	suite.NotContains(storage2Params, "bond0")
	suite.Contains(storage2Params, "rd.net.dhcp.retry=5")
	suite.Contains(storage2Params, "rd.peerdns=0")
	suite.Contains(storage2Params, "ip=mgmt0:dhcp")
	suite.NotContains(storage2Params, "auto6")
	suite.NotContains(storage2Params, "vlan")
	suite.NotContains(storage2Params, "bootdev")
	suite.NotContains(storage2Params, "hwprobe")

	// Verify kernel params for storage node with 2 NICs
	storage3Params := allBootParameters["x3000c0s17b0n0"].Params
	suite.Contains(storage3Params, "ifname=mgmt0:14:02:ec:d9:3e:90")
	suite.Contains(storage3Params, "ifname=sun0:14:02:ec:d9:3e:91")
	suite.Contains(storage3Params, "ifname=mgmt1:94:40:c9:5b:e5:70")
	suite.Contains(storage3Params, "ifname=sun1:94:40:c9:5b:e5:71")
	suite.NotContains(storage3Params, "bond0")
	suite.Contains(storage3Params, "rd.net.dhcp.retry=5")
	suite.Contains(storage3Params, "rd.peerdns=0")
	suite.Contains(storage3Params, "ip=mgmt0:dhcp")
	suite.NotContains(storage3Params, "auto6")
	suite.NotContains(storage3Params, "vlan")
	suite.NotContains(storage3Params, "bootdev")
	suite.NotContains(storage3Params, "hwprobe")

	bssGlobalHostRecords := getBSSGlobalHostRecords(managementNCNs, suite.slsNetworks)

	// When comparing the host_records the ordering could be different, but are functionally equivalent.
	suite.sortHostRecords(bssGlobalHostRecords)
	suite.sortHostRecords(suite.expectedGlobalHostRecords)
	suite.Equal(suite.expectedGlobalHostRecords, bssGlobalHostRecords)
}

func (suite *UpgradeBSSMetadataSuite) TestUpdateBSS_oneToOneTwo_CHNOnly() {
	//
	// Load in BSS test data
	//
	allBootParametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/csm1.0_bss_bootparameters.json")
	suite.NoError(err)

	var allBootParametersArray []bssTypes.BootParams
	err = json.Unmarshal(allBootParametersRaw, &allBootParametersArray)
	suite.NoError(err)

	allBootParameters := map[string]bssTypes.BootParams{}
	for _, bootParameters := range allBootParametersArray {
		if len(bootParameters.Hosts) != 1 {
			continue
		}
		allBootParameters[bootParameters.Hosts[0]] = bootParameters
	}

	// Verify the CSM 1.0 Global boot parameters contain the can-if and can-gw keys
	suite.Contains(allBootParameters["Global"].CloudInit.MetaData, "can-if")
	suite.Contains(allBootParameters["Global"].CloudInit.MetaData, "can-gw")

	// The SLS Network data is expected to be already configured for CSM 1.2
	networksRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/chn-only/csm1.2_sls_networks.json")
	suite.NoError(err)
	err = json.Unmarshal(networksRaw, &suite.slsNetworks)
	suite.NoError(err)
	suite.NotEmpty(suite.slsNetworks)

	// Load in the expected global BSS boot parameters for CSM 1.2
	expectedGlobalBootparametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/chn-only/csm1.2_expected_global_bootparameters.json")
	suite.NoError(err)
	err = json.Unmarshal(expectedGlobalBootparametersRaw, &suite.expectedGlobalBootparameters)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalBootparameters)

	// Extract the host_records from the global BSS boot parameters
	suite.expectedGlobalHostRecords = bss.HostRecords{}
	err = mapstructure.Decode(suite.expectedGlobalBootparameters.CloudInit.MetaData["host_records"], &suite.expectedGlobalHostRecords)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalHostRecords)

	//
	// Setup test HTTP servers for BSS and SLS
	//
	bssTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.T().Logf("Received BSS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/boot/v1/bootparameters" {
			host := r.URL.Query().Get("name")
			bootParameters, ok := allBootParameters[host]
			if !ok {
				suite.T().Logf("Unable to find host %s in BSS Bootparameters", host)
				suite.T().Fail()
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]bssTypes.BootParams{bootParameters})
			return
		} else if r.Method == http.MethodPatch && r.URL.Path == "/boot/v1/bootparameters" {
			var bootParameters bssTypes.BootParams
			err = json.NewDecoder(r.Body).Decode(&bootParameters)
			suite.NoError(err)

			if len(bootParameters.Hosts) != 1 {
				// For this test case we expect for 1 host to be set.
				suite.T().Logf("Received unexpected BSS Put request with %d hosts specified", len(bootParameters.Hosts))
				suite.T().Fail()
				w.WriteHeader(http.StatusBadRequest)
			}

			host := bootParameters.Hosts[0]
			if _, present := allBootParameters[host]; !present {
				suite.T().Logf("Received unexpected BSS Put request for nonexistent host %s", host)
				suite.T().Fail()
			}

			allBootParameters[host] = bootParameters
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bssTS.Close()

	slsTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.T().Logf("Received SLS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/v1/networks" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(suite.slsNetworks)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer slsTS.Close()

	// Configure BSS and SLS clients to use our test HTTP Servers
	bssClient = bss.NewBSSClient(bssTS.URL, bssTS.Client(), "")
	slsClient = sls.NewSLSClient(slsTS.URL, slsTS.Client(), "")

	// Function under test
	updateBSS10to12()

	// The global cloud-init meta-data should no longer contain the keys can-if and can-gw
	globalBootParameters := allBootParameters["Global"]
	suite.NotContains(globalBootParameters.CloudInit.MetaData, "can-if")
	suite.NotContains(globalBootParameters.CloudInit.MetaData, "can-gw")

	// Verify Global cloud-init data has updated host_records entries
	// When comparing the host_records the ordering could be different, but are functionally equivalent.
	globalHostRecords := suite.getHostRecords(globalBootParameters)
	suite.sortHostRecords(globalHostRecords)
	suite.sortHostRecords(suite.expectedGlobalHostRecords)
	suite.Equal(suite.expectedGlobalHostRecords, globalHostRecords)

	// Verify IPAM data for a single node:
	worker3BootParameters := allBootParameters["x3000c0s11b0n0"]
	var worker3IPAM bss.CloudInitIPAM
	err = mapstructure.Decode(worker3BootParameters.CloudInit.MetaData["ipam"], &worker3IPAM)
	suite.NoError(err)

	suite.Contains(worker3IPAM, "cmn")
	suite.Contains(worker3IPAM, "hmn")
	suite.Contains(worker3IPAM, "mtl")
	suite.Contains(worker3IPAM, "nmn")

	suite.Equal(bss.IPAMNetwork{Gateway: "10.102.4.1", CIDR: "10.102.4.12/25", ParentDevice: "bond0", VlanID: 7}, worker3IPAM["cmn"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.254.0.1", CIDR: "10.254.1.20/17", ParentDevice: "bond0", VlanID: 4}, worker3IPAM["hmn"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.1.0.1", CIDR: "10.1.1.10/16", ParentDevice: "bond0", VlanID: 0}, worker3IPAM["mtl"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.252.0.1", CIDR: "10.252.1.12/17", ParentDevice: "bond0", VlanID: 2}, worker3IPAM["nmn"])

	// The CHN should not be present in the BSS IPAM structure, as that network is managed differently
	suite.NotContains(worker3IPAM, "chn")
	suite.NotContains(worker3IPAM, "can")

	bssGlobalHostRecords := getBSSGlobalHostRecords(managementNCNs, suite.slsNetworks)

	// When comparing the host_records the ordering could be different, but are functionally equivalent.
	suite.sortHostRecords(bssGlobalHostRecords)
	suite.sortHostRecords(suite.expectedGlobalHostRecords)
	suite.Equal(suite.expectedGlobalHostRecords, bssGlobalHostRecords)
}
func (suite *UpgradeBSSMetadataSuite) TestUpdateBSS_oneToOneTwo_CANandCHN() {
	//
	// Load in BSS test data
	//
	allBootParametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/csm1.0_bss_bootparameters.json")
	suite.NoError(err)

	var allBootParametersArray []bssTypes.BootParams
	err = json.Unmarshal(allBootParametersRaw, &allBootParametersArray)
	suite.NoError(err)

	allBootParameters := map[string]bssTypes.BootParams{}
	for _, bootParameters := range allBootParametersArray {
		if len(bootParameters.Hosts) != 1 {
			continue
		}
		allBootParameters[bootParameters.Hosts[0]] = bootParameters
	}

	// Verify the CSM 1.0 Global boot parameters contain the can-if and can-gw keys
	suite.Contains(allBootParameters["Global"].CloudInit.MetaData, "can-if")
	suite.Contains(allBootParameters["Global"].CloudInit.MetaData, "can-gw")

	// The SLS Network data is expected to be already configured for CSM 1.2
	networksRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/can-chn/csm1.2_sls_networks.json")
	suite.NoError(err)
	err = json.Unmarshal(networksRaw, &suite.slsNetworks)
	suite.NoError(err)
	suite.NotEmpty(suite.slsNetworks)

	// Load in the expected global BSS boot parameters for CSM 1.2
	expectedGlobalBootparametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/can-chn/csm1.2_expected_global_bootparameters.json")
	suite.NoError(err)
	err = json.Unmarshal(expectedGlobalBootparametersRaw, &suite.expectedGlobalBootparameters)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalBootparameters)

	// Extract the host_records from the global BSS boot parameters
	suite.expectedGlobalHostRecords = bss.HostRecords{}
	err = mapstructure.Decode(suite.expectedGlobalBootparameters.CloudInit.MetaData["host_records"], &suite.expectedGlobalHostRecords)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalHostRecords)

	//
	// Setup test HTTP servers for BSS and SLS
	//
	bssTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.T().Logf("Received BSS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/boot/v1/bootparameters" {
			host := r.URL.Query().Get("name")
			bootParameters, ok := allBootParameters[host]
			if !ok {
				suite.T().Logf("Unable to find host %s in BSS Bootparameters", host)
				suite.T().Fail()
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]bssTypes.BootParams{bootParameters})
			return
		} else if r.Method == http.MethodPatch && r.URL.Path == "/boot/v1/bootparameters" {
			var bootParameters bssTypes.BootParams
			err = json.NewDecoder(r.Body).Decode(&bootParameters)
			suite.NoError(err)

			if len(bootParameters.Hosts) != 1 {
				// For this test case we expect for 1 host to be set.
				suite.T().Logf("Received unexpected BSS Put request with %d hosts specified", len(bootParameters.Hosts))
				suite.T().Fail()
				w.WriteHeader(http.StatusBadRequest)
			}

			host := bootParameters.Hosts[0]
			if _, present := allBootParameters[host]; !present {
				suite.T().Logf("Received unexpected BSS Put request for nonexistent host %s", host)
				suite.T().Fail()
			}

			allBootParameters[host] = bootParameters
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bssTS.Close()

	slsTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.T().Logf("Received SLS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/v1/networks" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(suite.slsNetworks)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer slsTS.Close()

	// Configure BSS and SLS clients to use our test HTTP Servers
	bssClient = bss.NewBSSClient(bssTS.URL, bssTS.Client(), "")
	slsClient = sls.NewSLSClient(slsTS.URL, slsTS.Client(), "")

	// Function under test
	updateBSS10to12()

	// The global cloud-init meta-data should no longer contain the keys can-if and can-gw
	globalBootParameters := allBootParameters["Global"]
	suite.NotContains(globalBootParameters.CloudInit.MetaData, "can-if")
	suite.NotContains(globalBootParameters.CloudInit.MetaData, "can-gw")

	// Verify Global cloud-init data has updated host_records entries
	// When comparing the host_records the ordering could be different, but are functionally equivalent.
	globalHostRecords := suite.getHostRecords(globalBootParameters)
	suite.sortHostRecords(globalHostRecords)
	suite.sortHostRecords(suite.expectedGlobalHostRecords)
	suite.Equal(suite.expectedGlobalHostRecords, globalHostRecords)

	// Verify IPAM data for a single node:
	worker3BootParameters := allBootParameters["x3000c0s11b0n0"]
	var worker3IPAM bss.CloudInitIPAM
	err = mapstructure.Decode(worker3BootParameters.CloudInit.MetaData["ipam"], &worker3IPAM)
	suite.NoError(err)

	suite.Contains(worker3IPAM, "can")
	suite.Contains(worker3IPAM, "cmn")
	suite.Contains(worker3IPAM, "hmn")
	suite.Contains(worker3IPAM, "mtl")
	suite.Contains(worker3IPAM, "nmn")

	suite.Equal(bss.IPAMNetwork{Gateway: "10.102.4.129", CIDR: "10.102.4.141/26", ParentDevice: "bond0", VlanID: 6}, worker3IPAM["can"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.102.4.1", CIDR: "10.102.4.12/25", ParentDevice: "bond0", VlanID: 7}, worker3IPAM["cmn"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.254.0.1", CIDR: "10.254.1.20/17", ParentDevice: "bond0", VlanID: 4}, worker3IPAM["hmn"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.1.0.1", CIDR: "10.1.1.10/16", ParentDevice: "bond0", VlanID: 0}, worker3IPAM["mtl"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.252.0.1", CIDR: "10.252.1.12/17", ParentDevice: "bond0", VlanID: 2}, worker3IPAM["nmn"])

	// The CHN should not be present in the BSS IPAM structure, as that network is managed differently
	suite.NotContains(worker3IPAM, "chn")

	bssGlobalHostRecords := getBSSGlobalHostRecords(managementNCNs, suite.slsNetworks)

	// When comparing the host_records the ordering could be different, but are functionally equivalent.
	suite.sortHostRecords(bssGlobalHostRecords)
	suite.sortHostRecords(suite.expectedGlobalHostRecords)
	suite.Equal(suite.expectedGlobalHostRecords, bssGlobalHostRecords)
}
func (suite *UpgradeBSSMetadataSuite) TestUpdateBSS_oneToOneTwo_NoCANorCHN() {
	//
	// Load in BSS test data
	//
	allBootParametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/csm1.0_bss_bootparameters.json")
	suite.NoError(err)

	var allBootParametersArray []bssTypes.BootParams
	err = json.Unmarshal(allBootParametersRaw, &allBootParametersArray)
	suite.NoError(err)

	allBootParameters := map[string]bssTypes.BootParams{}
	for _, bootParameters := range allBootParametersArray {
		if len(bootParameters.Hosts) != 1 {
			continue
		}
		allBootParameters[bootParameters.Hosts[0]] = bootParameters
	}

	// Verify the CSM 1.0 Global boot parameters contain the can-if and can-gw keys
	suite.Contains(allBootParameters["Global"].CloudInit.MetaData, "can-if")
	suite.Contains(allBootParameters["Global"].CloudInit.MetaData, "can-gw")

	// The SLS Network data is expected to be already configured for CSM 1.2
	networksRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/no-can-or-chn/csm1.2_sls_networks.json")
	suite.NoError(err)
	err = json.Unmarshal(networksRaw, &suite.slsNetworks)
	suite.NoError(err)
	suite.NotEmpty(suite.slsNetworks)

	// Load in the expected global BSS boot parameters for CSM 1.2
	expectedGlobalBootparametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/no-can-or-chn/csm1.2_expected_global_bootparameters.json")
	suite.NoError(err)
	err = json.Unmarshal(expectedGlobalBootparametersRaw, &suite.expectedGlobalBootparameters)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalBootparameters)

	// Extract the host_records from the global BSS boot parameters
	suite.expectedGlobalHostRecords = bss.HostRecords{}
	err = mapstructure.Decode(suite.expectedGlobalBootparameters.CloudInit.MetaData["host_records"], &suite.expectedGlobalHostRecords)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalHostRecords)

	//
	// Setup test HTTP servers for BSS and SLS
	//
	bssTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.T().Logf("Received BSS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/boot/v1/bootparameters" {
			host := r.URL.Query().Get("name")
			bootParameters, ok := allBootParameters[host]
			if !ok {
				suite.T().Logf("Unable to find host %s in BSS Bootparameters", host)
				suite.T().Fail()
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]bssTypes.BootParams{bootParameters})
			return
		} else if r.Method == http.MethodPatch && r.URL.Path == "/boot/v1/bootparameters" {
			var bootParameters bssTypes.BootParams
			err = json.NewDecoder(r.Body).Decode(&bootParameters)
			suite.NoError(err)

			if len(bootParameters.Hosts) != 1 {
				// For this test case we expect for 1 host to be set.
				suite.T().Logf("Received unexpected BSS Put request with %d hosts specified", len(bootParameters.Hosts))
				suite.T().Fail()
				w.WriteHeader(http.StatusBadRequest)
			}

			host := bootParameters.Hosts[0]
			if _, present := allBootParameters[host]; !present {
				suite.T().Logf("Received unexpected BSS Put request for nonexistent host %s", host)
				suite.T().Fail()
			}

			allBootParameters[host] = bootParameters
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bssTS.Close()

	slsTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.T().Logf("Received SLS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/v1/networks" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(suite.slsNetworks)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer slsTS.Close()

	// Configure BSS and SLS clients to use our test HTTP Servers
	bssClient = bss.NewBSSClient(bssTS.URL, bssTS.Client(), "")
	slsClient = sls.NewSLSClient(slsTS.URL, slsTS.Client(), "")

	// Function under test
	err = updateBSS10to12()

	suite.EqualError(err, "No CAN or CHN network defined in SLS")
}
func TestUpgradeBSSMetadataSuite(t *testing.T) {
	suite.Run(t, new(UpgradeBSSMetadataSuite))
}
