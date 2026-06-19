# Changelog

All notable changes to **muxdev** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.0](https://github.com/yarkingulacti/muxdev-cli/compare/v1.5.0...v1.6.0) (2026-06-19)


### Added

* **configure:** confirm auto-discovered port before applying ([de0ccdb](https://github.com/yarkingulacti/muxdev-cli/commit/de0ccdb74eda6bb9392459aebacbfda572dcfe46))

## [1.5.0](https://github.com/yarkingulacti/muxdev-cli/compare/v1.4.0...v1.5.0) (2026-06-19)


### Added

* **tui:** add graceful quit and interactive update prompt ([49d20f2](https://github.com/yarkingulacti/muxdev-cli/commit/49d20f2cdb3c53f6ea2f560c419a97b4f6420dc9))

## [1.4.0](https://github.com/yarkingulacti/muxdev-cli/compare/v1.3.1...v1.4.0) (2026-06-19)


### Added

* **help:** improve help center UI and fix cross-platform tests ([9e66c3f](https://github.com/yarkingulacti/muxdev-cli/commit/9e66c3f8b0939908c06021ca11c5c89f5d04ab21))
* **tui:** add runtime re-run picker ([cfcf3a0](https://github.com/yarkingulacti/muxdev-cli/commit/cfcf3a01bce60a0893e03436b0b733b378b4867f))


### Fixed

* **ci:** idempotent release pipeline and sync release-please ([4fbb26f](https://github.com/yarkingulacti/muxdev-cli/commit/4fbb26fa40b869d6cba7c2f5f0ef3453412b3cc2))

## [1.3.2](https://github.com/yarkingulacti/muxdev-cli/compare/v1.3.1...v1.3.2) (2026-06-19)

### Added

- Redesigned interactive help center with search, summaries, and rich formatting

### Fixed

- CI on macOS (platform-specific test skips)
- Idempotent Goreleaser + Nexus publish in release workflows

## [1.3.1](https://github.com/yarkingulacti/muxdev-cli/compare/v1.3.0...v1.3.1) (2026-06-19)

### Fixed

- Self-update replace when install dir and system temp are on different filesystems

## [1.3.0](https://github.com/yarkingulacti/muxdev-cli/compare/v1.2.0...v1.3.0) (2026-06-19)

### Added

- Persistent session logs with `muxdev logs` CLI and TUI viewer
- Interactive help wiki (`muxdev help`, `-h` / `--help` per command)
- Runtime log scroll history (PgUp/PgDn, Ctrl+U/D) and pagination footer
- Layered port resolution via `BindPortForService` and process-tree port kill

### Fixed

- Graceful shutdown: stop services and release ports on TUI quit
- Nexus manifest `base_url` rewrite when clients fetch via public proxy URL
- Nexus upload overwrite (delete-then-PUT) for `latest.json` and release assets

## [1.2.0](https://github.com/yarkingulacti/muxdev-cli/compare/v1.1.0...v1.2.0) (2026-06-19)

### Added

- Port conflict detection and resolution when starting dev services
- Nexus-backed self-update via `MUXDEV_UPDATE_URL` and automated CI publish on release
- Release notes sourced from `CHANGELOG.md` (no raw git-log output on GitHub Releases)
- Nexus automation scripts: `release-nexus.sh`, `verify-nexus.sh`, `test-nexus.sh`

### Fixed

- Windows CI build for `portkill` (platform-specific build tags)
- Goreleaser binary path in Nexus publish workflow
- Release pipeline ordering: Release Please creates tag first, then Goreleaser + Nexus upload

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
