// +build integration
// +build shcd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/Cray-HPE/cray-site-init/pkg/csi"
)

// The user should populate this directory with the shcd.xlsx files for the systems they want to test
const _testdataShcds = "../testdata/shcds/"

var canus = []struct {
	systemName string
	version    string
	tabs       string
	corners    string
	csmVersion string
}{
	{
		systemName: "baldar",
		version:    "V1",
		tabs:       "25G_10G,NMN,HMN",
		corners:    "I14,S65,J16,T23,J20,U46",
		csmVersion: "1.0",
	},
	{
		systemName: "drax",
		version:    "V1",
		tabs:       "HMN",
		corners:    "J20,U36",
		csmVersion: "1.0",
	},
	{
		systemName: "ego",
		version:    "V1",
		tabs:       "NMN,HMN",
		corners:    "I13,S18,J20,U36",
		csmVersion: "1.0",
	},
	{
		systemName: "fanta",
		version:    "V1",
		tabs:       "HMN",
		corners:    "J20,U43",
		csmVersion: "1.2",
	},
	{
		systemName: "hela",
		version:    "Full",
		tabs:       "MTN,10G_25G_40G_100G,NMN,HMN",
		corners:    "K15,U36,I37,T107,J15,T20,J20,U38",
		csmVersion: "1.2",
	},
	{
		systemName: "odin",
		version:    "Full",
		tabs:       "10G_25G_40G_100G,NMN,HMN,MTN_TDS",
		corners:    "I37,T125,J15,T24,J20,U51,K15,U36",
		csmVersion: "1.2",
	},
	{
		systemName: "redbull",
		version:    "V1",
		tabs:       "HMN",
		corners:    "J20,U31",
		csmVersion: "1.2",
	},
	// {
	// 	systemName: "rocket",
	// 	version:    "V1",
	// 	tabs:       "40G_10G,NMN,HMN",
	// 	corners:    "I12,T38,I9,S24,J20,V36",
	//  csmVersion: "1.0",
	// },
	{
		systemName: "thanos",
		version:    "V1",
		tabs:       "HMN",
		corners:    "J20,U36",
		csmVersion: "1.0",
	},
	{
		systemName: "sif",
		version:    "V1",
		tabs:       "HMN",
		corners:    "J20,U35",
		csmVersion: "1.0",
	},
}

// CanuCommand runs arbitrary canu commands
func CanuCommand(args ...string) {
	// run canu
	cmd := exec.Command("canu", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}

// WalkShcds walks the testdata/shcds directory and checks for shcd.xlsx files matching a regex of the system name
func WalkShcds(root string, re *regexp.Regexp) ([]string, error) {
	var files []string
	// walk the testdata/shcds directory and look for any .xlsx files that match the system name
	// these will be used to generate an shcd.json, which will then be used in the subsequent tests
	err := filepath.Walk(_testdataShcds, func(path string, info os.FileInfo, err error) error {

		// if it not a directory and it is an excel file and it matches the regex
		if !info.IsDir() && filepath.Ext(path) == ".xlsx" && re.MatchString(path) {
			// add the file to the list of shcd files that will be used
			files = append(files, path)
		}
		return err
	})

	if err != nil {
		return files, err
	}

	return files, err
}

// TestConfigureShcds runs canu validate against the shcd.xlsx and then generates csi seed files
func TestConfigShcd_GenerateSeeds(t *testing.T) {

	var systemDir string
	var shcds []string

	// for each test case
	for _, test := range canus {

		// make a dir to hold the generated configs
		systemDir = filepath.Join(_testdataShcds, test.systemName)
		os.Mkdir(systemDir, 0755)

		// case-insensitive regex to use for matching system names to the filenames
		fnameRegex := "(?i)(.*)(" + test.systemName + ")(.*)"
		re := regexp.MustCompile(fnameRegex)

		// walk the testdata/shcds directory and look for any .xlsx files that match the system name
		files, err := WalkShcds(_testdataShcds, re)

		for f := range files {
			shcds = append(shcds, files[f])
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	// For each file found in the testdata/shcds directory
	for s := range shcds {
		// Start a loop of all the test cases to see if we can find a matching system name
		for _, test := range canus {

			// case-insensitive regex to use for matching system names to the filenames
			fnameRegex := "(?i)(.*)(" + test.systemName + ")(.*)"
			re := regexp.MustCompile(fnameRegex)

			// If there is a match, run 'canu' and 'csi config shcd'
			if re.MatchString(shcds[s]) {
				// Get the mac addresses from the leaf-bmc switch of the system
				log.Println("Running canu for " + test.systemName + "...")

				// Run canu against the shcd.xlsx file to generate the shcd.json
				CanuCommand("validate",
					"shcd",
					"-a", test.version,
					"--shcd", shcds[s],
					"--tabs", test.tabs,
					"--corners", test.corners,
					"--json",
					"--out", filepath.Join(_testdataShcds, test.systemName, test.systemName+".json"))

				// Generate the seed files from the newly-crafted shcd.json
				csi.ExecuteCommandC(rootCmd, []string{"config", "shcd",
					filepath.Join(_testdataShcds, test.systemName, test.systemName+".json"),
					"-SHAN",
					"-j", _schemaFile})

				// Move the files into the folder for the system so everything is together at the end of the test
				filesToMove := []string{hmnConnections, switchMetadata, applicationNodeConfig, ncnMetadata}

				for _, f := range filesToMove {
					err := os.Rename(f, filepath.Join(_testdataShcds, test.systemName, f))

					if err != nil {
						log.Fatal(err)
					}

				}
				break
			} else {
				continue
			}
		}
	}
}
