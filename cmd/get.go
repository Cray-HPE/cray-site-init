package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get DEST URL",
	Short: "Downloads artifacts",
	Long: `Downloads artifacts such as the release tarball
	or kernel, initrd, or squashfs images`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		GetArtifact(args[0], args[1])
	},
}

// GetArtifact downloads kernels, initrd, and squashfs images
func GetArtifact(dirpath string, url string) (err error) {
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
				return fmt.Errorf("bad status: %s", resp.Status)
			}

			// Writer the body to file
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				return err
			}
		}

		defer resp.Body.Close()

	} else {
		return fmt.Errorf("missing or malformed URL: %v", url)
	}

	return nil
}

func init() {
	pitCmd.AddCommand(getCmd)
	getCmd.DisableAutoGenTag = true
}
