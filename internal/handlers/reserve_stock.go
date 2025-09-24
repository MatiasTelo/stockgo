package handlers

import (
	"github.com/MatiasTelo/stockgo/internal/models"
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type ReserveStockHandler struct {
	stockService *service.StockService
}

func NewReserveStockHandler(stockService *service.StockService) *ReserveStockHandler {
	return &ReserveStockHandler{
		stockService: stockService,
	}
}

// POST /api/stock/reserve
func (h *ReserveStockHandler) Handle(c *fiber.Ctx) error {
	var req models.ReserveStockRequest

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

	if req.OrderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order_id is required",
		})
	}

	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "quantity must be greater than 0",
		})
	}

	err := h.stockService.ReserveStock(c.Context(), &req)
	if err != nil {
		if err.Error() == "article not found: stock not found for article_id: "+req.ArticleID {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Article not found",
			})
		}

		// Check if it's an insufficient stock error
		if len(err.Error()) > 22 && err.Error()[:22] == "error reserving stock:" {
			if len(err.Error()) > 45 && err.Error()[23:42] == "insufficient stock:" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.Error()[23:], // Remove "error reserving stock: " prefix
				})
			}
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to reserve stock",
			"details": err.Error(),
		})
	}

	// Get updated stock info to return
	stock, _ := h.stockService.GetStock(c.Context(), req.ArticleID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Stock reserved successfully",
		"reservation": fiber.Map{
			"article_id": req.ArticleID,
			"order_id":   req.OrderID,
			"quantity":   req.Quantity,
		},
		"stock": stock,
	})
}

// POST /api/stock/articles/:articleId/reserve
func (h *ReserveStockHandler) HandleByPath(c *fiber.Ctx) error {
	articleID := c.Params("articleId")
	if articleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "article_id is required",
		})
	}

	var req struct {
		OrderID  string `json:"order_id" validate:"required"`
		Quantity int    `json:"quantity" validate:"min=1"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if req.OrderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order_id is required",
		})
	}

	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "quantity must be greater than 0",
		})
	}

	reserveReq := &models.ReserveStockRequest{
		ArticleID: articleID,
		OrderID:   req.OrderID,
		Quantity:  req.Quantity,
	}

	err := h.stockService.ReserveStock(c.Context(), reserveReq)
	if err != nil {
		if err.Error() == "article not found: stock not found for article_id: "+articleID {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Article not found",
			})
		}

		// Check if it's an insufficient stock error
		if len(err.Error()) > 22 && err.Error()[:22] == "error reserving stock:" {
			if len(err.Error()) > 45 && err.Error()[23:42] == "insufficient stock:" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.Error()[23:], // Remove "error reserving stock: " prefix
				})
			}
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to reserve stock",
			"details": err.Error(),
		})
	}

	// Get updated stock info to return
	stock, _ := h.stockService.GetStock(c.Context(), articleID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Stock reserved successfully",
		"reservation": fiber.Map{
			"article_id": articleID,
			"order_id":   req.OrderID,
			"quantity":   req.Quantity,
		},
		"stock": stock,
	})
}