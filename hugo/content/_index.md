---
layout: default
title: Cray Site Init
params:
  description: "Cray Site Initializer documentation"
---

# Cray Site Initializer (`csi`)

#### `csi` creates, validates, installs, and upgrades a CRAY supercomputer or HPCaaS platform.

It supports initializing a set of configuration files from command line flags and/or Shasta 1.3 configuration files.

{{% notice info %}}
These instructions are only valid for the Shasta v1.4+ release with CSM \>=0.8.0 and CSI \>=1.5.0

For CSI API documentation relevant to all versions of CSI, see the [Commands section]({{ < ref "commands/index" >}})
{{% /notice %}}

# Site Survey {{% button href="https://github.com/Cray-HPE/cray-site-init/releases" icon="fas fa-download" %}}Get csi{{% /button %}}

The Site Survey CSM Questionnaire outlines information needed to begin an installation.  It details the information that should be gathered prior to running `csi config init`, which will create the configuration payload needed to begin an installation.

## Seed Files

`csi` relies on several "seed" files:

- `hmn_connections.json`: maps which switch ports on the LeafBMC switches are cabled to what the River node BMCs, PDUs, or other hardware[307-HMN-CONNECTIONS](#)
- `ncn_metadata.csv`: maps the MACs of the Management NCNs to their xnames and is used to initialize the management cluster ([301-NCN-METADATA-BMC](#) and [302-NCN-METADATA-BONDX](#))
- `switch_metadata.csv`: maps the switch xname, brand, type, and model for the management switches in the system ([305-SWITCH-METADATA](#))

These three files represent the minimum set of inputs `csi` needs to generate a new configuration payload.  

> Details for making these three files can be found in the Cray Systems Management (CSM) documentation.

To begin, put the three files in a directory:

```
# ls -l
total 56
-rw-r--r--   1 root  12790  5641 Jul  6 15:23 hmn_connections.json
-rw-r--r--   1 root  12790  1002 Jul  6 15:23 ncn_metadata.csv
-rw-r--r--   1 root  12790    96 Jul  6 15:23 switch_metadata.csv
```

## Initializing a config

Once the seed files are in place, begin crafting the flags needed by running `csi config init empty`. This will create
a `system_config.yaml` file to be filled out. Some keys in this file will have defaults, and some will be blank and require 
editing.

### Required flags

#### `system-name`

The name of the system.  This value is prepended to the `--site-domain` to create the FQDN for the system.  

{{% notice warning %}}
This parameter impacts the Subject Name of some certificates and **cannot be easily changed after CSM has been installed**.
{{% /notice %}}

#### `site-domain`

The site domain to be used in the FQDN.  This corresponds to the site
DNS zone that will be associated with the Shasta system being
installed.  **Site domain selection is also critical in creation of system certificates certificates during installation**.  Changing these certificates post-installation is not a supported activity.

{{% notice warning %}}
The combination of --system-name and --site-domain are used to generate a set of certificates that form the basis of TLS security across CSM.  Post-Install adjustments of these values in Shasta 1.4 is not supported.
{{% /notice %}}

#### `csm-version`

The target CSM major and minor version being installed (e.g. 1.7). This ensures CSI generates configuration designed for
the version of CSM being installed.

#### Additional required flags

For the remaining flags, see the target CSM version's docs in Docs-CSM https://cray-hpe.github.io/docs-csm/en-17/install/pre-installation/#3-create-system-configuration
