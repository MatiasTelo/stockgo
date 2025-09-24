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
		"data":  lowStockInfo,
		"count": len(lowStockInfo),
		"message": "Low stock items retrieved successfully",
	})
}

// GET /api/stock/alerts/summary
func (h *LowStockHandler) GetAlertsSummary(c *fiber.Ctx) error {
	lowStocks, err := h.stockService.GetLowStocks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve low stock summary",
			"details": err.Error(),
		})
	}

	allStocks, err := h.stockService.GetAllStocks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve all stocks for summary",
			"details": err.Error(),
		})
	}

	// Calcular estadísticas
	totalArticles := len(allStocks)
	lowStockArticles := len(lowStocks)
	
	criticalStocks := 0
	for _, stock := range lowStocks {
		if stock.Quantity == 0 {
			criticalStocks++
		}
	}

	var lowStockPercentage float64
	if totalArticles > 0 {
		lowStockPercentage = float64(lowStockArticles) / float64(totalArticles) * 100
	}

	return c.JSON(fiber.Map{
		"summary": fiber.Map{
			"total_articles":        totalArticles,
			"low_stock_articles":    lowStockArticles,
			"critical_stock_articles": criticalStocks, // Articles with 0 quantity
			"low_stock_percentage":  lowStockPercentage,
		},
		"low_stock_items": lowStocks,
		"generated_at":    fiber.Map{
			"timestamp": c.Context().Value("timestamp"),
		},
	})
}