package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MatiasTelo/stockgo/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockEventRepository struct {
	db *pgxpool.Pool
}

func NewStockEventRepository(db *pgxpool.Pool) *StockEventRepository {
	return &StockEventRepository{
		db: db,
	}
}

// CreateStockEvent crea un nuevo evento de stock
func (r *StockEventRepository) CreateStockEvent(ctx context.Context, event *models.StockEvent) error {
	query := `
		INSERT INTO stock_events (id, article_id, event_type, quantity, order_id, reason, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	event.ID = uuid.New()
	event.CreatedAt = time.Now()

	// Asegurar que metadata sea un JSON válido
	metadata := event.Metadata
	if metadata == "" {
		metadata = "{}"
	}

	_, err := r.db.Exec(ctx, query,
		event.ID, event.ArticleID, event.EventType, event.Quantity,
		event.OrderID, event.Reason, metadata, event.CreatedAt)

	if err != nil {
		return fmt.Errorf("error creating stock event: %w", err)
	}

	return nil
}

// GetStockEventsByArticleID obtiene eventos por ID del artículo
func (r *StockEventRepository) GetStockEventsByArticleID(ctx context.Context, articleID string, limit int) ([]*models.StockEvent, error) {
	query := `
		SELECT id, article_id, event_type, quantity, order_id, reason, metadata, created_at
		FROM stock_events
		WHERE article_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	
	rows, err := r.db.Query(ctx, query, articleID, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying stock events: %w", err)
	}
	defer rows.Close()

	var events []*models.StockEvent
	for rows.Next() {
		var event models.StockEvent
		err := rows.Scan(
			&event.ID, &event.ArticleID, &event.EventType, &event.Quantity,
			&event.OrderID, &event.Reason, &event.Metadata, &event.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning stock event: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}

// GetStockEventsByOrderID obtiene eventos por ID de orden
func (r *StockEventRepository) GetStockEventsByOrderID(ctx context.Context, orderID string) ([]*models.StockEvent, error) {
	query := `
		SELECT id, article_id, event_type, quantity, order_id, reason, metadata, created_at
		FROM stock_events
		WHERE order_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("error querying stock events by order: %w", err)
	}
	defer rows.Close()

	var events []*models.StockEvent
	for rows.Next() {
		var event models.StockEvent
		err := rows.Scan(
			&event.ID, &event.ArticleID, &event.EventType, &event.Quantity,
			&event.OrderID, &event.Reason, &event.Metadata, &event.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning stock event: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}

// GetAllStockEvents obtiene todos los eventos con paginación
func (r *StockEventRepository) GetAllStockEvents(ctx context.Context, offset, limit int) ([]*models.StockEvent, error) {
	query := `
		SELECT id, article_id, event_type, quantity, order_id, reason, metadata, created_at
		FROM stock_events
		ORDER BY created_at DESC
		OFFSET $1 LIMIT $2
	`
	
	rows, err := r.db.Query(ctx, query, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying stock events: %w", err)
	}
	defer rows.Close()

	var events []*models.StockEvent
	for rows.Next() {
		var event models.StockEvent
		err := rows.Scan(
			&event.ID, &event.ArticleID, &event.EventType, &event.Quantity,
			&event.OrderID, &event.Reason, &event.Metadata, &event.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning stock event: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}