#!/bin/sh

if [ -z "${SERVER_DIR:-}" ]; then
	printf 'SERVER_DIR must be set before loading sqlc-common.sh\n' >&2
	exit 1
fi

# shellcheck source=../tools/sqlc.lock
. "$SERVER_DIR/tools/sqlc.lock"

if [ -z "${SQLC_BIN+x}" ]; then
	SQLC_BIN="$SERVER_DIR/.tools/bin/sqlc"
fi

require_sqlc() {
	if [ "$("$SQLC_BIN" version 2>/dev/null || true)" != "v$SQLC_VERSION" ]; then
		printf 'sqlc v%s is required; run make sqlc-bootstrap\n' "$SQLC_VERSION" >&2
		exit 1
	fi
}
