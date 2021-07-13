---
date: 2021-07-07T16:41:32-05:00
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

### SEE ALSO

* [csi config](/commands/csi_config/)	 - Interact with a config in a named directory

