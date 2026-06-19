#!/usr/bin/env bash
# Offline Nexus manifest tests + optional live verify/upload.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=nexus-common.sh
source "$SCRIPT_DIR/nexus-common.sh"

LIVE=0
UPLOAD_TAG=""

usage() {
  cat <<EOF
Usage: $0 [options]

Runs Nexus automation tests:
  1. manifest generator unit checks
  2. local HTTP fixture + verify-nexus.sh + muxdev update --check
  3. optional live Nexus verify (--live) or full publish (--upload TAG)

Options:
  --live           Verify remote latest.json (requires published artifacts)
  --upload TAG     Build, upload, and verify TAG on Nexus (requires NEXUS_AUTH)
  -h, --help       Show help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --live)
      LIVE=1
      shift
      ;;
    --upload)
      UPLOAD_TAG="${2:-}"
      if [[ -z "$UPLOAD_TAG" ]]; then
        nexus_err "--upload requires a tag"
      fi
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      nexus_err "unknown argument: $1"
      ;;
  esac
done

pass=0
fail=0

ok() {
  printf 'PASS: %s\n' "$1"
  pass=$((pass + 1))
}

bad() {
  printf 'FAIL: %s\n' "$1" >&2
  fail=$((fail + 1))
}

test_manifest_generator() {
  local tmp dist manifest
  tmp="$(mktemp -d)"
  dist="$tmp/dist"
  mkdir -p "$dist"

  printf 'deadbeef  muxdev_9.9.9_linux_amd64.tar.gz\n' >"$dist/checksums.txt"
  tar -czf "$dist/muxdev_9.9.9_linux_amd64.tar.gz" -C "$dist" checksums.txt

  manifest="$tmp/latest.json"
  nexus_generate_manifest "9.9.9" "v9.9.9" "http://example/stable/v9.9.9" "$dist" "$manifest"

  python3 - "$manifest" <<'PY'
import json, sys
from pathlib import Path
data = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert data["tag"] == "v9.9.9"
assert data["assets"]["linux_amd64"] == "muxdev_9.9.9_linux_amd64.tar.gz"
assert data["checksums"].endswith("/checksums.txt")
PY

  rm -rf "$tmp"
  ok "manifest generator"
}

test_local_fixture() {
  local tmp publish release muxdev_bin port server_pid manifest_url
  tmp="$(mktemp -d)"
  publish="$tmp/stable"
  release="$publish/v0.0.0-test"
  mkdir -p "$release"

  printf 'cafebabe  muxdev_0.0.0-test_linux_amd64.tar.gz\n' >"$release/checksums.txt"
  tar -czf "$release/muxdev_0.0.0-test_linux_amd64.tar.gz" -C "$release" checksums.txt

  port="$(python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
)"

  manifest_url="http://127.0.0.1:${port}/stable/latest.json"
  base_url="http://127.0.0.1:${port}/stable/v0.0.0-test"
  nexus_generate_manifest "0.0.0-test" "v0.0.0-test" "$base_url" "$release" "$publish/latest.json"
  cp "$publish/latest.json" "$release/manifest.json"

  python3 -m http.server "$port" --directory "$tmp" >/dev/null 2>&1 &
  server_pid=$!
  for _ in $(seq 1 30); do
    if curl -fsS "$manifest_url" >/dev/null 2>&1; then
      break
    fi
    sleep 0.1
  done
  cleanup() {
    kill "$server_pid" >/dev/null 2>&1 || true
    rm -rf "$tmp"
  }
  trap cleanup RETURN

  muxdev_bin="$ROOT/tmp/muxdev-test"
  (cd "$ROOT" && go build -o "$muxdev_bin" ./cmd/muxdev)

  MUXDEV_BIN="$muxdev_bin" "$SCRIPT_DIR/verify-nexus.sh" "$manifest_url"

  ok "local fixture verify + muxdev update --check"
  cleanup
  trap - RETURN
}

test_go_update_package() {
  (cd "$ROOT" && go test ./internal/update/... -count=1 >/dev/null)
  ok "go update package tests"
}

run_live_verify() {
  nexus_defaults
  if ! "$SCRIPT_DIR/verify-nexus.sh"; then
    bad "live Nexus verify"
    return
  fi
  ok "live Nexus verify"
}

run_live_upload() {
  chmod +x "$SCRIPT_DIR/release-nexus.sh"
  if ! "$SCRIPT_DIR/release-nexus.sh" "$UPLOAD_TAG"; then
    bad "live Nexus upload for $UPLOAD_TAG"
    return
  fi
  ok "live Nexus upload for $UPLOAD_TAG"
}

printf '== Nexus automation tests ==\n'

test_manifest_generator
test_go_update_package
test_local_fixture

if [[ "$LIVE" -eq 1 ]]; then
  run_live_verify
fi

if [[ -n "$UPLOAD_TAG" ]]; then
  run_live_upload
fi

printf '\n== summary: %d passed, %d failed ==\n' "$pass" "$fail"
if [[ "$fail" -ne 0 ]]; then
  exit 1
fi
