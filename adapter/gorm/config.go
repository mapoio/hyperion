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

// loadConfig loads configuration from hyperion.Config into Config struct.
func loadConfig(src hyperion.Config, dst *Config) error {
	// Try loading with "database" prefix
	if err := src.Unmarshal("database", dst); err != nil {
		// If that fails, try without prefix (root level)
		if err := src.Unmarshal("", dst); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	return nil
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
