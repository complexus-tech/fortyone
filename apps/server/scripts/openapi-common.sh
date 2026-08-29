#!/bin/sh

set -eu

readonly OAPI_CODEGEN_VERSION="2.8.0"
readonly OAPI_CODEGEN_BIN="${OAPI_CODEGEN_BIN:-$SERVER_DIR/.tools/bin/oapi-codegen}"

require_oapi_codegen() {
	if [ ! -x "$OAPI_CODEGEN_BIN" ]; then
		printf 'oapi-codegen is not installed; run make oapi-bootstrap\n' >&2
		exit 1
	fi
	installed_version="$($OAPI_CODEGEN_BIN -version 2>/dev/null | tail -n 1)"
	if [ "$installed_version" != "v$OAPI_CODEGEN_VERSION" ]; then
		printf 'oapi-codegen version mismatch: got %s, want v%s\n' "$installed_version" "$OAPI_CODEGEN_VERSION" >&2
		exit 1
	fi
}

# normalize_generated_openapi_go keeps generated Go compatible with the
# repository's staticcheck policy. oapi-codegen v2.8.0 emits one capitalized
# internal error string from its std-http middleware template; the outer
# RequiredHeaderError and its public response text remain unchanged.
normalize_generated_openapi_go() {
	generated_path="$1"
	normalized_path="${generated_path}.normalized"
	sed 's/fmt.Errorf("Header parameter /fmt.Errorf("header parameter /g' \
		"$generated_path" > "$normalized_path"
	mv "$normalized_path" "$generated_path"
	gofmt -w "$generated_path"
}
