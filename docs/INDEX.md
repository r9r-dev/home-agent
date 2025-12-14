# Documentation Index

Complete documentation guide for Home Agent.

## Getting Started

Start here if you're new to the project:

1. **[README.md](../README.md)** - Project overview and main documentation
2. **[QUICKSTART.md](../QUICKSTART.md)** - Get running in 5 minutes
3. **[IMPLEMENTATION_SUMMARY.md](../IMPLEMENTATION_SUMMARY.md)** - What was built

## Development

For developers working on the project:

1. **[DEVELOPMENT.md](../DEVELOPMENT.md)** - Developer reference guide
   - Project structure
   - Common tasks
   - Code style
   - Debugging tips

2. **[ARCHITECTURE.md](../ARCHITECTURE.md)** - Technical architecture
   - System design
   - Component hierarchy
   - Data flow
   - Communication protocol

3. **[Frontend README](../frontend/README.md)** - Frontend-specific docs
   - Tech stack
   - Components
   - WebSocket protocol
   - Build instructions

## Testing

Quality assurance and testing:

1. **[TESTING.md](../TESTING.md)** - Testing procedures
   - Unit tests
   - Integration tests
   - Manual testing
   - Success criteria

## Deployment

Production deployment guides:

1. **[DEPLOYMENT.md](../DEPLOYMENT.md)** - Complete deployment guide
   - Server setup
   - SSL configuration
   - Monitoring
   - Troubleshooting

2. **[nginx.conf.example](../nginx.conf.example)** - Nginx configuration
3. **[home-agent.service.example](../home-agent.service.example)** - Systemd service

## Release Management

Version tracking and changes:

1. **[CHANGELOG.md](../CHANGELOG.md)** - Version history
   - Release notes
   - Breaking changes
   - Migration guides

## Quick Reference

### File Structure

```
home-agent/
├── README.md              # Main documentation
├── QUICKSTART.md          # Quick setup guide
├── DEVELOPMENT.md         # Developer guide
├── TESTING.md             # Testing guide
├── DEPLOYMENT.md          # Deployment guide
├── ARCHITECTURE.md        # Technical architecture
├── CHANGELOG.md           # Version history
├── IMPLEMENTATION_SUMMARY.md  # What was built
│
├── frontend/              # Frontend application
│   ├── README.md         # Frontend docs
│   ├── src/
│   │   ├── components/   # UI components
│   │   ├── services/     # WebSocket service
│   │   └── stores/       # State management
│   └── package.json
│
├── backend/               # Go backend
│   ├── main.go
│   ├── handlers/
│   ├── services/
│   └── models/
│
└── docs/                  # Additional documentation
    └── INDEX.md          # This file
```

### Common Commands

```bash
# Development
./start-dev.sh              # Start everything
cd frontend && npm run dev  # Frontend only
cd backend && go run main.go  # Backend only

# Building
./build-prod.sh             # Build for production
cd frontend && npm run build  # Frontend only
cd backend && go build      # Backend only

# Testing
cd frontend && npm run check  # Type check
cd backend && go test ./...   # Run tests

# Deployment
sudo systemctl start home-agent   # Start service
sudo systemctl status home-agent  # Check status
sudo journalctl -u home-agent -f  # View logs
```

## Documentation by Role

### For End Users
1. [README.md](../README.md) - Overview
2. [QUICKSTART.md](../QUICKSTART.md) - Setup

### For Developers
1. [DEVELOPMENT.md](../DEVELOPMENT.md) - Development guide
2. [ARCHITECTURE.md](../ARCHITECTURE.md) - Architecture
3. [Frontend README](../frontend/README.md) - Frontend details
4. [TESTING.md](../TESTING.md) - Testing

### For DevOps
1. [DEPLOYMENT.md](../DEPLOYMENT.md) - Deployment
2. [nginx.conf.example](../nginx.conf.example) - Nginx config
3. [home-agent.service.example](../home-agent.service.example) - Systemd service

### For Project Managers
1. [IMPLEMENTATION_SUMMARY.md](../IMPLEMENTATION_SUMMARY.md) - What was built
2. [CHANGELOG.md](../CHANGELOG.md) - Version history

## Topics

### Architecture
- [System Architecture](../ARCHITECTURE.md#overview)
- [Frontend Architecture](../ARCHITECTURE.md#frontend-architecture)
- [Backend Architecture](../ARCHITECTURE.md#backend-architecture)
- [Communication Protocol](../ARCHITECTURE.md#communication-protocol)

### Development
- [Project Structure](../DEVELOPMENT.md#project-structure)
- [Adding Components](../DEVELOPMENT.md#adding-a-new-frontend-component)
- [Styling Guidelines](../DEVELOPMENT.md#styling-guidelines)
- [Debugging](../DEVELOPMENT.md#debugging)

### Deployment
- [Server Setup](../DEPLOYMENT.md#server-setup)
- [SSL Configuration](../DEPLOYMENT.md#setup-ssl-certificate)
- [Monitoring](../DEPLOYMENT.md#monitoring)
- [Security](../DEPLOYMENT.md#security-hardening)

### Testing
- [Unit Testing](../TESTING.md#unit-testing)
- [Integration Testing](../TESTING.md#integration-testing)
- [Manual Testing](../TESTING.md#manual-testing-scenarios)
- [Performance Testing](../TESTING.md#performance-testing)

## API Reference

### WebSocket Protocol

**Client → Server:**
```json
{
  "type": "message",
  "content": "User message text"
}
```

**Server → Client:**
```json
{
  "type": "chunk",
  "content": "Response chunk"
}
```

See [ARCHITECTURE.md](../ARCHITECTURE.md#communication-protocol) for full protocol.

### Component API

See [Frontend README](../frontend/README.md#key-components) for component props and events.

## Configuration

### Environment Variables

**Frontend:**
```env
VITE_WS_URL=ws://localhost:8080/ws
```

**Backend:**
```env
ANTHROPIC_API_KEY=sk-ant-...
PORT=8080
HOST=localhost
DB_PATH=./home-agent.db
```

See [DEVELOPMENT.md](../DEVELOPMENT.md#environment-variables) for details.

## Troubleshooting

Common issues and solutions:

1. **WebSocket Connection Fails**
   - See [TESTING.md](../TESTING.md#test-2-websocket-connection)
   - See [DEPLOYMENT.md](../DEPLOYMENT.md#websocket-connection-fails)

2. **Build Errors**
   - See [DEVELOPMENT.md](../DEVELOPMENT.md#troubleshooting)
   - See [TESTING.md](../TESTING.md#debugging-tips)

3. **Deployment Issues**
   - See [DEPLOYMENT.md](../DEPLOYMENT.md#troubleshooting)

## Contributing

Guidelines for contributing:

1. Read [DEVELOPMENT.md](../DEVELOPMENT.md)
2. Follow code style guidelines
3. Write tests
4. Update documentation
5. Submit pull request

## Support

Getting help:

1. Check relevant documentation section
2. Search existing issues
3. Review troubleshooting guides
4. Open new issue with details

## Resources

### External Documentation
- [Svelte Documentation](https://svelte.dev/docs)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Go Documentation](https://golang.org/doc/)
- [WebSocket Protocol](https://tools.ietf.org/html/rfc6455)

### Tools
- [Svelte REPL](https://svelte.dev/repl)
- [Go Playground](https://play.golang.org/)
- [WebSocket Echo Test](https://websocket.org/echo.html)

## Documentation Updates

To update documentation:

1. Edit relevant markdown file
2. Update this index if needed
3. Update CHANGELOG.md
4. Commit changes

---

**Last Updated**: December 11, 2024
**Version**: 0.1.0

For questions about documentation, open an issue on GitHub.
