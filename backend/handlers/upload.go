package handlers

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UploadHandler handles file uploads
type UploadHandler struct {
	uploadDir string
	maxSize   int64 // Max file size in bytes
}

// UploadResponse represents the response after a successful upload
type UploadResponse struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Type     string `json:"type"` // "image" or "file"
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// NewUploadHandler creates a new UploadHandler
func NewUploadHandler(uploadDir string) *UploadHandler {
	return &UploadHandler{
		uploadDir: uploadDir,
		maxSize:   10 * 1024 * 1024, // 10MB default
	}
}

// AllowedMimeTypes defines allowed file types
var AllowedMimeTypes = map[string]string{
	// Images
	"image/png":  "image",
	"image/jpeg": "image",
	"image/jpg":  "image",
	"image/gif":  "image",
	"image/webp": "image",
	// Documents
	"application/pdf":  "file",
	"text/plain":       "file",
	"text/markdown":    "file",
	"application/json": "file",
	"text/csv":         "file",
	"application/xml":  "file",
	"text/xml":         "file",
	"text/yaml":        "file",
	"application/x-yaml": "file",
	// Code files (often sent as text/plain or application/octet-stream)
	"text/html":       "file",
	"text/css":        "file",
	"text/javascript": "file",
	"application/javascript": "file",
}

// AllowedExtensions for files that might not have proper MIME types
var AllowedExtensions = map[string]string{
	".png":  "image",
	".jpg":  "image",
	".jpeg": "image",
	".gif":  "image",
	".webp": "image",
	".pdf":  "file",
	".txt":  "file",
	".md":   "file",
	".json": "file",
	".csv":  "file",
	".xml":  "file",
	".yaml": "file",
	".yml":  "file",
	".html": "file",
	".css":  "file",
	".js":   "file",
	".ts":   "file",
	".go":   "file",
	".py":   "file",
	".rs":   "file",
	".java": "file",
	".c":    "file",
	".cpp":  "file",
	".h":    "file",
	".sh":   "file",
	".sql":  "file",
	".log":  "file",
}

// RegisterRoutes registers upload routes
func (h *UploadHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/api/upload", h.HandleUpload)
	app.Get("/api/uploads/:sessionId/:filename", h.ServeFile)
	app.Delete("/api/uploads/:id", h.DeleteFile)
}

// HandleUpload handles file upload requests
func (h *UploadHandler) HandleUpload(c *fiber.Ctx) error {
	// Get session ID from form (optional, will use "temp" if not provided)
	sessionID := c.FormValue("session_id")
	if sessionID == "" {
		sessionID = "temp"
	}

	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided",
		})
	}

	// Check file size
	if file.Size > h.maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("File too large. Maximum size is %d MB", h.maxSize/(1024*1024)),
		})
	}

	// Get MIME type and validate
	mimeType := file.Header.Get("Content-Type")
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// Determine file type (image or file)
	fileType := ""

	// First try MIME type
	if t, ok := AllowedMimeTypes[mimeType]; ok {
		fileType = t
	}

	// Fall back to extension if MIME type not recognized
	if fileType == "" {
		if t, ok := AllowedExtensions[ext]; ok {
			fileType = t
		}
	}

	// Reject if type is not recognized
	if fileType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File type not allowed",
		})
	}

	// Generate unique ID for the file
	fileID := uuid.New().String()

	// Create safe filename
	safeFilename := fmt.Sprintf("%s-%s", fileID[:8], sanitizeFilename(file.Filename))

	// Create session upload directory
	sessionDir := filepath.Join(h.uploadDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		log.Printf("Failed to create upload directory: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	// Full path for the file
	filePath := filepath.Join(sessionDir, safeFilename)

	// Save the file
	if err := c.SaveFile(file, filePath); err != nil {
		log.Printf("Failed to save file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	log.Printf("File uploaded: %s (session: %s, type: %s, size: %d)", safeFilename, sessionID, fileType, file.Size)

	// Return upload response
	return c.JSON(UploadResponse{
		ID:       fileID,
		Filename: file.Filename,
		Path:     fmt.Sprintf("/api/uploads/%s/%s", sessionID, safeFilename),
		Type:     fileType,
		Size:     file.Size,
		MimeType: mimeType,
	})
}

// ServeFile serves uploaded files
func (h *UploadHandler) ServeFile(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	filename := c.Params("filename")

	// Sanitize path to prevent directory traversal
	if strings.Contains(sessionID, "..") || strings.Contains(filename, "..") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid path",
		})
	}

	filePath := filepath.Join(h.uploadDir, sessionID, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	return c.SendFile(filePath)
}

// DeleteFile deletes an uploaded file
func (h *UploadHandler) DeleteFile(c *fiber.Ctx) error {
	fileID := c.Params("id")
	sessionID := c.Query("session_id", "temp")

	// Find and delete the file
	sessionDir := filepath.Join(h.uploadDir, sessionID)

	files, err := os.ReadDir(sessionDir)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileID[:8]+"-") {
			filePath := filepath.Join(sessionDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to delete file: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to delete file",
				})
			}
			log.Printf("File deleted: %s", filePath)
			return c.JSON(fiber.Map{"deleted": fileID})
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "File not found",
	})
}

// sanitizeFilename removes potentially dangerous characters from filename
func sanitizeFilename(filename string) string {
	// Get just the base filename (no path)
	filename = filepath.Base(filename)

	// Replace spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")

	// Remove any characters that are not alphanumeric, dash, underscore, or dot
	var result strings.Builder
	for _, r := range filename {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			result.WriteRune(r)
		}
	}

	sanitized := result.String()
	if sanitized == "" {
		return "file"
	}

	return sanitized
}

// GetFileContent reads and returns the content of a text file
func (h *UploadHandler) GetFileContent(sessionID, filename string) ([]byte, error) {
	filePath := filepath.Join(h.uploadDir, sessionID, filename)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
