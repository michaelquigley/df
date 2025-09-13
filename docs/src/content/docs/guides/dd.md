---
title: dynamic foundation for data
description: Complete reference for the dd package - bidirectional data binding between Go structs and maps.
---

**Convert between Go structs and maps with ease**

The `dd` package provides bidirectional data binding between Go structs and `map[string]any`, enabling seamless integration with any data system. Since maps are foundational to all data structures, this facilitates integration with networks, databases, files, and APIs.

## Quick Reference

### 1. Basic Binding - Hello World

**Convert struct to map and back**

```go
import "github.com/michaelquigley/df/dd"

// struct → map
user := User{Name: "John", Age: 30}
data, err := dd.Unbind(user)
// data: map[string]any{"name": "John", "age": 30}

// map → struct (modern way)
userData := map[string]any{"name": "Alice", "age": 25}
user, err := dd.New[User](userData)

// map → struct (manual allocation)
var user User
err := dd.Bind(&user, userData)
```

### 2. Struct Tags - Field Control

**Control field mapping and validation**

```go
type User struct {
    Name     string `dd:"+required"`         // required field
    Email    string `dd:"email_address"`     // custom field name
    Age      int    `dd:",+required"`        // default name, required
    Password string `dd:",+secret"`          // hidden in output
    Internal string `dd:"-"`                 // skip completely
    Active   bool                            // uses snake_case: "active"
}
```

**Tag Options:**
- `dd:"custom_name"` - custom field name
- `dd:"+required"` - field is required
- `dd:",+secret"` - hidden in inspect output
- `dd:"-"` - exclude from binding
- No tag = automatic snake_case conversion

### 3. Type Coercion - Automatic Conversion

**Automatic type conversion between compatible types**

```go
// Input data with different types
data := map[string]any{
    "port":     "8080",        // string → int
    "timeout":  30.5,          // float → int
    "enabled":  "true",        // string → bool
    "duration": "5m",          // string → time.Duration
}

type Config struct {
    Port     int           `dd:"port"`
    Timeout  int           `dd:"timeout"`
    Enabled  bool          `dd:"enabled"`
    Tags     []string      `dd:"tags"`
    Duration time.Duration `dd:"duration"`
}

config, err := dd.New[Config](data)
// All fields converted automatically
```

### 4. File I/O - Direct Persistence

**Read/write JSON and YAML files directly**

```go
// From files
config, err := dd.BindFromJSON[Config]("config.json")
config, err := dd.BindFromYAML[Config]("config.yaml")

// To files
err := dd.UnbindToJSON(config, "output.json")
err := dd.UnbindToYAML(config, "output.yaml")

// With formatting options
err := dd.UnbindToJSONIndent(config, "pretty.json", "", "  ")
```

### 5. Nested Structures - Complex Data

**Handle deeply nested data structures**

```go
type User struct {
    Name    string   `dd:"name"`
    Profile *Profile `dd:"profile"`     // pointer to nested struct
    Tags    []Tag    `dd:"tags"`        // slice of structs
}

type Profile struct {
    Bio     string `dd:"bio"`
    Website string `dd:"website"`
}

type Tag struct {
    Name  string `dd:"name"`
    Color string `dd:"color"`
}

// Nested data automatically handled
data := map[string]any{
    "name": "John",
    "profile": map[string]any{
        "bio":     "Developer",
        "website": "john.dev",
    },
    "tags": []any{
        map[string]any{"name": "go", "color": "blue"},
        map[string]any{"name": "web", "color": "green"},
    },
}

user, err := dd.New[User](data)
```

### 6. Validation - Required Fields and Errors

**Field validation and error handling**

```go
type User struct {
    Name  string `dd:"+required"`
    Email string `dd:"+required"`
    Age   int    `dd:",+required"`
}

// Missing required field
data := map[string]any{
    "name": "John",
    // email missing
    "age": 30,
}

user, err := dd.New[User](data)
// err: "User.Email: required field missing"

// Check for specific error types
if bindErr, ok := err.(*dd.BindError); ok {
    fmt.Printf("Field: %s, Error: %s\n", bindErr.Field, bindErr.Message)
}
```

### 7. Merge - Configuration Layering

**Overlay data onto existing structs with defaults**

```go
// Start with defaults
config := &ServerConfig{
    Host:    "localhost",
    Port:    8080,
    Timeout: 30,
    Debug:   false,
}

// Overlay user configuration (only overrides specified fields)
userConfig := map[string]any{
    "host":  "api.example.com",
    "debug": true,
    // port and timeout preserved from defaults
}

err := dd.Merge(config, userConfig)
// Result: Host="api.example.com", Port=8080, Timeout=30, Debug=true
```

### 8. Custom Converters - Specialized Types

**Handle custom types with validation**

```go
type Email string

type EmailConverter struct{}

func (c *EmailConverter) FromRaw(raw interface{}) (interface{}, error) {
    s, ok := raw.(string)
    if !ok {
        return nil, fmt.Errorf("expected string for email")
    }
    if !strings.Contains(s, "@") {
        return nil, fmt.Errorf("invalid email format")
    }
    return Email(s), nil
}

func (c *EmailConverter) ToRaw(value interface{}) (interface{}, error) {
    email, ok := value.(Email)
    if !ok {
        return nil, fmt.Errorf("expected Email type")
    }
    return string(email), nil
}

// Use converter
opts := &dd.Options{
    Converters: map[reflect.Type]dd.Converter{
        reflect.TypeOf(Email("")): &EmailConverter{},
    },
}

type User struct {
    Email Email `dd:"email"`
}

user, err := dd.New[User](data, opts) // validates email format
```

### 9. Custom Marshaling - Full Control

**Complete control over binding/unbinding**

```go
type CustomTime struct {
    time.Time
}

// Control how this type is created from data
func (c *CustomTime) UnmarshalDf(data any) error {
    dateStr, ok := data.(string)
    if !ok {
        return fmt.Errorf("expected string for CustomTime")
    }
    t, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return err
    }
    c.Time = t
    return nil
}

// Control how this type becomes data
func (c CustomTime) MarshalDf() (any, error) {
    return c.Time.Format("2006-01-02"), nil
}

// dd automatically uses these methods
```

### 10. Dynamic Types - Runtime Polymorphism

**Different types based on runtime data**

```go
// Types that implement Dynamic interface
type EmailAction struct {
    Recipient string `dd:"recipient"`
    Subject   string `dd:"subject"`
}

func (e EmailAction) Type() string { return "email" }
func (e EmailAction) ToMap() (map[string]any, error) {
    return map[string]any{
        "recipient": e.Recipient,
        "subject":   e.Subject,
    }, nil
}

type SlackAction struct {
    Channel string `dd:"channel"`
    Message string `dd:"message"`
}

func (s SlackAction) Type() string { return "slack" }
func (s SlackAction) ToMap() (map[string]any, error) {
    return map[string]any{
        "channel": s.Channel,
        "message": s.Message,
    }, nil
}

// Use in polymorphic fields
type Notification struct {
    Name   string    `dd:"name"`
    Action dd.Dynamic `dd:"action"`
}

// Configure type discrimination
opts := &dd.Options{
    DynamicBinders: map[string]func(map[string]any) (dd.Dynamic, error){
        "email": func(m map[string]any) (dd.Dynamic, error) {
            action, err := dd.New[EmailAction](m)
            return *action, err
        },
        "slack": func(m map[string]any) (dd.Dynamic, error) {
            action, err := dd.New[SlackAction](m)
            return *action, err
        },
    },
}

// Data with type discriminator
data := map[string]any{
    "name": "Welcome",
    "action": map[string]any{
        "type":      "email",           // discriminator
        "recipient": "user@example.com",
        "subject":   "Welcome!",
    },
}

notification, err := dd.New[Notification](data, opts)
```

### 11. Object References - Linked Data

**Handle object references with cycle detection**

```go
type User struct {
    ID   string `dd:"id"`
    Name string `dd:"name"`
}

func (u *User) GetId() string { return u.ID }

type Document struct {
    ID     string                `dd:"id"`
    Title  string                `dd:"title"`
    Author *dd.Pointer[*User]    `dd:"author"`
}

func (d *Document) GetId() string { return d.ID }

// Data with $ref references
data := map[string]any{
    "users": []any{
        map[string]any{"id": "user1", "name": "Alice"},
    },
    "documents": []any{
        map[string]any{
            "id":     "doc1",
            "title":  "Guide",
            "author": map[string]any{"$ref": "user1"},
        },
    },
}

// Two-phase process
var container DataContainer
dd.Bind(&container, data)  // Phase 1: bind with $ref strings
dd.Link(&container)        // Phase 2: resolve references

// Access resolved objects
author := container.Documents[0].Author.Resolve()
```

### 12. Advanced Linking - Performance and Control

**Advanced reference resolution with caching**

```go
// Create linker with options
linker := dd.NewLinker(dd.LinkerOptions{
    EnableCaching:           true,   // cache object registries
    AllowPartialResolution: false,  // fail if any refs unresolved
})

// Multi-stage linking for complex scenarios
linker.Register(&users)      // register objects from multiple sources
linker.Register(&documents)
linker.Register(&projects)

err := linker.ResolveReferences() // resolve all at once

// OR use convenience method
err := linker.Link(&container) // register + resolve in one call
```

## Core Functions

| Function | Purpose | Use Case |
|----------|---------|----------|
| `dd.New[T](data)` | Create struct from map | Type-safe allocation |
| `dd.Bind(&struct, data)` | Populate existing struct | Manual allocation control |
| `dd.Unbind(struct)` | Convert struct to map | Serialization, APIs |
| `dd.Merge(&struct, data)` | Overlay data on defaults | Configuration systems |
| `dd.BindFromJSON[T](file)` | Load from JSON file | Configuration loading |
| `dd.UnbindToYAML(struct, file)` | Save to YAML file | Configuration persistence |
| `dd.Link(&container)` | Resolve object references | Complex data relationships |

## Common Patterns

### Configuration Loading
```go
// Multi-layer configuration
config := getDefaultConfig()
dd.MergeFromYAML(config, "app.yaml")        // base config
dd.MergeFromYAML(config, "app.prod.yaml")   // environment
dd.Merge(config, getEnvOverrides())         // environment vars
```

### API Integration
```go
// HTTP API → struct
resp, _ := http.Get("https://api.example.com/user")
var data map[string]any
json.NewDecoder(resp.Body).Decode(&data)
user, _ := dd.New[User](data)

// struct → HTTP API
userData, _ := dd.Unbind(user)
json.NewEncoder(w).Encode(userData)
```

### Database Integration
```go
// Database row → struct
row := db.QueryRow("SELECT data FROM users WHERE id = ?", id)
var jsonData string
row.Scan(&jsonData)
var data map[string]any
json.Unmarshal([]byte(jsonData), &data)
user, _ := dd.New[User](data)
```

---

*See [dd/examples/](../../../dd/examples/) for complete working examples of each feature.*