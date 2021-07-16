/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package csi

import (
	"bytes"

	"github.com/spf13/cobra"
)

// stringInSlice is shorthand
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// ExecuteCommandC runs a cobra command
func ExecuteCommandC(root *cobra.Command, args []string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}
