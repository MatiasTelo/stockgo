package handlers

import (
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

type ConfirmReservationHandler struct {
	stockService *service.StockService
}

type ConfirmReservationRequest struct {
	ArticleID string `json:"article_id" validate:"required"`
	OrderID   string `json:"order_id" validate:"required"`
	Reason    string `json:"reason,omitempty"`
}

func NewConfirmReservationHandler(stockService *service.StockService) *ConfirmReservationHandler {
	return &ConfirmReservationHandler{
		stockService: stockService,
	}
}

// PUT /api/stock/confirm-reservation
func (h *ConfirmReservationHandler) Handle(c *fiber.Ctx) error {
	var req ConfirmReservationRequest

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

	// Confirmar la reserva usando el servicio
	err := h.stockService.ConfirmReservationByOrderID(c.Context(), req.OrderID, req.ArticleID, req.Reason)
	if err != nil {
		if err.Error() == "no active reservation found for this order and article" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No active reservation found for the specified order_id and article_id",
			})
		}
		if err.Error() == "reservation has already been cancelled" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Cannot confirm a cancelled reservation",
			})
		}
		if err.Error() == "reservation has already been confirmed" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Reservation has already been confirmed",
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
