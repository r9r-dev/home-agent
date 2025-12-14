# Quick Start Guide

Get Home Agent running in 5 minutes.

## Prerequisites

- Go 1.21+ installed
- Node.js 18+ installed
- Claude API key

## Setup

### 1. Clone & Navigate

```bash
cd home-agent
```

### 2. Configure Environment

```bash
# Copy example env file
cp .env.example .env

# Edit with your API key
# .env should contain:
ANTHROPIC_API_KEY=your_api_key_here
PORT=8080
```

### 3. Run Development

**Option A: Automatic (Recommended)**

```bash
./start-dev.sh
```

This will:
- Install all dependencies
- Build backend
- Start backend on port 8080
- Start frontend dev server on port 5173

**Option B: Manual**

Terminal 1 (Backend):
```bash
cd backend
go mod download
go run main.go
```

Terminal 2 (Frontend):
```bash
cd frontend
npm install
npm run dev
```

### 4. Open Browser

Development: http://localhost:5173
Or production: http://localhost:8080

### 5. Start Chatting

- Type a message in the input box
- Press Enter to send
- Watch the response stream in

## Common Commands

```bash
# Development
./start-dev.sh              # Start everything

# Production
./build-prod.sh             # Build for production
cd backend && ./home-agent  # Run production server

# Testing
cd frontend && npm run check    # Type check
cd backend && go test ./...     # Run tests
```

## Troubleshooting

### Port already in use

Change port in `.env`:
```env
PORT=3000
```

### WebSocket connection failed

1. Ensure backend is running
2. Check firewall settings
3. Verify no proxy blocking WebSocket

### Frontend won't start

```bash
cd frontend
rm -rf node_modules
npm install
```

### Backend won't start

```bash
cd backend
go mod download
go mod tidy
```

## Next Steps

- Read [README.md](README.md) for full documentation
- Check [DEVELOPMENT.md](DEVELOPMENT.md) for development guide
- See [TESTING.md](TESTING.md) for testing procedures

## Quick Tips

- Use Shift+Enter for new lines in messages
- Scroll behavior: auto-scrolls when at bottom
- Connection reconnects automatically if lost
- Markdown and code syntax highlighting supported

## Need Help?

1. Check the logs in terminal
2. Open browser console (F12)
3. Review documentation files
4. Open an issue on GitHub

Enjoy using Home Agent!
