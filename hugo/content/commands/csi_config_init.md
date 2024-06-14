---
date: 2024-06-06T09:39:24-05:00
title: "csi config init"
layout: default
---
## csi config init

Generates a Shasta configuration payload

### Synopsis

init generates a scaffolding the Shasta configuration payload. It is based on several input files:
	1. The hmn_connections.json which describes the cabling for the BMCs on the NCNs
	2. The ncn_metadata.csv file documents the MAC addresses of the NCNs to be used in this installation
	   NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
	3. The switch_metadata.csv file which documents the Xname, Brand, Type, and Model of each switch. Types are CDU, LeafBMC, Leaf, and Spine
	   Switch Xname,Type,Brand,Model

	** NB **
	For systems that use non-sequential cabinet id numbers, an additional mapping file is necessary and must be indicated
	with the --cabinets-yaml flag.
	** NB **

	** NB **
	For additional control of the application node identification during the SLS Input File generation, an additional config file is necessary
	and must be indicated with the --application-node-config-yaml flag.

	Allows control of the following in the SLS Input File:
	1. System specific prefix for Applications node
	2. Specify HSM Subroles for system-specific application nodes
	3. Specify Application node Aliases
	** NB **

	In addition, there are many flags to impact the layout of the system. The defaults are generally fine except for the networking flags.
	

```
csi config init [flags]
```

### Options

```
      --application-node-config-yaml string   YAML to control Application node identification during the SLS Input File generation
      --bgp-asn string                        The autonomous system number for BGP router (default "65533")
      --bgp-chn-asn string                    The autonomous system number for CHN BGP clients (default "65530")
      --bgp-cmn-asn string                    The autonomous system number for CMN BGP clients (default "65532")
      --bgp-nmn-asn string                    The autonomous system number for NMN BGP clients (default "65531")
      --bgp-peer-types strings                Comma-separated list of which set of switches to use as metallb peers: spine (default), leaf and/or edge (default [spine])
      --bican-user-network-name string        Name of the network over which non-admin users access the system [CAN, CHN, HSN]
      --bootstrap-ncn-bmc-pass string         Password for connecting to the BMC on the initial NCNs
      --bootstrap-ncn-bmc-user string         Username for connecting to the BMC on the initial NCNs
      --cabinets-yaml string                  YAML file listing the ids for all cabinets by type
      --can-bootstrap-vlan int                Bootstrap VLAN for the CAN (default 6)
      --can-cidr string                       Overall IPv4 CIDR for all Customer Access subnets
      --can-dynamic-pool string               Overall IPv4 CIDR for dynamic Customer Access load balancer addresses
      --can-gateway string                    Gateway for NCNs on the CAN (User)
      --can-static-pool string                Overall IPv4 CIDR for static Customer Access load balancer addresses
      --chn-cidr string                       Overall IPv4 CIDR for all Customer High-Speed subnets
      --chn-dynamic-pool string               Overall IPv4 CIDR for dynamic Customer High-Speed load balancer addresses
      --chn-gateway string                    Gateway for NCNs on the CHN (User)
      --chn-static-pool string                Overall IPv4 CIDR for static Customer High-Speed load balancer addresses
      --cmn-bootstrap-vlan int                Bootstrap VLAN for the CMN (default 7)
      --cmn-cidr string                       Overall IPv4 CIDR for all Customer Management subnets
      --cmn-dynamic-pool string               Overall IPv4 CIDR for dynamic Customer Management load balancer addresses
      --cmn-external-dns string               IP Address in the cmn-static-pool for the external dns service "site-to-system lookups"
      --cmn-gateway string                    Gateway for NCNs on the CMN (Administrative/Management)
      --cmn-static-pool string                Overall IPv4 CIDR for static Customer Management load balancer addresses
      --csm-version string                    Version of CSM being installed (e.g. <major>.<minor> such as "1.5" or "v1.5").
      --first-master-hostname string          Hostname of the first master node (default "ncn-m002")
  -h, --help                                  help for init
      --hill-cabinets int                     Number of Hill Cabinets
      --hmn-bootstrap-vlan int                Bootstrap VLAN for the HMN (default 4)
      --hmn-cidr string                       Overall IPv4 CIDR for all Hardware Management subnets (default "10.254.0.0/17")
      --hmn-connections string                HMN Connections JSON Location (For generating an SLS File) (default "hmn_connections.json")
      --hmn-dynamic-pool string               Overall IPv4 CIDR for dynamic Hardware Management load balancer addresses (default "10.94.100.0/24")
      --hmn-mtn-cidr string                   IPv4 CIDR for grouped Mountain Hardware Management subnets (default "10.104.0.0/17")
      --hmn-rvr-cidr string                   IPv4 CIDR for grouped River Hardware Management subnets (default "10.107.0.0/17")
      --hmn-static-pool string                Overall IPv4 CIDR for static Hardware Management load balancer addresses
      --hsn-cidr string                       Overall IPv4 CIDR for all HSN subnets (default "10.253.0.0/16")
      --hsn-dynamic-pool string               Overall IPv4 CIDR for dynamic High Speed load balancer addresses
      --hsn-static-pool string                Overall IPv4 CIDR for static High Speed load balancer addresses
      --install-ncn string                    Hostname of the node to be used for installation (default "ncn-m001")
      --install-ncn-bond-members string       List of devices to use to form a bond on the install ncn (default "p1p1,p1p2")
      --k8s-api-auditing-enabled              Enable the kubernetes auditing API
      --management-net-ips int                Additional number of IP addresses to reserve in each vlan for network equipment
      --mountain-cabinets int                 Number of Mountain Cabinets (default 4)
      --mtl-cidr string                       Overall IPv4 CIDR for all Provisioning subnets (default "10.1.1.0/16")
      --ncn-metadata string                   CSV for mapping the mac addresses of the NCNs to their xnames (default "ncn_metadata.csv")
      --ncn-mgmt-node-auditing-enabled        Enable management node auditing
      --nmn-bootstrap-vlan int                Bootstrap VLAN for the NMN (default 2)
      --nmn-cidr string                       Overall IPv4 CIDR for all Node Management subnets (default "10.252.0.0/17")
      --nmn-dynamic-pool string               Overall IPv4 CIDR for dynamic Node Management load balancer addresses (default "10.92.100.0/24")
      --nmn-mtn-cidr string                   IPv4 CIDR for grouped Mountain Node Management subnets (default "10.100.0.0/17")
      --nmn-rvr-cidr string                   IPv4 CIDR for grouped River Node Management subnets (default "10.106.0.0/17")
      --nmn-static-pool string                Overall IPv4 CIDR for static Node Management load balancer addresses
      --notify-zones string                   Comma-separated list of the zones to be allowed transfer
      --ntp-peers strings                     Comma-separated list of NCNs that will peer together (default [ncn-m001,ncn-m002,ncn-m003,ncn-w001,ncn-w002,ncn-w003,ncn-s001,ncn-s002,ncn-s003])
      --ntp-pools strings                     Comma-separated list of upstream NTP pool(s)
      --ntp-servers strings                   Comma-separated list of upstream NTP server(s); ncn-m001 should always be in this list (default [ncn-m001])
      --ntp-timezone string                   Timezone to be used on the NCNs and across the system (default "UTC")
      --primary-server-name string            Desired name for the primary DNS server (default "primary")
      --retain-unused-user-network            Use the supernet mask and gateway for NCNs and Switches
      --river-cabinets int                    Number of River Cabinets (default 1)
      --secondary-servers string              Comma-separated list of FQDN/IP for all DNS servers to notify when zone changes are made
      --site-dns string                       Site Network DNS Server
      --site-domain string                    Site Domain Name
      --site-gw string                        Site Network IPv4 Gateway
      --site-ip string                        Site Network Information in the form ipaddress/prefix like 192.168.1.1/24
      --site-nic string                       Network Interface on install-ncn that will be connected to the site network (default "em1")
      --starting-hill-cabinet int             Starting ID number for Hill Cabinets (default 9000)
      --starting-mountain-NID int             Starting NID for Compute Nodes (default 1000)
      --starting-mountain-cabinet int         Starting ID number for Mountain Cabinets (default 1000)
      --starting-river-NID int                Starting NID for Compute Nodes (default 1)
      --starting-river-cabinet int            Starting ID number for River Cabinets (default 3000)
      --supernet                              Use the supernet mask and gateway for NCNs and Switches (default true)
      --switch-metadata string                CSV for mapping the switch xname, brand, type, and model (default "switch_metadata.csv")
      --system-name string                    Name of the System (default "sn-2024")
```

### Options inherited from parent commands

```
  -c, --config string   CSI config file
```

### SEE ALSO

* [csi config](/commands/csi_config/)	 - Interact with a Shasta config
* [csi config init empty](/commands/csi_config_init_empty/)	 - Write a empty config file.

