---
date: 2024-06-06T09:39:24-05:00
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
      --delete stringArray   For each cloud-init object you wish to remove provide just the key and it will be removed regardless of value
  -h, --help                 help for bss-update-cloud-init
      --limit stringArray    Limit updates to just the xnames specified
      --set stringArray      For each cloud-init object you wish to update or add list it in the format of key=value
      --user-data string     json-formatted file with cloud-init user-data
```

### Options inherited from parent commands

```
  -c, --config string   CSI config file
```

### SEE ALSO

* [csi handoff](/commands/csi_handoff/)	 - Runs migration steps to transition from LiveCD

