package dl

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Options configures the df logging system, compatible with pfxlog.Options
type Options struct {
	Level           slog.Level
	UseJSON         bool
	UseColor        bool
	AbsoluteTime    bool
	StartTimestamp  time.Time
	TimestampFormat string
	TrimPrefix      string
	Output          io.Writer // output destination, defaults to os.Stdout
	CustomHandler   slog.Handler

	// level labels
	ErrorLabel   string
	WarningLabel string
	InfoLabel    string
	DebugLabel   string

	// colors
	TimestampColor string
	FunctionColor  string
	ChannelColor   string
	FieldsColor    string
	DefaultFgColor string // used for resetting colors
	ErrorColor     string
	WarningColor   string
	InfoColor      string
	DebugColor     string
}

// DefaultOptions creates a default configuration with sensible defaults
func DefaultOptions() *Options {
	out := &Options{
		Level:           slog.LevelInfo,
		UseColor:        isTerminal() && shouldUseColor(),
		TimestampFormat: "2006-01-02 15:04:05.000",
		StartTimestamp:  time.Now(),
		Output:          os.Stdout,
		ErrorLabel:      "  ERROR",
		WarningLabel:    "WARNING",
		InfoLabel:       "   INFO",
		DebugLabel:      "  DEBUG",
		TimestampColor:  "\033[90m", // dark gray
		FunctionColor:   "\033[36m", // cyan
		ChannelColor:    "\033[35m", // magenta
		FieldsColor:     "\033[33m", // yellow
		DefaultFgColor:  "\033[0m",  // reset
		ErrorColor:      "\033[31m", // red
		WarningColor:    "\033[33m", // yellow
		InfoColor:       "\033[37m", // white
		DebugColor:      "\033[34m", // blue
	}
	if out.UseColor && !out.UseJSON {
		out.ErrorLabel = out.ErrorColor + out.ErrorLabel + out.DefaultFgColor
		out.WarningLabel = out.WarningColor + out.WarningLabel + out.DefaultFgColor
		out.InfoLabel = out.InfoColor + out.InfoLabel + out.DefaultFgColor
		out.DebugLabel = out.DebugColor + out.DebugLabel + out.DefaultFgColor
	}
	return out
}

// SetTrimPrefix sets the function trim prefix
func (o *Options) SetTrimPrefix(prefix string) *Options {
	o.TrimPrefix = prefix
	return o
}

// SetLevel allows setting the level threshold
func (o *Options) SetLevel(level slog.Level) *Options {
	o.Level = level
	return o
}

// Color enables colored output with default color scheme
func (o *Options) Color() *Options {
	o.UseColor = true
	return o
}

// NoColor disables colored output
func (o *Options) NoColor() *Options {
	o.UseColor = false
	return o
}

// JSON enables JSON output format
func (o *Options) JSON() *Options {
	o.UseJSON = true
	return o
}

// Pretty enables pretty-printed output format (default)
func (o *Options) Pretty() *Options {
	o.UseJSON = false
	return o
}

// SetOutput sets the output destination
func (o *Options) SetOutput(w io.Writer) *Options {
	o.Output = w
	return o
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// shouldUseColor checks environment variables to determine if color should be used
func shouldUseColor() bool {
	if env := os.Getenv("DFLOG_USE_COLOR"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			return val
		}
	}
	return true
}

// getDefaultFgColor returns the default foreground color (reset sequence)
func (o *Options) getDefaultFgColor() string {
	if o.UseColor {
		return o.DefaultFgColor
	}
	return ""
}
