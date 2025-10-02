package hyperion

// Logger is the structured logging abstraction.
// Implementations should provide thread-safe logging capabilities.
type Logger interface {
	// Debug logs a debug-level message with optional key-value fields.
	Debug(msg string, fields ...any)

	// Info logs an info-level message with optional key-value fields.
	Info(msg string, fields ...any)

	// Warn logs a warning-level message with optional key-value fields.
	Warn(msg string, fields ...any)

	// Error logs an error-level message with optional key-value fields.
	Error(msg string, fields ...any)

	// Fatal logs a fatal-level message and exits the application.
	Fatal(msg string, fields ...any)

	// With returns a new Logger with additional fields.
	// The original logger is not modified.
	With(fields ...any) Logger

	// WithError returns a new Logger with an error field.
	WithError(err error) Logger

	// SetLevel sets the minimum log level.
	SetLevel(level LogLevel)

	// GetLevel returns the current log level.
	GetLevel() LogLevel

	// Sync flushes any buffered log entries.
	// Applications should call Sync before exiting.
	Sync() error
}

// LogLevel represents the severity level of a log message.
type LogLevel int

const (
	// DebugLevel logs are typically voluminous and are usually disabled in production.
	DebugLevel LogLevel = iota

	// InfoLevel is the default logging priority.
	InfoLevel

	// WarnLevel logs are more important than Info, but don't need individual human review.
	WarnLevel

	// ErrorLevel logs are high-priority errors that should be reviewed.
	ErrorLevel

	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}
