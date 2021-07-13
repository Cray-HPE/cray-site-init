---
date: 2021-07-07T16:41:32-05:00
title: "csi handoff bss-update-param"
layout: default
---
## csi handoff bss-update-param

runs migration steps to update kernel parameters for NCNs

### Synopsis

Allows for the updating of kernel parameters in BSS for all the NCNs

```
csi handoff bss-update-param [flags]
```

### Options

```
      --delete stringArray   For each kernel parameter you wish to remove provide just the key and it will be removed regardless of value
  -h, --help                 help for bss-update-param
      --initrd string        New value to set for the initrd
      --kernel string        New value to set for the kernel
      --limit stringArray    Limit updates to just the xnames specified
      --set stringArray      For each kernel parameter you wish to update or add list it in the format of key=value
```

### SEE ALSO

* [csi handoff](/commands/csi_handoff/)	 - runs migration steps to transition from LiveCD

