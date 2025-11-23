package database

import (
	"context"
	"fmt"
)

// TxFunc is a function that executes within a transaction.
// If the function returns an error, the transaction is rolled back.
// If the function returns nil, the transaction is committed.
type TxFunc func(ctx context.Context, tx Tx) error

// WithTx executes a function within a transaction.
// Automatically handles commit/rollback based on the function's return value.
//
// Example:
//
//	err := database.WithTx(ctx, db, func(ctx context.Context, tx Tx) error {
//	    // All operations here are within a transaction
//	    _, err := tx.Exec(ctx, "INSERT INTO users ...")
//	    if err != nil {
//	        return err // Transaction will be rolled back
//	    }
//
//	    _, err = tx.Exec(ctx, "INSERT INTO profiles ...")
//	    return err // If nil, transaction commits; if error, rollback
//	})
func WithTx(ctx context.Context, db DB, fn TxFunc) error {
	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback in case of panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // Re-panic after rollback
		}
	}()

	// Execute function
	if err := fn(ctx, tx); err != nil {
		// Function returned error, rollback
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Success, commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetQuerier returns either the transaction if present in context,
// or the database otherwise.
// This allows repositories to work transparently with or without transactions.
//
// Example:
//
//	func (r *UserRepository) Create(ctx context.Context, user *User) error {
//	    q := database.GetQuerier(ctx, r.db)
//	    _, err := q.Exec(ctx, "INSERT INTO users ...")
//	    return err
//	}
func GetQuerier(ctx context.Context, db DB) Querier {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return db
}

// ============================================================================
// Transaction Context
// ============================================================================

type txKey struct{}

// WithTxContext stores a transaction in the context.
// This allows passing transactions through the call stack without
// explicitly passing the Tx parameter.
//
// Example:
//
//	tx, _ := db.BeginTx(ctx, nil)
//	ctx = database.WithTxContext(ctx, tx)
//
//	// Now any repository method can access the transaction
//	err := userRepo.Create(ctx, user)
func WithTxContext(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext extracts a transaction from context.
// Returns nil if no transaction is present.
func TxFromContext(ctx context.Context) Tx {
	if tx, ok := ctx.Value(txKey{}).(Tx); ok {
		return tx
	}
	return nil
}

// InTx returns true if the context contains a transaction.
func InTx(ctx context.Context) bool {
	return TxFromContext(ctx) != nil
}
