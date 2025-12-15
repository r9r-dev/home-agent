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

# Global variable to store the API key (set by install_service)
INSTALLED_API_KEY=""

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
# Sets INSTALLED_API_KEY global variable
install_service() {
    info "Installing systemd service..."

    # Get the user who ran sudo
    local run_user="${SUDO_USER:-root}"
    local run_group=$(id -gn "$run_user" 2>/dev/null || echo "root")

    # Check for existing API key
    local existing_key=$(get_existing_api_key)

    if [[ -n "$existing_key" ]]; then
        info "Using existing API key from service file"
        INSTALLED_API_KEY="$existing_key"
    else
        info "Generating new API key..."
        INSTALLED_API_KEY=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p)
    fi

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
Environment="PROXY_API_KEY=${INSTALLED_API_KEY}"

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
}

# Enable and start service
start_service() {
    info "Enabling and starting service..."
    systemctl enable "${SERVICE_NAME}" 2>/dev/null || true
    systemctl restart "${SERVICE_NAME}"
    success "Service started"
}

# Detect primary IP address
detect_ip() {
    # Try to get the primary IP (the one used for default route)
    local ip=""

    # Method 1: ip route (Linux)
    if command -v ip &> /dev/null; then
        ip=$(ip route get 1 2>/dev/null | awk '{print $7; exit}')
    fi

    # Method 2: hostname -I (Linux fallback)
    if [[ -z "$ip" ]] && command -v hostname &> /dev/null; then
        ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi

    # Method 3: ifconfig (macOS/BSD)
    if [[ -z "$ip" ]] && command -v ifconfig &> /dev/null; then
        ip=$(ifconfig 2>/dev/null | grep 'inet ' | grep -v '127.0.0.1' | head -1 | awk '{print $2}')
    fi

    # Fallback
    if [[ -z "$ip" ]]; then
        ip="<YOUR_HOST_IP>"
    fi

    echo "$ip"
}

# Extract existing API key from service file
get_existing_api_key() {
    if [[ -f "${SERVICE_FILE}" ]]; then
        local key=$(grep 'PROXY_API_KEY=' "${SERVICE_FILE}" | grep -v '^#' | sed 's/.*PROXY_API_KEY=//' | tr -d '"' | tr -d "'" | head -1)
        if [[ -n "$key" && "$key" != "" ]]; then
            echo "$key"
        fi
    fi
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

# Show Home Agent configuration instructions
show_home_agent_config() {
    local api_key=$1
    local host_ip=$(detect_ip)

    echo ""
    echo "  ╔═══════════════════════════════════════════════════════════════╗"
    echo "  ║           Home Agent Configuration                            ║"
    echo "  ╚═══════════════════════════════════════════════════════════════╝"
    echo ""
    echo "  Add these environment variables to your Home Agent container:"
    echo ""
    echo "  ┌─────────────────────────────────────────────────────────────┐"
    echo "  │  CLAUDE_PROXY_URL=http://${host_ip}:9090"
    echo "  │  CLAUDE_PROXY_KEY=${api_key}"
    echo "  └─────────────────────────────────────────────────────────────┘"
    echo ""
    echo "  For docker-compose, add to your .env file:"
    echo ""
    echo "    HOST_IP=${host_ip}"
    echo "    CLAUDE_PROXY_KEY=${api_key}"
    echo ""
    echo "  Then run: docker compose up -d"
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

    # Stop service if running (to allow binary replacement)
    if systemctl is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
        info "Stopping service for upgrade..."
        systemctl stop "${SERVICE_NAME}"
    fi

    # Install binary
    install_binary "$version"

    # Install or update service
    if [[ "$is_upgrade" == false ]] || [[ ! -f "${SERVICE_FILE}" ]]; then
        # New installation - generate API key
        install_service
    else
        # Upgrade - preserve existing API key
        info "Service file already exists, preserving configuration..."
        INSTALLED_API_KEY=$(get_existing_api_key)
        if [[ -z "$INSTALLED_API_KEY" ]]; then
            warn "No API key found in service file, regenerating..."
            install_service
        fi
    fi

    # Enable and start/restart service
    info "Enabling service for auto-start..."
    systemctl enable "${SERVICE_NAME}" 2>/dev/null || true

    if systemctl is-active --quiet "${SERVICE_NAME}"; then
        info "Restarting service..."
        systemctl restart "${SERVICE_NAME}"
        success "Service restarted with new version"
    else
        info "Starting service..."
        systemctl start "${SERVICE_NAME}"
        success "Service started"
    fi

    # Wait a moment for service to fully start
    sleep 2

    echo ""
    success "Installation complete!"
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
    show_home_agent_config "$INSTALLED_API_KEY"
fi
