---
date: 2024-06-06T09:39:24-05:00
title: "csi handoff"
layout: default
---
## csi handoff

Runs migration steps to transition from LiveCD

### Synopsis

A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the production version inside the Kubernetes cluster.

### Options

```
  -h, --help   help for handoff
```

### Options inherited from parent commands

```
  -c, --config string   CSI config file
```

### SEE ALSO

* [csi](/commands/csi/)	 - Cray Site Init. For new sites, re-installs, and upgrades.
* [csi handoff bss-metadata](/commands/csi_handoff_bss-metadata/)	 - runs migration steps to build BSS entries for all NCNs
* [csi handoff bss-update-cloud-init](/commands/csi_handoff_bss-update-cloud-init/)	 - runs migration steps to update cloud-init parameters for NCNs
* [csi handoff bss-update-param](/commands/csi_handoff_bss-update-param/)	 - runs migration steps to update kernel parameters for NCNs

