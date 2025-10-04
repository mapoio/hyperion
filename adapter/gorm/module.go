package gorm

import (
	"context"

	"go.uber.org/fx"

	"github.com/mapoio/hyperion"
)

// Module provides GORM-based Database and UnitOfWork implementations.
//
// Usage:
//
//	fx.New(
//	    hyperion.CoreModule,
//	    viper.Module,  // Provides Config
//	    gorm.Module,   // Provides Database and UnitOfWork
//	    myapp.Module,
//	).Run()
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
var Module = fx.Module("hyperion.adapter.gorm",
	fx.Provide(
		fx.Annotate(
			NewGormProvider,
			fx.As(new(hyperion.Database)),
		),
	),
	fx.Provide(
		fx.Annotate(
			NewGormUnitOfWorkProvider,
			fx.As(new(hyperion.UnitOfWork)),
		),
	),
	fx.Invoke(registerLifecycle),
)

// NewGormProvider creates a GORM database.
func NewGormProvider(cfg hyperion.Config) (hyperion.Database, error) {
	return NewGormDatabase(cfg)
}

// NewGormUnitOfWorkProvider creates a GORM UnitOfWork.
func NewGormUnitOfWorkProvider(db hyperion.Database) hyperion.UnitOfWork {
	return NewGormUnitOfWork(db)
}

// registerLifecycle registers database lifecycle hooks with fx.
func registerLifecycle(lc fx.Lifecycle, db hyperion.Database) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return db.Close()
		},
	})
}
