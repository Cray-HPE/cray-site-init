/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	hms_s3 "stash.us.cray.com/HMS/hms-s3"
)

// loftsmanCmd represents the loftsman command
var uploadSLSFile = &cobra.Command{
	Use:   "upload-sls-file",
	Short: "Upload the sls_input_file.json file into the SLS S3 Bucket",
	Long: `Upload the given sls_input_file.json file into the SLS S3 Bucket.
	Example: csi upload-sls-file --sls-file /path/to/sls_input_file.json

	Upload the given sls_input_file.json file into the SLS S3 Bucket, and delete the key "uploaded" if present in the SLS S3 bucket.
	The "upload" flag is added by the SLS loader when it successfully loads SLS with data, and it is used as flag to determine if 
	the SLS loader has been used previously to upload the SLS file. If it is present the loader will not upload the SLS file.

	Example: csi upload-sls-file --sls-file /path/to/sls_input_file.json --remove-upload-flag
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize the global viper
		v := viper.GetViper()
		uploadSLSInputFile(v)
	},
}

func init() {
	rootCmd.AddCommand(uploadSLSFile)

	home := homedir.HomeDir()
	uploadSLSFile.Flags().String("kubeconfig", filepath.Join(home, ".kube", "config"),
		"Absolute path to the kubeconfig file")
	if home == "" {
		_ = uploadSLSFile.MarkFlagRequired("kubeconfig")
	}

	uploadSLSFile.Flags().String("s3-secret", "sls-s3-credentials",
		"Secret to use for connecting to S3")
	uploadSLSFile.Flags().String("s3-bucket", "sls",
		"Bucket to create and upload the SLS input file to")

	uploadSLSFile.Flags().String("sls-file", "sls_input_file.json",
		"Path to the SLS Input File to Upload")
	uploadSLSFile.Flags().Bool("remove-upload-flag", false,
		"Remove the upload flag added by the SLS loader")
}

func uploadFileWithoutACL(myS3Client *hms_s3.S3Client, filePath, s3KeyName string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to open SLS file:", err)
	}
	defer file.Close()

	_, err = myS3Client.PutFile(s3KeyName, file)
	if err != nil {
		log.Fatal("Unable to update SLS file to S3:", err)
	}
}

func deleteFile(myS3Client *hms_s3.S3Client, s3KeyName string) {
	_, err := myS3Client.DeleteObject(s3KeyName)
	if err != nil {
		log.Fatal("Unable to delete file from S3:", err)
	}
}

func uploadSLSInputFile(v *viper.Viper) {
	// Built kubeconfig.
	config, err := clientcmd.BuildConfigFromFlags("", v.GetString("kubeconfig"))
	if err != nil {
		log.Fatal("Unable to build kubernetes config:", err)
	}

	// Create the clientset.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("Unable to setup kubernetes client:", err)
	}

	// Get the S3 credentails from Kubernetes.
	s3SecretName := v.GetString("s3-secret")
	log.Println("Retrieving S3 credentails (", s3SecretName, ") for SLS")
	s3Secret, err := clientset.CoreV1().Secrets("services").Get(context.TODO(), s3SecretName, v1.GetOptions{})
	if err != nil {
		log.Fatal("Unable to SLS S3 secret from k8s:", err)
	}

	// Normally the HMS S3 library uses environment variables but since the vast majority are just arguments to this
	// program manually create the object for connection info.
	s3Connection := hms_s3.ConnectionInfo{
		AccessKey: string(s3Secret.Data["access_key"]),
		SecretKey: string(s3Secret.Data["secret_key"]),
		Endpoint:  string(s3Secret.Data["s3_endpoint"]),
		Bucket:    v.GetString("s3-bucket"),
		Region:    "default",
	}

	if err := s3Connection.Validate(); err != nil {
		log.Fatal("S3 connection info validation failed:", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	myS3Client, err := hms_s3.NewS3Client(s3Connection, httpClient)
	if err != nil {
		log.Fatal("Failed to setup S3 Client:", err)
	}

	// Note: There is no need to create the SLS bucket, as it is automatically created when the Storage Nodes are stood up

	// Remove the Upload Flag if present from the SLS Bucket
	if v.GetBool("remove-upload-flag") {
		log.Println("Deleting upload flag (if present)")
		deleteFile(myS3Client, "uploaded") // This key should never change
	}

	// Upload the SLS file
	slsFilePath := v.GetString("sls-file")
	log.Printf("Uploading SLS file: %s\n", slsFilePath)
	uploadFileWithoutACL(myS3Client, slsFilePath, "sls_input_file.json") // This key name should never change

	log.Println("Successfully uploaded SLS Input File.")
}
