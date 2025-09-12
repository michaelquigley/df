package dl

import (
	"log/slog"
	"os"
	"sync"
)

// Channel represents a configured logging channel with its options and logger
type Channel struct {
	Logger  *slog.Logger
	Options *Options
}

// ChannelManager manages per-channel logging with independent destinations
type ChannelManager struct {
	channels       map[string]*Channel
	defaultChannel *Channel
	mu             sync.RWMutex
}

// NewChannelManager creates a new channel log manager
func NewChannelManager(defaultOpts *Options) *ChannelManager {
	if defaultOpts == nil {
		defaultOpts = DefaultOptions()
	}

	// create the default logger
	var defaultLogger *slog.Logger
	if defaultOpts.CustomHandler != nil {
		defaultLogger = slog.New(defaultOpts.CustomHandler)
	} else {
		defaultLogger = slog.New(NewDfHandler(defaultOpts))
	}

	return &ChannelManager{
		channels: make(map[string]*Channel),
		defaultChannel: &Channel{
			Logger:  defaultLogger,
			Options: defaultOpts,
		},
	}
}

// GetDefaultLogger returns the default logger
func (cm *ChannelManager) GetDefaultLogger() *slog.Logger {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.defaultChannel.Logger
}

// GetDefaultOptions returns a copy of the default log options
func (cm *ChannelManager) GetDefaultOptions() *Options {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.copyOptions(cm.defaultChannel.Options)
}

// GetDefaultChannel returns a copy of the default log channel
func (cm *ChannelManager) GetDefaultChannel() *Channel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return &Channel{
		Logger:  cm.defaultChannel.Logger,
		Options: cm.copyOptions(cm.defaultChannel.Options),
	}
}

// ConfigureDefaultChannel updates the default options and recreates the default logger
func (cm *ChannelManager) ConfigureDefaultChannel(opts *Options) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if opts == nil {
		opts = DefaultOptions()
	}

	// recreate default logger
	var defaultLogger *slog.Logger
	if opts.CustomHandler != nil {
		defaultLogger = slog.New(opts.CustomHandler)
	} else {
		defaultLogger = slog.New(NewDfHandler(opts))
	}

	cm.defaultChannel = &Channel{
		Logger:  defaultLogger,
		Options: cm.copyOptions(opts),
	}
}

// IsChannelConfigured returns true if the channel has been explicitly configured
func (cm *ChannelManager) IsChannelConfigured(name string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, exists := cm.channels[name]
	return exists
}

// GetChannelLogger returns a logger for the specified channel
func (cm *ChannelManager) GetChannelLogger(name string) *slog.Logger {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if channel, exists := cm.channels[name]; exists {
		return channel.Logger
	}

	return cm.defaultChannel.Logger
}

// GetChannelOptions returns a copy of the options for a specific channel
// Returns nil if the channel is not configured
func (cm *ChannelManager) GetChannelOptions(name string) *Options {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if channel, exists := cm.channels[name]; exists {
		return cm.copyOptions(channel.Options)
	}
	return nil
}

// GetChannel returns the full Channel for a specific channel
// Returns nil if the channel is not configured
func (cm *ChannelManager) GetChannel(name string) *Channel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if channel, exists := cm.channels[name]; exists {
		return &Channel{
			Logger:  channel.Logger,
			Options: cm.copyOptions(channel.Options),
		}
	}
	return nil
}

// ConfigureChannel sets a specific logger configuration for a channel
func (cm *ChannelManager) ConfigureChannel(name string, opts *Options) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if opts == nil {
		opts = DefaultOptions()
	}

	handler := cm.createHandlerForChannel(name, opts)
	cm.channels[name] = &Channel{
		Logger:  slog.New(handler),
		Options: cm.copyOptions(opts),
	}
}

// RemoveChannel removes a channel configuration, causing it to revert to defaults
func (cm *ChannelManager) RemoveChannel(name string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.channels, name)
}

// ListConfiguredChannels returns the names of all configured channels
func (cm *ChannelManager) ListConfiguredChannels() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels := make([]string, 0, len(cm.channels))
	for name := range cm.channels {
		channels = append(channels, name)
	}
	return channels
}

// createHandlerForChannel creates a handler for a specific channel
func (cm *ChannelManager) createHandlerForChannel(channelName string, opts *Options) slog.Handler {
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

// copyOptions creates a deep copy of Options
func (cm *ChannelManager) copyOptions(opts *Options) *Options {
	if opts == nil {
		return nil
	}

	out := *opts
	// note: this assumes Options fields are either value types or
	// interfaces that don't need deep copying. if Options contains
	// pointer fields that should be deep copied, add that logic here.
	return &out
}
