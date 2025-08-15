# df

A lightweight Go library for binding and unbinding structured data to/from Go structs using reflection. df serves as the foundational layer for building dynamic, configuration-driven Go applications that can reconfigure their internal architecture based on runtime configuration.

## Features

- **Bind** data from maps, JSON, and YAML files to Go structs
- **New[T]** generic function for automatic allocation and binding
- **Merge** data from maps, JSON, and YAML files into pre-built Go structs (default settings, etc.)
- **Unbind** Go structs back to maps, JSON, and YAML files
- **Flexible field mapping** with `df` struct tags
- **Type coercion** for primitives, pointers, slices, and nested structs
- **Custom field converters** for specialized type conversion and validation
- **Dynamic field resolution** for polymorphic data structures
- **Pointer references** with cycle handling for complex object relationships (see `Pointer`)
- **Custom marshaling/unmarshaling** with `Marshaler` and `Unmarshaler` interfaces
- **Round-trip compatibility** between bind and unbind operations

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/michaelquigley/df"
)

type User struct {
    Name   string `df:"name,required"`
    Email  string `df:"email"`
    Age    int    `df:"age"`
    Active bool   `df:"active"`
}

func main() {
    // Input data
    data := map[string]any{
        "name":   "John Doe",
        "email":  "john@example.com", 
        "age":    30,
        "active": true,
    }
    
    // Option 1: Use New[T] for automatic allocation
    user, err := df.New[User](data)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("%+v\n", *user) // {Name:John Doe Email:john@example.com Age:30 Active:true}
    
    // Option 2: Use Bind with pre-allocated struct  
    var user2 User
    err = df.Bind(&user2, data)
    if err != nil {
        panic(err)
    }
    
    // Unbind back to map
    result, err := df.Unbind(user)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("%+v\n", result) // map[active:true age:30 email:john@example.com name:John Doe]
}
```

## Vision: Dynamic System Construction

df serves as the foundational layer for building dynamic, configuration-driven Go applications. While traditional Go applications have fixed structures determined at compile time, df enables systems that can reconfigure their internal architecture based on runtime configuration.

### The Three Layers

1. **Data Binding Layer** (Current): Robust mapping between structured data and Go types
2. **Component Registry Layer** (Planned): Dynamic instantiation of registered component types  
3. **System Orchestration Layer** (Planned): Lifecycle management and dependency injection

### From Static to Dynamic

```go
// traditional static approach
server := &http.Server{
    Handler: &MyHandler{},
    Addr:    ":8080",
}

// df-enabled dynamic approach  
config := map[string]any{
    "type": "http_server",
    "addr": ":8080", 
    "handler": map[string]any{
        "type": "my_handler",
        "routes": []any{...},
    },
}

var server Component
df.Bind(&server, config)  // creates the right concrete types
```

This foundation enables applications that can be reconfigured without recompilation, supporting use cases like:
- **Plugin architectures** - Load and configure components dynamically
- **A/B testing** - Switch between different component implementations
- **Environment-specific topologies** - Different system layouts per environment  
- **Configuration-driven composition** - Assemble complex systems from simple parts

## Struct Tags

Control field binding behavior with `df` struct tags:

```go
type Example struct {
    Name     string `df:"custom_name,required"` // Custom field name, required
    Email    string `df:"email"`                // Custom field name
    Age      int    `df:",required"`            // Default name (snake_case), required  
    Internal string `df:"-"`                    // Skip this field
    Default  string                             // Uses snake_case: "default"
}
```

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
var container Container
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

## Current Capabilities

The current df implementation provides the essential data binding layer with these key capabilities:

### Core Binding Operations
- **Bidirectional data mapping** between Go structs and structured data (JSON, YAML, maps)
- **Type-safe conversion** with support for primitives, pointers, slices, and nested structures
- **Flexible field mapping** via struct tags with custom naming and validation rules

### Advanced Features  
- **Polymorphic data structures** via the Dynamic interface for runtime type selection
- **Object references** with cycle-safe pointer resolution using df.Pointer[T]
- **Custom marshaling/unmarshaling** with `Marshaler` and `Unmarshaler` interfaces
- **Round-trip compatibility** ensuring data integrity across bind/unbind operations

### Foundation for Dynamic Systems
Today's df provides the building blocks that future layers will leverage:
- **Structured data normalization** - Converting various input formats to Go types
- **Type discrimination** - Runtime selection of concrete types based on configuration
- **Object relationship mapping** - Managing complex interconnected data structures

## Roadmap

df is the foundational component in a _dynamic framework_ approach to building Go applications. The next phases will build upon this solid data binding foundation:

### Phase 2: Component Registry
- Registration and discovery of component types
- Factory pattern integration with df binding
- Plugin loading and configuration

### Phase 3: System Orchestration  
- Dependency injection and lifecycle management
- Configuration validation and schema enforcement
- Hot-reload capabilities for runtime reconfiguration

## License

See [LICENSE](LICENSE) file for details.
