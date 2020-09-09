/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"stash.us.cray.com/MTL/sic/pkg/ipam"
)

// addNetworkCmd represents the addNetwork command
var addNetworkCmd = &cobra.Command{
	Use:   "addNetwork",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetString("template") != "" {
			network, err := ipam.LoadNetworkFromFile(viper.GetString("template"))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(network)
		}
		fmt.Println("addNetwork called")
		// LoadConfig()
	},
}

func init() {
	configCmd.AddCommand(addNetworkCmd)

	var defaultCidr net.IPNet
	var defaultName, filename string
	var defaultVlan int16

	addNetworkCmd.Flags().String("template", filename, "YAML file matching the Network Struct")
	if err := viper.BindPFlag("template", addNetworkCmd.Flags().Lookup("template")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
	addNetworkCmd.Flags().IPNet("cidr", defaultCidr, "Subnet CIDR for the network")
	addNetworkCmd.Flags().String("name", defaultName, "Name of the Network")
	addNetworkCmd.Flags().Int16("vlan", defaultVlan, "Vlan ID for Network")
}
