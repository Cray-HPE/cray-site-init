/*
 *
 *  MIT License
 *
 *  (C) Copyright 2020-2022 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */

package cmd

import (
	"fmt"
	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"strings"
)

var (
	kernel string
	initrd string
)

var handoffBSSUpdateParamCmd = &cobra.Command{
	Use:   "bss-update-param",
	Short: "runs migration steps to update kernel parameters for NCNs",
	Long:  "Allows for the updating of kernel parameters in BSS for all the NCNs",
	Run: func(cmd *cobra.Command, args []string) {
		setupCommon()

		log.Println("Updating NCN kernel parameters...")
		updateNCNKernelParams()
		log.Println("Done updating NCN kernel parameters.")
	},
}

func init() {
	handoffCmd.AddCommand(handoffBSSUpdateParamCmd)
	handoffBSSUpdateParamCmd.DisableAutoGenTag = true

	handoffBSSUpdateParamCmd.Flags().StringArrayVar(&paramsToUpdate, "set", []string{},
		"For each kernel parameter you wish to update or add list it in the format of key=value")
	handoffBSSUpdateParamCmd.Flags().StringArrayVar(&paramsToDelete, "delete", []string{},
		"For each kernel parameter you wish to remove provide just the key and it will be removed "+
			"regardless of value")
	handoffBSSUpdateParamCmd.Flags().StringVar(&kernel, "kernel", "",
		"New value to set for the kernel")
	handoffBSSUpdateParamCmd.Flags().StringVar(&initrd, "initrd", "",
		"New value to set for the initrd")
	handoffBSSUpdateParamCmd.Flags().StringArrayVar(&limitToXnames, "limit", []string{},
		"Limit updates to just the xnames specified")
}

func updateParams(params []string, setParams []paramTuple, addParams []paramTuple, deleteParams []string) []string {
	// If we were told to delete any, do that first so that a subsequent set can be truly fresh.
	for _, deleteParam := range deleteParams {
		for i := range params {
			thisParam := &params[i]

			potentialParamSplit := strings.Split(*thisParam, "=")

			if len(potentialParamSplit) < 2 {
				// Ignore any params that don't have key=value format.
				continue
			}

			if potentialParamSplit[0] == deleteParam {
				*thisParam = ""
			}
		}
	}

	// For each of the given set parameters check to see if the value is already set. If it is, overwrite it.
	// If it is not however then add it to the parameters.
	for _, setParam := range setParams {
		found := false
		setParamString := fmt.Sprintf("%s=%s", setParam.key, setParam.value)

		for i := range params {
			thisParam := &params[i]

			potentialParamSplit := strings.Split(*thisParam, "=")

			if len(potentialParamSplit) != 2 {
				// Ignore any params that don't have key=value format.
				continue
			}

			if potentialParamSplit[0] == setParam.key {
				found = true

				*thisParam = setParamString
			}
		}

		// If we didn't find it, append it.
		if !found {
			params = append(params, setParamString)
		}
	}

	// For each of the given add parameters we will just add the parameter without checking if the key exists.
	// This is for parameters like "ifname" that need to be added multiple times with different values.
	// These items will be added to the front of the params.
	for _, addParam := range addParams {
		addParamString := fmt.Sprintf("%s=%s", addParam.key, addParam.value)

		params = append([]string{addParamString}, params...)
	}

	// Just to get rid of the whitespace.
	var finalParts []string
	for _, part := range params {
		if part != "" {
			finalParts = append(finalParts, part)
		}
	}

	return finalParts
}

func updateNCNKernelParams() {
	limitManagementNCNs, setParams := setupHandoffCommon()

	for _, ncn := range limitManagementNCNs {
		// Get the BSS bootparameters for this NCN.
		bssEntry := getBSSBootparametersForXname(ncn.Xname)

		params := strings.Split(bssEntry.Params, " ")

		finalParts := updateParams(params, setParams, []paramTuple{}, paramsToDelete)

		// Create a whole new structure to PATCH this entry with to not touch other pieces of the structure.
		newBSSEntry := bssTypes.BootParams{
			Hosts:  []string{ncn.Xname},
			Params: strings.Join(finalParts, " "),
		}

		// If the kernel and/or initrd are set, update them now.
		if kernel != "" {
			newBSSEntry.Kernel = kernel
		}
		if initrd != "" {
			newBSSEntry.Initrd = initrd
		}

		// Now write it back to BSS.
		uploadEntryToBSS(newBSSEntry, http.MethodPatch)
	}

}
