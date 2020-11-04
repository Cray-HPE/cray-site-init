# CRAY Site Initalizer

csi is a tool for managing install, upgrade, and disaster recovery of a Cray System Management cluster.

## Managing the Pile o' Files

One of the hallmarks of Shasta's complexity has been the number of configuration files required for the installation and the level of duplication within them.  With the 1.3 release, redundancies were reduced, but the number of files remained fairly high.  In 1.4, the number of files needed drops significantly.  Managing this reduction in files without adding more manual labor for the administrators is key to success with installing and upgrading a Shasta system with version 1.4.  The `csi` tool has subcommands for reading configuration files from a 1.3 system and writing them in the 1.4 format.  Additionally, `csi` can generate a tree of simple configuration files suitable for human editing that can then be verified by `csi` before the installation/upgrade can begin.

* `csi config init` can be run in a directory with appropriate files for a Shasta 1.3 system to create a Shasta 1.4 configuration.
* `csi config initSLS` is similar to `config init` except that it can use an SLS to generate the Shasta 1.4 configuration structure
* `csi config newSystem` will generate a brand new configuration for a new Shasta 1.4 system
* `csi config verify` will read a Shasta 1.4 configuration and verify that all required fields are filled in and compatible with each other
* `csi config export` will read a Shasta 1.4 configuration and output a single yaml configuration file for use by the installer

cray sls dumpstate list --format=json feeds back sls-common:SLSState

### TODO: Use SLSState and ipmitool from the LiveCD to "discover" the mac addresses of the NCNs

ipmitool can get us mac for lan1 on each NCN

### csi config dump

Print sections of the configuration object or the whole thing

### Artifacts

#### RPM

The RPM installs CRAY Site Init. `csi` into a SLES 15 Linux environment.

This RPM is installed into every LiveCD (Cray Pre-Install Toolkit).

```bash
redbull-ncn-w001-pit:/var/www/ephemeral/data # csi --help
csi is a tool for creating and validating the configuration of a Shasta system.

	It supports initializing a set of configuration from a variety of inputs including
	flags and/or Shasta 1.3 configuration files.  It can also validate that a set of
	configuration details are accurate before attempting to use them for installation

Usage:
  csi [flags]
  csi [command]

Available Commands:
  config      Interact with a config in a named directory
  help        Help about any command
  install     Install Cray System Management
  loftsman    A brief description of your command
  verify      A brief description of your command

Flags:
  -h, --help   help for csi

Use "csi [command] --help" for more information about a command.
```


### Reminder for using GOPRIVATE

`GOPRIVATE=*.us.cray.com go mod tidy`

<https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules>
