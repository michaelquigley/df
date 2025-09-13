---
title: Getting Started
description: Get up and running with the df framework - installation, basic concepts, and your first application.
---

The df framework is a comprehensive Go library for building dynamic, configuration-driven applications. This guide will get you started with the basics.

## Installation

Add df to your Go project:

```bash
go get github.com/michaelquigley/df
```

## Core Philosophy

df provides three integrated layers that work together:

1. **Data Binding Foundation** - Bidirectional conversion between Go structs and structured data
2. **Dependency Injection Container** - Object management and service discovery  
3. **Application Orchestration** - Complete lifecycle management for complex applications

## Your First df Application

Let's start with a simple example that demonstrates the basic data binding capabilities:

```go
package main

import (
    "fmt"
    "github.com/michaelquigley/df"
)

type User struct {
    Name     string `dd:"+required"`
    Email    string
    Age      int    
    Active   bool   `dd:"is_active"`
    Password string `dd:"+secret"`
}

func main() {
    // Input data from any source (JSON, YAML, environment, etc.)
    data := map[string]any{
        "name":      "John Doe",
        "email":     "john@example.com", 
        "age":       30,
        "is_active": true,
        "password":  "secret123",
    }

    // Bind data to struct in a type-safe way
    user, err := df.New[User](data)
    if err != nil {
        panic(err)
    }

    // Inspect the result (secrets are automatically hidden)
    output, _ := df.Inspect(user)
    fmt.Println(output)
    
    // Convert back to map format  
    userData, _ := df.Unbind(user)
    fmt.Printf("User data: %+v\n", userData)
}
```

This example shows:
- **Type-safe binding** with `df.New[T]()`
- **Struct tags** for field mapping (`dd:"is_active"`)
- **Required fields** validation (`dd:"+required"`)
- **Secret handling** (`dd:"+secret"` fields hidden in output)
- **Bidirectional conversion** with `df.Unbind()`

## Working with Files

df makes it easy to load configuration from files:

```go
type Config struct {
    AppName     string `dd:"app_name"`
    DatabaseURL string `dd:"database_url"`
    LogLevel    string `dd:"log_level"`
}

func main() {
    var config Config
    
    // Load from JSON file
    err := df.BindFromJSON(&config, "config.json")
    if err != nil {
        panic(err)
    }
    
    // Or load from YAML file
    err = df.BindFromYAML(&config, "config.yaml")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("App: %s, DB: %s, Log: %s\n", 
        config.AppName, config.DatabaseURL, config.LogLevel)
}
```

## Building a Complete Application

Here's a more comprehensive example showing the full application lifecycle:

```go
package main

import (
    "fmt"
    "log"
    "github.com/michaelquigley/df"
)

// Application configuration
type Config struct {
    AppName     string `dd:"app_name"`
    DatabaseURL string `dd:"database_url"`
    LogLevel    string `dd:"log_level"`
}

// A simple service
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

// Factory creates and configures the database service
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(a *df.Application[Config]) error {
    // Get configuration
    cfg, _ := df.Get[Config](a.C)
    
    // Create and register service
    db := &Database{URL: cfg.DatabaseURL}
    df.SetAs[*Database](a.C, db)
    
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

    // 3. Initialize: build + link dependencies
    if err := app.Initialize(); err != nil {
        log.Fatal(err)
    }

    // 4. Start all services
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }

    // 5. Use services
    db, _ := df.Get[*Database](app.C)
    fmt.Printf("database connected: %v\n", db.Connected)

    // 6. Graceful shutdown
    if err := app.Stop(); err != nil {
        log.Printf("shutdown error: %v", err)
    }
}
```

This example demonstrates:
- **Configuration-driven services** using factories
- **Dependency injection** through the container
- **Lifecycle management** with Start/Stop phases
- **Service discovery** using `df.Get[T]()`

## Key Concepts

### Data Binding
- Convert between Go structs and maps/JSON/YAML
- Type-safe with compile-time checking
- Automatic field name conversion (snake_case ↔ CamelCase)
- Support for nested structs, slices, and pointers

### Container
- Singleton and named object registration
- Type-based and interface-based queries
- Container introspection for debugging

### Application
- Orchestrates object creation and lifecycle
- Factory pattern for configuration-driven object creation
- Automatic dependency injection and linking

## Next Steps

Now that you understand the basics, dive deeper into specific areas:

- **[Data Binding](/guides/data-binding/)** - Learn about struct tags, type coercion, and file operations
- **[Dependency Injection](/guides/dependency-injection/)** - Master the container and service discovery
- **[Application Lifecycle](/guides/application-lifecycle/)** - Build complex applications with factories and lifecycle management
- **[Advanced Features](/guides/advanced-features/)** - Explore Dynamic types, object references, and custom converters