#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"

# shellcheck source=openapi-common.sh
. "$SCRIPT_DIR/openapi-common.sh"
# shellcheck source=sdk-common.sh
. "$SCRIPT_DIR/sdk-common.sh"
require_sdk_tools

temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-sdk-check.XXXXXX")"
trap 'rm -rf -- "$temporary_dir"' EXIT HUP INT TERM
generate_sdk_outputs "$temporary_dir"

diff -u "$GO_SDK_DIR/client.gen.go" "$temporary_dir/$SDK_GO_CLIENT_OUTPUT"
diff -u "$GO_SDK_DIR/metadata.gen.go" "$temporary_dir/$SDK_GO_METADATA_OUTPUT"
diff -u "$TYPESCRIPT_SDK_DIR/src/generated/schema.d.ts" "$temporary_dir/$SDK_TYPESCRIPT_SCHEMA_OUTPUT"
diff -u "$TYPESCRIPT_SDK_DIR/src/generated/metadata.ts" "$temporary_dir/$SDK_TYPESCRIPT_METADATA_OUTPUT"

(
	cd "$GO_SDK_DIR"
	go mod verify
	go mod tidy -diff
	go test -race -count=1 ./...
)
(
	cd "$EXTERNAL_SAMPLE_DIR"
	go mod verify
	go mod tidy -diff
	go test -race -count=1 ./...
)
pnpm --dir "$TYPESCRIPT_SDK_DIR" type-check
pnpm --dir "$TYPESCRIPT_SDK_DIR" test
