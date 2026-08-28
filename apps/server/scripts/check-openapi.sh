#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"
readonly CONFIG_PATH="api/openapi/v1/oapi-codegen.yaml"
readonly SPEC_PATH="api/openapi/v1/openapi.yaml"
readonly BUNDLE_PATH="api/openapi/v1/openapi.bundle.json"
readonly OUTPUT_PATH="internal/generated/openapi/v1/api.gen.go"

# shellcheck source=openapi-common.sh
. "$SCRIPT_DIR/openapi-common.sh"
require_oapi_codegen

temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-openapi-check.XXXXXX")"
trap 'rm -rf -- "$temporary_dir"' EXIT HUP INT TERM

cd "$SERVER_DIR"
mkdir -p "$temporary_dir/api/openapi"
cp -R api/openapi/v1 "$temporary_dir/api/openapi/v1"
cp go.mod go.sum "$temporary_dir/"
go run ./internal/tools/openapibundle \
	-input "$temporary_dir/$SPEC_PATH" \
	-output "$temporary_dir/$BUNDLE_PATH"
(
	cd "$temporary_dir"
	"$OAPI_CODEGEN_BIN" -config "$CONFIG_PATH" "$BUNDLE_PATH"
)
normalize_generated_openapi_go "$temporary_dir/$OUTPUT_PATH"
diff -u "$OUTPUT_PATH" "$temporary_dir/$OUTPUT_PATH"
