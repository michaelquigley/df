---
title: dynamic foundation for logging
description: Complete reference for the dl package - intelligent channel-based logging built on Go's slog.
---

**Intelligent channel-based logging built on Go's slog**

The `dl` package provides structured logging with channel-based routing, allowing different log categories to have independent destinations, formats, and configurations. Built on Go's standard `slog` package for performance and compatibility.

## Quick Reference

### 1. Basic Logging - Hello World

**Simple structured logging**

```go
import "github.com/michaelquigley/df/dl"

// Basic logging
dl.Log().Info("application started")
dl.Log().Error("connection failed")

// Contextual logging with attributes
dl.Log().With("user", "alice").Info("user logged in")
dl.Log().With("port", 8080).With("host", "localhost").Info("server listening")

// Builder pattern for complex context
logger := dl.Log().With("request_id", "req-123").With("user_id", "user-456")
logger.Info("processing request")
logger.With("duration", "250ms").Info("request completed")
```

### 2. Formatted Logging - Printf Style

**Traditional printf-style logging**

```go
// Formatted messages
dl.Log().Infof("server started on port %d", 8080)
dl.Log().Errorf("connection failed after %d attempts", retries)
dl.Log().Warnf("disk usage at %d%% on %s", usage, disk)

// With context
dl.Log().With("user", "alice").Infof("login attempt #%d", attempts)
```

### 3. Channel Routing - Categorized Logging

**Route different categories to different destinations**

```go
// Different categories use different channels
dl.ChannelLog("database").Info("connection established")
dl.ChannelLog("http").With("method", "GET").With("path", "/api/users").Info("request")
dl.ChannelLog("auth").With("user", "alice").Info("authentication successful")
dl.ChannelLog("errors").With("code", 500).Error("internal server error")

// Channels work with builder pattern
dbLogger := dl.ChannelLog("database").With("connection_id", "conn-123")
dbLogger.Info("query started")
dbLogger.With("duration", "45ms").Info("query completed")
```

### 4. Channel Configuration - Independent Settings

**Configure each channel independently**

```go
import "os"

// Database logs → file (no color)
dbFile, _ := os.Create("logs/database.log")
dl.ConfigureChannel("database", dl.DefaultOptions().SetOutput(dbFile))

// HTTP logs → stderr in JSON format
dl.ConfigureChannel("http", dl.DefaultOptions().JSON().SetOutput(os.Stderr))

// Error logs → console with colors
dl.ConfigureChannel("errors", dl.DefaultOptions().Color())

// Now each channel routes to its configured destination
dl.ChannelLog("database").Info("logged to file")      // → logs/database.log
dl.ChannelLog("http").Info("logged as JSON")          // → stderr (JSON)
dl.ChannelLog("errors").Error("logged with colors")   // → console (colored)
```

### 5. Output Formats - Pretty and JSON

**Different output formats for different needs**

```go
// Pretty format (human-readable, default for console)
dl.ConfigureChannel("console", dl.DefaultOptions().Pretty().Color())

// JSON format (machine-readable, for log aggregation)
dl.ConfigureChannel("api", dl.DefaultOptions().JSON())

// No color (for files)
logFile, _ := os.Create("app.log")
dl.ConfigureChannel("file", dl.DefaultOptions().Pretty().NoColor().SetOutput(logFile))

// Examples of output:
// Pretty: 2024-01-15 14:30:25    INFO user authenticated user_id=123
// JSON:   {"time":"2024-01-15T14:30:25Z","level":"INFO","msg":"user authenticated","user_id":123}
```

### 6. Log Levels - Filtering

**Control verbosity with log levels**

```go
import "log/slog"

// Configure different levels per channel
dl.ConfigureChannel("debug", dl.DefaultOptions().SetLevel(slog.LevelDebug))
dl.ConfigureChannel("errors", dl.DefaultOptions().SetLevel(slog.LevelError))

// Global initialization with level
dl.Init(dl.DefaultOptions().SetLevel(slog.LevelInfo))

// All standard levels available
dl.Log().Debug("debug message")    // only shown if level <= Debug
dl.Log().Info("info message")      // shown if level <= Info
dl.Log().Warn("warning message")   // shown if level <= Warn
dl.Log().Error("error message")    // shown if level <= Error
```

### 7. Multiple Destinations - Flexible Routing

**Route to multiple destinations simultaneously**

```go
// Create multiple outputs
consoleOut := os.Stdout
fileOut, _ := os.Create("app.log")
errorOut, _ := os.Create("errors.log")

// Configure different channels for different purposes
dl.ConfigureChannel("general", dl.DefaultOptions().Pretty().SetOutput(consoleOut))
dl.ConfigureChannel("audit", dl.DefaultOptions().JSON().SetOutput(fileOut))
dl.ConfigureChannel("alerts", dl.DefaultOptions().Color().SetOutput(errorOut))

// Route events appropriately
dl.ChannelLog("general").Info("application event")     // → console (pretty)
dl.ChannelLog("audit").Info("user action logged")      // → app.log (JSON)
dl.ChannelLog("alerts").Error("critical failure")      // → errors.log (colored)
```

### 8. Initialization - Global Settings

**Configure default behavior**

```go
// Initialize with custom defaults
dl.Init(dl.DefaultOptions().
    SetLevel(slog.LevelDebug).
    Color().
    SetTrimPrefix("github.com/mycompany/myapp"))

// All subsequent logging uses these defaults unless channel is configured
dl.Log().Debug("this will be shown and colored")

// Per-channel config overrides global defaults
dl.ConfigureChannel("api", dl.DefaultOptions().JSON().SetLevel(slog.LevelWarn))
dl.ChannelLog("api").Debug("this won't be shown (level too low)")
dl.ChannelLog("api").Warn("this will be shown as JSON")
```

### 9. Advanced Options - Fine Control

**Detailed configuration options**

```go
opts := dl.DefaultOptions()

// Time formatting
opts.TimestampFormat = "15:04:05.000"  // custom time format
opts.AbsoluteTime = true                // show absolute vs relative time

// Function name trimming
opts.TrimPrefix = "github.com/mycompany/myapp"  // trim from function names

// Custom colors
opts.ErrorColor = "\033[91m"    // bright red
opts.InfoColor = "\033[92m"     // bright green
opts.TimestampColor = "\033[90m" // dark gray

// Level labels
opts.ErrorLabel = " ERR"
opts.InfoLabel = "INFO"
opts.DebugLabel = " DBG"

dl.ConfigureChannel("custom", opts)
```

### 10. Context Integration - Request Tracking

**Integrate with request context and tracing**

```go
// Build context throughout request lifecycle
func handleRequest(w http.ResponseWriter, r *http.Request) {
    requestID := generateRequestID()
    
    // Base logger for this request
    reqLogger := dl.ChannelLog("http").
        With("request_id", requestID).
        With("method", r.Method).
        With("path", r.URL.Path)
    
    reqLogger.Info("request started")
    
    // Pass context to deeper functions
    processUser(reqLogger.With("component", "user_service"), userID)
    
    reqLogger.With("status", 200).Info("request completed")
}

func processUser(logger *dl.Builder, userID string) {
    // Logger already has request context + component
    logger.With("user_id", userID).Info("processing user")
    
    // Database operations with same context
    dbLogger := logger.With("table", "users")
    dbLogger.Info("querying user")
    dbLogger.With("rows", 1).Info("query completed")
}
```

### 11. Performance Patterns - High Volume

**Optimize for high-volume logging**

```go
// Pre-build loggers for hot paths
var (
    dbLogger   = dl.ChannelLog("database")
    httpLogger = dl.ChannelLog("http")
    authLogger = dl.ChannelLog("auth")
)

// Conditional logging to avoid expensive operations
if dl.Log().Enabled(context.Background(), slog.LevelDebug) {
    dl.Log().Debug("expensive debug info", "data", expensiveOperation())
}

// Use appropriate levels
dl.Log().Debug("detailed trace")  // disabled in production
dl.Log().Info("normal operation") // general information
dl.Log().Warn("something unusual") // attention needed
dl.Log().Error("operation failed") // errors requiring action
```

### 12. Integration Patterns - With Applications

**Integrate logging into larger applications**

```go
// Application startup
func main() {
    // Initialize logging first
    setupLogging()
    
    // Start application components
    app := da.NewApplication(config)
    da.WithFactory(app, &LoggerFactory{}) // provides dl loggers to other components
    
    app.Initialize()
    app.Start()
}

func setupLogging() {
    // Global setup
    dl.Init(dl.DefaultOptions().SetLevel(slog.LevelInfo))
    
    // Per-service channels
    dbFile, _ := os.Create("logs/database.log")
    dl.ConfigureChannel("database", dl.DefaultOptions().SetOutput(dbFile))
    
    apiFile, _ := os.Create("logs/api.log")
    dl.ConfigureChannel("api", dl.DefaultOptions().JSON().SetOutput(apiFile))
    
    // Error aggregation
    dl.ConfigureChannel("errors", dl.DefaultOptions().Color().SetLevel(slog.LevelError))
}

// Factory provides loggers to application components
type LoggerFactory struct{}

func (f *LoggerFactory) Build(app *da.Application[Config]) error {
    // Register channel-specific loggers
    da.SetNamed(app.R, "database", dl.ChannelLog("database"))
    da.SetNamed(app.R, "api", dl.ChannelLog("api"))
    da.SetNamed(app.R, "errors", dl.ChannelLog("errors"))
    return nil
}
```

## Core Functions

| Function | Purpose | Use Case |
|----------|---------|----------|
| `dl.Log()` | Default logger builder | General application logging |
| `dl.ChannelLog(name)` | Channel-specific logger | Categorized logging |
| `dl.ConfigureChannel(name, opts)` | Configure channel behavior | Route/format specific channels |
| `dl.Init(opts)` | Initialize global defaults | Application startup |
| `dl.DefaultOptions()` | Create configuration | Channel and global setup |

## Builder Methods

| Method | Purpose | Example |
|--------|---------|---------|
| `.With(key, value)` | Add context attribute | `.With("user_id", 123)` |
| `.Info(msg)` | Log info message | `.Info("operation completed")` |
| `.Warn(msg)` | Log warning message | `.Warn("disk space low")` |
| `.Error(msg)` | Log error message | `.Error("connection failed")` |
| `.Debug(msg)` | Log debug message | `.Debug("detailed trace")` |
| `.Infof(fmt, args...)` | Printf-style info | `.Infof("port %d", 8080)` |

## Options Methods

| Method | Purpose | Example |
|--------|---------|---------|
| `.SetOutput(w)` | Set output destination | `.SetOutput(os.Stderr)` |
| `.SetLevel(level)` | Set minimum level | `.SetLevel(slog.LevelDebug)` |
| `.JSON()` | Enable JSON format | `.JSON()` |
| `.Pretty()` | Enable pretty format | `.Pretty()` |
| `.Color()` | Enable colors | `.Color()` |
| `.NoColor()` | Disable colors | `.NoColor()` |

## Common Patterns

### Multi-Service Application
```go
// Service-specific channels
dl.ChannelLog("user-service").With("user_id", 123).Info("user updated")
dl.ChannelLog("order-service").With("order_id", 456).Info("order processed")
dl.ChannelLog("payment-service").With("transaction_id", 789).Error("payment failed")
```

### Request Correlation
```go
// All logs for a request share context
reqLogger := dl.ChannelLog("api").With("request_id", id).With("user_id", userID)
reqLogger.Info("request started")
reqLogger.With("duration", elapsed).Info("request completed")
```

### Environment-Specific Configuration
```go
if isProduction() {
    dl.Init(dl.DefaultOptions().JSON().SetLevel(slog.LevelWarn))
    dl.ConfigureChannel("audit", dl.DefaultOptions().JSON().SetOutput(auditFile))
} else {
    dl.Init(dl.DefaultOptions().Pretty().Color().SetLevel(slog.LevelDebug))
}
```

---

*See [dl/examples/](../../../dl/examples/) for complete working examples of each feature.*