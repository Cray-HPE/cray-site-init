# Shasta Instance Control

sic is a tool for managing install, upgrade, and disaster recovery of a Cray System Management cluster.

## Managing the Pile o' Files

One of the hallmarks of Shasta's complexity has been the number of configuration files required for the installation and the level of duplication within them.  With the 1.3 release, redundancies were reduced, but the number of files remained fairly high.  In 1.4, the number of files needed drops significantly.  Managing this reduction in files without adding more manual labor for the administrators is key to success with installing and upgrading a Shasta system with version 1.4.  The `sic` tool has subcommands for reading configuration files from a 1.3 system and writing them in the 1.4 format.  Additionally, `sic` can generate a tree of simple configuration files suitable for human editing that can then be verified by `sic` before the installation/upgrade can begin.

* `sic config init` can be run in a directory with appropriate files for a Shasta 1.3 system to create a Shasta 1.4 configuration.
* `sic config initSLS` is similar to `config init` except that it can use an SLS to generate the Shasta 1.4 configuration structure
* `sic config newSystem` will generate a brand new configuration for a new Shasta 1.4 system
* `sic config verify` will read a Shasta 1.4 configuration and verify that all required fields are filled in and compatible with each other
* `sic config export` will read a Shasta 1.4 configuration and output a single yaml configuration file for use by the installer

cray sls dumpstate list --format=json feeds back sls-common:SLSState

### TODO: Use SLSState and ipmitool from the LiveCD to "discover" the mac addresses of the NCNs

ipmitool can get us mac for lan1 on each NCN

### sic config dump

Print sections of the configuration object or the whole thing

### Artifacts

#### RPM

The RPM installs a system daemon for running the application container.
The RPM build does not run Go lint or Go unit tests.

### Docker Container

The container wraps build and test dependencies for repeatable builds.  All linting and tests run inside the container.
