/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"stash.us.cray.com/HMS/hms-bss/pkg/bssTypes"
	"strings"
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

	handoffBSSUpdateParamCmd.Flags().StringArrayVar(&paramsToUpdate, "set", []string{},
		"For each kernel parameter you wish to update or add list it in the format of key=value")
	handoffBSSUpdateParamCmd.Flags().StringArrayVar(&paramsToDelete, "delete", []string{},
		"For each kernel parameter you wish to remove provide just the key and it will be removed "+
			"regardless of value")
	handoffBSSUpdateParamCmd.Flags().StringArrayVar(&limitToXnames, "limit", []string{},
		"Limit updates to just the xnames specified")
}

func updateNCNKernelParams() {
	limitManagementNCNs, setParams := setupHandoffCommon()

	for _, ncn := range limitManagementNCNs {
		// Get the BSS bootparamaters for this NCN.
		bssEntry := getBSSBootparametersForXname(ncn.Xname)

		params := strings.Split(bssEntry.Params, " ")

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

		// If we were told to delete any, do that now.
		for _, deleteParam := range paramsToDelete {
			for i := range params {
				thisParam := &params[i]

				potentialParamSplit := strings.Split(*thisParam, "=")

				if len(potentialParamSplit) != 2 {
					// Ignore any params that don't have key=value format.
					continue
				}

				if potentialParamSplit[0] == deleteParam {
					*thisParam = ""
				}
			}
		}

		// Just to get rid of the whitespace.
		var finalParts []string
		for _, part := range params {
			if part != "" {
				finalParts = append(finalParts, part)
			}
		}

		// Create a whole new structure to PATCH this entry with so as to not touch other pieces of the structure.
		newBSSEntry := bssTypes.BootParams{
			Hosts:  []string{ncn.Xname},
			Params: strings.Join(finalParts, " "),
		}

		// Now write it back to BSS.
		uploadEntryToBSS(newBSSEntry, http.MethodPatch)
	}

}
