#!/usr/bin/env bash
set -euo pipefail

# scripts/package.sh - packages release artifacts and generates SHA256SUMS
# Usage: ./scripts/package.sh [OUTPUT_DIR]

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${1:-${ROOT_DIR}/release}"

mkdir -p "${OUT_DIR}"
BIN_DIR="${ROOT_DIR}/build/bin"

echo "Packaging release artifacts into ${OUT_DIR}..."

# Linux
if [ -f "${BIN_DIR}/captive-bypass" ]; then
    echo "Packaging Linux binary..."
    tar -czvf "${OUT_DIR}/captive-bypass-linux-amd64.tar.gz" -C "${BIN_DIR}" captive-bypass
fi

# macOS (.app bundle)
if [ -d "${BIN_DIR}/captive-bypass.app" ]; then
    echo "Packaging macOS application bundle..."
    if command -v ditto >/dev/null 2>&1; then
        ditto -c -k --keepParent "${BIN_DIR}/captive-bypass.app" "${OUT_DIR}/captive-bypass-darwin-universal.zip"
    else
        (cd "${BIN_DIR}" && zip -r "${OUT_DIR}/captive-bypass-darwin-universal.zip" captive-bypass.app)
    fi
fi

# Windows (.exe)
if [ -f "${BIN_DIR}/captive-bypass.exe" ]; then
    echo "Packaging Windows binary..."
    if command -v zip >/dev/null 2>&1; then
        (cd "${BIN_DIR}" && zip "${OUT_DIR}/captive-bypass-windows-amd64.zip" captive-bypass.exe)
    elif command -v 7z >/dev/null 2>&1; then
        7z a "${OUT_DIR}/captive-bypass-windows-amd64.zip" "${BIN_DIR}/captive-bypass.exe"
    fi
fi

# Generate SHA256SUMS
echo "Computing SHA256 checksums..."
(
    cd "${OUT_DIR}"
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum captive-bypass-* > SHA256SUMS 2>/dev/null || true
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 captive-bypass-* > SHA256SUMS 2>/dev/null || true
    fi
)

echo "Done. Artifacts in ${OUT_DIR}:"
ls -la "${OUT_DIR}"
