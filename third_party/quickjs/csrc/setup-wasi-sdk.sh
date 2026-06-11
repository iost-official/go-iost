#!/bin/bash
#
# Setup script for wasi-sdk on macOS (arm64/x86_64)
#
# This script downloads and installs wasi-sdk which is needed to compile
# C code to WebAssembly with WASI support.
#

set -e

WASI_SDK_VERSION="29"

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    WASI_SDK_URL="https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-${WASI_SDK_VERSION}/wasi-sdk-${WASI_SDK_VERSION}.0-arm64-macos.tar.gz"
elif [ "$ARCH" = "x86_64" ]; then
    WASI_SDK_URL="https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-${WASI_SDK_VERSION}/wasi-sdk-${WASI_SDK_VERSION}.0-x86_64-macos.tar.gz"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

INSTALL_DIR="${WASI_SDK_PATH:-$HOME/wasi-sdk}"

echo "=== wasi-sdk Setup Script ==="
echo ""
echo "Architecture: $ARCH"
echo "Version: $WASI_SDK_VERSION"
echo ""

# Check if already installed
if [ -f "$INSTALL_DIR/bin/clang" ]; then
    echo "wasi-sdk already installed at $INSTALL_DIR"
    "$INSTALL_DIR/bin/clang" --version
    exit 0
fi

echo "Installing wasi-sdk to: $INSTALL_DIR"
echo ""

# Create temp directory
TMPDIR=$(mktemp -d)
cleanup() {
    rm -rf "$TMPDIR"
}
trap cleanup EXIT

# Download
echo "Downloading wasi-sdk from:"
echo "  $WASI_SDK_URL"
echo ""
curl -L "$WASI_SDK_URL" -o "$TMPDIR/wasi-sdk.tar.gz" --progress-bar

# Extract
echo "Extracting..."
mkdir -p "$INSTALL_DIR"
tar xf "$TMPDIR/wasi-sdk.tar.gz" -C "$INSTALL_DIR" --strip-components=1

# Verify installation
if [ -f "$INSTALL_DIR/bin/clang" ]; then
    echo ""
    echo "=== wasi-sdk installed successfully! ==="
    echo ""
    echo "Installation path: $INSTALL_DIR"
    echo ""
    echo "Add to your shell profile (~/.zshrc or ~/.bashrc):"
    echo ""
    echo "  export WASI_SDK_PATH=$INSTALL_DIR"
    echo ""
    echo "Or build with:"
    echo ""
    echo "  make WASI_SDK_PATH=$INSTALL_DIR"
    echo ""
    "$INSTALL_DIR/bin/clang" --version
else
    echo "ERROR: Installation failed"
    exit 1
fi
