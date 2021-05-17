---
layout: page
title: Getting Started with `csi` (Cray Site Initializer)
<!-- permalink: /getting_started -->
---

`csi` is a command line utility that can generate config files needed for a 1.4+ Shasta installation and help setup an installable USB drive.

# Generate Shasta Config Files

`csi` can generate config files for a Shasta installation.  To begin, at least three seed files must exist to run the command.

1. `ncn_metadata.csv`
2. `switch_metdata.csv`
3. `hmn_connections.json`

Once those three files are in your current directory, you can run `csi config init`.  There are many flags to choose from, but here is an example command:

```bash
si config init \
    --bootstrap-ncn-bmc-user root \
    --bootstrap-ncn-bmc-pass changeme \
    --system-name eniac \
    --mountain-cabinets 4 \
    --starting-mountain-cabinet 1000 \
    --hill-cabinets 0 \
    --river-cabinets 1 \
    --can-cidr 10.103.11.0/24 \
    --can-external-dns 10.103.11.113 \
    --can-gateway 10.103.11.1 \
    --can-static-pool 10.103.11.112/28 \
    --can-dynamic-pool 10.103.11.128/25 \
    --nmn-cidr 10.252.0.0/17 \
    --hmn-cidr 10.254.0.0/17 \
    --ntp-pool time.nist.gov \
    --site-domain dev.cray.com \
    --site-ip 172.30.53.79/20 \
    --site-gw 172.30.48.1 \
    --site-nic p1p2 \
    --site-dns 172.30.84.40 \
    --install-ncn-bond-members p1p1,p10p1 \
    --application-node-config-yaml application_node_config.yaml \
    --cabinets-yaml cabinets.yaml \
    --hmn-mtn-cidr 10.104.0.0/17 \
    --nmn-mtn-cidr 10.100.0.0/17
```

The first time this command is run, a `system_config.yaml` is generated and contains all the info needed to run subsequent calls of `csi config init` without the need to type out all of the flags--all of the necessary data is contained in `system_config.yaml`.

# Setting Up A LiveCD

Once the config files are generated, a LiveCD can be setup.
