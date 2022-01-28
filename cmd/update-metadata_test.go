package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/
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

	// The SLS Network data is expected to be already configured for CSM 1.2
	networksRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/csm1.2_sls_networks.json")
	suite.NoError(err)
	err = json.Unmarshal(networksRaw, &suite.slsNetworks)
	suite.NoError(err)
	suite.NotEmpty(suite.slsNetworks)

	// Load in the expected global BSS boot parameters for CSM 1.2
	expectedGlobalBootparametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/csm1.0-csm1.2/csm1.2_expected_global_bootparameters.json")
	suite.NoError(err)
	err = json.Unmarshal(expectedGlobalBootparametersRaw, &suite.expectedGlobalBootparameters)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalBootparameters)

	// Extract the host_records from the global BSS boot parameters
	suite.expectedGlobalHostRecords = bss.HostRecords{}
	err = mapstructure.Decode(suite.expectedGlobalBootparameters.CloudInit.MetaData["host_records"], &suite.expectedGlobalHostRecords)
	suite.NoError(err)
	suite.NotEmpty(suite.expectedGlobalHostRecords)
}

func (suite *UpgradeBSSMetadataSuite) TestUpdateBSS_oneToOneTwo() {
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
				suite.T().Logf("Received unexpected BSS Put request for nonexistant host %s", host)
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
	updateBSS_oneToOneTwo()

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
	suite.Equal(bss.IPAMNetwork{Gateway: "10.102.4.1", CIDR: "10.102.4.27/25", ParentDevice: "bond0", VlanID: 7}, worker3IPAM["cmn"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.254.0.1", CIDR: "10.254.1.20/17", ParentDevice: "bond0", VlanID: 4}, worker3IPAM["hmn"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.1.0.1", CIDR: "10.1.1.10/16", ParentDevice: "bond0", VlanID: 0}, worker3IPAM["mtl"])
	suite.Equal(bss.IPAMNetwork{Gateway: "10.252.0.1", CIDR: "10.252.1.12/17", ParentDevice: "bond0", VlanID: 2}, worker3IPAM["nmn"])

	// The CHN should not be present in the BSS IPAM structure, as that network is managed differently
	suite.NotContains(worker3IPAM, "chn")
}

func (suite *UpgradeBSSMetadataSuite) TestGetBSSGlobalHostRecords() {
	globalHostRecords := getBSSGlobalHostRecords(managementNCNs, suite.slsNetworks)

	// When comparing the host_records the ordering could be different, but are functionally equivalent.
	suite.sortHostRecords(globalHostRecords)
	suite.sortHostRecords(suite.expectedGlobalHostRecords)
	suite.Equal(suite.expectedGlobalHostRecords, globalHostRecords)
}

func TestUpgradeBSSMetadataSuite(t *testing.T) {
	suite.Run(t, new(UpgradeBSSMetadataSuite))
}
