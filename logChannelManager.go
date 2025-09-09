package df

import (
	"log/slog"
	"os"
	"sync"
)

// LogChannel represents a configured logging channel with its options and logger
type LogChannel struct {
	Logger  *slog.Logger
	Options *LogOptions
}

// LogChannelManager manages per-channel logging with independent destinations
type LogChannelManager struct {
	channels       map[string]*LogChannel
	defaultChannel *LogChannel
	mu             sync.RWMutex
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
		channels: make(map[string]*LogChannel),
		defaultChannel: &LogChannel{
			Logger:  defaultLogger,
			Options: defaultOpts,
		},
	}
}

// GetDefaultLogger returns the default logger
func (cm *LogChannelManager) GetDefaultLogger() *slog.Logger {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.defaultChannel.Logger
}

// GetDefaultOptions returns a copy of the default log options
func (cm *LogChannelManager) GetDefaultOptions() *LogOptions {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.copyLogOptions(cm.defaultChannel.Options)
}

// GetDefaultChannel returns a copy of the default log channel
func (cm *LogChannelManager) GetDefaultChannel() *LogChannel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return &LogChannel{
		Logger:  cm.defaultChannel.Logger,
		Options: cm.copyLogOptions(cm.defaultChannel.Options),
	}
}

// ConfigureDefaultChannel updates the default options and recreates the default logger
func (cm *LogChannelManager) ConfigureDefaultChannel(opts *LogOptions) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if opts == nil {
		opts = DefaultLogOptions()
	}

	// recreate default logger
	var defaultLogger *slog.Logger
	if opts.CustomHandler != nil {
		defaultLogger = slog.New(opts.CustomHandler)
	} else {
		defaultLogger = slog.New(NewDfHandler(opts))
	}

	cm.defaultChannel = &LogChannel{
		Logger:  defaultLogger,
		Options: cm.copyLogOptions(opts),
	}
}

// IsChannelConfigured returns true if the channel has been explicitly configured
func (cm *LogChannelManager) IsChannelConfigured(name string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, exists := cm.channels[name]
	return exists
}

// GetChannelLogger returns a logger for the specified channel
func (cm *LogChannelManager) GetChannelLogger(name string) *slog.Logger {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if channel, exists := cm.channels[name]; exists {
		return channel.Logger
	}

	return cm.defaultChannel.Logger
}

// GetChannelOptions returns a copy of the options for a specific channel
// Returns nil if the channel is not configured
func (cm *LogChannelManager) GetChannelOptions(name string) *LogOptions {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if channel, exists := cm.channels[name]; exists {
		return cm.copyLogOptions(channel.Options)
	}
	return nil
}

// GetChannel returns the full LogChannel for a specific channel
// Returns nil if the channel is not configured
func (cm *LogChannelManager) GetChannel(name string) *LogChannel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if channel, exists := cm.channels[name]; exists {
		return &LogChannel{
			Logger:  channel.Logger,
			Options: cm.copyLogOptions(channel.Options),
		}
	}
	return nil
}

// ConfigureChannel sets a specific logger configuration for a channel
func (cm *LogChannelManager) ConfigureChannel(name string, opts *LogOptions) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if opts == nil {
		opts = DefaultLogOptions()
	}

	handler := cm.createHandlerForChannel(name, opts)
	cm.channels[name] = &LogChannel{
		Logger:  slog.New(handler),
		Options: cm.copyLogOptions(opts),
	}
}

// RemoveChannel removes a channel configuration, causing it to revert to defaults
func (cm *LogChannelManager) RemoveChannel(name string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.channels, name)
}

// ListConfiguredChannels returns the names of all configured channels
func (cm *LogChannelManager) ListConfiguredChannels() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels := make([]string, 0, len(cm.channels))
	for name := range cm.channels {
		channels = append(channels, name)
	}
	return channels
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

// copyLogOptions creates a deep copy of LogOptions
func (cm *LogChannelManager) copyLogOptions(opts *LogOptions) *LogOptions {
	if opts == nil {
		return nil
	}

	out := *opts
	// note: this assumes LogOptions fields are either value types or
	// interfaces that don't need deep copying. if LogOptions contains
	// pointer fields that should be deep copied, add that logic here.
	return &out
}
