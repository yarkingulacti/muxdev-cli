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

Bootstrap phase. Runtime and core extraction from the first consumer ([voice-synt](https://github.com/yarkingulacti/voice-synt)) are in progress.

## Usage (planned)

```bash
# From any project root with muxdev.yaml
muxdev

# Options (planned)
muxdev --focus=backend,ui
muxdev --no-interactive
muxdev --list
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

## License

MIT
