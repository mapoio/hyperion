// Package gorm provides a production-ready GORM v2 adapter for Hyperion framework.
//
// This adapter implements hyperion.Database, hyperion.Executor, and hyperion.UnitOfWork
// interfaces, providing database connectivity and declarative transaction management.
//
// # Supported Databases
//
// The adapter supports three database drivers:
//   - PostgreSQL (via gorm.io/driver/postgres)
//   - MySQL (via gorm.io/driver/mysql)
//   - SQLite (via gorm.io/driver/sqlite)
//
// # Installation
//
// Add the adapter to your application:
//
//	import (
//	    "go.uber.org/fx"
//	    gormadapter "github.com/mapoio/hyperion/adapter/gorm"
//	    "github.com/mapoio/hyperion"
//	)
//
//	app := fx.New(
//	    viper.Module,        // Provides Config
//	    gormadapter.Module,  // Provides Database and UnitOfWork
//	    fx.Invoke(func(db hyperion.Database) {
//	        // Database is ready to use
//	    }),
//	)
//
// # Configuration
//
// Configure the database via YAML configuration file:
//
//	database:
//	  driver: postgres
//	  host: localhost
//	  port: 5432
//	  username: dbuser
//	  password: dbpass
//	  database: mydb
//	  sslmode: disable
//
//	  # Connection pool settings
//	  max_open_conns: 25
//	  max_idle_conns: 5
//	  conn_max_lifetime: 5m
//	  conn_max_idle_time: 10m
//
//	  # GORM settings
//	  log_level: warn
//	  slow_threshold: 200ms
//	  skip_default_transaction: false
//	  prepare_stmt: true
//
// Alternatively, use a DSN string:
//
//	database:
//	  driver: postgres
//	  dsn: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
//
// # Basic Usage
//
// ## Querying Data
//
//	type UserRepository struct {
//	    db hyperion.Database
//	}
//
//	func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
//	    executor := r.db.Executor()
//	    db := executor.Unwrap().(*gorm.DB)
//
//	    var user User
//	    if err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
//	        return nil, err
//	    }
//	    return &user, nil
//	}
//
// ## Using Executor Interface
//
//	func (r *UserRepository) CountUsers(ctx hyperion.Context) (int64, error) {
//	    var count int64
//	    sql := "SELECT COUNT(*) FROM users"
//
//	    err := ctx.DB().Query(ctx, &count, sql)
//	    return count, err
//	}
//
// # Transaction Management
//
// The adapter provides declarative transaction management via UnitOfWork:
//
//	type UserService struct {
//	    uow      hyperion.UnitOfWork
//	    userRepo *UserRepository
//	}
//
//	func (s *UserService) RegisterUser(ctx hyperion.Context, email string) error {
//	    return s.uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
//	        // Create user
//	        user := &User{Email: email}
//	        if err := s.userRepo.Create(txCtx, user); err != nil {
//	            return err  // Automatic rollback
//	        }
//
//	        // Create profile
//	        profile := &Profile{UserID: user.ID}
//	        if err := s.profileRepo.Create(txCtx, profile); err != nil {
//	            return err  // Automatic rollback
//	        }
//
//	        return nil  // Automatic commit
//	    })
//	}
//
// # Transaction Options
//
// Set isolation level and read-only mode:
//
//	opts := &hyperion.TransactionOptions{
//	    Isolation: hyperion.IsolationLevelSerializable,
//	    ReadOnly:  false,
//	}
//
//	err := uow.WithTransactionOptions(ctx, opts, func(txCtx hyperion.Context) error {
//	    // Transaction with custom options
//	    return orderRepo.Create(txCtx, order)
//	})
//
// # Transaction Propagation
//
// Transactions automatically propagate through hyperion.Context:
//
//	func (r *UserRepository) Create(ctx hyperion.Context, user *User) error {
//	    // ctx.DB() returns transaction executor if inside WithTransaction
//	    // Otherwise, returns the default executor
//	    db := ctx.DB().Unwrap().(*gorm.DB)
//	    return db.Create(user).Error
//	}
//
// # Nested Transactions
//
// GORM automatically handles nested transactions using savepoints:
//
//	err := uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
//	    // Outer transaction
//
//	    // Nested transaction (creates savepoint)
//	    return uow.WithTransaction(txCtx, func(nestedCtx hyperion.Context) error {
//	        // Inner transaction
//	        return userRepo.Create(nestedCtx, user)
//	    })
//	})
//
// # Panic Recovery
//
// Transactions are automatically rolled back on panic:
//
//	err := uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
//	    userRepo.Create(txCtx, user)
//	    panic("something went wrong")  // Transaction rolled back
//	})
//
// # Health Checks
//
// Check database connectivity:
//
//	if err := db.Health(ctx); err != nil {
//	    log.Fatal("database unhealthy", err)
//	}
//
// # Graceful Shutdown
//
// The adapter automatically closes connections during fx shutdown:
//
//	app := fx.New(
//	    gormadapter.Module,
//	    // ... other modules
//	)
//
//	// Database connection closed automatically when app.Stop() is called
//
// # Performance Considerations
//
// The adapter adds minimal overhead (<5%) compared to native GORM:
//   - Executor interface: ~1-2% overhead
//   - Transaction management: ~2-3% overhead
//   - Total overhead: <5%
//
// # Best Practices
//
// 1. Always use hyperion.Context for database operations
// 2. Use WithTransaction for multi-step operations
// 3. Let transactions commit/rollback automatically
// 4. Use Unwrap() when you need GORM-specific features
// 5. Configure connection pool based on workload
// 6. Enable prepared statements for better performance
// 7. Set appropriate slow query threshold for monitoring
//
// # Limitations
//
// 1. Only supports GORM v2 (v1.25.0+)
// 2. Savepoints require database support (not all SQLite configurations)
// 3. Nested transaction depth limited by database
//
// # See Also
//
//   - hyperion.Database: Core database interface
//   - hyperion.Executor: Database operation executor
//   - hyperion.UnitOfWork: Transaction management
//   - hyperion.Context: Type-safe context with DB access
//   - GORM v2 documentation: https://gorm.io/docs/
package gorm
