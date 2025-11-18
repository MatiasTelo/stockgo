package handlers

import (
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type LowStockHandler struct {
	stockService *service.StockService
}

func NewLowStockHandler(stockService *service.StockService) *LowStockHandler {
	return &LowStockHandler{
		stockService: stockService,
	}
}

// GET /api/stock/low-stock
func (h *LowStockHandler) Handle(c *fiber.Ctx) error {
	stocks, err := h.stockService.GetLowStocks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve low stock items",
			"details": err.Error(),
		})
	}

	// Enriquecer la respuesta con información adicional
	var lowStockInfo []fiber.Map
	for _, stock := range stocks {
		lowStockInfo = append(lowStockInfo, fiber.Map{
			"article_id":        stock.ArticleID,
			"current_quantity":  stock.Quantity,
			"reserved":          stock.Reserved,
			"available":         stock.AvailableQuantity(),
			"min_stock":         stock.MinStock,
			"max_stock":         stock.MaxStock,
			"location":          stock.Location,
			"deficit":           stock.MinStock - stock.Quantity,
			"percentage_of_min": float64(stock.Quantity) / float64(stock.MinStock) * 100,
			"updated_at":        stock.UpdatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"data":    lowStockInfo,
		"count":   len(lowStockInfo),
		"message": "Low stock items retrieved successfully",
	})
}
