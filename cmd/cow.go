/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

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

package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// cowCmd represents the pitcow command
var cowCmd = &cobra.Command{
	Use:   "cow MOUNTPOINT /PATH/TO/CSI-GENERATED/FILES",
	Short: "Populates the cow partition with necessary config files",
	Long: `Populates the cow partition with necessary config files.

	This is what enables networking on boot by copying over the ifcfg files.`,
	// Arg is the path to the csi generated files
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		os.MkdirAll(filepath.Join(args[0], "rw/etc/sysconfig/network/"), 0755)
		os.MkdirAll(filepath.Join(args[0], "rw/etc/dnsmasq.d/"), 0755)

		// copy all csi-generated files to their correct place
		// Since we have persistence on the livecd, we can pre-create any directories we want,
		// populate them with files, and then they will be there when they boot.
		// this is how networking on boot is achieved as well as having dnsmasq configs already in place

		// pit-files are all interface config files, so they go to /etc/sysconfig/network
		copyAllFiles(filepath.Join(args[1], "pit-files/"), filepath.Join(args[0], "rw/etc/sysconfig/network/"))
		// dnsmasq.d are all dnsmasq configs, so they can go in /etc/dnsmasq.d
		copyAllFiles(filepath.Join(args[1], "dnsmasq.d/"), filepath.Join(args[0], "rw/etc/dnsmasq.d/"))

		// conman config enables the service to work at first boot when using a usb
		conmanSrc := filepath.Join(args[1], "conman.conf")
		conmanDest := filepath.Join(args[0], "rw/etc/conman.conf")
		fmt.Printf("%s> %s", PadRight(filepath.Base(conmanSrc), "-", 30), conmanDest)
		copyErr := copyFile(conmanSrc, conmanDest)
		if copyErr != nil {
			fmt.Printf("...Failed %q\n", copyErr)
		} else {
			fmt.Printf("...OK\n")
		}
	},
}

func init() {
	populateCmd.AddCommand(cowCmd)
	cowCmd.DisableAutoGenTag = true
}
