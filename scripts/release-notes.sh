#!/usr/bin/env bash
# Print GitHub release notes for TAG from CHANGELOG.md (Keep a Changelog / Release Please format).
set -euo pipefail

TAG="${1:-}"
if [[ -z "$TAG" ]]; then
  echo "usage: $0 vX.Y.Z" >&2
  exit 1
fi

VERSION="${TAG#v}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHANGELOG="${ROOT}/CHANGELOG.md"

if [[ ! -f "$CHANGELOG" ]]; then
  echo "missing $CHANGELOG" >&2
  exit 1
fi

python3 - "$CHANGELOG" "$VERSION" <<'PY'
import re
import sys
from pathlib import Path

path = Path(sys.argv[1])
version = sys.argv[2]
lines = path.read_text(encoding="utf-8").splitlines()

heading = re.compile(rf"^## \[{re.escape(version)}\]")
stop = re.compile(r"^## \[")
footer = re.compile(r"^\[[^\]]+\]:")

out: list[str] = []
capture = False
for line in lines:
    if footer.match(line):
        break
    if stop.match(line):
        if capture:
            break
        if heading.match(line):
            capture = True
            continue
    if capture:
        out.append(line)

text = "\n".join(out).strip()
if not text:
    raise SystemExit(f"no changelog section found for {version} in {path}")

print(text)
PY
