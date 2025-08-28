package df

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

var defaultLogger *slog.Logger

// InitLogging initializes the default logger with the provided options.
// if no options are provided, uses default options.
func InitLogging(opts ...*LogOptions) {
	var options *LogOptions
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultLogOptions()
	}

	handler := NewDfHandler(options)
	defaultLogger = slog.New(handler)
}

// Debug logs a debug message with optional key-value pairs
func Debug(msg string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Debug]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, fmt.Sprint(msg), pcs[0])
	addArgsToRecord(&r, args...)
	_ = defaultLogger.Handler().Handle(context.Background(), r)
}

// Info logs an info message with optional key-value pairs
func Info(msg string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Info]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprint(msg), pcs[0])
	addArgsToRecord(&r, args...)
	_ = defaultLogger.Handler().Handle(context.Background(), r)
}

// Warn logs a warning message with optional key-value pairs
func Warn(msg string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Warn]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprint(msg), pcs[0])
	addArgsToRecord(&r, args...)
	_ = defaultLogger.Handler().Handle(context.Background(), r)
}

// Error logs an error message with optional key-value pairs
func Error(msg string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Error]
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprint(msg), pcs[0])
	addArgsToRecord(&r, args...)
	_ = defaultLogger.Handler().Handle(context.Background(), r)
}

// Fatal logs a fatal error message and exits the program
func Fatal(msg string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Fatal]
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprint(msg), pcs[0])
	addArgsToRecord(&r, args...)
	_ = defaultLogger.Handler().Handle(context.Background(), r)
	os.Exit(1)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Debugf]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, fmt.Sprintf(format, args...), pcs[0])
	_ = defaultLogger.Handler().Handle(context.Background(), r)
}

// Infof logs a formatted info message
func Infof(format string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(format, args...), pcs[0])
	_ = defaultLogger.Handler().Handle(context.Background(), r)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Warnf]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprintf(format, args...), pcs[0])
	_ = defaultLogger.Handler().Handle(context.Background(), r)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Errorf]
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), pcs[0])
	_ = defaultLogger.Handler().Handle(context.Background(), r)
}

// Fatalf logs a formatted fatal error message and exits the program
func Fatalf(format string, args ...any) {
	ensureLogger()
	if !defaultLogger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Fatalf]
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), pcs[0])
	_ = defaultLogger.Handler().Handle(context.Background(), r)
	os.Exit(1)
}

// Logger returns a general logger builder for adding contextual attributes
func Logger() *LogBuilder {
	ensureLogger()
	return &LogBuilder{logger: defaultLogger}
}

// LoggerChannel creates a logger with a specific channel attribute for categorizing log entries
func LoggerChannel(name string) *LogBuilder {
	ensureLogger()
	return &LogBuilder{
		logger: defaultLogger,
		attrs:  []slog.Attr{slog.String("channel", name)},
	}
}

func ensureLogger() {
	if defaultLogger == nil {
		InitLogging()
	}
}

// addArgsToRecord converts key-value pairs to slog attributes and adds them to the record
func addArgsToRecord(r *slog.Record, args ...any) {
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := fmt.Sprintf("%v", args[i])
			value := args[i+1]
			r.AddAttrs(slog.Any(key, value))
		}
	}
}
