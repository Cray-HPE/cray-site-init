package cmd

/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"
	"io"
	"os"
	"path"
	// "os/exec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

var dataDir = "/mnt/data/"
var cephDir = "/mnt/data/ceph/"
var k8sDir = "/mnt/data/k8s/"

var kernelURL string
var initrdURL string
var k8sURL string
var cephURL string

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Downloads boot artifacts needed for PXE",
	Long: `Downloads kernel, initrd, and squashfs images needed
	for booting nodes via iPXE`,
	Run: func(cmd *cobra.Command, args []string) {
		makeReqDirs()
		GetArtifact(dataDir, viper.GetString("kernel"))
		GetArtifact(dataDir, viper.GetString("initrd"))
		GetArtifact(cephDir, viper.GetString("storage"))
		GetArtifact(k8sDir, viper.GetString("manager"))
	},
}

func makeReqDirs() {
	// Make dirs needed for images
	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(cephDir, 0755)
	os.MkdirAll(k8sDir, 0755)
}

// GetArtifact downloads kernels, initrd, and squashfs images
func GetArtifact(dirpath string, url string) (err error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	filename := resp.Request.URL.String()
	var fullPath = dirpath + path.Base(filename)

	fmt.Println("Saving...", fullPath)
	// Create the file
	out, err := os.Create(fullPath)
	if err != nil {
		return err
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

	return nil
}

func init() {
	spitCmd.AddCommand(getCmd)
	viper.SetEnvPrefix("spit") // will be uppercased automatically
	viper.AutomaticEnv()
	getCmd.Flags().StringVarP(&kernelURL, "kernel", "k", viper.GetString("kernel_url"), "URL of kernel file to download (env: SPIT_KERNEL_URL)")
	getCmd.Flags().StringVarP(&initrdURL, "initrd", "i", viper.GetString("initrd_url"), "URL of initrd file to download( env: SPIT_INITRD_URL)")
	getCmd.Flags().StringVarP(&k8sURL, "manager", "m", viper.GetString("manager_url"), "URL of manager/worker squashfs file to download (env: SPIT_MANAGER_URL)")
	getCmd.Flags().StringVarP(&cephURL, "storage", "s", viper.GetString("storage_url"), "URL of storage squashfs file to download (env: SPIT_STORAGE_URL)")

}
