//go:build integration
// +build integration

package gorm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mapoio/hyperion"
)

// TestIntegration_SQLite tests the complete flow with SQLite in-memory database.
func TestIntegration_SQLite(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	// Test health check
	ctx := context.Background()
	if err := db.Health(ctx); err != nil {
		t.Fatalf("Health() error = %v", err)
	}

	// Setup test table
	executor := db.Executor()
	createTableSQL := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if err := executor.Exec(ctx, createTableSQL); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test CRUD operations
	t.Run("Insert", func(t *testing.T) {
		err := executor.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Alice", "alice@example.com")
		if err != nil {
			t.Errorf("Insert failed: %v", err)
		}
	})

	t.Run("Query", func(t *testing.T) {
		type User struct {
			ID    int64
			Name  string
			Email string
		}

		var users []User
		err := executor.Query(ctx, &users, "SELECT id, name, email FROM users WHERE name = ?", "Alice")
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Email != "alice@example.com" {
			t.Errorf("Expected alice@example.com, got %s", users[0].Email)
		}
	})
}

// TestIntegration_Transaction tests transaction commit and rollback.
func TestIntegration_Transaction(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	ctx := context.Background()

	// Setup table
	executor.Exec(ctx, `CREATE TABLE accounts (id INTEGER PRIMARY KEY, balance INTEGER)`)
	executor.Exec(ctx, "INSERT INTO accounts (id, balance) VALUES (1, 1000), (2, 500)")

	uow := NewGormUnitOfWork(db)
	hctx := newTestContext(executor)

	t.Run("Successful transaction", func(t *testing.T) {
		err := uow.WithTransaction(hctx, func(txCtx hyperion.Context) error {
			// Transfer 200 from account 1 to account 2
			txDB := txCtx.DB()
			if err := txDB.Exec(txCtx, "UPDATE accounts SET balance = balance - 200 WHERE id = 1"); err != nil {
				return err
			}
			if err := txDB.Exec(txCtx, "UPDATE accounts SET balance = balance + 200 WHERE id = 2"); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			t.Errorf("Transaction failed: %v", err)
		}

		// Verify balances
		var balance int64
		executor.Query(ctx, &balance, "SELECT balance FROM accounts WHERE id = 1")
		if balance != 800 {
			t.Errorf("Account 1 balance = %d, want 800", balance)
		}

		executor.Query(ctx, &balance, "SELECT balance FROM accounts WHERE id = 2")
		if balance != 700 {
			t.Errorf("Account 2 balance = %d, want 700", balance)
		}
	})

	t.Run("Failed transaction rollback", func(t *testing.T) {
		// Reset balances
		executor.Exec(ctx, "UPDATE accounts SET balance = 1000 WHERE id = 1")
		executor.Exec(ctx, "UPDATE accounts SET balance = 500 WHERE id = 2")

		err := uow.WithTransaction(hctx, func(txCtx hyperion.Context) error {
			txDB := txCtx.DB()
			if err := txDB.Exec(txCtx, "UPDATE accounts SET balance = balance - 200 WHERE id = 1"); err != nil {
				return err
			}
			// Simulate error before second update
			return context.DeadlineExceeded
		})

		if err == nil {
			t.Error("Expected error, got nil")
		}

		// Verify balances were rolled back
		var balance int64
		executor.Query(ctx, &balance, "SELECT balance FROM accounts WHERE id = 1")
		if balance != 1000 {
			t.Errorf("Account 1 balance after rollback = %d, want 1000", balance)
		}

		executor.Query(ctx, &balance, "SELECT balance FROM accounts WHERE id = 2")
		if balance != 500 {
			t.Errorf("Account 2 balance after rollback = %d, want 500", balance)
		}
	})
}

// TestIntegration_IsolationLevels tests different transaction isolation levels.
func TestIntegration_IsolationLevels(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	ctx := context.Background()
	executor.Exec(ctx, `CREATE TABLE counters (id INTEGER PRIMARY KEY, value INTEGER)`)
	executor.Exec(ctx, "INSERT INTO counters (id, value) VALUES (1, 0)")

	uow := NewGormUnitOfWork(db)
	hctx := newTestContext(executor)

	levels := []hyperion.IsolationLevel{
		hyperion.IsolationLevelDefault,
		hyperion.IsolationLevelReadCommitted,
		hyperion.IsolationLevelSerializable,
	}

	for _, level := range levels {
		t.Run(fmt.Sprintf("IsolationLevel-%d", level), func(t *testing.T) {
			opts := &hyperion.TransactionOptions{
				Isolation: level,
				ReadOnly:  false,
			}

			err := uow.WithTransactionOptions(hctx, opts, func(txCtx hyperion.Context) error {
				txDB := txCtx.DB()
				return txDB.Exec(txCtx, "UPDATE counters SET value = value + 1 WHERE id = 1")
			})

			if err != nil {
				t.Errorf("Transaction with isolation level %d failed: %v", level, err)
			}
		})
	}

	// Verify all updates succeeded
	var value int64
	executor.Query(ctx, &value, "SELECT value FROM counters WHERE id = 1")
	if value != int64(len(levels)) {
		t.Errorf("Counter value = %d, want %d", value, len(levels))
	}
}

// TestIntegration_ConcurrentTransactions tests concurrent transaction execution.
func TestIntegration_ConcurrentTransactions(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	ctx := context.Background()
	executor.Exec(ctx, `CREATE TABLE counters (id INTEGER PRIMARY KEY, value INTEGER)`)
	executor.Exec(ctx, "INSERT INTO counters (id, value) VALUES (1, 0)")

	uow := NewGormUnitOfWork(db)
	hctx := newTestContext(executor)

	// Run 10 concurrent transactions
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			err := uow.WithTransaction(hctx, func(txCtx hyperion.Context) error {
				txDB := txCtx.DB()
				// Small sleep to increase chance of concurrent execution
				time.Sleep(10 * time.Millisecond)
				return txDB.Exec(txCtx, "UPDATE counters SET value = value + 1 WHERE id = 1")
			})
			if err != nil {
				t.Errorf("Concurrent transaction failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all transactions
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// Verify final value
	var value int64
	executor.Query(ctx, &value, "SELECT value FROM counters WHERE id = 1")
	if value != int64(concurrency) {
		t.Errorf("Counter value = %d, want %d", value, concurrency)
	}
}

// TestIntegration_ConnectionPool tests connection pool settings.
func TestIntegration_ConnectionPool(t *testing.T) {
	cfg := &mockConfig{
		data: map[string]any{
			"database": map[string]any{
				"driver":             DriverSQLite,
				"database":           ":memory:",
				"max_open_conns":     5,
				"max_idle_conns":     2,
				"conn_max_lifetime":  "1m",
				"conn_max_idle_time": "30s",
			},
		},
	}

	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	// Get underlying sql.DB to check pool stats
	gdb := db.(*gormDatabase)
	sqlDB, err := gdb.db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 5 {
		t.Errorf("MaxOpenConnections = %d, want 5", stats.MaxOpenConnections)
	}
}

// TestIntegration_NestedTransactions tests nested transaction behavior with savepoints.
func TestIntegration_NestedTransactions(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	ctx := context.Background()
	executor.Exec(ctx, `CREATE TABLE logs (id INTEGER PRIMARY KEY AUTOINCREMENT, message TEXT)`)

	uow := NewGormUnitOfWork(db)
	hctx := newTestContext(executor)

	err = uow.WithTransaction(hctx, func(txCtx1 hyperion.Context) error {
		txDB1 := txCtx1.DB()
		if err := txDB1.Exec(txCtx1, "INSERT INTO logs (message) VALUES (?)", "outer-1"); err != nil {
			return err
		}

		// Nested transaction (savepoint)
		err := uow.WithTransaction(txCtx1, func(txCtx2 hyperion.Context) error {
			txDB2 := txCtx2.DB()
			if err := txDB2.Exec(txCtx2, "INSERT INTO logs (message) VALUES (?)", "inner-1"); err != nil {
				return err
			}
			if err := txDB2.Exec(txCtx2, "INSERT INTO logs (message) VALUES (?)", "inner-2"); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		if err := txDB1.Exec(txCtx1, "INSERT INTO logs (message) VALUES (?)", "outer-2"); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Nested transaction failed: %v", err)
	}

	// Verify all logs were inserted
	var count int64
	executor.Query(ctx, &count, "SELECT COUNT(*) FROM logs")
	if count != 4 {
		t.Errorf("Log count = %d, want 4", count)
	}
}
