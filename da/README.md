# da - Dynamic Application

**Easily manage massive monoliths in code**

The `da` package provides a powerful application container for managing object dependencies, lifecycle, and service discovery in large, dynamic Go applications.

## Quick Start

```go
import "github.com/michaelquigley/df/da"

// Create application with config
app := da.NewApplication(config)

// Add factories to build objects
da.WithFactory(app, &DatabaseFactory{})
da.WithFactory(app, &LoggerFactory{})

// Initialize: build objects and link dependencies  
app.Initialize()

// Start all startable services
app.Start()

// Clean shutdown
defer app.Stop()
```

## Key Features

- **🏭 Factory Pattern**: Automatic object creation and registration
- **🔗 Dependency Injection**: Type-safe object retrieval and linking
- **♻️ Lifecycle Management**: Initialize → Start → Stop pattern
- **🔍 Service Discovery**: Find objects by type or interface
- **📋 Container Introspection**: Debug and inspect container contents
- **⚡ Thread-Safe**: Concurrent access to container objects

## Core Components

- **`Container`** - Object storage with singleton and named object support
- **`Application[C]`** - Orchestrates object creation and lifecycle
- **`Factory[C]`** - Interface for creating and registering objects
- **Lifecycle interfaces**: `Startable`, `Stoppable`, `Linkable`

## Object Management

**Register and retrieve objects**
```go
// Store singleton objects (one per type)
da.Set(container, database)
da.SetAs[DataStore](container, database)  // store as interface

// Retrieve objects by type
db, found := da.Get[*Database](container)
stores := da.OfType[DataStore](container)  // all matching interface

// Named objects (multiple per type)
da.SetNamed(container, "primary", primaryDB)
da.SetNamed(container, "cache", cacheDB)
```

## Factory Pattern

**Create objects with dependencies**
```go
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(app *da.Application[Config]) error {
    // Get config from container
    cfg, _ := da.Get[Config](app.R)
    
    // Create and register database
    db := &Database{URL: cfg.DatabaseURL}
    da.SetAs[*Database](app.R, db)
    return nil
}
```

## Lifecycle Management

**Objects implement lifecycle interfaces**
```go
type Database struct {
    URL       string
    Connected bool
}

// da.Startable - called during app.Start()
func (d *Database) Start() error {
    return d.Connect()
}

// da.Stoppable - called during app.Stop()  
func (d *Database) Stop() error {
    d.Connected = false
    return nil
}

// da.Linkable - called after all objects created
func (d *Database) Link(container *da.Container) error {
    logger, _ := da.Get[*Logger](container)
    d.logger = logger
    return nil
}
```

## Common Patterns

**Application Bootstrap**
```go
type AppConfig struct {
    DatabaseURL string `json:"database_url"`
    LogLevel    string `json:"log_level"`
}

// Load config and create application
config := loadConfig()
app := da.NewApplication(config)

// Register all factories
da.WithFactory(app, &DatabaseFactory{})
da.WithFactory(app, &CacheFactory{})
da.WithFactory(app, &HTTPServerFactory{})

// Initialize and start
app.Initialize()  // builds all objects
app.Start()       // starts all Startable objects
```

**Service Discovery**
```go
// Find all objects implementing an interface
allCaches := da.AsType[CacheService](container)
allStartables := da.AsType[da.Startable](container)

// Container introspection for debugging
data := container.Inspect()
fmt.Printf("Container has %d objects\n", data.Summary.Total)
```

## Examples

See [examples/](examples/) for tutorials on basic containers, dependency injection, and lifecycle management.

---
*Part of the [df framework](../README.md) - dynamic foundation for Go applications*