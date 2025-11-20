package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/rabbitmq/amqp091-go"
)

// OrderConfirmedConsumer maneja los eventos de órdenes confirmadas
type OrderConfirmedConsumer struct {
	stockService *service.StockService
	connection   *amqp091.Connection
	channel      *amqp091.Channel
}

// OrderConfirmedMessage representa el mensaje de orden confirmada
type OrderConfirmedMessage struct {
	OrderID     string              `json:"orderId"`
	CartID      string              `json:"cartId"`
	UserID      string              `json:"userId"`
	Articles    []ArticlePlacedData `json:"articles"`
	ConfirmedAt string              `json:"confirmed_at"`
}

func NewOrderConfirmedConsumer(stockService *service.StockService, conn *amqp091.Connection) (*OrderConfirmedConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	consumer := &OrderConfirmedConsumer{
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

func (c *OrderConfirmedConsumer) setupQueue() error {
	// Declarar exchange fanout
	err := c.channel.ExchangeDeclare(
		"orders_confirmed", // nombre del exchange
		"fanout",           // tipo fanout
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return err
	}

	// Declarar cola
	queue, err := c.channel.QueueDeclare(
		"orders_confirmed_stock", // nombre único para stock service
		true,                     // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		return err
	}

	// Bind cola al exchange (routing key vacío para fanout)
	return c.channel.QueueBind(
		queue.Name,         // queue name
		"",                 // routing key vacío para fanout
		"orders_confirmed", // exchange
		false,
		nil,
	)
}

// StartConsuming inicia el consumo de mensajes
func (c *OrderConfirmedConsumer) StartConsuming(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		"orders_confirmed_stock", // queue
		"",                       // consumer
		false,                    // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
	if err != nil {
		return err
	}

	log.Println("OrderConfirmedConsumer: Waiting for orders_confirmed messages...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("OrderConfirmedConsumer: Context cancelled, stopping consumer")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("OrderConfirmedConsumer: Channel closed")
					return
				}

				if err := c.handleMessage(ctx, msg); err != nil {
					log.Printf("OrderConfirmedConsumer: Error processing message: %v", err)
					msg.Nack(false, true) // requeue on error
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}

func (c *OrderConfirmedConsumer) handleMessage(ctx context.Context, msg amqp091.Delivery) error {
	var orderMsg OrderConfirmedMessage
	if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
		return err
	}

	log.Printf("OrderConfirmedConsumer: Processing order confirmed: %s with %d items", orderMsg.OrderID, len(orderMsg.Articles))

	// Confirmar las reservas (descontar stock)
	for _, item := range orderMsg.Articles {
		if err := c.stockService.ConfirmReservationByOrderID(ctx, orderMsg.OrderID, item.ArticleID, "Order confirmed via RabbitMQ"); err != nil {
			log.Printf("OrderConfirmedConsumer: Failed to confirm reservation for article %s in order %s: %v",
				item.ArticleID, orderMsg.OrderID, err)
			return err
		}

		log.Printf("OrderConfirmedConsumer: Successfully confirmed reservation for article %s in order %s",
			item.ArticleID, orderMsg.OrderID)
	}

	log.Printf("OrderConfirmedConsumer: Successfully confirmed all reservations for order: %s", orderMsg.OrderID)
	return nil
}

// Close cierra las conexiones del consumer
func (c *OrderConfirmedConsumer) Close() error {
	if c.channel != nil {
		return c.channel.Close()
	}
	return nil
}
