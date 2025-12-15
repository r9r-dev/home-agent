package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// APIKeyAuth returns a middleware that validates the X-API-Key header
func APIKeyAuth(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If no API key is configured, allow all requests
		if apiKey == "" {
			return c.Next()
		}

		providedKey := c.Get("X-API-Key")
		if providedKey == "" {
			log.Printf("Auth: Missing API key from %s", c.IP())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "API key required",
			})
		}

		if providedKey != apiKey {
			log.Printf("Auth: Invalid API key from %s", c.IP())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		return c.Next()
	}
}

// WebSocketAPIKeyAuth validates API key for WebSocket upgrade requests
func WebSocketAPIKeyAuth(apiKey string, c *fiber.Ctx) bool {
	if apiKey == "" {
		return true
	}

	providedKey := c.Get("X-API-Key")
	if providedKey != apiKey {
		log.Printf("WebSocket Auth: Invalid or missing API key from %s", c.IP())
		return false
	}

	return true
}
