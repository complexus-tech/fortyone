#!/bin/sh

if [ -z "${SERVER_DIR:-}" ]; then
	printf 'SERVER_DIR must be set before loading sdk-common.sh\n' >&2
	exit 1
fi

readonly WORKSPACE_DIR="$(CDPATH= cd -- "$SERVER_DIR/../.." && pwd)"
readonly TYPESCRIPT_SDK_DIR="$WORKSPACE_DIR/packages/sdk-typescript"
readonly GO_SDK_DIR="$SERVER_DIR/sdk/go"
readonly EXTERNAL_SAMPLE_DIR="$SERVER_DIR/examples/external-integration"
readonly SDK_CLIENT_CONFIG="api/openapi/v1/oapi-codegen-client.yaml"
readonly SDK_SPEC_PATH="api/openapi/v1/openapi.yaml"
readonly SDK_BUNDLE_PATH="api/openapi/v1/openapi.bundle.json"
readonly SDK_GO_CLIENT_OUTPUT="sdk/go/client.gen.go"
readonly SDK_GO_METADATA_OUTPUT="sdk/go/metadata.gen.go"
readonly SDK_TYPESCRIPT_SCHEMA_OUTPUT="packages/sdk-typescript/src/generated/schema.d.ts"
readonly SDK_TYPESCRIPT_METADATA_OUTPUT="packages/sdk-typescript/src/generated/metadata.ts"

# shellcheck source=../tools/sdk.lock
. "$SERVER_DIR/tools/sdk.lock"

require_sdk_tools() {
	require_oapi_codegen
	if ! command -v pnpm >/dev/null 2>&1; then
		printf 'pnpm is required to generate the TypeScript SDK\n' >&2
		exit 1
	fi
	if [ ! -x "$TYPESCRIPT_SDK_DIR/node_modules/.bin/openapi-typescript" ]; then
		printf 'TypeScript SDK dependencies are missing; run pnpm install from the repository root\n' >&2
		exit 1
	fi
	installed_version="$(pnpm --dir "$TYPESCRIPT_SDK_DIR" exec openapi-typescript --version 2>/dev/null | tail -n 1 | sed 's/^v//')"
	if [ "$installed_version" != "$OPENAPI_TYPESCRIPT_VERSION" ]; then
		printf 'openapi-typescript version mismatch: got %s, want %s\n' "$installed_version" "$OPENAPI_TYPESCRIPT_VERSION" >&2
		exit 1
	fi
}

generate_sdk_outputs() {
	output_root="$1"
	mkdir -p \
		"$output_root/sdk/go" \
		"$output_root/packages/sdk-typescript/src/generated" \
		"$output_root/api/openapi/v1"

	(
		cd "$SERVER_DIR"
		go run ./internal/tools/openapibundle \
			-input "$SDK_SPEC_PATH" \
			-output "$output_root/$SDK_BUNDLE_PATH"
	)
	(
		cd "$output_root"
		"$OAPI_CODEGEN_BIN" \
			-config "$SERVER_DIR/$SDK_CLIENT_CONFIG" \
			"$output_root/$SDK_BUNDLE_PATH"
	)
	pnpm --dir "$TYPESCRIPT_SDK_DIR" exec openapi-typescript \
		"$output_root/$SDK_BUNDLE_PATH" \
		--output "$output_root/$SDK_TYPESCRIPT_SCHEMA_OUTPUT" \
		--alphabetize \
		--read-write-markers
	(
		cd "$SERVER_DIR"
		go run ./internal/tools/sdkmetadata \
			-generated-root "$output_root" \
			-workspace-root "$WORKSPACE_DIR" \
			-input "$SDK_BUNDLE_PATH" \
			-go-output "$SDK_GO_METADATA_OUTPUT" \
			-typescript-output "$SDK_TYPESCRIPT_METADATA_OUTPUT" \
			-typescript-package "packages/sdk-typescript/package.json"
	)
	gofmt -w "$output_root/$SDK_GO_CLIENT_OUTPUT" "$output_root/$SDK_GO_METADATA_OUTPUT"
}
