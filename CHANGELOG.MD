# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.27] - 2021-04-05

### Changed

- CASMINST-1545: User provided application node prefix to HSM SubRole mappings specified via `application_node_config.yaml` now take precedence over the default HSM SubRole mappings within in CSI.

## [1.5.4] - 2020-12-10

### Added

- Support for the system_config.yaml to be a config file for csi instead of or in addition to the flags
- Improved support for 1.3 network configurations
- Flags/Configs for site-network connections
- Support for smaller CAN subnets
- Better error messages when ipam fails to find sufficient ip space
- Clearer .nmn/.hmn separation in hostnames and domains

## [1.5.3] - 2020-12-06

### Added

- Allow the switches and ncns to share the supernet with the --supernet option
- Tweaks to network layouts for pinned hostnames during install
- Bugfixes and minor tweaks
