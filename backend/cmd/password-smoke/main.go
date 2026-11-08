// Command password-smoke tests password hashing and lookup against PostgreSQL.
// Test users are created inside a transaction that is always rolled back.
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/auth"
    "github.com/nik003kk/DeskBooker/backend/internal/config"
    "github.com/nik003kk/DeskBooker/backend/internal/database"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
    "github.com/nik003kk/DeskBooker/backend/internal/repository"
)

func main() {
    if err := config.LoadLocalEnv(); err != nil { log.Fatal(err) }
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    pool, err := database.Open(ctx)
    if err != nil { log.Fatal(err) }
    defer pool.Close()

    tx, err := pool.Begin(ctx)
    if err != nil { log.Fatal(err) }
    defer func() { _ = tx.Rollback(context.Background()) }()

    if err := run(ctx, repository.New(tx)); err != nil { log.Fatal(err) }
    fmt.Println("DB-022 password smoke tests passed: hashed storage, email lookup, valid/invalid password, rollback")
}

func run(ctx context.Context, store *repository.Store) error {
    const password = "VeryStrongPass123!"
    hash, err := auth.HashPassword(password)
    if err != nil { return err }

    user, err := store.CreateUser(ctx, models.CreateUserInput{
        Email: "db022-password-smoke@example.test",
        PasswordHash: hash,
        FirstName: "Password",
        LastName: "Smoke",
    })
    if err != nil { return fmt.Errorf("create user: %w", err) }
    if user.PasswordHash == password || user.PasswordHash != hash {
        return errors.New("password was not stored as a bcrypt hash")
    }

    fetched, err := store.GetUserByEmail(ctx, "DB022-PASSWORD-SMOKE@EXAMPLE.TEST")
    if err != nil { return fmt.Errorf("email lookup: %w", err) }
    if fetched.ID != user.ID || fetched.PasswordHash != hash {
        return errors.New("stored hash does not match expected user")
    }
    if err := auth.VerifyPassword(fetched.PasswordHash, password); err != nil {
        return fmt.Errorf("correct password failed: %w", err)
    }
    if err := auth.VerifyPassword(fetched.PasswordHash, "WrongPassword123!"); !errors.Is(err, auth.ErrInvalidCredentials) {
        return fmt.Errorf("wrong password not rejected: %v", err)
    }
    if _, err := store.GetUserByEmail(ctx, "unknown-db022@example.test"); !errors.Is(err, pgx.ErrNoRows) {
        return fmt.Errorf("unknown user should return ErrNoRows, got %v", err)
    }
    return nil
}
