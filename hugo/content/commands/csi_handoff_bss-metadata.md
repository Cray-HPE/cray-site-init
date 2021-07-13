---
date: 2021-07-07T16:41:32-05:00
title: "csi handoff bss-metadata"
layout: default
---
## csi handoff bss-metadata

runs migration steps to build BSS entries for all NCNs

### Synopsis

Using PIT configuration builds kernel command line arguments and cloud-init metadata for each NCN

```
csi handoff bss-metadata [flags]
```

### Options

```
      --data-file string   data.json file with cloud-init configuration for each node and global
  -h, --help               help for bss-metadata
```

### SEE ALSO

* [csi handoff](/commands/csi_handoff/)	 - runs migration steps to transition from LiveCD

