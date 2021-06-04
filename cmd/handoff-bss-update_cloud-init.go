/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"stash.us.cray.com/HMS/hms-bss/pkg/bssTypes"
)

var ciData bssTypes.CloudInit

// "raw" to distiguish it from the purely string-based paramTuple used elsewhere
type rawParamTuple struct {
	key   string
	value interface{}
}

var handoffBSSUpdateCloudInitCmd = &cobra.Command{
	Use:   "bss-update-cloud-init",
	Short: "runs migration steps to update cloud-init parameters for NCNs",
	Long:  "Allows for the updating of cloud-init settings in BSS for all the NCNs",
	Run: func(cmd *cobra.Command, args []string) {
		setupCommon()

		log.Println("Updating NCN cloud-init parameters...")
		// Are we reading json from a file or the cli?
		if userDataJSON != "" {
			userDataFile, _ := ioutil.ReadFile(userDataJSON)
			err := json.Unmarshal(userDataFile, &ciData)
			if err != nil {
				log.Fatalln("Couldn't parse user-data file: ", err)
			}
			if ciData.UserData == nil {
				log.Fatalf("Could not find \"user-data\" key in %s\n", userDataJSON)
			}
			updateNCNCloudInitParamsFromFile()
		} else {
			// process data provided directly via cli
			updateNCNCloudInitParams()
		}
		log.Println("Done updating NCN cloud-init parameters.")
	},
}

func init() {
	handoffCmd.AddCommand(handoffBSSUpdateCloudInitCmd)

	handoffBSSUpdateCloudInitCmd.Flags().StringArrayVar(&paramsToUpdate, "set", []string{},
		"For each kernel parameter you wish to update or add list it in the format of key=value")
	handoffBSSUpdateCloudInitCmd.Flags().StringArrayVar(&paramsToDelete, "delete", []string{},
		"For each kernel parameter you wish to remove provide just the key and it will be removed "+
			"regardless of value")
	handoffBSSUpdateCloudInitCmd.Flags().StringArrayVar(&limitToXnames, "limit", []string{},
		"Limit updates to just the xnames specified")
	handoffBSSUpdateCloudInitCmd.Flags().StringVar(&userDataJSON, "user-data", "",
		"json-formatted file with global cloud-init user-data")
}

func getFinalJSONObject(key string, bssEntry *bssTypes.BootParams) (string, *map[string]interface{}) {
	// To make this as easy as possible to use all params are specified in the format of
	// user-data.key1.key2=value where each key is an object. Thus what we need to do is drill down into the
	// appropriate structure until we find the object we need to set. Note this logic does *not* support arrays.
	keyParts := strings.Split(key, ".")

	var object map[string]interface{}
	if keyParts[0] == "user-data" {
		if bssEntry.CloudInit.UserData == nil {
			bssEntry.CloudInit.UserData = make(map[string]interface{})
		}

		object = bssEntry.CloudInit.UserData
	} else if keyParts[0] == "meta-data" {
		if bssEntry.CloudInit.MetaData == nil {
			bssEntry.CloudInit.MetaData = make(map[string]interface{})
		}

		object = bssEntry.CloudInit.MetaData
	} else {
		log.Fatalf("Unknown root key: %s", keyParts[0])
	}

	keyIndex := 1
	nextKey := keyParts[keyIndex]
	var nextObject map[string]interface{}
	for keyIndex+1 < len(keyParts) {
		var ok bool
		if nextObject, ok = object[nextKey].(map[string]interface{}); !ok {
			// If it doesn't exist, create it.
			object[nextKey] = make(map[string]interface{})
			nextObject = object[nextKey].(map[string]interface{})
			log.Printf("Failed to find next key (%s in %s) in object, creating.", nextKey, key)
		}
		object = nextObject

		keyIndex++
		nextKey = keyParts[keyIndex]
	}

	return nextKey, &object
}

func updateNCNCloudInitParamsFromFile() {
	limitManagementNCNs, _ := setupHandoffCommon()

	for _, ncn := range limitManagementNCNs {
		// Get the BSS bootparamaters for this NCN.
		bssEntry := getBSSBootparametersForXname(ncn.Xname)

		for key, val := range ciData.UserData {
			object := bssEntry.CloudInit.UserData
			// key is user-data[key]
			object[key] = val
		}

		// Now write it back to BSS.
		uploadEntryToBSS(bssEntry, http.MethodPut)
	}
}

func updateNCNCloudInitParams() {
	limitManagementNCNs, setParams := setupHandoffCommon()

	for _, ncn := range limitManagementNCNs {
		// Get the BSS bootparamaters for this NCN.
		bssEntry := getBSSBootparametersForXname(ncn.Xname)

		// Create/update params.
		for _, setParam := range setParams {
			key, object := getFinalJSONObject(setParam.key, &bssEntry)
			objectVal := *object

			var value interface{}

			// Handle arrays of strings.
			var potentialArray []string
			arrayErr := json.Unmarshal([]byte(setParam.value), &potentialArray)
			if arrayErr == nil {
				// Must be an array.
				value = potentialArray
			} else {
				value = setParam.value
			}

			objectVal[key] = value
		}

		// Delete params.
		for _, deleteParam := range paramsToDelete {
			key, object := getFinalJSONObject(deleteParam, &bssEntry)

			delete(*object, key)
		}

		// Now write it back to BSS.
		uploadEntryToBSS(bssEntry, http.MethodPut)
	}
}
