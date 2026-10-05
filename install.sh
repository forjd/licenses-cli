#!/bin/sh
# Install the licenses CLI on macOS or Linux.
#   curl -fsSL https://raw.githubusercontent.com/forjd/licenses-cli/main/install.sh | sh
# Env: LICENSES_INSTALL_DIR (default /usr/local/bin if writable, else ~/.local/bin)
#      LICENSES_VERSION     (default latest, e.g. v0.1.0)
#      LICENSES_DOWNLOAD_URL (override the download base URL, e.g. a mirror;
#                             checksums.txt is fetched from the same place)
set -eu

# Everything runs from main, called on the last line, so a truncated
# download fails to parse instead of running half the script.
main() {
  repo="forjd/licenses-cli"
  version="${LICENSES_VERSION:-latest}"
  case "$version" in
    latest | v*) ;;
    *) version="v$version" ;;
  esac

  case "$(uname -s)" in
    Linux) os=linux ;;
    Darwin) os=darwin ;;
    *) die "Unsupported OS: $(uname -s). On Windows, use install.ps1." ;;
  esac
  case "$(uname -m)" in
    x86_64 | amd64) arch=amd64 ;;
    arm64 | aarch64) arch=arm64 ;;
    *) die "Unsupported architecture: $(uname -m)" ;;
  esac
  # An x86_64 shell under Rosetta on Apple Silicon: install the native build.
  if [ "$os" = darwin ] && [ "$arch" = amd64 ] &&
    [ "$(sysctl -n sysctl.proc_translated 2>/dev/null || true)" = 1 ]; then
    arch=arm64
  fi

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
    dir="${HOME:?HOME is not set; set LICENSES_INSTALL_DIR}/.local/bin"
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    sha256() { sha256sum "$1"; }
  elif command -v shasum >/dev/null 2>&1; then
    sha256() { shasum -a 256 "$1"; }
  else
    die "Need sha256sum or shasum to verify the download"
  fi
  if command -v curl >/dev/null 2>&1; then
    fetch() { curl -fsSL "$1" -o "$2"; }
  elif command -v wget >/dev/null 2>&1; then
    fetch() { wget -qO "$2" "$1"; }
  else
    die "Need curl or wget to download"
  fi

  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  # dash skips the EXIT trap when killed by a signal, so exit explicitly.
  trap 'exit 130' INT TERM HUP

  echo "Downloading $archive..."
  fetch "$base/$archive" "$tmp/$archive" || die "Download failed: $base/$archive"
  fetch "$base/checksums.txt" "$tmp/checksums.txt" || die "Download failed: $base/checksums.txt"

  want="$(awk -v f="$archive" '$2 == f { print $1 }' "$tmp/checksums.txt")"
  got="$(sha256 "$tmp/$archive" | cut -d' ' -f1)"
  if [ -z "$want" ] || [ "$want" != "$got" ]; then
    die "Checksum mismatch for $archive"
  fi

  tar -xzf "$tmp/$archive" -C "$tmp" licenses
  mkdir -p "$dir"
  mv "$tmp/licenses" "$dir/licenses"
  chmod +x "$dir/licenses"

  echo "Installed $("$dir/licenses" -version) to $dir/licenses"
  case ":$PATH:" in
    *":$dir:"*)
      found="$(command -v licenses 2>/dev/null || true)"
      if [ -n "$found" ] && [ "$found" != "$dir/licenses" ]; then
        echo "Note: $found comes first on your PATH. Remove it to use this install."
      fi
      ;;
    *) echo "Note: $dir is not on your PATH. Add it to your shell profile." ;;
  esac
}

die() {
  echo "$1" >&2
  exit 1
}

main "$@"
