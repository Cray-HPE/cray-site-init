---
date: 2021-07-07T16:41:32-05:00
title: "csi handoff bss-update-cloud-init"
layout: default
---
## csi handoff bss-update-cloud-init

runs migration steps to update cloud-init parameters for NCNs

### Synopsis

Allows for the updating of cloud-init settings in BSS for all the NCNs

```
csi handoff bss-update-cloud-init [flags]
```

### Options

```
      --delete stringArray   For each kernel parameter you wish to remove provide just the key and it will be removed regardless of value
  -h, --help                 help for bss-update-cloud-init
      --limit stringArray    Limit updates to just the xnames specified
      --set stringArray      For each kernel parameter you wish to update or add list it in the format of key=value
      --user-data string     json-formatted file with global cloud-init user-data
```

### SEE ALSO

* [csi handoff](/commands/csi_handoff/)	 - runs migration steps to transition from LiveCD

