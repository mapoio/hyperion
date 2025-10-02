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
