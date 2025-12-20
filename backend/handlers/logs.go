package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/ronan/home-agent/services"
)

// LogHandler handles log-related endpoints
type LogHandler struct {
	logService *services.LogService
}

// NewLogHandler creates a new LogHandler
func NewLogHandler(logService *services.LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

// LogStatusResponse represents the log status response
type LogStatusResponse struct {
	Status  services.LogLevel `json:"status"`
	Entries []services.LogEntry `json:"entries"`
}

// RegisterRoutes registers log-related routes
func (lh *LogHandler) RegisterRoutes(app *fiber.App) {
	// Get all logs and current status
	app.Get("/api/logs", func(c *fiber.Ctx) error {
		return c.JSON(LogStatusResponse{
			Status:  lh.logService.GetStatus(),
			Entries: lh.logService.GetEntries(),
		})
	})

	// Get just the status (for polling indicator)
	app.Get("/api/logs/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": lh.logService.GetStatus(),
		})
	})

	// Clear the status indicators (acknowledge warnings/errors)
	app.Post("/api/logs/clear", func(c *fiber.Ctx) error {
		lh.logService.ClearStatus()
		return c.JSON(fiber.Map{"cleared": true})
	})

	// WebSocket for real-time log streaming
	app.Use("/ws/logs", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/logs", websocket.New(lh.handleLogWebSocket))
}

// handleLogWebSocket streams logs to connected clients
func (lh *LogHandler) handleLogWebSocket(c *websocket.Conn) {
	// Subscribe to log entries
	sub := lh.logService.Subscribe()
	defer lh.logService.Unsubscribe(sub)

	// Send current status immediately
	status := struct {
		Type   string            `json:"type"`
		Status services.LogLevel `json:"status"`
	}{
		Type:   "status",
		Status: lh.logService.GetStatus(),
	}
	if data, err := json.Marshal(status); err == nil {
		c.WriteMessage(websocket.TextMessage, data)
	}

	// Set up ping/pong for connection health
	c.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Channel to signal close
	done := make(chan struct{})

	// Goroutine to read messages (needed for ping/pong)
	go func() {
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				close(done)
				return
			}
		}
	}()

	// Ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Stream logs to client
	for {
		select {
		case entry, ok := <-sub:
			if !ok {
				return
			}

			msg := struct {
				Type  string              `json:"type"`
				Entry services.LogEntry   `json:"entry"`
			}{
				Type:  "log",
				Entry: entry,
			}

			if data, err := json.Marshal(msg); err == nil {
				if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
					return
				}
			}

		case <-ticker.C:
			if err := c.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				return
			}

		case <-done:
			return
		}
	}
}
