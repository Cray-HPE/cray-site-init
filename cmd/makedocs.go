/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"time"
	"strings"
	"path/filepath"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// makedocsCmd represents the makedocs command
var makedocsCmd = &cobra.Command{
	Use:   "makedocs [directory]",
	Short: "Create a set of markdown files for the docs in the [directory] (docs/ is the default)",
	Run: func(cmd *cobra.Command, args []string) {
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
				log.Fatalf("Error accessing %v :%v", basepath, err)
			}
		}
		const fmTemplate = `---
date: %s
title: "%s"
layout: default
---
`

		filePrepender := func(filename string) string {
			now := time.Now().Format(time.RFC3339)
			name := filepath.Base(filename)
			base := strings.TrimSuffix(name, path.Ext(name))
			// url := "/commands/" + strings.ToLower(base) + "/"
			return fmt.Sprintf(fmTemplate, now, strings.Replace(base, "_", " ", -1))
		}

		linkHandler := func(name string) string {
			base := strings.TrimSuffix(name, path.Ext(name))
			return "/commands/" + strings.ToLower(base) + "/"
		}

		os.Mkdir(basepath, 0777)
		err = doc.GenMarkdownTreeCustom(rootCmd, basepath, filePrepender, linkHandler)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(makedocsCmd)
	makedocsCmd.DisableAutoGenTag = true
}
