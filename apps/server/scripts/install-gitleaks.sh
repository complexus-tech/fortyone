#!/usr/bin/env bash

set -euo pipefail

readonly version="8.27.2"
readonly install_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/.tools/bin"
readonly binary_path="${install_dir}/gitleaks"

if [[ -x "${binary_path}" ]] && "${binary_path}" version 2>/dev/null | grep -Fq "${version}"; then
  exit 0
fi

case "$(uname -s)-$(uname -m)" in
  Linux-x86_64)
    readonly asset="gitleaks_${version}_linux_x64.tar.gz"
    readonly expected_sha256="141c3b2dede46d8b3a53b47116da756bd223decc0374797559a6b50ecba5590c"
    ;;
  Linux-aarch64 | Linux-arm64)
    readonly asset="gitleaks_${version}_linux_arm64.tar.gz"
    readonly expected_sha256="fd59a77b3d898ab14782264bf7a22db457871db56debc5d7ac3e30b64b379921"
    ;;
  Darwin-x86_64)
    readonly asset="gitleaks_${version}_darwin_x64.tar.gz"
    readonly expected_sha256="aa79c412d76872d4917e6c53f784fd247576ded0d06c17262dc0299e4cc8e79f"
    ;;
  Darwin-arm64)
    readonly asset="gitleaks_${version}_darwin_arm64.tar.gz"
    readonly expected_sha256="ae969ca6b04c8621bae4dbb707cb4293264904c0e890901f0643c266d5e02bea"
    ;;
  *)
    echo "unsupported platform for pinned gitleaks: $(uname -s) $(uname -m)" >&2
    exit 1
    ;;
esac

readonly download_url="https://github.com/gitleaks/gitleaks/releases/download/v${version}/${asset}"
temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/fortyone-gitleaks.XXXXXX")"
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
  echo "gitleaks checksum mismatch: expected ${expected_sha256}, received ${actual_sha256}" >&2
  exit 1
fi

tar -xzf "${temporary_dir}/${asset}" -C "${temporary_dir}" gitleaks
mkdir -p "${install_dir}"
install -m 0755 "${temporary_dir}/gitleaks" "${binary_path}.tmp"
mv -f "${binary_path}.tmp" "${binary_path}"

if ! "${binary_path}" version | grep -Fq "${version}"; then
  echo "installed gitleaks did not report expected version ${version}" >&2
  exit 1
fi
