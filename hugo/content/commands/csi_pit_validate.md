---
date: 2021-07-07T16:41:32-05:00
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
  -c, --ceph                  Validate that Ceph is working
  -h, --help                  help for validate
  -k, --k8s                   Validate that Kubernetes is working
  -l, --livecd-preflight      Run LiveCD pre-flight tests
  -p, --livecd-provisioning   Run LiveCD provisioning tests
  -n, --ncn-preflight         Run NCN pre-flight tests
  -N, --network               Run network tests
  -S, --services              Run services tests
```

### SEE ALSO

* [csi pit](/commands/csi_pit/)	 - Manipulate or Create a LiveCD (Pre-Install Toolkit)

