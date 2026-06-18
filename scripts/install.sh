#!/usr/bin/env bash
# Install muxdev from GitHub releases.
set -euo pipefail

REPO="yarkingulacti/muxdev-cli"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
VERSION="${VERSION:-latest}"

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
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | head -n1 \
    | cut -d'"' -f4
}

main() {
  need_cmd tar
  need_cmd curl

  local os arch tag asset url tmpdir binary
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
  trap 'rm -rf "$tmpdir"' EXIT

  printf 'Downloading %s\n' "$url"
  curl -fsSL "$url" -o "${tmpdir}/archive"

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
