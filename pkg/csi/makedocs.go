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

package csi

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Cray-HPE/cray-site-init/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const fmTemplate = `---
date: %s
title: "%s"
layout: default
---
`

// DocsCommand represents the makedocs subcommand.
func DocsCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "makedocs [directory]",
		DisableAutoGenTag: true,
		Short:             "Create a set of markdown files for the docs in the [directory] (docs/ is the default)",
		Run: func(c *cobra.Command, args []string) {
			var destinationDirectory string
			if len(args) < 1 {
				destinationDirectory = "docs/" // This is the default without passing an argument
			} else {
				destinationDirectory = args[0]
			}
			basepath, _ := filepath.Abs(filepath.Clean(destinationDirectory))
			_, err := os.Stat(basepath)
			if err != nil {
				// Assert that the error is actually a PathError or bail
				_, ok := err.(*os.PathError)
				if !ok {
					log.Fatalf(
						"Error accessing %v :%v",
						basepath,
						err,
					)
				}
			}

			filePrepender := func(filename string) string {
				now := cli.Runtime.Format(time.RFC3339)
				name := filepath.Base(filename)
				base := strings.TrimSuffix(
					name,
					path.Ext(name),
				)
				// url := "/commands/" + strings.ToLower(base) + "/"
				return fmt.Sprintf(
					fmTemplate,
					now,
					strings.Replace(
						base,
						"_",
						" ",
						-1,
					),
				)
			}

			linkHandler := func(name string) string {
				base := strings.TrimSuffix(
					name,
					path.Ext(name),
				)
				return "/commands/" + strings.ToLower(base) + "/"
			}

			err = os.Mkdir(
				basepath,
				0777,
			)
			if err != nil {
				log.Fatal(err)
			}
			err = doc.GenMarkdownTreeCustom(
				NewCommand(),
				basepath,
				filePrepender,
				linkHandler,
			)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	return c
}
