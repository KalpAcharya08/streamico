package database

// =============================================================================
// 📖 LEARNING NOTES — MySQL Connection Pool
// =============================================================================
//
// WHAT IS A CONNECTION POOL?
// Opening a database connection is expensive (network handshake, authentication).
// A "pool" keeps multiple connections open and reuses them:
//
//   Without pool: Open → Query → Close → Open → Query → Close (slow!)
//   With pool:    [Pool of 10 connections] → borrow → query → return (fast!)
//
// Go's database/sql INCLUDES built-in pooling — unlike PostgreSQL where we
// needed a separate pgxpool library. Just configure MaxOpenConns, etc.
//
// WHY go-sql-driver/mysql?
// It's the most popular and battle-tested MySQL driver for Go.
// It implements Go's standard database/sql interface, so switching databases
// later is easier (same API, different driver).
//
// DSN FORMAT (Data Source Name):
//   user:password@tcp(host:port)/dbname?parseTime=true
//   Example: streaming:secret@tcp(localhost:3306)/streaming?parseTime=true
//
// parseTime=true → automatically converts MySQL DATETIME to Go time.Time
// =============================================================================

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // blank import registers the MySQL driver
	"go.uber.org/zap"
)

// MySQLConfig holds all settings needed to connect to MySQL.
type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// DSN builds the MySQL connection string.
// Format: user:password@tcp(host:port)/dbname?parseTime=true&charset=utf8mb4
//
// Example: streaming:secret@tcp(localhost:3306)/streaming?parseTime=true
func (c MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		c.User, c.Password, c.Host, c.Port, c.DBName,
	)
}

// NewMySQLPool creates a connection pool to MySQL.
//
// Parameters:
//   - ctx    → context for timeout/cancellation
//   - cfg    → database connection settings
//   - logger → for logging connection status
//
// Returns:
//   - *sql.DB → the connection pool (use this in your services)
//   - error   → if connection fails
func NewMySQLPool(ctx context.Context, cfg MySQLConfig, logger *zap.Logger) (*sql.DB, error) {
	// ─── Step 1: Open the database connection ────────────────────────
	// sql.Open doesn't actually connect — it just validates the DSN.
	// The actual connection happens on the first query or Ping.
	logger.Info("connecting to MySQL",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.String("database", cfg.DBName),
	)

	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// ─── Step 2: Configure pool settings ─────────────────────────────
	// Go's database/sql has built-in pool management via these settings.
	db.SetMaxOpenConns(25)                  // Max open connections at any time
	db.SetMaxIdleConns(5)                   // Keep at least 5 idle connections
	db.SetConnMaxLifetime(1 * time.Hour)    // Recycle connections after 1 hour
	db.SetConnMaxIdleTime(30 * time.Minute) // Close idle connections after 30 min

	// ─── Step 3: Verify the connection works (ping test) ─────────────
	// PingContext actually connects and checks the database is reachable.
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("✅ connected to MySQL successfully",
		zap.Int("max_open_connections", 25),
		zap.Int("max_idle_connections", 5),
	)

	return db, nil
}
