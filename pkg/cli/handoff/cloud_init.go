/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

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

package handoff

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	"github.com/spf13/cobra"
)

var ciData bssTypes.CloudInit

// NewHandoffCloudInitCommand creates the bss-update-cloud-init subcommand.
func NewHandoffCloudInitCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "bss-update-cloud-init",
		DisableAutoGenTag: true,
		Short:             "runs migration steps to update cloud-init parameters for NCNs",
		Long:              "Allows for the updating of cloud-init settings in BSS for all the NCNs",
		Run: func(c *cobra.Command, args []string) {
			setupCommon()

			log.Println("Updating NCN cloud-init parameters...")
			// Are we reading json from a file or the cli?
			if userDataJSON != "" {
				userDataFile, _ := os.ReadFile(userDataJSON)
				err := json.Unmarshal(
					userDataFile,
					&ciData,
				)
				if err != nil {
					log.Fatalln(
						"Couldn't parse user-data file: ",
						err,
					)
				}
				updateNCNCloudInitParamsFromFile()
			} else {
				// process data provided directly via cli
				updateNCNCloudInitParams()
			}
			log.Println("Done updating NCN cloud-init parameters.")
		},
	}

	c.Flags().StringArrayVar(
		&paramsToUpdate,
		"set",
		[]string{},
		"For each cloud-init object you wish to update or add list it in the format of key=value",
	)
	c.Flags().StringArrayVar(
		&paramsToDelete,
		"delete",
		[]string{},
		"For each cloud-init object you wish to remove provide just the key and it will be removed "+
			"regardless of value",
	)
	c.Flags().StringArrayVar(
		&limitToXnames,
		"limit",
		[]string{},
		"Limit updates to just the xnames specified",
	)
	c.Flags().StringVar(
		&userDataJSON,
		"user-data",
		"",
		"json-formatted file with cloud-init user-data",
	)
	return c
}

func getFinalJSONObject(
	key string, bssEntry *bssTypes.BootParams,
) (
	string, *map[string]interface{},
) {
	// To make this as easy as possible to use all params are specified in the format of
	// user-data.key1.key2=value where each key is an object. Thus what we need to do is drill down into the
	// appropriate structure until we find the object we need to set. Note this logic does *not* support arrays.
	keyParts := strings.Split(
		key,
		".",
	)

	var object map[string]interface{}
	switch keyParts[0] {
		case "user-data":
		if bssEntry.CloudInit.UserData == nil {
			bssEntry.CloudInit.UserData = make(map[string]interface{})
		}

		object = bssEntry.CloudInit.UserData
	case "meta-data":
		if bssEntry.CloudInit.MetaData == nil {
			bssEntry.CloudInit.MetaData = make(map[string]interface{})
		}

		object = bssEntry.CloudInit.MetaData
	default:
		log.Fatalf(
			"Unknown root key: %s",
			keyParts[0],
		)
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
			log.Printf(
				"Failed to find next key (%s in %s) in object, creating.",
				nextKey,
				key,
			)
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
		// Get the BSS bootparameters for this NCN.
		bssEntry := getBSSBootparametersForXname(ncn.Xname)

		if ciData.UserData != nil {

			for key, val := range ciData.UserData {
				log.Printf(
					"Updating %s on %s\n",
					key,
					ncn.Xname,
				)

				object := bssEntry.CloudInit.UserData
				// key is user-data[key]
				object[key] = val
			}
		}

		if ciData.MetaData != nil {

			for key, val := range ciData.MetaData {
				log.Printf(
					"Updating %s on %s\n",
					key,
					ncn.Xname,
				)

				object := bssEntry.CloudInit.MetaData
				// key is user-data[key]
				object[key] = val
			}
		}

		// Now write it back to BSS.
		log.Printf(
			"Writing back to BSS: %s\n",
			ncn.Xname,
		)
		uploadEntryToBSS(
			bssEntry,
			http.MethodPut,
		)
	}
}

func updateNCNCloudInitParams() {
	limitManagementNCNs, setParams := setupHandoffCommon()

	for _, ncn := range limitManagementNCNs {
		// Get the BSS bootparameters for this NCN.
		bssEntry := getBSSBootparametersForXname(ncn.Xname)

		// Create/update params.
		for _, setParam := range setParams {
			key, object := getFinalJSONObject(
				setParam.key,
				&bssEntry,
			)
			objectVal := *object

			var value interface{}

			// Handle arrays of strings.
			var potentialArray []string
			arrayErr := json.Unmarshal(
				[]byte(setParam.value),
				&potentialArray,
			)
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
			key, object := getFinalJSONObject(
				deleteParam,
				&bssEntry,
			)

			delete(
				*object,
				key,
			)
		}

		// Now write it back to BSS.
		uploadEntryToBSS(
			bssEntry,
			http.MethodPut,
		)
	}
}
