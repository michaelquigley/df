package df

import (
	"fmt"
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

	handler := NewDfHandler(options)
	defaultLogger = slog.New(handler)
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
