package handlers

import (
	"github.com/MatiasTelo/stockgo/internal/models"
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AddArticleHandler struct {
	stockService *service.StockService
}

func NewAddArticleHandler(stockService *service.StockService) *AddArticleHandler {
	return &AddArticleHandler{
		stockService: stockService,
	}
}

// POST /api/stock/articles
func (h *AddArticleHandler) Handle(c *fiber.Ctx) error {
	var req models.CreateStockRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validaciones básicas
	if req.ArticleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "article_id is required",
		})
	}

	if req.Quantity < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "quantity cannot be negative",
		})
	}

	if req.MinStock < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "min_stock cannot be negative",
		})
	}

	if req.MaxStock > 0 && req.MaxStock < req.MinStock {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "max_stock cannot be less than min_stock",
		})
	}

	stock, err := h.stockService.CreateStock(c.Context(), &req)
	if err != nil {
		if err.Error() == "article with ID "+req.ArticleID+" already exists" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to create stock",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Article added successfully",
		"data":    stock,
	})
}
