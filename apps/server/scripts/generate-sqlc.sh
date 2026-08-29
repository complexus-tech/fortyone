#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly SERVER_DIR="$(dirname -- "$SCRIPT_DIR")"

# shellcheck source=sqlc-common.sh
. "$SCRIPT_DIR/sqlc-common.sh"
require_sqlc

cd "$SERVER_DIR"
temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-sqlc-generate.XXXXXX")"
inputs_file="$temporary_dir/inputs"
outputs_file="$temporary_dir/outputs"
replaced_outputs_file="$temporary_dir/replaced-outputs"
staging_paths_owned=false
generation_committed=false

cleanup() {
	if [ "$staging_paths_owned" = true ] && [ -f "$outputs_file" ]; then
		if [ "$generation_committed" != true ]; then
			while IFS= read -r output; do
				parent="$(dirname -- "$output")"
				base="$(basename -- "$output")"
				backup="$parent/.${base}.sqlc-backup.$$"
				if [ -e "$backup" ]; then
					rm -rf -- "$output"
					mv "$backup" "$output"
				elif [ -f "$replaced_outputs_file" ] && grep -Fqx "$output" "$replaced_outputs_file"; then
					rm -rf -- "$output"
				fi
			done <"$outputs_file"
		fi
		while IFS= read -r output; do
			parent="$(dirname -- "$output")"
			base="$(basename -- "$output")"
			staged="$parent/.${base}.sqlc-staged.$$"
			backup="$parent/.${base}.sqlc-backup.$$"
			rm -rf -- "$backup"
			rm -rf -- "$staged"
		done <"$outputs_file"
	fi
	rm -rf -- "$temporary_dir"
}
trap cleanup EXIT HUP INT TERM

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

# Validate the complete temporary result before touching checked-in output.
go run ./internal/tools/sqlcconfig -config "$temporary_dir/sqlc.yaml" validate

# Stage every generated directory beside its destination first. Replacement is
# then a same-filesystem rename, with the old directory available for rollback.
while IFS= read -r output; do
	parent="$(dirname -- "$output")"
	base="$(basename -- "$output")"
	staged="$parent/.${base}.sqlc-staged.$$"
	backup="$parent/.${base}.sqlc-backup.$$"
	if [ -e "$staged" ] || [ -e "$backup" ]; then
		printf 'refusing to reuse an existing SQLC staging or backup path for %s\n' "$output" >&2
		exit 1
	fi
done <"$outputs_file"
staging_paths_owned=true

while IFS= read -r output; do
	parent="$(dirname -- "$output")"
	base="$(basename -- "$output")"
	staged="$parent/.${base}.sqlc-staged.$$"
	cp -R "$temporary_dir/$output" "$staged"
done <"$outputs_file"

while IFS= read -r output; do
	parent="$(dirname -- "$output")"
	base="$(basename -- "$output")"
	staged="$parent/.${base}.sqlc-staged.$$"
	backup="$parent/.${base}.sqlc-backup.$$"
	if [ -e "$output" ]; then
		mv "$output" "$backup"
	fi
	# Record ownership before installing the staged directory. On interruption,
	# cleanup can now distinguish a newly created output from an untouched one.
	printf '%s\n' "$output" >>"$replaced_outputs_file"
	if ! mv "$staged" "$output"; then
		exit 1
	fi
done <"$outputs_file"

# Backups remain available until every configured output has been replaced.
# The cleanup trap removes them only after this commit point.
generation_committed=true
