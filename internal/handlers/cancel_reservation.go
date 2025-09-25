package handlers

import (
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type CancelReservationHandler struct {
	stockService *service.StockService
}

type CancelReservationRequest struct {
	ArticleID string `json:"article_id" validate:"required"`
	OrderID   string `json:"order_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"min=1"`
}

func NewCancelReservationHandler(stockService *service.StockService) *CancelReservationHandler {
	return &CancelReservationHandler{
		stockService: stockService,
	}
}

// DELETE /api/stock/reservations
func (h *CancelReservationHandler) Handle(c *fiber.Ctx) error {
	var req CancelReservationRequest

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

	err := h.stockService.CancelReservation(c.Context(), req.OrderID, req.ArticleID, req.Quantity)
	if err != nil {
		if len(err.Error()) > 20 && err.Error()[:20] == "reservation not found:" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Reservation not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to cancel reservation",
			"details": err.Error(),
		})
	}

	// Get updated stock info to return
	stock, _ := h.stockService.GetStock(c.Context(), req.ArticleID)

	return c.JSON(fiber.Map{
		"message": "Reservation cancelled successfully",
		"cancelled_reservation": fiber.Map{
			"article_id": req.ArticleID,
			"order_id":   req.OrderID,
		},
		"stock": stock,
	})
}

// DELETE /api/stock/orders/:orderId/reservations/:articleId
func (h *CancelReservationHandler) HandleByPath(c *fiber.Ctx) error {
	orderID := c.Params("orderId")
	articleID := c.Params("articleId")

	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order_id is required",
		})
	}

	if articleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "article_id is required",
		})
	}

	var req struct {
		Quantity int `json:"quantity" validate:"min=1"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body. Quantity is required",
			"details": err.Error(),
		})
	}

	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "quantity must be greater than 0",
		})
	}

	err := h.stockService.CancelReservation(c.Context(), orderID, articleID, req.Quantity)
	if err != nil {
		if len(err.Error()) > 20 && err.Error()[:20] == "reservation not found:" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Reservation not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to cancel reservation",
			"details": err.Error(),
		})
	}

	// Get updated stock info to return
	stock, _ := h.stockService.GetStock(c.Context(), articleID)

	return c.JSON(fiber.Map{
		"message": "Reservation cancelled successfully",
		"cancelled_reservation": fiber.Map{
			"article_id": articleID,
			"order_id":   orderID,
		},
		"stock": stock,
	})
}

// POST /api/stock/reservations/confirm
func (h *CancelReservationHandler) ConfirmReservation(c *fiber.Ctx) error {
	var req struct {
		ArticleID string `json:"article_id" validate:"required"`
		OrderID   string `json:"order_id" validate:"required"`
		Quantity  int    `json:"quantity" validate:"min=1"`
	}

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

	err := h.stockService.ConfirmReservation(c.Context(), req.OrderID, req.ArticleID, req.Quantity)
	if err != nil {
		if len(err.Error()) > 20 && err.Error()[:20] == "reservation not found:" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Reservation not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to confirm reservation",
			"details": err.Error(),
		})
	}

	// Get updated stock info to return
	stock, _ := h.stockService.GetStock(c.Context(), req.ArticleID)

	return c.JSON(fiber.Map{
		"message": "Reservation confirmed successfully",
		"confirmed_reservation": fiber.Map{
			"article_id": req.ArticleID,
			"order_id":   req.OrderID,
		},
		"stock": stock,
	})
}

// POST /api/stock/orders/:orderId/reservations/:articleId/confirm
func (h *CancelReservationHandler) ConfirmReservationByPath(c *fiber.Ctx) error {
	orderID := c.Params("orderId")
	articleID := c.Params("articleId")

	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order_id is required",
		})
	}

	if articleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "article_id is required",
		})
	}

	var req struct {
		Quantity int `json:"quantity" validate:"min=1"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body. Quantity is required",
			"details": err.Error(),
		})
	}

	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "quantity must be greater than 0",
		})
	}

	err := h.stockService.ConfirmReservation(c.Context(), orderID, articleID, req.Quantity)
	if err != nil {
		if len(err.Error()) > 20 && err.Error()[:20] == "reservation not found:" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Reservation not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to confirm reservation",
			"details": err.Error(),
		})
	}

	// Get updated stock info to return
	stock, _ := h.stockService.GetStock(c.Context(), articleID)

	return c.JSON(fiber.Map{
		"message": "Reservation confirmed successfully",
		"confirmed_reservation": fiber.Map{
			"article_id": articleID,
			"order_id":   orderID,
		},
		"stock": stock,
	})
}
