# Runtime

muxdev is implemented in **Go** as a single cross-platform binary.

## Stack

| Layer | Choice |
|-------|--------|
| CLI | [cobra](https://github.com/spf13/cobra) |
| Config | `gopkg.in/yaml.v3` — `muxdev.yaml` |
| TUI | [bubbletea](https://github.com/charmbracelet/bubbletea) |
| Process runner | `os/exec` + platform-specific signal handling |

## Why Go

- Single static binary for Windows, macOS, and Linux
- Native Windows support without WSL
- Mature TUI ecosystem (bubbletea)
- Straightforward cross-compilation and Goreleaser distribution

## Reference implementation

The original Bash TUI prototype (service picker, log multiplexing, shutdown) is the behavioral reference. It is not shipped; behavior is ported into Go.

## Platform notes

| OS | Shell wrapper | Command execution |
|----|---------------|-------------------|
| Linux / macOS | `/bin/sh -c` | Native shell commands |
| Windows | `cmd.exe /C` | Native commands; `bash script.sh` requires WSL or Git Bash |

## Development

```bash
# Run from repo (wrapper)
./bin/muxdev --list --config testdata/muxdev.yaml

# Or build directly
go build -o muxdev ./cmd/muxdev
./muxdev --version
```

## Roadmap

- [x] Go module + CLI skeleton
- [x] `muxdev.yaml` loader and validation
- [x] `--list`, `--focus`, `--no-interactive`
- [x] Interactive TUI (picker + log viewport)
- [x] Goreleaser + GitHub Actions (linux/macos/windows)
- [x] Install script (`scripts/install.sh`)
- [ ] Homebrew / Scoop manifests
- [x] Example consumer project integration
