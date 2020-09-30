/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"stash.us.cray.com/MTL/sic/pkg/shasta"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load [path]",
	Short: "load a valid Shasta 1.4 directory of configuration",
	Long: `Load a set of files that represent a Shasta 1.4 system.
	Often load is used with init which generates the files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			panic(errors.New("path needs to be provided"))
		}
		basepath, err := filepath.Abs(filepath.Clean(args[0]))
		if err != nil {
			panic(err)
		}
		sysconfig, err := loadSystemConfig(filepath.Join(basepath, "system_config.yaml"))
		networks, err := extractNetworks(filepath.Join(basepath, "networks"))

		fmt.Println("Loaded ", sysconfig.SystemName, sysconfig.SiteDomain)
		for _, v := range networks {
			fmt.Println(v.Name, ":", v.CIDR, len(v.Subnets), "Subnets")
		}

	},
}

func init() {
	configCmd.AddCommand(loadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func extractNetworks(basepath string) ([]shasta.IPV4Network, error) {
	var networks []shasta.IPV4Network
	err := filepath.Walk(basepath,
		func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsRegular() {
				fmt.Println("Processing", path, "as IPV4Network -", info.Size())
				network := loadNetwork(path)
				// fmt.Println("Network is", network.Name)
				networks = append(networks, network)
			}
			return nil
		})
	return networks, err
}

func loadSystemConfig(path string) (shasta.SystemConfig, error) {
	var c shasta.SystemConfig
	info, err := os.Stat(path)
	fmt.Println("Processing", path, "as SystemConfig -", info.Size())
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		return c, err
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return c, err
	}
	return c, nil
}

func loadNetwork(path string) shasta.IPV4Network {
	var c shasta.IPV4Network
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}
