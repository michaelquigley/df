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

// With adds a key-value pair to the log context and returns a new builder.
// this allows for fluent chaining of contextual information.
func (b *Builder) With(key string, value any) *Builder {
	return &Builder{
		logger: b.logger,
		attrs:  append(b.attrs, slog.Any(key, value)),
	}
}

// Debug logs a debug message with the accumulated attributes
func (b *Builder) Debug(msg string) {
	if !b.logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Debug]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, fmt.Sprint(msg), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Debugf logs a formatted debug message with the accumulated attributes
func (b *Builder) Debugf(format string, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Debugf]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, fmt.Sprintf(format, args...), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Info logs an info message with the accumulated attributes
func (b *Builder) Info(msg string) {
	if !b.logger.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Info]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprint(msg), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Infof logs a formatted info message with the accumulated attributes
func (b *Builder) Infof(format string, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(format, args...), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Warn logs a warning message with the accumulated attributes
func (b *Builder) Warn(msg string) {
	if !b.logger.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Warn]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprint(msg), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Warnf logs a formatted warning message with the accumulated attributes
func (b *Builder) Warnf(format string, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Warnf]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprintf(format, args...), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Error logs an error message with the accumulated attributes
func (b *Builder) Error(msg string) {
	if !b.logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Error]
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprint(msg), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Errorf logs a formatted error message with the accumulated attributes
func (b *Builder) Errorf(format string, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Errorf]
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
}

// Fatal logs a fatal error message with the accumulated attributes and exits the program
func (b *Builder) Fatal(msg string) {
	if !b.logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Fatal]
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprint(msg), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
	os.Exit(1)
}

// Fatalf logs a formatted fatal error message with the accumulated attributes and exits the program
func (b *Builder) Fatalf(format string, args ...any) {
	if !b.logger.Enabled(context.Background(), slog.LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Fatalf]
	r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), pcs[0])
	for _, attr := range b.attrs {
		r.AddAttrs(attr)
	}
	_ = b.logger.Handler().Handle(context.Background(), r)
	os.Exit(1)
}
