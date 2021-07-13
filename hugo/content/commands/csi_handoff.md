---
date: 2021-07-07T16:41:32-05:00
title: "csi handoff"
layout: default
---
## csi handoff

runs migration steps to transition from LiveCD

### Synopsis

A series of subcommands that facilitate the migration of assets/configuration/etc from the LiveCD to the production version inside the Kubernetes cluster.

### Options

```
  -h, --help   help for handoff
```

### SEE ALSO

* [csi](/commands/csi/)	 - Cray Site Init. for new sites ore re-installs and upgrades.
* [csi handoff bss-metadata](/commands/csi_handoff_bss-metadata/)	 - runs migration steps to build BSS entries for all NCNs
* [csi handoff bss-update-cloud-init](/commands/csi_handoff_bss-update-cloud-init/)	 - runs migration steps to update cloud-init parameters for NCNs
* [csi handoff bss-update-param](/commands/csi_handoff_bss-update-param/)	 - runs migration steps to update kernel parameters for NCNs
* [csi handoff ncn-images](/commands/csi_handoff_ncn-images/)	 - runs migration steps to transition from LiveCD

