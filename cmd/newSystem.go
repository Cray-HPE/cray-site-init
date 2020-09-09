/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"stash.us.cray.com/MTL/sic/pkg/ipam"
)

// newSystemCmd represents the newSystem command
var newSystemCmd = &cobra.Command{
	Use:   "newSystem [path]",
	Short: "Create default configs for a new system (skeleton)",
	Long: `Create new configs in the provided directory.
	The new configs will have the correct structure and some sensible defaults, but no localized config yet.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			panic(errors.New("path needs to be provided"))
		}
		basepath, err := filepath.Abs(filepath.Clean(args[0]))
		if err != nil {
			panic(err)
		}
		dirs := []string{
			filepath.Join(basepath, "site_networks"),
			filepath.Join(basepath, "internal_networks"),
			filepath.Join(basepath, "certificates"),
			filepath.Join(basepath, "credentials"),
		}

		if err := os.Mkdir(basepath, 0777); err != nil {
			panic(err)
		}
		for _, dir := range dirs {
			if err := os.Mkdir(dir, 0777); err != nil {
				panic(err)
			}
		}

		fmt.Println(cmd.Flags().GetString("machine_name"))
	},
}

func init() {
	configCmd.AddCommand(newSystemCmd)

	newSystemCmd.Flags().Int16("river_cabinets", 5, "Number of Mountain Cabinets")
	newSystemCmd.Flags().Int16("mountain_cabinets", 5, "Number of Mountain Cabinets")
	newSystemCmd.Flags().String("machine_name", "sn-2024", "Name of the System")
	newSystemCmd.Flags().String("site_domain", "cray.io", "Site Domain Name")
	newSystemCmd.Flags().String("internal_domain", "unicos.shasta", "Internal Domain Name")
	newSystemCmd.Flags().IPNet("nmn_cidr", *ipam.NmnCIDR, "Overall CIDR for all Node Management subnets")
	newSystemCmd.Flags().IPNet("hmn_cidr", *ipam.HmnCIDR, "Overall CIDR for all Hardware Management subnets")
	newSystemCmd.Flags().IPNet("can_cidr", *ipam.CanCIDR, "Overall CIDR for all Customer Access subnets")
	newSystemCmd.Flags().IPNet("mtl_cidr", *ipam.MtlCIDR, "Overall CIDR for all Provisioning subnets")
	newSystemCmd.Flags().IPNet("slingshot_cidr", *ipam.HsnCIDR, "Overall CIDR for all Slingshot subnets")

}
