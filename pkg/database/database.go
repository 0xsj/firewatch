package database

import (
	"context"
	"database/sql"
)

// DB is the database interface (port).
// All database adapters must implement this interface.
//
// This abstraction allows:
//   - Swapping database implementations
//   - Easy testing with mocks
//   - Query execution without vendor lock-in
type DB interface {
	// Exec executes a query without returning any rows.
	// Used for INSERT, UPDATE, DELETE statements.
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Query executes a query that returns rows.
	// Used for SELECT statements.
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes a query that returns at most one row.
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Get scans a single row into dest.
	// Convenient wrapper around QueryRow + Scan.
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// Select scans multiple rows into dest (slice).
	// Convenient wrapper around Query + Scan loop.
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// BeginTx starts a new transaction.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)

	// Ping verifies connection to the database.
	Ping(ctx context.Context) error

	// Close closes the database connection.
	Close() error

	// Stats returns database statistics.
	Stats() sql.DBStats
}

// Tx is the transaction interface.
// Provides the same query methods as DB but within a transaction context.
type Tx interface {
	// Exec executes a query within the transaction.
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Query executes a query within the transaction.
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes a query within the transaction.
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Get scans a single row within the transaction.
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// Select scans multiple rows within the transaction.
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// Commit commits the transaction.
	Commit() error

	// Rollback rolls back the transaction.
	Rollback() error
}

// Querier is an interface that can execute queries.
// Both DB and Tx implement this interface.
type Querier interface {
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
