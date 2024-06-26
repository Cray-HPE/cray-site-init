---
date: 2021-07-07T16:41:32-05:00
title: "csi handoff ncn-images"
layout: default
---
## csi handoff ncn-images

runs migration steps to transition from LiveCD

### Synopsis

A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the production version inside the Kubernetes cluster.

```
csi handoff ncn-images [flags]
```

### Options

```
      --ceph-initrd-path string     Path to the initrd image to upload for CEPH NCNs
      --ceph-kernel-path string     Path to the kernel image to upload for CEPH NCNs
      --ceph-squashfs-path string   Path to the squashfs image to upload for CEPH NCNs
  -h, --help                        help for ncn-images
      --k8s-initrd-path string      Path to the initrd image to upload for K8s NCNs
      --k8s-kernel-path string      Path to the kernel image to upload for K8s NCNs
      --k8s-squashfs-path string    Path to the squashfs image to upload for K8s NCNs
      --kubeconfig string           Absolute path to the kubeconfig file (default "/Users/jsalmela/.kube/config")
      --s3-bucket string            Bucket to create and upload NCN images to (default "ncn-images")
      --s3-secret string            Secret to use for connecting to S3 (default "sts-s3-credentials")
```

### SEE ALSO

* [csi handoff](/commands/csi_handoff/)	 - runs migration steps to transition from LiveCD

