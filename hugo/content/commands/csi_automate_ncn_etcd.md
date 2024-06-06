---
date: 2024-06-06T09:39:24-05:00
title: "csi automate ncn etcd"
layout: default
---
## csi automate ncn etcd

tools used to automate actions to bare-metal etcd

### Synopsis

A series of subcommands that automates administrative tasks to the bare-metal etcd cluster.

```
csi automate ncn etcd [flags]
```

### Options

```
      --action string           The etcd action to perform (is-member/is-healthy)
      --endpoints stringArray   List of endpoints to connect to (default [ncn-m001.nmn:2379,ncn-m002.nmn:2379,ncn-m003.nmn:2379])
  -h, --help                    help for etcd
      --ncn string              The NCN to perform the action on
```

### Options inherited from parent commands

```
  -c, --config string   CSI config file
```

### SEE ALSO

* [csi automate ncn](/commands/csi_automate_ncn/)	 - tools used to automate NCN activities

