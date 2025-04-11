# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.12.0] - 2025-01-29

### Security

- Update module dependencies
- Remove deprecated Version from docker compose files
- Replace "docker-compose" with "docker compose" in runUnitTest.sh

## [1.11.0] - 2024-11-22

### Changed

- Updated go to 1.23

## [1.10.1] - 2021-08-10

### Changed

- updated docker file and added .github

## [1.10.0] - 2021-08-03

### Changed

- Upgraded AWS SDK to v1.40.14.

## [1.9.2] - 2021-07-21

### Changed

- Replaced all references to Stash with GitHub.

## [1.9.1] - 2021-07-19

### Changed

- Add support for building within the CSM Jenkins.

## [1.9.0] - 2021-06-28

### Security

- CASMHMS-4898 - Updated base container images for security updates.

## [1.8.2] - 2021-05-04

### Changed

- Updated docker-compose files to pull images from Artifactory instead of DTR.

## [1.8.1] - 2021-04-19

### Changed

- Updated Dockerfiles to pull base images from Artifactory instead of DTR.

## [1.8.0] - 2021-01-27

### Changed

- Updated copyrights/license info in source files.

## [1.7.0] - 2021-01-14

### Changed

- Updated license file.

## [1.6.0] - 2021-01-04

### Changed

- CASMHMS-4323 - Added delete operation for objects. Added a validation function to the ConnectionInfo struct.

## [1.5.0] - 2020-12-01

### Changed

- CASMINST-462 - Added logic necessary for creating buckets with ACL's and uploading directly from files without first reading the contents into a byte array.

## [1.4.1] - 2020-10-21

### Security

- CASMHMS-4105 - Updated base Golang Alpine image to resolve libcrypto vulnerability.

## [1.4.0] - 2020-08-03

### Changed

- CASMHMS-3408 - Updated hms-s3 to use the latest trusted baseOS images.

## [1.3.0] - 2020-06-17

### Changed

- Cleaned up style a bit by removing logging and ensuring errors are passed back appropriately and allow for the override of the default HTTP client.

## [1.2.0] - 2020-06-16

### Removed

- CASMHMS-3037 - Removed unused CT test library file that was mistakenly pulled into repository.

## [1.1.0] - 2020-05-01

### Changed

- Added configurable logging, and removed interface

## [1.0.0] - 2019-05-13

### Added

- This is the initial release of the `hms-s3` repo.

### Changed

### Deprecated

### Removed

### Fixed

### Security
