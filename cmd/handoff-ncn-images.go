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
	"regexp"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	hms_s3 "stash.us.cray.com/HMS/hms-s3"
)

const s3ACL = "public-read"

// We need these to be constant so that later when we do the BSS handoff we know the right value.
const k8sPath = "k8s"
const cephPath = "ceph"

const kernelName = "kernel"
const initrdName = "initrd"
const squashFSName = "filesystem.squashfs"

var (
	s3Client *hms_s3.S3Client

	kubeconfig string

	s3SecretName string
	s3BucketName string

	k8sKernelPath   string
	k8sInitrdPath   string
	k8sSquashFSPath string

	cephKernelPath   string
	cephInitrdPath   string
	cephSquashFSPath string
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
