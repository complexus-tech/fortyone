#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"
readonly REPOSITORY_DIR="$(git -C "$SERVER_DIR" rev-parse --show-toplevel)"
readonly BASE_REF="${OPENAPI_BASE_REF:-}"
readonly OASDIFF_VERSION="${OASDIFF_VERSION:-}"
readonly OASDIFF_BIN="${OASDIFF_BIN:-}"
readonly CONTRACT_PATH="apps/server/api/openapi/v1/openapi.yaml"

if [ -z "$BASE_REF" ]; then
	printf 'OPENAPI_BASE_REF is required (use the pull-request base commit SHA)\n' >&2
	exit 1
fi
if [ -z "$OASDIFF_VERSION" ]; then
	printf 'OASDIFF_VERSION is required\n' >&2
	exit 1
fi
if [ ! -x "$OASDIFF_BIN" ] || ! "$OASDIFF_BIN" --version 2>/dev/null | grep -Fq "${OASDIFF_VERSION#v}"; then
	printf 'oasdiff %s is required; run make oasdiff-bootstrap\n' "$OASDIFF_VERSION" >&2
	exit 1
fi

# Resolve first so an option-shaped or ambiguous caller value never reaches
# git archive. CI supplies an immutable pull-request base SHA.
resolved_base="$(git -C "$REPOSITORY_DIR" rev-parse --verify --end-of-options "${BASE_REF}^{commit}")"

temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-openapi-breaking.XXXXXX")"
trap 'rm -rf -- "$temporary_dir"' EXIT HUP INT TERM

if ! git -C "$REPOSITORY_DIR" cat-file -e "${resolved_base}:${CONTRACT_PATH}" 2>/dev/null; then
	printf 'No API v1 contract exists at base commit %s; compatibility check is not applicable.\n' "$resolved_base"
	exit 0
fi

git -C "$REPOSITORY_DIR" archive "$resolved_base" -- "$(dirname -- "$CONTRACT_PATH")" \
	| tar -x -C "$temporary_dir"

base_spec="$temporary_dir/$CONTRACT_PATH"
revision_spec="$SERVER_DIR/api/openapi/v1/openapi.yaml"

# ERR and WARN changes both require an explicit compatibility decision. The
# command receives only resolved/local paths after --, avoiding option parsing.
"$OASDIFF_BIN" breaking \
	--fail-on WARN \
	--format githubactions \
	-- "$base_spec" "$revision_spec"
