package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

// InsufficientStockPublisher maneja la publicación de alertas de stock insuficiente
type InsufficientStockPublisher struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
}

// InsufficientStockAlert representa el mensaje de stock insuficiente
type InsufficientStockAlert struct {
	OrderID    string   `json:"order_id"`
	ArticleIDs []string `json:"article_ids"`
}

func NewInsufficientStockPublisher(conn *amqp091.Connection) (*InsufficientStockPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declarar exchange
	err = ch.ExchangeDeclare(
		"ecommerce", // name
		"topic",     // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		ch.Close()
		return nil, err
	}

	return &InsufficientStockPublisher{
		connection: conn,
		channel:    ch,
	}, nil
}

// PublishInsufficientStock publica un mensaje de stock insuficiente
func (p *InsufficientStockPublisher) PublishInsufficientStock(ctx context.Context, orderID string, articleIDs []string) error {
	alert := InsufficientStockAlert{
		OrderID:    orderID,
		ArticleIDs: articleIDs,
	}

	body, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	err = p.channel.PublishWithContext(
		ctx,
		"ecommerce",          // exchange
		"insufficient_stock", // routing key
		false,                // mandatory
		false,                // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
		},
	)

	if err != nil {
		log.Printf("InsufficientStockPublisher: Failed to publish alert for order %s: %v", orderID, err)
		return err
	}

	log.Printf("InsufficientStockPublisher: Published insufficient stock alert for order %s with %d articles", orderID, len(articleIDs))
	return nil
}

// Close cierra el canal del publisher
func (p *InsufficientStockPublisher) Close() error {
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}
