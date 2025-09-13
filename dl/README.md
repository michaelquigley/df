# dl - Dynamic Logging

**Intelligent channel-based logging built on Go's slog**

The `dl` package provides structured logging with channel-based routing, allowing different log categories to have independent destinations, formats, and configurations.

## Quick Start

```go
import "github.com/michaelquigley/df/dl"

// Basic logging
dl.Log().Info("application started")
dl.Log().With("user", "alice").Info("user logged in")

// Channel-based logging
dl.ChannelLog("database").With("table", "users").Info("query executed")
dl.ChannelLog("http").With("status", 200).Info("request processed")
```

## Key Features

- **📡 Channel Routing**: Route logs to different destinations by category
- **⚙️ Per-Channel Config**: Independent format, level, and output per channel
- **🎨 Multiple Formats**: Pretty console output or structured JSON
- **🔗 Builder Pattern**: Fluent API with `.With()` for contextual attributes
- **🎯 Smart Defaults**: Works immediately, configure only what you need
- **🔄 Thread-Safe**: Concurrent logging across all channels

## Core Functions

- **`dl.Log()`** - Default logger builder
- **`dl.ChannelLog(name)`** - Channel-specific logger builder  
- **`dl.ConfigureChannel(name, opts)`** - Configure channel behavior
- **`dl.DefaultOptions()`** - Create configuration options

## Channel Configuration

**Route channels to different destinations**
```go
// Database logs → file (no color)
dbFile, _ := os.Create("logs/database.log")
dl.ConfigureChannel("database", dl.DefaultOptions().SetOutput(dbFile))

// HTTP logs → stderr as JSON
dl.ConfigureChannel("http", dl.DefaultOptions().JSON().SetOutput(os.Stderr))

// Error logs → console (colored)
dl.ConfigureChannel("errors", dl.DefaultOptions().Color())
```

## Common Patterns

**Contextual Logging**
```go
// Build context with chained attributes
logger := dl.ChannelLog("auth").
    With("user_id", 123).
    With("session", "abc-456")

logger.Info("login attempt")
logger.With("success", true).Info("authentication completed")
```

**Application Integration**
```go
// Initialize with custom defaults
dl.Init(dl.DefaultOptions().SetLevel(slog.LevelDebug).Color())

// Different channels for different concerns
dl.ChannelLog("database").Info("connection established")
dl.ChannelLog("cache").Warn("memory usage high") 
dl.ChannelLog("api").Error("rate limit exceeded")
```

**Format Examples**

*Pretty Console Output:*
```
2024-01-15 14:30:25    INFO user authenticated user_id=123 session=abc-456
2024-01-15 14:30:26   ERROR |database| connection failed host=db.example.com
```

*JSON Output:*
```json
{"time":"2024-01-15T14:30:25Z","level":"INFO","msg":"user authenticated","user_id":123}
{"time":"2024-01-15T14:30:26Z","level":"ERROR","channel":"database","msg":"connection failed"}
```

## Examples

See [examples/](examples/) for tutorials on basic logging, channel routing, and output formatting.

---
*Part of the [df framework](../README.md) - dynamic foundation for Go applications*