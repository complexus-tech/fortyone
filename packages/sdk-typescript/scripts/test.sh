#!/bin/sh

set -eu

readonly SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
readonly PACKAGE_DIR="$(dirname -- "$SCRIPT_DIR")"
temporary_dir="$(mktemp -d "$PACKAGE_DIR/.sdk-test.XXXXXX")"
trap 'rm -rf -- "$temporary_dir"' EXIT HUP INT TERM

"$PACKAGE_DIR/node_modules/.bin/tsc" \
	--project "$PACKAGE_DIR/tsconfig.test.json" \
	--outDir "$temporary_dir"
node --test "$temporary_dir"/src/*.test.js
