---
date: 2024-06-06T09:39:24-05:00
title: "csi handoff bss-metadata"
layout: default
---
## csi handoff bss-metadata

runs migration steps to build BSS entries for all NCNs

### Synopsis

Using PIT configuration builds kernel command line arguments and cloud-init metadata for each NCN

```
csi handoff bss-metadata [flags]
```

### Options

```
      --data-file string                 data.json file with cloud-init configuration for each node and global
  -h, --help                             help for bss-metadata
      --kubernetes-ims-image-id string   The Kubernetes IMS_IMAGE_ID UUID value
      --storage-ims-image-id string      The storage-ceph IMS_IMAGE_ID UUID value
```

### Options inherited from parent commands

```
  -c, --config string   CSI config file
```

### SEE ALSO

* [csi handoff](/commands/csi_handoff/)	 - Runs migration steps to transition from LiveCD

