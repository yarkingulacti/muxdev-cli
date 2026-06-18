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

### curl (Linux / macOS / Git Bash)

```bash
curl -fsSL https://raw.githubusercontent.com/yarkingulacti/muxdev-cli/main/scripts/install.sh | bash
```

### Homebrew

```bash
brew tap yarkingulacti/tap
brew install muxdev
```

### Scoop (Windows)

```powershell
scoop bucket add yarkingulacti https://github.com/yarkingulacti/scoop-bucket
scoop install muxdev
```

### winget (Windows)

```powershell
winget install yarkingulacti.muxdev
```

### go install

```bash
go install github.com/yarkingulacti/muxdev-cli/cmd/muxdev@latest
```

## Update

```bash
muxdev update --check     # check only (exit 2 if update available)
muxdev update             # self-update for direct installs
muxdev version            # full build metadata
muxdev version --short    # 0.1.0
```

Package manager installs should use their native upgrade commands (`brew upgrade`, `scoop update`, `winget upgrade`).

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

# Create muxdev.yaml interactively
muxdev init

# Edit existing config interactively
muxdev configure

# Explicit config path
muxdev --config ./muxdev.yaml --list
```

## Project manifest

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

See [docs/release.md](docs/release.md) for SemVer and Goreleaser shipping.

## Git workflow

See [docs/git-workflow.md](docs/git-workflow.md) for branch flow: `feature/*` → PR → `dev` → PR → `master` → release.

## License

MIT
