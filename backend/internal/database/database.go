package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	cf "github.com/seankim658/skullking/internal/config"
	l "github.com/seankim658/skullking/internal/logger"
)

var log = l.WithComponent(l.AppLog, "database-database")

// TODO : Should we eventually move to dependency injection rather than a global variable
var DB *sql.DB

// Interface that *sql.DB and *sql.Tx satisfy, allowing functions to accept either for
// executing queries. Used for functions that need to operate within a transaction or
// directly on the database connection pool.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Initializes a connection to the PostgreSQL database, sets up a connection pool, and pings
// the database to ensure connectivity.
func Connect(cfg *cf.Config) error {
	logger := l.WithSource(log, "Connect")
	logger.Info().Msg("Attempting to connect to the database...")

	// Construct the Data Source Name (DSN) for PostgreSQL
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	var err error
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create database connection pool")
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	DB.SetMaxOpenConns(25)                 // Maximum number of open connections to the database
	DB.SetMaxIdleConns(25)                 // Maximum number of connections in the idle connection pool
	DB.SetConnMaxLifetime(5 * time.Minute) // Maximum amount of time a connection may be reused
	DB.SetConnMaxIdleTime(1 * time.Minute) // Maximum amount of time a connection can be idle in the pool

	logger.Info().Msg("Pinging database to verify connection...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = DB.PingContext(ctx); err != nil {
		defer func() {
			if closeErr := DB.Close(); closeErr != nil {
				logger.Error().Err(closeErr).Msg("Failed to close database connection pool after ping failure")
			}
		}()
		logger.Error().Err(err).Msg("Failed to ping database")
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().Msg("Successfully connected to the database and verified connection")
	return nil
}

// Closes the database connection pool
func Close() {
	logger := l.WithSource(log, "Close")
	if DB != nil {
		logger.Info().Msg("Closing database connection pool...")
		if err := DB.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close database connection pool")
		} else {
			logger.Info().Msg("Database connection pool closed successfully")
		}
	}
}

// Returns a DBTX interface, which is used by the transaction if tx is not nil, otherwise
// it will be the global DB connection pool.
func GetQuerier(tx *sql.Tx) DBTX {
	if tx != nil {
		return tx
	}
	return DB
}
