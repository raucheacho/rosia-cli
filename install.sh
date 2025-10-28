#!/bin/bash
# Rosia CLI Installation Script for Unix Systems (Linux/macOS)
# Usage: curl -fsSL https://raw.githubusercontent.com/raucheacho/rosia-cli/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="raucheacho/rosia-cli"
BINARY_NAME="rosia"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
CONFIG_DIR="${HOME}/.rosia"
PROFILES_DIR="${CONFIG_DIR}/profiles"

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        *)          
            echo -e "${RED}Error: Unsupported operating system$(uname -s)${NC}"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64)     arch="amd64";;
        amd64)      arch="amd64";;
        arm64)      arch="arm64";;
        aarch64)    arch="arm64";;
        armv7l)     arch="armv7";;
        *)          
            echo -e "${RED}Error: Unsupported architecture $(uname -m)${NC}"
            exit 1
            ;;
    esac
    
    echo "${os}_${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        echo -e "${RED}Error: Failed to fetch latest version${NC}"
        exit 1
    fi
    
    echo "$version"
}

# Download and install binary
install_binary() {
    local version="$1"
    local platform="$2"
    local download_url="https://github.com/${REPO}/releases/download/${version}/rosia_${version#v}_${platform}.tar.gz"
    local tmp_dir=$(mktemp -d)
    
    echo -e "${BLUE}Downloading Rosia CLI ${version} for ${platform}...${NC}"
    
    # Download archive
    if ! curl -fsSL "$download_url" -o "${tmp_dir}/rosia.tar.gz"; then
        echo -e "${RED}Error: Failed to download from ${download_url}${NC}"
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    # Extract archive
    echo -e "${BLUE}Extracting archive...${NC}"
    tar -xzf "${tmp_dir}/rosia.tar.gz" -C "$tmp_dir"
    
    # Check if we have write permission to install directory
    if [ ! -w "$INSTALL_DIR" ]; then
        echo -e "${YELLOW}Installing to ${INSTALL_DIR} requires sudo privileges${NC}"
        sudo mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Copy profiles if they exist in the archive
    if [ -d "${tmp_dir}/profiles" ]; then
        echo -e "${BLUE}Installing default profiles...${NC}"
        mkdir -p "$PROFILES_DIR"
        cp -r "${tmp_dir}/profiles/"* "$PROFILES_DIR/"
    fi
    
    # Cleanup
    rm -rf "$tmp_dir"
    
    echo -e "${GREEN}✓ Rosia CLI installed successfully to ${INSTALL_DIR}/${BINARY_NAME}${NC}"
}

# Setup configuration directory
setup_config() {
    if [ ! -d "$CONFIG_DIR" ]; then
        echo -e "${BLUE}Creating configuration directory at ${CONFIG_DIR}...${NC}"
        mkdir -p "$CONFIG_DIR"
        mkdir -p "${CONFIG_DIR}/trash"
        mkdir -p "${CONFIG_DIR}/plugins"
        mkdir -p "$PROFILES_DIR"
    fi
    
    # Create default config if it doesn't exist
    local config_file="${HOME}/.rosiarc.json"
    if [ ! -f "$config_file" ]; then
        echo -e "${BLUE}Creating default configuration file...${NC}"
        cat > "$config_file" <<EOF
{
  "trash_retention_days": 3,
  "profiles": ["node", "python", "rust", "flutter", "go"],
  "ignore_paths": [],
  "plugins": [],
  "concurrency": 0,
  "telemetry_enabled": false
}
EOF
        echo -e "${GREEN}✓ Default configuration created at ${config_file}${NC}"
    fi
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        local version=$("$BINARY_NAME" version 2>&1 | head -n 1)
        echo -e "${GREEN}✓ Installation verified: ${version}${NC}"
        return 0
    else
        echo -e "${RED}Error: Installation verification failed${NC}"
        echo -e "${YELLOW}Please ensure ${INSTALL_DIR} is in your PATH${NC}"
        return 1
    fi
}

# Print usage instructions
print_usage() {
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Rosia CLI has been installed successfully!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${BLUE}Quick Start:${NC}"
    echo -e "  ${BINARY_NAME} scan ~/projects          # Scan for cleanable files"
    echo -e "  ${BINARY_NAME} ui ~/projects            # Launch interactive TUI"
    echo -e "  ${BINARY_NAME} clean                    # Clean detected targets"
    echo -e "  ${BINARY_NAME} stats                    # View cleaning statistics"
    echo ""
    echo -e "${BLUE}Configuration:${NC}"
    echo -e "  Config file: ${HOME}/.rosiarc.json"
    echo -e "  Profiles:    ${PROFILES_DIR}"
    echo -e "  Trash:       ${CONFIG_DIR}/trash"
    echo ""
    echo -e "${BLUE}Documentation:${NC}"
    echo -e "  ${BINARY_NAME} --help                   # Show help"
    echo -e "  https://github.com/${REPO}"
    echo ""
}

# Main installation flow
main() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  Rosia CLI Installer${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    # Check for required commands
    for cmd in curl tar; do
        if ! command -v "$cmd" &> /dev/null; then
            echo -e "${RED}Error: Required command '$cmd' not found${NC}"
            exit 1
        fi
    done
    
    # Detect platform
    local platform=$(detect_platform)
    echo -e "${BLUE}Detected platform: ${platform}${NC}"
    
    # Get latest version
    local version=$(get_latest_version)
    echo -e "${BLUE}Latest version: ${version}${NC}"
    echo ""
    
    # Install binary
    install_binary "$version" "$platform"
    
    # Setup configuration
    setup_config
    
    # Verify installation
    echo ""
    verify_installation
    
    # Print usage instructions
    print_usage
}

# Run main function
main "$@"
