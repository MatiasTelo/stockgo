package handlers

import (
	"strconv"

	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type GetArticleEventsHandler struct {
	stockService *service.StockService
}

func NewGetArticleEventsHandler(stockService *service.StockService) *GetArticleEventsHandler {
	return &GetArticleEventsHandler{
		stockService: stockService,
	}
}

// GET /api/stock/articles/:articleId/events
// Requiere autenticación mediante token Bearer
func (h *GetArticleEventsHandler) Handle(c *fiber.Ctx) error {
	// El token ya fue validado por el middleware AuthMiddleware
	// y está disponible en c.Locals("token")

	articleID := c.Params("articleId")
	if articleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "article_id is required",
		})
	}

	// Parse limit parameter
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	events, err := h.stockService.GetStockEvents(c.Context(), articleID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve stock events",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  events,
		"count": len(events),
	})
}
