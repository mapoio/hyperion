package gorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/mapoio/hyperion"
)

// gormDatabase implements hyperion.Database interface using GORM v2.
// It provides database connectivity, health checks, and graceful shutdown.
type gormDatabase struct {
	db *gorm.DB
}

// Ensure interface compliance at compile time.
var _ hyperion.Database = (*gormDatabase)(nil)

// Executor returns the default database executor.
// The returned executor can be used for non-transactional operations.
func (d *gormDatabase) Executor() hyperion.Executor {
	return &gormExecutor{
		db:   d.db,
		isTx: false,
	}
}

// Health checks the database connection by pinging the underlying sql.DB.
// Returns an error if the connection is not healthy.
func (d *gormDatabase) Health(ctx context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close closes the underlying database connection pool.
// This should be called during application shutdown.
func (d *gormDatabase) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}
