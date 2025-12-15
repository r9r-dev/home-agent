package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/ronan/claude-proxy/handlers"
	"github.com/ronan/claude-proxy/middleware"
	"github.com/ronan/claude-proxy/services"
)

// Version is set at build time via -ldflags
var Version = "dev"

// Config holds the application configuration
type Config struct {
	Port      string
	Host      string
	ClaudeBin string
	APIKey    string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	config := Config{
		Port:      getEnv("PROXY_PORT", "9090"),
		Host:      getEnv("PROXY_HOST", "0.0.0.0"),
		ClaudeBin: getEnv("CLAUDE_BIN", "claude"),
		APIKey:    getEnv("PROXY_API_KEY", ""),
	}

	log.Println("Configuration loaded:")
	log.Printf("  Host: %s", config.Host)
	log.Printf("  Port: %s", config.Port)
	log.Printf("  Claude binary: %s", config.ClaudeBin)
	if config.APIKey != "" {
		log.Println("  API Key: configured")
	} else {
		log.Println("  API Key: not configured (no authentication)")
	}

	return config
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Handle --version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("claude-proxy %s\n", Version)
		os.Exit(0)
	}

	log.Println("Starting Claude Proxy Service...")
	log.Printf("Version: %s", Version)

	// Load configuration
	config := loadConfig()

	// Initialize Claude service
	claudeService := services.NewClaudeService(config.ClaudeBin)

	// Test Claude binary
	if err := claudeService.TestClaudeBinary(); err != nil {
		log.Printf("Warning: Claude binary test failed: %v", err)
		log.Println("Make sure the 'claude' CLI is installed and accessible")
	} else {
		log.Println("Claude binary test successful")
	}

	// Initialize handlers
	httpHandler := handlers.NewHTTPHandler(claudeService)
	wsHandler := handlers.NewWebSocketHandler(claudeService, config.APIKey)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Claude Proxy",
		DisableStartupMessage: false,
		EnablePrintRoutes:     true,
		ServerHeader:          "Claude Proxy",
	})

	// Middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, X-API-Key",
	}))

	// Health check (no auth required)
	app.Get("/health", httpHandler.HealthCheck)

	// API routes with authentication
	api := app.Group("/api", middleware.APIKeyAuth(config.APIKey))
	api.Post("/title", httpHandler.GenerateTitle)

	// WebSocket routes (auth handled in handler)
	wsHandler.RegisterRoutes(app)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("\nShutting down gracefully...")
		app.Shutdown()
	}()

	// Start server
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	log.Printf("Server starting on http://%s", addr)
	log.Printf("WebSocket endpoint: ws://%s/ws", addr)
	log.Printf("Health check: http://%s/health", addr)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server stopped")
}
