# Viper Configuration Adapter for Hyperion

Production-ready Viper adapter providing configuration management with hot-reload capabilities for Hyperion framework.

## Features

- **Multiple Formats**: YAML, JSON, TOML, HCL, INI, ENV support
- **Environment Variables**: Automatic environment variable binding
- **Hot Reload**: Watch configuration files for changes
- **Nested Configuration**: Access nested keys with dot notation
- **Type Safety**: Type-safe configuration access methods
- **Default Values**: Graceful fallback to defaults
- **Zero Dependencies**: Clean integration with Hyperion interfaces

## Installation

```bash
go get github.com/mapoio/hyperion/adapter/viper
```

## Quick Start

### 1. Create Configuration File

Create a `config.yaml` file:

```yaml
app:
  name: myapp
  version: 1.0.0
  debug: true

server:
  host: 0.0.0.0
  port: 8080
  timeout: 30s

log:
  level: info
  encoding: json
  output: stdout

database:
  driver: postgres
  host: localhost
  port: 5432
  database: mydb
```

### 2. Initialize Application

```go
package main

import (
    "go.uber.org/fx"
    viperadapter "github.com/mapoio/hyperion/adapter/viper"
    "github.com/mapoio/hyperion"
)

func main() {
    app := fx.New(
        hyperion.CoreModule,  // Core infrastructure (required)
        viperadapter.Module,  // Config provider
        fx.Invoke(run),
    )
    app.Run()
}

func run(cfg hyperion.Config) {
    // Configuration is ready to use
    appName := cfg.GetString("app.name")
    port := cfg.GetInt("server.port")
}
```

### 3. Use Configuration

#### Basic Access

```go
type AppService struct {
    cfg hyperion.Config
}

func (s *AppService) GetSettings() {
    // String values
    name := s.cfg.GetString("app.name")
    host := s.cfg.GetString("server.host")

    // Integer values
    port := s.cfg.GetInt("server.port")

    // Boolean values
    debug := s.cfg.GetBool("app.debug")

    // Check if key exists
    if s.cfg.IsSet("database.driver") {
        driver := s.cfg.GetString("database.driver")
    }
}
```

#### Unmarshal to Struct

```go
type ServerConfig struct {
    Host    string        `mapstructure:"host"`
    Port    int           `mapstructure:"port"`
    Timeout time.Duration `mapstructure:"timeout"`
}

type AppConfig struct {
    Name    string `mapstructure:"name"`
    Version string `mapstructure:"version"`
    Debug   bool   `mapstructure:"debug"`
}

func (s *AppService) LoadConfig() error {
    var serverCfg ServerConfig
    if err := s.cfg.Unmarshal("server", &serverCfg); err != nil {
        return fmt.Errorf("failed to unmarshal server config: %w", err)
    }

    var appCfg AppConfig
    if err := s.cfg.Unmarshal("app", &appCfg); err != nil {
        return fmt.Errorf("failed to unmarshal app config: %w", err)
    }

    return nil
}
```

#### Hot Reload

```go
func (s *AppService) WatchConfig() {
    s.cfg.Watch(func(cfg hyperion.Config) {
        log.Println("Configuration changed, reloading...")

        // Reload settings
        newPort := cfg.GetInt("server.port")
        if newPort != s.currentPort {
            s.restartServer(newPort)
        }
    })
}
```

## Configuration Structure

### Recommended Structure

```yaml
# Application settings
app:
  name: string
  version: string
  debug: bool
  environment: string  # development, staging, production

# Server settings
server:
  host: string
  port: int
  timeout: duration
  read_timeout: duration
  write_timeout: duration

# Logging
log:
  level: string       # debug, info, warn, error
  encoding: string    # json, console
  output: string      # stdout, stderr, file path

# Database
database:
  driver: string      # postgres, mysql, sqlite
  host: string
  port: int
  username: string
  password: string
  database: string

# Cache
cache:
  type: string        # memory, redis
  max_size: int
  ttl: duration

# Tracing
tracing:
  enabled: bool
  endpoint: string
  sample_rate: float
```

## Environment Variables

### Automatic Binding

Environment variables are automatically bound with `HYPERION_` prefix:

```bash
export HYPERION_APP_NAME="myapp"
export HYPERION_SERVER_PORT=8080
export HYPERION_LOG_LEVEL="debug"
```

Access in code:

```go
name := cfg.GetString("app.name")      // Gets HYPERION_APP_NAME
port := cfg.GetInt("server.port")      // Gets HYPERION_SERVER_PORT
level := cfg.GetString("log.level")    // Gets HYPERION_LOG_LEVEL
```

### Override Precedence

Configuration values are resolved in the following order (highest to lowest):

1. **Explicit Set** - Values set programmatically
2. **Environment Variables** - `HYPERION_*` variables
3. **Configuration File** - Values from config.yaml
4. **Default Values** - Fallback defaults

## Advanced Usage

### Multiple Configuration Files

```go
package main

import (
    "go.uber.org/fx"
    viperadapter "github.com/mapoio/hyperion/adapter/viper"
)

func main() {
    app := fx.New(
        fx.Provide(
            fx.Annotate(
                func() (hyperion.Config, error) {
                    return viperadapter.NewProvider(
                        viperadapter.WithConfigFiles("config.yaml", "local.yaml"),
                        viperadapter.WithEnvPrefix("MYAPP"),
                    )
                },
                fx.As(new(hyperion.Config)),
            ),
        ),
        fx.Invoke(run),
    )
    app.Run()
}
```

### Configuration Validation

```go
type DatabaseConfig struct {
    Driver   string `mapstructure:"driver" validate:"required,oneof=postgres mysql sqlite"`
    Host     string `mapstructure:"host" validate:"required"`
    Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
    Database string `mapstructure:"database" validate:"required"`
}

func ValidateConfig(cfg hyperion.Config) error {
    var dbCfg DatabaseConfig
    if err := cfg.Unmarshal("database", &dbCfg); err != nil {
        return err
    }

    validate := validator.New()
    return validate.Struct(dbCfg)
}
```

## Best Practices

### 1. Use Struct Unmarshaling

```go
// ✅ Good - Type-safe, documented
type Config struct {
    Database DatabaseConfig `mapstructure:"database"`
    Server   ServerConfig   `mapstructure:"server"`
}

var cfg Config
if err := config.Unmarshal("", &cfg); err != nil {
    log.Fatal(err)
}

// ❌ Avoid - Error-prone, no type safety
host := config.GetString("database.host")
port := config.GetInt("database.port")
```

### 2. Validate Early

```go
func NewService(cfg hyperion.Config) (*Service, error) {
    var config ServiceConfig
    if err := cfg.Unmarshal("service", &config); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }

    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    return &Service{config: config}, nil
}
```

### 3. Use Sensible Defaults

```go
type Config struct {
    Host    string        `mapstructure:"host"`
    Port    int           `mapstructure:"port"`
    Timeout time.Duration `mapstructure:"timeout"`
}

func (c *Config) SetDefaults() {
    if c.Host == "" {
        c.Host = "localhost"
    }
    if c.Port == 0 {
        c.Port = 8080
    }
    if c.Timeout == 0 {
        c.Timeout = 30 * time.Second
    }
}
```

### 4. Document Configuration

```go
// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
    // Driver specifies the database driver (postgres, mysql, sqlite).
    Driver string `mapstructure:"driver"`

    // Host is the database server hostname or IP address.
    Host string `mapstructure:"host"`

    // Port is the database server port (default: 5432 for postgres).
    Port int `mapstructure:"port"`
}
```

## Testing

### Mock Configuration

```go
type mockConfig struct {
    data map[string]any
}

func (m *mockConfig) GetString(key string) string {
    if v, ok := m.data[key].(string); ok {
        return v
    }
    return ""
}

func (m *mockConfig) GetInt(key string) int {
    if v, ok := m.data[key].(int); ok {
        return v
    }
    return 0
}

// Use in tests
func TestService(t *testing.T) {
    cfg := &mockConfig{
        data: map[string]any{
            "app.name": "test",
            "server.port": 8080,
        },
    }

    service := NewService(cfg)
    // Test service
}
```

## Troubleshooting

### Configuration Not Loading

**Problem**: Configuration values are empty or default

**Solutions**:
1. Check file path is correct (relative to working directory)
2. Verify YAML/JSON syntax is valid
3. Check environment variable names (use `HYPERION_` prefix)
4. Enable debug logging to see what Viper is loading

### Environment Variables Not Working

**Problem**: Environment variables are not overriding config file

**Solutions**:
1. Ensure variable names use `HYPERION_` prefix
2. Use uppercase with underscores: `HYPERION_APP_NAME`
3. Nested keys use underscores: `HYPERION_DATABASE_HOST`

### Hot Reload Not Triggering

**Problem**: Configuration changes not detected

**Solutions**:
1. Check file permissions (Viper needs read access)
2. Verify `Watch()` is called after file is loaded
3. On some systems, atomic writes (mv) may not trigger fsnotify
4. Try using symlinks or direct file edits

## Performance

- **Memory**: ~1-2 MB overhead for typical configuration
- **Startup**: < 10ms to load and parse configuration
- **Watch**: Minimal CPU usage (<1%) when watching files
- **Access**: O(1) for key lookups using internal map

## Integration Examples

### With Zap Logger

```go
app := fx.New(
    hyperion.CoreModule,  // Core infrastructure (required)
    viperadapter.Module,
    zapadapter.Module,  // Reads log.* from config
    fx.Invoke(func(logger hyperion.Logger, cfg hyperion.Config) {
        logger.Info("application started",
            "name", cfg.GetString("app.name"),
            "version", cfg.GetString("app.version"),
        )
    }),
)
```

### With GORM Database

```go
app := fx.New(
    hyperion.CoreModule,  // Core infrastructure (required)
    viperadapter.Module,
    gormadapter.Module,  // Reads database.* from config
    fx.Invoke(func(db hyperion.Database, cfg hyperion.Config) {
        // Database configured automatically
    }),
)
```

## License

Same as Hyperion framework.

## Contributing

See main Hyperion repository for contribution guidelines.

## Support

- Documentation: https://github.com/mapoio/hyperion
- Issues: https://github.com/mapoio/hyperion/issues
- Discussions: https://github.com/mapoio/hyperion/discussions
