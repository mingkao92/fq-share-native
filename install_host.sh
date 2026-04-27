#!/usr/bin/env bash
set -euo pipefail

DEFAULT_REPO="${FQ_SHARE_GITHUB_REPO:-mingkao92/fq-share-native}"

usage() {
  echo "Usage: $0 <extension-id> [chrome|chromium] [github-repo] [release-tag|latest]"
  echo "Example (simple): $0 abcdefghijklmnopqrstuvwxyz"
  echo "Example (custom repo/tag): $0 abcdefghijklmnopqrstuvwxyz chrome mingkao92/fq-share-native v0.2.0"
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

EXTENSION_ID="$1"
BROWSER="${2:-chrome}"
REPO="${3:-${DEFAULT_REPO}}"
VERSION="${4:-latest}"

if [[ "${BROWSER}" != "chrome" && "${BROWSER}" != "chromium" ]]; then
  echo "Invalid browser: ${BROWSER}. Use chrome or chromium."
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl not found. Please install curl first."
  exit 1
fi
if ! command -v tar >/dev/null 2>&1; then
  echo "tar not found. Please install tar first."
  exit 1
fi

OS_NAME="$(uname -s)"
ARCH_NAME="$(uname -m)"

case "${OS_NAME}" in
  Darwin) TARGET_OS="darwin" ;;
  Linux) TARGET_OS="linux" ;;
  *)
    echo "Unsupported OS: ${OS_NAME}"
    exit 1
    ;;
esac

case "${ARCH_NAME}" in
  x86_64|amd64) TARGET_ARCH="amd64" ;;
  aarch64|arm64) TARGET_ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: ${ARCH_NAME}"
    exit 1
    ;;
esac

ASSET="free-quick-share-host-${TARGET_OS}-${TARGET_ARCH}.tar.gz"
if [[ "${VERSION}" == "latest" ]]; then
  DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
else
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
fi

TMP_DIR="$(mktemp -d)"
ARCHIVE_PATH="${TMP_DIR}/${ASSET}"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

echo "Downloading: ${DOWNLOAD_URL}"
curl -fL --retry 3 --connect-timeout 10 "${DOWNLOAD_URL}" -o "${ARCHIVE_PATH}"
tar -xzf "${ARCHIVE_PATH}" -C "${TMP_DIR}"

HOST_BIN="${TMP_DIR}/free-quick-share-host"
if [[ ! -x "${HOST_BIN}" ]]; then
  FOUND_BIN="$(find "${TMP_DIR}" -maxdepth 3 -type f -name 'free-quick-share-host' | head -n1 || true)"
  if [[ -z "${FOUND_BIN}" ]]; then
    echo "Downloaded archive does not contain free-quick-share-host"
    exit 1
  fi
  HOST_BIN="${FOUND_BIN}"
  chmod +x "${HOST_BIN}"
fi

"${HOST_BIN}" install --extension-id "${EXTENSION_ID}" --browser "${BROWSER}"
