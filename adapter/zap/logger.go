package zap

import (
	"context"
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
	sugar         *zap.SugaredLogger
	atom          zap.AtomicLevel
	core          *zap.Logger
	contextLogger *contextLogger // Context-aware logger for trace correlation
}

// Ensure zapLogger implements hyperion.Logger interface.
var _ hyperion.Logger = (*zapLogger)(nil)

// Ensure zapLogger implements hyperion.ContextAwareLogger interface.
var _ hyperion.ContextAwareLogger = (*zapLogger)(nil)

// Config holds configuration for Zap logger.
// Fields are ordered for optimal memory alignment.
type Config struct {
	OtlpConfig *OtlpLogConfig `mapstructure:"otlp"`     // OTLP logs export configuration (8 bytes pointer)
	FileConfig *FileConfig    `mapstructure:"file"`     // File rotation configuration (8 bytes pointer)
	Level      string         `mapstructure:"level"`    // Log level: debug, info, warn, error, fatal (16 bytes)
	Encoding   string         `mapstructure:"encoding"` // Encoding format: json or console (16 bytes)
	Output     string         `mapstructure:"output"`   // Output destination: stdout, stderr, or file path (16 bytes)
}

// OtlpLogConfig holds OTLP logs export configuration.
type OtlpLogConfig struct {
	Enabled     bool   `mapstructure:"enabled"`      // Whether to enable OTLP logs export
	Endpoint    string `mapstructure:"endpoint"`     // OTLP gRPC endpoint (e.g., "localhost:4317")
	ServiceName string `mapstructure:"service_name"` // Service name for logs
}

// FileConfig holds file rotation configuration.
// Fields are ordered for optimal memory alignment.
type FileConfig struct {
	Path       string `mapstructure:"path"`        // Log file path (16 bytes)
	MaxSize    int    `mapstructure:"max_size"`    // Max file size in MB before rotation (8 bytes)
	MaxBackups int    `mapstructure:"max_backups"` // Max number of old log files to keep (8 bytes)
	MaxAge     int    `mapstructure:"max_age"`     // Max days to retain old log files (8 bytes)
	Compress   bool   `mapstructure:"compress"`    // Whether to compress rotated files (1 byte)
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

	// Configure output writers
	var writers []zapcore.WriteSyncer

	// Add standard output writer
	var stdWriter io.Writer
	switch logCfg.Output {
	case "stdout":
		stdWriter = os.Stdout
	case "stderr":
		stdWriter = os.Stderr
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

		stdWriter = &lumberjack.Logger{
			Filename:   logCfg.FileConfig.Path,
			MaxSize:    logCfg.FileConfig.MaxSize,
			MaxBackups: logCfg.FileConfig.MaxBackups,
			MaxAge:     logCfg.FileConfig.MaxAge,
			Compress:   logCfg.FileConfig.Compress,
		}
	}
	writers = append(writers, zapcore.AddSync(stdWriter))

	// Add OTLP logs exporter if enabled
	var otlpBridge *otlpLogBridge
	if logCfg.OtlpConfig != nil && logCfg.OtlpConfig.Enabled {
		var err error
		otlpBridge, err = createOtlpLogBridge(logCfg.OtlpConfig)
		if err != nil {
			// OTLP is optional - log warning but continue with stdout logging
			// This allows the logger to work even when OTLP collector is unavailable
			fmt.Fprintf(os.Stderr, "[Zap] Warning: failed to create OTLP log bridge: %v\n", err)
			fmt.Fprintf(os.Stderr, "[Zap] Continuing with stdout-only logging\n")
		} else {
			writers = append(writers, otlpBridge)
		}
	}

	// Build core with multi-writer
	var coreWriter zapcore.WriteSyncer
	if len(writers) == 1 {
		coreWriter = writers[0]
	} else {
		coreWriter = zapcore.NewMultiWriteSyncer(writers...)
	}

	core := zapcore.NewCore(
		encoder,
		coreWriter,
		atom,
	)

	// Wrap core with OTel bridge for automatic trace context injection
	otelCore := newOtelCore(core)

	// Create logger with OTel-wrapped core
	zapCore := zap.New(otelCore, zap.AddCaller(), zap.AddCallerSkip(1))

	return &zapLogger{
		sugar:         zapCore.Sugar(),
		atom:          atom,
		core:          zapCore,
		contextLogger: newContextLogger(zapCore),
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
	childCore := l.core.With(convertToZapFields(fields)...)
	return &zapLogger{
		sugar:         l.sugar.With(fields...),
		atom:          l.atom,
		core:          childCore,
		contextLogger: newContextLogger(childCore),
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

// WithContext returns a context-aware logger that automatically injects
// trace context (trace_id and span_id) into all log entries.
// This implements the hyperion.ContextAwareLogger interface.
func (l *zapLogger) WithContext(ctx context.Context) hyperion.Logger {
	return newContextAwareLogger(ctx, l)
}

// levelMapping defines bidirectional mapping between hyperion and zap log levels.
var levelMapping = map[hyperion.LogLevel]zapcore.Level{
	hyperion.DebugLevel: zapcore.DebugLevel,
	hyperion.InfoLevel:  zapcore.InfoLevel,
	hyperion.WarnLevel:  zapcore.WarnLevel,
	hyperion.ErrorLevel: zapcore.ErrorLevel,
	hyperion.FatalLevel: zapcore.FatalLevel,
}

// reverseLevelMapping provides reverse lookup from zap to hyperion levels.
var reverseLevelMapping = map[zapcore.Level]hyperion.LogLevel{
	zapcore.DebugLevel: hyperion.DebugLevel,
	zapcore.InfoLevel:  hyperion.InfoLevel,
	zapcore.WarnLevel:  hyperion.WarnLevel,
	zapcore.ErrorLevel: hyperion.ErrorLevel,
	zapcore.FatalLevel: hyperion.FatalLevel,
}

// toZapLevel converts hyperion.LogLevel to zapcore.Level.
func toZapLevel(level hyperion.LogLevel) zapcore.Level {
	if zapLevel, ok := levelMapping[level]; ok {
		return zapLevel
	}
	return zapcore.InfoLevel // default
}

// fromZapLevel converts zapcore.Level to hyperion.LogLevel.
func fromZapLevel(level zapcore.Level) hyperion.LogLevel {
	if hyperionLevel, ok := reverseLevelMapping[level]; ok {
		return hyperionLevel
	}
	return hyperion.InfoLevel // default
}

// convertToZapFields converts variadic fields to zap.Field slice.
// This helper handles the conversion from sugared fields to structured fields.
func convertToZapFields(fields ...any) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	return zapFields
}
