package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// PostgresTx is a PostgreSQL transaction wrapper.
// Implements the database.Tx interface.
type PostgresTx struct {
	tx *sqlx.Tx
}

// ============================================================================
// database.Tx Interface Implementation
// ============================================================================

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

// ============================================================================
// Additional Helpers
// ============================================================================

// Tx returns the underlying *sqlx.Tx.
func (t *PostgresTx) Tx() *sqlx.Tx {
	return t.tx
}

// NamedExec executes a named query within the transaction.
func (t *PostgresTx) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return t.tx.NamedExecContext(ctx, query, arg)
}
