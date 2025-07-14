---
date: 2025-07-10T15:00:19-05:00
title: "csi patch pit ca"
layout: default
---
## csi patch pit ca

Patch CA certificates into the PIT's cloud-init meta-data.

### Synopsis


Patches the Pre-Install Toolkit's (PIT) cloud-init meta-data, adding Certificate Authority (CA) certificates from a
given Shasta configuration (shasta-cfg).


```
csi patch pit ca [flags]
```

### Options

```
      --customizations-file string      path to customizations.yaml (shasta-cfg)
  -h, --help                            help for ca
      --sealed-secret-key-file string   Path to sealed secrets/shasta-cfg private key
      --sealed-secret-name string       Path to cloud-init metadata seed file (default "gen_platform_ca_1")
```

### Options inherited from parent commands

```
      --cloud-init-seed-file string   Path to cloud-init metadata seed file
  -c, --config string                 Path to a CSI config file (default is $PWD/system_config.yaml).
      --csm-api-url string            (for use against a completed CSM installation) The URL to a CSM API. (default "https://api-gw-service-nmn.local")
  -i, --input-dir string              A directory to read input files from (--config will take precedence, but only for system_config.yaml).
      --k8s-namespace string          (for use against a completed CSM installation) The namespace that the --k8s-secret-name belongs to. (default "default")
      --k8s-secret-name string        (for use against a completed CSM installation) The name of the Kubernetes secret to look for an OpenID credential in for CSM APIs (a.k.a. TOKEN=). (default "admin-client-auth")
```

### SEE ALSO

* [csi patch pit](/commands/csi_patch_pit/)	 - Patch aspects of the Pre-Install Toolkit (PIT) environment.

