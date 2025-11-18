package handlers

import (
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type GetAllArticlesHandler struct {
	stockService *service.StockService
}

func NewGetAllArticlesHandler(stockService *service.StockService) *GetAllArticlesHandler {
	return &GetAllArticlesHandler{
		stockService: stockService,
	}
}

// GET /api/stock/articles
// Requiere autenticación mediante token Bearer
func (h *GetAllArticlesHandler) Handle(c *fiber.Ctx) error {
	// El token ya fue validado por el middleware AuthMiddleware
	// y está disponible en c.Locals("token")

	stocks, err := h.stockService.GetAllStocks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve stocks",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  stocks,
		"count": len(stocks),
	})
}
