/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"net"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"stash.us.cray.com/MTL/sic/pkg/ipam"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// subnetCmd represents the subnet command
var subnetCmd = &cobra.Command{
	Use:   "subnet [name]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := viper.GetViper()
		var n shasta.IPV4Subnet
		err := v.Unmarshal(&n)
		if err != nil {
			fmt.Printf("unable to decode into struct, %v \n", err)
		}
		n.Name = args[0]
		_, network, err := net.ParseCIDR(v.GetString("within"))
		viperSize, err := strconv.Atoi(v.GetString("size"))
		n.CIDR, err = ipam.SubnetWithin(*network, viperSize)
		bs, err := yaml.Marshal(&n)
		fmt.Print(string(bs))
	},
}

func init() {
	rawCmd.AddCommand(subnetCmd)

	subnetCmd.Flags().String("full_name", "", "Long Descriptive Name for the Subnet")
	subnetCmd.Flags().Int("size", 16, "Number of ip addresses in the subnet")
	subnetCmd.Flags().Int16("vlan_id", 0, "Preferred VlanID")
	subnetCmd.Flags().String("comment", "", "Subnet Comment")
	subnetCmd.Flags().IP("gateway", net.IP{}, "Subnet Gateway")
	subnetCmd.Flags().IPNet("within", net.IPNet{}, "Overall IPv4 CIDR for all Provisioning subnets")
}
