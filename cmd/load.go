/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	shcd_parser "stash.us.cray.com/HMS/hms-shcd-parser/pkg/shcd-parser"
	csiFiles "stash.us.cray.com/MTL/csi/internal/files"
	"stash.us.cray.com/MTL/csi/pkg/shasta"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load <path>",
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
		sysconfig, _ := loadSystemConfig(filepath.Join(basepath, "system_config.yaml"))
		networks, _ := extractNetworks(filepath.Join(basepath, "networks"))

		log.Println("Loaded ", sysconfig.SystemName, sysconfig.SiteDomain)
		for _, v := range networks {
			log.Println(v.Name, ":", v.CIDR, len(v.Subnets), "Subnets")
		}

	},
}

func init() {
	configCmd.AddCommand(loadCmd)
}

func loadSystemConfig(path string) (sysconf shasta.SystemConfig, err error) {
	err = csiFiles.ReadYAMLConfig(path, &sysconf)
	return
}

func loadNetwork(path string) (network shasta.IPV4Network, err error) {
	err = csiFiles.ReadJSONConfig(path, &network)
	return
}

func extractNetworks(basepath string) ([]shasta.IPV4Network, error) {
	// TODO: Handle incoming error?
	var networks []shasta.IPV4Network
	err := filepath.Walk(basepath,
		func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsRegular() {
				log.Println("Processing", path, "as IPV4Network -", info.Size())
				network, err := loadNetwork(path)
				if err != nil {
					log.Printf("Failed loading network from %v: %v\n", path, err)
				}
				networks = append(networks, network)
			}
			return nil
		})
	return networks, err
}

func loadHMNConnectionsFile(path string) (rows []shcd_parser.HMNRow, err error) {
	err = csiFiles.ReadJSONConfig(path, &rows)
	return
}

func loadNCNMetadataFile(path string) (ncns []*shasta.LogicalNCN, err error) {
	// I know this is a little silly, but it improves readability and
	// gives us future flexibility
	return csiFiles.ReadNodeCSV(path)
}

func loadMgmtMetadataFile(path string) (switches []*shasta.ManagementSwitch, err error) {
	// I know this is a little silly, but it improves readability and
	// gives us future flexibility
	return csiFiles.ReadSwitchCSV(path)
}
