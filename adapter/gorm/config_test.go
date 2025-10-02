package gorm

import (
	"testing"
	"time"

	"github.com/mapoio/hyperion"
)

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
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
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

	if cfg.PrepareStmt != true {
		t.Error("DefaultConfig().PrepareStmt = false, want true")
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
