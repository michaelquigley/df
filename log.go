package df

import (
	"log/slog"
)

var defaultChannelManager *LogChannelManager

// InitLogging initializes the logging system with the provided options.
// if no options are provided, uses default options.
func InitLogging(opts ...*LogOptions) {
	var options *LogOptions
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultLogOptions()
	}

	// initialize channel manager (which creates the default logger internally)
	defaultChannelManager = NewLogChannelManager(options)
}

// Log returns a general logger builder for adding contextual attributes
func Log() *LogBuilder {
	ensureInit()
	return &LogBuilder{logger: defaultChannelManager.GetDefaultLogger()}
}

// ChannelLog creates a logger with a specific channel attribute for categorizing log entries
func ChannelLog(name string) *LogBuilder {
	ensureInit()

	logger := defaultChannelManager.GetChannelLogger(name)

	// if this channel is not configured (using default logger), add channel attribute for backward compatibility
	if !defaultChannelManager.IsChannelConfigured(name) {
		return &LogBuilder{
			logger: logger,
			attrs:  []slog.Attr{slog.String(ChannelKey, name)},
		}
	}

	// configured channels have their own loggers with built-in channel names
	return &LogBuilder{logger: logger}
}

// ConfigureChannel sets a specific logger configuration for a channel
func ConfigureChannel(name string, opts *LogOptions) {
	ensureInit()
	defaultChannelManager.ConfigureChannel(name, opts)
}

// ReconfigureChannel reconfigures an existing channel (alias for ConfigureChannel)
func ReconfigureChannel(name string, opts *LogOptions) {
	ensureInit()
	ConfigureChannel(name, opts)
}

// RemoveChannel removes a channel configuration, causing it to revert to defaults
func RemoveChannel(name string) {
	ensureInit()
	defaultChannelManager.RemoveChannel(name)
}

func ensureInit() {
	if defaultChannelManager == nil {
		InitLogging()
	}
}
