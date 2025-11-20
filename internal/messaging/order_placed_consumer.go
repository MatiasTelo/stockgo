package messaging

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/MatiasTelo/stockgo/internal/models"
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/rabbitmq/amqp091-go"
)

// OrderPlacedConsumer maneja los eventos de órdenes creadas
type OrderPlacedConsumer struct {
	stockService               *service.StockService
	connection                 *amqp091.Connection
	channel                    *amqp091.Channel
	insufficientStockPublisher *InsufficientStockPublisher
}

// OrderPlacedMessage representa el mensaje de orden creada
type OrderPlacedMessage struct {
	OrderID  string              `json:"orderId"`
	CartID   string              `json:"cartId"`
	UserID   string              `json:"userId"`
	Articles []ArticlePlacedData `json:"articles"`
}

// ArticlePlacedData representa un artículo en la orden
type ArticlePlacedData struct {
	ArticleID string `json:"articleId"`
	Quantity  int    `json:"quantity"`
}

func NewOrderPlacedConsumer(stockService *service.StockService, conn *amqp091.Connection, insufficientStockPublisher *InsufficientStockPublisher) (*OrderPlacedConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	consumer := &OrderPlacedConsumer{
		stockService:               stockService,
		connection:                 conn,
		channel:                    ch,
		insufficientStockPublisher: insufficientStockPublisher,
	}

	if err := consumer.setupQueue(); err != nil {
		ch.Close()
		return nil, err
	}

	return consumer, nil
}

func (c *OrderPlacedConsumer) setupQueue() error {
	// Declarar exchange fanout
	err := c.channel.ExchangeDeclare(
		"orders_placed", // nombre del exchange
		"fanout",        // tipo fanout
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return err
	}

	// Declarar cola
	queue, err := c.channel.QueueDeclare(
		"orders_placed_stock", // nombre único para stock service
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return err
	}

	// Bind cola al exchange (routing key vacío para fanout)
	return c.channel.QueueBind(
		queue.Name,      // queue name
		"",              // routing key vacío para fanout
		"orders_placed", // exchange
		false,
		nil,
	)
}

// StartConsuming inicia el consumo de mensajes
func (c *OrderPlacedConsumer) StartConsuming(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		"orders_placed_stock", // queue
		"",                    // consumer
		false,                 // auto-ack
		false,                 // exclusive
		false,                 // no-local
		false,                 // no-wait
		nil,                   // args
	)
	if err != nil {
		return err
	}

	log.Println("OrderPlacedConsumer: Waiting for orders_placed messages...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("OrderPlacedConsumer: Context cancelled, stopping consumer")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("OrderPlacedConsumer: Channel closed")
					return
				}

				if err := c.handleMessage(ctx, msg); err != nil {
					log.Printf("OrderPlacedConsumer: Error processing message: %v", err)

					// Verificar si es un error recuperable o no recuperable
					if c.isRecoverableError(err) {
						log.Printf("OrderPlacedConsumer: Recoverable error, requeuing message")
						msg.Nack(false, true) // requeue only for recoverable errors
					} else {
						log.Printf("OrderPlacedConsumer: Non-recoverable error, rejecting message: %v", err)
						msg.Nack(false, false) // reject without requeuing
					}
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}

func (c *OrderPlacedConsumer) handleMessage(ctx context.Context, msg amqp091.Delivery) error {
	var orderMsg OrderPlacedMessage
	if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
		return err
	}

	log.Printf("OrderPlacedConsumer: Processing order placed: %s with %d items", orderMsg.OrderID, len(orderMsg.Articles))

	// Array para recopilar artículos con stock insuficiente
	var insufficientStockArticles []string
	// Array para tracking de artículos reservados exitosamente
	var successfullyReserved []string
	// Flag para indicar si hubo algún error
	var hasError error

	// Intentar reservar stock para cada artículo en la orden
	for _, item := range orderMsg.Articles {
		req := &models.ReserveStockRequest{
			ArticleID: item.ArticleID,
			Quantity:  item.Quantity,
			OrderID:   orderMsg.OrderID,
		}

		if err := c.stockService.ReserveStock(ctx, req); err != nil {
			log.Printf("OrderPlacedConsumer: Failed to reserve stock for article %s in order %s: %v",
				item.ArticleID, orderMsg.OrderID, err)

			// Si es error de stock insuficiente, agregarlo a la lista
			if strings.Contains(err.Error(), "insufficient stock") || strings.Contains(err.Error(), "not enough stock") {
				insufficientStockArticles = append(insufficientStockArticles, item.ArticleID)
			}

			// Guardar el error pero continuar verificando los demás artículos
			hasError = err
		} else {
			// Marcar como exitosamente reservado
			successfullyReserved = append(successfullyReserved, item.ArticleID)
			log.Printf("OrderPlacedConsumer: Successfully reserved %d units of article %s for order %s",
				item.Quantity, item.ArticleID, orderMsg.OrderID)
		}
	}

	// Si hubo artículos con stock insuficiente, publicar mensaje y rechazar la orden
	if len(insufficientStockArticles) > 0 {
		log.Printf("OrderPlacedConsumer: Order %s has insufficient stock for %d article(s): %v",
			orderMsg.OrderID, len(insufficientStockArticles), insufficientStockArticles)

		// Publicar mensaje con TODOS los artículos sin stock
		if c.insufficientStockPublisher != nil {
			if err := c.insufficientStockPublisher.PublishInsufficientStock(ctx, orderMsg.OrderID, insufficientStockArticles); err != nil {
				log.Printf("OrderPlacedConsumer: Failed to publish insufficient stock alert: %v", err)
			} else {
				log.Printf("OrderPlacedConsumer: Published insufficient stock alert for order %s with %d articles",
					orderMsg.OrderID, len(insufficientStockArticles))
			}
		}

		// Compensar las reservas que se hicieron exitosamente
		if len(successfullyReserved) > 0 {
			log.Printf("OrderPlacedConsumer: Compensating %d successful reservations", len(successfullyReserved))
			c.compensateReservations(ctx, orderMsg.OrderID, orderMsg.Articles)
		}

		return hasError
	}

	// Si hubo algún error pero no fue de stock insuficiente, compensar y retornar error
	if hasError != nil {
		log.Printf("OrderPlacedConsumer: Error processing order %s, compensating reservations", orderMsg.OrderID)
		c.compensateReservations(ctx, orderMsg.OrderID, orderMsg.Articles)
		return hasError
	}

	log.Printf("OrderPlacedConsumer: Successfully processed order placed: %s", orderMsg.OrderID)
	return nil
}

// compensateReservations cancela las reservas ya hechas en caso de error
func (c *OrderPlacedConsumer) compensateReservations(ctx context.Context, orderID string, items []ArticlePlacedData) {
	log.Printf("OrderPlacedConsumer: Compensating reservations for failed order: %s", orderID)

	for _, item := range items {
		if err := c.stockService.CancelReservationByOrderID(ctx, orderID, item.ArticleID, "Compensation for failed order processing"); err != nil {
			log.Printf("OrderPlacedConsumer: Failed to compensate reservation for article %s: %v", item.ArticleID, err)
		}
	}
}

// isRecoverableError determina si un error es recuperable o no
func (c *OrderPlacedConsumer) isRecoverableError(err error) bool {
	errorMsg := err.Error()

	// Errores no recuperables (no reencolar)
	nonRecoverableErrors := []string{
		"already has an active reservation",
		"duplicate order",
		"invalid order format",
		"article not found",
		"insufficient stock",
	}

	for _, nonRecoverable := range nonRecoverableErrors {
		if strings.Contains(errorMsg, nonRecoverable) {
			return false
		}
	}

	// Por defecto, consideramos otros errores como recuperables
	return true
}

// Close cierra las conexiones del consumer
func (c *OrderPlacedConsumer) Close() error {
	if c.channel != nil {
		return c.channel.Close()
	}
	return nil
}
