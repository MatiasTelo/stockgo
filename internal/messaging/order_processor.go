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

// HandleOrderStatusChanged maneja el cambio de estado de una orden
func (p *OrderEventProcessor) HandleOrderStatusChanged(ctx context.Context, status *OrderStatusChangedMessage) error {
	log.Printf("Processing order status change: %s -> %s", status.OrderID, status.Status)

	switch status.Status {
	case "CONFIRMED", "PAID", "PROCESSING":
		// Confirmar las reservas (descontar stock)
		return p.confirmOrderReservations(ctx, status)
	case "CANCELLED", "EXPIRED":
		// Cancelar las reservas (liberar stock)
		return p.cancelOrderReservations(ctx, status)
	default:
		log.Printf("Unhandled order status: %s", status.Status)
		return nil
	}
}

func (p *OrderEventProcessor) confirmOrderReservations(ctx context.Context, status *OrderStatusChangedMessage) error {
	log.Printf("Confirming reservations for order: %s", status.OrderID)

	// Si no tenemos los items en el mensaje, necesitaremos obtenerlos de otra fuente
	// Por ahora asumimos que los items están incluidos en el mensaje
	if len(status.Items) == 0 {
		log.Printf("No items provided in status change message for order %s", status.OrderID)
		return nil
	}

	for _, item := range status.Items {
		if err := p.stockService.ConfirmReservation(ctx, status.OrderID, item.ArticleID); err != nil {
			log.Printf("Failed to confirm reservation for article %s in order %s: %v", 
				item.ArticleID, status.OrderID, err)
			return err
		}

		log.Printf("Successfully confirmed reservation for article %s in order %s", 
			item.ArticleID, status.OrderID)
	}

	log.Printf("Successfully confirmed all reservations for order: %s", status.OrderID)
	return nil
}

func (p *OrderEventProcessor) cancelOrderReservations(ctx context.Context, status *OrderStatusChangedMessage) error {
	log.Printf("Canceling reservations for order: %s", status.OrderID)

	// Si no tenemos los items en el mensaje, necesitaremos obtenerlos de otra fuente
	if len(status.Items) == 0 {
		log.Printf("No items provided in status change message for order %s", status.OrderID)
		return nil
	}

	for _, item := range status.Items {
		if err := p.stockService.CancelReservation(ctx, status.OrderID, item.ArticleID); err != nil {
			log.Printf("Failed to cancel reservation for article %s in order %s: %v", 
				item.ArticleID, status.OrderID, err)
			// Continuamos con los otros artículos aunque falle uno
		} else {
			log.Printf("Successfully canceled reservation for article %s in order %s", 
				item.ArticleID, status.OrderID)
		}
	}

	log.Printf("Processed reservation cancellations for order: %s", status.OrderID)
	return nil
}

// compensateReservations cancela las reservas ya hechas en caso de error
func (p *OrderEventProcessor) compensateReservations(ctx context.Context, orderID string, items []OrderItem) {
	log.Printf("Compensating reservations for failed order: %s", orderID)

	for _, item := range items {
		if err := p.stockService.CancelReservation(ctx, orderID, item.ArticleID); err != nil {
			log.Printf("Failed to compensate reservation for article %s: %v", item.ArticleID, err)
		}
	}
}