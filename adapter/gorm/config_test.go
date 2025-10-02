package gorm

import (
	"testing"
	"time"

	"github.com/mapoio/hyperion"
)

// boolPtr returns a pointer to a boolean value.
// Helper function for tests to easily create boolean pointers.
func boolPtr(b bool) *bool {
	return &b
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Valid SQLite config",
			config: &Config{
				Driver:   DriverSQLite,
				Database: "test.db",
			},
			wantErr: false,
		},
		{
			name: "Valid PostgreSQL config",
			config: &Config{
				Driver:   DriverPostgres,
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Password: "pass",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "Valid MySQL config",
			config: &Config{
				Driver:   DriverMySQL,
				Host:     "localhost",
				Port:     3306,
				Username: "user",
				Password: "pass",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "Valid config with DSN",
			config: &Config{
				Driver: DriverPostgres,
				DSN:    "postgres://user:pass@localhost:5432/testdb",
			},
			wantErr: false,
		},
		{
			name:    "Missing driver",
			config:  &Config{},
			wantErr: true,
		},
		{
			name: "Invalid driver",
			config: &Config{
				Driver: "mssql",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL without database name",
			config: &Config{
				Driver: DriverPostgres,
				Host:   "localhost",
			},
			wantErr: true,
		},
		{
			name: "MySQL without database name",
			config: &Config{
				Driver: DriverMySQL,
				Host:   "localhost",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_BuildDSN(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name: "PostgreSQL DSN",
			config: &Config{
				Driver:   DriverPostgres,
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Password: "pass",
				Database: "testdb",
				SSLMode:  "disable",
			},
			want: "host=localhost port=5432 user=user password=pass dbname=testdb sslmode=disable",
		},
		{
			name: "MySQL DSN",
			config: &Config{
				Driver:   DriverMySQL,
				Host:     "localhost",
				Port:     3306,
				Username: "user",
				Password: "pass",
				Database: "testdb",
				Charset:  "utf8mb4",
			},
			want: "user:pass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "SQLite DSN",
			config: &Config{
				Driver:   DriverSQLite,
				Database: "test.db",
			},
			want: "test.db",
		},
		{
			name: "Explicit DSN overrides",
			config: &Config{
				Driver:   DriverPostgres,
				DSN:      "postgres://custom:dsn@host:1234/db",
				Host:     "ignored",
				Database: "ignored",
			},
			want: "postgres://custom:dsn@host:1234/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.buildDSN()
			if got != tt.want {
				t.Errorf("buildDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Open(t *testing.T) {
	// Test SQLite (can actually open)
	cfg := &Config{
		Driver:                 DriverSQLite,
		Database:               ":memory:",
		MaxOpenConns:           10,
		MaxIdleConns:           5,
		ConnMaxLifetime:        5 * time.Minute,
		ConnMaxIdleTime:        10 * time.Minute,
		LogLevel:               "silent",
		SlowThreshold:          200 * time.Millisecond,
		SkipDefaultTransaction: boolPtr(true),
		PrepareStmt:            boolPtr(true),
	}

	db, err := cfg.Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB() error = %v", err)
	}

	// Verify connection pool settings
	if sqlDB.Stats().MaxOpenConnections != 10 {
		t.Errorf("MaxOpenConnections = %d, want 10", sqlDB.Stats().MaxOpenConnections)
	}

	// Clean up
	sqlDB.Close()
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Driver != DriverSQLite {
		t.Errorf("DefaultConfig().Driver = %s, want sqlite", cfg.Driver)
	}

	if cfg.MaxOpenConns != 25 {
		t.Errorf("DefaultConfig().MaxOpenConns = %d, want 25", cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns != 5 {
		t.Errorf("DefaultConfig().MaxIdleConns = %d, want 5", cfg.MaxIdleConns)
	}

	if cfg.LogLevel != "warn" {
		t.Errorf("DefaultConfig().LogLevel = %s, want warn", cfg.LogLevel)
	}

	if cfg.PrepareStmt == nil || !*cfg.PrepareStmt {
		t.Error("DefaultConfig().PrepareStmt should be true")
	}
}

func TestConfig_GetLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
	}{
		{"Silent", "silent"},
		{"Error", "error"},
		{"Warn", "warn"},
		{"Info", "info"},
		{"Default", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{LogLevel: tt.logLevel}
			logger := cfg.getLogger()
			if logger == nil {
				t.Error("getLogger() returned nil")
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Test with database prefix
	cfg := &mockConfig{
		data: map[string]any{
			"database": map[string]any{
				"driver":         DriverPostgres,
				"host":           "testhost",
				"port":           5433,
				"username":       "testuser",
				"password":       "testpass",
				"database":       "testdb",
				"max_open_conns": 50,
				"max_idle_conns": 10,
			},
		},
	}

	dbConfig := DefaultConfig()
	err := loadConfig(cfg, dbConfig)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if dbConfig.Driver != DriverPostgres {
		t.Errorf("Driver = %s, want postgres", dbConfig.Driver)
	}

	if dbConfig.Host != "testhost" {
		t.Errorf("Host = %s, want testhost", dbConfig.Host)
	}

	if dbConfig.Port != 5433 {
		t.Errorf("Port = %d, want 5433", dbConfig.Port)
	}

	if dbConfig.MaxOpenConns != 50 {
		t.Errorf("MaxOpenConns = %d, want 50", dbConfig.MaxOpenConns)
	}
}

func TestNewGormDatabase(t *testing.T) {
	tests := []struct {
		name    string
		cfg     hyperion.Config
		wantErr bool
	}{
		{
			name: "Valid SQLite config",
			cfg: &mockConfig{
				data: map[string]any{
					"database": map[string]any{
						"driver":   DriverSQLite,
						"database": ":memory:",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid driver",
			cfg: &mockConfig{
				data: map[string]any{
					"database": map[string]any{
						"driver": "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing database for non-SQLite",
			cfg: &mockConfig{
				data: map[string]any{
					"database": map[string]any{
						"driver": DriverPostgres,
						"host":   "localhost",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewGormDatabase(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGormDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && db != nil {
				// Clean up
				if gdb, ok := db.(*gormDatabase); ok {
					if sqlDB, err := gdb.db.DB(); err == nil {
						sqlDB.Close()
					}
				}
			}
		})
	}
}

func TestNewGormUnitOfWork(t *testing.T) {
	// Test with valid gormDatabase
	cfg := &mockConfig{
		data: map[string]any{
			"database": map[string]any{
				"driver":   DriverSQLite,
				"database": ":memory:",
			},
		},
	}

	db, err := NewGormDatabase(cfg)
	if err != nil {
		t.Fatalf("NewGormDatabase() error = %v", err)
	}
	defer func() {
		if gdb, ok := db.(*gormDatabase); ok {
			if sqlDB, err := gdb.db.DB(); err == nil {
				sqlDB.Close()
			}
		}
	}()

	uow := NewGormUnitOfWork(db)
	if uow == nil {
		t.Error("NewGormUnitOfWork() returned nil")
	}
}

func TestNewGormUnitOfWork_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewGormUnitOfWork() should panic with invalid database type")
		}
	}()

	// Create a mock database that's not a *gormDatabase
	type mockDB struct {
		hyperion.Database
	}
	NewGormUnitOfWork(&mockDB{})
}

func TestLoadConfig_Fallback(t *testing.T) {
	// Test without database prefix - the Unmarshal will still look for "database" key
	// when key="" is provided
	cfg := &mockConfig{
		data: map[string]any{
			"database": map[string]any{
				"driver":         DriverSQLite,
				"database":       "test.db",
				"max_open_conns": 30,
			},
		},
	}

	dbConfig := &Config{
		Driver:       "mysql", // Different default to verify it gets overridden
		MaxOpenConns: 10,      // Different default to verify it gets overridden
	}
	err := loadConfig(cfg, dbConfig)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if dbConfig.Driver != DriverSQLite {
		t.Errorf("Driver = %s, want sqlite", dbConfig.Driver)
	}

	if dbConfig.MaxOpenConns != 30 {
		t.Errorf("MaxOpenConns = %d, want 30", dbConfig.MaxOpenConns)
	}
}

func TestLoadConfig_PreservesDefaults(t *testing.T) {
	// Test that defaults are preserved when config file has no database section
	// This addresses Codex P1 issue: Viper/mapstructure zeroes the destination struct
	cfg := &mockConfig{
		data: map[string]any{
			"app": map[string]any{
				"name": "myapp",
			},
		},
	}

	// Start with defaults
	dbConfig := DefaultConfig()

	// Load config (should preserve defaults since no database section)
	err := loadConfig(cfg, dbConfig)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	// Verify defaults are preserved
	if dbConfig.Driver != DriverSQLite {
		t.Errorf("Driver = %s, want %s", dbConfig.Driver, DriverSQLite)
	}

	if dbConfig.Port != 5432 {
		t.Errorf("Port = %d, want 5432", dbConfig.Port)
	}

	if dbConfig.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns = %d, want 25", dbConfig.MaxOpenConns)
	}

	if dbConfig.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", dbConfig.MaxIdleConns)
	}

	if dbConfig.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("ConnMaxLifetime = %v, want 5m", dbConfig.ConnMaxLifetime)
	}

	if dbConfig.ConnMaxIdleTime != 10*time.Minute {
		t.Errorf("ConnMaxIdleTime = %v, want 10m", dbConfig.ConnMaxIdleTime)
	}

	if dbConfig.LogLevel != "warn" {
		t.Errorf("LogLevel = %s, want warn", dbConfig.LogLevel)
	}

	if dbConfig.PrepareStmt == nil || !*dbConfig.PrepareStmt {
		t.Errorf("PrepareStmt = %v, want true", dbConfig.PrepareStmt)
	}

	if dbConfig.Database != "hyperion.db" {
		t.Errorf("Database = %s, want hyperion.db", dbConfig.Database)
	}
}

func TestLoadConfig_PartialOverride(t *testing.T) {
	// Test that only provided values are overridden, rest preserved
	cfg := &mockConfig{
		data: map[string]any{
			"database": map[string]any{
				"driver":         DriverPostgres,
				"host":           "customhost",
				"max_open_conns": 100,
				// Other fields not provided - should keep defaults
			},
		},
	}

	// Start with defaults
	dbConfig := DefaultConfig()

	// Load config
	err := loadConfig(cfg, dbConfig)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	// Verify overridden values
	if dbConfig.Driver != DriverPostgres {
		t.Errorf("Driver = %s, want %s", dbConfig.Driver, DriverPostgres)
	}

	if dbConfig.Host != "customhost" {
		t.Errorf("Host = %s, want customhost", dbConfig.Host)
	}

	if dbConfig.MaxOpenConns != 100 {
		t.Errorf("MaxOpenConns = %d, want 100", dbConfig.MaxOpenConns)
	}

	// Verify preserved defaults
	if dbConfig.Port != 5432 {
		t.Errorf("Port = %d, want 5432 (default)", dbConfig.Port)
	}

	if dbConfig.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5 (default)", dbConfig.MaxIdleConns)
	}

	if dbConfig.LogLevel != "warn" {
		t.Errorf("LogLevel = %s, want warn (default)", dbConfig.LogLevel)
	}
}

func TestLoadConfig_RootLevel(t *testing.T) {
	// Test root-level config (without "database" prefix)
	// This addresses Codex P1 issue: root-level config was being ignored
	cfg := &mockConfig{
		data: map[string]any{
			"driver":         DriverMySQL,
			"host":           "roothost",
			"port":           3306,
			"username":       "rootuser",
			"password":       "rootpass",
			"database":       "rootdb",
			"max_open_conns": 50,
		},
	}

	// Start with defaults
	dbConfig := DefaultConfig()

	// Load config
	err := loadConfig(cfg, dbConfig)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	// Verify root-level values are loaded
	if dbConfig.Driver != DriverMySQL {
		t.Errorf("Driver = %s, want %s", dbConfig.Driver, DriverMySQL)
	}

	if dbConfig.Host != "roothost" {
		t.Errorf("Host = %s, want roothost", dbConfig.Host)
	}

	if dbConfig.Port != 3306 {
		t.Errorf("Port = %d, want 3306", dbConfig.Port)
	}

	if dbConfig.Username != "rootuser" {
		t.Errorf("Username = %s, want rootuser", dbConfig.Username)
	}

	if dbConfig.Password != "rootpass" {
		t.Errorf("Password = %s, want rootpass", dbConfig.Password)
	}

	if dbConfig.Database != "rootdb" {
		t.Errorf("Database = %s, want rootdb", dbConfig.Database)
	}

	if dbConfig.MaxOpenConns != 50 {
		t.Errorf("MaxOpenConns = %d, want 50", dbConfig.MaxOpenConns)
	}

	// Verify defaults are preserved for unset fields
	if dbConfig.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5 (default)", dbConfig.MaxIdleConns)
	}
}

func TestLoadConfig_RootOverridesPrefixed(t *testing.T) {
	// Test that root-level config takes precedence over prefixed config
	cfg := &mockConfig{
		data: map[string]any{
			"database": map[string]any{
				"driver":                   DriverSQLite,
				"host":                     "prefixedhost",
				"port":                     5432,
				"username":                 "prefixuser",
				"password":                 "prefixpass",
				"database":                 "prefixdb",
				"dsn":                      "prefixdsn",
				"sslmode":                  "disable",
				"charset":                  "utf8",
				"max_open_conns":           10,
				"max_idle_conns":           2,
				"conn_max_lifetime":        int64(1 * time.Second),
				"conn_max_idle_time":       int64(2 * time.Second),
				"slow_threshold":           int64(100 * time.Millisecond),
				"log_level":                "error",
				"skip_default_transaction": true,
				"prepare_stmt":             false,
				"auto_migrate":             false,
			},
			// Root-level values should override prefixed ones
			"driver":   DriverMySQL,
			"port":     3306,
			"username": "rootuser",
			"charset":  "utf8mb4",
		},
	}

	// Start with defaults
	dbConfig := DefaultConfig()

	// Load config
	err := loadConfig(cfg, dbConfig)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	// Verify root-level values take precedence
	if dbConfig.Driver != DriverMySQL {
		t.Errorf("Driver = %s, want %s (root should override prefixed)", dbConfig.Driver, DriverMySQL)
	}

	if dbConfig.Port != 3306 {
		t.Errorf("Port = %d, want 3306 (root should override prefixed)", dbConfig.Port)
	}

	if dbConfig.Username != "rootuser" {
		t.Errorf("Username = %s, want rootuser (root should override prefixed)", dbConfig.Username)
	}

	if dbConfig.Charset != "utf8mb4" {
		t.Errorf("Charset = %s, want utf8mb4 (root should override prefixed)", dbConfig.Charset)
	}

	// Verify prefixed values are used when root doesn't override
	if dbConfig.Host != "prefixedhost" {
		t.Errorf("Host = %s, want prefixedhost (from prefixed)", dbConfig.Host)
	}

	if dbConfig.Password != "prefixpass" {
		t.Errorf("Password = %s, want prefixpass (from prefixed)", dbConfig.Password)
	}

	if dbConfig.Database != "prefixdb" {
		t.Errorf("Database = %s, want prefixdb (from prefixed)", dbConfig.Database)
	}

	if dbConfig.DSN != "prefixdsn" {
		t.Errorf("DSN = %s, want prefixdsn (from prefixed)", dbConfig.DSN)
	}

	if dbConfig.SSLMode != "disable" {
		t.Errorf("SSLMode = %s, want disable (from prefixed)", dbConfig.SSLMode)
	}

	if dbConfig.MaxOpenConns != 10 {
		t.Errorf("MaxOpenConns = %d, want 10 (from prefixed)", dbConfig.MaxOpenConns)
	}

	if dbConfig.MaxIdleConns != 2 {
		t.Errorf("MaxIdleConns = %d, want 2 (from prefixed)", dbConfig.MaxIdleConns)
	}

	if dbConfig.ConnMaxLifetime != 1000000000 {
		t.Errorf("ConnMaxLifetime = %v, want 1s (from prefixed)", dbConfig.ConnMaxLifetime)
	}

	if dbConfig.ConnMaxIdleTime != 2000000000 {
		t.Errorf("ConnMaxIdleTime = %v, want 2s (from prefixed)", dbConfig.ConnMaxIdleTime)
	}

	if dbConfig.SlowThreshold != 100000000 {
		t.Errorf("SlowThreshold = %v, want 100ms (from prefixed)", dbConfig.SlowThreshold)
	}

	if dbConfig.LogLevel != "error" {
		t.Errorf("LogLevel = %s, want error (from prefixed)", dbConfig.LogLevel)
	}

	if dbConfig.SkipDefaultTransaction == nil || !*dbConfig.SkipDefaultTransaction {
		t.Errorf("SkipDefaultTransaction = %v, want true (from prefixed)", dbConfig.SkipDefaultTransaction)
	}
}

func TestLoadConfig_BooleanFalseValues(t *testing.T) {
	// Test that boolean fields can be explicitly set to false
	// This addresses Codex P1 issue: Boolean overrides ignore explicit false values
	cfg := &mockConfig{
		data: map[string]any{
			"database": map[string]any{
				"driver":                   DriverSQLite,
				"database":                 ":memory:",
				"prepare_stmt":             false, // Explicitly set to false (default is true)
				"skip_default_transaction": false, // Explicitly set to false
				"auto_migrate":             false, // Explicitly set to false
			},
		},
	}

	dbConfig := DefaultConfig()
	// Verify defaults before loading
	if dbConfig.PrepareStmt == nil || !*dbConfig.PrepareStmt {
		t.Fatal("Default PrepareStmt should be true")
	}

	err := loadConfig(cfg, dbConfig)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	// Verify that explicit false values override defaults
	if dbConfig.PrepareStmt == nil || *dbConfig.PrepareStmt {
		t.Errorf("PrepareStmt = %v, want false (explicitly set)", dbConfig.PrepareStmt)
	}

	if dbConfig.SkipDefaultTransaction == nil || *dbConfig.SkipDefaultTransaction {
		t.Errorf("SkipDefaultTransaction = %v, want false (explicitly set)", dbConfig.SkipDefaultTransaction)
	}

	if dbConfig.AutoMigrate == nil || *dbConfig.AutoMigrate {
		t.Errorf("AutoMigrate = %v, want false (explicitly set)", dbConfig.AutoMigrate)
	}
}

func TestLoadConfig_BooleanRootLevelFalse(t *testing.T) {
	// Test that root-level boolean fields can be set to false
	cfg := &mockConfig{
		data: map[string]any{
			"driver":       DriverSQLite,
			"database":     ":memory:",
			"prepare_stmt": false, // Root level explicit false
		},
	}

	dbConfig := DefaultConfig()
	err := loadConfig(cfg, dbConfig)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if dbConfig.PrepareStmt == nil || *dbConfig.PrepareStmt {
		t.Errorf("PrepareStmt = %v, want false (root level explicit)", dbConfig.PrepareStmt)
	}
}
