# CRAY Site Initalizer (CSI)

CSI is a tool for managing install, upgrade, and disaster recovery of a Cray System Management cluster.

**This deprecates CrayCTL** for shasta-1.4 as the primary orchestrator tool.

# Usage

CSI can be installed into any local Go-lang environment.

> Note: You will need to add CRAY to the [GOPRIVATE lib][1] for a clean run: `GOPRIVATE=*.us.cray.com go mod tidy`

You can build CSI two ways:
1. Using the `makefile`
    ```bash
    $> make
    $> ./bin/csi --help
    ```
2. Calling Go
    ```bash
    $> go build -o bin/csi ./main.go
    $> ./bin/csi --help
    ```

CSI is also built for distributing through Linux package managers.

#### OpenSuSE / SLES

```bash
# Add repo.
repo=http://car.dev.cray.com/artifactory/shasta-premium/MTL/sle15_sp2_ncn/x86_64/dev/master/
zypper addrepo --no-gpgcheck --refresh "$repo" metal_x86-64

# Install.
zypper install cray-site-init
``` 

## Managing the Pile o' Files

One of the hallmarks of Shasta's complexity has been the number of configuration files required for the installation and the level of duplication within them.  With the 1.3 release, redundancies were reduced, but the number of files remained fairly high.  In 1.4, the number of files needed drops significantly.  Managing this reduction in files without adding more manual labor for the administrators is key to success with installing and upgrading a Shasta system with version 1.4.  The `csi` tool has subcommands for reading configuration files from a 1.3 system and writing them in the 1.4 format.  Additionally, `csi` can generate a tree of simple configuration files suitable for human editing that can then be verified by `csi` before the installation/upgrade can begin.

## Commands

## Config

> `csi config dump`

Print sections of the configuration object or the whole thing

> `csi config init`

Can be run in a directory with appropriate files for a Shasta 1.3 system to create a Shasta 1.4 configuration.

> `csi config initSLS` 

Similar to `config init` except that it can use an SLS to generate the Shasta 1.4 configuration structure

> `csi config newSystem`

Generate a brand-new configuration for a new Shasta 1.4+ system

> `csi config verify` 

Reads a Shasta 1.4 configuration and verify all required fields and whether they're compatible with each other.

> `csi config export` 

Read a Shasta 1.4 configuration and output a single yaml configuration file for use by the installer.

## Pre-Install Toolkit `pit`

> `csi pit get`

Get the artifacts needed for deploying a CRAY.
- SquashFS Images for Storage and Kubernetes class nodes
- Boot artifacts (initrd and kernel)
- Specify where to store files with `-c` and `-d` for configs and data, respecitively

> `csi pit format`

Formats and creates a liveCD off an attached USB stick.

> `csi pit validate`

Verifies the liveCD after launch.

 

[1]: https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules