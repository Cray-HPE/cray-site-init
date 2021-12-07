/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package cmd

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"log"
	"os/exec"
	"path/filepath"
)

var (
	utilsS3BucketName string
)

// handoffCmd represents the handoff command
var handoffUploadUtils = &cobra.Command{
	Use:   "upload-utils",
	Short: "uploads utilities to S3",
	Long:  "Uploads utilities to S3 so they may be used throughout the cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		setupS3(utilsS3BucketName)

		fmt.Println("Uploading utilities.")
		uploadCSI()
	},
}

func init() {
	handoffCmd.AddCommand(handoffUploadUtils)
	handoffUploadUtils.DisableAutoGenTag = true

	home := homedir.HomeDir()
	handoffUploadUtils.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"),
		"Absolute path to the kubeconfig file")
	if home == "" {
		_ = handoffUploadUtils.MarkFlagRequired("kubeconfig")
	}

	handoffUploadUtils.Flags().StringVar(&utilsS3BucketName, "s3-bucket", "ncn-utils",
		"Bucket to create and upload NCN utils to")
}

func uploadCSI() {
	// Create public-read bucket.
	_, err := s3Client.CreateBucketWithACL(utilsS3BucketName, s3ACL)
	if err != nil {
		awsErr := err.(awserr.Error)

		if awsErr.Code() == s3.ErrCodeBucketAlreadyExists {
			log.Printf("Bucket %s already exists.\n", utilsS3BucketName)
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Sucessfully created %s bucket.\n", utilsS3BucketName)
	}

	csiPath, err := exec.LookPath("csi")
	if err != nil {
		log.Fatal(err)
	}

	uploadFile(csiPath, "csi")
	fmt.Println("Successfully uploaded CSI.")
}
