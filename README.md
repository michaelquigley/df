# df

A lightweight Go library for binding and unbinding structured data to/from Go structs using reflection. df serves as the foundational layer for building dynamic, configuration-driven Go applications that can reconfigure their internal architecture based on runtime configuration.

## Features

- **Bind** data from maps, JSON, and YAML files to Go structs
- **Unbind** Go structs back to maps, JSON, and YAML files
- **Flexible field mapping** with `df` struct tags
- **Type coercion** for primitives, pointers, slices, and nested structs
- **Dynamic field resolution** for polymorphic data structures
- **Pointer references** with cycle handling for complex object relationships
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
    // Bind from map
    data := map[string]any{
        "name":   "John Doe",
        "email":  "john@example.com", 
        "age":    30,
        "active": true,
    }
    
    var user User
    err := df.Bind(&user, data)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("%+v\n", user) // {Name:John Doe Email:john@example.com Age:30 Active:true}
    
    // Unbind back to map
    result, err := df.Unbind(&user)
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
// Traditional static approach
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
df.Bind(&server, config)  // Creates the right concrete types
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
                var action EmailAction
                err := df.Bind(&action, m)
                return action, err
            },
            "slack": func(m map[string]any) (df.Dynamic, error) {
                var action SlackAction
                err := df.Bind(&action, m)
                return action, err
            },
        },
    }
    
    var notification Notification
    err := df.Bind(&notification, data, opts)
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
    ID     string          `df:"id"`
    Title  string          `df:"title"`
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