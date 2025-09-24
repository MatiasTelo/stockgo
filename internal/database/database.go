package database

import (
	"context"
	"fmt"
	"log"

	"github.com/MatiasTelo/stockgo/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Database struct {
	PG    *pgxpool.Pool
	Redis *redis.Client
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	// Conexión a PostgreSQL
	pg, err := pgxpool.New(context.Background(), cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("error connecting to PostgreSQL: %w", err)
	}

	// Verificar conexión a PostgreSQL
	if err := pg.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("error pinging PostgreSQL: %w", err)
	}

	log.Println("Connected to PostgreSQL successfully")

	// Conexión a Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Verificar conexión a Redis
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
	} else {
		log.Println("Connected to Redis successfully")
	}

	return &Database{
		PG:    pg,
		Redis: rdb,
	}, nil
}

func (db *Database) Close() {
	if db.PG != nil {
		db.PG.Close()
	}
	if db.Redis != nil {
		db.Redis.Close()
	}
}