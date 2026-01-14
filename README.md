# df - Dynamic Foundation

**A Go framework for building dynamic, configuration-driven applications**

The `df` (dynamic foundation) framework enables applications that can reconfigure their internal architecture based on runtime configuration. It provides the essential building blocks for creating flexible, manageable systems at any scale.

## Framework Components

The `df` framework consists of three complementary packages:

### [`dd`](dd/) - dynamic foundation for data
**Convert between Go structs and maps with ease**

Bidirectional data binding between Go structs and `map[string]any`. Since maps are foundational to all data systems, this enables seamless integration with any network protocol, database, object store, or file format.

```go
// struct → map → JSON/YAML/database
user := User{Name: "John", Age: 30}
data, _ := dd.Unbind(user)

// map → struct (from config, API, etc.)
userData := map[string]any{"name": "Alice", "age": 25}
user, _ := dd.New[User](userData)
```

### [`dl`](dl/) - dynamic foundation for logging
**Intelligent channel-based logging built on Go's slog**

Route different log categories to independent destinations with per-channel configuration. Database logs can go to files, HTTP logs to JSON format, errors to colored console output.

```go
// Different channels route to different destinations
dl.ChannelLog("database").With("table", "users").Info("query executed")
dl.ChannelLog("http").With("status", 200).Info("request processed")
dl.ChannelLog("errors").With("code", 500).Error("internal error")
```

### [`da`](da/) - dynamic foundation for applications
**Easily manage massive monoliths in code**

This is not "dependency injection". This is an idiomatic, clear, consistent approach to managing instantiation and lifecycle for large multi-component applications. Define your own container struct with explicit types and let `da` handle wiring and lifecycle.

```go
// User-defined container with explicit types
type App struct {
    Config   *Config   `da:"-"`
    Database *Database `da:"order=1"`
    Cache    *Cache    `da:"order=2"`
    API      *Server   `da:"order=10"`
}

app := &App{Config: cfg, Database: db, Cache: cache, API: server}
da.Run(app)  // wire -> start -> wait for signal -> stop
```

## When to Use df

- **Configuration-driven applications** that need to adapt behavior based on runtime config
- **Large monolithic applications** that need organized component management
- **Systems integration** where data flows between different formats and protocols
- **Plugin architectures** with dynamic component loading and lifecycle management
- **Microservice orchestration** with shared infrastructure and service discovery

## Getting Started

Start with the [documentation](https://michaelquigley.github.io/df/).

Each package works independently:

- **Start with `dd`** for struct ↔ map conversion and configuration loading
- **Add `dl`** for intelligent logging with channel routing
- **Use `da`** for application container and lifecycle management

See examples in each package: [dd/examples/](dd/examples/), [dl/examples/](dl/examples/), [da/examples/](da/examples/)

## License

See [LICENSE](LICENSE) file for details.