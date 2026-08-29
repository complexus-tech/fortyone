#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"

# shellcheck source=sqlc-common.sh
. "$SCRIPT_DIR/sqlc-common.sh"
require_sqlc

if [ -z "${SQLC_DATABASE_URL:-}" ]; then
	printf 'SQLC_DATABASE_URL is required for database-backed sqlc vet\n' >&2
	exit 1
fi

cd "$SERVER_DIR"
go run ./internal/tools/migrationstate \
	-migrations internal/migrations
"$SQLC_BIN" vet -f sqlc.yaml
