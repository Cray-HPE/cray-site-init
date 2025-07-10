---
date: 2025-07-10T15:00:19-05:00
title: "csi pit"
layout: default
---
## csi pit

Manipulate or Create a LiveCD (Pre-Install Toolkit)

### Synopsis


Interact with the Pre-Install Toolkit (LiveCD);
create, validate, or re-create a new or old USB stick with
the liveCD tool. Fetches artifacts for deployment.

```
csi pit [flags]
```

### Options

```
  -h, --help   help for pit
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

* [csi](/commands/csi/)	 - Cray Site Initializer (csi)
* [csi pit format](/commands/csi_pit_format/)	 - Formats a disk as a LiveCD
* [csi pit validate](/commands/csi_pit_validate/)	 - Runs unit tests

