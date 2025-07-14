---
date: 2025-07-10T15:00:19-05:00
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
  -c, --config string            Path to a CSI config file (default is $PWD/system_config.yaml).
      --csm-api-url string       (for use against a completed CSM installation) The URL to a CSM API. (default "https://api-gw-service-nmn.local")
  -i, --input-dir string         A directory to read input files from (--config will take precedence, but only for system_config.yaml).
      --k8s-namespace string     (for use against a completed CSM installation) The namespace that the --k8s-secret-name belongs to. (default "default")
      --k8s-secret-name string   (for use against a completed CSM installation) The name of the Kubernetes secret to look for an OpenID credential in for CSM APIs (a.k.a. TOKEN=). (default "admin-client-auth")
```

### SEE ALSO

* [csi handoff](/commands/csi_handoff/)	 - Runs migration steps to transition from LiveCD

