#!/usr/bin/env bash
#
# install.sh — download the pptx-to-md binary for this platform from GitHub
# Releases and install it. Asset naming mirrors ../.goreleaser.yml:
#
#   pptx-to-md_<version>_<os>_<arch>.tar.gz   (zip on windows)
#
# Environment overrides:
#   PPTX2MD_VERSION   pin a release tag (e.g. v1.2.3); default: latest
#   PPTX2MD_BIN_DIR   install location;               default: $HOME/.local/bin
#
set -euo pipefail

REPO="greenstevester/pptx2md-go"
BINARY="pptx-to-md"
INSTALL_DIR="${PPTX2MD_BIN_DIR:-$HOME/.local/bin}"

err() { printf 'error: %s\n' "$*" >&2; exit 1; }

# --- platform detection ---
case "$(uname -s)" in
  Darwin) os="darwin" ;;
  Linux)  os="linux" ;;
  *) err "unsupported OS '$(uname -s)'. On Windows, download the .zip from https://github.com/$REPO/releases/latest" ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) err "unsupported architecture '$(uname -m)'" ;;
esac

# --- resolve version ---
tag="${PPTX2MD_VERSION:-}"
if [ -z "$tag" ]; then
  tag="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null \
    | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name" *: *"([^"]+)".*/\1/' || true)"
fi
[ -n "$tag" ] || err "could not determine a release tag — has a release been published yet? See https://github.com/$REPO/releases"
version="${tag#v}"

asset="${BINARY}_${version}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$tag"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $asset ($tag)…"
curl -fsSL -o "$tmp/$asset" "$base/$asset" || err "download failed: $base/$asset"

# --- verify checksum (best effort) ---
if curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt" 2>/dev/null; then
  expected="$(awk -v f="$asset" '$2 == f {print $1}' "$tmp/checksums.txt")"
  if [ -n "$expected" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      actual="$(sha256sum "$tmp/$asset" | awk '{print $1}')"
    else
      actual="$(shasum -a 256 "$tmp/$asset" | awk '{print $1}')"
    fi
    [ "$actual" = "$expected" ] || err "checksum mismatch for $asset (expected $expected, got $actual)"
    echo "Checksum OK."
  else
    echo "warning: $asset not listed in checksums.txt; skipping verification" >&2
  fi
else
  echo "warning: checksums.txt unavailable; skipping verification" >&2
fi

# --- extract & install ---
tar -xzf "$tmp/$asset" -C "$tmp"
[ -f "$tmp/$BINARY" ] || err "$BINARY not found inside $asset"
mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"

echo "Installed $BINARY $version to $INSTALL_DIR/$BINARY"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "note: $INSTALL_DIR is not on your PATH — add: export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac
"$INSTALL_DIR/$BINARY" --help >/dev/null 2>&1 && echo "Verified: $BINARY runs."
