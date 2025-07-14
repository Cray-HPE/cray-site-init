---
date: 2025-07-10T15:00:19-05:00
title: "csi config template cloud-init disks"
layout: default
---
## csi config template cloud-init disks

Process cloud-init disk templates

### Synopsis

Process cloud-iniit meta-data for disks. Includes bootcmd, fs_setup, and mounts cloud-init user-data

```
csi config template cloud-init disks [flags]
```

### Options

```
  -h, --help   help for disks
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

* [csi config template cloud-init](/commands/csi_config_template_cloud-init/)	 - Process cloud-init templates

