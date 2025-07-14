---
date: 2025-07-10T15:00:19-05:00
title: "csi handoff"
layout: default
---
## csi handoff

Runs migration steps to transition from LiveCD

### Synopsis

A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the production version inside the Kubernetes cluster.

### Options

```
  -h, --help   help for handoff
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
* [csi handoff bss-metadata](/commands/csi_handoff_bss-metadata/)	 - runs migration steps to build BSS entries for all NCNs
* [csi handoff bss-update-cloud-init](/commands/csi_handoff_bss-update-cloud-init/)	 - runs migration steps to update cloud-init parameters for NCNs
* [csi handoff bss-update-param](/commands/csi_handoff_bss-update-param/)	 - runs migration steps to update kernel parameters for NCNs

