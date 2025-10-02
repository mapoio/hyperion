# GORM Database Adapter for Hyperion

Production-ready GORM v2 adapter providing database connectivity and declarative transaction management for Hyperion framework.

## Features

- **Multiple Database Support**: PostgreSQL, MySQL, SQLite
- **Declarative Transactions**: Automatic commit/rollback with panic recovery
- **Connection Pooling**: Configurable connection pool settings
- **Transaction Propagation**: Type-safe transaction context propagation
- **Nested Transactions**: Automatic savepoint management
- **Health Checks**: Built-in database health monitoring
- **Low Overhead**: < 5% performance overhead vs native GORM

## Installation

```bash
go get github.com/mapoio/hyperion/adapter/gorm
```

## Quick Start

### 1. Configure Database

Create a `config.yaml` file:

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  username: myuser
  password: mypassword
  database: mydb
  sslmode: disable

  # Connection pool (optional)
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

  # GORM settings (optional)
  log_level: warn
  slow_threshold: 200ms
  prepare_stmt: true
```

### 2. Initialize Application

```go
package main

import (
    "go.uber.org/fx"
    viperadapter "github.com/mapoio/hyperion/adapter/viper"
    gormadapter "github.com/mapoio/hyperion/adapter/gorm"
    "github.com/mapoio/hyperion"
)

func main() {
    app := fx.New(
        viperadapter.Module,  // Config provider
        gormadapter.Module,   // Database provider
        fx.Invoke(run),
    )
    app.Run()
}

func run(db hyperion.Database, uow hyperion.UnitOfWork) {
    // Database is ready to use
}
```

### 3. Use Database

#### Basic Query

```go
type UserRepository struct {
    db hyperion.Database
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    executor := r.db.Executor()
    db := executor.Unwrap().(*gorm.DB)

    var user User
    if err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

#### Transaction

```go
type UserService struct {
    uow      hyperion.UnitOfWork
    userRepo *UserRepository
}

func (s *UserService) RegisterUser(ctx hyperion.Context, email string) error {
    return s.uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
        // Create user
        user := &User{Email: email}
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err  // Automatic rollback
        }

        // Create profile
        profile := &Profile{UserID: user.ID}
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err  // Automatic rollback
        }

        return nil  // Automatic commit
    })
}
```

## Configuration Reference

### Basic Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `driver` | string | `sqlite` | Database driver: `postgres`, `mysql`, `sqlite` |
| `host` | string | `localhost` | Database host |
| `port` | int | `5432` | Database port |
| `username` | string | - | Database username |
| `password` | string | - | Database password |
| `database` | string | - | Database name (or file path for SQLite) |
| `dsn` | string | - | Complete DSN string (overrides individual params) |

### Connection Pool

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `max_open_conns` | int | `25` | Maximum open connections |
| `max_idle_conns` | int | `5` | Maximum idle connections |
| `conn_max_lifetime` | duration | `5m` | Connection max lifetime |
| `conn_max_idle_time` | duration | `10m` | Connection max idle time |

### GORM Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `log_level` | string | `warn` | Log level: `silent`, `error`, `warn`, `info` |
| `slow_threshold` | duration | `200ms` | Slow query threshold |
| `skip_default_transaction` | bool | `false` | Skip default transaction for single operations |
| `prepare_stmt` | bool | `true` | Use prepared statements |

### Driver-Specific Options

#### PostgreSQL

```yaml
database:
  driver: postgres
  sslmode: disable  # disable, require, verify-ca, verify-full
```

#### MySQL

```yaml
database:
  driver: mysql
  charset: utf8mb4  # Character set
```

## Advanced Usage

### Transaction Options

Set isolation level and read-only mode:

```go
opts := &hyperion.TransactionOptions{
    Isolation: hyperion.IsolationLevelSerializable,
    ReadOnly:  false,
}

err := uow.WithTransactionOptions(ctx, opts, func(txCtx hyperion.Context) error {
    // Transaction with custom options
    return orderRepo.Create(txCtx, order)
})
```

### Nested Transactions

GORM automatically handles nested transactions using savepoints:

```go
err := uow.WithTransaction(ctx, func(txCtx hyperion.Context) error {
    // Outer transaction
    userRepo.Create(txCtx, user)

    // Nested transaction (creates savepoint)
    return uow.WithTransaction(txCtx, func(nestedCtx hyperion.Context) error {
        // Inner transaction
        return profileRepo.Create(nestedCtx, profile)
    })
})
```

### Health Checks

Check database connectivity:

```go
if err := db.Health(ctx); err != nil {
    log.Fatal("database unhealthy", err)
}
```

### Accessing Native GORM

Use `Unwrap()` to access GORM-specific features:

```go
func (r *UserRepository) ComplexQuery(ctx hyperion.Context) ([]User, error) {
    db := ctx.DB().Unwrap().(*gorm.DB)

    var users []User
    err := db.WithContext(ctx).
        Preload("Profile").
        Where("age > ?", 18).
        Order("created_at DESC").
        Find(&users).Error

    return users, err
}
```

## Supported Databases

### PostgreSQL

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  username: user
  password: pass
  database: mydb
  sslmode: disable
```

Tested with PostgreSQL 14+

### MySQL

```yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  username: user
  password: pass
  database: mydb
  charset: utf8mb4
```

Tested with MySQL 8.0+

### SQLite

```yaml
database:
  driver: sqlite
  database: /path/to/database.db
```

Or in-memory:

```yaml
database:
  driver: sqlite
  database: ":memory:"
```

## Best Practices

1. **Always use hyperion.Context** for database operations
2. **Use WithTransaction** for multi-step operations requiring atomicity
3. **Let transactions commit/rollback automatically** - avoid manual commit/rollback
4. **Use Unwrap()** when you need GORM-specific features
5. **Configure connection pool** based on your workload
6. **Enable prepared statements** for better performance (`prepare_stmt: true`)
7. **Set slow query threshold** for monitoring (`slow_threshold: 200ms`)
8. **Use health checks** in readiness probes

## Performance Considerations

The adapter adds minimal overhead compared to native GORM:

- **Executor interface**: ~1-2% overhead
- **Transaction management**: ~2-3% overhead
- **Total overhead**: < 5%

### Optimization Tips

1. **Connection pooling**: Adjust `max_open_conns` based on workload
2. **Prepared statements**: Enabled by default, reduces parsing overhead
3. **Skip default transaction**: Set `skip_default_transaction: true` for read-heavy workloads
4. **Batch operations**: Use GORM's batch insert for bulk operations

## Troubleshooting

### Connection Refused

```
Error: failed to connect to database: dial tcp: connection refused
```

**Solution**: Verify database host, port, and ensure database is running.

### Too Many Connections

```
Error: pq: sorry, too many clients already
```

**Solution**: Reduce `max_open_conns` or increase database max connections.

### Transaction Deadlock

```
Error: deadlock detected
```

**Solution**: Use appropriate isolation level or implement retry logic.

### Slow Queries

**Solution**:
- Check `log_level: info` to see slow query logs
- Adjust `slow_threshold` to identify slow queries
- Add database indexes

## Limitations

1. **GORM v2 only**: Supports GORM v1.25.0+, no support for GORM v1
2. **Savepoints**: Nested transactions require database support (not all SQLite configurations)
3. **Nested depth**: Limited by database capabilities

## Examples

See [examples](../../examples) directory for complete working examples.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## License

See [LICENSE](../../LICENSE) for details.
