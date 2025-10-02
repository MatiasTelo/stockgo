package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/MatiasTelo/stockgo/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQService struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  *config.RabbitMQConfig
}

type OrderCreatedMessage struct {
	OrderID   string      `json:"order_id"`
	Items     []OrderItem `json:"items"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type OrderItem struct {
	ArticleID string  `json:"article_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderStatusChangedMessage struct {
	OrderID   string      `json:"order_id"`
	Items     []OrderItem `json:"items,omitempty"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type OrderConfirmedMessage struct {
	OrderID   string      `json:"order_id"`
	Items     []OrderItem `json:"items"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type OrderCancelledMessage struct {
	OrderID   string      `json:"order_id"`
	Items     []OrderItem `json:"items"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type LowStockAlertMessage struct {
	ArticleID       string    `json:"article_id"`
	CurrentQuantity int       `json:"current_quantity"`
	MinStock        int       `json:"min_stock"`
	AlertedAt       time.Time `json:"alerted_at"`
}

func NewRabbitMQService(cfg *config.RabbitMQConfig) (*RabbitMQService, error) {
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	service := &RabbitMQService{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}

	// Declarar exchange y colas
	if err := service.setupExchangesAndQueues(); err != nil {
		service.Close()
		return nil, fmt.Errorf("failed to setup exchanges and queues: %w", err)
	}

	log.Println("Connected to RabbitMQ successfully")
	return service, nil
}

func (r *RabbitMQService) setupExchangesAndQueues() error {
	// Declarar exchange principal
	err := r.channel.ExchangeDeclare(
		r.config.Exchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Cola para escuchar eventos de órdenes
	orderQueue := "stock.orders"
	_, err = r.channel.QueueDeclare(
		orderQueue, // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare order queue: %w", err)
	}

	// Bind para órdenes creadas
	err = r.channel.QueueBind(
		orderQueue,        // queue name
		"order.created",   // routing key
		r.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind order created queue: %w", err)
	}

	// Bind para órdenes confirmadas
	err = r.channel.QueueBind(
		orderQueue,        // queue name
		"order.confirmed", // routing key
		r.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind order confirmed queue: %w", err)
	}

	// Bind para órdenes canceladas
	err = r.channel.QueueBind(
		orderQueue,        // queue name
		"order.cancelled", // routing key
		r.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind order cancelled queue: %w", err)
	}

	// Cola para alertas de stock bajo
	alertQueue := "stock.alerts"
	_, err = r.channel.QueueDeclare(
		alertQueue, // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare alert queue: %w", err)
	}

	return nil
}

// PublishLowStockAlert publica una alerta de stock bajo
func (r *RabbitMQService) PublishLowStockAlert(ctx context.Context, articleID string, currentQuantity, minStock int) error {
	message := LowStockAlertMessage{
		ArticleID:       articleID,
		CurrentQuantity: currentQuantity,
		MinStock:        minStock,
		AlertedAt:       time.Now(),
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal low stock alert: %w", err)
	}

	err = r.channel.PublishWithContext(
		ctx,
		r.config.Exchange, // exchange
		"stock.alert.low", // routing key
		false,             // mandatory
		false,             // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish low stock alert: %w", err)
	}

	log.Printf("Published low stock alert for article %s (current: %d, min: %d)",
		articleID, currentQuantity, minStock)
	return nil
}

// ConsumeOrderEvents consume eventos de órdenes
func (r *RabbitMQService) ConsumeOrderEvents(ctx context.Context, handler OrderEventHandler) error {
	msgs, err := r.channel.Consume(
		"stock.orders", // queue
		"",             // consumer
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("Starting to consume order events...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping order event consumption")
				return
			case msg := <-msgs:
				r.handleOrderMessage(msg, handler)
			}
		}
	}()

	return nil
}

func (r *RabbitMQService) handleOrderMessage(msg amqp091.Delivery, handler OrderEventHandler) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic handling message: %v", r)
			msg.Nack(false, false) // No requeue on panic
		}
	}()

	log.Printf("Received message with routing key: %s", msg.RoutingKey)

	switch msg.RoutingKey {
	case "order.created":
		var orderMsg OrderCreatedMessage
		if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
			log.Printf("Failed to unmarshal order created message: %v", err)
			msg.Nack(false, false)
			return
		}

		if err := handler.HandleOrderCreated(context.Background(), &orderMsg); err != nil {
			log.Printf("Failed to handle order created: %v", err)
			msg.Nack(false, true) // Requeue for retry
			return
		}

	case "order.confirmed":
		var confirmedMsg OrderConfirmedMessage
		if err := json.Unmarshal(msg.Body, &confirmedMsg); err != nil {
			log.Printf("Failed to unmarshal order confirmed message: %v", err)
			msg.Nack(false, false)
			return
		}

		if err := handler.HandleOrderConfirmed(context.Background(), &confirmedMsg); err != nil {
			log.Printf("Failed to handle order confirmed: %v", err)
			msg.Nack(false, true) // Requeue for retry
			return
		}

	case "order.cancelled":
		var cancelledMsg OrderCancelledMessage
		if err := json.Unmarshal(msg.Body, &cancelledMsg); err != nil {
			log.Printf("Failed to unmarshal order cancelled message: %v", err)
			msg.Nack(false, false)
			return
		}

		if err := handler.HandleOrderCancelled(context.Background(), &cancelledMsg); err != nil {
			log.Printf("Failed to handle order cancelled: %v", err)
			msg.Nack(false, true) // Requeue for retry
			return
		}

	default:
		log.Printf("Unknown routing key: %s", msg.RoutingKey)
		msg.Nack(false, false)
		return
	}

	msg.Ack(false)
}

// OrderEventHandler define la interfaz para manejar eventos de órdenes
type OrderEventHandler interface {
	HandleOrderCreated(ctx context.Context, order *OrderCreatedMessage) error
	HandleOrderConfirmed(ctx context.Context, order *OrderConfirmedMessage) error
	HandleOrderCancelled(ctx context.Context, order *OrderCancelledMessage) error
}

// Close cierra la conexión
func (r *RabbitMQService) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
