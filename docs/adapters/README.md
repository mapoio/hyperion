# Hyperion Adapters Documentation

Complete documentation for all official Hyperion adapters.

## Available Adapters

### Configuration
- **[Viper](../../adapter/viper/README.md)** - Configuration management with hot-reload
  - Multiple format support (YAML, JSON, TOML, etc.)
  - Environment variable binding
  - Configuration watching and hot-reload

### Logging
- **[Zap](../../adapter/zap/README.md)** - High-performance structured logging
  - Blazing fast (1M+ logs/sec)
  - JSON and Console encoders
  - Log rotation with lumberjack

### Database
- **[GORM](../../adapter/gorm/README.md)** - Database connectivity and ORM
  - PostgreSQL, MySQL, SQLite support
  - Declarative transaction management
  - Connection pooling

## Quick Start

### Basic Application Setup

```go
package main

import (
    "go.uber.org/fx"

    viperadapter "github.com/mapoio/hyperion/adapter/viper"
    zapadapter "github.com/mapoio/hyperion/adapter/zap"
    gormadapter "github.com/mapoio/hyperion/adapter/gorm"
    "github.com/mapoio/hyperion"
)

func main() {
    app := fx.New(
        // Essential adapters
        viperadapter.Module,  // Config
        zapadapter.Module,    // Logger
        gormadapter.Module,   // Database

        // Your application
        fx.Invoke(run),
    )
    app.Run()
}

func run(
    cfg hyperion.Config,
    logger hyperion.Logger,
    db hyperion.Database,
    uow hyperion.UnitOfWork,
) {
    logger.Info("application started",
        "name", cfg.GetString("app.name"),
        "version", cfg.GetString("app.version"),
    )

    // Your application logic here
}
```

### Configuration File

Create `config.yaml`:

```yaml
app:
  name: myapp
  version: 1.0.0
  environment: production

log:
  level: info
  encoding: json
  output: /var/log/app.log

database:
  driver: postgres
  host: localhost
  port: 5432
  username: myuser
  password: mypassword
  database: mydb
```

## Implementation Reports

Detailed implementation reports for each adapter:

- [Viper Implementation Report](./reports/viper-implementation-report.md)
- [Zap Implementation Report](./reports/zap-implementation-report.md)
- [GORM Implementation Report](./reports/gorm-implementation-report.md)

## Adapter Development

### Creating a New Adapter

1. **Create Adapter Directory**
   ```bash
   mkdir -p adapter/myadapter
   cd adapter/myadapter
   ```

2. **Initialize Go Module**
   ```bash
   go mod init github.com/mapoio/hyperion/adapter/myadapter
   ```

3. **Implement Core Interface**
   ```go
   package myadapter

   import (
       "github.com/mapoio/hyperion"
   )

   type myAdapter struct {
       // fields
   }

   // Ensure interface compliance
   var _ hyperion.MyInterface = (*myAdapter)(nil)

   func New(cfg hyperion.Config) (hyperion.MyInterface, error) {
       // Implementation
   }
   ```

4. **Export fx.Module**
   ```go
   // module.go
   package myadapter

   import (
       "go.uber.org/fx"
       "github.com/mapoio/hyperion"
   )

   var Module = fx.Module("hyperion.adapter.myadapter",
       fx.Provide(
           fx.Annotate(
               New,
               fx.As(new(hyperion.MyInterface)),
           ),
       ),
   )
   ```

5. **Write Tests**
   - Unit tests: `*_test.go`
   - Integration tests: `integration_test.go`
   - Target: >80% coverage

6. **Add Documentation**
   - `README.md` - User guide with examples
   - `doc.go` - Package documentation
   - Implementation report (after completion)

7. **Register in Workspace**
   ```bash
   cd ../../
   go work use ./adapter/myadapter
   ```

8. **Update Makefile**
   ```makefile
   MODULES := hyperion adapter/viper adapter/zap adapter/gorm adapter/myadapter
   ```

### Adapter Quality Checklist

- [ ] Implements required Hyperion interface(s)
- [ ] Provides fx.Module for dependency injection
- [ ] Reads configuration via `hyperion.Config`
- [ ] Test coverage >= 80%
- [ ] Includes integration tests
- [ ] Has comprehensive README.md
- [ ] Has godoc package documentation
- [ ] Passes all linters
- [ ] Zero breaking changes to core

## Adapter Architecture

### Interface Pattern

```go
// Core defines interface
package hyperion

type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

```go
// Adapter implements interface
package ristretto

import "github.com/mapoio/hyperion"

type cache struct {
    client *ristretto.Cache
}

var _ hyperion.Cache = (*cache)(nil)

func New(cfg hyperion.Config) (hyperion.Cache, error) {
    // Implementation
}
```

### Dependency Injection

```go
// Adapter exports fx.Module
var Module = fx.Module("hyperion.adapter.ristretto",
    fx.Provide(
        fx.Annotate(
            New,
            fx.As(new(hyperion.Cache)),
        ),
    ),
)
```

```go
// Application imports adapter
import ristrettoadapter "github.com/mapoio/hyperion/adapter/ristretto"

app := fx.New(
    ristrettoadapter.Module,
    fx.Invoke(func(cache hyperion.Cache) {
        // Use cache
    }),
)
```

### Configuration

```go
// Adapter reads from hyperion.Config
func New(cfg hyperion.Config) (hyperion.Cache, error) {
    var config Config
    if err := cfg.Unmarshal("cache", &config); err != nil {
        return nil, err
    }

    // Use config to initialize
}
```

## Best Practices

### 1. Zero Lock-In

Adapters should be **swappable** without changing application code:

```go
// Application code (unchanged)
type Service struct {
    cache hyperion.Cache
}

// Swap adapter via configuration
// Old: import ristretto "github.com/mapoio/hyperion/adapter/ristretto"
// New: import redis "github.com/mapoio/hyperion/adapter/redis"
```

### 2. Configuration-Driven

```go
// ✅ Good - Configurable
func New(cfg hyperion.Config) (*Adapter, error) {
    var config Config
    cfg.Unmarshal("adapter", &config)
    return newAdapter(config)
}

// ❌ Avoid - Hardcoded
func New() *Adapter {
    return &Adapter{
        timeout: 30 * time.Second,
        maxConns: 100,
    }
}
```

### 3. Sensible Defaults

```go
type Config struct {
    MaxSize int           `mapstructure:"max_size"`
    TTL     time.Duration `mapstructure:"ttl"`
}

func (c *Config) SetDefaults() {
    if c.MaxSize == 0 {
        c.MaxSize = 1000
    }
    if c.TTL == 0 {
        c.TTL = 5 * time.Minute
    }
}
```

### 4. Lifecycle Management

```go
var Module = fx.Module("hyperion.adapter.mymodule",
    fx.Provide(New),
    fx.Invoke(registerLifecycle),
)

func registerLifecycle(lc fx.Lifecycle, adapter *Adapter) {
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            return adapter.Connect()
        },
        OnStop: func(ctx context.Context) error {
            return adapter.Close()
        },
    })
}
```

### 5. Error Wrapping

```go
func (a *adapter) Operation() error {
    if err := a.client.Do(); err != nil {
        return fmt.Errorf("adapter operation failed: %w", err)
    }
    return nil
}
```

## Testing Adapters

### Unit Tests

```go
func TestAdapter_Operation(t *testing.T) {
    cfg := &mockConfig{
        data: map[string]any{
            "adapter.host": "localhost",
            "adapter.port": 8080,
        },
    }

    adapter, err := New(cfg)
    if err != nil {
        t.Fatalf("New() error = %v", err)
    }

    // Test operations
}
```

### Integration Tests

```go
// +build integration

func TestAdapter_Integration(t *testing.T) {
    // Setup real dependencies (database, cache, etc.)
    container := setupTestContainer(t)
    defer container.Close()

    // Test with real connections
    adapter := newAdapter(container.Config())
    // ... test
}
```

## Roadmap

### Planned Adapters

**Epic 2 - Essential Adapters (v2.1)**:
- ✅ Viper (Configuration)
- ✅ Zap (Logging)
- ✅ GORM (Database)
- ⏳ Ristretto (Cache)

**Epic 3 - Infrastructure Adapters (v2.2)**:
- ⏳ Redis (Distributed Cache)
- ⏳ NATS (Messaging)
- ⏳ OpenTelemetry (Tracing)
- ⏳ MinIO (Object Storage)

**Epic 4 - Advanced Adapters (v2.3)**:
- ⏳ Temporal (Workflow)
- ⏳ Vault (Secrets)
- ⏳ Consul (Service Discovery)

## Contributing

### Submitting a New Adapter

1. **Discuss First**: Open an issue to discuss the adapter
2. **Follow Standards**: Use existing adapters as templates
3. **Complete Documentation**: README, godoc, and implementation report
4. **Test Coverage**: Minimum 80% coverage required
5. **Review Process**: Submit PR and respond to feedback

### Adapter Maintenance

- **Bug Fixes**: Critical bugs should be fixed ASAP
- **Dependencies**: Keep third-party dependencies up-to-date
- **Breaking Changes**: Require major version bump
- **Deprecation**: Give 6 months notice before removal

## Support

- **Documentation**: This directory and adapter READMEs
- **Examples**: See `examples/` directory in main repo
- **Issues**: https://github.com/mapoio/hyperion/issues
- **Discussions**: https://github.com/mapoio/hyperion/discussions

## License

All adapters follow the same license as Hyperion framework.
