#!/bin/bash
#
# Claude Proxy Installer
# Install or update Claude Proxy service on Linux
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/r9r-dev/home-agent/main/claude-proxy/install.sh | bash
#
# Or with options:
#   curl -fsSL ... | bash -s -- --version v1.0.0
#   curl -fsSL ... | bash -s -- --uninstall
#

set -e

# Configuration
GITHUB_REPO="r9r-dev/home-agent"
BINARY_NAME="claude-proxy"
INSTALL_DIR="/opt/claude-proxy"
SERVICE_NAME="claude-proxy"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Print banner
banner() {
    echo ""
    echo "  ╔═══════════════════════════════════════╗"
    echo "  ║       Claude Proxy Installer          ║"
    echo "  ║   Claude CLI Gateway for Home Agent   ║"
    echo "  ╚═══════════════════════════════════════╝"
    echo ""
}

# Detect architecture
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64)  echo "amd64" ;;
        aarch64) echo "arm64" ;;
        armv7l)  echo "arm" ;;
        *)       error "Unsupported architecture: $arch" ;;
    esac
}

# Detect OS
detect_os() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $os in
        linux)  echo "linux" ;;
        darwin) echo "darwin" ;;
        *)      error "Unsupported OS: $os" ;;
    esac
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root (use sudo)"
    fi
}

# Check dependencies
check_deps() {
    local missing=()

    for cmd in curl tar systemctl; do
        if ! command -v $cmd &> /dev/null; then
            missing+=($cmd)
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        error "Missing dependencies: ${missing[*]}"
    fi
}

# Get latest version from GitHub
get_latest_version() {
    local latest=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [[ -z "$latest" ]]; then
        error "Failed to get latest version from GitHub"
    fi
    echo "$latest"
}

# Get installed version
get_installed_version() {
    if [[ -x "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
        "${INSTALL_DIR}/${BINARY_NAME}" --version 2>/dev/null | head -1 || echo "unknown"
    else
        echo "not installed"
    fi
}

# Download and install binary
install_binary() {
    local version=$1
    local os=$(detect_os)
    local arch=$(detect_arch)

    # Construct download URL
    local filename="${BINARY_NAME}-${os}-${arch}.tar.gz"
    local url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${filename}"

    info "Downloading ${BINARY_NAME} ${version} for ${os}/${arch}..."
    info "URL: ${url}"

    # Create temp directory
    local tmp_dir=$(mktemp -d)
    trap "rm -rf ${tmp_dir}" EXIT

    # Download
    if ! curl -fsSL -o "${tmp_dir}/${filename}" "$url"; then
        error "Failed to download from ${url}"
    fi

    # Extract
    info "Extracting..."
    tar -xzf "${tmp_dir}/${filename}" -C "${tmp_dir}"

    # Install
    info "Installing to ${INSTALL_DIR}..."
    mkdir -p "${INSTALL_DIR}"
    cp "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    success "Binary installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Install systemd service
install_service() {
    info "Installing systemd service..."

    # Get the user who ran sudo
    local run_user="${SUDO_USER:-root}"
    local run_group=$(id -gn "$run_user" 2>/dev/null || echo "root")

    # Generate API key suggestion
    local suggested_key=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p)

    cat > "${SERVICE_FILE}" << EOF
[Unit]
Description=Claude Proxy Service - Claude CLI Gateway
Documentation=https://github.com/${GITHUB_REPO}
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=${run_user}
Group=${run_group}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${BINARY_NAME}

# Environment
Environment="PROXY_PORT=9090"
Environment="PROXY_HOST=0.0.0.0"
Environment="CLAUDE_BIN=claude"
# Set your API key below (uncomment and set)
#Environment="PROXY_API_KEY=${suggested_key}"

# Restart policy
Restart=always
RestartSec=5s

# Limits
LimitNOFILE=65535

# Security
PrivateTmp=true
NoNewPrivileges=true

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    success "Service installed to ${SERVICE_FILE}"

    echo ""
    warn "IMPORTANT: Configure your API key!"
    echo "  1. Edit the service file:"
    echo "     sudo nano ${SERVICE_FILE}"
    echo ""
    echo "  2. Uncomment and set PROXY_API_KEY:"
    echo "     Environment=\"PROXY_API_KEY=${suggested_key}\""
    echo ""
    echo "  3. Reload and restart:"
    echo "     sudo systemctl daemon-reload"
    echo "     sudo systemctl restart ${SERVICE_NAME}"
    echo ""
}

# Enable and start service
start_service() {
    info "Enabling and starting service..."
    systemctl enable "${SERVICE_NAME}" 2>/dev/null || true
    systemctl restart "${SERVICE_NAME}"
    success "Service started"
}

# Show status
show_status() {
    echo ""
    echo "Service Status:"
    echo "───────────────"
    systemctl status "${SERVICE_NAME}" --no-pager -l || true
    echo ""
    echo "Useful commands:"
    echo "  sudo systemctl status ${SERVICE_NAME}    # Check status"
    echo "  sudo systemctl restart ${SERVICE_NAME}   # Restart"
    echo "  sudo journalctl -u ${SERVICE_NAME} -f    # View logs"
    echo ""
    echo "Test the service:"
    echo "  curl http://localhost:9090/health"
    echo ""
}

# Uninstall
uninstall() {
    info "Uninstalling ${BINARY_NAME}..."

    # Stop and disable service
    systemctl stop "${SERVICE_NAME}" 2>/dev/null || true
    systemctl disable "${SERVICE_NAME}" 2>/dev/null || true

    # Remove files
    rm -f "${SERVICE_FILE}"
    rm -rf "${INSTALL_DIR}"

    systemctl daemon-reload

    success "Uninstalled successfully"
}

# Main installation
install() {
    local version=$1
    local is_upgrade=false

    # Check if already installed
    local current=$(get_installed_version)
    if [[ "$current" != "not installed" ]]; then
        is_upgrade=true
        info "Current version: ${current}"
    fi

    # Get version to install
    if [[ -z "$version" ]]; then
        info "Fetching latest version..."
        version=$(get_latest_version)
    fi
    info "Version to install: ${version}"

    # Install binary
    install_binary "$version"

    # Install service (only if new install)
    if [[ "$is_upgrade" == false ]] || [[ ! -f "${SERVICE_FILE}" ]]; then
        install_service
    else
        info "Service file already exists, skipping..."
    fi

    # Restart service if running
    if systemctl is-active --quiet "${SERVICE_NAME}"; then
        info "Restarting service..."
        systemctl restart "${SERVICE_NAME}"
        success "Service restarted with new version"
    fi

    echo ""
    success "Installation complete!"

    if [[ "$is_upgrade" == false ]]; then
        echo ""
        echo "Next steps:"
        echo "  1. Configure API key in ${SERVICE_FILE}"
        echo "  2. Start the service: sudo systemctl start ${SERVICE_NAME}"
        echo "  3. Test: curl http://localhost:9090/health"
    fi
}

# Parse arguments
VERSION=""
UNINSTALL=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --version|-v)
            VERSION="$2"
            shift 2
            ;;
        --uninstall|-u)
            UNINSTALL=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --version, -v VERSION   Install specific version"
            echo "  --uninstall, -u         Uninstall claude-proxy"
            echo "  --help, -h              Show this help"
            echo ""
            echo "Examples:"
            echo "  curl -fsSL URL | sudo bash"
            echo "  curl -fsSL URL | sudo bash -s -- --version v1.0.0"
            echo "  curl -fsSL URL | sudo bash -s -- --uninstall"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

# Main
banner
check_root
check_deps

if [[ "$UNINSTALL" == true ]]; then
    uninstall
else
    install "$VERSION"
    show_status
fi
