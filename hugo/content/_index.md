---
layout: default
title: Cray Site Init
---
# Cray Site Initializer (`csi`)

#### `csi` creates, validates, installs, and upgrades a CRAY supercomputer or HPCaaS platform.

It supports initializing a set of configuration files from command line flags and/or Shasta 1.3 configuration files.  

{{% notice info %}}
These instructions are only valid for the Shasta v1.4+ release with CSM \>=0.8.0 and CSI \>=1.5.0
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

## Required Flags for `csi config init`

Once the seed files are in place, begin crafting the flags needed for running `csi config init`

### `--system-name`

The name of the system.  This value is prepended to the `--site-domain` to create the FQDN for the system.  

{{% notice warning %}}
This parameter impacts the Subject Name of some certificates and **cannot be easily changed after CSM has been installed**.
{{% /notice %}}

### `--site-domain`

The site domain to be used in the FQDN.  This corresponds to the site
DNS zone that will be associated with the Shasta system being
installed.  **Site domain selection is also critical in creation of system certificates certificates during installation**.  Changing these certificates post-installation is not a supported activity.

{{% notice warning %}}
The combination of --system-name and --site-domain are used to generate a set of certificates that form the basis of TLS security across CSM.  Post-Install adjustments of these values in Shasta 1.4 is not supported.
{{% /notice %}}

#### Customer Access Network (CAN) IP Addresses

The Customer Access Network is a set of site-provided IP addresses that
CSM uses to grant access to services. CSM manages these IP addresses and
provides name resolution to services within the CAN. Some services,
namely DNS services, can be specified ahead of time so that site DNS
servers can be configured with an upstream IP address for the zone, even
before the system arrives on site. Keep in mind that site networking
will need to route these IP addresses according to the diagram above.
There are a set of ip addresses on the CAN that will be static and
therefore, once set by the application, user intervention is needed to
change them. The DNS Server is an example of a static reservation. CSM
developers strive to keep the number of ip addresses needed in the
static pool to a minimum while using the dynamic pool for most services.
IP addresses within the CAN CIDR, but outside the two named pools are
used for direct host ip addresses where necessary.

Both pools are used by CSM to provide resilient, load-balanced access to
all services via a software load balancer called
[MetalLB](https://metallb.universe.tf/).

### `--can-cidr`

A CIDR block that defines a site-routable IP range that is allocated to
the CAN.  Shasta will assign IPs within the CIDR block to endpoints that
are connected to the CAN.  Systems without the CAN, air-gap for
instance, still need to define the CAN gateway, CAN CIDR block, as well
as CAN static and dynamic pools, but the network need not transit
outside the system.

### `--can-gateway`

The CAN uses a site-routable CIDR block.  The CAN Gateway parameter is
used to configure which IP should be used as the gateway for the CAN
routes.

> example: 10.103.2.1

### `--can-dynamic-pool`

A CIDR block that defines a site IP range that is set aside for load
balanced services that may be dynamically allocated.  This is used for
ephemeral services like UAIs that request a dynamic IP on boot.  This
CIDR block must be a subset of "CAN CIDR Block".

> example: 10.102.11.128/25

### `--can-static-pool`

A CIDR block that defines a site IP range that is set aside for load
balanced services that must be statically defined rather than
dynamically allocated.  This CIDR block must be a subset of "CAN CIDR
Block".

> example: 10.102.11.112/28

### `--cmn-external-dns`

The CMN IP address to reserved for the CSM DNS endpoint which must be
part of the CMN Static Pool.  This is generally referred to as the
site-to-system lookup target.  The DNS zone \<system-name>.\<site-domain>
will be served by a DNS server at this IP address.

#### Site Parameters

The following parameters are related to the site and generally should
match the expected values for the production environment except where
specifically noted below.

### `--ipv4-resolvers`

The site DNS resolvers that are used by the PIT as part of the
installation to resolve site services and upstream internet addresses
(where applicable). This setting is generally the same as --site-dns,
but remains a separate configuration item to support installation in a
network that is different than the final production network.

> example: 8.8.8.8,9.9.9.9

### `--site-ip`

The Site IP/CIDR that will be used to by the Pre-Install Toolkit (PIT) for initial setup
before the CAN is configured

You will use this connection to SSH into the PIT once it has booted.

> example: 10.103.15.194/30

### `--site-gw`

The IP of the site gateway to be used with the Site IP for routing
during installation and recovery operations.

> example: 10.103.15.193

### `--site-dns`

The IP of the site DNS server(s) that Shasta should forward to via the
Site IP.  This is usually the same as `--ipv4-resolvers` when
installations are performed within the final production network.

> example: 8.8.8.8,9.9.9.9

### `--bootstrap-ncn-bmc-user`

The username that should be used when authenticating to the NCN BMCs.
This should be set to the username provided by manufacturing.  Changing
BMC credentials is not safe during an install, but is supported through
HSM APIs via the admin guide.

### `--bootstrap-ncn-bmc-pass`

The password that should be used when authenticating to the NCN BMCs.
This should be set to the password provided by manufacturing.  Changing
BMC credentials is not safe during an install, but is supported through
HSM APIs via the admin guide.

### `--ncn-metadata`

If `ncn_metadata.csv` is in the current directory (this is recommended), this flag is not needed (since it defaults to `./ncn_metadata.csv`), but you can pass it a different filepath if needed.

### `--switch-metadata`

If `switch_metadata.csv` is in the current directory (this is recommended), this flag is not needed (since it defaults to `./switch_metadata.csv`), but you can pass it a different filepath if needed.

### `--hmn-connections`

If `hmn_connections.json` is in the current directory (this is recommended), this flag is not needed (since it defaults to `./hmn_connections.json`), but you can pass it a different filepath if needed.

## Optional Flags for `csi config init`

### `--application-node-config-yaml`

For additional control of the application node identification during the
SLS Input File generation, an additional config file is necessary and
must be indicated with the `--application-node-config-yaml` flag. Allows
for the control of the following within the generated SLS Input File:

1.  System specific prefix for Applications Nodes.
2.  Specify HSM Subroles for system specific application nodes
3.  Specify Application Node Aliases

**It is recommended** to pre-specify the Application node alias using
this file, otherwise they will need to be manually added after CSM is
installed using
the [306-SLS-ADD-UAN-ALIAS](#)
procedure in the CSM documentation.

For additional information on constructing the
`application_node_config.yaml` file please refer
to [308-APPLICATION-NODE-CONFIG](#)
in the CSM v1.4 documentation.

### `--cabinets-yaml`

For systems that use non-sequential cabinet id numbers, an additional
mapping file is necessary and must be indicated with the `--cabinets-yaml`
flag.

For systems with Mountain or Hill cabinets, this configuration file can
be used to specify the NMN and HMN VLANs for the cabinets. On a Mountain
or Hill cabinet the CEC controls the VLAN settings for the NMN and HMN
networks on the cabinet.

For additional information on construction the `cabinets.yaml` file please
refer
to [310-CABINETS](#)
in the CSM v1.4 documentation.

# Changing information after an install is completed

Most of the information on this page can be adjusted after installation with appropriate procedures and APIs.  The **one major exception is the machine name and site domain because they are part of the certificate chain** for so many other things.  Detailed procedures for making these changes is out of scope for the site-survey, but can be found at the [accompanying page](#) and in the Administration Guide.  
