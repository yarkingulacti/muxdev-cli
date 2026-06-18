# muxdev

**Multiplexed dev stack runner** — config-driven local development orchestrator with an interactive terminal UI.

## Problem

Monorepo dev scripts grow into large, project-coupled bash files: service pickers, log multiplexing, resize handling, and shutdown logic all live beside application code.

## Solution

`muxdev` is a standalone CLI that reads a project manifest (`muxdev.yaml`) and provides:

- Interactive service picker (multiselect, dependencies)
- Focused log streaming across multiple processes
- Fixed-layout TUI (header metadata card, scrollable logs, keyboard shortcuts)
- Clean startup and shutdown

## Repository

https://github.com/yarkingulacti/muxdev-cli

## Status

Go CLI with interactive TUI (service picker + log panel), non-interactive mode, and cross-platform release tooling.

## Build

```bash
go build -o muxdev ./cmd/muxdev
```

Or run from the repo without installing:

```bash
./bin/muxdev --version
```

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/yarkingulacti/muxdev-cli/main/scripts/install.sh | bash
```

## Usage

```bash
# Interactive TUI (picker + log panel)
muxdev

# List configured services
muxdev --list

# Run a subset (includes dependencies)
muxdev --focus=backend,ui

# Plain multiplexed logs (CI / pipes)
muxdev --no-interactive

# Explicit config path
muxdev --config ./muxdev.yaml --list
```

## Project manifest (planned)

```yaml
name: My App
subtitle: Local development stack

services:
  backend:
    label: Backend
    command: bash apps/backend/run-dev.sh
    port: "${BACKEND_PORT}"
    depends_on: []

  ui:
    label: Web UI
    command: bash scripts/dev-ui.sh
    port: "${UI_PORT}"
    depends_on: [backend]
```

## First consumer

[voice-synt](https://github.com/yarkingulacti/voice-synt) — voice synthesis platform. The current dev TUI (`scripts/lib/dev-*.sh`) is the reference implementation to extract into this repo.

## Runtime

See [docs/runtime.md](docs/runtime.md) for the runtime decision framework (Bash vs Go vs alternatives).

## Release & distribution

See [docs/release.md](docs/release.md) for SemVer (Release Please), Goreleaser shipping, install channels (GitHub, Homebrew, Scoop, winget), and the update mechanism (`muxdev update`).

## License

MIT
