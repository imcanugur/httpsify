#!/usr/bin/env bash

# HTTPSify Installation Script
# This script automatically detects your OS and architecture, downloads the 
# latest release from GitHub, and installs it to your system.

set -e

# Configuration
REPO="imcanugur/httpsify"
BINARY_NAME="httpsify"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

printf "${BLUE}==>${NC} Detecting system... "

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
    linux*)     OS='linux';;
    darwin*)    OS='darwin';;
    msys*|cygwin*|mingw*) OS='windows';;
    *)          printf "${RED}Error: Unsupported OS: ${OS}${NC}\n"; exit 1;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64)     ARCH='amd64';;
    arm64|aarch64) ARCH='arm64';;
    *)          printf "${RED}Error: Unsupported architecture: ${ARCH}${NC}\n"; exit 1;;
esac

printf "${GREEN}${OS}-${ARCH}${NC}\n"

# Get latest release tag
printf "${BLUE}==>${NC} Fetching latest release... "
LATEST_TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${LATEST_TAG}" ]; then
    printf "${RED}Error: Could not fetch latest release tag.${NC}\n"
    exit 1
fi
printf "${GREEN}${LATEST_TAG}${NC}\n"

# Construct download URL
EXTENSION=""
if [ "${OS}" = "windows" ]; then
    EXTENSION=".exe"
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BINARY_NAME}-${OS}-${ARCH}${EXTENSION}"

# Download binary
printf "${BLUE}==>${NC} Downloading ${BINARY_NAME}... "
TMP_FILE=$(mktemp)
curl -sL "${DOWNLOAD_URL}" -o "${TMP_FILE}"

if [ ! -s "${TMP_FILE}" ]; then
    printf "${RED}Error: Downloaded file is empty. Check if the release asset exists: ${DOWNLOAD_URL}${NC}\n"
    rm "${TMP_FILE}"
    exit 1
fi
printf "${GREEN}done${NC}\n"

# Install binary
printf "${BLUE}==>${NC} Installing to ${INSTALL_DIR}/${BINARY_NAME}... "
chmod +x "${TMP_FILE}"

# Check for sudo if needed
SUDO=""
if [ ! -w "${INSTALL_DIR}" ]; then
    SUDO="sudo"
fi

${SUDO} mv "${TMP_FILE}" "${INSTALL_DIR}/${BINARY_NAME}"
printf "${GREEN}done${NC}\n"

printf "\n${GREEN}HTTPSify has been successfully installed!${NC}\n"
printf "Run it with: ${BLUE}sudo ${BINARY_NAME}${NC}\n"
