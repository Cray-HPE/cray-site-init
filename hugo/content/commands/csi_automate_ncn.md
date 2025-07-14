---
date: 2025-07-10T15:00:19-05:00
title: "csi automate ncn"
layout: default
---
## csi automate ncn

tools used to automate NCN activities

### Synopsis

A series of subcommands that automates NCN administrative tasks.

### Options

```
  -h, --help                help for ncn
  -k, --kubeconfig string   Absolute path to the kubeconfig file
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

* [csi automate](/commands/csi_automate/)	 - Tools used to automate system lifecycle events
* [csi automate ncn etcd](/commands/csi_automate_ncn_etcd/)	 - tools used to automate actions to bare-metal etcd
* [csi automate ncn kubernetes](/commands/csi_automate_ncn_kubernetes/)	 - tools used to automate actions to Kubernetes
* [csi automate ncn preflight](/commands/csi_automate_ncn_preflight/)	 - tools used to automate preflight checks

