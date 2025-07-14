---
date: 2025-07-10T15:00:19-05:00
title: "csi pit format"
layout: default
---
## csi pit format

Formats a disk as a LiveCD

### Synopsis

Formats a disk as a LiveCD using an ISO.

```
csi pit format DISK ISO SIZE [flags]
```

### Options

```
  -h, --help                  help for format
  -w, --write-script string   Path to the write-livecd.sh script (default "/usr/local/bin/write-livecd.sh")
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

* [csi pit](/commands/csi_pit/)	 - Manipulate or Create a LiveCD (Pre-Install Toolkit)

