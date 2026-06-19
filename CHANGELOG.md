# Changelog

All notable changes to **muxdev** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0](https://github.com/yarkingulacti/muxdev-cli/compare/v1.0.0...v1.1.0) (2026-06-18)

### Added

- Interactive `muxdev configure` wizard to create and edit `muxdev.yaml`
- `muxdev init` entry point for new projects
- Local development guide (`docs/local-development.md`)

## [1.0.0](https://github.com/yarkingulacti/muxdev-cli/releases/tag/v1.0.0) (2026-06-18)

### Added

- Config-driven local dev stack runner with interactive TUI (`muxdev`)
- `muxdev.yaml` service definitions: commands, ports, dependencies
- Service picker, log multiplexing, and `--focus` / `--list` CLI modes
- Built-in updater (`muxdev update`) with install-method detection
- Multi-platform release artifacts via Goreleaser (Linux, macOS, Windows; amd64 & arm64)
- `install.sh` bootstrap script with SHA-256 verification
- Release Please + SemVer release pipeline

[1.1.0]: https://github.com/yarkingulacti/muxdev-cli/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/yarkingulacti/muxdev-cli/releases/tag/v1.0.0
