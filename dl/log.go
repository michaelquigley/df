package dl

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

var defaultChannelManager *ChannelManager

// Init initializes the logging system with the provided options.
// if no options are provided, uses default options.
func Init(opts ...*Options) {
	var options *Options
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultOptions()
	}

	// initialize channel manager (which creates the default logger internally)
	defaultChannelManager = NewChannelManager(options)
}

// ConfigureChannel sets a specific logger configuration for a channel
func ConfigureChannel(name string, opts *Options) {
	ensureInit()
	defaultChannelManager.ConfigureChannel(name, opts)
}

// RemoveChannel removes a channel configuration, causing it to revert to defaults
func RemoveChannel(name string) {
	ensureInit()
	defaultChannelManager.RemoveChannel(name)
}

// Debug logs a debug message using the default logger
func Debug(msg any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Debug]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

// Debugf logs a formatted debug message using the default logger
func Debugf(format any, args ...any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Debugf]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

// Info logs an info message using the default logger
func Info(msg any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Info]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

// Infof logs a formatted info message using the default logger
func Infof(format any, args ...any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

// Warn logs a warning message using the default logger
func Warn(msg any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Warn]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

// Warnf logs a formatted warning message using the default logger
func Warnf(format any, args ...any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Warnf]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

// Error logs an error message using the default logger
func Error(msg any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Error]
	r := slog.NewRecord(time.Now(), slog.LevelError, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

// Errorf logs a formatted error message using the default logger
func Errorf(format any, args ...any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Errorf]
	r := slog.NewRecord(time.Now(), slog.LevelError, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}

// Fatal logs a fatal error message using the default logger and exits the program
func Fatal(msg any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Fatal]
	r := slog.NewRecord(time.Now(), slog.LevelError, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
	os.Exit(1)
}

// Fatalf logs a formatted fatal error message using the default logger and exits the program
func Fatalf(format any, args ...any) {
	ensureInit()
	logger := defaultChannelManager.GetDefaultLogger()
	if !logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Fatalf]
	r := slog.NewRecord(time.Now(), slog.LevelError, message, pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
	os.Exit(1)
}

// Log returns a general logger builder for adding contextual attributes
func Log() *Builder {
	ensureInit()
	return &Builder{logger: defaultChannelManager.GetDefaultLogger()}
}

// ChannelLog creates a logger with a specific channel attribute for categorizing log entries
func ChannelLog(name string) *Builder {
	ensureInit()

	logger := defaultChannelManager.GetChannelLogger(name)

	// if this channel is not configured (using default logger), add channel attribute for backward compatibility
	if !defaultChannelManager.IsChannelConfigured(name) {
		return &Builder{
			logger: logger,
			attrs:  []slog.Attr{slog.String(ChannelKey, name)},
		}
	}

	// configured channels have their own loggers with built-in channel names
	return &Builder{logger: logger}
}

func ensureInit() {
	if defaultChannelManager == nil {
		Init()
	}
}
