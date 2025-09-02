package df

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const ChannelKey = "channel"

// NewDfHandler creates a handler that supports both pretty and JSON modes
func NewDfHandler(opts *LogOptions) slog.Handler {
	if opts == nil {
		opts = DefaultLogOptions()
	}

	if opts.UseJSON {
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     opts.Level,
			AddSource: true,
		})
	}

	return NewPrettyHandler(opts.Level, opts)
}

// PrettyHandler is a direct port of pfxlog's PrettyHandler for df
type PrettyHandler struct {
	level   slog.Level
	options *LogOptions
	lock    sync.Mutex
	attrs   []slog.Attr
}

// NewPrettyHandler creates a new pretty handler - direct port from pfxlog
func NewPrettyHandler(level slog.Level, options *LogOptions) slog.Handler {
	return &PrettyHandler{level: level, options: options}
}

// Enabled implements slog.Handler.Enabled
func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle implements slog.Handler.Handle - direct port from pfxlog
func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	var out strings.Builder

	var timeLabel string
	if h.options.AbsoluteTime {
		timeLabel = "[" + time.Now().Format(h.options.TimestampFormat) + "]"
	} else {
		seconds := time.Since(h.options.StartTimestamp).Seconds()
		timeLabel = fmt.Sprintf("[%8.3f]", seconds)
	}
	out.WriteString(h.options.TimestampColor + timeLabel + h.options.getDefaultFgColor())

	var level string
	switch r.Level {
	case slog.LevelError:
		level = h.options.ErrorLabel
	case slog.LevelWarn:
		level = h.options.WarningLabel
	case slog.LevelInfo:
		level = h.options.InfoLabel
	case slog.LevelDebug:
		level = h.options.DebugLabel
	}
	out.WriteString(" " + level)

	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	functionStr := f.Function
	if h.options.TrimPrefix != "" {
		functionStr = strings.TrimPrefix(functionStr, h.options.TrimPrefix)
	}
	out.WriteString(" " + h.options.FunctionColor + functionStr + h.options.getDefaultFgColor())

	// collect handler attributes
	allAttrs := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	allAttrs = append(allAttrs, h.attrs...)

	// collect record attributes
	fieldsMap := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		allAttrs = append(allAttrs, a)
		return true
	})

	// process all attributes
	for _, a := range allAttrs {
		if a.Key != ChannelKey {
			fieldsMap[a.Key] = a.Value.Any()
		} else {
			out.WriteString(h.options.ChannelColor + " |" + a.Value.String() + "|" + h.options.getDefaultFgColor())
		}
	}

	fieldsBytes, err := json.Marshal(fieldsMap)
	if err != nil {
		return err
	}
	if len(fieldsBytes) > 2 {
		out.WriteString(" " + h.options.FieldsColor + string(fieldsBytes) + h.options.getDefaultFgColor())
	}

	out.WriteString(" " + r.Message)

	h.lock.Lock()
	fmt.Println(out.String())
	h.lock.Unlock()

	return nil
}

// WithAttrs implements slog.Handler.WithAttrs
func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{level: h.level, options: h.options, attrs: attrs}
}

// WithGroup implements slog.Handler.WithGroup
func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return h
}
