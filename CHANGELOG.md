# Changelog

All notable changes to **muxdev** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0](https://github.com/yarkingulacti/muxdev-cli/compare/v1.1.0...v1.2.0) (2026-06-19)


### Added

* add port conflict handling and nexus release pipeline ([bd97d36](https://github.com/yarkingulacti/muxdev-cli/commit/bd97d36b7b125f3d9930ef4a9ff4eae65b46a564))


### Fixed

* **ci:** add portkill build tag and skip brew without tap token ([8f1694d](https://github.com/yarkingulacti/muxdev-cli/commit/8f1694d4c47b04514f0532ed2850312878f73cac))
* **ci:** resolve goreleaser path in nexus publish workflow ([b44758f](https://github.com/yarkingulacti/muxdev-cli/commit/b44758fc9b20422ae0a313b1ebe32bb130792b3a))
* **release:** use changelog.md for github release notes ([c839c96](https://github.com/yarkingulacti/muxdev-cli/commit/c839c966ed3b8eb1f1abc872605b8d21819698b4))

## [1.1.0](https://github.com/yarkingulacti/muxdev-cli/compare/v1.0.0...v1.1.0) (2026-06-18)

### Added

- Interactive `muxdev configure` wizard to create and edit `muxdev.yaml`
- `muxdev init` entry point with welcome flow for new projects
- Automatic port discovery when defining a service command
- Local development guide (`docs/local-development.md`)

### Fixed

- `install.sh` GitHub Release download URL resolution and checksum verification

## [1.0.0](https://github.com/yarkingulacti/muxdev-cli/releases/tag/v1.0.0) (2026-06-18)

### Added

- Config-driven local dev stack runner with interactive TUI (`muxdev`)
- `muxdev.yaml` service definitions: commands, ports, and `depends_on`
- Service picker, multiplexed logs, and graceful shutdown
- CLI modes: `--list`, `--focus`, `--no-interactive`, and `--config`
- Built-in updater (`muxdev update`) with install-method detection
- `muxdev version` with semver, commit, and build date
- Multi-platform release artifacts (Linux, macOS, Windows; amd64 and arm64)
- `install.sh` bootstrap installer with SHA-256 verification
- Goreleaser release pipeline and Release Please semver automation
- CI test matrix across Linux, macOS, and Windows

[1.1.0]: https://github.com/yarkingulacti/muxdev-cli/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/yarkingulacti/muxdev-cli/releases/tag/v1.0.0
