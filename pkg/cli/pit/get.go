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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"
)

func getCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "get DEST URL",
		DisableAutoGenTag: true,
		Short:             "Downloads artifacts",
		Long: `Downloads artifacts such as the release tarball
	or kernel, initrd, or squashfs images`,
		Args: cobra.ExactArgs(2),
		Run: func(c *cobra.Command, args []string) {
			GetArtifact(
				args[0],
				args[1],
			)
		},
	}
	return c
}

// GetArtifact downloads kernels, initrd, and squashfs images
func GetArtifact(
	dirpath string, url string,
) (err error) {
	if url != "" {
		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		filename := resp.Request.URL.String()
		var fullPath = dirpath + "/" + path.Base(filename)

		// Downloading non-existent files gives a bunch of html, which csi doesn't need
		if resp.Header.Get("Content-Type") == "text/html;charset=UTF-8" {

		} else {

			// Create the file
			out, err := os.Create(fullPath)
			if err != nil {
				fmt.Println(err)
			}
			defer out.Close()

			// Check server response
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf(
					"bad status: %s",
					resp.Status,
				)
			}

			// Writer the body to file
			_, err = io.Copy(
				out,
				resp.Body,
			)
			if err != nil {
				return err
			}
		}

		defer resp.Body.Close()

	} else {
		return fmt.Errorf(
			"missing or malformed URL: %v",
			url,
		)
	}

	return nil
}
