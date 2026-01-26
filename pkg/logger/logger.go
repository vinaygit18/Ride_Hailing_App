package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger
type Logger struct {
	*zap.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string
	Format string
	Output string
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Choose encoder (json or console)
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &Logger{logger}, nil
}

// Helper methods for common logging patterns

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// With creates a child logger with additional fields
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Common field constructors

// String adds a string field
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int adds an int field
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64 adds an int64 field
func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Float64 adds a float64 field
func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

// Bool adds a bool field
func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// Err adds an error field
func Err(err error) zap.Field {
	return zap.Error(err)
}

// Any adds an arbitrary field
func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// Duration adds a duration field
func Duration(key string, value interface{}) zap.Field {
	if d, ok := value.(interface{ Seconds() float64 }); ok {
		return zap.Float64(key+"_seconds", d.Seconds())
	}
	return zap.Any(key, value)
}
