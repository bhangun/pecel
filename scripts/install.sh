#!/bin/bash

# Pecel Enhanced Installation Script
set -e

VERSION="v1.1.0"
REPO="github.com/bhangun/pecel"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="pecel"
OS="$(uname -s)"
ARCH="$(uname -m)"

# Detect OS and Architecture
case "$OS" in
    Linux*)     PLATFORM="linux" ;;
    Darwin*)    PLATFORM="darwin" ;;
    *)          echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
    x86_64)     ARCH="amd64" ;;
    arm64)      ARCH="arm64" ;;
    aarch64)    ARCH="arm64" ;;
    *)          echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════╗"
    echo "║          Pecel CLI Installer v1.1.0          ║"
    echo "╚══════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Check for existing installation
check_existing() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        CURRENT_VERSION=$($BINARY_NAME -version 2>/dev/null || echo "unknown")
        print_info "Found existing installation: $CURRENT_VERSION"
        return 0
    fi
    return 1
}

# Download and install
install_binary() {
    print_info "Downloading pecel $VERSION for $PLATFORM/$ARCH..."
    
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}-${PLATFORM}-${ARCH}"
    
    # Create temp directory
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # Download binary
    if ! curl -sSL -o "$TEMP_DIR/$BINARY_NAME" "$DOWNLOAD_URL"; then
        print_error "Failed to download binary from $DOWNLOAD_URL"
        print_info "Trying alternative URL..."
        DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/${BINARY_NAME}-${PLATFORM}-${ARCH}"
        curl -sSL -o "$TEMP_DIR/$BINARY_NAME" "$DOWNLOAD_URL" || {
            print_error "All download attempts failed"
            exit 1
        }
    fi
    
    # Make binary executable
    chmod +x "$TEMP_DIR/$BINARY_NAME"
    
    # Verify binary works
    if ! "$TEMP_DIR/$BINARY_NAME" --version &> /dev/null; then
        print_error "Downloaded binary appears to be invalid"
        exit 1
    fi
    
    # Show binary info
    BINARY_INFO=$("$TEMP_DIR/$BINARY_NAME" --version 2>/dev/null || echo "Unknown version")
    print_info "Binary version: $BINARY_INFO"
    
    # Install to system
    print_info "Installing to $INSTALL_DIR..."
    sudo mkdir -p "$INSTALL_DIR"
    sudo mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    
    # Verify installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        print_success "Installation completed successfully!"
        print_info "Run '$BINARY_NAME --help' to see all options"
        
        # Show example commands
        echo -e "${YELLOW}"
        echo "Examples:"
        echo "  $BINARY_NAME -i ./src -o output.txt"
        echo "  $BINARY_NAME --ext .go,.txt --format json --compress"
        echo "  $BINARY_NAME --config config.json --verbose"
        echo -e "${NC}"
    else
        print_warn "Binary installed but may not be in PATH"
        print_info "You may need to add $INSTALL_DIR to your PATH"
        print_info "Or run directly: $INSTALL_DIR/$BINARY_NAME"
    fi
}

# Uninstall function
uninstall() {
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_info "Removing $INSTALL_DIR/$BINARY_NAME..."
        sudo rm -f "$INSTALL_DIR/$BINARY_NAME"
        print_success "Pecel has been uninstalled"
    else
        print_warn "Pecel is not installed"
    fi
    exit 0
}

# Update function
update() {
    print_info "Checking for updates..."
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "$VERSION")
    
    if [ "$LATEST_VERSION" != "$VERSION" ]; then
        print_info "New version available: $LATEST_VERSION (current: $VERSION)"
        VERSION=$LATEST_VERSION
        install_binary
    else
        print_info "You have the latest version ($VERSION)"
    fi
    exit 0
}

# Install completion
install_completion() {
    print_info "Installing shell completion..."
    
    # Detect shell
    SHELL_NAME=$(basename "$SHELL")
    
    case "$SHELL_NAME" in
        bash)
            COMPLETION_DIR="/etc/bash_completion.d"
            if [ -d "$COMPLETION_DIR" ]; then
                sudo curl -sSL "https://raw.githubusercontent.com/$REPO/main/completions/bash/pecel" \
                    -o "$COMPLETION_DIR/pecel" || true
                print_info "Bash completion installed (may require restart)"
            fi
            ;;
        zsh)
            COMPLETION_DIR="/usr/local/share/zsh/site-functions"
            if [ -d "$COMPLETION_DIR" ]; then
                sudo curl -sSL "https://raw.githubusercontent.com/$REPO/main/completions/zsh/_pecel" \
                    -o "$COMPLETION_DIR/_pecel" || true
                print_info "Zsh completion installed"
            fi
            ;;
    esac
}

# Show help
show_help() {
    print_banner
    echo "Usage: $0 [OPTION]"
    echo
    echo "Options:"
    echo "  install     Install pecel (default)"
    echo "  update      Update to latest version"
    echo "  uninstall   Remove pecel"
    echo "  completion  Install shell completion"
    echo "  help        Show this help message"
    echo
    echo "One-line install:"
    echo "  curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash"
    echo
    echo "With options:"
    echo "  curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash -s update"
    echo
    exit 0
}

# Main execution
main() {
    print_banner
    
    case "${1:-install}" in
        install)    
            check_existing && print_warn "Overwriting existing installation"
            install_binary
            if [ "$2" != "--no-completion" ]; then
                install_completion
            fi
            ;;
        update)     
            update 
            ;;
        uninstall)  
            uninstall 
            ;;
        completion) 
            install_completion 
            ;;
        help)       
            show_help 
            ;;
        *)          
            print_error "Unknown option: $1" 
            show_help 
            ;;
    esac
}

main "$@"