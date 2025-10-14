package service

import (
	"context"
	"fmt"

	"github.com/MatiasTelo/stockgo/internal/models"
	"github.com/MatiasTelo/stockgo/internal/repository"
)

type StockService struct {
	stockRepo        *repository.StockRepository
	eventRepo        *repository.StockEventRepository
	messagingService MessagePublisher
}

type MessagePublisher interface {
	PublishLowStockAlert(ctx context.Context, articleID string, currentQuantity, minStock int) error
}

func NewStockService(
	stockRepo *repository.StockRepository,
	eventRepo *repository.StockEventRepository,
	messagingService MessagePublisher,
) *StockService {
	return &StockService{
		stockRepo:        stockRepo,
		eventRepo:        eventRepo,
		messagingService: messagingService,
	}
}

// CreateStock crea un nuevo artículo en el inventario
func (s *StockService) CreateStock(ctx context.Context, req *models.CreateStockRequest) (*models.Stock, error) {
	// Validar que el artículo no exista
	existingStock, _ := s.stockRepo.GetStockByArticleID(ctx, req.ArticleID)
	if existingStock != nil {
		return nil, fmt.Errorf("article with ID %s already exists", req.ArticleID)
	}

	stock := &models.Stock{
		ArticleID: req.ArticleID,
		Quantity:  req.Quantity,
		Reserved:  0,
		MinStock:  req.MinStock,
		MaxStock:  req.MaxStock,
		Location:  req.Location,
	}

	if err := s.stockRepo.CreateStock(ctx, stock); err != nil {
		return nil, fmt.Errorf("error creating stock: %w", err)
	}

	// Crear evento de stock
	event := &models.StockEvent{
		ArticleID: req.ArticleID,
		EventType: models.EventTypeAdd,
		Quantity:  req.Quantity,
		Reason:    "Nuevo artículo agregado al inventario",
	}

	if err := s.eventRepo.CreateStockEvent(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Could not create stock event: %v\n", err)
	}

	return stock, nil
}

// ReplenishStock repone stock de un artículo existente
func (s *StockService) ReplenishStock(ctx context.Context, articleID string, quantity int, reason string) (*models.Stock, error) {
	stock, err := s.stockRepo.GetStockByArticleID(ctx, articleID)
	if err != nil {
		return nil, fmt.Errorf("article not found: %w", err)
	}

	newQuantity := stock.Quantity + quantity
	if err := s.stockRepo.UpdateStockQuantity(ctx, articleID, newQuantity); err != nil {
		return nil, fmt.Errorf("error updating stock: %w", err)
	}

	// Crear evento de stock
	event := &models.StockEvent{
		ArticleID: articleID,
		EventType: models.EventTypeReplenish,
		Quantity:  quantity,
		Reason:    reason,
	}

	if err := s.eventRepo.CreateStockEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Could not create stock event: %v\n", err)
	}

	// Obtener el stock actualizado
	return s.stockRepo.GetStockByArticleID(ctx, articleID)
}

// DeductStock descuenta stock directamente
func (s *StockService) DeductStock(ctx context.Context, articleID string, quantity int, reason string) (*models.Stock, error) {
	stock, err := s.stockRepo.GetStockByArticleID(ctx, articleID)
	if err != nil {
		return nil, fmt.Errorf("article not found: %w", err)
	}

	if stock.Quantity < quantity {
		return nil, fmt.Errorf("insufficient stock: available %d, requested %d", stock.Quantity, quantity)
	}

	newQuantity := stock.Quantity - quantity
	if err := s.stockRepo.UpdateStockQuantity(ctx, articleID, newQuantity); err != nil {
		return nil, fmt.Errorf("error updating stock: %w", err)
	}

	// Crear evento de stock
	event := &models.StockEvent{
		ArticleID: articleID,
		EventType: models.EventTypeDeduct,
		Quantity:  quantity,
		Reason:    reason,
	}

	if err := s.eventRepo.CreateStockEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Could not create stock event: %v\n", err)
	}

	// Obtener el stock actualizado y verificar si está bajo
	updatedStock, _ := s.stockRepo.GetStockByArticleID(ctx, articleID)
	if updatedStock != nil && updatedStock.IsLowStock() {
		s.messagingService.PublishLowStockAlert(ctx, articleID, updatedStock.Quantity, updatedStock.MinStock)
	}

	return updatedStock, nil
}

// ReserveStock reserva una cantidad de stock para una orden
func (s *StockService) ReserveStock(ctx context.Context, req *models.ReserveStockRequest) error {
	// Verificar si ya existe una reserva activa para este order_id y article_id específicos
	hasReservation, err := s.eventRepo.HasActiveReservation(ctx, req.OrderID, req.ArticleID)
	if err != nil {
		return fmt.Errorf("error checking existing reservations: %w", err)
	}

	if hasReservation {
		return fmt.Errorf("order %s already has an active reservation for article %s", req.OrderID, req.ArticleID)
	}

	// Verificar que hay stock suficiente y reservarlo
	if err := s.stockRepo.ReserveStock(ctx, req.ArticleID, req.Quantity); err != nil {
		return fmt.Errorf("error reserving stock: %w", err)
	}

	// Crear evento de stock
	event := &models.StockEvent{
		ArticleID: req.ArticleID,
		EventType: models.EventTypeReserve,
		Quantity:  req.Quantity,
		OrderID:   &req.OrderID,
		Reason:    fmt.Sprintf("Stock reservado para orden %s", req.OrderID),
	}

	if err := s.eventRepo.CreateStockEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Could not create stock event: %v\n", err)
	}

	return nil
}

// GetStock obtiene información de stock por artículo
func (s *StockService) GetStock(ctx context.Context, articleID string) (*models.Stock, error) {
	return s.stockRepo.GetStockByArticleID(ctx, articleID)
}

// GetAllStocks obtiene todos los stocks
func (s *StockService) GetAllStocks(ctx context.Context) ([]*models.Stock, error) {
	return s.stockRepo.GetAllStocks(ctx)
}

// GetStockEvents obtiene eventos de stock por artículo
func (s *StockService) GetStockEvents(ctx context.Context, articleID string, limit int) ([]*models.StockEvent, error) {
	if limit <= 0 {
		limit = 50 // límite por defecto
	}
	return s.eventRepo.GetStockEventsByArticleID(ctx, articleID, limit)
}

// GetLowStocks obtiene artículos con stock bajo
func (s *StockService) GetLowStocks(ctx context.Context) ([]*models.Stock, error) {
	return s.stockRepo.GetLowStocks(ctx)
}

// CancelReservationByOrderID cancela una reserva usando order_id y article_id
func (s *StockService) CancelReservationByOrderID(ctx context.Context, orderID, articleID, reason string) error {
	// Buscar eventos de reserva para este order_id y article_id
	events, err := s.eventRepo.GetStockEventsByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("error getting events for order: %w", err)
	}

	var reserveEvent *models.StockEvent
	var hasCancel, hasConfirm bool

	// Buscar el evento de reserva y verificar si ya fue cancelada o confirmada
	for _, event := range events {
		if event.ArticleID == articleID {
			switch event.EventType {
			case models.EventTypeReserve:
				reserveEvent = event
			case models.EventTypeCancelReserve:
				hasCancel = true
			case models.EventTypeDeduct: // DEDUCT representa confirmación
				hasConfirm = true
			}
		}
	}

	if reserveEvent == nil {
		return fmt.Errorf("no active reservation found for this order and article")
	}

	if hasCancel {
		return fmt.Errorf("reservation has already been cancelled")
	}

	if hasConfirm {
		return fmt.Errorf("reservation has already been confirmed")
	}

	// Liberar el stock reservado
	if err := s.stockRepo.CancelReservation(ctx, articleID, reserveEvent.Quantity); err != nil {
		return fmt.Errorf("error canceling stock reservation: %w", err)
	}

	// Crear evento de cancelación
	cancelReason := reason
	if cancelReason == "" {
		cancelReason = fmt.Sprintf("Reserva cancelada para orden %s", orderID)
	}

	event := &models.StockEvent{
		ArticleID: articleID,
		EventType: models.EventTypeCancelReserve,
		Quantity:  reserveEvent.Quantity,
		OrderID:   &orderID,
		Reason:    cancelReason,
	}

	if err := s.eventRepo.CreateStockEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Could not create stock event: %v\n", err)
	}

	return nil
}

// ConfirmReservationByOrderID confirma una reserva usando order_id y article_id
func (s *StockService) ConfirmReservationByOrderID(ctx context.Context, orderID, articleID, reason string) error {
	// Buscar eventos de reserva para este order_id y article_id
	events, err := s.eventRepo.GetStockEventsByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("error getting events for order: %w", err)
	}

	var reserveEvent *models.StockEvent
	var hasCancel, hasConfirm bool

	// Buscar el evento de reserva y verificar si ya fue cancelada o confirmada
	for _, event := range events {
		if event.ArticleID == articleID {
			switch event.EventType {
			case models.EventTypeReserve:
				reserveEvent = event
			case models.EventTypeCancelReserve:
				hasCancel = true
			case models.EventTypeDeduct: // DEDUCT representa confirmación
				hasConfirm = true
			}
		}
	}

	if reserveEvent == nil {
		return fmt.Errorf("no active reservation found for this order and article")
	}

	if hasCancel {
		return fmt.Errorf("reservation has already been cancelled")
	}

	if hasConfirm {
		return fmt.Errorf("reservation has already been confirmed")
	}

	// Confirmar la reserva (descontar stock y liberar reserved)
	if err := s.stockRepo.ConfirmReservation(ctx, articleID, reserveEvent.Quantity); err != nil {
		return fmt.Errorf("error confirming reservation: %w", err)
	}

	// Crear evento de confirmación
	confirmReason := reason
	if confirmReason == "" {
		confirmReason = fmt.Sprintf("Stock descontado por confirmación de orden %s", orderID)
	}

	event := &models.StockEvent{
		ArticleID: articleID,
		EventType: models.EventTypeDeduct,
		Quantity:  reserveEvent.Quantity,
		OrderID:   &orderID,
		Reason:    confirmReason,
	}

	if err := s.eventRepo.CreateStockEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Could not create stock event: %v\n", err)
	}

	// Verificar si el stock está bajo después de la confirmación
	stock, _ := s.stockRepo.GetStockByArticleID(ctx, articleID)
	if stock != nil && stock.IsLowStock() {
		s.messagingService.PublishLowStockAlert(ctx, articleID, stock.Quantity, stock.MinStock)
	}

	return nil
}
