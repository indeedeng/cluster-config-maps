# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2023-10-17
### Changed
- Updated dependencies on k8s apis

## [0.2.0] - 2022-03-29
### Changed
- Removed volumeID from prometheus metrics tags
### Fixed
- Added locking around publish requests to prevent concurrent publish/unpublish requests for the same volumeID causing inconsistent state.

## [0.1.1] - 2021-11-09
### Fixed
- Added edge case handling for node unpublish volume requests where the volume was already cleaned up.

## [0.1.0] - 2021-11-09
### Added
- Metrics for successful, failed, and cleanup csi operations

## [0.0.3] - 2021-09-30
### Fixed
- Fixed permission of the volume dir, from `0700` to `0755`

## [0.0.2] - 2021-09-29
### Fixed
- Fixed permission handling with defaultMode
### Added
- Added cleanup logic to delete unused cluster config map data in the daemonset hostpath dir on NodeUnpublishVolume requests

## [0.0.1] - 2021-08-27
### Added
- Initial release
