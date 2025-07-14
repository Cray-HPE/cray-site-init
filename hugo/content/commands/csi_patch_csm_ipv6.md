---
date: 2025-07-10T15:00:19-05:00
title: "csi patch csm ipv6"
layout: default
---
## csi patch csm ipv6

Retroactively adds IPv6 data to CSM.

### Synopsis


Patches a Cray System Management (CSM) deployment, modifying the System Layout Service (SLS) and the
Boot Script Service (BSS) for IPv6 enablement.

This command runs as a dry-run unless its --commit flag is present. As a dry-run, no changes are committed
to BSS or SLS, and all discovered changes are written to the local filesystem.

Backups of BSS and SLS will be created for inspection or rollback (using a backup to rollback to requires using the
CSM API, or its direct tools such as CrayCLI).

Only certain SLS subnets defined within the Customer Management Network and (if CSM has one configured) the
Customer High-Speed Network are targeted, by default the targeted subnets are:
- network_hardware
- bootstrap_dhcp

NOTE: Any IPv6 enablement already present in SLS and BSS, such as reserved IP addresses, will not be overwritten unless
the force flag is given.

```
csi patch csm ipv6 [flags]
```

### Options

```
  -b, --backup-dir string     The directory to write backup files to (defaults to a timestamped directory within the current working directory).
      --chn-cidr6 string      Overall IPv6 CIDR for all CHN subnets.
      --chn-gateway6 string   IPv6 Gateway for NCNs on the CHN.
      --cmn-cidr6 string      Overall IPv6 CIDR for all CMN subnets.
      --cmn-gateway6 string   IPv6 Gateway for NCNs on the CMN.
  -w, --commit                Write all proposed changes into CSM; commit changes to BSS and SLS.
  -f, --force                 Force updating IPv6 reservations, subnet CIDRs, and more.
  -h, --help                  help for ipv6
  -r, --remove                Remove all IPv6 enablement data from BSS and SLS.
      --subnets strings       Comma-separated list of SLS subnets to patch. (default [network_hardware,bootstrap_dhcp])
```

### Options inherited from parent commands

```
  -c, --config string            Path to a CSI config file (default is $PWD/system_config.yaml).
      --csm-api-url string       (for use against a completed CSM installation) The URL to a CSM API. (default "https://api-gw-service-nmn.local")
  -i, --input-dir string         A directory to read input files from (--config will take precedence, but only for system_config.yaml).
      --k8s-namespace string     (for use against a completed CSM installation) The namespace that the --k8s-secret-name belongs to. (default "default")
      --k8s-secret-name string   (for use against a completed CSM installation) The name of the Kubernetes secret to look for an OpenID credential in for CSM APIs (a.k.a. TOKEN=). (default "admin-client-auth")
```

### SEE ALSO

* [csi patch csm](/commands/csi_patch_csm/)	 - Patch aspects of Cray System Management (CSM).

