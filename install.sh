#!/bin/sh
# Install the licenses CLI on macOS or Linux.
#   curl -fsSL https://raw.githubusercontent.com/forjd/licenses-cli/main/install.sh | sh
# Env: LICENSES_INSTALL_DIR (default /usr/local/bin if writable, else ~/.local/bin)
#      LICENSES_VERSION     (default latest, e.g. v0.1.0)
#      LICENSES_DOWNLOAD_URL (override the download base URL, e.g. a mirror)
set -eu

repo="forjd/licenses-cli"
version="${LICENSES_VERSION:-latest}"

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) echo "Unsupported OS: $(uname -s). On Windows, use install.ps1." >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

if [ "$version" = latest ]; then
  base="https://github.com/$repo/releases/latest/download"
else
  base="https://github.com/$repo/releases/download/$version"
fi
base="${LICENSES_DOWNLOAD_URL:-$base}"
archive="licenses_${os}_${arch}.tar.gz"

if [ -n "${LICENSES_INSTALL_DIR:-}" ]; then
  dir="$LICENSES_INSTALL_DIR"
elif [ -w /usr/local/bin ]; then
  dir=/usr/local/bin
else
  dir="$HOME/.local/bin"
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $archive..."
curl -fsSL "$base/$archive" -o "$tmp/$archive"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt"

want="$(grep " $archive\$" "$tmp/checksums.txt" | cut -d' ' -f1)"
if command -v sha256sum >/dev/null 2>&1; then
  got="$(sha256sum "$tmp/$archive" | cut -d' ' -f1)"
else
  got="$(shasum -a 256 "$tmp/$archive" | cut -d' ' -f1)"
fi
if [ -z "$want" ] || [ "$want" != "$got" ]; then
  echo "Checksum mismatch for $archive" >&2
  exit 1
fi

tar -xzf "$tmp/$archive" -C "$tmp" licenses
mkdir -p "$dir"
mv "$tmp/licenses" "$dir/licenses"
chmod +x "$dir/licenses"

echo "Installed $("$dir/licenses" -version) to $dir/licenses"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "Note: $dir is not on your PATH. Add it to your shell profile." ;;
esac
