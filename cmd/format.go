package cmd

/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// formatCmd represents the format command
var formatCmd = &cobra.Command{
	Use:   "format DISK ISO SIZE",
	Short: "Formats a disk as a LiveCD",
	Long:  `Formats a disk as a LiveCD using an ISO.`,
	// ValidArgs: []string{"disk", "iso", "size"},
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		writeLiveCD(args[0], args[1], args[2])
	},
}

var isoURL = viper.GetString("iso_url")

var isoName = viper.GetString("iso_name")

var toolkit = viper.GetString("repo_url")

var pitdata = viper.GetString("disk_label")

func writeLiveCD(device string, iso string, size string) {
	// git clone https://stash.us.cray.com/scm/mtl/shasta-pre-install-toolkit.git

	// get current directory
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	var writeScript = filepath.Join(path, "preinstall-toolkit/scripts/write-livecd.sh")

	// ./shasta-pre-install-toolkit/scripts/write-livecd.sh /dev/sdd $(pwd)/shasta-pre-install-toolkit-latest.iso 20000
	// format the device as the livecd
	cmd := exec.Command(writeScript, device, iso, size)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", stdoutStderr)

	// mount /dev/disk/by-label/PITDATA /mnt/
	fmt.Printf("Mount %s to /mnt before continuing\n", pitdata)
}

func init() {
	spitCmd.AddCommand(formatCmd)
	viper.SetEnvPrefix("spit") // will be uppercased automatically
	viper.AutomaticEnv()
	formatCmd.Flags().StringVarP(&isoURL, "iso-url", "u", viper.GetString("iso_url"), "URL the SPIT ISO to download (env: SPIT_ISO_URL)")
	formatCmd.Flags().StringVarP(&isoName, "iso-name", "n", viper.GetString("iso_name"), "Local filename of the iso to download (env: SPIT_ISO_NAME)")
	formatCmd.Flags().StringVarP(&toolkit, "repo-url", "r", viper.GetString("repo_url"), "URL of the git repo for the preinstall toolkit (env: SPIT_REPO_URL)")
	formatCmd.Flags().StringVarP(&pitdata, "disk-label", "l", viper.GetString("disk_label"), "URL of the git repo for the preinstall toolkit (env: SPIT_DISK_LABEL)")
	formatCmd.Flags().BoolP("force", "f", false, "Force overwrite the disk without warning")
}
