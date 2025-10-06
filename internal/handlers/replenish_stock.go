package handlers

import (
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type ReplenishStockHandler struct {
	stockService *service.StockService
}

type ReplenishStockRequest struct {
	ArticleID string `json:"article_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"min=1"`
	Reason    string `json:"reason"`
}

func NewReplenishStockHandler(stockService *service.StockService) *ReplenishStockHandler {
	return &ReplenishStockHandler{
		stockService: stockService,
	}
}

// PUT /api/stock/replenish
func (h *ReplenishStockHandler) Handle(c *fiber.Ctx) error {
	var req ReplenishStockRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validaciones
	if req.ArticleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "article_id is required",
		})
	}

	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "quantity must be greater than 0",
		})
	}

	// Set default reason if not provided
	if req.Reason == "" {
		req.Reason = "Stock replenishment"
	}

	stock, err := h.stockService.ReplenishStock(c.Context(), req.ArticleID, req.Quantity, req.Reason)
	if err != nil {
		if err.Error() == "article not found: stock not found for article_id: "+req.ArticleID {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Article not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to replenish stock",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stock replenished successfully",
		"data":    stock,
		"replenished": fiber.Map{
			"article_id": req.ArticleID,
			"quantity":   req.Quantity,
			"reason":     req.Reason,
		},
	})
}
