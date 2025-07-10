---
date: 2025-07-10T15:00:19-05:00
title: "csi handoff bss-update-cloud-init"
layout: default
---
## csi handoff bss-update-cloud-init

runs migration steps to update cloud-init parameters for NCNs

### Synopsis

Allows for the updating of cloud-init settings in BSS for all the NCNs

```
csi handoff bss-update-cloud-init [flags]
```

### Options

```
      --delete stringArray   For each cloud-init object you wish to remove provide just the key and it will be removed regardless of value
  -h, --help                 help for bss-update-cloud-init
      --limit stringArray    Limit updates to just the xnames specified
      --set stringArray      For each cloud-init object you wish to update or add list it in the format of key=value
      --user-data string     json-formatted file with cloud-init user-data
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

