---
date: 2021-07-07T16:41:32-05:00
title: "csi config init"
layout: default
---
## csi config init

init generates the directory structure for a new system rooted in a directory matching the system-name argument

### Synopsis

init generates a scaffolding the Shasta configuration payload.  It is based on several input files:
	1. The hmn_connections.json which describes the cabling for the BMCs on the NCNs
	2. The ncn_metadata.csv file documents the MAC addresses of the NCNs to be used in this installation
	   NCN xname,NCN Role,NCN Subrole,BMC MAC,BMC Switch Port,NMN MAC,NMN Switch Port
	3. The switch_metadata.csv file which documents the Xname, Brand, Type, and Model of each switch.  Types are CDU, Leaf, Aggregation, and Spine
	   Switch Xname,Type,Brand,Model

	** NB **
	For systems that use non-sequential cabinet id numbers, an additional mapping file is necessary and must be indicated
	with the --cabinets-yaml flag.
	** NB **

	** NB **
	For additional control of the application node identification durring the SLS Input File generation, an additional config file is necessary
	and must be indicated with the --application-node-config-yaml flag.

	Allows control of the following in the SLS Input File:
	1. System specific prefix for Applications node
	2. Specify HSM Subroles for system specifc application nodes
	3. Specify Application node Aliases
	** NB **

	In addition, there are many flags to impact the layout of the system.  The defaults are generally fine except for the networking flags.
	

```
csi config init [flags]
```

### Options

```
      --system-name string                    Name of the System (default "sn-2024")
      --site-domain string                    Site Domain Name (default "dev.cray.com")
      --ntp-pools strings                     Comma-seperated list of upstream NTP pool(s)
      --ntp-servers strings                   Comma-seperated list of upstream NTP server(s) ncn-m001 should always be in this list (default [ncn-m001])
      --ntp-peers strings                     Comma-seperated list of NCNs that will peer together (default [ncn-m001,ncn-m002,ncn-m003,ncn-w001,ncn-w002,ncn-w003,ncn-s001,ncn-s002,ncn-s003])
      --ntp-timezone string                   Timezone to be used on the NCNs and across the system (default "UTC")
      --ipv4-resolvers string                 List of IP Addresses for DNS (default "8.8.8.8, 9.9.9.9")
      --v2-registry string                    URL for default v2 registry used for both helm and containers (default "https://registry.nmn/")
      --rpm-repository string                 URL for default rpm repository (default "https://packages.nmn/repository/shasta-master")
      --can-gateway string                    Gateway for NCNs on the CAN
      --cmn-gateway string                    Gateway for NCNs on the CMN
      --ceph-cephfs-image string              The container image for the cephfs provisioner (default "dtr.dev.cray.com/cray/cray-cephfs-provisioner:0.1.0-nautilus-1.3")
      --ceph-rbd-image string                 The container image for the ceph rbd provisioner (default "dtr.dev.cray.com/cray/cray-rbd-provisioner:0.1.0-nautilus-1.3")
      --docker-image-registry string          Upstream docker registry for use during the install (default "dtr.dev.cray.com")
      --install-ncn string                    Hostname of the node to be used for installation (default "ncn-m001")
      --install-ncn-bond-members string       List of devices to use to form a bond on the install ncn (default "p1p1,p1p2")
      --site-ip string                        Site Network Information in the form ipaddress/prefix like 192.168.1.1/24
      --site-gw string                        Site Network IPv4 Gateway
      --site-dns string                       Site Network DNS Server which can be different from the upstream ipv4-resolvers if necessary
      --site-nic string                       Network Interface on install-ncn that will be connected to the site network (default "em1")
      --nmn-cidr string                       Overall IPv4 CIDR for all Node Management subnets (default "10.252.0.0/17")
      --nmn-mtn-cidr string                   IPv4 CIDR for grouped Mountain Node Management subnets (default "10.100.0.0/17")
      --nmn-rvr-cidr string                   IPv4 CIDR for grouped River Node Management subnets (default "10.106.0.0/17")
      --hmn-cidr string                       Overall IPv4 CIDR for all Hardware Management subnets (default "10.254.0.0/17")
      --hmn-mtn-cidr string                   IPv4 CIDR for grouped Mountain Hardware Management subnets (default "10.104.0.0/17")
      --hmn-rvr-cidr string                   IPv4 CIDR for grouped River Hardware Management subnets (default "10.107.0.0/17")
      --can-cidr string                       Overall IPv4 CIDR for all Customer Access subnets (default "10.102.11.0/24")
      --can-static-pool string                Overall IPv4 CIDR for static Customer Access addresses (default "10.102.11.112/28")
      --can-dynamic-pool string               Overall IPv4 CIDR for dynamic Customer Access addresses (default "10.102.11.128/25")
      --cmn-cidr string                       Overall IPv4 CIDR for all Customer Management subnets (default "10.103.6.0/24")
      --cmn-static-pool string                Overall IPv4 CIDR for static Customer Management addresses (default "10.103.6.112/28")
      --cmn-dynamic-pool string               Overall IPv4 CIDR for dynamic Customer Management addresses (default "10.103.6.128/25")
      --cmn-external-dns string               IP Address in the cmn-static-pool for the external dns service "site-to-system lookups"
      --mtl-cidr string                       Overall IPv4 CIDR for all Provisioning subnets (default "10.1.1.0/16")
      --hsn-cidr string                       Overall IPv4 CIDR for all HSN subnets (default "10.253.0.0/16")
      --supernet                              Use the supernet mask and gateway for NCNs and Switches (default true)
      --nmn-bootstrap-vlan int                Bootstrap VLAN for the NMN (default 2)
      --hmn-bootstrap-vlan int                Bootstrap VLAN for the HMN (default 4)
      --can-bootstrap-vlan int                Bootstrap VLAN for the CAN (default 7)
      --macvlan-bootstrap-vlan int            Bootstrap VLAN for MacVlan (default 2)
      --mountain-cabinets int                 Number of Mountain Cabinets (default 4)
      --starting-mountain-cabinet int         Starting ID number for Mountain Cabinets (default 1000)
      --river-cabinets int                    Number of River Cabinets (default 1)
      --starting-river-cabinet int            Starting ID number for River Cabinets (default 3000)
      --hill-cabinets int                     Number of Hill Cabinets
      --starting-hill-cabinet int             Starting ID number for Hill Cabinets (default 9000)
      --starting-river-NID int                Starting NID for Compute Nodes (default 1)
      --starting-mountain-NID int             Starting NID for Compute Nodes (default 1000)
      --bgp-asn string                        The autonomous system number for BGP conversations (default "65533")
      --bgp-peers string                      Which set of switches to use as metallb peers, spine (default) or aggregation (default "spine")
      --management-net-ips int                Additional number of ip addresses to reserve in each vlan for network equipment
      --k8s-api-auditing-enabled              Enable the kubernetes auditing API
      --ncn-mgmt-node-auditing-enabled        Enable management node auditing
      --bootstrap-ncn-bmc-user string         Username for connecting to the BMC on the initial NCNs
      --bootstrap-ncn-bmc-pass string         Password for connecting to the BMC on the initial NCNs
      --hmn-connections string                HMN Connections JSON Location (For generating an SLS File) (default "hmn_connections.json")
      --ncn-metadata string                   CSV for mapping the mac addresses of the NCNs to their xnames (default "ncn_metadata.csv")
      --switch-metadata string                CSV for mapping the switch xname, brand, type, and model (default "switch_metadata.csv")
      --cabinets-yaml string                  YAML file listing the ids for all cabinets by type
      --application-node-config-yaml string   YAML to control Application node identification durring the SLS Input File generation
      --manifest-release string               Loftsman Manifest Release Version (leave blank to prevent manifest generation)
  -h, --help                                  help for init
```

### SEE ALSO

* [csi config](/commands/csi_config/)	 - Interact with a config in a named directory

