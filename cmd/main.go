package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatiasTelo/stockgo/internal/config"
	"github.com/MatiasTelo/stockgo/internal/database"
	"github.com/MatiasTelo/stockgo/internal/handlers"
	"github.com/MatiasTelo/stockgo/internal/messaging"
	"github.com/MatiasTelo/stockgo/internal/repository"
	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Conectar a bases de datos
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Configurar RabbitMQ
	rabbitMQ, err := messaging.NewRabbitMQService(&cfg.RabbitMQ)
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v", err)
		// La aplicación puede funcionar sin RabbitMQ, pero con funcionalidad limitada
	}
	defer func() {
		if rabbitMQ != nil {
			rabbitMQ.Close()
		}
	}()

	// Crear repositorios
	stockRepo := repository.NewStockRepository(db.PG, db.Redis)
	eventRepo := repository.NewStockEventRepository(db.PG)

	// Crear publisher para low stock
	var lowStockPublisher messaging.MessagePublisher
	if rabbitMQ != nil {
		lowStockPublisher, err = messaging.NewLowStockPublisher(rabbitMQ.GetConnection())
		if err != nil {
			log.Printf("Warning: Failed to create low stock publisher: %v", err)
		}
	}

	// Crear servicios
	stockService := service.NewStockService(stockRepo, eventRepo, lowStockPublisher)

	// Crear handlers
	addArticleHandler := handlers.NewAddArticleHandler(stockService)
	replenishHandler := handlers.NewReplenishStockHandler(stockService)
	deductHandler := handlers.NewDeductStockHandler(stockService)
	reserveHandler := handlers.NewReserveStockHandler(stockService)
	cancelHandler := handlers.NewCancelReservationHandler(stockService)
	confirmHandler := handlers.NewConfirmReservationHandler(stockService)
	lowStockHandler := handlers.NewLowStockHandler(stockService)

	// Configurar Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
				"code":  code,
			})
		},
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	})

	// Middlewares
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"service":   "stock-service",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	api := app.Group("/api")
	v1 := api.Group("/stock")

	// Article management routes
	v1.Post("/articles", addArticleHandler.Handle)
	v1.Get("/articles", addArticleHandler.GetAllArticles)
	v1.Get("/articles/:articleId", addArticleHandler.GetArticle)
	v1.Get("/articles/:articleId/events", addArticleHandler.GetArticleEvents)

	// Stock operations routes
	v1.Put("/replenish", replenishHandler.Handle)

	v1.Put("/deduct", deductHandler.Handle)

	// Reservation routes
	v1.Put("/reserve", reserveHandler.Handle)

	v1.Put("/cancel-reservation", cancelHandler.Handle)

	v1.Put("/confirm-reservation", confirmHandler.Handle)

	// Low stock and alerts routes
	v1.Get("/low-stock", lowStockHandler.Handle)
	v1.Get("/alerts/summary", lowStockHandler.GetAlertsSummary)

	// Configurar consumidores de RabbitMQ
	if rabbitMQ != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Insufficient Stock Publisher
		insufficientStockPublisher, err := messaging.NewInsufficientStockPublisher(rabbitMQ.GetConnection())
		if err != nil {
			log.Printf("Warning: Failed to create insufficient stock publisher: %v", err)
		} else {
			defer insufficientStockPublisher.Close()
		}

		// Order Placed Consumer
		orderPlacedConsumer, err := messaging.NewOrderPlacedConsumer(stockService, rabbitMQ.GetConnection(), insufficientStockPublisher)
		if err != nil {
			log.Printf("Warning: Failed to create order placed consumer: %v", err)
		} else {
			if err := orderPlacedConsumer.StartConsuming(ctx); err != nil {
				log.Printf("Warning: Failed to start order placed consumer: %v", err)
			} else {
				defer orderPlacedConsumer.Close()
			}
		}

		// Order Confirmed Consumer
		orderConfirmedConsumer, err := messaging.NewOrderConfirmedConsumer(stockService, rabbitMQ.GetConnection())
		if err != nil {
			log.Printf("Warning: Failed to create order confirmed consumer: %v", err)
		} else {
			if err := orderConfirmedConsumer.StartConsuming(ctx); err != nil {
				log.Printf("Warning: Failed to start order confirmed consumer: %v", err)
			} else {
				defer orderConfirmedConsumer.Close()
			}
		}

		// Order Canceled Consumer
		orderCanceledConsumer, err := messaging.NewOrderCanceledConsumer(stockService, rabbitMQ.GetConnection())
		if err != nil {
			log.Printf("Warning: Failed to create order canceled consumer: %v", err)
		} else {
			if err := orderCanceledConsumer.StartConsuming(ctx); err != nil {
				log.Printf("Warning: Failed to start order canceled consumer: %v", err)
			} else {
				defer orderCanceledConsumer.Close()
			}
		}
	}

	// Configurar graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Iniciar servidor
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Stock service starting on %s", serverAddr)
	log.Printf("API documentation available at http://%s/api/stock", serverAddr)

	if err := app.Listen(serverAddr); err != nil {
		log.Fatal("Server failed to start:", err)
	}

	log.Println("Stock service stopped")
}
