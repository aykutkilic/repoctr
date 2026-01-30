# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `repo-ctr version` command to display current version
- `repo-ctr update` command for self-updating from GitHub releases
  - Displays release notes for all versions since current version
  - SHA256 checksum verification for downloaded binaries
  - `--check` flag to only check for updates without installing
  - `--force` flag to update even if already on latest version
  - `--skip-checksum` flag to bypass checksum verification (not recommended)
- Version embedding via ldflags at build time
- SHA256 checksum file generation in release workflow

### Security
- HTTP client timeout (60 seconds) to prevent hanging requests
- URL validation to ensure downloads only from GitHub domains
- Binary integrity verification via SHA256 checksums
- Atomic binary replacement with rollback on failure

## [0.1.1] - 2026-01-30

### Changed
- Filter LOC by runtime type
- Improve install script

## [0.1.0] - 2026-01-30

### Added
- Initial release
- Project discovery for Go, Python, JavaScript/TypeScript, Java, .NET, Rust, Dart, C/C++
- Lines of code statistics with hierarchical project support
- Multiple output formats: human-readable, YAML, JSON, XML, CSV
- `repo-ctr init` command to create projects.yaml template
- `repo-ctr identify` command to auto-discover projects
- `repo-ctr stats` command to show LOC statistics
