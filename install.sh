#!/bin/bash

# This script automates the installation of the clambda tool.
# It checks for the specified version (or fetches the latest one),
# downloads the binary, and installs it on the system.

set -e

# Check for required tools: curl and tar.
# These tools are necessary for downloading and extracting the clambda binary.
if ! command -v curl &> /dev/null; then
    echo "curl could not be found"
    exit 1
fi

if ! command -v tar &> /dev/null; then
    echo "tar could not be found"
    exit 1
fi

# Determine the version of clambda to install.
# If no version is specified as a command line argument, fetch the latest version.
if [ -z "$1" ]; then
    VERSION=$(curl -s https://api.github.com/repos/shikira/clambda/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo "Failed to fetch the latest version"
        exit 1
    fi
else
    VERSION=$1
fi

# Normalize the version string by removing any leading 'v'.
VERSION=${VERSION#v}

# Detect the architecture of the current system.
# This script supports x86_64, arm64, and i386 architectures.
ARCH=$(uname -m)
case $ARCH in
    x86_64|amd64) ARCH="x86_64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    i386|i686) ARCH="i386" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Detect the operating system (OS) of the current system.
# This script supports Linux, Darwin (macOS) and Windows operating systems.
OS=$(uname -s)
case $OS in
    Linux) OS="Linux" ;;
    Darwin) OS="Darwin" ;;
    MINGW*|MSYS*|CYGWIN*) OS="Windows" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Construct the download URL for the clambda binary based on the version, OS, and architecture.
FILE_NAME="clambda_${VERSION}_${OS}_${ARCH}.tar.gz"
if [ "$OS" = "Windows" ]; then
    FILE_NAME="clambda_${VERSION}_${OS}_${ARCH}.zip"
fi

URL="https://github.com/shikira/clambda/releases/download/v${VERSION}/${FILE_NAME}"

# Download the clambda binary.
echo "Downloading clambda version ${VERSION} for ${OS}/${ARCH}..."
if ! curl -L -o "$FILE_NAME" "$URL"; then
    echo "Failed to download clambda"
    exit 1
fi

# Install clambda.
# This involves extracting the binary and moving it to /usr/local/bin.
echo "Installing clambda..."
if [ "$OS" = "Windows" ]; then
    if ! command -v unzip &> /dev/null; then
        echo "unzip could not be found"
        exit 1
    fi
    if ! unzip -o "$FILE_NAME"; then
        echo "Failed to extract clambda"
        exit 1
    fi
else
    if ! tar -xzf "$FILE_NAME"; then
        echo "Failed to extract clambda"
        exit 1
    fi
fi

# Move binary to /usr/local/bin (or a Windows-appropriate location)
if [ "$OS" = "Windows" ]; then
    # For Windows, just inform the user to add it to PATH manually
    echo "clambda.exe extracted successfully."
    echo "Please add the current directory to your PATH or move clambda.exe to a directory in your PATH."
else
    if ! sudo mv clambda /usr/local/bin/clambda; then
        echo "Failed to install clambda to /usr/local/bin"
        echo "You may need to run this script with appropriate permissions or manually move the binary."
        exit 1
    fi
    # Make sure it's executable
    sudo chmod +x /usr/local/bin/clambda
fi

# Clean up by removing the downloaded archive file.
rm "$FILE_NAME"

echo "clambda installation complete."
echo "Run 'clambda --help' to see how to use clambda."
