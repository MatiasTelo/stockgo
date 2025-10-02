package main

import (
	"fmt"
	"log"
	"os"

	"github.com/MatiasTelo/stockgo/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run migrate_runner.go [up|down|version|force <version>]")
	}

	command := os.Args[1]

	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Crear instancia de migrate
	m, err := migrate.New(
		"file://migrations",
		cfg.DatabaseURL())
	if err != nil {
		log.Fatal("Failed to create migrate instance:", err)
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Failed to run migrations:", err)
		}
		fmt.Println("Migrations applied successfully!")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Failed to rollback migrations:", err)
		}
		fmt.Println("Migrations rolled back successfully!")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatal("Failed to get migration version:", err)
		}
		fmt.Printf("Current migration version: %d (dirty: %t)\n", version, dirty)

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run migrate_runner.go force <version>")
		}
		version := os.Args[2]
		var v int
		if _, err := fmt.Sscanf(version, "%d", &v); err != nil {
			log.Fatal("Invalid version number:", version)
		}
		if err := m.Force(v); err != nil {
			log.Fatal("Failed to force migration version:", err)
		}
		fmt.Printf("Forced migration to version: %d\n", v)

	default:
		log.Fatal("Unknown command:", command)
	}
}
