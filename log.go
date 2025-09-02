package df

import (
	"log/slog"
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

	if options.OverrideHandler != nil {
		defaultLogger = slog.New(options.OverrideHandler)
	} else {
		defaultLogger = slog.New(NewDfHandler(options))
	}
}

// Log returns a general logger builder for adding contextual attributes
func Log() *LogBuilder {
	ensureInit()
	return &LogBuilder{logger: defaultLogger}
}

// ChannelLog creates a logger with a specific channel attribute for categorizing log entries
func ChannelLog(name string) *LogBuilder {
	ensureInit()
	return &LogBuilder{
		logger: defaultLogger,
		attrs:  []slog.Attr{slog.String("channel", name)},
	}
}

func ensureInit() {
	if defaultLogger == nil {
		InitLogging()
	}
}
