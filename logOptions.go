package df

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

// LogOptions configures the df logging system, compatible with pfxlog.Options
type LogOptions struct {
	Level           slog.Level
	UseJSON         bool
	UseColor        bool
	AbsoluteTime    bool
	StartTimestamp  time.Time
	TimestampFormat string
	TrimPrefix      string

	// level labels - matching pfxlog names
	ErrorLabel   string
	WarningLabel string // pfxlog uses WarningLabel, not WarnLabel
	InfoLabel    string
	DebugLabel   string

	// colors
	TimestampColor string
	FunctionColor  string
	ChannelColor   string
	FieldsColor    string
	DefaultFgColor string // pfxlog uses this for reset
	ErrorColor     string
	WarningColor   string
	InfoColor      string
	DebugColor     string
}

// DefaultLogOptions creates a default configuration with sensible defaults
func DefaultLogOptions() *LogOptions {
	out := &LogOptions{
		Level:           slog.LevelInfo,
		UseColor:        isTerminal() && shouldUseColor(),
		TimestampFormat: "2006-01-02 15:04:05.000",
		StartTimestamp:  time.Now(),
		ErrorLabel:      "ERROR",
		WarningLabel:    "WARN ",
		InfoLabel:       "INFO ",
		DebugLabel:      "DEBUG",
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

// Color enables colored output with default color scheme
func (o *LogOptions) Color() *LogOptions {
	o.UseColor = true
	return o
}

// NoColor disables colored output
func (o *LogOptions) NoColor() *LogOptions {
	o.UseColor = false
	return o
}

// JSON enables JSON output format
func (o *LogOptions) JSON() *LogOptions {
	o.UseJSON = true
	return o
}

// Pretty enables pretty-printed output format (default)
func (o *LogOptions) Pretty() *LogOptions {
	o.UseJSON = false
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
	if env := os.Getenv("PFXLOG_USE_COLOR"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			return val
		}
	}
	return true
}

// getDefaultFgColor returns the default foreground color (reset sequence)
func (o *LogOptions) getDefaultFgColor() string {
	if o.UseColor {
		return o.DefaultFgColor
	}
	return ""
}
