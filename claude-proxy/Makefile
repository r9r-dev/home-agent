.PHONY: build run clean install test

# Build variables
BINARY_NAME=claude-proxy
BUILD_DIR=./build
INSTALL_DIR=/opt/claude-proxy

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the service locally
run:
	go run .

# Run with environment variables
run-dev:
	PROXY_PORT=9090 PROXY_HOST=127.0.0.1 go run .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Run tests
test:
	go test ./...

# Install to system (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	sudo mkdir -p $(INSTALL_DIR)
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installing systemd service..."
	sudo cp deploy/claude-proxy.service /etc/systemd/system/
	sudo systemctl daemon-reload
	@echo "Installation complete!"
	@echo ""
	@echo "To start the service:"
	@echo "  sudo systemctl enable claude-proxy"
	@echo "  sudo systemctl start claude-proxy"
	@echo ""
	@echo "To check status:"
	@echo "  sudo systemctl status claude-proxy"

# Uninstall from system (requires sudo)
uninstall:
	@echo "Stopping and removing $(BINARY_NAME)..."
	-sudo systemctl stop claude-proxy
	-sudo systemctl disable claude-proxy
	-sudo rm /etc/systemd/system/claude-proxy.service
	-sudo rm -rf $(INSTALL_DIR)
	sudo systemctl daemon-reload
	@echo "Uninstallation complete!"

# Download dependencies
deps:
	go mod download
	go mod tidy
