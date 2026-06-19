#!/usr/bin/env bash
# Build release artifacts, upload to Nexus, and verify latest.json.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=nexus-common.sh
source "$SCRIPT_DIR/nexus-common.sh"

BUILD=1
SKIP_VERIFY=0

usage() {
  cat <<EOF
Usage: $0 [options] <tag>

Build (optional), upload Goreleaser dist/ artifacts to Nexus, then verify.

Options:
  --no-build       Skip Goreleaser build; use existing dist/
  --skip-verify    Upload only; do not run verify-nexus.sh
  -h, --help       Show help

Environment:
  NEXUS_URL, NEXUS_REPO, NEXUS_CHANNEL, NEXUS_AUTH, DIST_DIR
  GORELEASER       Path to goreleaser binary (default: goreleaser)

Example:
  NEXUS_AUTH=user:password $0 v1.0.0
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-build)
      BUILD=0
      shift
      ;;
    --skip-verify)
      SKIP_VERIFY=1
      shift
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    -*)
      nexus_err "unknown option: $1"
      ;;
    *)
      break
      ;;
  esac
done

TAG="${1:-}"
if [[ -z "$TAG" ]]; then
  usage
  exit 1
fi

if ! [[ "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  nexus_err "invalid tag: $TAG (expected vX.Y.Z)"
fi

nexus_defaults

if [[ "$BUILD" -eq 1 ]]; then
  goreleaser_bin=""
  if [[ -n "${GORELEASER:-}" && -x "$GORELEASER" ]]; then
    goreleaser_bin="$GORELEASER"
  elif command -v goreleaser >/dev/null 2>&1; then
    goreleaser_bin="$(command -v goreleaser)"
  else
    candidate="$(go env GOPATH 2>/dev/null)/bin/goreleaser"
    if [[ -x "$candidate" ]]; then
      goreleaser_bin="$candidate"
    fi
  fi
  if [[ -z "$goreleaser_bin" ]]; then
    nexus_err "goreleaser not found — install it or pass --no-build with an existing dist/"
  fi

  worktree="$(mktemp -d)"
  cleanup_worktree() {
    git -C "$ROOT" worktree remove "$worktree" --force >/dev/null 2>&1 || rm -rf "$worktree"
  }
  trap cleanup_worktree EXIT

  printf 'building release artifacts for %s\n' "$TAG"
  git -C "$ROOT" worktree add --detach "$worktree" "$TAG"
  (
    cd "$worktree"
    "$goreleaser_bin" release --clean --skip=publish --skip=validate
  )
  rm -rf "$ROOT/dist"
  cp -a "$worktree/dist" "$ROOT/dist"
  cleanup_worktree
  trap - EXIT
fi

chmod +x "$SCRIPT_DIR/upload-nexus.sh"
(
  cd "$ROOT"
  DIST_DIR="${DIST_DIR:-dist}" "$SCRIPT_DIR/upload-nexus.sh" "$TAG"
)

if [[ "$SKIP_VERIFY" -eq 0 ]]; then
  chmod +x "$SCRIPT_DIR/verify-nexus.sh"
  "$SCRIPT_DIR/verify-nexus.sh"
fi

printf 'release complete: %s\n' "$(nexus_manifest_url "$(nexus_publish_base "$NEXUS_URL" "$NEXUS_REPO" "$NEXUS_CHANNEL")")"
