---
date: 2024-06-06T09:39:24-05:00
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
  -c, --config string   CSI config file
```

### SEE ALSO

* [csi pit](/commands/csi_pit/)	 - Manipulate or Create a LiveCD (Pre-Install Toolkit)

