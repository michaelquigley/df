# df

A comprehensive Go framework for building dynamic, configuration-driven applications. df provides a complete stack from low-level data binding to high-level application orchestration, enabling systems that can reconfigure their internal architecture based on runtime configuration.

## Overview

df consists of three integrated layers:

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

## Key Features

### Data Binding Layer
- **Bind/Unbind** data between Go structs and structured formats
- **New[T]** generic function for type-safe allocation and binding
- **Merge** data into pre-initialized structs for configuration layering
- **Flexible field mapping** with `df` struct tags and validation
- **Custom converters** for specialized type handling
- **Round-trip compatibility** ensuring data integrity

### Container Layer  
- **Singleton objects** registered and retrieved by type
- **Named objects** supporting multiple instances of the same type
- **Type queries** with `OfType[T]()` and `AsType[T]()` functions
- **Container inspection** for debugging and monitoring

### Application Layer
- **Configuration-driven** object creation through factory registration
- **Lifecycle interfaces** (`Startable`, `Stoppable`, `Linkable`) for managed services
- **Dependency injection** with automatic resolution during linking phase
- **Graceful startup/shutdown** with proper dependency ordering

## Quick Start

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

    // 6. Inspect container contents
    fmt.Println("\n=== container contents ===")
    output, _ := app.C.Inspect(df.InspectHuman)
    fmt.Println(output)

    // 7. Graceful shutdown
    if err := app.Stop(); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

### Data Binding Only

For simple data binding without the application framework:

```go
type User struct {
    Name     string `df:"required"`
    Email    string
    Age      int    
    Active   bool   `df:"is_active"`
    Password string `df:"secret"`
}

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
```

## Container: Dependency Injection Made Simple

The `Container` type provides a modern dependency injection system with both singleton and named object registration:

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

The `Application` type orchestrates object creation, dependency injection, and lifecycle management:

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

### Configuration Integration

```go
// Load configuration from multiple sources
app := df.NewApplication(defaultConfig)

// Layer 1: Configuration file
app.Initialize("config.yaml")

// Layer 2: Environment-specific overrides  
app.Initialize("config.prod.yaml")

// Layer 3: Build and start
app.Start()
```

## Struct Tags

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

## Embedded Structs

df fully supports Go's embedded struct feature, automatically flattening embedded fields during binding and unbinding operations:

```go
type Person struct {
    Name string `df:"name"`
    Age  int    `df:"age"`
}

type Employee struct {
    Person // embedded struct - fields are promoted to parent level
    Title  string `df:"title"`
    Salary int    `df:"salary"`
}

// Input data - embedded fields appear at the top level
data := map[string]any{
    "name":   "John Doe",    // from embedded Person
    "age":    30,            // from embedded Person
    "title":  "Engineer",
    "salary": 75000,
}

// Works with all binding functions
employee, err := df.New[Employee](data)
// employee.Name == "John Doe" (promoted from embedded Person)

result, err := df.Unbind(employee)  
// result flattens embedded fields: {"name": "John Doe", "age": 30, ...}
```

**Key features:**
- **Field promotion**: Embedded struct fields appear at the parent level in data
- **Pointer embedding**: Supports both value (`Person`) and pointer (`*Person`) embedding
- **Smart allocation**: Pointer embedded structs only allocated when their fields are present
- **Deep nesting**: Multiple levels of embedding work seamlessly
- **Tag inheritance**: Embedded fields respect their original `df` struct tags

## New[T] vs Bind

df provides two ways to populate structs from data:

### New[T] - Generic Type-Safe Allocation

The `New[T]` function provides a modern, type-safe approach using Go generics. It automatically allocates a new instance of type `T` and returns a pointer to the populated struct:

```go
type Config struct {
    Host string `df:"host"`
    Port int    `df:"port"`
}

data := map[string]any{
    "host": "localhost",
    "port": 8080,
}

// Automatic allocation with compile-time type safety
config, err := df.New[Config](data)
if err != nil {
    // handle error
}
// config is *Config, ready to use
```

### Bind - Manual Allocation Control

The `Bind` function provides more control over object allocation, useful when you need to:
- Bind to pre-initialized structs with default values
- Control where and how objects are allocated
- Work with interfaces or complex allocation patterns

```go
// Option 1: Zero-value allocation
var config Config
err := df.Bind(&config, data)

// Option 2: Pre-initialized with defaults
config := Config{
    Host: "0.0.0.0",  // default value
    Port: 3000,       // default value
}
err := df.Bind(&config, data) // only overrides provided fields

// Option 3: Custom allocation
config := &Config{}
err := df.Bind(config, data)
```

### When to Use Each

- **Use `New[T]`** when you want simple, type-safe allocation for most common use cases
- **Use `Bind`** when you need control over allocation, pre-initialized defaults, or working with interfaces

Both functions support the same features: struct tags, nested structures, custom converters, dynamic fields, etc.

## Merge - Building Defaults Systems

The `Merge` function provides a powerful way to build configuration systems with sensible defaults. Unlike `Bind` and `New[T]` which populate empty structs, `Merge` overlays external data onto pre-initialized structs, preserving any existing values that aren't overridden.

### Basic Merge Usage

```go
type ServerConfig struct {
    Host    string `df:"host"`
    Port    int    `df:"port"`
    Timeout int    `df:"timeout"`
    Debug   bool   `df:"debug"`
}

// Start with a struct containing sensible defaults
config := &ServerConfig{
    Host:    "localhost",
    Port:    8080,
    Timeout: 30,
    Debug:   false,
}

// External configuration (from file, environment, CLI, etc.)
userConfig := map[string]any{
    "host": "api.example.com",
    "debug": true,
    // Note: port and timeout are not specified
}

// Merge preserves unspecified defaults
err := df.Merge(config, userConfig)
// Result: Host="api.example.com", Port=8080 (preserved), 
//         Timeout=30 (preserved), Debug=true
```

### Layered Configuration Systems

`Merge` enables sophisticated configuration hierarchies where defaults can be progressively overridden:

```go
type AppConfig struct {
    Server   ServerConfig   `df:"server"`
    Database DatabaseConfig `df:"database"`
    Features []string       `df:"features"`
}

// Layer 1: Application defaults
config := &AppConfig{
    Server: ServerConfig{
        Host:    "localhost",
        Port:    8080,
        Timeout: 30,
        Debug:   false,
    },
    Database: DatabaseConfig{
        Host:     "localhost",
        Port:     5432,
        Database: "myapp",
        SSL:      true,
    },
    Features: []string{"basic", "auth"},
}

// Layer 2: Environment-specific overrides
envConfig := map[string]any{
    "server": map[string]any{
        "host": "prod-server.example.com",
        "timeout": 60,
    },
    "database": map[string]any{
        "host": "prod-db.example.com",
    },
}

err := df.Merge(config, envConfig)

// Layer 3: User-specific overrides
userConfig := map[string]any{
    "server": map[string]any{
        "debug": true,
    },
    "features": []string{"basic", "auth", "premium"},
}

err = df.Merge(config, userConfig)

// Final result combines all layers:
// - Server.Host: "prod-server.example.com" (from env)
// - Server.Port: 8080 (preserved from defaults)
// - Server.Timeout: 60 (from env)
// - Server.Debug: true (from user)
// - Database.Host: "prod-db.example.com" (from env)
// - Database.Port: 5432 (preserved from defaults)
// - Features: ["basic", "auth", "premium"] (from user, replaces entirely)
```

### Configuration Sources Integration

`Merge` works seamlessly with multiple configuration sources:

```go
// Start with compiled-in defaults
config := getDefaultConfig()

// Layer 1: Configuration file
if configExists("app.yaml") {
    err := df.MergeFromYAML(config, "app.yaml")
    if err != nil {
        return err
    }
}

// Layer 2: Environment variables (converted to map)
envVars := getEnvironmentOverrides()
err := df.Merge(config, envVars)

// Layer 3: Command line flags (converted to map)
cliFlags := getCLIOverrides()
err = df.Merge(config, cliFlags)

// Final config reflects the complete hierarchy
```

### Nested Struct Preservation

`Merge` intelligently handles nested structures, preserving defaults at all levels:

```go
type DatabaseConfig struct {
    Host     string        `df:"host"`
    Port     int          `df:"port"`
    Pool     PoolConfig   `df:"pool"`
    Features []string     `df:"features"`
}

type PoolConfig struct {
    MinSize int `df:"min_size"`
    MaxSize int `df:"max_size"`
    Timeout int `df:"timeout"`
}

// Defaults with nested configuration
config := &DatabaseConfig{
    Host: "localhost",
    Port: 5432,
    Pool: PoolConfig{
        MinSize: 5,
        MaxSize: 20,
        Timeout: 30,
    },
    Features: []string{"ssl", "pooling"},
}

// Partial override - only changes MaxSize
override := map[string]any{
    "pool": map[string]any{
        "max_size": 50,
    },
}

err := df.Merge(config, override)
// Result preserves Host, Port, Pool.MinSize, Pool.Timeout, Features
// Only Pool.MaxSize changes to 50
```

### Key Benefits

- **Intuitive Defaults**: Define sensible defaults directly in Go structs
- **Selective Overrides**: Users only specify what they want to change
- **Deep Merging**: Nested structures are merged intelligently
- **Configuration Hierarchies**: Layer multiple configuration sources
- **Type Safety**: Leverage Go's type system for configuration validation
- **Backward Compatibility**: Adding new fields with defaults doesn't break existing configs

### Merge vs Bind vs New[T]

| Function | Use Case | Object State | Best For |
|----------|----------|--------------|----------|
| `New[T]` | Create from scratch | Empty → Populated | Simple binding, new objects |
| `Bind` | Populate existing | Any → Populated | Manual allocation control |
| `Merge` | Overlay onto defaults | Defaults → Enhanced | Configuration systems, defaults |

See [examples/df_defaults](examples/df_defaults) for a complete working example of building configuration systems with `Merge`.

## Custom Marshaling and Unmarshaling

For types that require custom logic for binding and unbinding, `df` supports the `Unmarshaler` and `Marshaler` interfaces. This allows a type to take full control over how it is converted from or to structured data.

### The Unmarshaler Interface

A type implements the `Unmarshaler` interface by defining an `UnmarshalDf` method. When `df.Bind` encounters a type that satisfies this interface, it will call this method to populate the struct, bypassing the default reflection-based binding logic for that type.

```go
// Unmarshaler is the interface implemented by types that can unmarshal a
// df description of themselves.
type Unmarshaler interface {
    UnmarshalDf(data any) error
}
```

#### Example

```go
import (
    "fmt"
    "time"
)

// CustomTime wraps time.Time to support a custom date format.
type CustomTime struct {
    time.Time
}

// UnmarshalDf implements the df.Unmarshaler interface.
func (c *CustomTime) UnmarshalDf(data any) error {
    if dateStr, ok := data.(string); ok {
        t, err := time.Parse("2006-01-02", dateStr)
        if err != nil {
            return err
        }
        c.Time = t
        return nil
    }
    return fmt.Errorf("expected string for CustomTime, got %T", data)
}
```

### The Marshaler Interface

A type implements the `Marshaler` interface by defining a `MarshalDf` method. When `df.Unbind` encounters a type that satisfies this interface, it will call this method to convert the type into its data representation, bypassing the default reflection-based unbinding logic.

```go
// Marshaler is the interface implemented by types that can marshal themselves
// into a df description.
type Marshaler interface {
    MarshalDf() (any, error)
}
```

#### Example

```go
// MarshalDf implements the df.Marshaler interface.
func (c CustomTime) MarshalDf() (any, error) {
    return c.Time.Format("2006-01-02"), nil
}
```

With these interfaces, you can integrate types that don't follow standard struct conventions seamlessly into the `df` binding and unbinding process.

## Custom Field Converters

For specialized type conversion and validation, df supports custom field converters through the `Converter` interface. This is particularly useful for domain-specific types, validation during binding, or handling multiple input formats for the same logical type.

### The Converter Interface

```go
// Converter defines a bidirectional type conversion interface for custom field types.
type Converter interface {
    // FromRaw converts a raw value (from the data map) to the target type.
    FromRaw(raw interface{}) (interface{}, error)
    
    // ToRaw converts a typed value back to a raw value for serialization.
    ToRaw(value interface{}) (interface{}, error)
}
```

### Example: Email Validation Converter

```go
// Email represents a validated email address
type Email string

// EmailConverter handles conversion and validation
type EmailConverter struct{}

func (c *EmailConverter) FromRaw(raw interface{}) (interface{}, error) {
    s, ok := raw.(string)
    if !ok {
        return nil, fmt.Errorf("expected string for email, got %T", raw)
    }
    
    // basic email validation
    if !strings.Contains(s, "@") {
        return nil, fmt.Errorf("invalid email format: %s", s)
    }
    
    return Email(s), nil
}

func (c *EmailConverter) ToRaw(value interface{}) (interface{}, error) {
    email, ok := value.(Email)
    if !ok {
        return nil, fmt.Errorf("expected Email, got %T", value)
    }
    return string(email), nil
}

// Usage
opts := &df.Options{
    Converters: map[reflect.Type]df.Converter{
        reflect.TypeOf(Email("")): &EmailConverter{},
    },
}

type User struct {
    Email Email `df:"email"`
    Name  string `df:"name"`
}

var user User
err := df.Bind(&user, data, opts) // validates email during binding
```

### Benefits

- **Type Safety**: Ensures data conforms to expected formats before binding
- **Validation**: Built-in validation during the binding process  
- **Flexibility**: Supports multiple input formats for the same logical type
- **Reusability**: Converters can be used across different struct definitions
- **Bidirectional**: Works seamlessly with both bind and unbind operations

See the `examples/df_converters` directory for a complete example with multiple converter types.

## Configuration Inspection

The `Inspect` function provides human-readable output for debugging bound configuration with global vertical alignment and secret field filtering:

```go
type Config struct {
    Host     string `df:"host"`
    Port     int    `df:"port"`
    APIKey   string `df:"api_key,secret"`
    Database *DBConfig `df:"database"`
}

type DBConfig struct {
    Host     string `df:"host"`
    Password string `df:"password,secret"`
}

// Inspect with secrets hidden (default)
output, _ := df.Inspect(config)
fmt.Println(output)
// Config {
//   host           : "localhost"
//   port           : 8080
//   api_key (secret): <set>
//   database       : DBConfig {
//     host           : "db.example.com"
//     password (secret): <set>
//   }
// }

// Inspect with secrets visible
output, _ = df.Inspect(config, &df.InspectOptions{ShowSecrets: true})
// Shows actual secret values instead of <set>/<unset>
```

### InspectOptions

```go
type InspectOptions struct {
    MaxDepth    int    // Recursion depth limit (default: 10)
    Indent      string // Indentation string (default: "  ")
    ShowSecrets bool   // Show secret field values (default: false)
}
```

## File Operations

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

## Dynamic Fields

Support polymorphic data structures with the `Dynamic` interface for fields that can be different concrete types based on runtime data.

### Basic Dynamic Interface

```go
type Dynamic interface {
    Type() string              // Returns the discriminator string
    ToMap() map[string]any    // Converts the struct to a map
}
```

### Example Implementation

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

type SlackAction struct {
    Channel string `df:"channel"`
    Message string `df:"message"`
}

func (s SlackAction) Type() string { return "slack" }
func (s SlackAction) ToMap() map[string]any {
    return map[string]any{
        "channel": s.Channel,
        "message": s.Message,
    }
}

// Use Dynamic fields in structs
type Notification struct {
    Name   string  `df:"name"`
    Action Dynamic `df:"action"`  // Polymorphic field
}
```

### Binding with Dynamic Binders

```go
func main() {
    // Input data with type discriminator
    data := map[string]any{
        "name": "Welcome Email",
        "action": map[string]any{
            "type":      "email",           // Discriminator field
            "recipient": "user@example.com",
            "subject":   "Welcome!",
        },
    }
    
    // Configure binders for different types
    opts := &df.Options{
        DynamicBinders: map[string]func(map[string]any) (df.Dynamic, error){
            "email": func(m map[string]any) (df.Dynamic, error) {
                // Use New[T] for cleaner allocation
                action, err := df.New[EmailAction](m)
                if err != nil {
                    return nil, err
                }
                return *action, nil
            },
            "slack": func(m map[string]any) (df.Dynamic, error) {
                action, err := df.New[SlackAction](m)
                if err != nil {
                    return nil, err
                }
                return *action, nil
            },
        },
    }
    
    // Use New[T] for the main struct too
    notification, err := df.New[Notification](data, opts)
    if err != nil {
        panic(err)
    }
    
    // Access the concrete type
    if emailAction, ok := notification.Action.(EmailAction); ok {
        fmt.Printf("Email to: %s\n", emailAction.Recipient)
    }
}
```

### Field-Specific Dynamic Binders

For more granular control, use `FieldDynamicBinders` to specify different binder sets per field:

```go
opts := &df.Options{
    FieldDynamicBinders: map[string]map[string]func(map[string]any) (df.Dynamic, error){
        "Notification.Action": {  // Field path
            "email": emailBinder,
            "slack": slackBinder,
        },
        "Workflow.Steps": {       // Different binders for different fields
            "http":  httpBinder,
            "delay": delayBinder,
        },
    },
}
```

### Dynamic Slices

Dynamic fields work seamlessly with slices:

```go
type Workflow struct {
    Name  string    `df:"name"`
    Steps []Dynamic `df:"steps"`  // Slice of polymorphic types
}

// Input data
data := map[string]any{
    "name": "User Onboarding",
    "steps": []any{
        map[string]any{"type": "email", "recipient": "user@example.com"},
        map[string]any{"type": "slack", "channel": "#general"},
    },
}
```

### Round-trip Compatibility

Dynamic fields maintain their discriminator during unbind operations:

```go
// Unbind preserves the "type" field automatically
result, err := df.Unbind(&notification)
// result["action"]["type"] == "email"
```

The `type` field in the input data determines which binder function is called to create the appropriate concrete type. During unbind, the `Type()` method ensures the discriminator is preserved for round-trip compatibility.

## Pointer References

Support object references with cycle handling using `df.Pointer[T]` and the `df.Identifiable` interface.

### Basic Usage

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
```

### Two-Phase Process

```go
// Phase 1: Bind data with $ref strings
var container DataContainer
df.Bind(&container, data)

// Phase 2: Resolve all pointer references
df.Link(&container)

// Access resolved objects
author := container.Documents[0].Author.Resolve()
```

### Advanced Linking with `df.Linker`

For more control over the linking process, use the `df.Linker` type. This is useful for scenarios like multi-stage linking or performance optimization with caching.

```go
// Create a linker with caching enabled
linker := df.NewLinker(df.LinkerOptions{
    EnableCaching: true,
})

// Use the linker to resolve references
linker.Link(&container) 
```

The `Linker` provides options for:
- **`EnableCaching`**: Caches object registries to speed up repeated linking operations on the same data. Disabled by default.
- **`AllowPartialResolution`**: Allows linking to succeed even if some references cannot be found.
- **Multi-stage linking**: Use `Register()` to collect objects from multiple sources before calling `ResolveReferences()`.

### JSON Structure

```json
{
  "users": [
    {"id": "user1", "name": "Alice"}
  ],
  "documents": [
    {
      "id": "doc1",
      "title": "Guide", 
      "author": {"$ref": "user1"}
    }
  ]
}
```

### Key Features

- **Type Safety**: Generic `Pointer[T]` ensures compile-time type checking
- **Cycle Support**: Two-phase binding naturally handles circular references
- **ID Namespacing**: Objects with same ID but different types don't clash (e.g., `User:1` vs `Document:1`)
- **Round-trip Compatible**: Bind/Link/Unbind preserves reference structure

See [examples/df_pointers](examples/df_pointers) for a complete working example.

## Architecture & Use Cases

df enables a range of application architectures from simple data binding to complex distributed systems:

### Simple Configuration Loading
```go
// Load application settings
type Config struct {
    Database DatabaseConfig `df:"database"`
    Server   ServerConfig   `df:"server"`
}

config, err := df.NewFromYAML[Config]("config.yaml")
```

### Dependency Injection Applications
```go
// Build applications with automatic dependency resolution
app := df.NewApplication(config)
df.WithFactory(app, &DatabaseFactory{})
df.WithFactory(app, &APIFactory{})
df.WithFactory(app, &WorkerFactory{})

app.Initialize()
app.Start()
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
df.WithFactory(app, &NotificationServiceFactory{})

app.Initialize()
app.Start()

// Services automatically discover and link to shared infrastructure
```

## Current Status

df provides a complete, production-ready framework with three integrated layers:

### ✅ Data Binding Layer (Complete)
- **Bidirectional binding** between Go structs and structured data formats
- **Type-safe conversion** with comprehensive type system support
- **Polymorphic data** via Dynamic interface for runtime type discrimination
- **Object references** with cycle-safe pointer resolution
- **Custom marshaling** with Marshaler/Unmarshaler interfaces
- **Configuration merging** for layered configuration systems

### ✅ Container Layer (Complete)  
- **Dependency injection** with singleton and named object registration
- **Type queries** for exact type matching and interface compatibility
- **Container introspection** with multiple output formats
- **Object lifecycle** management with automatic cleanup

### ✅ Application Layer (Complete)
- **Lifecycle orchestration** with configurable phases (build → link → start → stop)
- **Factory pattern** for configuration-driven object creation
- **Dependency injection** with automatic resolution during linking
- **Configuration integration** supporting multiple file sources

## Roadmap

Future enhancements will build upon this solid foundation:

### Enhanced Configuration
- **Schema validation** for configuration structures
- **Hot-reload** capabilities for runtime reconfiguration
- **Environment templating** for deployment-specific configurations

### Advanced Patterns
- **Plugin discovery** with automatic registration
- **Service mesh** integration for distributed applications  
- **Observability** integration with metrics and tracing

## License

See [LICENSE](LICENSE) file for details.
