package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MatiasTelo/stockgo/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReservationRepository struct {
	db *pgxpool.Pool
}

func NewReservationRepository(db *pgxpool.Pool) *ReservationRepository {
	return &ReservationRepository{
		db: db,
	}
}

// CreateReservation crea una nueva reserva
func (r *ReservationRepository) CreateReservation(ctx context.Context, reservation *models.StockReservation) error {
	query := `
		INSERT INTO stock_reservations (id, article_id, order_id, quantity, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	reservation.ID = uuid.New()
	reservation.CreatedAt = time.Now()
	reservation.UpdatedAt = time.Now()
	reservation.Status = string(models.ReservationStatusActive)
	
	// Reserva expira en 30 minutos por defecto
	if reservation.ExpiresAt.IsZero() {
		reservation.ExpiresAt = time.Now().Add(30 * time.Minute)
	}

	_, err := r.db.Exec(ctx, query,
		reservation.ID, reservation.ArticleID, reservation.OrderID,
		reservation.Quantity, reservation.Status, reservation.ExpiresAt,
		reservation.CreatedAt, reservation.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating reservation: %w", err)
	}

	return nil
}

// GetReservationByOrderAndArticle obtiene una reserva por orden y artículo
func (r *ReservationRepository) GetReservationByOrderAndArticle(ctx context.Context, orderID, articleID string) (*models.StockReservation, error) {
	query := `
		SELECT id, article_id, order_id, quantity, status, expires_at, created_at, updated_at
		FROM stock_reservations
		WHERE order_id = $1 AND article_id = $2 AND status = $3
	`
	
	var reservation models.StockReservation
	err := r.db.QueryRow(ctx, query, orderID, articleID, string(models.ReservationStatusActive)).Scan(
		&reservation.ID, &reservation.ArticleID, &reservation.OrderID,
		&reservation.Quantity, &reservation.Status, &reservation.ExpiresAt,
		&reservation.CreatedAt, &reservation.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("active reservation not found for order %s and article %s", orderID, articleID)
		}
		return nil, fmt.Errorf("error getting reservation: %w", err)
	}

	return &reservation, nil
}

// UpdateReservationStatus actualiza el estado de una reserva
func (r *ReservationRepository) UpdateReservationStatus(ctx context.Context, reservationID uuid.UUID, status models.ReservationStatus) error {
	query := `
		UPDATE stock_reservations 
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	
	result, err := r.db.Exec(ctx, query, string(status), time.Now(), reservationID)
	if err != nil {
		return fmt.Errorf("error updating reservation status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("reservation not found: %s", reservationID.String())
	}

	return nil
}

// GetExpiredReservations obtiene reservas expiradas
func (r *ReservationRepository) GetExpiredReservations(ctx context.Context) ([]*models.StockReservation, error) {
	query := `
		SELECT id, article_id, order_id, quantity, status, expires_at, created_at, updated_at
		FROM stock_reservations
		WHERE status = $1 AND expires_at < $2
	`
	
	rows, err := r.db.Query(ctx, query, string(models.ReservationStatusActive), time.Now())
	if err != nil {
		return nil, fmt.Errorf("error querying expired reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*models.StockReservation
	for rows.Next() {
		var reservation models.StockReservation
		err := rows.Scan(
			&reservation.ID, &reservation.ArticleID, &reservation.OrderID,
			&reservation.Quantity, &reservation.Status, &reservation.ExpiresAt,
			&reservation.CreatedAt, &reservation.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning reservation: %w", err)
		}
		reservations = append(reservations, &reservation)
	}

	return reservations, nil
}

// GetReservationsByOrderID obtiene todas las reservas de una orden
func (r *ReservationRepository) GetReservationsByOrderID(ctx context.Context, orderID string) ([]*models.StockReservation, error) {
	query := `
		SELECT id, article_id, order_id, quantity, status, expires_at, created_at, updated_at
		FROM stock_reservations
		WHERE order_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("error querying reservations by order: %w", err)
	}
	defer rows.Close()

	var reservations []*models.StockReservation
	for rows.Next() {
		var reservation models.StockReservation
		err := rows.Scan(
			&reservation.ID, &reservation.ArticleID, &reservation.OrderID,
			&reservation.Quantity, &reservation.Status, &reservation.ExpiresAt,
			&reservation.CreatedAt, &reservation.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning reservation: %w", err)
		}
		reservations = append(reservations, &reservation)
	}

	return reservations, nil
}