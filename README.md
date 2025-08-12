# df

A lightweight Go library for binding and unbinding structured data to/from Go structs using reflection. Supports JSON, YAML, and map[string]any data sources with customizable field mapping via struct tags.

## Features

- **Bind** data from maps, JSON, and YAML files to Go structs
- **Unbind** Go structs back to maps, JSON, and YAML files
- **Flexible field mapping** with `df` struct tags
- **Type coercion** for primitives, pointers, slices, and nested structs
- **Dynamic field resolution** for polymorphic data structures
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

## Roadmap

df is a foundational component in a _dynamic framework_ approach to building golang applications. A dynamic framework application is designed to reconfigure its internal landscape based on configuration structures.

## License

See [LICENSE](LICENSE) file for details.