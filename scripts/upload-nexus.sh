#!/usr/bin/env bash
# Upload Goreleaser dist/ artifacts to a Nexus raw repository and publish latest.json.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=nexus-common.sh
source "$SCRIPT_DIR/nexus-common.sh"

usage() {
  cat <<EOF
Usage: $0 <version>

Upload muxdev release artifacts from \${DIST_DIR}/ to Nexus.

Environment:
  NEXUS_URL       Base URL (default: http://5.178.111.150:8081)
  NEXUS_REPO      Repository name (default: muxdev-releases)
  NEXUS_CHANNEL   Channel path (default: stable)
  DIST_DIR        Goreleaser output directory (default: dist)
  NEXUS_AUTH      user:password for Basic auth (required for upload)

Example:
  NEXUS_AUTH=deploy:secret $0 v1.0.0
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

TAG="${1:-}"
if [[ -z "$TAG" ]]; then
  usage
  exit 1
fi

nexus_defaults

VERSION="${TAG#v}"
if [[ ! -d "$DIST_DIR" ]]; then
  nexus_err "missing dist directory: $DIST_DIR"
fi

PUBLISH_BASE="$(nexus_publish_base "$NEXUS_URL" "$NEXUS_REPO" "$NEXUS_CHANNEL")"
BASE="$(nexus_release_base "$PUBLISH_BASE" "$TAG")"

shopt -s nullglob
files=(
  "${DIST_DIR}/muxdev_${VERSION}_"*.tar.gz
  "${DIST_DIR}/muxdev_${VERSION}_"*.zip
  "${DIST_DIR}/checksums.txt"
)
if [[ ${#files[@]} -eq 0 ]]; then
  nexus_err "no release artifacts found in ${DIST_DIR} for version ${VERSION}"
fi

for f in "${files[@]}"; do
  name="$(basename "$f")"
  printf 'uploading %s\n' "$name"
  if ! nexus_curl --upload-file "$f" "${BASE}/${name}"; then
    nexus_err "upload failed for ${name} — set NEXUS_AUTH=user:pass"
  fi
done

manifest="$(mktemp)"
trap 'rm -f "$manifest"' EXIT

nexus_generate_manifest "$VERSION" "$TAG" "$BASE" "$DIST_DIR" "$manifest"

printf 'uploading manifest.json\n'
if ! nexus_curl --upload-file "$manifest" "${BASE}/manifest.json"; then
  nexus_err "upload failed for manifest.json"
fi

printf 'uploading latest.json\n'
if ! nexus_curl --upload-file "$manifest" "${PUBLISH_BASE}/latest.json"; then
  nexus_err "upload failed for latest.json"
fi

printf 'done: %s/latest.json\n' "$PUBLISH_BASE"
