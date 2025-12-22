# Clean up main.go

**Priority:** P3 (Low)
**Type:** Refactoring
**Component:** Backend
**Estimated Effort:** Medium

## Summary

Refactor `main.go` (388 lines) to separate concerns: configuration, dependency injection, routing, and server lifecycle.

## Current Issues

- Inline route handlers mixed with route definitions
- Configuration loading mixed with app setup
- No dependency injection pattern
- Hard to test individual components
- Middleware setup interleaved with routes

## Proposed Structure

```go
// cmd/server/main.go - reduced to ~80 lines
package main

import (
    "log"

    "home-agent/internal/config"
    "home-agent/internal/server"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    // Initialize application
    app, cleanup, err := server.Initialize(cfg)
    if err != nil {
        log.Fatalf("Initialization error: %v", err)
    }
    defer cleanup()

    // Start server with graceful shutdown
    server.Run(app, cfg.Server.Port)
}
```

```go
// internal/server/server.go
package server

func Initialize(cfg *config.Config) (*fiber.App, func(), error) {
    // Initialize database
    db, err := database.New(cfg.Database)
    if err != nil {
        return nil, nil, err
    }

    // Initialize services
    services := initServices(db, cfg)

    // Initialize handlers
    handlers := initHandlers(services)

    // Create Fiber app
    app := fiber.New(fiber.Config{...})

    // Setup middleware
    setupMiddleware(app, cfg)

    // Register routes
    RegisterRoutes(app, handlers)

    cleanup := func() {
        db.Close()
    }

    return app, cleanup, nil
}
```

```go
// internal/server/routes.go
package server

func RegisterRoutes(app *fiber.App, h *Handlers) {
    // Health check
    app.Get("/health", h.Health.Check)

    // API routes
    api := app.Group("/api")

    // Sessions
    sessions := api.Group("/sessions")
    sessions.Get("/", h.Session.List)
    sessions.Get("/:id", h.Session.Get)
    sessions.Get("/:id/messages", h.Session.GetMessages)
    sessions.Patch("/:id/model", h.Session.UpdateModel)
    sessions.Delete("/:id", h.Session.Delete)

    // Memory
    memory := api.Group("/memory")
    memory.Get("/", h.Memory.List)
    memory.Post("/", h.Memory.Create)
    memory.Get("/:id", h.Memory.Get)
    memory.Put("/:id", h.Memory.Update)
    memory.Delete("/:id", h.Memory.Delete)
    memory.Get("/export", h.Memory.Export)
    memory.Post("/import", h.Memory.Import)

    // ... more routes

    // WebSocket
    app.Get("/ws", h.WebSocket.Handle)

    // Static files (frontend)
    app.Static("/", h.Config.PublicDir)
}
```

```go
// internal/server/middleware.go
package server

func setupMiddleware(app *fiber.App, cfg *config.Config) {
    // CORS
    app.Use(cors.New(cors.Config{
        AllowOrigins: "*",
        AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
    }))

    // Logger
    app.Use(logger.New())

    // Recover
    app.Use(recover.New())

    // Custom error handler
    app.Use(errorHandler)
}
```

## Proposed Directory Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Slim entry point (~80 lines)
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration loading
│   ├── server/
│   │   ├── server.go            # App initialization
│   │   ├── routes.go            # Route registration
│   │   ├── middleware.go        # Middleware setup
│   │   └── handlers.go          # Handler struct
│   └── ...
```

## Tasks

- [ ] Create `internal/server/` package
- [ ] Extract route registration to `routes.go`
- [ ] Extract middleware setup to `middleware.go`
- [ ] Create `Handlers` struct for dependency injection
- [ ] Create `Initialize()` function
- [ ] Simplify `main.go` to initialization only
- [ ] Add graceful shutdown handling
- [ ] Update imports throughout codebase

## Acceptance Criteria

- [ ] `main.go` under 100 lines
- [ ] Routes defined in dedicated file
- [ ] Middleware configurable separately
- [ ] Clear dependency injection pattern
- [ ] Graceful shutdown works correctly

## References

- `ARCHITECTURE_REVIEW.md` section "4. Clean Up `main.go`"
- Current file: `backend/main.go`
- Depends on: Issue #010 (Configuration)

## Labels

```
priority: P3
type: refactoring
component: backend
```
