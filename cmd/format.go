package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/
import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

var writeScript = filepath.Join(viper.GetString("write_script"))

func writeLiveCD(device string, iso string, size string) {
	// format the device as the liveCD
	cmd := exec.Command(writeScript, device, iso, size)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdoutBuf.String(), stderrBuf.String()
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)

	fmt.Printf("Run these commands before using 'pit populate':\n")
	fmt.Printf("\tmkdir -pv /mnt/{cow,pitdata}\n")
	fmt.Printf("\tmount -L cow /mnt/cow && mount -L PITDATA /mnt/pitdata\n")
}

func init() {
	pitCmd.AddCommand(formatCmd)
	formatCmd.DisableAutoGenTag = true
	viper.SetEnvPrefix("pit") // will be uppercased automatically
	viper.AutomaticEnv()
	formatCmd.MarkFlagRequired("write-script")
	formatCmd.Flags().StringVarP(&writeScript, "write-script", "w", "/usr/local/bin/write-livecd.sh", "Path to the write-livecd.sh script")
}
