package zap

import (
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/mapoio/hyperion"
)

// zapLogger implements hyperion.Logger interface using Zap.
type zapLogger struct {
	sugar *zap.SugaredLogger
	atom  zap.AtomicLevel
	core  *zap.Logger
}

// Ensure zapLogger implements hyperion.Logger interface.
var _ hyperion.Logger = (*zapLogger)(nil)

// Config holds configuration for Zap logger.
type Config struct {
	Level      string      `mapstructure:"level"`    // Log level: debug, info, warn, error, fatal
	Encoding   string      `mapstructure:"encoding"` // Encoding format: json or console
	Output     string      `mapstructure:"output"`   // Output destination: stdout, stderr, or file path
	FileConfig *FileConfig `mapstructure:"file"`     // File rotation configuration
}

// FileConfig holds file rotation configuration.
type FileConfig struct {
	Path       string `mapstructure:"path"`        // Log file path
	MaxSize    int    `mapstructure:"max_size"`    // Max file size in MB before rotation
	MaxBackups int    `mapstructure:"max_backups"` // Max number of old log files to keep
	MaxAge     int    `mapstructure:"max_age"`     // Max days to retain old log files
	Compress   bool   `mapstructure:"compress"`    // Whether to compress rotated files
}

// NewZapLogger creates a new Zap logger instance.
// It reads configuration from the provided hyperion.Config under the "log" key.
// If no configuration is found, sensible defaults are used.
func NewZapLogger(cfg hyperion.Config) (hyperion.Logger, error) {
	// Read configuration
	logCfg := &Config{
		Level:    "info",
		Encoding: "json",
		Output:   "stdout",
	}

	if cfg != nil {
		if err := cfg.Unmarshal("log", logCfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal log config: %w", err)
		}
	}

	// Parse log level
	atom := zap.NewAtomicLevel()
	if err := atom.UnmarshalText([]byte(logCfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", logCfg.Level, err)
	}

	// Build encoder config
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Select encoder
	var encoder zapcore.Encoder
	switch logCfg.Encoding {
	case "console":
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", logCfg.Encoding)
	}

	// Configure output writer
	var writer io.Writer
	switch logCfg.Output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// Treat as file path
		if logCfg.FileConfig == nil {
			// Use default file config
			logCfg.FileConfig = &FileConfig{
				Path:       logCfg.Output,
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
			}
		} else {
			logCfg.FileConfig.Path = logCfg.Output
		}

		writer = &lumberjack.Logger{
			Filename:   logCfg.FileConfig.Path,
			MaxSize:    logCfg.FileConfig.MaxSize,
			MaxBackups: logCfg.FileConfig.MaxBackups,
			MaxAge:     logCfg.FileConfig.MaxAge,
			Compress:   logCfg.FileConfig.Compress,
		}
	}

	// Build core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(writer),
		atom,
	)

	// Create logger
	zapCore := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &zapLogger{
		sugar: zapCore.Sugar(),
		atom:  atom,
		core:  zapCore,
	}, nil
}

// Debug logs a debug message with optional fields.
func (l *zapLogger) Debug(msg string, fields ...any) {
	l.sugar.Debugw(msg, fields...)
}

// Info logs an info message with optional fields.
func (l *zapLogger) Info(msg string, fields ...any) {
	l.sugar.Infow(msg, fields...)
}

// Warn logs a warning message with optional fields.
func (l *zapLogger) Warn(msg string, fields ...any) {
	l.sugar.Warnw(msg, fields...)
}

// Error logs an error message with optional fields.
func (l *zapLogger) Error(msg string, fields ...any) {
	l.sugar.Errorw(msg, fields...)
}

// Fatal logs a fatal message with optional fields and exits the process.
func (l *zapLogger) Fatal(msg string, fields ...any) {
	l.sugar.Fatalw(msg, fields...)
}

// With creates a child logger with additional fields.
func (l *zapLogger) With(fields ...any) hyperion.Logger {
	return &zapLogger{
		sugar: l.sugar.With(fields...),
		atom:  l.atom,
		core:  l.core,
	}
}

// WithError creates a child logger with an error field.
func (l *zapLogger) WithError(err error) hyperion.Logger {
	return l.With("error", err)
}

// SetLevel changes the log level dynamically.
func (l *zapLogger) SetLevel(level hyperion.LogLevel) {
	l.atom.SetLevel(toZapLevel(level))
}

// GetLevel returns the current log level.
func (l *zapLogger) GetLevel() hyperion.LogLevel {
	return fromZapLevel(l.atom.Level())
}

// Sync flushes any buffered log entries.
func (l *zapLogger) Sync() error {
	return l.core.Sync()
}

// toZapLevel converts hyperion.LogLevel to zapcore.Level.
func toZapLevel(level hyperion.LogLevel) zapcore.Level {
	switch level {
	case hyperion.DebugLevel:
		return zapcore.DebugLevel
	case hyperion.InfoLevel:
		return zapcore.InfoLevel
	case hyperion.WarnLevel:
		return zapcore.WarnLevel
	case hyperion.ErrorLevel:
		return zapcore.ErrorLevel
	case hyperion.FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// fromZapLevel converts zapcore.Level to hyperion.LogLevel.
func fromZapLevel(level zapcore.Level) hyperion.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return hyperion.DebugLevel
	case zapcore.InfoLevel:
		return hyperion.InfoLevel
	case zapcore.WarnLevel:
		return hyperion.WarnLevel
	case zapcore.ErrorLevel:
		return hyperion.ErrorLevel
	case zapcore.FatalLevel:
		return hyperion.FatalLevel
	default:
		return hyperion.InfoLevel
	}
}
