#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"
readonly INSTALL_DIR="$SERVER_DIR/.tools/bin"
readonly SQLC_BIN="$INSTALL_DIR/sqlc"

# shellcheck source=sqlc-common.sh
. "$SCRIPT_DIR/sqlc-common.sh"

installed_version() {
	if [ ! -x "$SQLC_BIN" ]; then
		return 1
	fi

	[ "$("$SQLC_BIN" version)" = "v$SQLC_VERSION" ]
}

if installed_version; then
	printf '%s\n' "$SQLC_BIN"
	exit 0
fi

case "$(uname -s)" in
	Darwin) os="darwin" ;;
	Linux) os="linux" ;;
	*)
		printf 'unsupported operating system: %s\n' "$(uname -s)" >&2
		exit 1
		;;
esac

case "$(uname -m)" in
	x86_64 | amd64) arch="amd64" ;;
	arm64 | aarch64) arch="arm64" ;;
	*)
		printf 'unsupported architecture: %s\n' "$(uname -m)" >&2
		exit 1
		;;
esac

case "$os/$arch" in
	darwin/amd64) checksum="$SQLC_SHA256_DARWIN_AMD64" ;;
	darwin/arm64) checksum="$SQLC_SHA256_DARWIN_ARM64" ;;
	linux/amd64) checksum="$SQLC_SHA256_LINUX_AMD64" ;;
	linux/arm64) checksum="$SQLC_SHA256_LINUX_ARM64" ;;
esac

readonly archive="sqlc_${SQLC_VERSION}_${os}_${arch}.tar.gz"
readonly download_url="https://github.com/sqlc-dev/sqlc/releases/download/v${SQLC_VERSION}/${archive}"
temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-sqlc.XXXXXX")"
trap 'rm -rf -- "$temporary_dir"' EXIT HUP INT TERM

curl --fail --location --silent --show-error "$download_url" --output "$temporary_dir/$archive"

if command -v sha256sum >/dev/null 2>&1; then
	actual_checksum="$(sha256sum "$temporary_dir/$archive" | awk '{print $1}')"
else
	actual_checksum="$(shasum -a 256 "$temporary_dir/$archive" | awk '{print $1}')"
fi

if [ "$actual_checksum" != "$checksum" ]; then
	printf 'sqlc archive checksum mismatch: got %s, want %s\n' "$actual_checksum" "$checksum" >&2
	exit 1
fi

tar -xzf "$temporary_dir/$archive" -C "$temporary_dir"
mkdir -p "$INSTALL_DIR"
install -m 0755 "$temporary_dir/sqlc" "$SQLC_BIN"

if ! installed_version; then
	printf 'installed sqlc did not report v%s\n' "$SQLC_VERSION" >&2
	exit 1
fi

printf '%s\n' "$SQLC_BIN"
