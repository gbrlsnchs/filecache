# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.6.0] - 2018-10-22
### Changed
- Update radix tree dependency.

## [0.5.0] - 2018-10-03
### Added
- `NewSize` constructor.
- Benchmark tests.
- More test data.

### Changed
- Optimize directory walking by limiting the number of concurrent goroutines.
- Use `runtime.NumCPU()` as the default value of concurrent goroutines when using `New`.

## [0.4.1] - 2018-10-02
### Fixed
- Rename back function `Read` to `ReadDir`.

## [0.4.0] - 2018-10-02
### Changed
- Rename `ReadDir` method to `Load`.
- Rename `ReadDirContext` method to `LoadContext`.

## [0.3.0] - 2018-10-02
### Added
- `ReadDir` and `ReadDirContext` methods for `Cache`.

### Fixed
- Remove "Makefile" from changelog for `v0.1.0`.

## [0.2.0] - 2018-10-01
### Changed
- Enhance channeling and context canceling.

## 0.1.0 - 2018-09-25
### Added
- This changelog file.
- README file.
- MIT License.
- CI configuration files.
- Git ignore file.
- EditorConfig file.
- Source code.
- Go modules files.

[0.6.0]: https://github.com/gbrlsnchs/filecache/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/gbrlsnchs/filecache/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/gbrlsnchs/filecache/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/gbrlsnchs/filecache/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/gbrlsnchs/filecache/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/gbrlsnchs/filecache/compare/v0.1.0...v0.2.0
