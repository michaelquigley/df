---
title: df Guide
description: A comprehensive guide to building dynamic, configuration-driven applications with the df framework.
---

The df framework is a comprehensive Go library for building dynamic, configuration-driven applications. It provides a complete stack from low-level data binding to high-level application orchestration, enabling systems that can reconfigure their internal architecture based on runtime configuration.

## Quick Start

### Installation

```bash
go get github.com/michaelquigley/df
```

### Simple Data Binding

For basic data binding without the application framework:

```go
package main

import (
    "fmt"
    "github.com/michaelquigley/df"
)

type User struct {
    Name     string `df:"required"`
    Email    string
    Age      int    
    Active   bool   `df:"is_active"`
    Password string `df:"secret"`
}

func main() {
    // Input data
    data := map[string]any{
        "name":      "John Doe",
        "email":     "john@example.com", 
        "age":       30,
        "is_active": true,
        "password":  "secret123",
    }

    // Bind to struct
    user, err := df.New[User](data)
    if err != nil {
        panic(err)
    }

    // Inspect with secrets hidden
    output, _ := df.Inspect(user)
    fmt.Println(output)
}
```

### Complete Application Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/michaelquigley/df"
)

// Configuration struct
type Config struct {
    AppName     string `df:"app_name"`
    DatabaseURL string `df:"database_url"`
    LogLevel    string `df:"log_level"`
}

// Service implementations
type Database struct {
    URL string
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

type Logger struct {
    Level string
}

func (l *Logger) Start() error {
    fmt.Printf("starting logger with level: %s\n", l.Level)
    return nil
}

func (l *Logger) Info(msg string) {
    fmt.Printf("[INFO] %s\n", msg)
}

// Factories for dependency injection
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(a *df.Application[Config]) error {
    cfg, _ := df.Get[Config](a.C)
    db := &Database{URL: cfg.DatabaseURL}
    df.SetAs[*Database](a.C, db)
    return nil
}

type LoggerFactory struct{}

func (f *LoggerFactory) Build(a *df.Application[Config]) error {
    cfg, _ := df.Get[Config](a.C)
    logger := &Logger{Level: cfg.LogLevel}
    df.SetAs[*Logger](a.C, logger)
    return nil
}

func main() {
    // 1. Create configuration
    cfg := Config{
        AppName:     "MyApp",
        DatabaseURL: "postgres://localhost:5432/mydb",
        LogLevel:    "info",
    }

    // 2. Build application with factories
    app := df.NewApplication(cfg)
    df.WithFactory(app, &DatabaseFactory{})
    df.WithFactory(app, &LoggerFactory{})

    // 3. Initialize: build + link dependencies
    if err := app.Initialize(); err != nil {
        log.Fatal(err)
    }

    // 4. Start all services
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }

    // 5. Use services
    logger, _ := df.Get[*Logger](app.C)
    logger.Info("application started successfully")

    db, _ := df.Get[*Database](app.C)
    fmt.Printf("database connected: %v\n", db.Connected)

    // 6. Graceful shutdown
    if err := app.Stop(); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

## Core Concepts

### Three Integrated Layers

df consists of three integrated layers that work together:

#### 1. Data Binding Foundation
- **Bidirectional binding** between Go structs and structured data (JSON, YAML, maps)
- **Type-safe conversion** with support for primitives, pointers, slices, and nested structures  
- **Polymorphic data** via Dynamic interface for runtime type discrimination
- **Object references** with cycle-safe pointer resolution

#### 2. Dependency Injection Container
- **Object management** with singleton and named registration patterns
- **Type queries** for exact type matching and interface compatibility
- **Container introspection** with multiple output formats (human, JSON, YAML)

#### 3. Application Orchestration
- **Lifecycle management** with configurable phases (build → link → start → stop)
- **Factory pattern** for configuration-driven object creation
- **Dependency injection** through automatic linking of compatible objects
- **Service discovery** via container-based object lookup

## Data Binding

### Basic Functions

df provides several core functions for data binding:

- **`df.New[T](data)`** - Generic function for type-safe allocation and binding
- **`df.Bind(struct, data)`** - Bind data into pre-allocated struct
- **`df.Unbind(struct)`** - Convert struct back to map
- **`df.Merge(struct, data)`** - Merge data into pre-initialized struct for configuration layering

### Struct Tags

Control field binding behavior with `df` struct tags:

```go
type Example struct {
    Name     string `df:"custom_name,required"` // Custom field name, required
    Email    string `df:"email"`                // Custom field name
    Age      int    `df:",required"`            // Default name (snake_case), required  
    Password string `df:",secret"`              // Secret field (hidden in Inspect)
    Internal string `df:"-"`                    // Skip this field
    Default  string                             // Uses snake_case: "default"
}
```

### File Operations

```go
// Bind from JSON file
err := df.BindFromJSON(&user, "user.json")

// Bind from YAML file  
err := df.BindFromYAML(&user, "user.yaml")

// Unbind to JSON file
err := df.UnbindToJSON(&user, "output.json")

// Unbind to YAML file
err := df.UnbindToYAML(&user, "output.yaml")
```

## Container: Dependency Injection

The `Container` type provides a modern dependency injection system:

### Basic Container Usage

```go
// Create container
container := df.NewContainer()

// Register singleton objects by type
database := &Database{URL: "localhost:5432"}
container.Set(database)

// Register named objects (multiple instances)
container.SetNamed("primary", &Database{URL: "primary-db:5432"})
container.SetNamed("cache", &Database{URL: "cache-db:6379"})

// Retrieve objects
db, found := df.Get[*Database](container)           // Get singleton
primary, found := df.GetNamed[*Database](container, "primary") // Get named

// Query by type (returns all instances)
allDatabases := df.OfType[*Database](container)     // [singleton, primary, cache]

// Query by interface (returns compatible objects)
startables := df.AsType[df.Startable](container)    // All objects implementing Startable
```

### Container Introspection

```go
// Human-readable output
output, _ := container.Inspect(df.InspectHuman)
fmt.Println(output)

// Machine-readable formats
jsonOutput, _ := container.Inspect(df.InspectJSON)
yamlOutput, _ := container.Inspect(df.InspectYAML)
```

## Application: Complete Lifecycle Management

The `Application` type orchestrates object creation, dependency injection, and lifecycle management.

### Application Phases

1. **Build**: Factories create and register objects in the container
2. **Link**: Objects implementing `Linkable` establish dependencies  
3. **Start**: Objects implementing `Startable` are initialized
4. **Stop**: Objects implementing `Stoppable` are gracefully shut down

### Lifecycle Interfaces

```go
// Linkable: Establish dependencies after all objects are created
type DatabaseService struct {
    db *Database
}

func (s *DatabaseService) Link(c *df.Container) error {
    db, found := df.Get[*Database](c)
    if !found {
        return errors.New("database not found")
    }
    s.db = db
    return nil
}

// Startable: Initialize resources
func (s *DatabaseService) Start() error {
    return s.db.Connect()
}

// Stoppable: Clean up resources  
func (s *DatabaseService) Stop() error {
    return s.db.Disconnect()
}
```

### Factory Pattern

```go
type DatabaseServiceFactory struct{}

func (f *DatabaseServiceFactory) Build(app *df.Application[Config]) error {
    service := &DatabaseService{}
    df.SetAs[*DatabaseService](app.C, service)
    return nil
}

// Register factory with application
app := df.NewApplication(config)
df.WithFactory(app, &DatabaseServiceFactory{})
```

## Advanced Features

### Configuration Merging with Defaults

The `Merge` function enables sophisticated configuration hierarchies:

```go
// Start with sensible defaults
config := &ServerConfig{
    Host:    "localhost",
    Port:    8080,
    Timeout: 30,
    Debug:   false,
}

// Layer 1: Configuration file
if configExists("app.yaml") {
    err := df.MergeFromYAML(config, "app.yaml")
}

// Layer 2: Environment variables
envVars := getEnvironmentOverrides()
err := df.Merge(config, envVars)

// Layer 3: Command line flags
cliFlags := getCLIOverrides()
err = df.Merge(config, cliFlags)
```

### Dynamic Fields (Polymorphic Data)

Support polymorphic data structures with the `Dynamic` interface:

```go
// Define concrete types that implement Dynamic
type EmailAction struct {
    Recipient string `df:"recipient"`
    Subject   string `df:"subject"`
}

func (e EmailAction) Type() string { return "email" }
func (e EmailAction) ToMap() map[string]any {
    return map[string]any{
        "recipient": e.Recipient,
        "subject":   e.Subject,
    }
}

// Use Dynamic fields in structs
type Notification struct {
    Name   string  `df:"name"`
    Action Dynamic `df:"action"`  // Polymorphic field
}

// Configure binders for different types
opts := &df.Options{
    DynamicBinders: map[string]func(map[string]any) (df.Dynamic, error){
        "email": func(m map[string]any) (df.Dynamic, error) {
            action, err := df.New[EmailAction](m)
            if err != nil {
                return nil, err
            }
            return *action, nil
        },
    },
}
```

### Object References

Support object references with cycle handling using `df.Pointer[T]`:

```go
type User struct {
    ID   string `df:"id"`
    Name string `df:"name"`
}

func (u *User) GetId() string { return u.ID }

type Document struct {
    ID     string             `df:"id"`
    Title  string             `df:"title"`
    Author *df.Pointer[*User] `df:"author"`
}

func (d *Document) GetId() string { return d.ID }

// Two-phase process
var container DataContainer
df.Bind(&container, data)  // Phase 1: Bind data with $ref strings
df.Link(&container)        // Phase 2: Resolve all pointer references

// Access resolved objects
author := container.Documents[0].Author.Resolve()
```

## Best Practices

### Configuration Management

1. **Use layered configuration** with `Merge` for flexible deployment scenarios
2. **Define sensible defaults** in your structs
3. **Mark sensitive fields** with the `secret` tag
4. **Use required tags** for critical configuration values

### Dependency Injection

1. **Implement lifecycle interfaces** (`Startable`, `Stoppable`, `Linkable`) for managed services
2. **Use factories** for configuration-driven object creation
3. **Leverage type queries** to discover services by interface
4. **Inspect containers** during development for debugging

### Application Architecture

1. **Separate concerns** into distinct layers (config, services, controllers)
2. **Use the factory pattern** for configurable object creation
3. **Implement graceful shutdown** with proper dependency ordering
4. **Leverage container introspection** for monitoring and debugging

## Architecture Patterns

### Simple Configuration Loading
```go
// Load application settings
type Config struct {
    Database DatabaseConfig `df:"database"`
    Server   ServerConfig   `df:"server"`
}

config, err := df.NewFromYAML[Config]("config.yaml")
```

### Plugin-Based Systems
```go
// Load and configure plugins dynamically
container := df.NewContainer()

// Register plugin types
pluginConfigs := loadPluginConfigs()
for _, pluginConfig := range pluginConfigs {
    plugin, err := df.New[Plugin](pluginConfig)
    container.SetNamed(plugin.Name(), plugin)
}

// Find all plugins implementing specific interfaces  
authPlugins := df.AsType[AuthenticationPlugin](container)
```

### Microservice Orchestration
```go
// Coordinate multiple services with shared dependencies
app := df.NewApplication(serviceConfig)

// Shared infrastructure
df.WithFactory(app, &DatabaseFactory{})
df.WithFactory(app, &MessageQueueFactory{})

// Service-specific components
df.WithFactory(app, &UserServiceFactory{})
df.WithFactory(app, &OrderServiceFactory{})

app.Initialize()
app.Start()
```

## Next Steps

- Explore the [API Reference](/reference/df/) for detailed function documentation
- Check out the examples in the repository for complete working applications
- Read about advanced features like custom converters and marshaling interfaces
