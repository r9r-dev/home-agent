# Development Guide

Quick reference for developing Home Agent.

## Project Structure

```
home-agent/
├── backend/              # Go backend
│   ├── main.go          # Entry point
│   ├── handlers/        # Request handlers
│   ├── services/        # Business logic
│   ├── models/          # Data models
│   └── public/          # Built frontend (generated)
│
├── frontend/            # Svelte frontend
│   ├── src/
│   │   ├── components/  # UI components
│   │   ├── services/    # WebSocket, etc.
│   │   ├── stores/      # Svelte stores
│   │   ├── App.svelte   # Root component
│   │   └── main.ts      # Entry point
│   └── public/          # Static assets
│
├── start-dev.sh         # Development startup
├── build-prod.sh        # Production build
└── README.md            # Main documentation
```

## Quick Commands

### Development

```bash
# Start everything (recommended)
./start-dev.sh

# Or manually:

# Backend only
cd backend
go run main.go

# Frontend only (requires backend running)
cd frontend
npm run dev
```

### Building

```bash
# Build everything for production
./build-prod.sh

# Or manually:

# Frontend
cd frontend
npm run build

# Backend
cd backend
go build
```

### Testing

```bash
# Frontend type check
cd frontend
npm run check

# Backend test
cd backend
go test ./...
```

## Development Workflow

### Adding a New Frontend Component

1. Create component in `frontend/src/components/`
2. Import in parent component
3. Add types if needed
4. Test with `npm run check`

Example:
```typescript
// frontend/src/components/NewComponent.svelte
<script lang="ts">
  export let prop: string;
</script>

<div>{prop}</div>

<style>
  div { color: var(--color-text-primary); }
</style>
```

### Modifying WebSocket Protocol

1. Update backend handler (`backend/handlers/websocket.go`)
2. Update frontend service (`frontend/src/services/websocket.ts`)
3. Update types if needed
4. Test connection thoroughly

### Adding a New Store

1. Create in `frontend/src/stores/`
2. Export store and methods
3. Use in components with `$store` syntax

Example:
```typescript
// frontend/src/stores/newStore.ts
import { writable } from 'svelte/store';

export const newStore = writable(initialValue);
```

### Styling Guidelines

Use CSS variables from `app.css`:

```css
/* Colors */
--color-bg-primary
--color-bg-secondary
--color-text-primary
--color-primary

/* Spacing */
--spacing-sm, --spacing-md, --spacing-lg

/* Other */
--radius-md
--transition-normal
```

## Backend Development

### Adding a New Handler

1. Create in `backend/handlers/`
2. Register route in `main.go`
3. Handle errors properly

Example:
```go
// backend/handlers/new.go
func NewHandler(w http.ResponseWriter, r *http.Request) {
    // Handler logic
}

// backend/main.go
http.HandleFunc("/api/new", handlers.NewHandler)
```

### Working with Claude API

Edit `backend/services/claude.go`:

```go
func (s *ClaudeService) SendMessage(message string) (string, error) {
    // API call logic
}
```

### Database Changes

Edit `backend/models/database.go`:

```go
type Session struct {
    ID        string
    Messages  []Message
    CreatedAt time.Time
}
```

## Environment Variables

### Frontend (.env)

```env
VITE_WS_URL=ws://localhost:8080/ws
```

Access in code:
```typescript
const wsUrl = import.meta.env.VITE_WS_URL;
```

### Backend (.env)

```env
ANTHROPIC_API_KEY=sk-ant-...
PORT=8080
HOST=localhost
DB_PATH=./home-agent.db
```

Access in code:
```go
apiKey := os.Getenv("ANTHROPIC_API_KEY")
```

## Debugging

### Frontend

```bash
# Browser dev tools
- Console: errors and logs
- Network: WebSocket messages
- Elements: inspect styles
- Sources: breakpoints

# Svelte dev tools
Install browser extension
```

### Backend

```bash
# Add logging
log.Printf("Debug: %v", value)

# Run with race detector
go run -race main.go

# Debug with delve
dlv debug
```

### WebSocket Debugging

Use `wscat` to test WebSocket:

```bash
# Install
npm install -g wscat

# Connect
wscat -c ws://localhost:8080/ws

# Send message
> {"type": "message", "content": "Hello"}
```

## Common Tasks

### Update Dependencies

Frontend:
```bash
cd frontend
npm update
npm audit fix
```

Backend:
```bash
cd backend
go get -u ./...
go mod tidy
```

### Clean Build Artifacts

```bash
# Frontend
cd frontend
rm -rf node_modules dist

# Backend
cd backend
rm -f home-agent
rm -rf public
```

### Reset Database

```bash
rm backend/*.db backend/*.db-*
```

### Change Ports

Backend port in `.env`:
```env
PORT=3000
```

Frontend dev port in `vite.config.ts`:
```typescript
server: {
  port: 5173,
}
```

## Code Style

### Frontend (TypeScript/Svelte)

- Use TypeScript strict mode
- Prefer `const` over `let`
- Use descriptive variable names
- Add JSDoc comments for public functions
- Keep components small and focused
- Use stores for shared state

### Backend (Go)

- Follow Go conventions
- Use `gofmt` for formatting
- Handle all errors
- Add comments for exported functions
- Keep handlers thin
- Use services for business logic

## Git Workflow

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes
git add .
git commit -m "Add new feature"

# Push
git push origin feature/new-feature

# Create pull request
```

## Performance Tips

### Frontend

- Use Svelte's reactivity (`$:`)
- Avoid unnecessary re-renders
- Lazy load large components
- Optimize images and assets
- Use CSS transforms for animations

### Backend

- Use goroutines for concurrent operations
- Buffer channels appropriately
- Close connections properly
- Use connection pooling
- Cache when appropriate

## Troubleshooting

### "Cannot connect to WebSocket"

1. Check backend is running
2. Verify port 8080 is open
3. Check firewall settings
4. Verify VITE_WS_URL is correct

### "Frontend not loading"

1. Check `npm install` completed
2. Verify Node.js version (18+)
3. Clear `node_modules` and reinstall
4. Check browser console for errors

### "Backend build fails"

1. Check Go version (1.21+)
2. Run `go mod download`
3. Check for syntax errors
4. Verify all imports

### "Types not working"

1. Run `npm run check`
2. Check `tsconfig.json`
3. Verify all `.d.ts` files
4. Restart TypeScript server in IDE

## Resources

### Documentation

- [Svelte Docs](https://svelte.dev/docs)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Go Documentation](https://golang.org/doc/)
- [Vite Guide](https://vitejs.dev/guide/)

### Tools

- [Svelte REPL](https://svelte.dev/repl) - Test Svelte code
- [Go Playground](https://play.golang.org/) - Test Go code
- [WebSocket Echo](https://websocket.org/echo.html) - Test WebSocket

## Support

- Check [TESTING.md](TESTING.md) for test procedures
- See [README.md](README.md) for setup instructions
- Open an issue for bugs or questions
