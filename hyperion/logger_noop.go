package hyperion

// noopLogger is a no-op implementation of Logger interface.
type noopLogger struct {
	level LogLevel
}

// NewNoOpLogger creates a new no-op Logger implementation.
func NewNoOpLogger() Logger {
	return &noopLogger{level: InfoLevel}
}

func (l *noopLogger) Debug(msg string, fields ...any) {}
func (l *noopLogger) Info(msg string, fields ...any)  {}
func (l *noopLogger) Warn(msg string, fields ...any)  {}
func (l *noopLogger) Error(msg string, fields ...any) {}
func (l *noopLogger) Fatal(msg string, fields ...any) {}
func (l *noopLogger) With(fields ...any) Logger       { return l }
func (l *noopLogger) WithError(err error) Logger      { return l }
func (l *noopLogger) SetLevel(level LogLevel)         { l.level = level }
func (l *noopLogger) GetLevel() LogLevel              { return l.level }
func (l *noopLogger) Sync() error                     { return nil }
