package hyperion

import "context"

// Database provides database connectivity and transaction management.
type Database interface {
	// Executor returns the database executor.
	Executor() Executor

	// Health checks the database connection.
	Health(ctx context.Context) error

	// Close closes the database connection.
	Close() error
}

// Executor is the unified database operation interface.
// It abstracts GORM, sqlx, and other database libraries.
type Executor interface {
	// Exec executes a query without returning rows.
	Exec(ctx context.Context, sql string, args ...any) error

	// Query executes a query that returns rows into dest.
	Query(ctx context.Context, dest any, sql string, args ...any) error

	// Begin starts a new transaction and returns a transaction executor.
	Begin(ctx context.Context) (Executor, error)

	// Commit commits the current transaction.
	// Returns an error if not in a transaction.
	Commit() error

	// Rollback rolls back the current transaction.
	// Returns an error if not in a transaction.
	Rollback() error

	// Unwrap returns the underlying database implementation.
	// This allows accessing implementation-specific features when needed.
	// Returns nil if no underlying implementation exists.
	Unwrap() any
}

// UnitOfWork manages transaction boundaries.
// It provides a declarative way to handle database transactions.
type UnitOfWork interface {
	// WithTransaction executes fn within a database transaction.
	// If fn returns an error, the transaction is rolled back.
	// Otherwise, the transaction is committed.
	//
	// The Context passed to fn will have its DB() method return
	// the transaction executor instead of the default executor.
	WithTransaction(ctx Context, fn func(txCtx Context) error) error

	// WithTransactionOptions executes fn within a database transaction with options.
	WithTransactionOptions(ctx Context, opts *TransactionOptions, fn func(txCtx Context) error) error
}

// TransactionOptions configures transaction behavior.
type TransactionOptions struct {
	// Isolation sets the transaction isolation level.
	Isolation IsolationLevel

	// ReadOnly indicates the transaction should be read-only.
	ReadOnly bool
}

// IsolationLevel represents the transaction isolation level.
type IsolationLevel int

const (
	// IsolationLevelDefault uses the database's default isolation level.
	IsolationLevelDefault IsolationLevel = iota

	// IsolationLevelReadUncommitted allows reading uncommitted changes.
	IsolationLevelReadUncommitted

	// IsolationLevelReadCommitted prevents reading uncommitted changes.
	IsolationLevelReadCommitted

	// IsolationLevelRepeatableRead ensures the same data is read throughout the transaction.
	IsolationLevelRepeatableRead

	// IsolationLevelSerializable provides the highest isolation level.
	IsolationLevelSerializable
)
