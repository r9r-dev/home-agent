package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/ronan/home-agent/handlers"
	"github.com/ronan/home-agent/models"
	"github.com/ronan/home-agent/services"
)

// Config holds the application configuration
type Config struct {
	Port         string
	DatabasePath string
	ClaudeBin    string
	PublicDir    string
}

// loadConfig loads configuration from environment variables with defaults
func loadConfig() Config {
	config := Config{
		Port:         getEnv("PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./data/homeagent.db"),
		ClaudeBin:    getEnv("CLAUDE_BIN", "claude"),
		PublicDir:    getEnv("PUBLIC_DIR", "./public"),
	}

	log.Println("Configuration loaded:")
	log.Printf("  Port: %s", config.Port)
	log.Printf("  Database: %s", config.DatabasePath)
	log.Printf("  Claude binary: %s", config.ClaudeBin)
	log.Printf("  Public directory: %s", config.PublicDir)

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

// ensureDirectories creates necessary directories if they don't exist
func ensureDirectories(config Config) error {
	// Ensure database directory exists
	dbDir := filepath.Dir(config.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	log.Printf("Database directory ensured: %s", dbDir)
	return nil
}

func main() {
	log.Println("Starting Home Agent backend...")

	// Load configuration
	config := loadConfig()

	// Ensure necessary directories exist
	if err := ensureDirectories(config); err != nil {
		log.Fatalf("Failed to ensure directories: %v", err)
	}

	// Initialize database
	db, err := models.InitDB(config.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize services
	sessionManager := services.NewSessionManager(db)
	claudeService := services.NewClaudeService(config.ClaudeBin)

	// Test Claude binary
	if err := claudeService.TestClaudeBinary(); err != nil {
		log.Printf("Warning: Claude binary test failed: %v", err)
		log.Println("Make sure the 'claude' CLI is installed and accessible")
	}

	// Initialize handlers
	chatHandler := handlers.NewChatHandler(sessionManager, claudeService)
	wsHandler := handlers.NewWebSocketHandler(chatHandler)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Home Agent",
		DisableStartupMessage: false,
		EnablePrintRoutes:     true,
		ServerHeader:          "Home Agent",
		ErrorHandler:          customErrorHandler,
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
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "home-agent",
		})
	})

	// API info endpoint
	app.Get("/api/info", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "Home Agent API",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"websocket": "/ws",
				"health":    "/health",
			},
		})
	})

	// Sessions API
	app.Get("/api/sessions", func(c *fiber.Ctx) error {
		sessions, err := sessionManager.ListSessions()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(sessions)
	})

	app.Get("/api/sessions/:id", func(c *fiber.Ctx) error {
		sessionID := c.Params("id")
		session, err := sessionManager.GetSession(sessionID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(session)
	})

	app.Get("/api/sessions/:id/messages", func(c *fiber.Ctx) error {
		sessionID := c.Params("id")
		messages, err := sessionManager.GetMessages(sessionID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(messages)
	})

	app.Delete("/api/sessions/:id", func(c *fiber.Ctx) error {
		sessionID := c.Params("id")
		if err := sessionManager.DeleteSession(sessionID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"deleted": sessionID})
	})

	// Register WebSocket routes
	wsHandler.RegisterRoutes(app)

	// Serve static files from public directory (for frontend)
	// This should be last so WebSocket and API routes take precedence
	if _, err := os.Stat(config.PublicDir); err == nil {
		log.Printf("Serving static files from: %s", config.PublicDir)
		app.Static("/", config.PublicDir)

		// Serve index.html for SPA routing
		app.Get("/*", func(c *fiber.Ctx) error {
			return c.SendFile(filepath.Join(config.PublicDir, "index.html"))
		})
	} else {
		log.Printf("Warning: Public directory not found: %s", config.PublicDir)
		log.Println("Static file serving disabled")
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("\nShutting down gracefully...")
		app.Shutdown()
	}()

	// Start server
	addr := fmt.Sprintf(":%s", config.Port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Println("WebSocket endpoint: ws://localhost" + addr + "/ws")

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server stopped")
}

// customErrorHandler handles Fiber errors
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	log.Printf("Error: %v (status: %d)", err, code)

	return c.Status(code).JSON(fiber.Map{
		"error":  true,
		"status": code,
		"message": err.Error(),
	})
}
