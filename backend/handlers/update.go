package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	gorillaws "github.com/gorilla/websocket"
)

// UpdateHandler handles update-related requests
type UpdateHandler struct {
	proxyURL string
	apiKey   string
	client   *http.Client
}

// NewUpdateHandler creates a new update handler
func NewUpdateHandler(proxyURL, apiKey string) *UpdateHandler {
	return &UpdateHandler{
		proxyURL: proxyURL,
		apiKey:   apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// VersionInfo represents version information for a component
type VersionInfo struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"updateAvailable"`
}

// UpdateStatus represents the update status response
type UpdateStatus struct {
	Backend VersionInfo `json:"backend"`
	Proxy   VersionInfo `json:"proxy"`
}

// CheckForUpdates checks for available updates
func (h *UpdateHandler) CheckForUpdates(c *fiber.Ctx) error {
	req, err := http.NewRequest("GET", h.proxyURL+"/api/update/check", nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request"})
	}

	if h.apiKey != "" {
		req.Header.Set("X-API-Key", h.apiKey)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "Failed to connect to proxy"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": string(body)})
	}

	var status UpdateStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse response"})
	}

	return c.JSON(status)
}

// StartBackendUpdate triggers backend update
func (h *UpdateHandler) StartBackendUpdate(c *fiber.Ctx) error {
	req, err := http.NewRequest("POST", h.proxyURL+"/api/update/backend", nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request"})
	}

	if h.apiKey != "" {
		req.Header.Set("X-API-Key", h.apiKey)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "Failed to connect to proxy"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": string(body)})
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse response"})
	}

	return c.JSON(result)
}

// StartProxyUpdate triggers proxy update
func (h *UpdateHandler) StartProxyUpdate(c *fiber.Ctx) error {
	req, err := http.NewRequest("POST", h.proxyURL+"/api/update/proxy", nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request"})
	}

	if h.apiKey != "" {
		req.Header.Set("X-API-Key", h.apiKey)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "Failed to connect to proxy"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": string(body)})
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse response"})
	}

	return c.JSON(result)
}

// RegisterRoutes registers update routes
func (h *UpdateHandler) RegisterRoutes(app *fiber.App) {
	// Check for updates
	app.Get("/api/update/check", h.CheckForUpdates)

	// Start backend update
	app.Post("/api/update/backend", h.StartBackendUpdate)

	// Start proxy update
	app.Post("/api/update/proxy", h.StartProxyUpdate)

	// WebSocket for update logs
	app.Use("/ws/update", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/update", websocket.New(h.WebSocketUpdateLogs))
}

// WebSocketUpdateLogs proxies WebSocket connection for update logs
func (h *UpdateHandler) WebSocketUpdateLogs(c *websocket.Conn) {
	defer c.Close()

	// Build the proxy WebSocket URL
	proxyWSURL := h.proxyURL
	// Convert http:// to ws:// or https:// to wss://
	if len(proxyWSURL) > 7 && proxyWSURL[:7] == "http://" {
		proxyWSURL = "ws://" + proxyWSURL[7:]
	} else if len(proxyWSURL) > 8 && proxyWSURL[:8] == "https://" {
		proxyWSURL = "wss://" + proxyWSURL[8:]
	}
	proxyWSURL = proxyWSURL + "/ws/update"

	// Add API key as query parameter
	if h.apiKey != "" {
		proxyWSURL = proxyWSURL + "?key=" + h.apiKey
	}

	// Connect to proxy WebSocket
	dialer := gorillaws.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	proxyConn, resp, err := dialer.Dial(proxyWSURL, nil)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to connect to proxy: %v", err)
		if resp != nil {
			errMsg = fmt.Sprintf("%s (status: %d)", errMsg, resp.StatusCode)
		}
		c.WriteJSON(fiber.Map{"type": "error", "error": errMsg})
		return
	}
	defer proxyConn.Close()

	// Proxy messages from backend proxy to client
	done := make(chan struct{})

	// Read from proxy, write to client
	go func() {
		defer close(done)
		for {
			_, message, err := proxyConn.ReadMessage()
			if err != nil {
				return
			}
			if err := c.WriteMessage(gorillaws.TextMessage, message); err != nil {
				return
			}
		}
	}()

	// Read from client, write to proxy (for any commands)
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				proxyConn.Close()
				return
			}
			if err := proxyConn.WriteMessage(gorillaws.TextMessage, message); err != nil {
				return
			}
		}
	}()

	<-done
}
