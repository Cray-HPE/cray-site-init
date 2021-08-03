# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


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
