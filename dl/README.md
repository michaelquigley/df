# DLF - Dynamic Logging Foundation

This package contains the logging layer extracted from the main `df` library.

## Overview

The `dlf` package provides a comprehensive logging foundation with the following features:

- **Fluent Builder API**: `Builder` type with method chaining for contextual logging
- **Channel-based Logging**: Support for multiple named logging channels with independent configurations
- **Multiple Output Formats**: Pretty-printed and JSON output formats
- **Flexible Configuration**: Comprehensive options for colors, timestamps, output destinations
- **Thread-safe**: All operations are safe for concurrent use

## Key Types

- **`Builder`**: Fluent logging interface with contextual attributes
- **`Options`**: Configuration options for loggers and handlers
- **`ChannelManager`**: Manages multiple named logging channels
- **`Channel`**: Individual logging channel with its own configuration
- **`PrettyHandler`**: Custom slog handler for human-readable output

## Key Functions

- **`Init(opts)`**: Initialize the logging system
- **`Log()`**: Get the default logger builder
- **`ChannelLog(name)`**: Get a channel-specific logger builder
- **`ConfigureChannel(name, opts)`**: Configure a specific channel
- **`DefaultOptions()`**: Create default configuration options

## Example Usage

```go
import "github.com/michaelquigley/df/dl"

// Initialize logging
dlf.Init(dlf.DefaultOptions().SetLevel(slog.LevelDebug))

// Basic logging
dlf.Log().Info("application started")

// Contextual logging
dlf.Log().With("user", "alice").With("action", "login").Info("user action completed")

// Channel-based logging
dlf.ChannelLog("auth").Info("authentication successful")
dlf.ChannelLog("database").Warn("connection pool running low")

// Configure channels
opts := dlf.DefaultOptions().JSON().SetLevel(slog.LevelError)
dlf.ConfigureChannel("errors", opts)
```