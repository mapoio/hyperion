package gorm

import (
	"database/sql"

	"gorm.io/gorm"

	"github.com/mapoio/hyperion"
)

// gormUnitOfWork implements hyperion.UnitOfWork interface using GORM.
// It provides declarative transaction management with automatic commit/rollback.
type gormUnitOfWork struct {
	db *gorm.DB
}

// Ensure interface compliance at compile time.
var _ hyperion.UnitOfWork = (*gormUnitOfWork)(nil)

// WithTransaction executes fn within a database transaction.
// If fn returns an error or panics, the transaction is rolled back.
// Otherwise, the transaction is committed.
//
// The Context passed to fn will have its DB() method return
// the transaction executor instead of the default executor.
//
// Example:
//
//	err := uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
//	    // All operations using txCtx.DB() will be part of the transaction
//	    return userRepo.Create(txCtx, user)
//	})
func (u *gormUnitOfWork) WithTransaction(ctx hyperion.Context, fn func(txCtx hyperion.Context) error) error {
	return u.WithTransactionOptions(ctx, nil, fn)
}

// WithTransactionOptions executes fn within a database transaction with custom options.
// It supports setting isolation level and read-only mode.
//
// Example:
//
//	opts := &hyperion.TransactionOptions{
//	    Isolation: hyperion.IsolationLevelSerializable,
//	    ReadOnly:  false,
//	}
//	err := uow.WithTransactionOptions(ctx, opts, func(txCtx hyperion.Context) error {
//	    return orderRepo.Create(txCtx, order)
//	})
func (u *gormUnitOfWork) WithTransactionOptions(
	ctx hyperion.Context,
	opts *hyperion.TransactionOptions,
	fn func(txCtx hyperion.Context) error,
) error {
	// Convert options if provided
	var txOpts *sql.TxOptions
	if opts != nil {
		txOpts = &sql.TxOptions{
			Isolation: toSQLIsolation(opts.Isolation),
			ReadOnly:  opts.ReadOnly,
		}
	}

	// Use GORM's transaction helper which handles panic recovery
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create transaction executor
		txExecutor := &gormExecutor{
			db:   tx,
			isTx: true,
		}

		// Inject transaction executor into context
		txCtx := hyperion.WithDB(ctx, txExecutor)

		// Execute user function
		// GORM's Transaction() already handles panic recovery and rollback
		return fn(txCtx)
	}, txOpts)
}

// toSQLIsolation converts hyperion.IsolationLevel to sql.IsolationLevel.
func toSQLIsolation(level hyperion.IsolationLevel) sql.IsolationLevel {
	switch level {
	case hyperion.IsolationLevelDefault:
		return sql.LevelDefault
	case hyperion.IsolationLevelReadUncommitted:
		return sql.LevelReadUncommitted
	case hyperion.IsolationLevelReadCommitted:
		return sql.LevelReadCommitted
	case hyperion.IsolationLevelRepeatableRead:
		return sql.LevelRepeatableRead
	case hyperion.IsolationLevelSerializable:
		return sql.LevelSerializable
	default:
		return sql.LevelDefault
	}
}
