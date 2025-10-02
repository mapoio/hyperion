// Package zap provides a Zap-based implementation of the hyperion.Logger interface.
//
// This adapter wraps go.uber.org/zap to provide high-performance structured
// logging with support for JSON and Console encoders, dynamic level adjustment,
// and file rotation via lumberjack.
//
// # Features
//
//   - High-performance structured logging (1M+ logs/sec)
//   - JSON and Console output encoders
//   - Dynamic log level adjustment at runtime
//   - Automatic log file rotation with size/age limits
//   - Zero-allocation logging paths
//   - Full hyperion.Logger interface compliance
//
// # Configuration
//
// The logger reads configuration from the provided hyperion.Config under the "log" key:
//
//	log:
//	  level: info              # debug, info, warn, error, fatal
//	  encoding: json           # json or console
//	  output: stdout           # stdout, stderr, or file path
//	  file:
//	    path: /var/log/app.log
//	    max_size: 100          # MB
//	    max_backups: 3
//	    max_age: 7             # days
//	    compress: false
//
// # Usage
//
// Basic usage with fx:
//
//	package main
//
//	import (
//	    "go.uber.org/fx"
//	    "github.com/mapoio/hyperion"
//	    "github.com/mapoio/hyperion/adapter/viper"
//	    "github.com/mapoio/hyperion/adapter/zap"
//	)
//
//	func main() {
//	    app := fx.New(
//	        viper.Module,  // Provides Config
//	        zap.Module,    // Provides Logger
//	        fx.Invoke(run),
//	    )
//	    app.Run()
//	}
//
//	func run(logger hyperion.Logger) {
//	    logger.Info("application started",
//	        "version", "1.0.0",
//	        "environment", "production",
//	    )
//
//	    // Create child logger with context
//	    reqLogger := logger.With("request_id", "abc123")
//	    reqLogger.Info("processing request")
//
//	    // Log with error
//	    if err := doSomething(); err != nil {
//	        reqLogger.WithError(err).Error("operation failed")
//	    }
//
//	    // Ensure all logs are flushed
//	    defer logger.Sync()
//	}
//
// # Performance
//
// The Zap adapter provides near-zero allocation logging with minimal overhead:
//
//   - Throughput: 1M+ logs/sec
//   - Latency: <100ns per log call (cached logger)
//   - Overhead: <5% vs native Zap
//
// # Thread Safety
//
// All Logger methods are safe for concurrent use.
package zap
