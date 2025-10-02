package gorm

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/mapoio/hyperion"
)

const (
	// DriverPostgres represents the PostgreSQL driver.
	DriverPostgres = "postgres"
	// DriverMySQL represents the MySQL driver.
	DriverMySQL = "mysql"
	// DriverSQLite represents the SQLite driver.
	DriverSQLite = "sqlite"
)

// Config represents the database configuration.
// Fields are ordered for optimal memory alignment (larger types first).
type Config struct {
	// String fields (16 bytes on 64-bit: 8-byte pointer + 8-byte length)
	Driver   string `json:"driver" yaml:"driver"`       // Driver specifies the database driver (postgres, mysql, sqlite)
	DSN      string `json:"dsn" yaml:"dsn"`             // DSN allows providing a complete connection string
	Host     string `json:"host" yaml:"host"`           // Connection host
	Username string `json:"username" yaml:"username"`   // Connection username
	Password string `json:"password" yaml:"password"`   // Connection password
	Database string `json:"database" yaml:"database"`   // Database name
	SSLMode  string `json:"sslmode" yaml:"sslmode"`     // PostgreSQL SSL mode
	Charset  string `json:"charset" yaml:"charset"`     // MySQL charset
	LogLevel string `json:"log_level" yaml:"log_level"` // Log level: silent, error, warn, info

	// Duration fields (8 bytes each)
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	SlowThreshold   time.Duration `json:"slow_threshold" yaml:"slow_threshold"`

	// Int fields (8 bytes on 64-bit)
	MaxOpenConns int `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns int `json:"max_idle_conns" yaml:"max_idle_conns"`
	Port         int `json:"port" yaml:"port"`

	// Bool fields (1 byte each) - put last to minimize padding
	SkipDefaultTransaction bool `json:"skip_default_transaction" yaml:"skip_default_transaction"`
	PrepareStmt            bool `json:"prepare_stmt" yaml:"prepare_stmt"`
	AutoMigrate            bool `json:"auto_migrate" yaml:"auto_migrate"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Driver:                 DriverSQLite,
		Database:               "hyperion.db",
		Host:                   "localhost",
		Port:                   5432,
		SSLMode:                "disable",
		Charset:                "utf8mb4",
		MaxOpenConns:           25,
		MaxIdleConns:           5,
		ConnMaxLifetime:        5 * time.Minute,
		ConnMaxIdleTime:        10 * time.Minute,
		LogLevel:               "warn",
		SlowThreshold:          200 * time.Millisecond,
		SkipDefaultTransaction: false,
		PrepareStmt:            true,
		AutoMigrate:            false,
	}
}

// NewGormDatabase creates a new GORM database instance from configuration.
// It supports PostgreSQL, MySQL, and SQLite drivers.
func NewGormDatabase(cfg hyperion.Config) (hyperion.Database, error) {
	// Load configuration from hyperion.Config
	dbConfig := DefaultConfig()
	if err := loadConfig(cfg, dbConfig); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := dbConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Open database connection
	db, err := dbConfig.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &gormDatabase{db: db}, nil
}

// NewGormUnitOfWork creates a new UnitOfWork from a Database instance.
func NewGormUnitOfWork(db hyperion.Database) hyperion.UnitOfWork {
	gdb, ok := db.(*gormDatabase)
	if !ok {
		panic("db must be a *gormDatabase instance")
	}
	return &gormUnitOfWork{db: gdb.db}
}

// loadConfig loads configuration from hyperion.Config and merges with defaults.
// It unmarshals into a temporary struct to preserve default values for unset fields.
// Tries loading from both "database" prefix and root level, merging both results.
func loadConfig(src hyperion.Config, dst *Config) error {
	// Unmarshal into temporary struct to avoid overwriting defaults
	var temp Config

	// Try loading with "database" prefix first
	// Note: Viper returns nil even when key doesn't exist, so we can't rely on error
	_ = src.Unmarshal("database", &temp)

	// Also try root level config (without prefix) to support both styles:
	// 1. database.driver (prefixed)
	// 2. driver (root level)
	// Both are valid and will be merged (root level takes precedence if both exist)
	var rootTemp Config
	_ = src.Unmarshal("", &rootTemp)

	// Merge root-level values into temp (root takes precedence over prefixed)
	mergeConfigValues(&temp, &rootTemp)

	// Merge non-zero values from temp into dst, preserving defaults
	if temp.Driver != "" {
		dst.Driver = temp.Driver
	}
	if temp.Host != "" {
		dst.Host = temp.Host
	}
	if temp.Port != 0 {
		dst.Port = temp.Port
	}
	if temp.Username != "" {
		dst.Username = temp.Username
	}
	if temp.Password != "" {
		dst.Password = temp.Password
	}
	if temp.Database != "" {
		dst.Database = temp.Database
	}
	if temp.DSN != "" {
		dst.DSN = temp.DSN
	}
	if temp.SSLMode != "" {
		dst.SSLMode = temp.SSLMode
	}
	if temp.Charset != "" {
		dst.Charset = temp.Charset
	}
	if temp.MaxOpenConns != 0 {
		dst.MaxOpenConns = temp.MaxOpenConns
	}
	if temp.MaxIdleConns != 0 {
		dst.MaxIdleConns = temp.MaxIdleConns
	}
	if temp.ConnMaxLifetime != 0 {
		dst.ConnMaxLifetime = temp.ConnMaxLifetime
	}
	if temp.ConnMaxIdleTime != 0 {
		dst.ConnMaxIdleTime = temp.ConnMaxIdleTime
	}
	if temp.SlowThreshold != 0 {
		dst.SlowThreshold = temp.SlowThreshold
	}
	if temp.LogLevel != "" {
		dst.LogLevel = temp.LogLevel
	}

	// Boolean fields - only override if explicitly set to true
	// (since false is the zero value, we can't distinguish between unset and false)
	if temp.SkipDefaultTransaction {
		dst.SkipDefaultTransaction = temp.SkipDefaultTransaction
	}
	if temp.PrepareStmt {
		dst.PrepareStmt = temp.PrepareStmt
	}
	if temp.AutoMigrate {
		dst.AutoMigrate = temp.AutoMigrate
	}

	return nil
}

// mergeConfigValues merges non-zero values from src into dst.
// This allows root-level config to override prefixed config.
func mergeConfigValues(dst, src *Config) {
	if src.Driver != "" {
		dst.Driver = src.Driver
	}
	if src.Host != "" {
		dst.Host = src.Host
	}
	if src.Port != 0 {
		dst.Port = src.Port
	}
	if src.Username != "" {
		dst.Username = src.Username
	}
	if src.Password != "" {
		dst.Password = src.Password
	}
	if src.Database != "" {
		dst.Database = src.Database
	}
	if src.DSN != "" {
		dst.DSN = src.DSN
	}
	if src.SSLMode != "" {
		dst.SSLMode = src.SSLMode
	}
	if src.Charset != "" {
		dst.Charset = src.Charset
	}
	if src.MaxOpenConns != 0 {
		dst.MaxOpenConns = src.MaxOpenConns
	}
	if src.MaxIdleConns != 0 {
		dst.MaxIdleConns = src.MaxIdleConns
	}
	if src.ConnMaxLifetime != 0 {
		dst.ConnMaxLifetime = src.ConnMaxLifetime
	}
	if src.ConnMaxIdleTime != 0 {
		dst.ConnMaxIdleTime = src.ConnMaxIdleTime
	}
	if src.SlowThreshold != 0 {
		dst.SlowThreshold = src.SlowThreshold
	}
	if src.LogLevel != "" {
		dst.LogLevel = src.LogLevel
	}
	if src.SkipDefaultTransaction {
		dst.SkipDefaultTransaction = src.SkipDefaultTransaction
	}
	if src.PrepareStmt {
		dst.PrepareStmt = src.PrepareStmt
	}
	if src.AutoMigrate {
		dst.AutoMigrate = src.AutoMigrate
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}

	validDrivers := map[string]bool{
		DriverPostgres: true,
		DriverMySQL:    true,
		DriverSQLite:   true,
	}

	if !validDrivers[c.Driver] {
		return fmt.Errorf("unsupported driver: %s (supported: postgres, mysql, sqlite)", c.Driver)
	}

	if c.DSN == "" {
		// Validate individual connection parameters
		if c.Driver != DriverSQLite && c.Database == "" {
			return fmt.Errorf("database name is required")
		}
	}

	return nil
}

// Open opens a database connection with the configured settings.
func (c *Config) Open() (*gorm.DB, error) {
	// Configure GORM
	gormConfig := &gorm.Config{
		Logger:                 c.getLogger(),
		SkipDefaultTransaction: c.SkipDefaultTransaction,
		PrepareStmt:            c.PrepareStmt,
	}

	// Open connection based on driver
	var db *gorm.DB
	var err error

	switch c.Driver {
	case DriverPostgres:
		db, err = gorm.Open(postgres.Open(c.buildDSN()), gormConfig)
	case DriverMySQL:
		db, err = gorm.Open(mysql.Open(c.buildDSN()), gormConfig)
	case DriverSQLite:
		db, err = gorm.Open(sqlite.Open(c.buildDSN()), gormConfig)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", c.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(c.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(c.ConnMaxIdleTime)

	return db, nil
}

// buildDSN builds a DSN string based on the driver and configuration.
func (c *Config) buildDSN() string {
	// Use explicit DSN if provided
	if c.DSN != "" {
		return c.DSN
	}

	// Build DSN based on driver
	switch c.Driver {
	case DriverPostgres:
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
		)
	case DriverMySQL:
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset,
		)
	case DriverSQLite:
		return c.Database
	default:
		return ""
	}
}

// getLogger returns a GORM logger based on the configured log level.
func (c *Config) getLogger() logger.Interface {
	var logLevel logger.LogLevel

	switch c.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Warn
	}

	return logger.Default.LogMode(logLevel)
}
