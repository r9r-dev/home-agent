# Docker CI/CD Setup

**Date**: 2025-12-15
**ID**: 01

## Summary

Setup GitHub Actions workflows for CI/CD with automatic Docker image publishing to GitHub Container Registry (ghcr.io).

## Changes

### Modified Files

- **Dockerfile**: Changed runtime base image from `alpine:latest` to `node:20-alpine` to support Claude CLI installation via npm. Replaced placeholder installation with `npm install -g @anthropic-ai/claude-code`.

- **DOCKER.md**: Updated CI/CD Integration section with actual workflow documentation. Added Production Deployment section with example docker-compose configuration.

### New Files

- **.github/workflows/ci.yml**: CI workflow that runs on push to main and pull requests.
  - Tests frontend (TypeScript type checking)
  - Tests backend (Go tests)
  - Builds Docker image (without push)
  - Uses GitHub Actions cache for faster builds

- **.github/workflows/release.yml**: Release workflow that runs on version tags (v*).
  - Builds and pushes Docker image to ghcr.io
  - Uses GITHUB_TOKEN for authentication (no PAT required)
  - Generates semantic version tags (e.g., 1.2.3, 1.2, latest)

## Security Considerations

- Uses GITHUB_TOKEN for GHCR authentication (automatic, no secrets to manage)
- Minimal permissions: `contents: read`, `packages: write`
- Production docker-compose not committed to repo (server-specific)
- Non-root user in Docker image (homeagent:1000)

## Usage

### Creating a Release

```bash
git tag v1.0.0
git push origin v1.0.0
```

### Pulling the Image

```bash
docker pull ghcr.io/r9r-dev/home-agent:latest
```
