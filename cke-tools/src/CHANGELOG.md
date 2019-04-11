# Change Log

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [1.5.1] - 2019-03-07

### Changed

- Fix empty-dir script (#13).

## [1.5.0] - 2019-03-07

### Added

- Add empty-dir script (#12).

## [1.4.0] - 2019-03-04

### Removed

- Remove CNI config file from install-tools (#11).

## [1.3.0] - 2018-12-25

### Added
- Add `etcdbackup` service (#10).

### Removed
- Remove `go.sum` (#9).

## [1.2.1] - 2018-09-21

### Changed
- fix bug in `write_files`.

## [1.2.0] - 2018-09-20

### Added
- `write_files` script to extract tar archive under a root directory.

### Removed
- `write_file` script.

## [1.1.1] - 2018-09-20

### Changed
- Fix wrong permission bug in `write_file`.

## [1.1.0] - 2018-09-18

### Added
- Utilities to install CNI plugins (`install-cni`)

## [1.0.0] - 2018-09-18

### Added
- Opt in to [Go modules](https://github.com/golang/go/wiki/Modules).
- Utilities to create files and directories.

## [0.2] - 2018-09-06

### Changed
- Fix warning log.

## [0.1] - 2018-08-05

### Added
- Add rivers: an TCP reverse proxy for kubernetes apiservers (#1).

[Unreleased]: https://github.com/cybozu-go/cke-tools/compare/v1.5.1...HEAD
[1.5.1]: https://github.com/cybozu-go/cke-tools/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/cybozu-go/cke-tools/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/cybozu-go/cke-tools/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/cybozu-go/cke-tools/compare/v1.2.1...v1.3.0
[1.2.1]: https://github.com/cybozu-go/cke-tools/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/cybozu-go/cke-tools/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/cybozu-go/cke-tools/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/cybozu-go/cke-tools/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/cybozu-go/cke-tools/compare/v0.2...v1.0.0
[0.2]: https://github.com/cybozu-go/cke-tools/compare/v0.1...v0.2
[0.1]: https://github.com/cybozu-go/cke-tools/compare/b797246...v0.1
