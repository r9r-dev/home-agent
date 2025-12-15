#!/bin/bash
set -e

# Claude Proxy Installation Script
# This script builds and installs the Claude Proxy service

INSTALL_DIR="/opt/claude-proxy"
SERVICE_FILE="/etc/systemd/system/claude-proxy.service"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "==================================="
echo "Claude Proxy Installation Script"
echo "==================================="
echo ""

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root (use sudo)"
   exit 1
fi

# Get the actual user (not root)
ACTUAL_USER="${SUDO_USER:-$USER}"
ACTUAL_GROUP=$(id -gn "$ACTUAL_USER")

echo "Installing for user: $ACTUAL_USER"
echo "Install directory: $INSTALL_DIR"
echo ""

# Build the binary
echo "Step 1: Building claude-proxy..."
cd "$PROJECT_DIR"
sudo -u "$ACTUAL_USER" go build -ldflags="-s -w" -o claude-proxy .
echo "Build complete!"
echo ""

# Create install directory
echo "Step 2: Creating installation directory..."
mkdir -p "$INSTALL_DIR"
cp claude-proxy "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/claude-proxy"
echo "Binary installed to $INSTALL_DIR/claude-proxy"
echo ""

# Update service file with correct user
echo "Step 3: Installing systemd service..."
cp "$SCRIPT_DIR/claude-proxy.service" "$SERVICE_FILE"

# Update the User and Group in the service file
sed -i "s/User=rlamour/User=$ACTUAL_USER/" "$SERVICE_FILE"
sed -i "s/Group=staff/Group=$ACTUAL_GROUP/" "$SERVICE_FILE"

systemctl daemon-reload
echo "Service file installed to $SERVICE_FILE"
echo ""

# Generate API key suggestion
API_KEY=$(openssl rand -hex 32)
echo "Step 4: Configuration"
echo ""
echo "To configure the service, edit: $SERVICE_FILE"
echo ""
echo "Suggested API key (add to Environment in service file):"
echo "  Environment=\"PROXY_API_KEY=$API_KEY\""
echo ""
echo "After editing, reload and restart the service:"
echo "  sudo systemctl daemon-reload"
echo "  sudo systemctl restart claude-proxy"
echo ""

# Enable and start service
echo "Step 5: Enabling and starting service..."
systemctl enable claude-proxy
systemctl start claude-proxy

# Show status
echo ""
echo "==================================="
echo "Installation Complete!"
echo "==================================="
echo ""
echo "Service status:"
systemctl status claude-proxy --no-pager || true
echo ""
echo "Useful commands:"
echo "  sudo systemctl status claude-proxy   # Check status"
echo "  sudo systemctl restart claude-proxy  # Restart service"
echo "  sudo journalctl -u claude-proxy -f   # View logs"
echo ""
echo "Test the service:"
echo "  curl http://localhost:9090/health"
echo ""
