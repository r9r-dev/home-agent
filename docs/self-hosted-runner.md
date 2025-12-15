# Self-Hosted GitHub Actions Runner

This guide explains how to set up a self-hosted GitHub Actions runner for faster CI/CD builds.

## Prerequisites

- Linux server (Ubuntu/Debian recommended)
- Docker installed and running
- Go 1.21+ installed
- At least 2GB RAM, 10GB disk space

## Installation

### 1. Create a dedicated user (recommended)

```bash
sudo useradd -m -s /bin/bash github-runner
sudo usermod -aG docker github-runner
sudo su - github-runner
```

### 2. Download the runner

Go to your repository on GitHub:
**Settings > Actions > Runners > New self-hosted runner**

GitHub will provide the exact commands. Example:

```bash
mkdir actions-runner && cd actions-runner

# Download (check GitHub for latest version)
curl -o actions-runner-linux-x64-2.321.0.tar.gz -L \
  https://github.com/actions/runner/releases/download/v2.321.0/actions-runner-linux-x64-2.321.0.tar.gz

tar xzf ./actions-runner-linux-x64-2.321.0.tar.gz
```

### 3. Configure the runner

```bash
# Token is provided by GitHub in the setup page
./config.sh --url https://github.com/r9r-dev/home-agent --token YOUR_TOKEN
```

When prompted:
- **Runner group**: Press Enter for default
- **Runner name**: Choose a descriptive name (e.g., `vps-runner`)
- **Labels**: Press Enter for default (or add custom labels)
- **Work folder**: Press Enter for default (`_work`)

### 4. Install as a service

```bash
sudo ./svc.sh install
sudo ./svc.sh start
```

Verify it's running:

```bash
sudo ./svc.sh status
```

## Required Tools

Install these tools on your runner:

### Go

```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Docker (if not already installed)

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker github-runner
```

### QEMU (for multi-arch builds)

```bash
sudo apt-get update
sudo apt-get install -y qemu-user-static binfmt-support
```

## Management Commands

```bash
# Check status
sudo ./svc.sh status

# Stop the runner
sudo ./svc.sh stop

# Start the runner
sudo ./svc.sh start

# Uninstall the service
sudo ./svc.sh uninstall
```

## Troubleshooting

### Runner not picking up jobs

1. Check the runner is online in GitHub Settings > Actions > Runners
2. Verify the service is running: `sudo ./svc.sh status`
3. Check logs: `journalctl -u actions.runner.*.service -f`

### Docker permission denied

```bash
sudo usermod -aG docker github-runner
# Then restart the runner service
sudo ./svc.sh stop
sudo ./svc.sh start
```

### Cache not working

The workflow uses local cache at `/tmp/.buildx-cache`. Ensure the runner has write access:

```bash
sudo mkdir -p /tmp/.buildx-cache
sudo chown github-runner:github-runner /tmp/.buildx-cache
```

## Security Considerations

- Only use self-hosted runners with private repositories or trusted public repositories
- The runner has access to your server, so ensure proper isolation
- Consider using Docker-in-Docker or rootless Docker for additional security
- Regularly update the runner software
