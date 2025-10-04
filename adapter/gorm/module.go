package gorm

import (
	"context"

	"go.uber.org/fx"

	"github.com/mapoio/hyperion"
)

// Module provides GORM database adapter as hyperion.Database and hyperion.UnitOfWork
// via fx dependency injection.
//
// Usage:
//
//	app := fx.New(
//	    viper.Module,  // Provides Config
//	    gorm.Module,   // Provides Database and UnitOfWork
//	    fx.Invoke(func(db hyperion.Database, uow hyperion.UnitOfWork) {
//	        // Use database and unit of work
//	    }),
//	)
//
// The module automatically handles database lifecycle:
//   - Opens connection during application startup
//   - Closes connection during graceful shutdown
//
// Configuration example (config.yaml):
//
//	database:
//	  driver: postgres
//	  host: localhost
//	  port: 5432
//	  username: dbuser
//	  password: dbpass
//	  database: mydb
//	  max_open_conns: 25
//	  max_idle_conns: 5
var Module = fx.Module("hyperion.adapter.gorm",
	fx.Decorate(
		fx.Annotate(
			NewGormDatabase,
			fx.As(new(hyperion.Database)),
		),
	),
	fx.Provide(
		fx.Annotate(
			NewGormUnitOfWork,
			fx.As(new(hyperion.UnitOfWork)),
		),
	),
	fx.Invoke(registerLifecycle),
)

// registerLifecycle registers database lifecycle hooks with fx.
// It ensures the database connection is properly closed during shutdown.
func registerLifecycle(lc fx.Lifecycle, db hyperion.Database) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return db.Close()
		},
	})
}
