package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"
	"net/http"

	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

var (
	oneToOneTwo  bool
	ipamNetNames []string
	ipamKey      map[string]interface{}
)

// The Ipam type represents all the metadata needed by the cloud-init network module inside a go struct
type Ipam struct {
	IpamNetworks
}

type IpamNetworks struct {
	CAN IpamNetwork `json:"can"`
	CMN IpamNetwork `json:"cmn"`
	HMN IpamNetwork `json:"hmn"`
	MTL IpamNetwork `json:"mtl"`
	NMN IpamNetwork `json:"nmn"`
}

type IpamNetwork struct {
	Gateway      string `json:"gateway"`
	CIDR         string `json:"ip"`
	ParentDevice string `json:"parent_device"`
	VlanID       int    `json:"vlanid"`
}

// metadataCmd represents the upgrade command
var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Upgrades metadata",
	Long: `
	Upgrades cloud-init metadata and pushes it to BSS`,
	Run: func(cmd *cobra.Command, args []string) {
		v := viper.GetViper()
		v.BindPFlags(cmd.Flags())

		if oneToOneTwo {
			fmt.Println("Begining upgrade of cloud-init metadata...")
			// query SLS for new information needed for the upgrade
			// setupCommon()

			// All the networks we need
			ipamNetNames = []string{
				"CAN",
				"CMN",
				"HMN",
				"MTL",
				"NMN"}

			// This will become the payload that is pushed to BSS
			ipamKey = make(map[string]interface{})

			for _, n := range ipamNetNames {
				fmt.Println("Gathering", n, "networking from SLS...")

				// Query the required keys from SLS
				// FIXME: these need to actually query SLS
				gateway := "gw"
				cidr := "cidr"
				parentDevice := "parent"
				vlanID := 2

				// Instantiate a new IpamNetwork type with the queried info from SLS
				ipamNet := IpamNetwork{
					Gateway:      gateway,
					CIDR:         cidr,
					ParentDevice: parentDevice,
					VlanID:       vlanID,
				}

				// Print the data discovered for each network
				fmt.Println(ipamNet)

				// Add the dict into the ipam key we're building
				ipamKey[n] = ipamNet

			}
		}

		// Show the user the final key that will be added to BSS
		fmt.Println(ipamKey)

		limitManagementNCNs, _ := setupHandoffCommon()

		for _, ncn := range limitManagementNCNs {
			ipamEntry := bssTypes.BootParams{
				CloudInit: bssTypes.CloudInit{
					MetaData: ipamKey,
				},
			}

			// PATCH in the new data to BSS
			fmt.Printf("Upgrading ipam for %s\n", ncn.Xname)
			uploadEntryToBSS(ipamEntry, http.MethodPatch)
		}

	},
}

func init() {
	upgradeCmd.AddCommand(metadataCmd)
	metadataCmd.DisableAutoGenTag = true

	metadataCmd.Flags().SortFlags = true
	metadataCmd.Flags().BoolVarP(&oneToOneTwo, "1-0-to-1-2", "", false, "Upgrade CSM 1.0 metadata to 1.2 metadata")
}
