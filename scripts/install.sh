#!/bin/bash
# AIQ Installation Script for Unix/Linux/macOS
# Automatically detects latest version, architecture, and installs aiq binary

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# GitHub repository
REPO="sunetic/aiq"
GITHUB_API="https://api.github.com/repos/${REPO}/releases/latest"
GITHUB_RELEASES="https://github.com/${REPO}/releases/download"

# Installation directory
INSTALL_DIR="${HOME}/.aiq/bin"
BINARY_NAME="aiq"

# Detect latest version
echo "Detecting latest version..."
# Try releases API first
LATEST_VERSION=$(curl -s --max-time 10 "${GITHUB_API}" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

# Fallback to tags API if releases API fails
if [ -z "$LATEST_VERSION" ]; then
    echo -e "${YELLOW}Releases API failed, trying tags API...${NC}"
    LATEST_VERSION=$(curl -s --max-time 10 "https://api.github.com/repos/${REPO}/tags" 2>/dev/null | grep '"name":' | head -1 | sed -E 's/.*"([^"]+)".*/\1/' || echo "")
fi

# Final fallback
if [ -z "$LATEST_VERSION" ]; then
    echo -e "${YELLOW}Warning: Failed to fetch latest version from GitHub API. Using v0.0.1 as fallback.${NC}"
    LATEST_VERSION="v0.0.1"
fi
echo -e "${GREEN}Latest version: ${LATEST_VERSION}${NC}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: ${ARCH}${NC}"
        exit 1
        ;;
esac

# Map OS
case "$OS" in
    darwin)
        OS="darwin"
        ;;
    linux)
        OS="linux"
        ;;
    *)
        echo -e "${RED}Error: Unsupported OS: ${OS}${NC}"
        exit 1
        ;;
esac

PLATFORM="${OS}-${ARCH}"
echo -e "${GREEN}Detected platform: ${PLATFORM}${NC}"

# Construct download URL
GITHUB_URL="${GITHUB_RELEASES}/${LATEST_VERSION}/${BINARY_NAME}-${PLATFORM}"

# Create installation directory
mkdir -p "${INSTALL_DIR}"
echo -e "Install directory: ${GREEN}${INSTALL_DIR}${NC}"

# Download binary
echo "Downloading binary..."
BINARY_PATH="${INSTALL_DIR}/${BINARY_NAME}"

# Download from GitHub (no timeout - user can Ctrl+C if too slow)
echo "Download URL: ${GITHUB_URL}"
if curl -fSL --progress-bar "${GITHUB_URL}" -o "${BINARY_PATH}.tmp"; then
    echo -e "${GREEN}Downloaded successfully${NC}"
else
    echo -e "${RED}Error: Failed to download binary${NC}"
    echo -e "${YELLOW}Please check your network or download manually from:${NC}"
    echo -e "  ${GITHUB_URL}"
    exit 1
fi

# Make binary executable
chmod +x "${BINARY_PATH}.tmp"
mv "${BINARY_PATH}.tmp" "${BINARY_PATH}"

# Detect shell for PATH command
SHELL_NAME=$(basename "$SHELL")
SHELL_CONFIG=""

case "$SHELL_NAME" in
    zsh)
        SHELL_CONFIG="~/.zshrc"
        PATH_CMD="echo 'export PATH=\"\$HOME/.aiq/bin:\$PATH\"' >> ~/.zshrc"
        ;;
    bash)
        SHELL_CONFIG="~/.bashrc"
        PATH_CMD="echo 'export PATH=\"\$HOME/.aiq/bin:\$PATH\"' >> ~/.bashrc"
        ;;
    *)
        SHELL_CONFIG="~/.profile"
        PATH_CMD="echo 'export PATH=\"\$HOME/.aiq/bin:\$PATH\"' >> ~/.profile"
        ;;
esac

# Verify installation
echo "Verifying installation..."
if [ -f "${BINARY_PATH}" ] && [ -x "${BINARY_PATH}" ]; then
    echo -e "${GREEN}Installation successful!${NC}"
    echo ""
    
    # Check if PATH already contains INSTALL_DIR
    if echo "$PATH" | grep -q "${HOME}/.aiq/bin"; then
        echo -e "${GREEN}PATH already contains ${INSTALL_DIR}${NC}"
    else
        echo -e "${YELLOW}To use 'aiq' command, add it to your PATH:${NC}"
        echo -e "  ${GREEN}${PATH_CMD}${NC}"
        echo ""
        echo -e "${YELLOW}Then run:${NC}"
        echo -e "  ${GREEN}source ${SHELL_CONFIG}${NC}"
    fi
else
    echo -e "${RED}Warning: Installation completed but verification failed.${NC}"
    echo -e "${YELLOW}Please check if ${BINARY_PATH} exists and is executable.${NC}"
    exit 1
fi
