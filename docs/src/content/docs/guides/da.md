---
title: dynamic foundation for applications
description: Complete reference for the da package - dependency injection and application lifecycle management.
---

**Easily manage massive monoliths in code**

The `da` package provides two complementary approaches to application lifecycle management:

- **Concrete containers** - define your own container struct with explicit types
- **Dynamic containers** - factory pattern with reflection-based object storage

---

## Concrete Containers

Define your own container struct with explicit types. Components implement `Wireable[C]` to receive dependencies.

### Basic Usage

```go
import "github.com/michaelquigley/df/da"

// Define container with explicit types
type App struct {
    Config   *Config   `da:"-"`        // skip - not a component
    Database *Database `da:"order=1"`  // wire/start first
    Cache    *Cache    `da:"order=2"`  // wire/start second
    API      *Server   `da:"order=10"` // wire/start last
}

// Load configuration
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

### Component Wiring

Components implement `Wireable[C]` to receive the container and wire dependencies:

```go
type UserService struct {
    db    *Database
    cache *Cache
}

// Wireable[App] - receives container for dependency wiring
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

### Configuration Loading

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

### Struct Tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `da:"-"` | Skip field | `Config *Config \`da:"-"\`` |
| `da:"order=N"` | Process order | `DB *Database \`da:"order=1"\`` |

### Lifecycle Functions

| Function | Purpose |
|----------|---------|
| `da.Wire[C](c)` | Call `Wire(c)` on all `Wireable[C]` components |
| `da.Start[C](c)` | Call `Start()` on all `Startable` components |
| `da.Stop[C](c)` | Call `Stop()` on all `Stoppable` components (reverse order) |
| `da.Run[C](c)` | Wire → Start → wait for signal → Stop |
| `da.WaitForSignal()` | Block until SIGINT/SIGTERM |

### Nested Structs and Collections

Traversal automatically handles nested structs, slices, and maps:

```go
type App struct {
    Config *Config `da:"-"`

    // Nested struct - fields are traversed
    Services struct {
        Auth  *AuthService  `da:"order=10"`
        Users *UserService  `da:"order=20"`
    }

    // Slices of pointers - each element processed
    Workers []*Worker

    // Maps with pointer values - each value processed
    Handlers map[string]*Handler
}
```

---

## Dynamic Containers

Factory pattern with reflection-based object storage for configuration-driven object creation.

## Quick Reference

### 1. Basic Container - Hello World

**Store and retrieve objects**

```go
import "github.com/michaelquigley/df/da"

// Create container
container := da.NewContainer()

// Store objects (singletons)
database := &Database{URL: "localhost:5432"}
da.Set(container, database)

// Retrieve objects by type
db, found := da.Get[*Database](container)
fmt.Printf("Found: %v, URL: %s\n", found, db.URL)

// Store as interface
da.SetAs[DataStore](container, database)
store, found := da.Get[DataStore](container)
```

### 2. Named Objects - Multiple Instances

**Store multiple instances of the same type**

```go
// Store named instances
primary := &Database{URL: "primary-db:5432"}
cache := &Database{URL: "cache-db:6379"}

da.SetNamed(container, "primary", primary)
da.SetNamed(container, "cache", cache)

// Retrieve by name
primaryDB, found := da.GetNamed[*Database](container, "primary")
cacheDB, found := da.GetNamed[*Database](container, "cache")

// Get all instances of a type
allDBs := da.OfType[*Database](container)
fmt.Printf("Found %d databases\n", len(allDBs))
```

### 3. Type Queries - Service Discovery

**Find objects by type or interface**

```go
// Find all objects of exact type
databases := da.OfType[*Database](container)

// Find all objects implementing interface
startables := da.AsType[da.Startable](container)
stoppables := da.AsType[da.Stoppable](container)

// Use discovered services
for _, startable := range startables {
    err := startable.Start()
    fmt.Printf("Started service: %T\n", startable)
}
```

### 4. Application Creation - Lifecycle Management

**Create applications with configuration**

```go
type Config struct {
    DatabaseURL string `dd:"database_url"`
    LogLevel    string `dd:"log_level"`
    Port        int    `dd:"port"`
}

// Create application with configuration
config := Config{
    DatabaseURL: "postgres://localhost:5432/mydb",
    LogLevel:    "info",
    Port:        8080,
}

app := da.NewApplication(config)

// Access configuration and container
fmt.Printf("Config: %+v\n", app.Config)
fmt.Printf("Container has %d objects\n", len(app.R.GetAll()))
```

### 5. Factories - Object Creation

**Create objects with dependency injection**

```go
// Factory creates and registers objects
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(app *da.Application[Config]) error {
    // Access configuration
    cfg := app.Config
    
    // Create object
    db := &Database{
        URL:       cfg.DatabaseURL,
        Connected: false,
    }
    
    // Register in container (as singleton)
    da.Set(app.R, db)
    
    // Also register as interface
    da.SetAs[DataStore](app.R, db)
    
    return nil
}

// Register factory with application
da.WithFactory(app, &DatabaseFactory{})
```

### 6. Lifecycle Interfaces - Start/Stop/Link

**Objects that participate in application lifecycle**

```go
type Database struct {
    URL       string
    Connected bool
    logger    *Logger  // will be injected
}

// da.Startable - called during app.Start()
func (d *Database) Start() error {
    fmt.Printf("connecting to database: %s\n", d.URL)
    d.Connected = true
    return nil
}

// da.Stoppable - called during app.Stop()
func (d *Database) Stop() error {
    fmt.Printf("disconnecting from database\n")
    d.Connected = false
    return nil
}

// da.Linkable - called after all objects created
func (d *Database) Link(container *da.Container) error {
    // Inject dependencies
    logger, found := da.Get[*Logger](container)
    if found {
        d.logger = logger
    }
    return nil
}
```

### 7. Application Phases - Initialize/Start/Stop

**Complete application lifecycle**

```go
// Phase 1: Build objects
app := da.NewApplication(config)
da.WithFactory(app, &DatabaseFactory{})
da.WithFactory(app, &LoggerFactory{})
da.WithFactory(app, &APIServerFactory{})

// Phase 2: Initialize (build + link)
err := app.Initialize()
if err != nil {
    log.Fatal("Failed to initialize:", err)
}

// Phase 3: Start all services
err = app.Start()
if err != nil {
    log.Fatal("Failed to start:", err)
}

// Application running...

// Phase 4: Graceful shutdown
err = app.Stop()
if err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

### 8. Factory Functions - Simple Factories

**Function-based factories for simple cases**

```go
// Function factory
loggerFactory := da.FactoryFunc[Config](func(app *da.Application[Config]) error {
    logger := &Logger{
        Level: app.Config.LogLevel,
    }
    da.Set(app.R, logger)
    return nil
})

// Register function factory
da.WithFactory(app, loggerFactory)

// OR use inline
da.WithFactory(app, da.FactoryFunc[Config](func(app *da.Application[Config]) error {
    server := &HTTPServer{
        Port: app.Config.Port,
    }
    da.Set(app.R, server)
    return nil
}))
```

### 9. Container Introspection - Debugging

**Inspect container contents for debugging**

```go
// Get inspection data
data := container.Inspect()

fmt.Printf("Container Summary:\n")
fmt.Printf("  Total objects: %d\n", data.Summary.Total)
fmt.Printf("  Singletons: %d\n", data.Summary.Singletons)
fmt.Printf("  Named: %d\n", data.Summary.Named)

fmt.Printf("\nObjects:\n")
for i, obj := range data.Objects {
    fmt.Printf("  [%d] %s (%s): %s\n", 
        i, obj.Type, obj.Storage, obj.Value)
}

// Machine-readable formats
jsonData, _ := json.MarshalIndent(data, "", "  ")
fmt.Println(string(jsonData))
```

### 10. Advanced Factories - Complex Dependencies

**Factories with complex dependency resolution**

```go
type APIServerFactory struct{}

func (f *APIServerFactory) Build(app *da.Application[Config]) error {
    cfg := app.Config
    
    // Get dependencies (must be created by other factories)
    db, found := da.Get[*Database](app.R)
    if !found {
        return errors.New("database not found")
    }
    
    logger, found := da.Get[*Logger](app.R)
    if !found {
        return errors.New("logger not found")
    }
    
    // Create complex object with dependencies
    server := &APIServer{
        Port:     cfg.Port,
        Database: db,
        Logger:   logger,
        Routes:   setupRoutes(),
    }
    
    da.Set(app.R, server)
    return nil
}

// Factory registration order doesn't matter - dependencies resolved during linking
da.WithFactory(app, &APIServerFactory{})  // depends on db + logger
da.WithFactory(app, &DatabaseFactory{})   // no dependencies
da.WithFactory(app, &LoggerFactory{})     // no dependencies
```

### 11. Multiple Configurations - Environment Management

**Different configurations for different environments**

```go
type Environment string

const (
    Development Environment = "development"
    Production  Environment = "production"
    Testing     Environment = "testing"
)

func createApp(env Environment) *da.Application[Config] {
    config := getConfigForEnvironment(env)
    app := da.NewApplication(config)
    
    // Common factories
    da.WithFactory(app, &LoggerFactory{})
    da.WithFactory(app, &DatabaseFactory{})
    
    // Environment-specific factories
    switch env {
    case Development:
        da.WithFactory(app, &DevServerFactory{})
        da.WithFactory(app, &MockPaymentFactory{})
    case Production:
        da.WithFactory(app, &ProdServerFactory{})
        da.WithFactory(app, &StripePaymentFactory{})
    case Testing:
        da.WithFactory(app, &TestServerFactory{})
        da.WithFactory(app, &MockEverythingFactory{})
    }
    
    return app
}
```

### 12. Plugin Architecture - Dynamic Loading

**Load and manage plugins dynamically**

```go
// Plugin interface
type Plugin interface {
    Name() string
    Initialize(container *da.Container) error
    da.Startable
    da.Stoppable
}

// Plugin factory loads plugins from configuration
type PluginFactory struct{}

func (f *PluginFactory) Build(app *da.Application[Config]) error {
    pluginConfigs := app.Config.Plugins
    
    for _, pluginConfig := range pluginConfigs {
        // Load plugin dynamically (from file, registry, etc.)
        plugin, err := loadPlugin(pluginConfig.Name, pluginConfig.Config)
        if err != nil {
            return fmt.Errorf("failed to load plugin %s: %w", pluginConfig.Name, err)
        }
        
        // Register plugin by name
        da.SetNamed(app.R, pluginConfig.Name, plugin)
        
        // Also register as Plugin interface
        da.SetNamed[Plugin](app.R, pluginConfig.Name, plugin)
    }
    
    return nil
}

// Find and manage all plugins
func managePlugins(container *da.Container) {
    plugins := da.AsType[Plugin](container)
    
    fmt.Printf("Found %d plugins\n", len(plugins))
    for _, plugin := range plugins {
        fmt.Printf("  - %s\n", plugin.Name())
        plugin.Start()
    }
}
```

## Core Functions

| Function | Purpose | Use Case |
|----------|---------|----------|
| `da.NewContainer()` | Create container | Object storage |
| `da.Set(container, obj)` | Store singleton | Register objects |
| `da.Get[T](container)` | Retrieve singleton | Access objects |
| `da.SetNamed(container, name, obj)` | Store named object | Multiple instances |
| `da.GetNamed[T](container, name)` | Retrieve named object | Access by name |
| `da.OfType[T](container)` | Find all of type | Service discovery |
| `da.AsType[T](container)` | Find all implementing interface | Interface queries |
| `da.NewApplication(config)` | Create application | Lifecycle management |
| `da.WithFactory(app, factory)` | Register factory | Object creation |

## Lifecycle Methods

| Method | Purpose | When Called |
|--------|---------|-------------|
| `app.Initialize()` | Build + link objects | Application startup |
| `app.Start()` | Start all services | After initialization |
| `app.Stop()` | Stop all services | Application shutdown |

## Lifecycle Interfaces

| Interface | Method | Purpose |
|-----------|--------|---------|
| `da.Startable` | `Start() error` | Initialize resources |
| `da.Stoppable` | `Stop() error` | Clean up resources |
| `da.Linkable` | `Link(*da.Container) error` | Inject dependencies |

## Factory Pattern

| Type | Use Case | Example |
|------|----------|---------|
| `Factory[C]` | Complex object creation | Database connections |
| `FactoryFunc[C]` | Simple object creation | Configuration objects |

## Common Patterns

### Service Registration
```go
// Register services with multiple interfaces
service := &UserService{}
da.Set(app.R, service)                    // as concrete type
da.SetAs[UserRepository](app.R, service)  // as repository interface
da.SetAs[UserValidator](app.R, service)   // as validator interface
```

### Dependency Injection
```go
// Automatic dependency injection via Link method
func (s *UserService) Link(container *da.Container) error {
    s.db, _ = da.Get[*Database](container)
    s.logger, _ = da.Get[*Logger](container)
    s.cache, _ = da.GetNamed[*Redis](container, "user-cache")
    return nil
}
```

### Graceful Shutdown
```go
// Shutdown handling
c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)

go func() {
    <-c
    fmt.Println("Shutting down...")
    app.Stop()
    os.Exit(0)
}()
```

### Health Checks
```go
// Container-based health checks
func healthCheck(container *da.Container) map[string]bool {
    health := make(map[string]bool)
    
    if db, found := da.Get[*Database](container); found {
        health["database"] = db.Connected
    }
    
    services := da.AsType[HealthChecker](container)
    for _, service := range services {
        health[service.Name()] = service.IsHealthy()
    }
    
    return health
}
```

---

*See [da/examples/](../../../da/examples/) for complete working examples of each feature.*