#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-v2026.06.30}"
RELEASE_DIR="release/${VERSION}"

rm -rf "${RELEASE_DIR}"
mkdir -p "${RELEASE_DIR}"

echo "==> Building IssGo ${VERSION}"
echo

# Linux
echo "  linux/amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/issgo" . && \
  tar czf "${RELEASE_DIR}/issgo_linux_amd64.tar.gz" -C "${RELEASE_DIR}" issgo

echo "  linux/arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/issgo" . && \
  tar czf "${RELEASE_DIR}/issgo_linux_arm64.tar.gz" -C "${RELEASE_DIR}" issgo

# macOS
echo "  darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/issgo" . && \
  tar czf "${RELEASE_DIR}/issgo_darwin_amd64.tar.gz" -C "${RELEASE_DIR}" issgo

echo "  darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/issgo" . && \
  tar czf "${RELEASE_DIR}/issgo_darwin_arm64.tar.gz" -C "${RELEASE_DIR}" issgo

# Windows
echo "  windows/amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "${RELEASE_DIR}/issgo.exe" . && \
  (cd "${RELEASE_DIR}" && zip issgo_windows_amd64.zip issgo.exe)

# Clean up temp binary
rm -f "${RELEASE_DIR}/issgo" "${RELEASE_DIR}/issgo.exe"

# Checksums
echo
echo "==> Generating checksums..."
cd "${RELEASE_DIR}"
sha256sum *.tar.gz *.zip > checksums.txt

echo
echo "==> Done! Artifacts in ${RELEASE_DIR}/"
echo
ls -lh "${RELEASE_DIR}"
echo
echo "Files to upload:"
ls "${RELEASE_DIR}"
