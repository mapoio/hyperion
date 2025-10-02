package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/mapoio/hyperion"
)

// mockConfig implements hyperion.Config for testing.
type mockConfig struct {
	data map[string]any
}

func (m *mockConfig) Get(key string) any {
	return m.data[key]
}

func (m *mockConfig) GetString(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

func (m *mockConfig) GetInt(key string) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return 0
}

func (m *mockConfig) GetInt64(key string) int64 {
	if v, ok := m.data[key].(int64); ok {
		return v
	}
	if v, ok := m.data[key].(int); ok {
		return int64(v)
	}
	return 0
}

func (m *mockConfig) GetBool(key string) bool {
	if v, ok := m.data[key].(bool); ok {
		return v
	}
	return false
}

func (m *mockConfig) GetFloat64(key string) float64 {
	if v, ok := m.data[key].(float64); ok {
		return v
	}
	return 0
}

func (m *mockConfig) GetStringSlice(key string) []string {
	if v, ok := m.data[key].([]string); ok {
		return v
	}
	return nil
}

func (m *mockConfig) IsSet(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *mockConfig) AllKeys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *mockConfig) Unmarshal(key string, v any) error {
	cfg, ok := v.(*Config)
	if !ok {
		return nil
	}

	var dbData map[string]any
	if key == "" {
		// Root level - use entire data map directly
		dbData = m.data
	} else {
		// Specific key - extract nested map
		if data, ok := m.data[key].(map[string]any); ok {
			dbData = data
		}
	}

	if dbData != nil {
		if driver, ok := dbData["driver"].(string); ok {
			cfg.Driver = driver
		}
		if database, ok := dbData["database"].(string); ok {
			cfg.Database = database
		}
		if host, ok := dbData["host"].(string); ok {
			cfg.Host = host
		}
		if port, ok := dbData["port"].(int); ok {
			cfg.Port = port
		}
		if username, ok := dbData["username"].(string); ok {
			cfg.Username = username
		}
		if password, ok := dbData["password"].(string); ok {
			cfg.Password = password
		}
		if dsn, ok := dbData["dsn"].(string); ok {
			cfg.DSN = dsn
		}
		if sslmode, ok := dbData["sslmode"].(string); ok {
			cfg.SSLMode = sslmode
		}
		if charset, ok := dbData["charset"].(string); ok {
			cfg.Charset = charset
		}
		if maxOpenConns, ok := dbData["max_open_conns"].(int); ok {
			cfg.MaxOpenConns = maxOpenConns
		}
		if maxIdleConns, ok := dbData["max_idle_conns"].(int); ok {
			cfg.MaxIdleConns = maxIdleConns
		}
		if connMaxLifetime, ok := dbData["conn_max_lifetime"].(int64); ok {
			cfg.ConnMaxLifetime = time.Duration(connMaxLifetime)
		}
		if connMaxIdleTime, ok := dbData["conn_max_idle_time"].(int64); ok {
			cfg.ConnMaxIdleTime = time.Duration(connMaxIdleTime)
		}
		if slowThreshold, ok := dbData["slow_threshold"].(int64); ok {
			cfg.SlowThreshold = time.Duration(slowThreshold)
		}
		if logLevel, ok := dbData["log_level"].(string); ok {
			cfg.LogLevel = logLevel
		}
		if skipDefaultTransaction, ok := dbData["skip_default_transaction"].(bool); ok {
			cfg.SkipDefaultTransaction = skipDefaultTransaction
		}
		if prepareStmt, ok := dbData["prepare_stmt"].(bool); ok {
			cfg.PrepareStmt = prepareStmt
		}
		if autoMigrate, ok := dbData["auto_migrate"].(bool); ok {
			cfg.AutoMigrate = autoMigrate
		}
	}
	return nil
}

func newSQLiteConfig() hyperion.Config {
	return &mockConfig{
		data: map[string]any{
			"database": map[string]any{
				"driver":   DriverSQLite,
				"database": ":memory:",
			},
		},
	}
}

func TestGormDatabase_InterfaceCompliance(t *testing.T) {
	var _ hyperion.Database = (*gormDatabase)(nil)
}

func TestGormDatabase_Executor(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	executor := db.Executor()
	if executor == nil {
		t.Fatal("Executor() returned nil")
	}

	// Verify it's a gormExecutor
	gormExec, ok := executor.(*gormExecutor)
	if !ok {
		t.Fatal("Executor() did not return *gormExecutor")
	}

	if gormExec.isTx {
		t.Error("Executor() returned transaction executor, want non-transaction")
	}
}

func TestGormDatabase_Health(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Health(ctx); err != nil {
		t.Errorf("Health() error = %v, want nil", err)
	}
}

func TestGormDatabase_Close(t *testing.T) {
	cfg := newSQLiteConfig()
	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Health check should fail after close
	ctx := context.Background()
	if err := db.Health(ctx); err == nil {
		t.Error("Health() after Close() should return error")
	}
}

func TestNewGormDatabase_InvalidDriver(t *testing.T) {
	cfg := &mockConfig{
		data: map[string]any{
			"database": map[string]any{
				"driver":   "invalid",
				"database": "test.db",
			},
		},
	}

	_, err := NewGormDatabase(cfg)
	if err == nil {
		t.Error("NewGormDatabase() with invalid driver should return error")
	}
}

func TestNewGormDatabase_MissingDriver(t *testing.T) {
	// Test config validation with empty driver
	dbConfig := &Config{
		Driver:   "", // Empty driver
		Database: "test.db",
	}

	err := dbConfig.Validate()
	if err == nil {
		t.Error("Validate() without driver should return error")
	}
}
