package gorm

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"github.com/mapoio/hyperion"
)

func TestGormExecutor_InterfaceCompliance(t *testing.T) {
	var _ hyperion.Executor = (*gormExecutor)(nil)
}

func TestGormExecutor_Exec(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	ctx := context.Background()

	// Create table
	sql := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL
	)`

	if err := executor.Exec(ctx, sql); err != nil {
		t.Errorf("Exec() error = %v, want nil", err)
	}

	// Insert data
	insertSQL := "INSERT INTO users (name, email) VALUES (?, ?)"
	if err := executor.Exec(ctx, insertSQL, "John Doe", "john@example.com"); err != nil {
		t.Errorf("Exec() insert error = %v, want nil", err)
	}
}

func TestGormExecutor_Query(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	ctx := context.Background()

	// Setup table and data
	executor.Exec(ctx, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)`)
	executor.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Alice", "alice@example.com")
	executor.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Bob", "bob@example.com")

	// Query single value
	var count int64
	if err := executor.Query(ctx, &count, "SELECT COUNT(*) FROM users"); err != nil {
		t.Errorf("Query() error = %v, want nil", err)
	}

	if count != 2 {
		t.Errorf("Query() count = %d, want 2", count)
	}

	// Query multiple rows
	type User struct {
		ID    int64
		Name  string
		Email string
	}

	var users []User
	if err := executor.Query(ctx, &users, "SELECT id, name, email FROM users ORDER BY id"); err != nil {
		t.Errorf("Query() error = %v, want nil", err)
	}

	if len(users) != 2 {
		t.Errorf("Query() len(users) = %d, want 2", len(users))
	}

	if users[0].Name != "Alice" {
		t.Errorf("Query() users[0].Name = %s, want Alice", users[0].Name)
	}
}

func TestGormExecutor_Begin(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	ctx := context.Background()

	// Begin transaction
	txExecutor, err := executor.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v, want nil", err)
	}

	// Verify it's a transaction executor
	gormTx, ok := txExecutor.(*gormExecutor)
	if !ok {
		t.Fatal("Begin() did not return *gormExecutor")
	}

	if !gormTx.isTx {
		t.Error("Begin() returned non-transaction executor, want transaction")
	}
}

func TestGormExecutor_Commit(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	ctx := context.Background()

	// Setup table
	executor.Exec(ctx, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)

	// Begin transaction
	txExecutor, err := executor.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	// Insert in transaction
	if err := txExecutor.Exec(ctx, "INSERT INTO users (name) VALUES (?)", "Alice"); err != nil {
		t.Errorf("Exec() in transaction error = %v", err)
	}

	// Commit
	if err := txExecutor.Commit(); err != nil {
		t.Errorf("Commit() error = %v, want nil", err)
	}

	// Verify data was committed
	var count int64
	executor.Query(ctx, &count, "SELECT COUNT(*) FROM users")
	if count != 1 {
		t.Errorf("After commit, count = %d, want 1", count)
	}
}

func TestGormExecutor_Rollback(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	ctx := context.Background()

	// Setup table
	executor.Exec(ctx, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)

	// Begin transaction
	txExecutor, err := executor.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	// Insert in transaction
	if err := txExecutor.Exec(ctx, "INSERT INTO users (name) VALUES (?)", "Bob"); err != nil {
		t.Errorf("Exec() in transaction error = %v", err)
	}

	// Rollback
	if err := txExecutor.Rollback(); err != nil {
		t.Errorf("Rollback() error = %v, want nil", err)
	}

	// Verify data was rolled back
	var count int64
	executor.Query(ctx, &count, "SELECT COUNT(*) FROM users")
	if count != 0 {
		t.Errorf("After rollback, count = %d, want 0", count)
	}
}

func TestGormExecutor_CommitNonTransaction(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()

	// Try to commit non-transaction executor
	err = executor.Commit()
	if err == nil {
		t.Error("Commit() on non-transaction executor should return error")
	}
}

func TestGormExecutor_RollbackNonTransaction(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()

	// Try to rollback non-transaction executor
	err = executor.Rollback()
	if err == nil {
		t.Error("Rollback() on non-transaction executor should return error")
	}
}

func TestGormExecutor_Unwrap(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	unwrapped := executor.Unwrap()

	if unwrapped == nil {
		t.Error("Unwrap() returned nil")
	}

	// Verify it's a *gorm.DB
	if _, ok := unwrapped.(*gorm.DB); !ok {
		t.Error("Unwrap() did not return *gorm.DB")
	}
}
