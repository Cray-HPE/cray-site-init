---
date: 2025-07-10T15:00:19-05:00
title: "csi config gen-sls"
layout: default
---
## csi config gen-sls

Generates SLS input file

### Synopsis

Generates SLS input file based on a Shasta configuration and
	HMN connections file. By default, cabinets are assumed to be one River, the
	rest Mountain.

```
csi config gen-sls [options] <path> [flags]
```

### Options

```
  -h, --help                   help for gen-sls
      --hill-cabinets int      Number of River cabinets
      --river-cabinets int16   Number of River cabinets (default 1)
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

* [csi config](/commands/csi_config/)	 - HPC configuration

