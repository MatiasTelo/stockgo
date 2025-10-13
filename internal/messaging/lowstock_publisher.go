package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// LowStockPublisher maneja la publicación de alertas de stock bajo
type LowStockPublisher struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
}

// LowStockAlert representa el mensaje de alerta de stock bajo
type LowStockAlert struct {
	ArticleID       string    `json:"article_id"`
	CurrentQuantity int       `json:"current_quantity"`
	MinQuantity     int       `json:"min_quantity"`
	AlertedAt       time.Time `json:"alerted_at"`
	Location        string    `json:"location,omitempty"`
}

func NewLowStockPublisher(conn *amqp091.Connection) (*LowStockPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	publisher := &LowStockPublisher{
		connection: conn,
		channel:    ch,
	}

	if err := publisher.setupExchange(); err != nil {
		ch.Close()
		return nil, err
	}

	return publisher, nil
}

func (p *LowStockPublisher) setupExchange() error {
	// Declarar exchange principal
	err := p.channel.ExchangeDeclare(
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

	// Declarar cola para alertas de stock bajo
	queue, err := p.channel.QueueDeclare(
		"stock.alerts.lowstock", // name
		true,                    // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		return err
	}

	// Bind cola al exchange
	return p.channel.QueueBind(
		queue.Name,        // queue name
		"stock.alert.low", // routing key
		"ecommerce",       // exchange
		false,
		nil,
	)
}

// PublishLowStockAlert publica una alerta de stock bajo
func (p *LowStockPublisher) PublishLowStockAlert(ctx context.Context, articleID string, currentQuantity, minStock int) error {
	alert := LowStockAlert{
		ArticleID:       articleID,
		CurrentQuantity: currentQuantity,
		MinQuantity:     minStock,
		AlertedAt:       time.Now(),
	}

	body, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	err = p.channel.PublishWithContext(
		ctx,
		"ecommerce",       // exchange
		"stock.alert.low", // routing key
		false,             // mandatory
		false,             // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent, // persistent message
			Timestamp:    time.Now(),
			MessageId:    articleID + "-" + time.Now().Format("20060102-150405"),
			Body:         body,
		},
	)

	if err != nil {
		log.Printf("LowStockPublisher: Failed to publish low stock alert for article %s: %v", articleID, err)
		return err
	}

	log.Printf("LowStockPublisher: Published low stock alert for article %s (current: %d, min: %d)",
		articleID, currentQuantity, minStock)

	return nil
}

// PublishLowStockAlertWithLocation publica una alerta de stock bajo con ubicación
func (p *LowStockPublisher) PublishLowStockAlertWithLocation(ctx context.Context, articleID string, currentQuantity, minStock int, location string) error {
	alert := LowStockAlert{
		ArticleID:       articleID,
		CurrentQuantity: currentQuantity,
		MinQuantity:     minStock,
		AlertedAt:       time.Now(),
		Location:        location,
	}

	body, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	err = p.channel.PublishWithContext(
		ctx,
		"ecommerce",       // exchange
		"stock.alert.low", // routing key
		false,             // mandatory
		false,             // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
			MessageId:    articleID + "-" + time.Now().Format("20060102-150405"),
			Body:         body,
		},
	)

	if err != nil {
		log.Printf("LowStockPublisher: Failed to publish low stock alert for article %s: %v", articleID, err)
		return err
	}

	log.Printf("LowStockPublisher: Published low stock alert for article %s at location %s (current: %d, min: %d)",
		articleID, location, currentQuantity, minStock)

	return nil
}

// Close cierra las conexiones del publisher
func (p *LowStockPublisher) Close() error {
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}
