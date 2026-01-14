# da - Dynamic Application

**Easily manage massive monoliths in code**

The `da` package provides two complementary approaches to application lifecycle management:

- **Concrete containers** - define your own container struct with explicit types
- **Dynamic containers** - factory pattern with reflection-based object storage

## Quick Start - Concrete Containers

```go
import "github.com/michaelquigley/df/da"

// Define your container with explicit types
type App struct {
    Config   *Config   `da:"-"`        // skip - not a component
    Database *Database `da:"order=1"`  // wire/start first
    Cache    *Cache    `da:"order=2"`  // wire/start second
    API      *Server   `da:"order=10"` // wire/start last
}

// Load config
cfg := &Config{}
da.Config(cfg, da.FileLoader("config.yaml"))

// Build container explicitly
app := &App{
    Config:   cfg,
    Database: NewDatabase(cfg.DatabaseURL),
    Cache:    NewCache(cfg.CacheURL),
    API:      NewServer(cfg.Port),
}

// Run application (wire -> start -> wait for signal -> stop)
da.Run(app)
```

## Quick Start - Dynamic Containers

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

- **Concrete Containers**: Type-safe, explicit dependency wiring
- **Dynamic Containers**: Factory pattern with automatic object creation
- **Lifecycle Management**: Wire/Initialize → Start → Stop pattern
- **Service Discovery**: Find objects by type or interface
- **Configuration Loading**: Flexible file-based config with loaders
- **Container Introspection**: Debug and inspect container contents

## Core Components

### Concrete Containers
- **`Wireable[C]`** - Interface for type-safe dependency wiring
- **`Wire[C]`/`Start[C]`/`Stop[C]`/`Run[C]`** - Lifecycle functions
- **`Loader`** - Configuration loading interface
- **Struct tags**: `da:"order=N"` for ordering, `da:"-"` to skip

### Dynamic Containers
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

## Concrete Container Pattern

**Define components that wire to the container**
```go
type UserService struct {
    db    *Database
    cache *Cache
}

// Wireable[App] - receives concrete container for dependency wiring
func (s *UserService) Wire(app *App) error {
    s.db = app.Database
    s.cache = app.Cache
    return nil
}

// Startable - called during da.Start()
func (s *UserService) Start() error {
    return s.db.Ping()
}

// Stoppable - called during da.Stop()
func (s *UserService) Stop() error {
    return nil
}
```

**Configuration loading**
```go
cfg := &Config{}

// Load from required file
da.Config(cfg, da.FileLoader("config.yaml"))

// Load with optional overrides
da.Config(cfg,
    da.FileLoader("config.yaml"),
    da.OptionalFileLoader("config.local.yaml"),
)

// Chain multiple loaders
da.Config(cfg, da.ChainLoader(
    da.FileLoader("base.yaml"),
    da.OptionalFileLoader("env.yaml"),
))
```

**Struct tags for ordering**
```go
type App struct {
    Config   *Config   `da:"-"`        // skip - not a component
    Database *Database `da:"order=1"`  // process first
    Cache    *Cache    `da:"order=2"`  // process second
    API      *Server   `da:"order=100"` // process last
}
```

## Examples

See [examples/](examples/) for tutorials:
- `da_01_hello_world` - Dynamic container with factories
- `da_02_concrete_container` - Concrete container pattern

---
*Part of the [df framework](../README.md) - dynamic foundation for Go applications*