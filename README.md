# Home Agent

A simple web interface to chat with Claude.

## Getting Started

### With Docker

```bash
docker pull ghcr.io/r9r-dev/home-agent:latest
docker run -d -p 8080:8080 -e ANTHROPIC_API_KEY=your_key ghcr.io/r9r-dev/home-agent:latest
```

Open http://localhost:8080

### Without Docker

```bash
cp .env.example .env
# Edit .env and add your ANTHROPIC_API_KEY

./start-dev.sh
```

Open http://localhost:5173

## Configuration

Create a `.env` file with:

```
ANTHROPIC_API_KEY=your_api_key_here
```

## Documentation

See [docs/](docs/INDEX.md) for detailed guides.

## Contributing

See [docs/development.md](docs/development.md) for setup instructions.

## License

MIT
