---
date: 2024-06-06T09:39:24-05:00
title: "csi automate ncn preflight"
layout: default
---
## csi automate ncn preflight

tools used to automate preflight checks

### Synopsis

A series of subcommands that automates preflight checks around shutdown/reboot/rebuilt NCN lifecycle activities.

```
csi automate ncn preflight [flags]
```

### Options

```
      --action string       The etcd action to perform (verify-loss-acceptable/standardize-hostname)
  -h, --help                help for preflight
      --hostnames strings   Hostname(s) to standardize so it will work with Ansible
      --ncns strings        The NCNs to assume go away (at least temporarily) for the action
```

### Options inherited from parent commands

```
  -c, --config string   CSI config file
```

### SEE ALSO

* [csi automate ncn](/commands/csi_automate_ncn/)	 - tools used to automate NCN activities

