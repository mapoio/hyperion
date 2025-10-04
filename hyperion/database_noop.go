package hyperion

import (
	"context"
	"errors"
)

var (
	// ErrNoOpDatabase is returned by no-op database operations.
	ErrNoOpDatabase = errors.New("no-op database: no adapter provided")
)

// noopDatabase is a no-op implementation of Database interface.
type noopDatabase struct{}

// NewNoOpDatabase creates a new no-op Database implementation.
func NewNoOpDatabase() Database {
	return &noopDatabase{}
}

func (db *noopDatabase) Executor() Executor {
	return &noopExecutor{}
}

func (db *noopDatabase) Health(ctx context.Context) error {
	return ErrNoOpDatabase
}

func (db *noopDatabase) Close() error {
	return nil
}

// noopExecutor is a no-op implementation of Executor interface.
type noopExecutor struct{}

func (e *noopExecutor) Exec(ctx context.Context, sql string, args ...any) error {
	return ErrNoOpDatabase
}

func (e *noopExecutor) Query(ctx context.Context, dest any, sql string, args ...any) error {
	return ErrNoOpDatabase
}

func (e *noopExecutor) Begin(ctx context.Context) (Executor, error) {
	return e, nil
}

func (e *noopExecutor) Commit() error   { return nil }
func (e *noopExecutor) Rollback() error { return nil }
func (e *noopExecutor) Unwrap() any     { return nil }

// noopUnitOfWork is a no-op implementation of UnitOfWork interface.
type noopUnitOfWork struct{}

// NewNoOpUnitOfWork creates a new no-op UnitOfWork implementation.
func NewNoOpUnitOfWork() UnitOfWork {
	return &noopUnitOfWork{}
}

func (u *noopUnitOfWork) WithTransaction(ctx Context, fn func(txCtx Context) error) error {
	// No-op: just execute the function with the original context
	return fn(ctx)
}

func (u *noopUnitOfWork) WithTransactionOptions(ctx Context, opts *TransactionOptions, fn func(txCtx Context) error) error {
	// No-op: ignore options and just execute the function
	return fn(ctx)
}
