---
title: How to learn df...
description: Overview of the df framework for building dynamic, configuration-driven applications in Go.
---

The df framework is a comprehensive Go library for building dynamic, configuration-driven applications. It provides a complete stack from low-level data binding to high-level application orchestration, enabling systems that can reconfigure their internal architecture based on runtime configuration.

## What is df?

df serves as the foundation for dynamic, configuration-driven applications. It's an evolution of concepts first espoused by IoC and dependency-injection frameworks like Spring, completely reimagined for modern Go development.

The library allows you to:
- Define data structure shapes in Go and populate them from various sources at runtime
- Build "application containers" that make managing large, sprawling, dynamic software infrastructures manageable  
- Provide "glue" that keeps large, dynamic applications well organized
- Use performant data binding as primary persistence for any aspect of your application

## Framework Architecture

df consists of three integrated layers that work together:

### 1. Data Binding Foundation
- **Bidirectional binding** between Go structs and structured data (JSON, YAML, maps)
- **Type-safe conversion** with support for primitives, pointers, slices, and nested structures  
- **Polymorphic data** via Dynamic interface for runtime type discrimination
- **Object references** with cycle-safe pointer resolution

### 2. Dependency Injection Container
- **Object management** with singleton and named registration patterns
- **Type queries** for exact type matching and interface compatibility
- **Container introspection** with multiple output formats (human, JSON, YAML)

### 3. Application Orchestration
- **Lifecycle management** with configurable phases (build → link → start → stop)
- **Factory pattern** for configuration-driven object creation
- **Dependency injection** through automatic linking of compatible objects
- **Service discovery** via container-based object lookup

## Learning Path

Follow these guides in order to master the df framework:

### 1. [Getting Started](/guides/getting-started/)
Start here to learn the basics:
- Installation and setup
- Core concepts and philosophy  
- Your first df application
- Basic data binding examples

### 2. [Data Binding](/guides/data-binding/)
Master struct binding and configuration:
- Core binding functions (`df.New`, `df.Bind`, `df.Unbind`, `df.Merge`)
- Struct tags and field mapping
- Type coercion and conversion
- File operations (JSON/YAML)
- Configuration layering patterns

### 3. [Dependency Injection](/guides/dependency-injection/)
Learn container-based object management:
- Object registration (singleton and named)
- Service discovery and type queries
- Dependency injection patterns
- Container introspection
- Testing with mocks

### 4. [Application Lifecycle](/guides/application-lifecycle/)
Build complete applications with orchestration:
- Application phases (Build → Link → Start → Stop)
- Factory pattern for object creation
- Lifecycle interfaces (`Startable`, `Stoppable`, `Linkable`)
- Configuration-driven architecture
- Error handling and recovery

### 5. [Advanced Features](/guides/advanced-features/)
Explore sophisticated patterns:
- Dynamic types for polymorphic data
- Object references with cycle handling
- Custom converters and marshalers
- Plugin architectures
- Performance optimization

## Quick Examples

### Simple Data Binding
```go
type User struct {
    Name     string `df:"+required"`
    Email    string
    Age      int    
    Active   bool   `df:"is_active"`
    Password string `df:"+secret"`
}

data := map[string]any{
    "name":      "John Doe",
    "email":     "john@example.com", 
    "age":       30,
    "is_active": true,
    "password":  "secret123",
}

user, err := df.New[User](data)
```

### Configuration Loading
```go
type Config struct {
    Database DatabaseConfig `df:"database"`
    Server   ServerConfig   `df:"server"`
}

config, err := df.NewFromYAML[Config]("config.yaml")
```

### Complete Application
```go
app := df.NewApplication(config)
df.WithFactory(app, &DatabaseFactory{})
df.WithFactory(app, &UserServiceFactory{})

app.Initialize() // Build + Link phases
app.Start()      // Start all services
// ... application runs ...
app.Stop()       // Graceful shutdown
```

## Key Features

- **Bidirectional Binding**: Core functions `df.Bind()`, `df.Unbind()`, `df.New[T]()`, and `df.Merge()` for struct-to-map conversion
- **Struct Tag Configuration**: Field mapping using `df` tags with support for custom names, required fields, and field exclusion
- **Type Coercion**: Comprehensive type handling including primitives, pointers, slices, time.Duration, and nested structs with automatic coercion
- **Polymorphic Data**: `df.Dynamic` interface with global and field-specific type binders for runtime type discrimination
- **Object References**: Generic `df.Pointer[T]` type with two-phase bind-and-link process supporting cycles and caching
- **File I/O**: Direct JSON and YAML file binding/unbinding with `BindFromJSON/YAML()` and `UnbindToJSON/YAML()` functions
- **Custom Conversion**: `df.Converter` interface and `df.Marshaler`/`df.Unmarshaler` interfaces for specialized type handling

## Use Cases

### Configuration Management
Layer configuration from defaults → files → environment → command line flags with `df.Merge()`.

### Plugin Systems
Load and configure plugins dynamically based on runtime configuration with the container and Dynamic types.

### Microservice Orchestration
Coordinate multiple services with shared dependencies using factories and lifecycle management.

### Data Processing Pipelines
Build flexible data processing systems where pipeline components are configured at runtime.

### Infrastructure as Code
Define infrastructure components in configuration and have them created and linked automatically.

## Getting Started

Ready to dive in? **[Start with the Getting Started guide](/guides/getting-started/)** to learn the fundamentals and build your first df application.

For specific topics, jump to:
- **[Data Binding](/guides/data-binding/)** - Master struct binding and configuration
- **[Dependency Injection](/guides/dependency-injection/)** - Learn container-based object management
- **[Application Lifecycle](/guides/application-lifecycle/)** - Build complete applications with orchestration
- **[Advanced Features](/guides/advanced-features/)** - Explore Dynamic types, object references, and custom converters
