---
date: 2025-07-10T15:00:19-05:00
title: "csi patch pit"
layout: default
---
## csi patch pit

Patch aspects of the Pre-Install Toolkit (PIT) environment.

### Synopsis


Patch commands targeting the Pre-Install Toolkit (PIT) environment's generated data files or services


### Options

```
      --cloud-init-seed-file string   Path to cloud-init metadata seed file
  -h, --help                          help for pit
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

* [csi patch](/commands/csi_patch/)	 - Patch commands for modifying system contexts.
* [csi patch pit ca](/commands/csi_patch_pit_ca/)	 - Patch CA certificates into the PIT's cloud-init meta-data.
* [csi patch pit packages](/commands/csi_patch_pit_packages/)	 - Patch packages and repositories into the PIT's cloud-init meta-data.

