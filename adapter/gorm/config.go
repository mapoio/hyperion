package gorm

import (
	"fmt"
	"time"

	"dario.cat/mergo"
	"github.com/go-playground/validator/v10"
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
	Driver   string `json:"driver" yaml:"driver" validate:"required,oneof=postgres mysql sqlite"` // Driver specifies the database driver (postgres, mysql, sqlite)
	DSN      string `json:"dsn" yaml:"dsn"`                                                       // DSN allows providing a complete connection string
	Host     string `json:"host" yaml:"host" validate:"omitempty,hostname|ip"`                    // Connection host
	Username string `json:"username" yaml:"username"`                                             // Connection username
	Password string `json:"password" yaml:"password"`                                             // Connection password
	Database string `json:"database" yaml:"database"`                                             // Database name (required for non-SQLite unless DSN provided)
	SSLMode  string `json:"sslmode" yaml:"sslmode" validate:"omitempty,oneof=disable require verify-ca verify-full"`
	Charset  string `json:"charset" yaml:"charset"`     // MySQL charset
	LogLevel string `json:"log_level" yaml:"log_level"` // Log level: silent, error, warn, info

	// Duration fields (8 bytes each)
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime" validate:"omitempty,min=0"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time" validate:"omitempty,min=0"`
	SlowThreshold   time.Duration `json:"slow_threshold" yaml:"slow_threshold" validate:"omitempty,min=0"`

	// Int fields (8 bytes on 64-bit)
	MaxOpenConns int `json:"max_open_conns" yaml:"max_open_conns" validate:"omitempty,min=0"`
	MaxIdleConns int `json:"max_idle_conns" yaml:"max_idle_conns" validate:"omitempty,min=0"`
	Port         int `json:"port" yaml:"port" validate:"omitempty,min=1,max=65535"`

	// Bool fields (1 byte each) - use pointers to distinguish unset from false
	// Using pointers allows us to differentiate between "not provided" (nil) and "explicitly set to false"
	SkipDefaultTransaction *bool `json:"skip_default_transaction" yaml:"skip_default_transaction"`
	PrepareStmt            *bool `json:"prepare_stmt" yaml:"prepare_stmt"`
	AutoMigrate            *bool `json:"auto_migrate" yaml:"auto_migrate"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	skipDefaultTransaction := false
	prepareStmt := true
	autoMigrate := false

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
		SkipDefaultTransaction: &skipDefaultTransaction,
		PrepareStmt:            &prepareStmt,
		AutoMigrate:            &autoMigrate,
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

// loadConfig loads configuration from hyperion.Config and merges with defaults using mergo.
// It unmarshals from both "database" prefix and root level, with root taking precedence.
// Uses mergo for clean struct merging, with special handling for boolean pointers.
func loadConfig(src hyperion.Config, dst *Config) error {
	// Load prefixed configuration (database.*)
	var prefixed Config
	_ = src.Unmarshal("database", &prefixed)

	// Load root-level configuration (driver, host, etc.)
	var root Config
	_ = src.Unmarshal("", &root)

	// Merge configurations: root overrides prefixed
	// mergo.WithOverride allows root values to override prefixed values
	if err := mergo.Merge(&prefixed, root, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge root config: %w", err)
	}

	// Merge loaded config into dst, preserving defaults for unset fields
	// mergo.WithOverride ensures user config overrides defaults
	if err := mergo.Merge(dst, prefixed, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	// Special handling for boolean pointers: mergo doesn't override non-nil pointers
	// We manually copy pointer fields if they are set in the source (non-nil)
	if prefixed.SkipDefaultTransaction != nil {
		dst.SkipDefaultTransaction = prefixed.SkipDefaultTransaction
	}
	if prefixed.PrepareStmt != nil {
		dst.PrepareStmt = prefixed.PrepareStmt
	}
	if prefixed.AutoMigrate != nil {
		dst.AutoMigrate = prefixed.AutoMigrate
	}

	return nil
}

// Validate checks if the configuration is valid using validator.
// Validation rules are defined in struct tags (validate:"...").
// Additional custom validation is performed after struct validation.
func (c *Config) Validate() error {
	// Run struct-level validation based on tags
	validate := validator.New()

	if err := validate.Struct(c); err != nil {
		// Return validation error with field details
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Format first validation error for better user experience
			firstErr := validationErrors[0]
			return fmt.Errorf("validation failed for field '%s': %s (value: '%v')",
				firstErr.Field(),
				firstErr.Tag(),
				firstErr.Value())
		}
		return err
	}

	// Custom validation: Database is required for non-SQLite drivers unless DSN is provided
	if c.DSN == "" && c.Driver != DriverSQLite && c.Database == "" {
		return fmt.Errorf("database name is required for driver '%s'", c.Driver)
	}

	return nil
}

// Open opens a database connection with the configured settings.
func (c *Config) Open() (*gorm.DB, error) {
	// Configure GORM - dereference boolean pointers with defaults
	skipTx := false
	if c.SkipDefaultTransaction != nil {
		skipTx = *c.SkipDefaultTransaction
	}
	prepare := true
	if c.PrepareStmt != nil {
		prepare = *c.PrepareStmt
	}

	gormConfig := &gorm.Config{
		Logger:                 c.getLogger(),
		SkipDefaultTransaction: skipTx,
		PrepareStmt:            prepare,
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
