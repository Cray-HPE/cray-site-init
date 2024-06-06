---
date: 2024-06-06T09:39:24-05:00
title: "csi patch packages"
layout: default
---
## csi patch packages

Patch cloud-init metadata with repositories and packages

### Synopsis


Patch cloud-init metadata (in place) with a list of repositories to add, and packages to install, during cloud-init from
CSM's cloud-init.yaml.

```
csi patch packages [flags]
```

### Options

```
      --config-file string   Path to cloud-init.yaml
  -h, --help                 help for packages
```

### Options inherited from parent commands

```
      --cloud-init-seed-file string   Path to cloud-init metadata seed file
  -c, --config string                 CSI config file
```

### SEE ALSO

* [csi patch](/commands/csi_patch/)	 - Apply patch operations

