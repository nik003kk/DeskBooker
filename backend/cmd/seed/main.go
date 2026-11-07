// Command seed loads repeatable test fixtures into an isolated development schema.
// Usage: APP_ENV=development go run ./cmd/seed
package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nik003kk/DeskBooker/backend/internal/config"
	"github.com/nik003kk/DeskBooker/backend/internal/database"
)

//go:embed seed.sql
var seedSQL string

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func requireDevelopment(env string, args []string) error {
	if len(args) != 0 {
		return errors.New("uporaba: APP_ENV=development go run ./cmd/seed")
	}
	if env != "development" {
		return errors.New("seed je dovoljen le ob APP_ENV=development; nikoli v produkciji")
	}
	return nil
}

func run(args []string) error {
	if err := config.LoadLocalEnv(); err != nil {
		return fmt.Errorf("napaka pri nalaganju konfiguracije: %w", err)
	}
	if err := requireDevelopment(os.Getenv("APP_ENV"), args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := database.Open(ctx)
	if err != nil {
		// Do not print connection errors that could contain credentials.
		return errors.New("ni povezave s PostgreSQL (preveri Docker, port in .env)")
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return errors.New("ni mogoče začeti transakcije za seed")
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	// Simple protocol supports the multiple SQL statements embedded above.
	// Transaction ensures all fixture tables are updated together or not at all.
	if _, err := tx.Exec(ctx, seedSQL, pgx.QueryExecModeSimpleProtocol); err != nil {
		return errors.New("seed ni uspel (preveri, ali je migracija 000002 izvedena)")
	}
	if err := tx.Commit(ctx); err != nil {
		return errors.New("ni mogoče potrditi seed transakcije")
	}
	log.Println("Razvojni testni podatki so uspešno vstavljeni ali posodobljeni.")
	return nil
}
