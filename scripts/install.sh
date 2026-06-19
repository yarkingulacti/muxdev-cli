#!/usr/bin/env bash
# Install or update muxdev from GitHub releases.
set -euo pipefail

REPO="yarkingulacti/muxdev-cli"
GITHUB_API="${GITHUB_API:-https://api.github.com}"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
VERSION="${VERSION:-latest}"
CHECK_ONLY="${CHECK_ONLY:-0}"

err() {
  printf 'install: %s\n' "$1" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"
}

detect_os() {
  case "$(uname -s)" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) err "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) err "unsupported architecture: $(uname -m)" ;;
  esac
}

resolve_version() {
  if [[ "$VERSION" != "latest" ]]; then
    printf '%s' "$VERSION"
    return
  fi
  need_cmd curl
  local api_url="${GITHUB_API}/repos/${REPO}/releases/latest"
  local curl_args=(-fsSL -H "Accept: application/vnd.github+json")
  if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    curl_args+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
  fi
  curl "${curl_args[@]}" "$api_url" \
    | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n1
}

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  else
    err "sha256sum or shasum required for checksum verification"
  fi
}

verify_checksum() {
  local archive="$1"
  local asset="$2"
  local checksums="$3"
  local expected actual

  expected="$(printf '%s\n' "$checksums" | awk -v asset="$asset" '$NF == asset { print $1; exit }')"
  if [[ -z "$expected" ]]; then
    err "checksum for $asset not found"
  fi
  actual="$(sha256_file "$archive")"
  if [[ "$actual" != "$expected" ]]; then
    err "checksum mismatch for $asset"
  fi
}

main() {
  need_cmd tar
  need_cmd curl

  local os arch tag asset url tmpdir binary checksums
  os="$(detect_os)"
  arch="$(detect_arch)"
  tag="$(resolve_version)"

  if [[ -z "$tag" ]]; then
    err "could not resolve release version"
  fi

  if [[ "$os" == "windows" ]]; then
    asset="muxdev_${tag#v}_${os}_${arch}.zip"
  else
    asset="muxdev_${tag#v}_${os}_${arch}.tar.gz"
  fi

  url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir:-}"' EXIT

  checksums="$(curl -fsSL -H "Accept: application/vnd.github+json" \
    "${GITHUB_API}/repos/${REPO}/releases/tags/${tag}" \
    | sed -n 's/.*"browser_download_url":[[:space:]]*"\([^"]*checksums.txt\)".*/\1/p' | head -n1)"
  if [ -z "$checksums" ]; then
    checksums="https://github.com/${REPO}/releases/download/${tag}/checksums.txt"
  fi
  checksums="$(curl -fsSL "$checksums")"

  if [[ "$CHECK_ONLY" == "1" ]]; then
    printf 'Latest release: %s (%s)\n' "$tag" "$asset"
    exit 0
  fi

  printf 'Downloading %s\n' "$url"
  curl -fsSL "$url" -o "${tmpdir}/archive"
  verify_checksum "${tmpdir}/archive" "$asset" "$checksums"

  if [[ "$os" == "windows" ]]; then
    need_cmd unzip
    unzip -q "${tmpdir}/archive" -d "$tmpdir"
    binary="muxdev.exe"
  else
    tar -xzf "${tmpdir}/archive" -C "$tmpdir"
    binary="muxdev"
  fi

  mkdir -p "$INSTALL_DIR"
  install -m 0755 "${tmpdir}/${binary}" "${INSTALL_DIR}/${binary}"

  printf 'Installed muxdev %s to %s\n' "$tag" "${INSTALL_DIR}/${binary}"
  if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
    printf 'Add to PATH: export PATH="%s:$PATH"\n' "$INSTALL_DIR"
  fi
}

main "$@"
