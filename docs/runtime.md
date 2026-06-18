# Runtime decision

This document frames the choice of implementation language/runtime for muxdev.

## Candidates

| Runtime | Pros | Cons | Fit with current code |
|---------|------|------|------------------------|
| **Bash** | ~800 lines of TUI already working in voice-synt; zero new runtime; fastest path to extraction | Hard to distribute and test; long-term maintenance cost | Full — move `dev-tui`, `dev-runner`, `dev-picker`, etc. |
| **Go + bubbletea** | Single static binary; solid signal/resize handling; `go install` | 2–4 week rewrite | Same `muxdev.yaml`; engine swap only |
| **Rust + ratatui** | Fast, modern TUI ecosystem | Steeper learning curve; rewrite cost | Same manifest |
| **Node + ink** | npm ecosystem | Requires Node; shell orchestration still needed | Partial |

## Recommended path

### Phase 1 — Bash (short term)

- Move generic libs from voice-synt into `muxdev/lib/`
- Introduce `muxdev.yaml` config loader
- Keep voice-synt as thin wrapper + project-specific preflight (Postgres, Prisma)

**When to choose:** You want voice-synt working with muxdev within days, not weeks.

### Phase 2 — Freeze config schema (medium term)

- Stabilize `muxdev.yaml` fields: `name`, `services`, `preflight`, `env`
- Bash implementation enters maintenance mode (bugfixes only)

### Phase 3 — Go binary (long term, optional)

- Reimplement TUI/runner in Go against the same manifest
- Ship `muxdev` as a single binary via GitHub releases or `go install`
- voice-synt consumers change nothing except installing the binary

**When to choose:** You need reliable distribution, CI tests, or plan to open-source / publish to npm/Homebrew.

## Decision questions

Answer these before committing to Phase 1 vs jumping to Go:

1. **Audience** — Only your own projects, or public npm/binary distribution?
2. **Distribution** — Is `bash + lib/` acceptable, or is a single binary required?
3. **Testing** — How much automated CI do you need (unit tests for TUI logic, integration with fake services)?
4. **Timeline** — Days (Bash) vs weeks (Go rewrite)?

## Current recommendation

**Start with Bash (Phase 1).** The voice-synt TUI is production-tested. Extract it, define `muxdev.yaml`, integrate voice-synt. Revisit Go when the config schema is stable and distribution becomes a pain point.

## Next steps after decision

- [ ] Extract `lib/` from voice-synt `scripts/lib/dev-{tui,runner,picker,colors,logo,service-wrap}.sh`
- [ ] Implement `muxdev.yaml` parser (bash + `yq` or minimal JSON subset)
- [ ] Add `bin/muxdev` entry that loads config from cwd
- [ ] voice-synt: `pnpm dev` → `muxdev` + project `muxdev.yaml`
