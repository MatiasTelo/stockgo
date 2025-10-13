package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/MatiasTelo/stockgo/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQService provides a basic connection to RabbitMQ
type RabbitMQService struct {
	conn   *amqp091.Connection
	config *config.RabbitMQConfig
}

// MessagePublisher interface for publishing messages
type MessagePublisher interface {
	PublishLowStockAlert(ctx context.Context, articleID string, currentQuantity, minStock int) error
}

func NewRabbitMQService(cfg *config.RabbitMQConfig) (*RabbitMQService, error) {
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	service := &RabbitMQService{
		conn:   conn,
		config: cfg,
	}

	log.Println("Connected to RabbitMQ successfully")
	return service, nil
}

// GetConnection returns the RabbitMQ connection
func (r *RabbitMQService) GetConnection() *amqp091.Connection {
	return r.conn
}

// Close closes the RabbitMQ connection
func (r *RabbitMQService) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
