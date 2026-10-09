#!/bin/sh
# Install the latest whyd release on any Linux with systemd:
#   curl -fsSL https://raw.githubusercontent.com/vmaltarello/whyd/main/install.sh | sh
# Installs to /usr/local/bin (asks for sudo) or, with PREFIX=~/.local, to ~/.local/bin.
# WHYD_VERSION=v0.1.0 installs a specific release.
set -eu

repo="vmaltarello/whyd"
prefix="${PREFIX:-/usr/local}"
version="${WHYD_VERSION:-latest}"

die() { echo "whyd install: $*" >&2; exit 1; }

[ "$(uname -s)" = Linux ] || die "whyd runs on Linux only"
command -v journalctl >/dev/null 2>&1 || echo "whyd install: journalctl not found; whyd needs systemd" >&2

case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) die "no release for $(uname -m); build from source: go install github.com/$repo/cmd/whyd@latest" ;;
esac

if [ "$version" = latest ]; then
  base="https://github.com/$repo/releases/latest/download"
else
  base="https://github.com/$repo/releases/download/$version"
fi

if command -v curl >/dev/null 2>&1; then
  get() { curl -fsSL -o "$2" "$1"; }
elif command -v wget >/dev/null 2>&1; then
  get() { wget -q -O "$2" "$1"; }
else
  die "needs curl or wget"
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
file="whyd_linux_$arch.tar.gz"

echo "downloading $file ($version)"
get "$base/$file" "$tmp/$file" || die "download failed: $base/$file"
get "$base/checksums.txt" "$tmp/checksums.txt" || die "download failed: $base/checksums.txt"
(cd "$tmp" && grep " $file\$" checksums.txt | sha256sum -c --quiet -) || die "checksum mismatch"
tar -xzf "$tmp/$file" -C "$tmp" whyd

dir="$prefix/bin"
if mkdir -p "$dir" 2>/dev/null && [ -w "$dir" ]; then
  install -m 755 "$tmp/whyd" "$dir/whyd"
else
  echo "installing to $dir (sudo)"
  sudo install -D -m 755 "$tmp/whyd" "$dir/whyd"
fi

echo "installed $("$dir/whyd" --version) to $dir/whyd"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "add $dir to your PATH to run it as whyd" ;;
esac
