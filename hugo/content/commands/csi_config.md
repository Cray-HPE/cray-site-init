---
date: 2025-07-10T15:00:19-05:00
title: "csi config"
layout: default
---
## csi config

HPC configuration

### Synopsis

Creates Cray-HPE site configuration files

### Options

```
  -h, --help   help for config
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

* [csi](/commands/csi/)	 - Cray Site Initializer (csi)
* [csi config dump](/commands/csi_config_dump/)	 - Dumps an existing config to STDOUT
* [csi config gen-sls](/commands/csi_config_gen-sls/)	 - Generates SLS input file
* [csi config init](/commands/csi_config_init/)	 - Generates a Shasta configuration payload
* [csi config shcd](/commands/csi_config_shcd/)	 - Generates hmn_connections.json, switch_metadata.csv, and application_node_config.yaml from an SHCD JSON file
* [csi config template](/commands/csi_config_template/)	 - Process file templates

