//
//  MIT License
//
//  (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.

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
		fmt.Printf("Successfully created %s bucket.\n", utilsS3BucketName)
	}

	csiPath, err := exec.LookPath("csi")
	if err != nil {
		log.Fatal(err)
	}

	uploadFile(csiPath, "csi")
	fmt.Println("Successfully uploaded CSI.")
}
