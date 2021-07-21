/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	"github.com/spf13/cobra"
	shcd_parser "stash.us.cray.com/HMS/hms-shcd-parser/pkg/shcd-parser"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load <path>",
	Short: "Load an existing Shasta configuration",
	Long: `Load a set of files that represent a Shasta system.
	Often, 'load' is used with 'init', which generates the files.`,
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
	loadCmd.DisableAutoGenTag = true

}

func loadSystemConfig(path string) (sysconf csi.SystemConfig, err error) {
	err = csiFiles.ReadYAMLConfig(path, &sysconf)
	return
}

func loadNetwork(path string) (network csi.IPV4Network, err error) {
	err = csiFiles.ReadJSONConfig(path, &network)
	return
}

func extractNetworks(basepath string) ([]csi.IPV4Network, error) {
	// TODO: Handle incoming error?
	var networks []csi.IPV4Network
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
