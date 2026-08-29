#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"

# shellcheck source=openapi-common.sh
. "$SCRIPT_DIR/openapi-common.sh"
# shellcheck source=sdk-common.sh
. "$SCRIPT_DIR/sdk-common.sh"
require_sdk_tools

temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-sdk-generate.XXXXXX")"
trap 'rm -rf -- "$temporary_dir"' EXIT HUP INT TERM
generate_sdk_outputs "$temporary_dir"

for output in \
	"$SDK_GO_CLIENT_OUTPUT" \
	"$SDK_GO_METADATA_OUTPUT" \
	"$SDK_TYPESCRIPT_SCHEMA_OUTPUT" \
	"$SDK_TYPESCRIPT_METADATA_OUTPUT"
do
	destination="$WORKSPACE_DIR/$output"
	if [ "${output#sdk/go/}" != "$output" ]; then
		destination="$SERVER_DIR/$output"
	fi
	mkdir -p "$(dirname -- "$destination")"
	install -m 0644 "$temporary_dir/$output" "$destination"
done
