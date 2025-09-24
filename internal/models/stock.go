package models

import (
	"time"

	"github.com/google/uuid"
)

// Stock representa el stock de un artículo
type Stock struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ArticleID    string    `json:"article_id" db:"article_id"`
	Quantity     int       `json:"quantity" db:"quantity"`
	Reserved     int       `json:"reserved" db:"reserved"`
	MinStock     int       `json:"min_stock" db:"min_stock"`
	MaxStock     int       `json:"max_stock" db:"max_stock"`
	Location     string    `json:"location" db:"location"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// AvailableQuantity retorna la cantidad disponible (no reservada)
func (s *Stock) AvailableQuantity() int {
	return s.Quantity - s.Reserved
}

// IsLowStock verifica si el stock está por debajo del mínimo
func (s *Stock) IsLowStock() bool {
	return s.Quantity <= s.MinStock
}

// CanReserve verifica si se puede reservar una cantidad específica
func (s *Stock) CanReserve(quantity int) bool {
	return s.AvailableQuantity() >= quantity
}

// CreateStockRequest representa la estructura para crear un nuevo artículo
type CreateStockRequest struct {
	ArticleID string `json:"article_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"min=0"`
	MinStock  int    `json:"min_stock" validate:"min=0"`
	MaxStock  int    `json:"max_stock" validate:"min=0"`
	Location  string `json:"location"`
}

// UpdateStockRequest representa la estructura para actualizar stock
type UpdateStockRequest struct {
	Quantity int `json:"quantity" validate:"min=0"`
	MinStock int `json:"min_stock" validate:"min=0"`
	MaxStock int `json:"max_stock" validate:"min=0"`
}

// ReserveStockRequest representa la estructura para reservar stock
type ReserveStockRequest struct {
	ArticleID string `json:"article_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"min=1"`
	OrderID   string `json:"order_id" validate:"required"`
}

// StockMovementRequest representa la estructura para movimientos de stock
type StockMovementRequest struct {
	ArticleID string `json:"article_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"min=1"`
	Reason    string `json:"reason"`
}