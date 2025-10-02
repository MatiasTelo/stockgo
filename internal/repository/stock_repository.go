package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatiasTelo/stockgo/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type StockRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewStockRepository(db *pgxpool.Pool, redis *redis.Client) *StockRepository {
	return &StockRepository{
		db:    db,
		redis: redis,
	}
}

// CreateStock crea un nuevo registro de stock
func (r *StockRepository) CreateStock(ctx context.Context, stock *models.Stock) error {
	query := `
		INSERT INTO stocks (id, article_id, quantity, reserved, min_stock, max_stock, location, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	stock.ID = uuid.New()
	stock.CreatedAt = time.Now()
	stock.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		stock.ID, stock.ArticleID, stock.Quantity, stock.Reserved,
		stock.MinStock, stock.MaxStock, stock.Location,
		stock.CreatedAt, stock.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating stock: %w", err)
	}

	// Invalidar cache
	r.invalidateStockCache(ctx, stock.ArticleID)

	return nil
}

// GetStockByArticleID obtiene el stock por ID del artículo
func (r *StockRepository) GetStockByArticleID(ctx context.Context, articleID string) (*models.Stock, error) {
	// Intentar obtener desde cache
	if r.redis != nil {
		cacheKey := fmt.Sprintf("stock:%s", articleID)
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var stock models.Stock
			if err := json.Unmarshal([]byte(cached), &stock); err == nil {
				return &stock, nil
			}
		}
	}

	// Si no está en cache, obtener de la base de datos
	query := `
		SELECT id, article_id, quantity, reserved, min_stock, max_stock, location, created_at, updated_at
		FROM stocks
		WHERE article_id = $1
	`
	
	var stock models.Stock
	err := r.db.QueryRow(ctx, query, articleID).Scan(
		&stock.ID, &stock.ArticleID, &stock.Quantity, &stock.Reserved,
		&stock.MinStock, &stock.MaxStock, &stock.Location,
		&stock.CreatedAt, &stock.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("stock not found for article_id: %s", articleID)
		}
		return nil, fmt.Errorf("error getting stock: %w", err)
	}

	// Guardar en cache
	if r.redis != nil {
		r.cacheStock(ctx, &stock)
	}

	return &stock, nil
}

// UpdateStockQuantity actualiza la cantidad de stock
func (r *StockRepository) UpdateStockQuantity(ctx context.Context, articleID string, quantity int) error {
	query := `
		UPDATE stocks 
		SET quantity = $1, updated_at = $2
		WHERE article_id = $3
	`
	
	result, err := r.db.Exec(ctx, query, quantity, time.Now(), articleID)
	if err != nil {
		return fmt.Errorf("error updating stock quantity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("stock not found for article_id: %s", articleID)
	}

	// Invalidar cache
	r.invalidateStockCache(ctx, articleID)

	return nil
}

// ReserveStock reserva una cantidad de stock
func (r *StockRepository) ReserveStock(ctx context.Context, articleID string, quantity int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verificar si hay suficiente stock disponible
	var currentQuantity, reserved int
	err = tx.QueryRow(ctx, 
		"SELECT quantity, reserved FROM stocks WHERE article_id = $1 FOR UPDATE", 
		articleID).Scan(&currentQuantity, &reserved)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("stock not found for article_id: %s", articleID)
		}
		return fmt.Errorf("error checking stock: %w", err)
	}

	availableQuantity := currentQuantity - reserved
	if availableQuantity < quantity {
		return fmt.Errorf("insufficient stock: available %d, requested %d", availableQuantity, quantity)
	}

	// Actualizar stock reservado
	_, err = tx.Exec(ctx,
		"UPDATE stocks SET reserved = reserved + $1, updated_at = $2 WHERE article_id = $3",
		quantity, time.Now(), articleID)
	
	if err != nil {
		return fmt.Errorf("error reserving stock: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	// Invalidar cache
	r.invalidateStockCache(ctx, articleID)

	return nil
}

// CancelReservation cancela una reserva de stock
func (r *StockRepository) CancelReservation(ctx context.Context, articleID string, quantity int) error {
	query := `
		UPDATE stocks 
		SET reserved = reserved - $1, updated_at = $2
		WHERE article_id = $3 AND reserved >= $1
	`
	
	result, err := r.db.Exec(ctx, query, quantity, time.Now(), articleID)
	if err != nil {
		return fmt.Errorf("error canceling reservation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("insufficient reserved stock or stock not found for article_id: %s", articleID)
	}

	// Invalidar cache
	r.invalidateStockCache(ctx, articleID)

	return nil
}

// ConfirmReservation confirma una reserva y descuenta el stock
func (r *StockRepository) ConfirmReservation(ctx context.Context, articleID string, quantity int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verificar que hay suficiente stock reservado
	var reserved int
	err = tx.QueryRow(ctx, 
		"SELECT reserved FROM stocks WHERE article_id = $1 FOR UPDATE", 
		articleID).Scan(&reserved)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("stock not found for article_id: %s", articleID)
		}
		return fmt.Errorf("error checking reserved stock: %w", err)
	}

	if reserved < quantity {
		return fmt.Errorf("insufficient reserved stock: reserved %d, requested %d", reserved, quantity)
	}

	// Descontar del stock y liberar la reserva
	_, err = tx.Exec(ctx,
		"UPDATE stocks SET quantity = quantity - $1, reserved = reserved - $1, updated_at = $2 WHERE article_id = $3",
		quantity, time.Now(), articleID)
	
	if err != nil {
		return fmt.Errorf("error confirming reservation: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	// Invalidar cache
	r.invalidateStockCache(ctx, articleID)

	return nil
}

// GetAllStocks obtiene todos los stocks
func (r *StockRepository) GetAllStocks(ctx context.Context) ([]*models.Stock, error) {
	query := `
		SELECT id, article_id, quantity, reserved, min_stock, max_stock, location, created_at, updated_at
		FROM stocks
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying stocks: %w", err)
	}
	defer rows.Close()

	var stocks []*models.Stock
	for rows.Next() {
		var stock models.Stock
		err := rows.Scan(
			&stock.ID, &stock.ArticleID, &stock.Quantity, &stock.Reserved,
			&stock.MinStock, &stock.MaxStock, &stock.Location,
			&stock.CreatedAt, &stock.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning stock: %w", err)
		}
		stocks = append(stocks, &stock)
	}

	return stocks, nil
}

// GetLowStocks obtiene stocks con cantidad baja
func (r *StockRepository) GetLowStocks(ctx context.Context) ([]*models.Stock, error) {
	query := `
		SELECT id, article_id, quantity, reserved, min_stock, max_stock, location, created_at, updated_at
		FROM stocks
		WHERE quantity <= min_stock
		ORDER BY (quantity - min_stock) ASC
	`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying low stocks: %w", err)
	}
	defer rows.Close()

	var stocks []*models.Stock
	for rows.Next() {
		var stock models.Stock
		err := rows.Scan(
			&stock.ID, &stock.ArticleID, &stock.Quantity, &stock.Reserved,
			&stock.MinStock, &stock.MaxStock, &stock.Location,
			&stock.CreatedAt, &stock.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning stock: %w", err)
		}
		stocks = append(stocks, &stock)
	}

	return stocks, nil
}

// Métodos auxiliares para cache
func (r *StockRepository) cacheStock(ctx context.Context, stock *models.Stock) {
	if r.redis == nil {
		return
	}

	cacheKey := fmt.Sprintf("stock:%s", stock.ArticleID)
	data, err := json.Marshal(stock)
	if err != nil {
		return
	}

	r.redis.Set(ctx, cacheKey, data, 5*time.Minute) // Cache por 5 minutos
}

func (r *StockRepository) invalidateStockCache(ctx context.Context, articleID string) {
	if r.redis == nil {
		return
	}

	cacheKey := fmt.Sprintf("stock:%s", articleID)
	r.redis.Del(ctx, cacheKey)
}