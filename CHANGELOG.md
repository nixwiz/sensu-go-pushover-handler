# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

## [1.2.1] - 2024-12-18

### Changed
- Update dependencies to address security issue

## [1.2.0] - 2024-05-07

### Changed
- Updated to use Go 1.22
- Updated dependencies

## [1.1.0] - 2022-03-11

### Changed
- Made changes to reflect deprecation of ioutil
- Updated to use Go 1.17
- Updated dependencies
- Changed testing of main()

### Added
- darwin-arm64 (Apple Silicon) to build releases

## [1.0.0] - 2021-06-05

### Changed
- Updated GitHub Actions to use Go 1.14
- Updated dependencies and fixed path to sensu-plugin-sdk
- Updated and fixed test

## [0.9.0] - 2021-01-27

### Changed
- Updated SDK and other modules
- Reorganized and updated README

### Added
- Added lint GitHubAction
- Output of Pushover request ID

## [0.8.0] - 2020-08-14

### Changed
- Updated SDK to 0.8.0
- Set secret bool to true for token and auth

## [0.7.0] - 2020-06-02

### Added
- Retry and Expire flags to allow sending emergency (Pri 2) alerts

## [0.6.4] - 2020-05-05

### Changed
- Removed unneeded bonsai entries

## [0.6.3] - 2020-05-05

### Changed
- Changed goreleaser targets

## [0.6.1] - 2020-05-05

### Changed
- Change resolveTemplate to return error instead of panic
- Updated to use SDK template

## [0.6.0] - 2020-02-18

### Added
- Added configurable message sound

### Changed
- Updated README to include new argument for setting API URL
- Minor cleanup
- Minor cleanup for golint
- Added goreportcard.com
- Improved test coverage
- README changes for secrets and a few other README fixes

## [0.5.1] - 2020-02-12

### Changed
- Made pushoverAPIURL a configurable variable to facilitate testing

### Added
- Tests, including GitHub Actions

## [0.5.0] - 2020-02-10

### Changed
- Fixed goreleaser deprecated archive to use archives
- Replaced Travis CI with GitHub Actions
- Use new Sensu SDK module

## [0.4.1] - 2019-12-17

### Changed
- Reformatted README for [Plugin Style Guide](https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md)

## [0.4.0] - 08/22/2019

### Changed
- Rewrote to use confighandler
- Updated to use modules
- Compiled with go1.12.9
- Minor documentation changes
- Remove v from version numbers for goreleaser

## [0.3.3] - 06/10/2019

### Changed
- Updated README.md and main.go for program name to be consistent

## [0.3.2] - 05/16/2019

### Changed
- Updated README.md for incorrect env variables, and need to roll version to update bonsai

## [0.3.1] - 03/29/2019

### Changed
- Updated .goreleaser.yml to fix versioning and sha512 checksum

### Added
- Sensu bonsai

## [0.3.0] - 03/05/2019

### Changed
- Changed the environment variables for consistency
- Updated the sample events

### Added
- Support for annotations

## [0.2.0] - 02/27/2019

### Changed
- Changed how message priorities are set

### Added
- Added command line arguments for message priorities

## [0.1.1] - 02/26/2019

### Added
- Fixed bug for status == 0

## [0.1.0] - 02/26/2019

### Added
- Initial release

