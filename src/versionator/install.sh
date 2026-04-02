#!/bin/bash
set -e

VERSION="${VERSION:-latest}"
INSTALL_DIR="/usr/local/bin"

echo "Installing versionator ${VERSION}..."

# Detect architecture
ARCH=$(uname -m)
case ${ARCH} in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: ${ARCH}"
        exit 1
        ;;
esac

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case ${OS} in
    linux|darwin)
        ;;
    *)
        echo "Unsupported OS: ${OS}"
        exit 1
        ;;
esac

# Determine download URL
REPO="benjaminabbitt/versionator"

if [ "${VERSION}" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/versionator-${OS}-${ARCH}"
else
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/versionator-${OS}-${ARCH}"
fi

echo "Downloading from: ${DOWNLOAD_URL}"

# Download and install
curl -fsSL "${DOWNLOAD_URL}" -o "${INSTALL_DIR}/versionator" || {
    # Fallback: try building from source if Go is available
    if command -v go &> /dev/null; then
        echo "Download failed, attempting to install from source..."
        go install "github.com/${REPO}@${VERSION}"

        # Move to install dir if installed to GOPATH
        if [ -f "${GOPATH}/bin/versionator" ]; then
            cp "${GOPATH}/bin/versionator" "${INSTALL_DIR}/versionator"
        fi
    else
        echo "Failed to download versionator and Go is not available for source installation"
        exit 1
    fi
}

chmod +x "${INSTALL_DIR}/versionator"

echo "versionator installed successfully!"
versionator output version 2>/dev/null || echo "Version: ${VERSION}"
