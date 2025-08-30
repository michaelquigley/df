---
title: Application Lifecycle
description: Build and orchestrate complex applications with factories, lifecycle management, and dependency injection.
---

The df Application type provides complete lifecycle orchestration for complex applications. It manages object creation through factories, automatic dependency injection, and structured startup/shutdown sequences.

## Application Basics

An `Application` orchestrates your entire application lifecycle:

```go
type Config struct {
    AppName     string `df:"app_name"`
    DatabaseURL string `df:"database_url"`
    LogLevel    string `df:"log_level"`
}

// Create application with configuration
config := Config{
    AppName:     "MyApp",
    DatabaseURL: "postgres://localhost:5432/mydb",
    LogLevel:    "info",
}

app := df.NewApplication(config)
```

## Application Phases

The application follows a structured lifecycle with four phases:

1. **Build**: Factories create and register objects in the container
2. **Link**: Objects implementing `Linkable` establish dependencies  
3. **Start**: Objects implementing `Startable` are initialized
4. **Stop**: Objects implementing `Stoppable` are gracefully shut down

### Phase Execution

```go
// Initialize runs Build + Link phases
err := app.Initialize()
if err != nil {
    log.Fatal("initialization failed:", err)
}

// Start phase
err = app.Start()
if err != nil {
    log.Fatal("startup failed:", err)
}

// Application running...

// Stop phase (graceful shutdown)
err = app.Stop()
if err != nil {
    log.Printf("shutdown error: %v", err)
}
```

## Factory Pattern

Factories create and configure objects based on application configuration:

### Basic Factory

```go
type Database struct {
    URL       string
    Connected bool
}

func (d *Database) Start() error {
    fmt.Printf("connecting to database: %s\n", d.URL)
    d.Connected = true
    return nil
}

func (d *Database) Stop() error {
    fmt.Printf("disconnecting from database\n")
    d.Connected = false
    return nil
}

// Factory creates the database service
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(a *df.Application[Config]) error {
    // Get application configuration
    cfg, _ := df.Get[Config](a.C)
    
    // Create and configure service
    db := &Database{URL: cfg.DatabaseURL}
    
    // Register in container
    df.SetAs[*Database](a.C, db)
    
    return nil
}

// Register factory with application
df.WithFactory(app, &DatabaseFactory{})
```

### Factory with Dependencies

```go
type UserService struct {
    db     *Database
    logger *Logger
    config UserServiceConfig
}

type UserServiceFactory struct{}

func (f *UserServiceFactory) Build(a *df.Application[Config]) error {
    // Get application config
    appConfig, _ := df.Get[Config](a.C)
    
    // Create service-specific config
    serviceConfig := UserServiceConfig{
        TableName: "users",
        CacheSize: 1000,
    }
    
    // Create service (dependencies will be injected in Link phase)
    service := &UserService{
        config: serviceConfig,
    }
    
    df.SetAs[*UserService](a.C, service)
    return nil
}
```

## Lifecycle Interfaces

Objects can implement lifecycle interfaces to participate in application phases:

### Linkable Interface
Establish dependencies after all objects are created:

```go
type UserService struct {
    db     *Database
    logger *Logger
    config UserServiceConfig
}

// Link establishes dependencies
func (s *UserService) Link(c *df.Container) error {
    var found bool
    
    s.db, found = df.Get[*Database](c)
    if !found {
        return errors.New("database dependency not found")
    }
    
    s.logger, found = df.Get[*Logger](c)
    if !found {
        return errors.New("logger dependency not found")
    }
    
    s.logger.Info("user service dependencies linked")
    return nil
}
```

### Startable Interface
Initialize resources and start services:

```go
func (s *UserService) Start() error {
    s.logger.Info("starting user service")
    
    // Initialize caches, connections, etc.
    err := s.initializeCache()
    if err != nil {
        return fmt.Errorf("cache initialization failed: %w", err)
    }
    
    s.logger.Info("user service started successfully")
    return nil
}

func (s *UserService) initializeCache() error {
    // Cache initialization logic
    return nil
}
```

### Stoppable Interface
Clean up resources during graceful shutdown:

```go
func (s *UserService) Stop() error {
    s.logger.Info("stopping user service")
    
    // Clean up resources
    err := s.cleanup()
    if err != nil {
        return fmt.Errorf("cleanup failed: %w", err)
    }
    
    s.logger.Info("user service stopped")
    return nil
}

func (s *UserService) cleanup() error {
    // Cleanup logic - close connections, flush caches, etc.
    return nil
}
```

## Configuration-Driven Architecture

Build applications that reconfigure themselves based on configuration:

### Multi-Environment Configuration

```go
type Config struct {
    Environment string         `df:"environment"`
    Database    DatabaseConfig `df:"database"`
    Cache       CacheConfig    `df:"cache"`
    Features    FeatureFlags   `df:"features"`
}

type DatabaseConfig struct {
    Type     string `df:"type"`
    URL      string `df:"url"`
    PoolSize int    `df:"pool_size"`
}

type CacheConfig struct {
    Enabled bool   `df:"enabled"`
    Type    string `df:"type"`
    TTL     int    `df:"ttl"`
}

// Factory adapts based on configuration
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(a *df.Application[Config]) error {
    cfg, _ := df.Get[Config](a.C)
    
    switch cfg.Database.Type {
    case "postgres":
        db := &PostgresDatabase{
            URL:      cfg.Database.URL,
            PoolSize: cfg.Database.PoolSize,
        }
        df.SetAs[DatabaseInterface](a.C, db)
        
    case "mysql":
        db := &MySQLDatabase{
            URL:      cfg.Database.URL,
            PoolSize: cfg.Database.PoolSize,
        }
        df.SetAs[DatabaseInterface](a.C, db)
        
    case "memory":
        db := &InMemoryDatabase{}
        df.SetAs[DatabaseInterface](a.C, db)
        
    default:
        return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
    }
    
    return nil
}
```

### Conditional Service Registration

```go
type CacheFactory struct{}

func (f *CacheFactory) Build(a *df.Application[Config]) error {
    cfg, _ := df.Get[Config](a.C)
    
    // Only register cache if enabled
    if !cfg.Cache.Enabled {
        return nil
    }
    
    switch cfg.Cache.Type {
    case "redis":
        cache := &RedisCache{TTL: cfg.Cache.TTL}
        df.SetAs[CacheInterface](a.C, cache)
        
    case "memory":
        cache := &MemoryCache{TTL: cfg.Cache.TTL}
        df.SetAs[CacheInterface](a.C, cache)
        
    default:
        return fmt.Errorf("unsupported cache type: %s", cfg.Cache.Type)
    }
    
    return nil
}
```

## Complete Application Example

Here's a comprehensive example showing all lifecycle phases:

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "github.com/michaelquigley/df"
)

// Configuration
type Config struct {
    AppName     string         `df:"app_name"`
    Environment string         `df:"environment"`
    Database    DatabaseConfig `df:"database"`
    Server      ServerConfig   `df:"server"`
}

type DatabaseConfig struct {
    URL      string `df:"url"`
    PoolSize int    `df:"pool_size"`
}

type ServerConfig struct {
    Host string `df:"host"`
    Port int    `df:"port"`
}

// Services
type Database struct {
    config DatabaseConfig
}

func (d *Database) Start() error {
    fmt.Printf("connecting to database: %s\n", d.config.URL)
    return nil
}

func (d *Database) Stop() error {
    fmt.Printf("disconnecting from database\n")
    return nil
}

type WebServer struct {
    config ServerConfig
    db     *Database
}

func (s *WebServer) Link(c *df.Container) error {
    db, found := df.Get[*Database](c)
    if !found {
        return errors.New("database not found")
    }
    s.db = db
    return nil
}

func (s *WebServer) Start() error {
    fmt.Printf("starting web server on %s:%d\n", s.config.Host, s.config.Port)
    return nil
}

func (s *WebServer) Stop() error {
    fmt.Printf("stopping web server\n")
    return nil
}

// Factories
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(a *df.Application[Config]) error {
    cfg, _ := df.Get[Config](a.C)
    db := &Database{config: cfg.Database}
    df.SetAs[*Database](a.C, db)
    return nil
}

type WebServerFactory struct{}

func (f *WebServerFactory) Build(a *df.Application[Config]) error {
    cfg, _ := df.Get[Config](a.C)
    server := &WebServer{config: cfg.Server}
    df.SetAs[*WebServer](a.C, server)
    return nil
}

func main() {
    // Load configuration
    config := Config{
        AppName:     "MyWebApp",
        Environment: "development",
        Database: DatabaseConfig{
            URL:      "postgres://localhost:5432/myapp",
            PoolSize: 10,
        },
        Server: ServerConfig{
            Host: "localhost",
            Port: 8080,
        },
    }

    // Build application
    app := df.NewApplication(config)
    df.WithFactory(app, &DatabaseFactory{})
    df.WithFactory(app, &WebServerFactory{})

    // Initialize (Build + Link)
    if err := app.Initialize(); err != nil {
        log.Fatal("initialization failed:", err)
    }

    // Start all services
    if err := app.Start(); err != nil {
        log.Fatal("startup failed:", err)
    }

    fmt.Println("application started successfully")

    // Wait for shutdown signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    fmt.Println("shutting down...")

    // Graceful shutdown
    if err := app.Stop(); err != nil {
        log.Printf("shutdown error: %v", err)
    }

    fmt.Println("application stopped")
}
```

## Error Handling and Recovery

### Factory Error Handling

```go
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(a *df.Application[Config]) error {
    cfg, found := df.Get[Config](a.C)
    if !found {
        return errors.New("configuration not found")
    }
    
    if cfg.Database.URL == "" {
        return errors.New("database URL is required")
    }
    
    db, err := connectToDatabase(cfg.Database.URL)
    if err != nil {
        return fmt.Errorf("database connection failed: %w", err)
    }
    
    df.SetAs[*Database](a.C, db)
    return nil
}
```

### Startup Error Recovery

```go
func (s *WebServer) Start() error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := s.attemptStart()
        if err == nil {
            return nil
        }
        
        fmt.Printf("startup attempt %d failed: %v\n", i+1, err)
        if i < maxRetries-1 {
            time.Sleep(time.Second * time.Duration(i+1))
        }
    }
    
    return fmt.Errorf("server startup failed after %d attempts", maxRetries)
}
```

### Partial Failure Handling

```go
func (app *Application[T]) startServices() error {
    startables := df.AsType[df.Startable](app.C)
    var errors []error
    
    for _, service := range startables {
        if err := service.Start(); err != nil {
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        // Log errors but continue if non-critical services failed
        for _, err := range errors {
            log.Printf("service startup error: %v", err)
        }
        
        // Only fail if critical services failed
        return app.checkCriticalServices()
    }
    
    return nil
}
```

## Testing Applications

### Test Application Builder

```go
func NewTestApplication() *df.Application[Config] {
    config := Config{
        AppName: "TestApp",
        Database: DatabaseConfig{
            URL: "memory://test",
        },
        Server: ServerConfig{
            Host: "localhost",
            Port: 0, // Random port
        },
    }
    
    app := df.NewApplication(config)
    df.WithFactory(app, &MockDatabaseFactory{})
    df.WithFactory(app, &TestServerFactory{})
    
    return app
}

func TestApplicationStartup(t *testing.T) {
    app := NewTestApplication()
    
    err := app.Initialize()
    assert.NoError(t, err)
    
    err = app.Start()
    assert.NoError(t, err)
    
    // Test services are running
    db, found := df.Get[*Database](app.C)
    assert.True(t, found)
    assert.True(t, db.IsConnected())
    
    // Cleanup
    err = app.Stop()
    assert.NoError(t, err)
}
```

## Best Practices

### Factory Design
- Keep factories focused on single services
- Use configuration to drive object creation  
- Handle missing dependencies gracefully
- Validate configuration in factories

### Dependency Management
- Use interface dependencies when possible
- Implement `Linkable` for complex dependencies
- Validate dependencies in Link phase
- Provide clear error messages for missing dependencies

### Lifecycle Management
- Implement graceful startup and shutdown
- Handle partial failures appropriately
- Use timeouts for long-running operations
- Log lifecycle events for debugging

### Configuration
- Use layered configuration (defaults → files → env → flags)
- Validate configuration early
- Support multiple environments
- Keep secrets separate from regular config

## Next Steps

Now that you understand application lifecycle management, explore:

- **[Advanced Features](/guides/advanced-features/)** - Dynamic types, object references, and custom converters
- **[Data Binding](/guides/data-binding/)** - Review struct binding and configuration techniques
- **[Dependency Injection](/guides/dependency-injection/)** - Deep dive into container patterns