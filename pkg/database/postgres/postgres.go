package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Config holds PostgreSQL connection configuration.
type Config struct {
	Host     string `env:"HOST"`
	Port     int    `env:"PORT"`
	Database string `env:"DATABASE"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	SSLMode  string `env:"SSL_MODE"`

	// Connection pool settings
	MaxOpenConns    int `env:"MAX_OPEN_CONNS"`
	MaxIdleConns    int `env:"MAX_IDLE_CONNS"`
	ConnMaxLifetime int `env:"CONN_MAX_LIFETIME"` // seconds
}

// DefaultConfig returns default PostgreSQL configuration.
func DefaultConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            5432,
		Database:        "postgres",
		User:            "postgres",
		Password:        "",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300, // 5 minutes
	}
}

// PostgresDB is a PostgreSQL database wrapper.
// Implements the database.DB interface.
type PostgresDB struct {
	db *sqlx.DB
}

// Connect creates a new PostgreSQL database connection.
func Connect(config Config) (*PostgresDB, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	// Open database connection
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// ============================================================================
// database.DB Interface Implementation
// ============================================================================

// Exec executes a query without returning rows.
func (d *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows.
func (d *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row.
func (d *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// Get scans a single row into dest.
func (d *PostgresDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return d.db.GetContext(ctx, dest, query, args...)
}

// Select scans multiple rows into dest.
func (d *PostgresDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return d.db.SelectContext(ctx, dest, query, args...)
}

// BeginTx starts a new transaction.
func (d *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (database.Tx, error) {
	tx, err := d.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &PostgresTx{tx: tx}, nil
}

// Close closes the database connection.
func (d *PostgresDB) Close() error {
	return d.db.Close()
}

// Ping verifies the database connection is alive.
func (d *PostgresDB) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Stats returns database statistics.
func (d *PostgresDB) Stats() sql.DBStats {
	return d.db.Stats()
}

// ============================================================================
// Additional Helpers
// ============================================================================

// DB returns the underlying *sqlx.DB.
func (d *PostgresDB) DB() *sqlx.DB {
	return d.db
}

// NamedExec executes a named query.
func (d *PostgresDB) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return d.db.NamedExecContext(ctx, query, arg)
}

// ============================================================================
// Transaction
// ============================================================================

// PostgresTx is a PostgreSQL transaction wrapper.
// Implements the database.Tx interface.
type PostgresTx struct {
	tx *sqlx.Tx
}

// Exec executes a query within the transaction.
func (t *PostgresTx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// Query executes a query within the transaction.
func (t *PostgresTx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// QueryRow executes a query within the transaction.
func (t *PostgresTx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// Get scans a single row within the transaction.
func (t *PostgresTx) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return t.tx.GetContext(ctx, dest, query, args...)
}

// Select scans multiple rows within the transaction.
func (t *PostgresTx) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return t.tx.SelectContext(ctx, dest, query, args...)
}

// Commit commits the transaction.
func (t *PostgresTx) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction.
func (t *PostgresTx) Rollback() error {
	return t.tx.Rollback()
}

// Tx returns the underlying *sqlx.Tx.
func (t *PostgresTx) Tx() *sqlx.Tx {
	return t.tx
}

// NamedExec executes a named query within the transaction.
func (t *PostgresTx) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return t.tx.NamedExecContext(ctx, query, arg)
}
