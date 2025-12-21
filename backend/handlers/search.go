package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ronan/home-agent/models"
)

// SearchHandler handles search-related API endpoints
type SearchHandler struct {
	db *models.DB
}

// NewSearchHandler creates a new SearchHandler
func NewSearchHandler(db *models.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

// SearchResponse represents the search API response
type SearchResponse struct {
	Results []*models.SearchResult `json:"results"`
	Total   int                    `json:"total"`
	Query   string                 `json:"query"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
}

// RegisterRoutes registers search API routes
func (h *SearchHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/search", h.Search)
}

// Search handles GET /api/search?q=term&limit=20&offset=0
func (h *SearchHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q", "")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	// Parse pagination parameters
	limit := c.QueryInt("limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	results, total, err := h.db.SearchMessages(query, limit, offset)
	if err != nil {
		log.Printf("Search failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Search failed",
		})
	}

	// Return empty array if no results
	if results == nil {
		results = []*models.SearchResult{}
	}

	return c.JSON(SearchResponse{
		Results: results,
		Total:   total,
		Query:   query,
		Limit:   limit,
		Offset:  offset,
	})
}
