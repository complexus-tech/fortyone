#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"
: "${MIGRATE_BIN:?MIGRATE_BIN must identify the golang-migrate executable}"

if [ "$#" -lt 1 ]; then
	printf 'usage: migrate.sh <up|down|version|force> [argument]\n' >&2
	exit 2
fi

readonly action="$1"
shift
case "$action" in
	up | down | version | force) ;;
	*)
		printf 'unsupported migration action: %s\n' "$action" >&2
		exit 2
		;;
esac

cd "$SERVER_DIR"
database_url="${DB_URL:-}"
if [ -z "$database_url" ]; then
	database_url="$(go run ./internal/tools/databaseurl)"
fi

"$MIGRATE_BIN" -path internal/migrations -database "$database_url" "$action" "$@"
