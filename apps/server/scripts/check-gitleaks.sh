#!/usr/bin/env bash

set -euo pipefail

readonly server_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly repository_root="$(git -C "${server_root}" rev-parse --show-toplevel)"
readonly gitleaks_bin="${GITLEAKS_BIN:-${server_root}/.tools/bin/gitleaks}"
readonly indexed_server_path="/apps/server"

if [[ ! -x "${gitleaks_bin}" ]]; then
  echo "gitleaks is not installed at ${gitleaks_bin}; run make security-bootstrap" >&2
  exit 1
fi

temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-gitleaks-index.XXXXXX")"
readonly temporary_dir
trap 'rm -rf -- "${temporary_dir}"' EXIT

# Scan the Git index, not the developer's filesystem. This includes staged
# changes and excludes ignored local .env files. CI checks out the pull-request
# commit into the index, so the complete reviewed API tree is scanned.
git -C "${repository_root}" checkout-index --all --prefix="${temporary_dir}/"

"${gitleaks_bin}" dir \
  --redact \
  --no-banner \
  --no-color \
  "${temporary_dir}${indexed_server_path}"
