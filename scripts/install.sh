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
API_RESPONSE=$(curl -s --max-time 10 "${GITHUB_API}" 2>&1)
CURL_EXIT_CODE=$?
TAGS_EXIT_CODE=0

if [ $CURL_EXIT_CODE -ne 0 ]; then
    echo -e "${YELLOW}Warning: curl failed with exit code ${CURL_EXIT_CODE}${NC}"
    if [ -n "$API_RESPONSE" ]; then
        echo -e "${YELLOW}Response: ${API_RESPONSE}${NC}"
    fi
fi

# Check for API rate limiting
if echo "$API_RESPONSE" | grep -q "rate limit exceeded"; then
    echo -e "${YELLOW}Warning: GitHub API rate limit exceeded. Trying tags API...${NC}"
    API_RESPONSE=""  # Clear response to trigger fallback
fi

LATEST_VERSION=$(echo "$API_RESPONSE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

# Fallback to tags API if releases API fails
if [ -z "$LATEST_VERSION" ]; then
    if [ -z "$API_RESPONSE" ] || ! echo "$API_RESPONSE" | grep -q "rate limit exceeded"; then
        echo -e "${YELLOW}Releases API failed, trying tags API...${NC}"
    fi
    TAGS_RESPONSE=$(curl -s --max-time 10 "https://api.github.com/repos/${REPO}/tags" 2>&1)
    TAGS_EXIT_CODE=$?
    
    # Check for API rate limiting in tags API too
    if echo "$TAGS_RESPONSE" | grep -q "rate limit exceeded"; then
        echo -e "${RED}Error: GitHub API rate limit exceeded for both releases and tags APIs.${NC}"
        echo -e "${YELLOW}This usually happens when making too many requests from the same IP address.${NC}"
        echo -e "${YELLOW}Solutions:${NC}"
        echo -e "  1. Wait a few minutes and try again"
        echo -e "  2. Use a GitHub token for higher rate limits (set GITHUB_TOKEN env var)"
        echo -e "  3. Manually download from: https://github.com/${REPO}/releases"
        exit 1
    fi
    
    if [ $TAGS_EXIT_CODE -ne 0 ]; then
        echo -e "${YELLOW}Warning: tags API curl failed with exit code ${TAGS_EXIT_CODE}${NC}"
        if [ -n "$TAGS_RESPONSE" ]; then
            echo -e "${YELLOW}Response: ${TAGS_RESPONSE}${NC}"
        fi
    fi
    
    LATEST_VERSION=$(echo "$TAGS_RESPONSE" | grep '"name":' | head -1 | sed -E 's/.*"([^"]+)".*/\1/' || echo "")
fi

# Fail if version detection failed
if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}Error: Failed to fetch latest version from GitHub API.${NC}"
    echo -e "${YELLOW}Diagnostic info:${NC}"
    echo -e "  - Releases API URL: ${GITHUB_API}"
    echo -e "  - Releases API curl exit code: ${CURL_EXIT_CODE}"
    echo -e "  - Tags API curl exit code: ${TAGS_EXIT_CODE}"
    echo -e "${YELLOW}Possible causes:${NC}"
    echo -e "  - Network connectivity issues"
    echo -e "  - GitHub API rate limiting"
    echo -e "  - Firewall or proxy blocking GitHub API"
    echo -e "${YELLOW}You can manually download from: https://github.com/${REPO}/releases${NC}"
    exit 1
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

# Create installation directory if it doesn't exist
if [ ! -d "${INSTALL_DIR}" ]; then
    mkdir -p "${INSTALL_DIR}"
    echo -e "Created install directory: ${GREEN}${INSTALL_DIR}${NC}"
else
    echo -e "Install directory: ${GREEN}${INSTALL_DIR}${NC}"
fi

# Check if binary already exists
BINARY_PATH="${INSTALL_DIR}/${BINARY_NAME}"
if [ -f "${BINARY_PATH}" ]; then
    echo -e "${YELLOW}Binary already exists at: ${BINARY_PATH}${NC}"
    read -p "Overwrite existing binary? [y/N]: " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Installation cancelled. Keeping existing binary.${NC}"
        exit 0
    fi
    echo -e "${GREEN}Proceeding with overwrite...${NC}"
fi

# Download binary
echo "Downloading binary..."
echo "Download URL: ${GITHUB_URL}"

# Download from GitHub (no timeout - user can Ctrl+C if too slow)
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
