package handlers

import (
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type DeductStockHandler struct {
	stockService *service.StockService
}

type DeductStockRequest struct {
	ArticleID string `json:"article_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"min=1"`
	Reason    string `json:"reason"`
}

func NewDeductStockHandler(stockService *service.StockService) *DeductStockHandler {
	return &DeductStockHandler{
		stockService: stockService,
	}
}

// PUT /api/stock/deduct
func (h *DeductStockHandler) Handle(c *fiber.Ctx) error {
	var req DeductStockRequest

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
		req.Reason = "Manual stock deduction"
	}

	stock, err := h.stockService.DeductStock(c.Context(), req.ArticleID, req.Quantity, req.Reason)
	if err != nil {
		if err.Error() == "article not found: stock not found for article_id: "+req.ArticleID {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Article not found",
			})
		}

		// Check if it's an insufficient stock error
		if len(err.Error()) > 19 && err.Error()[:19] == "insufficient stock:" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to deduct stock",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stock deducted successfully",
		"data":    stock,
		"deducted": fiber.Map{
			"article_id": req.ArticleID,
			"quantity":   req.Quantity,
			"reason":     req.Reason,
		},
	})
}
