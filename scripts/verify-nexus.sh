#!/usr/bin/env bash
# Verify a Nexus latest.json manifest and its release artifacts.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=nexus-common.sh
source "$SCRIPT_DIR/nexus-common.sh"

usage() {
  cat <<EOF
Usage: $0 [manifest-url]

Verify muxdev Nexus release metadata and artifacts.

If manifest-url is omitted, uses:
  \${NEXUS_URL}/repository/\${NEXUS_REPO}/\${NEXUS_CHANNEL}/latest.json

Environment:
  NEXUS_URL, NEXUS_REPO, NEXUS_CHANNEL, NEXUS_AUTH
  MUXDEV_BIN   Path to muxdev binary for update --check smoke test (optional)

Example:
  $0
  $0 http://5.178.111.150:8081/repository/muxdev-releases/stable/latest.json
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

nexus_defaults

MANIFEST_URL="${1:-$(nexus_manifest_url "$(nexus_publish_base "$NEXUS_URL" "$NEXUS_REPO" "$NEXUS_CHANNEL")")}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

manifest_file="$tmpdir/latest.json"
printf 'fetching %s\n' "$MANIFEST_URL"
if ! nexus_curl -o "$manifest_file" "$MANIFEST_URL"; then
  nexus_err "failed to fetch manifest from $MANIFEST_URL"
fi

python3 - "$manifest_file" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
data = json.loads(path.read_text(encoding="utf-8"))

required = ("version", "tag", "base_url", "assets")
for key in required:
    if not str(data.get(key, "")).strip():
        raise SystemExit(f"manifest missing {key}")

assets = data["assets"]
if not isinstance(assets, dict) or not assets:
    raise SystemExit("manifest assets must be a non-empty object")

for key, name in assets.items():
    if not key or not name:
        raise SystemExit("manifest contains empty asset entry")

checksums = data.get("checksums") or f"{data['base_url'].rstrip('/')}/checksums.txt"
print(f"manifest ok: tag={data['tag']} version={data['version']} assets={len(assets)}")
print(f"checksums={checksums}")

out = Path(sys.argv[1]).with_name("verify.env")
out.write_text(
    "\n".join(
        [
            f"TAG={data['tag']}",
            f"VERSION={data['version']}",
            f"BASE_URL={data['base_url'].rstrip('/')}",
            f"CHECKSUMS_URL={checksums}",
        ]
    )
    + "\n",
    encoding="utf-8",
)
PY

# shellcheck disable=SC1091
source "$tmpdir/verify.env"

printf 'fetching checksums %s\n' "$CHECKSUMS_URL"
checksums_file="$tmpdir/checksums.txt"
nexus_curl -o "$checksums_file" "$CHECKSUMS_URL"

python3 - "$manifest_file" "$checksums_file" <<'PY'
import json
import sys
from pathlib import Path

manifest = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
checksums = Path(sys.argv[2]).read_text(encoding="utf-8")
lines = [line.strip() for line in checksums.splitlines() if line.strip()]

if not lines:
    raise SystemExit("checksums.txt is empty")

indexed = {}
for line in lines:
    parts = line.split()
    if len(parts) < 2:
        raise SystemExit(f"invalid checksum line: {line!r}")
    indexed[parts[-1]] = parts[0]

for name in manifest["assets"].values():
    if name not in indexed:
        raise SystemExit(f"checksums.txt missing entry for {name}")

if "checksums.txt" in indexed:
    raise SystemExit("checksums.txt should not contain a self-entry")

print(f"checksums ok: {len(lines)} entries")
PY

fail=0
while IFS= read -r asset_name; do
  asset_url="${BASE_URL}/${asset_name}"
  printf 'HEAD %s\n' "$asset_name"
  if ! nexus_curl -o /dev/null -I "$asset_url"; then
    printf 'missing asset: %s\n' "$asset_url" >&2
    fail=1
  fi
done < <(python3 -c '
import json, sys
manifest = json.load(open(sys.argv[1], encoding="utf-8"))
for name in manifest["assets"].values():
    print(name)
' "$manifest_file")

if [[ "$fail" -ne 0 ]]; then
  nexus_err "one or more assets are unreachable"
fi

if [[ -n "${MUXDEV_BIN:-}" && -x "$MUXDEV_BIN" ]]; then
  printf 'running update check via %s\n' "$MUXDEV_BIN"
  set +e
  update_out="$(MUXDEV_UPDATE_URL="$MANIFEST_URL" "$MUXDEV_BIN" update --check 2>&1)"
  check_code=$?
  set -e
  printf '%s\n' "$update_out"
  if [[ "$check_code" -ne 0 ]]; then
    if [[ "$update_out" == *"Update available:"* || "$update_out" == *"Up to date."* ]]; then
      check_code=0
    fi
  fi
  if [[ "$check_code" -ne 0 && "$check_code" -ne 2 ]]; then
    nexus_err "muxdev update --check failed (exit $check_code)"
  fi
fi

printf 'verify ok: %s\n' "$MANIFEST_URL"
