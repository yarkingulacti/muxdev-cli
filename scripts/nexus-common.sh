#!/usr/bin/env bash
# Shared helpers for Nexus release scripts.

nexus_repo_root() {
  cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd
}

nexus_load_env() {
  local root
  root="$(nexus_repo_root)"
  if [[ -f "$root/.env" ]]; then
    set -a
    # shellcheck disable=SC1091
    source "$root/.env"
    set +a
  fi
}

nexus_defaults() {
  nexus_load_env
  NEXUS_URL="${NEXUS_URL:-http://127.0.0.1:8081}"
  NEXUS_REPO="${NEXUS_REPO:-muxdev-releases}"
  NEXUS_CHANNEL="${NEXUS_CHANNEL:-stable}"
  DIST_DIR="${DIST_DIR:-dist}"
  NEXUS_AUTH="${NEXUS_AUTH:-}"
}

nexus_err() {
  printf 'nexus: %s\n' "$1" >&2
  exit 1
}

nexus_curl() {
  if [[ -n "${NEXUS_AUTH:-}" ]]; then
    curl -fsS -u "$NEXUS_AUTH" "$@"
  else
    curl -fsS "$@"
  fi
}

nexus_publish_base() {
  local nexus_url="$1"
  local nexus_repo="$2"
  local channel="$3"
  printf '%s/repository/%s/%s' "${nexus_url%/}" "$nexus_repo" "$channel"
}

nexus_release_base() {
  local publish_base="$1"
  local tag="$2"
  printf '%s/%s' "$publish_base" "$tag"
}

nexus_manifest_url() {
  local publish_base="$1"
  printf '%s/latest.json' "$publish_base"
}

nexus_generate_manifest() {
  local version="$1"
  local tag="$2"
  local base_url="$3"
  local dist_dir="$4"
  local output="$5"
  python3 "$(dirname "${BASH_SOURCE[0]}")/nexus-manifest.py" \
    "$version" "$tag" "$base_url" "$dist_dir" -o "$output"
}
