#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"

# shellcheck source=openapi-common.sh
. "$SCRIPT_DIR/openapi-common.sh"

if [ -x "$OAPI_CODEGEN_BIN" ] && [ "$($OAPI_CODEGEN_BIN -version 2>/dev/null | tail -n 1)" = "v$OAPI_CODEGEN_VERSION" ]; then
	printf '%s\n' "$OAPI_CODEGEN_BIN"
	exit 0
fi

temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-oapi-codegen.XXXXXX")"
trap 'rm -rf -- "$temporary_dir"' EXIT HUP INT TERM

GOBIN="$temporary_dir" go install "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v$OAPI_CODEGEN_VERSION"
mkdir -p "$(dirname -- "$OAPI_CODEGEN_BIN")"
install -m 0755 "$temporary_dir/oapi-codegen" "$OAPI_CODEGEN_BIN"
require_oapi_codegen

printf '%s\n' "$OAPI_CODEGEN_BIN"
