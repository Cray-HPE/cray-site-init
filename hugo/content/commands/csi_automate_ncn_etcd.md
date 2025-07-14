---
date: 2025-07-10T15:00:19-05:00
title: "csi automate ncn etcd"
layout: default
---
## csi automate ncn etcd

tools used to automate actions to bare-metal etcd

### Synopsis

A series of subcommands that automates administrative tasks to the bare-metal etcd cluster.

```
csi automate ncn etcd [flags]
```

### Options

```
      --action string           The etcd action to perform (is-member/is-healthy)
      --endpoints stringArray   List of endpoints to connect to (default [ncn-m001.nmn:2379,ncn-m002.nmn:2379,ncn-m003.nmn:2379])
  -h, --help                    help for etcd
      --ncn string              The NCN to perform the action on
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

