package messaging

import (
	"context"
	"log"

	"github.com/MatiasTelo/stockgo/internal/models"
	"github.com/MatiasTelo/stockgo/internal/service"
)

// OrderEventProcessor maneja los eventos de órdenes
type OrderEventProcessor struct {
	stockService *service.StockService
}

func NewOrderEventProcessor(stockService *service.StockService) *OrderEventProcessor {
	return &OrderEventProcessor{
		stockService: stockService,
	}
}

// HandleOrderCreated maneja el evento de orden creada
func (p *OrderEventProcessor) HandleOrderCreated(ctx context.Context, order *OrderCreatedMessage) error {
	log.Printf("Processing order created: %s with %d items", order.OrderID, len(order.Items))

	// Reservar stock para cada artículo en la orden
	for _, item := range order.Items {
		req := &models.ReserveStockRequest{
			ArticleID: item.ArticleID,
			Quantity:  item.Quantity,
			OrderID:   order.OrderID,
		}

		if err := p.stockService.ReserveStock(ctx, req); err != nil {
			log.Printf("Failed to reserve stock for article %s in order %s: %v",
				item.ArticleID, order.OrderID, err)

			// Si falla la reserva, podrías implementar compensación aquí
			// Por ejemplo, cancelar reservas ya hechas para esta orden
			p.compensateReservations(ctx, order.OrderID, order.Items)
			return err
		}

		log.Printf("Successfully reserved %d units of article %s for order %s",
			item.Quantity, item.ArticleID, order.OrderID)
	}

	log.Printf("Successfully processed order created: %s", order.OrderID)
	return nil
}

// HandleOrderConfirmed maneja el evento de orden confirmada
func (p *OrderEventProcessor) HandleOrderConfirmed(ctx context.Context, order *OrderConfirmedMessage) error {
	log.Printf("Processing order confirmed: %s with %d items", order.OrderID, len(order.Items))

	// Confirmar las reservas (descontar stock)
	for _, item := range order.Items {
		if err := p.stockService.ConfirmReservation(ctx, order.OrderID, item.ArticleID, item.Quantity); err != nil {
			log.Printf("Failed to confirm reservation for article %s in order %s: %v",
				item.ArticleID, order.OrderID, err)
			return err
		}

		log.Printf("Successfully confirmed reservation for article %s in order %s",
			item.ArticleID, order.OrderID)
	}

	log.Printf("Successfully confirmed all reservations for order: %s", order.OrderID)
	return nil
}

// HandleOrderCancelled maneja el evento de orden cancelada
func (p *OrderEventProcessor) HandleOrderCancelled(ctx context.Context, order *OrderCancelledMessage) error {
	log.Printf("Processing order cancelled: %s with %d items", order.OrderID, len(order.Items))

	// Cancelar las reservas (liberar stock)
	for _, item := range order.Items {
		if err := p.stockService.CancelReservation(ctx, order.OrderID, item.ArticleID, item.Quantity); err != nil {
			log.Printf("Failed to cancel reservation for article %s in order %s: %v",
				item.ArticleID, order.OrderID, err)
			// Continuamos con los otros artículos aunque falle uno
		} else {
			log.Printf("Successfully canceled reservation for article %s in order %s",
				item.ArticleID, order.OrderID)
		}
	}

	log.Printf("Successfully canceled all reservations for order: %s", order.OrderID)
	return nil
}

// compensateReservations cancela las reservas ya hechas en caso de error
func (p *OrderEventProcessor) compensateReservations(ctx context.Context, orderID string, items []OrderItem) {
	log.Printf("Compensating reservations for failed order: %s", orderID)

	for _, item := range items {
		if err := p.stockService.CancelReservation(ctx, orderID, item.ArticleID, item.Quantity); err != nil {
			log.Printf("Failed to compensate reservation for article %s: %v", item.ArticleID, err)
		}
	}
}
