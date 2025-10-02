package gorm

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	"github.com/mapoio/hyperion"
)

// gormExecutor implements hyperion.Executor interface using GORM.
// It wraps a *gorm.DB instance and tracks whether it represents a transaction.
type gormExecutor struct {
	db   *gorm.DB
	isTx bool // Tracks if this executor is a transaction
}

// Ensure interface compliance at compile time.
var _ hyperion.Executor = (*gormExecutor)(nil)

// Exec executes a SQL statement without returning rows.
// It's typically used for INSERT, UPDATE, DELETE, or DDL statements.
func (e *gormExecutor) Exec(ctx context.Context, sql string, args ...any) error {
	result := e.db.WithContext(ctx).Exec(sql, args...)
	if result.Error != nil {
		return fmt.Errorf("exec failed: %w", result.Error)
	}
	return nil
}

// Query executes a SQL query and scans the results into dest.
// dest should be a pointer to a slice of structs or a pointer to a struct.
func (e *gormExecutor) Query(ctx context.Context, dest any, sql string, args ...any) error {
	result := e.db.WithContext(ctx).Raw(sql, args...).Scan(dest)
	if result.Error != nil {
		return fmt.Errorf("query failed: %w", result.Error)
	}
	return nil
}

// Begin starts a new transaction and returns a transaction executor.
// The returned executor will have isTx set to true.
func (e *gormExecutor) Begin(ctx context.Context) (hyperion.Executor, error) {
	tx := e.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", tx.Error)
	}

	return &gormExecutor{
		db:   tx,
		isTx: true,
	}, nil
}

// BeginTx starts a new transaction with custom options.
func (e *gormExecutor) BeginTx(ctx context.Context, opts *sql.TxOptions) (hyperion.Executor, error) {
	tx := e.db.WithContext(ctx).Begin(opts)
	if tx.Error != nil {
		return nil, fmt.Errorf("begin transaction with options failed: %w", tx.Error)
	}

	return &gormExecutor{
		db:   tx,
		isTx: true,
	}, nil
}

// Commit commits the current transaction.
// Returns an error if this executor is not a transaction.
func (e *gormExecutor) Commit() error {
	if !e.isTx {
		return fmt.Errorf("not in transaction: cannot commit non-transaction executor")
	}

	if err := e.db.Commit().Error; err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// Rollback rolls back the current transaction.
// Returns an error if this executor is not a transaction.
func (e *gormExecutor) Rollback() error {
	if !e.isTx {
		return fmt.Errorf("not in transaction: cannot rollback non-transaction executor")
	}

	if err := e.db.Rollback().Error; err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	return nil
}

// Unwrap returns the underlying *gorm.DB instance.
// This allows users to access GORM-specific features when needed.
// Returns nil if no underlying implementation exists.
func (e *gormExecutor) Unwrap() any {
	return e.db
}
