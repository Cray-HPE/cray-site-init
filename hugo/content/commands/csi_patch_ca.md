---
date: 2021-07-07T16:41:32-05:00
title: "csi patch ca"
layout: default
---
## csi patch ca

Patch cloud-init metadata with CA certs

### Synopsis


Patch cloud-init metadata (in place) with certificate authority (CA) certificates from
shasta-cfg (customizations.yaml). Decrypts CA material from named sealed secret using the shasta-cfg
private RSA key.

```
csi patch ca [flags]
```

### Options

```
      --cloud-init-seed-file string     Path to cloud-init metadata seed file
      --customizations-file string      path to customizations.yaml (shasta-cfg)
  -h, --help                            help for ca
      --sealed-secret-key-file string   Path to sealed secrets/shasta-cfg private key
      --sealed-secret-name string       Path to cloud-init metadata seed file (default "gen_platform_ca_1")
```

### SEE ALSO

* [csi patch](/commands/csi_patch/)	 - Apply patch operations

