package df

import (
	"log/slog"
	"os"
	"sync"
)

// LogChannelManager manages per-channel logging with independent destinations
type LogChannelManager struct {
	channels      map[string]*slog.Logger
	defaultLogger *slog.Logger
	mu            sync.RWMutex
}

// NewLogChannelManager creates a new channel log manager
func NewLogChannelManager(defaultOpts *LogOptions) *LogChannelManager {
	if defaultOpts == nil {
		defaultOpts = DefaultLogOptions()
	}

	// create the default logger
	var defaultLogger *slog.Logger
	if defaultOpts.CustomHandler != nil {
		defaultLogger = slog.New(defaultOpts.CustomHandler)
	} else {
		defaultLogger = slog.New(NewDfHandler(defaultOpts))
	}

	return &LogChannelManager{
		channels:      make(map[string]*slog.Logger),
		defaultLogger: defaultLogger,
	}
}

// GetDefaultLogger returns the default logger
func (cm *LogChannelManager) GetDefaultLogger() *slog.Logger {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.defaultLogger
}

// IsChannelConfigured returns true if the channel has been explicitly configured
func (cm *LogChannelManager) IsChannelConfigured(name string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, exists := cm.channels[name]
	return exists
}

// GetChannelLogger returns a logger for the specified channel
// If the channel is configured, returns the configured logger
// If unconfigured, returns the default logger
func (cm *LogChannelManager) GetChannelLogger(name string) *slog.Logger {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if logger, exists := cm.channels[name]; exists {
		return logger
	}

	// return default logger for unconfigured channels
	return cm.defaultLogger
}

// ConfigureChannel sets a specific logger configuration for a channel
func (cm *LogChannelManager) ConfigureChannel(name string, opts *LogOptions) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if opts == nil {
		opts = DefaultLogOptions()
	}

	handler := cm.createHandlerForChannel(name, opts)
	cm.channels[name] = slog.New(handler)
}

// RemoveChannel removes a channel configuration, causing it to revert to defaults
func (cm *LogChannelManager) RemoveChannel(name string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.channels, name)
}

// createHandlerForChannel creates a handler for a specific channel
func (cm *LogChannelManager) createHandlerForChannel(channelName string, opts *LogOptions) slog.Handler {
	if opts.CustomHandler != nil {
		return opts.CustomHandler
	}

	output := opts.Output
	if output == nil {
		output = os.Stdout
	}

	if opts.UseJSON {
		return slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level:     opts.Level,
			AddSource: true,
		})
	}

	return NewPrettyHandlerWithChannel(opts.Level, opts, channelName)
}
