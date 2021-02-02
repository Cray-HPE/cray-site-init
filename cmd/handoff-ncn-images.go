/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	hms_s3 "stash.us.cray.com/HMS/hms-s3"
)

const s3ACL = "public-read"

var (
	s3Client *hms_s3.S3Client

	kubeconfig string

	s3SecretName string
	s3BucketName string

	kernelPath   string
	kernelName   string
	initrdPath   string
	initrdName   string
	squashFSPath string
	squashFSName string
)

// handoffCmd represents the handoff command
var handoffNCNImagesCmd = &cobra.Command{
	Use:   "ncn-images",
	Short: "runs migration steps to transition from LiveCD",
	Long: "A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the " +
		"production version inside the Kubernetes cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Uploading NCN images into S3.")
		uploadNCNImagesS3()
	},
}

func init() {
	handoffCmd.AddCommand(handoffNCNImagesCmd)

	home := homedir.HomeDir()
	handoffNCNImagesCmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"),
		"Absolute path to the kubeconfig file")
	if home == "" {
		_ = handoffCmd.MarkFlagRequired("kubeconfig")
	}

	handoffNCNImagesCmd.Flags().StringVar(&s3SecretName, "s3-secret", "sds-s3-credentials",
		"Secret to use for connecting to S3")
	handoffNCNImagesCmd.Flags().StringVar(&s3BucketName, "s3-bucket", "ncn-images",
		"Bucket to create and upload NCN images to")

	handoffNCNImagesCmd.Flags().StringVar(&kernelPath, "kernel-path", "/var/www/kernel",
		"Path to the kernel image to upload")
	handoffNCNImagesCmd.Flags().StringVar(&kernelName, "kernel-name", "kernel",
		"Name to set for kernel in S3")
	handoffNCNImagesCmd.Flags().StringVar(&initrdPath, "initrd-path", "/var/www/initrd.img.xz",
		"Path to the initrd image to upload")
	handoffNCNImagesCmd.Flags().StringVar(&initrdName, "initrd-name", "initrd",
		"Name to set for initrd in S3")
	handoffNCNImagesCmd.Flags().StringVar(&squashFSPath, "squashfs-path", "",
		"Path to the squashfs image to upload")
	_ = handoffNCNImagesCmd.MarkFlagRequired("squashfs-path")
	handoffNCNImagesCmd.Flags().StringVar(&squashFSName, "squashfs-name", "filesystem.squashfs",
		"Name to set for squashfs in S3")
}

func uploadFile(filePath string, s3KeyName string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	_, err = s3Client.PutFileWithACL(s3KeyName, file, s3ACL)
	if err != nil {
		log.Panic(err)
	}
}

func uploadNCNImagesS3() {
	// Built kubeconfig.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Panic(err)
	}

	// Create the clientset.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err)
	}

	// Get the secret from Kubernetes.
	s3Secret, err := clientset.CoreV1().Secrets("services").Get(context.TODO(), s3SecretName, v1.GetOptions{})
	if err != nil {
		log.Panic(err)
	}

	// Normally the HMS S3 library uses environment variables but since the vast majority are just arguments to this
	// program manually create the object for connection info.
	s3Connection := hms_s3.ConnectionInfo{
		AccessKey: string(s3Secret.Data["access_key"]),
		SecretKey: string(s3Secret.Data["secret_key"]),
		Endpoint:  string(s3Secret.Data["s3_endpoint"]),
		Bucket:    s3BucketName,
		Region:    "default",
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	s3Client, err = hms_s3.NewS3Client(s3Connection, httpClient)
	if err != nil {
		log.Panic(err)
	}

	// Create public-read bucket.
	_, err = s3Client.CreateBucketWithACL(s3BucketName, s3ACL)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Sucessfully created %s bucket.\n", s3BucketName)

	// Upload the files.
	uploadFile(kernelPath, kernelName)
	fmt.Println("Successfully uploaded kernel.")

	uploadFile(initrdPath, initrdName)
	fmt.Println("Successfully uploaded initrd.")

	uploadFile(squashFSPath, squashFSName)
	fmt.Println("Successfully uploaded squash FS.")
}
