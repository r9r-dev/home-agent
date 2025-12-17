package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

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
	Port           string
	DatabasePath   string
	ClaudeBin      string
	PublicDir      string
	UploadDir      string // Directory for uploaded files (local path)
	WorkspacePath  string // Path prefix for Claude CLI (maps UploadDir for Claude's environment)
	ClaudeProxyURL string // If set, use proxy instead of local CLI
	ClaudeProxyKey string // API key for proxy authentication
}

// loadConfig loads configuration from environment variables with defaults
func loadConfig() Config {
	uploadDir := getEnv("UPLOAD_DIR", "./data/uploads")
	// Convert to absolute path for local file operations
	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		log.Printf("Warning: could not get absolute path for upload dir: %v", err)
		absUploadDir = uploadDir
	}

	// WORKSPACE_PATH is the path prefix that Claude CLI sees
	// Used when backend runs in container but Claude runs on host with mounted workspace
	// If not set, defaults to the absolute upload dir
	workspacePath := getEnv("WORKSPACE_PATH", "")

	config := Config{
		Port:           getEnv("PORT", "8080"),
		DatabasePath:   getEnv("DATABASE_PATH", "./data/homeagent.db"),
		ClaudeBin:      getEnv("CLAUDE_BIN", "claude"),
		PublicDir:      getEnv("PUBLIC_DIR", "./public"),
		UploadDir:      absUploadDir,
		WorkspacePath:  workspacePath,
		ClaudeProxyURL: getEnv("CLAUDE_PROXY_URL", ""),
		ClaudeProxyKey: getEnv("CLAUDE_PROXY_KEY", ""),
	}

	log.Println("Configuration loaded:")
	log.Printf("  Port: %s", config.Port)
	log.Printf("  Database: %s", config.DatabasePath)
	log.Printf("  Public directory: %s", config.PublicDir)
	log.Printf("  Upload directory: %s", config.UploadDir)
	if config.WorkspacePath != "" {
		log.Printf("  Workspace path (for Claude): %s", config.WorkspacePath)
	}

	if config.ClaudeProxyURL != "" {
		log.Printf("  Claude mode: proxy (%s)", config.ClaudeProxyURL)
		if config.ClaudeProxyKey != "" {
			log.Println("  Claude proxy key: configured")
		}
	} else {
		log.Printf("  Claude mode: local (%s)", config.ClaudeBin)
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

// ensureDirectories creates necessary directories if they don't exist
func ensureDirectories(config Config) error {
	// Ensure database directory exists
	dbDir := filepath.Dir(config.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}
	log.Printf("Database directory ensured: %s", dbDir)

	// Ensure upload directory exists
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}
	log.Printf("Upload directory ensured: %s", config.UploadDir)

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

	// Initialize Claude executor based on configuration
	var claudeExecutor services.ClaudeExecutor

	if config.ClaudeProxyURL != "" {
		// Use proxy executor for remote Claude CLI execution
		log.Printf("Initializing proxy Claude executor: %s", config.ClaudeProxyURL)
		claudeExecutor = services.NewProxyClaudeExecutor(services.ProxyConfig{
			ProxyURL: config.ClaudeProxyURL,
			APIKey:   config.ClaudeProxyKey,
			Timeout:  10 * time.Minute,
		})
	} else {
		// Use local executor for direct Claude CLI execution
		log.Printf("Initializing local Claude executor: %s", config.ClaudeBin)
		claudeExecutor = services.NewLocalClaudeExecutor(config.ClaudeBin)
	}

	// Test Claude connection
	if err := claudeExecutor.TestConnection(); err != nil {
		log.Printf("Warning: Claude executor test failed: %v", err)
		if config.ClaudeProxyURL != "" {
			log.Println("Make sure the Claude proxy service is running and accessible")
		} else {
			log.Println("Make sure the 'claude' CLI is installed and accessible")
		}
	}

	// Initialize handlers
	chatHandler := handlers.NewChatHandler(sessionManager, claudeExecutor, config.UploadDir, config.WorkspacePath, db)
	wsHandler := handlers.NewWebSocketHandler(chatHandler)
	uploadHandler := handlers.NewUploadHandler(config.UploadDir)
	memoryHandler := handlers.NewMemoryHandler(db)

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

	// Update session model
	app.Patch("/api/sessions/:id/model", func(c *fiber.Ctx) error {
		sessionID := c.Params("id")

		var body struct {
			Model string `json:"model"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// Validate model
		validModels := map[string]bool{"haiku": true, "sonnet": true, "opus": true}
		if !validModels[body.Model] {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid model. Must be one of: haiku, sonnet, opus"})
		}

		if err := sessionManager.UpdateSessionModel(sessionID, body.Model); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"session_id": sessionID, "model": body.Model})
	})

	// Settings API
	app.Get("/api/settings", func(c *fiber.Ctx) error {
		settings, err := db.GetAllSettings()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(settings)
	})

	app.Put("/api/settings/:key", func(c *fiber.Ctx) error {
		key := c.Params("key")

		var body struct {
			Value string `json:"value"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// Validate custom_instructions length (max 2000 chars)
		if key == "custom_instructions" && len(body.Value) > 2000 {
			return c.Status(400).JSON(fiber.Map{"error": "Custom instructions must be 2000 characters or less"})
		}

		if err := db.SetSetting(key, body.Value); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"key": key, "value": body.Value})
	})

	// System prompt API (for preview in frontend)
	app.Get("/api/system-prompt", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"prompt": services.GetSystemPrompt()})
	})

	// Register WebSocket routes
	wsHandler.RegisterRoutes(app)

	// Register upload routes
	uploadHandler.RegisterRoutes(app)

	// Register memory routes
	memoryHandler.RegisterRoutes(app)

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
