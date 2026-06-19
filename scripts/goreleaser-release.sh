#!/usr/bin/env bash
# Run Goreleaser for a tag; skip GitHub publish when release assets already exist.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  cat <<EOF
Usage: $0 <tag> [release-notes.md]

Build and optionally publish GitHub release assets for an existing tag checkout.
EOF
}

TAG="${1:-}"
NOTES="${2:-}"
if [[ -z "$TAG" ]]; then
  usage
  exit 1
fi

goreleaser_bin=""
if [[ -n "${GORELEASER:-}" && -x "$GORELEASER" ]]; then
  goreleaser_bin="$GORELEASER"
elif command -v goreleaser >/dev/null 2>&1; then
  goreleaser_bin="$(command -v goreleaser)"
else
  goreleaser_bin="$(go env GOPATH)/bin/goreleaser"
fi

ARGS=(release --clean --skip=validate)
if [[ -n "$NOTES" && -f "$NOTES" ]]; then
  ARGS+=(--release-notes="$NOTES")
fi

if gh release view "$TAG" >/dev/null 2>&1; then
  asset_count="$(gh release view "$TAG" --json assets --jq '.assets | length')"
  if [[ "${asset_count:-0}" -gt 0 ]]; then
    printf 'release %s already has %s assets — building only (--skip=publish)\n' "$TAG" "$asset_count"
    ARGS+=(--skip=publish)
  fi
fi

if [[ -z "${TAP_GITHUB_TOKEN:-}" ]]; then
  ARGS+=(--skip=homebrew --skip=scoop)
fi

exec "$goreleaser_bin" "${ARGS[@]}"
