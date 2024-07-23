/*
 MIT License

 (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
*/

package pit

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// WriteLiveCDFunc is a type for the writeLiveCD function to allow for dependency injection in testing
type WriteLiveCDFunc func(device string, iso string, size string) error

func formatCommand(formatFunc WriteLiveCDFunc) *cobra.Command {
	c := &cobra.Command{
		Use:   "format DISK ISO SIZE",
		Short: "Formats a disk as a LiveCD",
		Long:  `Formats a disk as a LiveCD using an ISO.`, // ValidArgs: []string{"disk", "iso", "size"},
		Args:  cobra.ExactArgs(3),
		RunE: func(c *cobra.Command, args []string) error {
			device := args[0]
			iso := args[1]
			size := args[2]
			fmt.Printf("Arguments received: device=%s, iso=%s, size=%s\n", device, iso, size) // Debugging statement
			err := formatFunc(device, iso, size)
			return err
		},
	}
	viper.SetEnvPrefix("pit") // will be uppercased automatically
	viper.AutomaticEnv()
	c.Flags().StringVarP(
		&writeScript,
		"write-script",
		"w",
		"/usr/local/bin/write-livecd.sh",
		"Path to the write-livecd.sh script",
	)
	return c
}

var writeScript = filepath.Join(viper.GetString("write_script"))

var writeLiveCD = WriteLiveCDFunc(func(device string, iso string, size string) error {
	// format the device as the liveCD
	cmd := exec.Command(
		writeScript,
		device,
		iso,
		size,
	)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(
		os.Stdout,
		&stdoutBuf,
	)
	cmd.Stderr = io.MultiWriter(
		os.Stderr,
		&stderrBuf,
	)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf(
			"cmd.Run() failed with %s",
			err,
		)
	}
	outStr, errStr := stdoutBuf.String(), stderrBuf.String()
	fmt.Printf(
		"\nout:\n%s\nerr:\n%s\n",
		outStr,
		errStr,
	)

	fmt.Printf("Run these commands before using 'pit populate':\n")
	fmt.Printf("\tmkdir -pv /mnt/{cow,pitdata}\n")
	fmt.Printf("\tmount -L cow /mnt/cow && mount -L PITDATA /mnt/pitdata\n")
	return nil
})
