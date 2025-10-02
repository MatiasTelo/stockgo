package models

import (
	"time"

	"github.com/google/uuid"
)

// StockEventType representa los tipos de eventos de stock
type StockEventType string

const (
	EventTypeAdd         StockEventType = "ADD"
	EventTypeReplenish   StockEventType = "REPLENISH"
	EventTypeDeduct      StockEventType = "DEDUCT"
	EventTypeReserve     StockEventType = "RESERVE"
	EventTypeCancelReserve StockEventType = "CANCEL_RESERVE"
	EventTypeLowStock    StockEventType = "LOW_STOCK"
)

// StockEvent representa un evento del historial de stock
type StockEvent struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	ArticleID string         `json:"article_id" db:"article_id"`
	EventType StockEventType `json:"event_type" db:"event_type"`
	Quantity  int            `json:"quantity" db:"quantity"`
	OrderID   *string        `json:"order_id,omitempty" db:"order_id"`
	Reason    string         `json:"reason" db:"reason"`
	Metadata  string         `json:"metadata,omitempty" db:"metadata"` // JSON para datos adicionales
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

// CreateStockEventRequest representa la estructura para crear un evento
type CreateStockEventRequest struct {
	ArticleID string         `json:"article_id" validate:"required"`
	EventType StockEventType `json:"event_type" validate:"required"`
	Quantity  int            `json:"quantity" validate:"min=0"`
	OrderID   *string        `json:"order_id,omitempty"`
	Reason    string         `json:"reason"`
	Metadata  string         `json:"metadata,omitempty"`
}

// StockReservation representa una reserva de stock
type StockReservation struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ArticleID string    `json:"article_id" db:"article_id"`
	OrderID   string    `json:"order_id" db:"order_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Status    string    `json:"status" db:"status"` // ACTIVE, CONFIRMED, CANCELLED
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ReservationStatus representa los estados de una reserva
type ReservationStatus string

const (
	ReservationStatusActive    ReservationStatus = "ACTIVE"
	ReservationStatusConfirmed ReservationStatus = "CONFIRMED"
	ReservationStatusCancelled ReservationStatus = "CANCELLED"
	ReservationStatusExpired   ReservationStatus = "EXPIRED"
)