package cmd

/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
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

var writeScript = filepath.Join(viper.GetString("write_script"))

func writeLiveCD(device string, iso string, size string) {
	// git clone https://stash.us.cray.com/scm/mtl/cray-pre-install-toolkit.git

	// ./cray-pre-install-toolkit/scripts/write-livecd.sh /dev/sdd $(pwd)/cray-pre-install-toolkit-latest.iso 20000
	// format the device as the livecd
	cmd := exec.Command(writeScript, device, iso, size)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)

	// mount /dev/disk/by-label/PITDATA /mnt/
	fmt.Printf("Mount %s to /mnt before continuing\n", pitdata)
}

func init() {
	pitCmd.AddCommand(formatCmd)
	viper.SetEnvPrefix("pit") // will be uppercased automatically
	viper.AutomaticEnv()
	formatCmd.Flags().StringVarP(&isoURL, "iso-url", "u", viper.GetString("iso_url"), "URL the PIT ISO to download (env: PIT_ISO_URL)")
	formatCmd.Flags().StringVarP(&isoName, "iso-name", "n", viper.GetString("iso_name"), "Local filename of the iso to download (env: PIT_ISO_NAME)")
	formatCmd.MarkFlagRequired("write-script")
	formatCmd.Flags().StringVarP(&writeScript, "write-script", "w", viper.GetString("write_script"), "Path to the write_livecd.sh script (env: PIT_WRITE_SCRIPT)")
	formatCmd.Flags().StringVarP(&toolkit, "repo-url", "r", viper.GetString("repo_url"), "URL of the git repo for the preinstall toolkit (env: PIT_REPO_URL)")
	formatCmd.Flags().StringVarP(&pitdata, "disk-label", "l", viper.GetString("disk_label"), "URL of the git repo for the preinstall toolkit (env: PIT_DISK_LABEL)")
	formatCmd.Flags().BoolP("force", "f", false, "Force overwrite the disk without warning")
}
