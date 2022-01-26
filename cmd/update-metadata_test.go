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
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func getHostRecords(t *testing.T, globalBootParameters bssTypes.BootParams) []bss.HostRecord {
	var globalHostRecords []bss.HostRecord
	err := mapstructure.Decode(globalBootParameters.CloudInit.MetaData["host_records"], &globalHostRecords)
	assert.NoError(t, err)

	return globalHostRecords
}

func findHostRecord(t *testing.T, hostRecords []bss.HostRecord, alias string) bss.HostRecord {
	for _, hostRecord := range hostRecords {
		for _, hostAlias := range hostRecord.Aliases {
			if hostAlias == alias {
				return hostRecord
			} 
		}
	}

	t.Errorf("Unable to find host record for %s", alias)
	t.Fail()
	return bss.HostRecord{} 
}

func sortHostRecords(hostRecords []bss.HostRecord) {
	sort.SliceStable(hostRecords, func(i, j int) bool{
		return hostRecords[i].IP < hostRecords[j].IP
	})
}

func TestUpgradeBSSMetadata(t *testing.T) {
	//
	// Load in BSS test data
	//
	allBootParametersRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/bss_bootparameters_csm1.0.json")
	assert.NoError(t, err)

	var allBootParametersArray []bssTypes.BootParams
	err = json.Unmarshal(allBootParametersRaw, &allBootParametersArray)
	assert.NoError(t, err)

	allBootParameters := map[string]bssTypes.BootParams{}
	for _, bootParameters := range allBootParametersArray {
		if len(bootParameters.Hosts) != 1 {
			continue
		}
		allBootParameters[bootParameters.Hosts[0]] = bootParameters
	}

	// Load in management NCNs from SLS
	ncnsRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/sls_management_ncns.json")
	assert.NoError(t, err)
	err = json.Unmarshal(ncnsRaw, &managementNCNs)
	assert.NoError(t, err)

	//
	// Setup test HTTP servers for BSS and SLS
	//
	bssTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Received BSS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/boot/v1/bootparameters" {
			host := r.URL.Query().Get("name")
			bootParameters, ok := allBootParameters[host]
			if !ok {
				t.Logf("Unable to find host %s in BSS Bootparameters", host)
				t.Fail()
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]bssTypes.BootParams{bootParameters})
			return
		} else if r.Method == http.MethodPatch && r.URL.Path == "/boot/v1/bootparameters" {
			var bootParameters bssTypes.BootParams
			err = json.NewDecoder(r.Body).Decode(&bootParameters)
			assert.NoError(t, err)

			if len(bootParameters.Hosts) != 1 {
				// For this test case we expect for 1 host to be set.
				t.Logf("Received unexpected BSS Put request with %d hosts specified", len(bootParameters.Hosts))
				t.Fail()
				w.WriteHeader(http.StatusBadRequest)
			}

			host := bootParameters.Hosts[0]
			if _, present := allBootParameters[host]; !present {
				t.Logf("Received unexpected BSS Put request for nonexistant host %s", host)
				t.Fail()
			}

			allBootParameters[host] = bootParameters
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bssTS.Close()

	slsTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Received SLS request for %s", r.URL)
		if r.Method == http.MethodGet && r.URL.Path == "/v1/search/hardware" && r.URL.Query().Get("extra_properties.Role") == "Management" {
		} else if r.Method == http.MethodGet && r.URL.Path == "/v1/networks" {
			// The SLS Network data is expected to be already configured for CSM 1.2
			networks, err := ioutil.ReadFile("../testdata/upgrade-bss/sls_networks.json")
			assert.NoError(t, err)

			w.WriteHeader(http.StatusOK)
			w.Write(networks)
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

	// Verify Global cloud-init data has updated host_records entries
	globalHostRecords := getHostRecords(t, allBootParameters["Global"])
	sortHostRecords(globalHostRecords)

	var expectedGlobalHostRecords bss.HostRecords
	expectedGlobalHostRecordsRaw, err := ioutil.ReadFile("../testdata/upgrade-bss/expected_global_host_records.json")
	assert.NoError(t, err)

	err = json.Unmarshal(expectedGlobalHostRecordsRaw, &expectedGlobalHostRecords)
	assert.NoError(t, err)
	sortHostRecords(expectedGlobalHostRecords)


	assert.Equal(t, expectedGlobalHostRecords, globalHostRecords)


	actualGlobalHostRecordsOut, _ := json.MarshalIndent(globalHostRecords, "", "  ")
	expectedGlobalHostRecordsOut, _ := json.MarshalIndent(expectedGlobalHostRecords, "", "  ")
	ioutil.WriteFile("actual_host_records.json", actualGlobalHostRecordsOut, 0644)
	ioutil.WriteFile("expected_host_records.json", expectedGlobalHostRecordsOut, 0644)
}