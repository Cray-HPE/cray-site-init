# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [2.0.1] - 2022-01-21

### Changed

- repatriated HMSError types from hms-xname
- updated arti link to artifactory.algol60.net

## [2.0.0] - 2021-12-13

### Changed

- CASMHMS-5180: Moved HMSTypes and related functions to the new hms-xname repo under the xnametypes package.

## [1.15.1] - 2021-08-09

### Changed

- Added GitHub configuration files and fixed snyk warning.

## [1.15.0] - 2021-07-16

### Changed

- Replaced Stash Go module name with GitHub version.

## [1.14.0] - 2021-07-01

### Changed

- Bumped version to represent migration to GitHub.

## [1.13.0] - 2021-06-28

### Security

- CASMHMS-4898 - Updated base container images for security updates.

## [1.12.2] - 2021-05-03

### Changed

- Allow valid nodeAccel type xnames for more than 8 GPUs

## [1.12.1] - 2021-04-02

### Changed

- Updated Dockerfiles to pull base images from Artifactory instead of DTR.

## [1.12.0] - 2021-01-26

### Added

- Update Licence info in source files.

## [1.11.1] - 2021-01-19

### Added

- Added a function to add a User-Agent header to an http request.

## [1.11.0] - 2021-01-14

### Changed

- fix versions..

## [1.9.0] - 2021-01-14

### Changed

- Updated license file.

## [1.8.5] - 2020-12-18

### Changed

- CASMHMS-4295 - Changed the regex for the partition hmstype to accept p# and p#.#

## [1.8.4] - 2020-12-2

### Added

- CASMHMS-4246 - Added CDUMgmtSwitch to HMS types.

## [1.8.3] - 2020-11-24

### Added

- CASMHMS-4239 - Added MgmtHLSwitch to HMS types.

## [1.8.2] - 2020-11-16

### Added 

- CASMHMS-4087 - Added NodeAccelRiser to HMS types.

## [1.8.1] - 2020-10-16

### Security

- CASMHMS-4105 - Updated base Golang Alpine image to resolve libcrypto vulnerability.

## [1.8.0] - 2020-09-02

### Changed

- CASMHMS-3922 - Updated components to include reservation disabled and locked status.

## [1.7.3] - 2020-05-01

### Changed

- CASMHMS-3403 - Updated regex used for cmmrectifiers to allow more than 3 xnames to be validated per chassis

## [1.7.2] - 2020-04-27

### Changed

- CASMHMS-2968 - Updated hms-base to use trusted baseOS.

## [1.7.1] - 2020-04-09

### Added

- The NodeEnclosurePowerSupply type to HMS types

## [1.7.0] - 2020-03-18

### Changed

- Changed the valid component role and subrole values to be extendable via configfile.

### Added

- A config file watcher to pick up any new roles/subroles defined in the config file.

## [1.6.4] - 2020-03-13

### Changed

- Changed the component state transition Ready->On to be invalid

## [1.6.3] - 2020-03-06

### Added

- Definitions for HMS hardware Class (River/Mountian/Hill)

## [1.6.2] - 2020-02-13

### Added

- Added functions for listing out all valid values for the enums defined in hms-base (HMS Type, State, Role, etc).

## [1.6.1] - 2020-01-29

### Added

- The Drive and StorageGroup types to HMS types

## [1.6.0] - 2019-12-11

### Changed

- Split this module into a separate package from hms-common

## [1.5.5] - 2019-12-02

### Added

- The SNMPAuthPass and SNMPPrivPass fields to the CompCredentials struct

## [1.5.4] - 2019-11-22

### Added

- Definitions for subroles

## [1.5.3] - 2019-10-04

### Added

- Extended securestorage mock Vault adapter to also function as a more
  generalized storage mechanism for complex unit test case scenarios.  All
  existing functionality is preserved. Use as a generalized store requires
  initializing InputLookup.Key (or InputLookupKeys.KeyPath) and setting
  LookupNum (or LookupKeysNum) to -1.

## [1.5.2] - 2019-10-03

### Fixed

- Synced up with the HMS Component Naming Convention.  Note that this introduces
some incompatibilties with previous versions.

## [1.5.1] - 2019-09-18

### Added

- Added the "Locked" component flag to base.

## [1.5.0] - 2019-08-13

### Added

- Added SMNetManager already in use in REDS/MEDS to common library.

## [1.4.2] - 2019-08-07

### Fixed

- Segmentation fault in decode logic of secure store when a nil structure (i.e., no results) are returned from Vault.

## [1.4.1] - 2019-08-01

### Added

- Management role to base

## [1.4.0] - 2019-07-30

### Added

- Added the securestorage package that performs basic actions (Store, Lookup, etc) on a chosen secure backing store. The initial list of backing stores only includes Vault.
- Added the compcredentials package that performs common component credential operations with the securestorage package.

## [1.3.0] - 2019-07-08

### Added

- Added HTTP library that utilizes retryablehttp to perform HTTP operations and optionally unmarshal the returned value into an interface.

## [1.2.0] - 2019-05-18

### Changed

- Added changes for CabinetPDU support
- Tweak to state change table

## [1.1.0] - 2019-05-13

### Removed

- Removed `hmsds`, `sharedtest`, `sm`, and `redfish` packages from this repo as they are actually SMD specific and therefore belong in that repo.

## [1.0.0] - 2019-05-13

### Added

- This is the initial release of the `hms-common` repo. It contains everything that was in `hms-services` at the time with the major exception of being `go mod` based now.

### Changed

### Deprecated

### Removed

### Fixed

### Security
