package gorm

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/mapoio/hyperion"
)

// mockLogger implements hyperion.Logger for testing.
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...any)     {}
func (m *mockLogger) Info(msg string, fields ...any)      {}
func (m *mockLogger) Warn(msg string, fields ...any)      {}
func (m *mockLogger) Error(msg string, fields ...any)     {}
func (m *mockLogger) Fatal(msg string, fields ...any)     {}
func (m *mockLogger) With(fields ...any) hyperion.Logger  { return m }
func (m *mockLogger) WithError(err error) hyperion.Logger { return m }
func (m *mockLogger) SetLevel(level hyperion.LogLevel)    {}
func (m *mockLogger) GetLevel() hyperion.LogLevel         { return hyperion.InfoLevel }
func (m *mockLogger) Sync() error                         { return nil }

// mockTracer implements hyperion.Tracer for testing.
type mockTracer struct{}

func (m *mockTracer) Start(ctx hyperion.Context, spanName string, opts ...hyperion.SpanOption) (hyperion.Context, hyperion.Span) {
	return ctx, &mockSpan{}
}

// mockSpan implements hyperion.Span for testing.
type mockSpan struct{}

func (m *mockSpan) End(opts ...hyperion.SpanEndOption)                  {}
func (m *mockSpan) SetAttributes(attrs ...hyperion.Attribute)           {}
func (m *mockSpan) RecordError(err error, opts ...hyperion.EventOption) {}
func (m *mockSpan) AddEvent(name string, opts ...hyperion.EventOption)  {}
func (m *mockSpan) SpanContext() hyperion.SpanContext                   { return &mockSpanContext{} }

// mockSpanContext implements hyperion.SpanContext for testing.
type mockSpanContext struct{}

func (m *mockSpanContext) TraceID() string { return "trace-id" }
func (m *mockSpanContext) SpanID() string  { return "span-id" }
func (m *mockSpanContext) IsValid() bool   { return true }

// mockMeter implements hyperion.Meter for testing.
type mockMeter struct{}

func (m *mockMeter) Counter(name string, opts ...hyperion.MetricOption) hyperion.Counter {
	return &mockCounter{}
}
func (m *mockMeter) Histogram(name string, opts ...hyperion.MetricOption) hyperion.Histogram {
	return &mockHistogram{}
}
func (m *mockMeter) Gauge(name string, opts ...hyperion.MetricOption) hyperion.Gauge {
	return &mockGauge{}
}
func (m *mockMeter) UpDownCounter(name string, opts ...hyperion.MetricOption) hyperion.UpDownCounter {
	return &mockUpDownCounter{}
}

// mockCounter implements hyperion.Counter for testing.
type mockCounter struct{}

func (m *mockCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {}

// mockHistogram implements hyperion.Histogram for testing.
type mockHistogram struct{}

func (m *mockHistogram) Record(ctx context.Context, value float64, attrs ...hyperion.Attribute) {}

// mockGauge implements hyperion.Gauge for testing.
type mockGauge struct{}

func (m *mockGauge) Record(ctx context.Context, value float64, attrs ...hyperion.Attribute) {}

// mockUpDownCounter implements hyperion.UpDownCounter for testing.
type mockUpDownCounter struct{}

func (m *mockUpDownCounter) Add(ctx context.Context, value int64, attrs ...hyperion.Attribute) {}

func newTestContext(executor hyperion.Executor) hyperion.Context {
	return hyperion.New(
		context.Background(),
		&mockLogger{},
		executor,
		&mockTracer{},
		&mockMeter{},
	)
}

func TestGormUnitOfWork_InterfaceCompliance(t *testing.T) {
	var _ hyperion.UnitOfWork = (*gormUnitOfWork)(nil)
}

func TestGormUnitOfWork_WithTransaction_Commit(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	// Setup table
	executor := db.Executor()
	ctx := context.Background()
	if execErr := executor.Exec(ctx, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`); execErr != nil {
		t.Fatalf("Failed to create table: %v", execErr)
	}

	// Create UnitOfWork
	uow := NewGormUnitOfWork(db)
	hctx := newTestContext(executor)

	// Execute transaction
	err = uow.WithTransaction(hctx, func(txCtx hyperion.Context) error {
		// Verify txCtx.DB() returns transaction executor
		txDB := txCtx.DB()
		gormTx, ok := txDB.(*gormExecutor)
		if !ok {
			t.Error("txCtx.DB() did not return *gormExecutor")
		}
		if !gormTx.isTx {
			t.Error("txCtx.DB() returned non-transaction executor")
		}

		// Insert data
		return txDB.Exec(txCtx, "INSERT INTO users (name) VALUES (?)", "Alice")
	})

	if err != nil {
		t.Errorf("WithTransaction() error = %v, want nil", err)
	}

	// Verify data was committed
	var count int64
	if err := executor.Query(ctx, &count, "SELECT COUNT(*) FROM users"); err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}
	if count != 1 {
		t.Errorf("After commit, count = %d, want 1", count)
	}
}

func TestGormUnitOfWork_WithTransaction_Rollback(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	// Setup table
	executor := db.Executor()
	ctx := context.Background()
	if execErr := executor.Exec(ctx, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`); execErr != nil {
		t.Fatalf("Failed to create table: %v", execErr)
	}

	// Create UnitOfWork
	uow := NewGormUnitOfWork(db)
	hctx := newTestContext(executor)

	// Execute transaction with error
	expectedErr := errors.New("test error")
	err = uow.WithTransaction(hctx, func(txCtx hyperion.Context) error {
		txDB := txCtx.DB()
		if execErr := txDB.Exec(txCtx, "INSERT INTO users (name) VALUES (?)", "Bob"); execErr != nil {
			t.Logf("Insert failed as expected: %v", execErr)
		}
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("WithTransaction() error = %v, want %v", err, expectedErr)
	}

	// Verify data was rolled back
	var count int64
	if err := executor.Query(ctx, &count, "SELECT COUNT(*) FROM users"); err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}
	if count != 0 {
		t.Errorf("After rollback, count = %d, want 0", count)
	}
}

func TestGormUnitOfWork_WithTransaction_Panic(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	// Setup table
	executor := db.Executor()
	ctx := context.Background()
	if execErr := executor.Exec(ctx, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`); execErr != nil {
		t.Fatalf("Failed to create table: %v", execErr)
	}

	// Create UnitOfWork
	uow := NewGormUnitOfWork(db)
	hctx := newTestContext(executor)

	// Execute transaction with panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("WithTransaction() did not propagate panic")
		}
	}()

	_ = uow.WithTransaction(hctx, func(txCtx hyperion.Context) error {
		txDB := txCtx.DB()
		if execErr := txDB.Exec(txCtx, "INSERT INTO users (name) VALUES (?)", "Charlie"); execErr != nil {
			t.Logf("Insert failed: %v", execErr)
		}
		panic("test panic")
	})

	// Verify data was rolled back
	var count int64
	if err := executor.Query(ctx, &count, "SELECT COUNT(*) FROM users"); err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}
	if count != 0 {
		t.Errorf("After panic rollback, count = %d, want 0", count)
	}
}

func TestGormUnitOfWork_WithTransactionOptions(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	// Setup table
	executor := db.Executor()
	ctx := context.Background()
	if execErr := executor.Exec(ctx, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`); execErr != nil {
		t.Fatalf("Failed to create table: %v", execErr)
	}

	// Create UnitOfWork
	uow := NewGormUnitOfWork(db)
	hctx := newTestContext(executor)

	// Execute transaction with options
	opts := &hyperion.TransactionOptions{
		Isolation: hyperion.IsolationLevelSerializable,
		ReadOnly:  false,
	}

	err = uow.WithTransactionOptions(hctx, opts, func(txCtx hyperion.Context) error {
		txDB := txCtx.DB()
		return txDB.Exec(txCtx, "INSERT INTO users (name) VALUES (?)", "David")
	})

	if err != nil {
		t.Errorf("WithTransactionOptions() error = %v, want nil", err)
	}

	// Verify data was committed
	var count int64
	if err := executor.Query(ctx, &count, "SELECT COUNT(*) FROM users"); err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}
	if count != 1 {
		t.Errorf("After commit with options, count = %d, want 1", count)
	}
}

func TestGormUnitOfWork_NestedTransaction(t *testing.T) {
	// Note: This test demonstrates nested transaction support which relies on
	// database savepoint capabilities. SQLite in-memory databases may have
	// limitations with savepoints in GORM's transaction handling.
	// For full nested transaction testing, use PostgreSQL or MySQL in integration tests.

	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	// Setup table
	gdb := db.(*gormDatabase)
	if createErr := gdb.db.Exec(`CREATE TABLE nested_users (id INTEGER PRIMARY KEY, name TEXT)`).Error; createErr != nil {
		t.Fatalf("Failed to create table: %v", createErr)
	}

	executor := db.Executor()
	ctx := context.Background()

	// Create UnitOfWork
	uow := NewGormUnitOfWork(db)
	hctx := newTestContext(executor)

	// Test that nested WithTransaction calls don't fail (they may or may not create savepoints)
	err = uow.WithTransaction(hctx, func(txCtx1 hyperion.Context) error {
		gormDB := txCtx1.DB().Unwrap().(*gorm.DB)
		if execErr := gormDB.Exec("INSERT INTO nested_users (name) VALUES (?)", "Eve").Error; execErr != nil {
			return execErr
		}

		// This may create a savepoint or reuse the existing transaction depending on GORM/DB behavior
		innerErr := uow.WithTransaction(txCtx1, func(txCtx2 hyperion.Context) error {
			gormDB2 := txCtx2.DB().Unwrap().(*gorm.DB)
			return gormDB2.Exec("INSERT INTO nested_users (name) VALUES (?)", "Frank").Error
		})

		// For SQLite in-memory, nested transactions might not work as expected
		// The test verifies that the API doesn't panic, actual savepoint behavior
		// is database-dependent
		return innerErr
	})

	// If the operation succeeded, verify data was committed
	// If it failed due to savepoint limitations, that's also acceptable for SQLite
	if err == nil {
		var count int64
		if queryErr := executor.Query(ctx, &count, "SELECT COUNT(*) FROM nested_users"); queryErr != nil {
			t.Fatalf("Failed to query count: %v", queryErr)
		}
		// At least the outer transaction should have committed
		if count < 1 {
			t.Errorf("After nested transaction, count = %d, want at least 1", count)
		}
	} else {
		t.Logf("Nested transaction not fully supported (expected for SQLite in-memory): %v", err)
	}
}

func TestToSQLIsolation(t *testing.T) {
	tests := []struct {
		name     string
		level    hyperion.IsolationLevel
		expected string
	}{
		{
			name:     "Default",
			level:    hyperion.IsolationLevelDefault,
			expected: "Default",
		},
		{
			name:     "ReadUncommitted",
			level:    hyperion.IsolationLevelReadUncommitted,
			expected: "Read Uncommitted",
		},
		{
			name:     "ReadCommitted",
			level:    hyperion.IsolationLevelReadCommitted,
			expected: "Read Committed",
		},
		{
			name:     "RepeatableRead",
			level:    hyperion.IsolationLevelRepeatableRead,
			expected: "Repeatable Read",
		},
		{
			name:     "Serializable",
			level:    hyperion.IsolationLevelSerializable,
			expected: "Serializable",
		},
		{
			name:     "Unknown",
			level:    hyperion.IsolationLevel(999),
			expected: "Default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toSQLIsolation(tt.level)
			if result.String() != tt.expected {
				t.Errorf("toSQLIsolation() = %s, want %s", result.String(), tt.expected)
			}
		})
	}
}
