<p align="center">
  <img src="https://img.shields.io/github/v/release/yarkingulacti/muxdev-cli?logo=github&label=release&style=for-the-badge" alt="release">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="go">
  <img src="https://img.shields.io/badge/license-MIT-blue?style=for-the-badge" alt="license">
  <img src="https://img.shields.io/github/stars/yarkingulacti/muxdev-cli?style=for-the-badge&logo=github" alt="stars">
</p>

# 🖥️ muxdev

**Multiplexed dev stack runner — config-driven local development orchestrator with an interactive terminal UI.**

Service picker · multiplexed logs · port conflict resolution · session history · self-update — one CLI, zero bash spaghetti in your app repo.

```bash
curl -fsSL https://raw.githubusercontent.com/yarkingulacti/muxdev-cli/master/scripts/install.sh | bash
```

```text
┌─ muxdev ────────────────────────────────────────────────────────────────┐
│  My App · Local development stack                          v1.3.2       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  ◉ backend   Backend API          :5005                         │    │
│  │  ◉ ui        Web UI               :3131                         │    │
│  └─────────────────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────────────────┤
│ [backend]  INFO  Uvicorn running on http://0.0.0.0:5005                 │
│ [ui]       ready  Local: http://localhost:3131                          │
│ [ui]       ✓      compiled successfully                                 │
├─────────────────────────────────────────────────────────────────────────┤
│  ↑↓ pick  Space toggle  Enter run  q quit  ? help  PgUp/PgDn scroll     │
└─────────────────────────────────────────────────────────────────────────┘
```

---

Monorepo dev scripts grow into large, project-coupled bash files: service pickers, log multiplexing, resize handling, and shutdown logic all live beside application code.

`muxdev` is a standalone CLI that reads a project manifest (`muxdev.yaml`) and keeps orchestration out of your application tree — with a fixed-layout TUI, clean startup/shutdown, and optional CI-friendly plain output.

> **Repository:** [github.com/yarkingulacti/muxdev-cli](https://github.com/yarkingulacti/muxdev-cli)

---

## Contents

- [Why muxdev?](#-why-muxdev)
- [60-second tour](#-60-second-tour)
- [Install](#-install)
- [Usage modes](#-usage-modes)
- [CLI reference](#-cli-reference)
- [Project manifest](#-project-manifest)
- [Self-update](#-self-update)
- [Self-hosted updates (Nexus)](#-self-hosted-updates-nexus)
- [Documentation](#-documentation)
- [License](#-license)

---

## 🤔 Why muxdev?

| | What you get |
|---|---|
| 🎯 **Config-driven** | Define services, ports, and `depends_on` in `muxdev.yaml` — no hard-coded bash in app repos. |
| 🖥️ **Interactive TUI** | Multiselect picker, focused log panel, keyboard shortcuts, log scroll history, pagination footer. |
| 🔌 **Port-aware** | Detects conflicts, resolves process trees, and binds ports via layered env resolution. |
| 📜 **Session logs** | Persists runtime output to platform paths; browse with `muxdev logs`. |
| 📖 **Built-in help** | Interactive wiki (`muxdev help`, `-h` per command) generated from CLI docs. |
| 🔄 **Self-update** | `muxdev update` from [GitHub Releases](https://github.com/yarkingulacti/muxdev-cli/releases) (default) or a Nexus manifest |
| 🧰 **Cross-platform** | Linux, macOS, Windows — amd64 & arm64; Homebrew, Scoop, winget, curl, `go install`. |

## ⚡ 60-second tour

```bash
# Interactive TUI — pick services, stream logs
muxdev

# List configured services
muxdev --list

# Run a subset (pulls in dependencies automatically)
muxdev --focus=backend,ui

# Plain multiplexed logs for CI / pipes
muxdev --no-interactive

# Scaffold or edit muxdev.yaml interactively
muxdev init
muxdev configure

# Check version & updates
muxdev version --short
muxdev update --check
```

That's the core loop: point `muxdev` at a project with `muxdev.yaml`, pick what to run, and get multiplexed logs with graceful shutdown when you quit.

## 📦 Install

### curl (Linux / macOS / Git Bash)

```bash
curl -fsSL https://raw.githubusercontent.com/yarkingulacti/muxdev-cli/master/scripts/install.sh | bash
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

### Build from source

```bash
git clone https://github.com/yarkingulacti/muxdev-cli.git
cd muxdev-cli
go build -o muxdev ./cmd/muxdev
# or: ./bin/muxdev --version
```

See [docs/local-development.md](docs/local-development.md) for a short local dev guide.

## 🎛️ Usage modes

| Mode | Command | When to use |
|------|---------|-------------|
| 🖥️ **Interactive** | `muxdev` | Daily dev — picker + log panel + shortcuts. |
| 🎯 **Focused** | `muxdev --focus=backend,ui` | Run a subset; dependencies included automatically. |
| 📋 **List** | `muxdev --list` | Inspect services, ports, and env sources. |
| 🤖 **CI / pipes** | `muxdev --no-interactive` | Plain multiplexed stdout/stderr, no TUI. |
| ⚙️ **Configure** | `muxdev init` / `muxdev configure` | Create or edit `muxdev.yaml` with the wizard. |
| 📜 **Logs** | `muxdev logs` | Browse persisted session logs from past runs. |

## 📟 CLI reference

<details>
<summary><strong>Click to expand — common invocations</strong></summary>

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

# Browse session logs
muxdev logs

# Interactive help wiki
muxdev help
muxdev run --help          # opens wiki topic for `run`

# Version & self-update
muxdev version
muxdev version --short
muxdev update --check      # exit 2 if update available
muxdev update --yes
```

</details>

## 📄 Project manifest

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

See [testdata/muxdev.yaml](testdata/muxdev.yaml) for a minimal project config you can copy into your repo.

## 🔄 Self-update

By default, `muxdev update` uses the public GitHub Releases API — no extra config needed:

```bash
muxdev update --check     # check only (exit 2 if update available)
muxdev update --yes       # download, verify checksum, replace binary
muxdev version --short    # e.g. 1.3.2
```

Re-install from the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/yarkingulacti/muxdev-cli/master/scripts/install.sh | bash
```

Package manager installs should use their native upgrade commands (`brew upgrade`, `scoop update`, `winget upgrade`).

## 🛰️ Self-hosted updates (Nexus)

Optional: point the updater at a Nexus (or other static) manifest instead of GitHub:

```bash
export MUXDEV_UPDATE_URL="https://apps.developeryarkin.com/repository/muxdev-releases/stable/latest.json"
muxdev update --check
muxdev update --yes
```

Optional auth (when anonymous read is disabled):

```bash
export MUXDEV_UPDATE_USER=your-user
export MUXDEV_UPDATE_TOKEN=your-token
```

Publish release artifacts:

```bash
# requires NEXUS_AUTH in .env or environment
./scripts/release-nexus.sh v1.3.2
./scripts/verify-nexus.sh
./scripts/test-nexus.sh          # offline + local fixture e2e
./scripts/test-nexus.sh --live   # verify remote latest.json
```

**CI release flow:** push to `master` → Release Please → Goreleaser → GitHub Release → Nexus upload (secrets: `NEXUS_URL`, `NEXUS_AUTH`, optional `NEXUS_REPO`, `TAP_GITHUB_TOKEN` for Homebrew/Scoop).

## 📚 Documentation

| Doc | Description |
|-----|-------------|
| [docs/local-development.md](docs/local-development.md) | Build and run from source |
| [docs/runtime.md](docs/runtime.md) | Runtime decision framework (Bash vs Go vs alternatives) |
| [docs/release.md](docs/release.md) | SemVer and Goreleaser shipping |
| [docs/git-workflow.md](docs/git-workflow.md) | Branch flow: `feature/*` → PR → `dev` → PR → `master` → release |
| [CHANGELOG.md](CHANGELOG.md) | Release history |

## 📜 License

MIT — see [LICENSE](LICENSE).
