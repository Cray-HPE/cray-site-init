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
	"os"
	"path/filepath"
	"regexp"
)

const s3ACL = "public-read"

// We need these to be constant so that later when we do the BSS handoff we know the right value.
const k8sPath = "k8s"
const cephPath = "ceph"

const kernelName = "kernel"
const initrdName = "initrd"
const squashFSName = "filesystem.squashfs"

var (
	k8sKernelPath   string
	k8sInitrdPath   string
	k8sSquashFSPath string

	cephKernelPath   string
	cephInitrdPath   string
	cephSquashFSPath string

	imagesS3BucketName string
)

// handoffCmd represents the handoff command
var handoffNCNImagesCmd = &cobra.Command{
	Use:   "ncn-images",
	Short: "runs migration steps to transition from LiveCD",
	Long: "A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the " +
		"production version inside the Kubernetes cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		setupS3(imagesS3BucketName)

		fmt.Println("Uploading NCN images into S3.")
		uploadNCNImagesS3()
	},
}

func init() {
	handoffCmd.AddCommand(handoffNCNImagesCmd)
	handoffNCNImagesCmd.DisableAutoGenTag = true

	home := homedir.HomeDir()
	handoffNCNImagesCmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"),
		"Absolute path to the kubeconfig file")
	if home == "" {
		_ = handoffNCNImagesCmd.MarkFlagRequired("kubeconfig")
	}

	handoffNCNImagesCmd.Flags().StringVar(&s3SecretName, "s3-secret", "sts-s3-credentials",
		"Secret to use for connecting to S3")
	handoffNCNImagesCmd.Flags().StringVar(&imagesS3BucketName, "s3-bucket", "ncn-images",
		"Bucket to create and upload NCN images to")

	handoffNCNImagesCmd.Flags().StringVar(&k8sKernelPath, "k8s-kernel-path", "",
		"Path to the kernel image to upload for K8s NCNs")
	handoffNCNImagesCmd.Flags().StringVar(&k8sInitrdPath, "k8s-initrd-path", "",
		"Path to the initrd image to upload for K8s NCNs")
	handoffNCNImagesCmd.Flags().StringVar(&k8sSquashFSPath, "k8s-squashfs-path", "",
		"Path to the squashfs image to upload for K8s NCNs")
	_ = handoffNCNImagesCmd.MarkFlagRequired("k8s-squashfs-path")

	handoffNCNImagesCmd.Flags().StringVar(&cephKernelPath, "ceph-kernel-path", "",
		"Path to the kernel image to upload for CEPH NCNs")
	handoffNCNImagesCmd.Flags().StringVar(&cephInitrdPath, "ceph-initrd-path", "",
		"Path to the initrd image to upload for CEPH NCNs")
	handoffNCNImagesCmd.Flags().StringVar(&cephSquashFSPath, "ceph-squashfs-path", "",
		"Path to the squashfs image to upload for CEPH NCNs")
	_ = handoffNCNImagesCmd.MarkFlagRequired("ceph-squashfs-path")
}

func uploadFile(filePath string, s3KeyName string) {
	fmt.Printf("Uploading file %s to S3 at s3://%s/%s...\n", filePath, s3Client.ConnInfo.Bucket, s3KeyName)

	file, err := os.Open(filePath)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	_, err = s3Client.UploadFileWithACL(s3KeyName, file, s3ACL)
	if err != nil {
		log.Panic(err)
	}
}

func uploadNCNImagesS3() {
	// Create public-read bucket.
	_, err := s3Client.CreateBucketWithACL(imagesS3BucketName, s3ACL)
	if err != nil {
		awsErr := err.(awserr.Error)

		if awsErr.Code() == s3.ErrCodeBucketAlreadyExists {
			log.Printf("Bucket %s already exists.\n", imagesS3BucketName)
		} else {
			log.Panic(err)
		}
	} else {
		fmt.Printf("Sucessfully created %s bucket.\n", imagesS3BucketName)
	}

	// Need to figure out the versions of these images.
	versionRegex := regexp.MustCompile(`.*-(.+[0-9])\.squashfs`)
	var k8sVersion string
	var cephVersion string

	k8sMatches := versionRegex.FindStringSubmatch(k8sSquashFSPath)
	if k8sMatches != nil {
		k8sVersion = k8sMatches[1]
	} else {
		log.Fatalf("Kubernetes image version invalid: %s", k8sSquashFSPath)
	}

	cephMatches := versionRegex.FindStringSubmatch(cephSquashFSPath)
	if cephMatches != nil {
		cephVersion = cephMatches[1]
	} else {
		log.Fatalf("Ceph image version invalid: %s", cephSquashFSPath)
	}

	// Upload the files.
	uploadFile(k8sKernelPath, fmt.Sprintf("%s/%s/%s", k8sPath, k8sVersion, kernelName))
	fmt.Println("Successfully uploaded K8s kernel.")
	uploadFile(k8sInitrdPath, fmt.Sprintf("%s/%s/%s", k8sPath, k8sVersion, initrdName))
	fmt.Println("Successfully uploaded K8s initrd.")
	uploadFile(k8sSquashFSPath, fmt.Sprintf("%s/%s/%s", k8sPath, k8sVersion, squashFSName))
	fmt.Println("Successfully uploaded K8s squash FS.")

	uploadFile(cephKernelPath, fmt.Sprintf("%s/%s/%s", cephPath, cephVersion, kernelName))
	fmt.Println("Successfully uploaded CEPH kernel.")
	uploadFile(cephInitrdPath, fmt.Sprintf("%s/%s/%s", cephPath, cephVersion, initrdName))
	fmt.Println("Successfully uploaded CEPH initrd.")
	uploadFile(cephSquashFSPath, fmt.Sprintf("%s/%s/%s", cephPath, cephVersion, squashFSName))
	fmt.Println("Successfully uploaded CEPH squash FS.")

	fmt.Printf("\n\nImage versions uploaded:\nKubernetes:\t%s\nCEPH:\t\t%s\n", k8sVersion, cephVersion)

	fmt.Printf("\n\nYou should run the following commands so the versions you just uploaded can be used in "+
		"other steps:\nexport KUBERNETES_VERSION=%s\nexport CEPH_VERSION=%s\n", k8sVersion, cephVersion)
}
