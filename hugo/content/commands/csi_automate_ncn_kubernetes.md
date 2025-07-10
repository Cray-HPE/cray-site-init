---
date: 2025-07-10T15:00:19-05:00
title: "csi automate ncn kubernetes"
layout: default
---
## csi automate ncn kubernetes

tools used to automate actions to Kubernetes

### Synopsis

A series of subcommands that automates administrative tasks to Kubernetes.

```
csi automate ncn kubernetes [flags]
```

### Options

```
      --action string   The etcd action to perform (cordon-ncn/uncordon-ncn/drain-ncn/delete-ncn/is-member)
  -h, --help            help for kubernetes
      --ncn string      The NCN to perform the action on
```

### Options inherited from parent commands

```
  -c, --config string            Path to a CSI config file (default is $PWD/system_config.yaml).
      --csm-api-url string       (for use against a completed CSM installation) The URL to a CSM API. (default "https://api-gw-service-nmn.local")
  -i, --input-dir string         A directory to read input files from (--config will take precedence, but only for system_config.yaml).
      --k8s-namespace string     (for use against a completed CSM installation) The namespace that the --k8s-secret-name belongs to. (default "default")
      --k8s-secret-name string   (for use against a completed CSM installation) The name of the Kubernetes secret to look for an OpenID credential in for CSM APIs (a.k.a. TOKEN=). (default "admin-client-auth")
  -k, --kubeconfig string        Absolute path to the kubeconfig file
```

### SEE ALSO

* [csi automate ncn](/commands/csi_automate_ncn/)	 - tools used to automate NCN activities

