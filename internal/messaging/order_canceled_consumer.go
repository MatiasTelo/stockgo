package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/rabbitmq/amqp091-go"
)

// OrderCanceledConsumer maneja los eventos de órdenes canceladas
type OrderCanceledConsumer struct {
	stockService *service.StockService
	connection   *amqp091.Connection
	channel      *amqp091.Channel
}

// OrderCanceledMessage representa el mensaje de orden cancelada
type OrderCanceledMessage struct {
	OrderID    string              `json:"orderId"`
	CartID     string              `json:"cartId"`
	UserID     string              `json:"userId"`
	Articles   []ArticlePlacedData `json:"articles"`
	CanceledAt string              `json:"canceled_at"`
	Reason     string              `json:"reason,omitempty"`
}

func NewOrderCanceledConsumer(stockService *service.StockService, conn *amqp091.Connection) (*OrderCanceledConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	consumer := &OrderCanceledConsumer{
		stockService: stockService,
		connection:   conn,
		channel:      ch,
	}

	if err := consumer.setupQueue(); err != nil {
		ch.Close()
		return nil, err
	}

	return consumer, nil
}

func (c *OrderCanceledConsumer) setupQueue() error {
	// Declarar exchange
	err := c.channel.ExchangeDeclare(
		"ecommerce", // name
		"topic",     // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return err
	}

	// Declarar cola
	queue, err := c.channel.QueueDeclare(
		"stock.order.canceled", // name
		true,                   // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return err
	}

	// Bind cola al exchange
	return c.channel.QueueBind(
		queue.Name,       // queue name
		"order.canceled", // routing key
		"ecommerce",      // exchange
		false,
		nil,
	)
}

// StartConsuming inicia el consumo de mensajes
func (c *OrderCanceledConsumer) StartConsuming(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		"stock.order.canceled", // queue
		"",                     // consumer
		false,                  // auto-ack
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		return err
	}

	log.Println("OrderCanceledConsumer: Starting to consume order.canceled messages")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("OrderCanceledConsumer: Context cancelled, stopping consumer")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("OrderCanceledConsumer: Channel closed")
					return
				}

				if err := c.handleMessage(ctx, msg); err != nil {
					log.Printf("OrderCanceledConsumer: Error processing message: %v", err)
					msg.Nack(false, true) // requeue on error
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}

func (c *OrderCanceledConsumer) handleMessage(ctx context.Context, msg amqp091.Delivery) error {
	var orderMsg OrderCanceledMessage
	if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
		return err
	}

	log.Printf("OrderCanceledConsumer: Processing order canceled: %s with %d items", orderMsg.OrderID, len(orderMsg.Articles))

	// Cancelar las reservas (liberar stock)
	reason := "Order canceled via RabbitMQ"
	if orderMsg.Reason != "" {
		reason = orderMsg.Reason
	}

	for _, item := range orderMsg.Articles {
		if err := c.stockService.CancelReservationByOrderID(ctx, orderMsg.OrderID, item.ArticleID, reason); err != nil {
			log.Printf("OrderCanceledConsumer: Failed to cancel reservation for article %s in order %s: %v",
				item.ArticleID, orderMsg.OrderID, err)
			// Continuamos con los otros artículos aunque falle uno
		} else {
			log.Printf("OrderCanceledConsumer: Successfully canceled reservation for article %s in order %s",
				item.ArticleID, orderMsg.OrderID)
		}
	}

	log.Printf("OrderCanceledConsumer: Successfully canceled all reservations for order: %s", orderMsg.OrderID)
	return nil
}

// Close cierra las conexiones del consumer
func (c *OrderCanceledConsumer) Close() error {
	if c.channel != nil {
		return c.channel.Close()
	}
	return nil
}
