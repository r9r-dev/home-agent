#!/bin/bash
#
# Claude Proxy SDK Installer
# Install or update Claude Proxy SDK (TypeScript) service on Linux
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/ronalove/home-agent/main/claude-proxy-sdk/install.sh | sudo bash
#
# Or with options:
#   curl -fsSL ... | sudo bash -s -- --uninstall
#

set -e

# Configuration
GITHUB_REPO="ronalove/home-agent"
SERVICE_NAME="claude-proxy-sdk"
INSTALL_DIR="/opt/claude-proxy-sdk"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

# Global variable to store the API key
INSTALLED_API_KEY=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Print banner
banner() {
    echo ""
    echo "  ╔═══════════════════════════════════════╗"
    echo "  ║     Claude Proxy SDK Installer        ║"
    echo "  ║   TypeScript + Claude Agent SDK       ║"
    echo "  ║             Powered by Bun            ║"
    echo "  ╚═══════════════════════════════════════╝"
    echo ""
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root (use sudo)"
    fi
}

# Check Bun
check_bun() {
    if ! command -v bun &> /dev/null; then
        warn "Bun not found"
        install_bun
        return
    fi

    success "Bun $(bun --version) found"
}

# Install Bun
install_bun() {
    info "Installing Bun..."

    # Get the user who ran sudo (install for them, not root)
    local run_user="${SUDO_USER:-root}"
    local run_home=$(eval echo ~$run_user)

    # Install Bun for the user
    if [[ "$run_user" != "root" ]]; then
        su - "$run_user" -c 'curl -fsSL https://bun.sh/install | bash'
        # Add to PATH for this session
        export PATH="$run_home/.bun/bin:$PATH"
    else
        curl -fsSL https://bun.sh/install | bash
        export PATH="/root/.bun/bin:$PATH"
    fi

    # Create symlink for system-wide access
    if [[ -f "$run_home/.bun/bin/bun" ]]; then
        ln -sf "$run_home/.bun/bin/bun" /usr/local/bin/bun
    fi

    success "Bun installed: $(bun --version)"
}

# Check dependencies
check_deps() {
    local missing=()

    for cmd in curl git systemctl; do
        if ! command -v $cmd &> /dev/null; then
            missing+=($cmd)
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        error "Missing dependencies: ${missing[*]}"
    fi
}

# Download source code
download_source() {
    info "Downloading source code..."

    # Create temp directory
    local tmp_dir=$(mktemp -d)
    trap "rm -rf ${tmp_dir}" EXIT

    # Clone only the claude-proxy-sdk directory
    cd "${tmp_dir}"
    git clone --depth 1 --filter=blob:none --sparse "https://github.com/${GITHUB_REPO}.git" repo
    cd repo
    git sparse-checkout set claude-proxy-sdk

    # Copy to install directory
    info "Installing to ${INSTALL_DIR}..."
    rm -rf "${INSTALL_DIR}"
    mkdir -p "${INSTALL_DIR}"
    cp -r claude-proxy-sdk/* "${INSTALL_DIR}/"

    success "Source code installed"
}

# Install dependencies with Bun
install_deps() {
    info "Installing dependencies with Bun..."
    cd "${INSTALL_DIR}"

    # Install all dependencies
    bun install

    # IMPORTANT: Fix ownership AFTER bun install
    # bun runs as root (via sudo) and creates files owned by root
    # The service user needs to own these files for self-updates
    local run_user="${SUDO_USER:-root}"
    local run_group=$(id -gn "$run_user" 2>/dev/null || echo "root")
    info "Setting ownership to ${run_user}:${run_group}..."
    chown -R "${run_user}:${run_group}" "${INSTALL_DIR}"

    success "Dependencies installed"
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

# Get bun path
get_bun_path() {
    local run_user="${SUDO_USER:-root}"
    local run_home=$(eval echo ~$run_user)

    # Check common locations
    if [[ -f "/usr/local/bin/bun" ]]; then
        echo "/usr/local/bin/bun"
    elif [[ -f "$run_home/.bun/bin/bun" ]]; then
        echo "$run_home/.bun/bin/bun"
    elif command -v bun &> /dev/null; then
        which bun
    else
        echo "/usr/local/bin/bun"
    fi
}

# Install systemd service
install_service() {
    info "Installing systemd service..."

    # Get the user who ran sudo
    local run_user="${SUDO_USER:-root}"
    local run_group=$(id -gn "$run_user" 2>/dev/null || echo "root")
    local run_home=$(eval echo ~$run_user)
    local bun_path=$(get_bun_path)

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
Description=Claude Proxy SDK - Claude Agent SDK Gateway
Documentation=https://github.com/${GITHUB_REPO}
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=${run_user}
Group=${run_group}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${bun_path} ${INSTALL_DIR}/src/index.ts

# Environment
Environment="PROXY_PORT=9090"
Environment="PROXY_HOST=0.0.0.0"
Environment="PROXY_API_KEY=${INSTALLED_API_KEY}"
Environment="HOME=${run_home}"
Environment="PATH=${run_home}/.bun/bin:/usr/local/bin:/usr/bin:/bin"

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

# Start service
start_service() {
    info "Enabling and starting service..."
    systemctl enable "${SERVICE_NAME}" 2>/dev/null || true
    systemctl restart "${SERVICE_NAME}"
    success "Service started"
}

# Detect primary IP address
detect_ip() {
    local ip=""

    if command -v ip &> /dev/null; then
        ip=$(ip route get 1 2>/dev/null | awk '{print $7; exit}')
    fi

    if [[ -z "$ip" ]] && command -v hostname &> /dev/null; then
        ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi

    if [[ -z "$ip" ]]; then
        ip="<YOUR_HOST_IP>"
    fi

    echo "$ip"
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
}

# Uninstall
uninstall() {
    info "Uninstalling ${SERVICE_NAME}..."

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
    local is_upgrade=false

    # Check if already installed
    if [[ -d "${INSTALL_DIR}" ]]; then
        is_upgrade=true
        info "Existing installation found, upgrading..."
    fi

    # Stop service if running
    if systemctl is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
        info "Stopping service for upgrade..."
        systemctl stop "${SERVICE_NAME}"
    fi

    # Download and install
    download_source
    install_deps

    # Install or update service
    if [[ "$is_upgrade" == false ]] || [[ ! -f "${SERVICE_FILE}" ]]; then
        install_service
    else
        info "Service file exists, preserving configuration..."
        INSTALLED_API_KEY=$(get_existing_api_key)
        if [[ -z "$INSTALLED_API_KEY" ]]; then
            warn "No API key found, regenerating..."
            install_service
        fi
    fi

    # Start service
    start_service

    # Wait for service to start
    sleep 2

    echo ""
    success "Installation complete!"
}

# Parse arguments
UNINSTALL=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --uninstall|-u)
            UNINSTALL=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --uninstall, -u   Uninstall claude-proxy-sdk"
            echo "  --help, -h        Show this help"
            echo ""
            echo "Examples:"
            echo "  curl -fsSL URL | sudo bash"
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
check_bun

if [[ "$UNINSTALL" == true ]]; then
    uninstall
else
    install
    show_status
    show_home_agent_config "$INSTALLED_API_KEY"
fi
