package cmd

/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

var dataDir, cephDir, k8sDir, kernelURL, initrdURL, k8sURL, cephURL string

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Downloads boot artifacts needed for PXE",
	Long: `Downloads kernel, initrd, and squashfs images needed
	for booting nodes via iPXE`,
	Run: func(cmd *cobra.Command, args []string) {
		makeReqDirs()
		GetArtifact(dataDir, kernelURL)
		GetArtifact(dataDir, initrdURL)
		GetArtifact(cephDir, cephURL)
		GetArtifact(k8sDir, k8sURL)
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
	if url != "" {
		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != 200 {
			fmt.Println(resp.StatusCode)
		}
		defer resp.Body.Close()

		filename := resp.Request.URL.String()
		var fullPath = dirpath + "/" + path.Base(filename)

		fmt.Printf("%d Saving...%s\n", resp.StatusCode, fullPath)
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

	} else {
		fmt.Println("Skipping empty URL")
	}

	return nil
}

func init() {
	pitCmd.AddCommand(getCmd)
	viper.SetEnvPrefix("pit") // will be uppercased automatically
	viper.AutomaticEnv()
	getCmd.Flags().StringVarP(&kernelURL, "kernel", "k", viper.GetString("kernel_url"), "URL of kernel file to download (env: PIT_KERNEL_URL)")
	getCmd.Flags().StringVarP(&initrdURL, "initrd", "i", viper.GetString("initrd_url"), "URL of initrd file to download( env: PIT_INITRD_URL)")
	getCmd.Flags().StringVarP(&k8sURL, "manager", "m", viper.GetString("manager_url"), "URL of manager/worker squashfs file to download (env: PIT_MANAGER_URL)")
	getCmd.Flags().StringVarP(&cephURL, "storage", "s", viper.GetString("storage_url"), "URL of storage squashfs file to download (env: PIT_STORAGE_URL)")
	getCmd.Flags().StringVarP(&dataDir, "data-dir", "d", viper.GetString("data_dir"), "Path to data folder where kernel, initrd, will be saved (env: PIT_DATA_DIR)")
	getCmd.Flags().StringVarP(&cephDir, "ceph-dir", "c", viper.GetString("ceph_dir"), "Path to where ceph image will be saved (env: PIT_CEPH_DIR)")
	getCmd.Flags().StringVarP(&k8sDir, "k8s-dir", "K", viper.GetString("k8s_dir"), "Path to where k8s image will be saved  (env: PIT_K8S_DIR)")

}
