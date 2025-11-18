package handlers

import (
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type GetArticleHandler struct {
	stockService *service.StockService
}

func NewGetArticleHandler(stockService *service.StockService) *GetArticleHandler {
	return &GetArticleHandler{
		stockService: stockService,
	}
}

// GET /api/stock/articles/:articleId
// Requiere autenticación mediante token Bearer
func (h *GetArticleHandler) Handle(c *fiber.Ctx) error {
	// El token ya fue validado por el middleware AuthMiddleware
	// y está disponible en c.Locals("token")

	articleID := c.Params("articleId")
	if articleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "article_id is required",
		})
	}

	stock, err := h.stockService.GetStock(c.Context(), articleID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Article not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": stock,
	})
}
