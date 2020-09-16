/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init --from-sls-file=<file> --from-1.3-dir=<dir> --from-sls=<sls url>",
	Short: "init",
	Long: `init generates a scaffolding the Shasta 1.4 configuration payload.  
	It can optionally load from the sls or files of an existing 1.3 system.
	Files used from the 1.3 system include:
	 - system_config.yml
	 - ncn_metadata.csv
	 - networks_derived.yml
	 - hmn_connections.json
	 - customer_var.yml`,
	Run: func(cmd *cobra.Command, args []string) {
		var slsState sls_common.SLSState
		var err error
		// Favor the SLS file.  Try parsing it first.
		if viper.GetString("slsfilepath") != "" {
			slsState, err = shasta.ParseSLSFile(viper.GetString("slsfilepath"))
			if err != nil {
				panic(err)
			}
			// If we couldn't populate the file, try the URL instead
		} else if viper.GetString("slsurl") != "" {
			slsState, err = shasta.ParseSLSfromURL(viper.GetString("slsurl"))
			if err != nil {
				panic(err)
			}
		}
		if unsafe.Sizeof(slsState) > 0 {
			// At this point, we should have a valid slsState
			networks := shasta.ConvertSLSNetworks(slsState)
			fmt.Println(networks)
		}
		if viper.GetString("oldconfigdir") != "" {
			LoadConfig()
			MergeNCNMetadata()
			MergeNetworksDerived()
			MergeSLSInput()
			MergeCustomerVar()
		}

		// InitializeConfiguration()
		// WriteConfigFile()
	},
}

func init() {
	configCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings
	initCmd.Flags().String("from-sls-file", "", "SLS File Location")
	viper.BindPFlag("slsfilepath", initCmd.Flags().Lookup("from-sls-file"))

	initCmd.Flags().String("from-1.3-dir", "", "Shasta 1.3 Configuration Directory")
	viper.BindPFlag("oldconfigdir", initCmd.Flags().Lookup("from-1.3-dir"))

	initCmd.Flags().String("from-sls", "", "Shasta 1.3 SLS dumpstate url")
	viper.BindPFlag("slsurl", initCmd.Flags().Lookup("from-sls"))

}
