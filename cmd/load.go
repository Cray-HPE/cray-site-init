/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
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
			log.Fatalln(errors.New("path needs to be provided"))
		}
		basepath, err := filepath.Abs(filepath.Clean(args[0]))
		if err != nil {
			log.Fatalln(err)
		}
		sysconfig, err := loadSystemConfig(filepath.Join(basepath, "system_config.yaml"))
		networks, err := extractNetworks(filepath.Join(basepath, "networks"))

		log.Println("Loaded ", sysconfig.SystemName, sysconfig.SiteDomain)
		for _, v := range networks {
			log.Println(v.Name, ":", v.CIDR, len(v.Subnets), "Subnets")
		}

	},
}

func init() {
	configCmd.AddCommand(loadCmd)
}

func extractNetworks(basepath string) ([]shasta.IPV4Network, error) {
	// TODO: Handle incoming error?
	var networks []shasta.IPV4Network
	err := filepath.Walk(basepath,
		func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsRegular() {
				log.Println("Processing", path, "as IPV4Network -", info.Size())
				network, er := loadNetwork(path)
				if err != nil {
					log.Printf("Unable to extract network from %v.  Error was: %v \n", path, er)
				}
				networks = append(networks, network)
			}
			return nil
		})
	return networks, err
}

func loadSystemConfig(path string) (shasta.SystemConfig, error) {
	var c shasta.SystemConfig
	info, err := os.Stat(path)
	log.Println("Processing", path, "as SystemConfig -", info.Size())
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}

func loadNetwork(path string) (shasta.IPV4Network, error) {
	var c shasta.IPV4Network
	info, _ := os.Stat(path)
	log.Printf("Processing %v as Network - (%v)", path, info.Size())
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(yamlFile, &c)
	return c, err
}
