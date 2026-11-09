package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const connectionTimeout = 5 * time.Second

// ConnectionString constructs a PostgreSQL connection URL from environment variables.
// DATABASE_URL takes precedence. For local Compose, POSTGRES_PASSWORD is required.
// Credentials are never embedded in source control.
func ConnectionString(getenv func(string) string) (string, error) {
	if dsn := getenv("DATABASE_URL"); dsn != "" {
		return dsn, nil
	}

	password := getenv("POSTGRES_PASSWORD")
	if password == "" {
		return "", errors.New("missing DATABASE_URL or POSTGRES_PASSWORD")
	}

	host := getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", errors.New("POSTGRES_PORT must be between 1 and 65535")
	}

	user := getenv("POSTGRES_USER")
	if user == "" {
		user = "deskbooker"
	}
	name := getenv("POSTGRES_DB")
	if name == "" {
		name = "deskbooker"
	}
	sslmode := getenv("POSTGRES_SSLMODE")
	if sslmode == "" {
		sslmode = "disable" // Local development only; configure TLS for production.
	}

	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + name,
	}
	query := dsn.Query()
	query.Set("sslmode", sslmode)
	dsn.RawQuery = query.Encode()

	return dsn.String(), nil
}

// Open creates a bounded pool and verifies the initial connection.
// The caller must close the pool using pool.Close().
func Open(ctx context.Context) (*pgxpool.Pool, error) {
	dsn, err := ConnectionString(os.Getenv)
	if err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL configuration: %w", err)
	}
	config.MaxConns = 10
	config.MinConns = 1

	connectCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, config)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("PostgreSQL is unavailable: %w", err)
	}
	return pool, nil
}
