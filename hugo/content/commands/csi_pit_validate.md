---
date: 2025-07-10T15:00:19-05:00
title: "csi pit validate"
layout: default
---
## csi pit validate

Runs unit tests

### Synopsis

Runs unit tests and validates a working livecd and NCN deployment.

```
csi pit validate [flags]
```

### Options

```
  -C, --ceph                  Validate that Ceph is working
  -h, --help                  help for validate
  -k, --k8s                   Validate that Kubernetes is working
  -l, --livecd-preflight      Run LiveCD pre-flight tests
  -p, --livecd-provisioning   Run LiveCD provisioning tests
  -n, --ncn-preflight         Run NCN pre-flight tests
      --postgres              Validate that Postgres clusters are healthy
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

