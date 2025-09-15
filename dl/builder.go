package dl

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Builder provides a fluent API for contextual logging, allowing attributes to be added
// before logging messages. preserves pfxlog's builder pattern semantics.
type Builder struct {
	logger *slog.Logger
	attrs  []slog.Attr
}

// convertMessage converts any type to a string for logging
func convertMessage(msg any) string {
	switch v := msg.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return fmt.Sprint(v)
	}
}

// convertFormattedMessage converts a format and args to a formatted string for logging
func convertFormattedMessage(format any, args ...any) string {
	switch v := format.(type) {
	case string:
		return fmt.Sprintf(v, args...)
	case error:
		return v.Error()
	default:
		return fmt.Sprint(v)
	}
}

// With adds a key-value pair to the log context and returns a new builder.
// this allows for fluent chaining of contextual information.
func (b *Builder) With(key string, value any) *Builder {
	return &Builder{
		logger: b.logger,
		attrs:  append(b.attrs, slog.Any(key, value)),
	}
}

// Debug logs a debug message with the accumulated attributes
func (b *Builder) Debug(msg any) {
	if !b.logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Debug]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Debugf logs a formatted debug message with the accumulated attributes
func (b *Builder) Debugf(format any, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Debugf]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Info logs an info message with the accumulated attributes
func (b *Builder) Info(msg any) {
	if !b.logger.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Info]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Infof logs a formatted info message with the accumulated attributes
func (b *Builder) Infof(format any, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Warn logs a warning message with the accumulated attributes
func (b *Builder) Warn(msg any) {
	if !b.logger.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Warn]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Warnf logs a formatted warning message with the accumulated attributes
func (b *Builder) Warnf(format any, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Warnf]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Error logs an error message with the accumulated attributes
func (b *Builder) Error(msg any) {
	if !b.logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Error]
	r := slog.NewRecord(time.Now(), slog.LevelError, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Errorf logs a formatted error message with the accumulated attributes
func (b *Builder) Errorf(format any, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Errorf]
	r := slog.NewRecord(time.Now(), slog.LevelError, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Fatal logs a fatal error message with the accumulated attributes and exits the program
func (b *Builder) Fatal(msg any) {
	if !b.logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	message := convertMessage(msg)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Fatal]
	r := slog.NewRecord(time.Now(), slog.LevelError, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
	os.Exit(1)
}

// Fatalf logs a formatted fatal error message with the accumulated attributes and exits the program
func (b *Builder) Fatalf(format any, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	message := convertFormattedMessage(format, args...)
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Fatalf]
	r := slog.NewRecord(time.Now(), slog.LevelError, message, pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
	os.Exit(1)
}
