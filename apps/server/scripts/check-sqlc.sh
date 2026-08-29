#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"

# shellcheck source=sqlc-common.sh
. "$SCRIPT_DIR/sqlc-common.sh"
require_sqlc

cd "$SERVER_DIR"
go run ./internal/tools/sqlcconfig -config sqlc.yaml validate
"$SQLC_BIN" compile -f sqlc.yaml

temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-sqlc-check.XXXXXX")"
inputs_file="$temporary_dir/inputs"
outputs_file="$temporary_dir/outputs"
trap 'rm -rf -- "$temporary_dir"' EXIT HUP INT TERM

go run ./internal/tools/sqlcconfig -config sqlc.yaml inputs >"$inputs_file"
go run ./internal/tools/sqlcconfig -config sqlc.yaml outputs >"$outputs_file"
cp sqlc.yaml "$temporary_dir/sqlc.yaml"

while IFS= read -r input; do
	destination="$temporary_dir/$input"
	mkdir -p "$(dirname -- "$destination")"
	cp -R "$input" "$destination"
done <"$inputs_file"

(
	cd "$temporary_dir"
	"$SQLC_BIN" generate -f sqlc.yaml
)

while IFS= read -r output; do
	diff -ru "$output" "$temporary_dir/$output"
done <"$outputs_file"
