---
date: 2025-07-10T15:00:19-05:00
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
  -c, --config string            Path to a CSI config file (default is $PWD/system_config.yaml).
      --csm-api-url string       (for use against a completed CSM installation) The URL to a CSM API. (default "https://api-gw-service-nmn.local")
  -i, --input-dir string         A directory to read input files from (--config will take precedence, but only for system_config.yaml).
      --k8s-namespace string     (for use against a completed CSM installation) The namespace that the --k8s-secret-name belongs to. (default "default")
      --k8s-secret-name string   (for use against a completed CSM installation) The name of the Kubernetes secret to look for an OpenID credential in for CSM APIs (a.k.a. TOKEN=). (default "admin-client-auth")
  -k, --kubeconfig string        Absolute path to the kubeconfig file
```

### SEE ALSO

* [csi automate ncn](/commands/csi_automate_ncn/)	 - tools used to automate NCN activities

