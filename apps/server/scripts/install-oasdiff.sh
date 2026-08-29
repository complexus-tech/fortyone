#!/usr/bin/env bash

set -euo pipefail

readonly version="1.28.0"
readonly install_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/.tools/bin"
readonly binary_path="${install_dir}/oasdiff"

installed_version() {
  [[ -x "${binary_path}" ]] && "${binary_path}" --version 2>/dev/null | grep -Fq "${version}"
}

if installed_version; then
  printf '%s\n' "${binary_path}"
  exit 0
fi

case "$(uname -s)-$(uname -m)" in
  Linux-x86_64 | Linux-amd64)
    readonly asset="oasdiff_${version}_linux_amd64.tar.gz"
    readonly expected_sha256="e0ef076f2cf953d922addc04be9c3851cf3ec18f7678d2b94d44cea23dca51b5"
    ;;
  Linux-aarch64 | Linux-arm64)
    readonly asset="oasdiff_${version}_linux_arm64.tar.gz"
    readonly expected_sha256="cb15a381472321ac602cc252e65018d03feba7e6449a0854e1181680444d4051"
    ;;
  Darwin-x86_64 | Darwin-arm64)
    readonly asset="oasdiff_${version}_darwin_all.tar.gz"
    readonly expected_sha256="ff76474bf47bfb806d1711aa3e962b8e55570badcd462fa487b80aa532a823db"
    ;;
  *)
    echo "unsupported platform for pinned oasdiff: $(uname -s) $(uname -m)" >&2
    exit 1
    ;;
esac

readonly download_url="https://github.com/oasdiff/oasdiff/releases/download/v${version}/${asset}"
temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-oasdiff.XXXXXX")"
readonly temporary_dir
trap 'rm -rf -- "${temporary_dir}"' EXIT

curl --fail --silent --show-error --location \
  --proto '=https' --tlsv1.2 \
  --output "${temporary_dir}/${asset}" \
  "${download_url}"

if command -v sha256sum >/dev/null 2>&1; then
  actual_sha256="$(sha256sum "${temporary_dir}/${asset}" | awk '{print $1}')"
else
  actual_sha256="$(shasum -a 256 "${temporary_dir}/${asset}" | awk '{print $1}')"
fi
readonly actual_sha256

if [[ "${actual_sha256}" != "${expected_sha256}" ]]; then
  echo "oasdiff checksum mismatch: expected ${expected_sha256}, received ${actual_sha256}" >&2
  exit 1
fi

tar -xzf "${temporary_dir}/${asset}" -C "${temporary_dir}" oasdiff
mkdir -p "${install_dir}"
install -m 0755 "${temporary_dir}/oasdiff" "${binary_path}.tmp"
mv -f "${binary_path}.tmp" "${binary_path}"

if ! installed_version; then
  echo "installed oasdiff did not report expected version ${version}" >&2
  exit 1
fi

printf '%s\n' "${binary_path}"
