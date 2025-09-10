---
title: Data Binding
description: Master struct binding, tags, type coercion, and file operations in the df framework.
---

Data binding is the foundation of the df framework. It provides bidirectional conversion between Go structs and structured data (maps, JSON, YAML) with comprehensive type handling and configuration options.

## Core Binding Functions

df provides four primary functions for data binding:

### df.New\[T\] - Type-Safe Allocation
Creates and populates a new struct instance:

```go
type User struct {
    Name  string `df:"+required"`
    Email string
    Age   int
}

data := map[string]any{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
}

user, err := df.New[User](data)
if err != nil {
    return err
}
// user is now *User with populated fields
```

### df.Bind() - Populate Existing Struct
Binds data into a pre-allocated struct:

```go
var user User
err := df.Bind(&user, data)
if err != nil {
    return err
}
// user struct is now populated
```

### df.Unbind() - Convert to Map
Converts a struct back to a map:

```go
userData, err := df.Unbind(&user)
if err != nil {
    return err
}
// userData is map[string]any{"name": "John Doe", ...}
```

### df.Merge() - Layer Configuration
Merges data into an existing struct, preserving existing values:

```go
// Start with defaults
config := &ServerConfig{
    Host:    "localhost",
    Port:    8080,
    Timeout: 30,
}

// Layer environment variables
envOverrides := map[string]any{
    "port": 9000,
    "host": "0.0.0.0",
}
err := df.Merge(config, envOverrides)
// config.Port is now 9000, config.Timeout remains 30
```

## Struct Tags

Control binding behavior with `df` struct tags:

### Basic Tag Syntax

```go
type Example struct {
    Name     string `df:"custom_name"`       // Custom field name
    Email    string `df:"zmail,+required"`    // Custom name + required
    Age      int    `df:",+required"`         // Default name + required  
    Password string `df:",+secret"`           // Default name + secret
    Internal string `df:"-"`                 // Skip this field
    Default  string                          // Uses snake_case: "default"
}
```

### Tag Options

#### Custom Field Names
Map struct fields to different data keys:

```go
type User struct {
    FirstName string `df:"First"`
    LastName  string `df:"Last"`
    EmailAddr string `df:"Email"`
}

data := map[string]any{
    "First": "John",
    "Last":  "Doe", 
    "Email":      "john@example.com",
}
```

#### Required Fields
Mark fields as mandatory:

```go
type Config struct {
    APIKey   string `df:"api_key,+required"`
    Database string `df:"db_url,+required"`
    LogLevel string `df:"log_level"` // Optional
}

// Binding will fail if api_key or db_url are missing
```

#### Secret Fields
Hide sensitive data from inspection:

```go
type Credentials struct {
    Username string `df:"username"`
    Password string `df:"password,+secret"`
    APIToken string `df:"api_token,+secret"`
}

creds, _ := df.New[Credentials](data)
output, _ := df.Inspect(creds)
// Password and APIToken will show as "[hidden]" in output
```

#### Excluded Fields
Skip fields entirely:

```go
type User struct {
    Name     string `df:"name"`
    Email    string `df:"email"`
    internal string `df:"-"`        // Never bound
    computed int    `df:"-"`        // Never bound
}
```

## Type Coercion

df automatically handles type conversion between compatible types:

### Primitive Types

```go
type Config struct {
    Port    int     `df:"port"`
    Timeout float64 `df:"timeout"`
    Enabled bool    `df:"enabled"`
    Name    string  `df:"name"`
}

// All of these work:
data := map[string]any{
    "port":    "8080",        // string -> int
    "timeout": 30,            // int -> float64
    "enabled": "true",        // string -> bool
    "name":    123,           // int -> string
}
```

### Time.Duration

```go
type Config struct {
    Timeout   time.Duration `df:"timeout"`
    KeepAlive time.Duration `df:"keep_alive"`
}

data := map[string]any{
    "timeout":    "30s",      // string -> Duration
    "keep_alive": 300000000,  // int64 nanoseconds -> Duration
}
```

### Pointers
Automatic pointer handling with nil safety:

```go
type User struct {
    Name  string  `df:"name"`
    Age   *int    `df:"age"`    // Optional field
    Email *string `df:"email"`  // Optional field
}

data := map[string]any{
    "name": "John",
    "age":  30,
    // email is missing - will be nil
}
```

### Slices and Arrays

```go
type Config struct {
    Hosts    []string `df:"hosts"`
    Ports    []int    `df:"ports"`
    Features []bool   `df:"features"`
}

data := map[string]any{
    "hosts":    []any{"web1", "web2", "web3"},
    "ports":    []any{8080, 9000, 3000},
    "features": []any{true, false, true},
}
```

### Nested Structs

```go
type Address struct {
    Street  string `df:"street"`
    City    string `df:"city"`
    Country string `df:"country"`
}

type User struct {
    Name    string  `df:"name"`
    Address Address `df:"address"`
}

data := map[string]any{
    "name": "John Doe",
    "address": map[string]any{
        "street":  "123 Main St",
        "city":    "New York",
        "country": "USA",
    },
}
```

## File Operations

df provides convenient functions for file-based binding:

### JSON Files

```go
type Config struct {
    Database DatabaseConfig `df:"database"`
    Server   ServerConfig   `df:"server"`
}

// Read from JSON file
var config Config
err := df.BindFromJSON(&config, "config.json")

// Write to JSON file
err = df.UnbindToJSON(&config, "output.json")
```

### YAML Files

```go
// Read from YAML file
err := df.BindFromYAML(&config, "config.yaml")

// Write to YAML file  
err = df.UnbindToYAML(&config, "output.yaml")
```

### Generic File Loading
Create instances directly from files:

```go
// Load and create in one step
config, err := df.NewFromJSON[Config]("config.json")
config, err := df.NewFromYAML[Config]("config.yaml")
```

## Configuration Layering

Use `df.Merge()` to build sophisticated configuration hierarchies:

```go
type ServerConfig struct {
    Host    string `df:"host"`
    Port    int    `df:"port"`
    Debug   bool   `df:"debug"`
    Timeout int    `df:"timeout"`
}

// Layer 1: Sensible defaults
config := &ServerConfig{
    Host:    "localhost",
    Port:    8080,
    Debug:   false,
    Timeout: 30,
}

// Layer 2: Configuration file
if fileExists("app.yaml") {
    err := df.MergeFromYAML(config, "app.yaml")
    if err != nil {
        return err
    }
}

// Layer 3: Environment variables  
envOverrides := map[string]any{
    "host": os.Getenv("HOST"),
    "port": os.Getenv("PORT"),
}
err := df.Merge(config, envOverrides)

// Layer 4: Command line flags
if *debugFlag {
    cliOverrides := map[string]any{"debug": true}
    err = df.Merge(config, cliOverrides)
}
```

## Error Handling

df provides detailed error information for binding failures:

```go
type User struct {
    Name  string `df:"+required"`
    Email string `df:"email,+required"`
    Age   int    `df:"age"`
}

data := map[string]any{
    "name": "John",
    // Missing required "email" field
    "age": "invalid", // Invalid type for int
}

user, err := df.New[User](data)
if err != nil {
    // Error will describe missing required field and type conversion failure
    fmt.Printf("binding error: %v\n", err)
}
```

## Best Practices

### Field Naming
- Use snake_case in data, CamelCase in structs
- Leverage automatic conversion when possible
- Use explicit tags only when needed

```go
// Good - relies on automatic conversion
type Config struct {
    DatabaseURL string // maps to "database_url"
    LogLevel    string // maps to "log_level"
}

// Only when you need custom mapping
type Config struct {
    DatabaseURL string `df:"db_connection_string"`
}
```

### Required vs Optional Fields
- Mark critical configuration as required
- Use pointers for truly optional fields
- Provide sensible defaults where appropriate

```go
type Config struct {
    APIKey   string  `df:"api_key,+required"`     // Must be provided
    LogLevel string  `df:"log_level"`            // Has default
    Debug    *bool   `df:"debug"`                // Truly optional
}
```

### Secret Management
- Always mark sensitive fields as secret
- Never log or display secret values
- Consider separate credential structures

```go
type DatabaseConfig struct {
    Host     string `df:"host"`
    Port     int    `df:"port"`
    Database string `df:"database"`
    Username string `df:"username"`
    Password string `df:"password,+secret"`
}
```

### Configuration Validation
- Validate after binding
- Use custom validation logic
- Fail fast with clear error messages

```go
config, err := df.New[Config](data)
if err != nil {
    return fmt.Errorf("config binding failed: %w", err)
}

if config.Port < 1 || config.Port > 65535 {
    return fmt.Errorf("invalid port: %d", config.Port)
}
```

## Next Steps

Now that you understand data binding, learn about:

- **[Dependency Injection](/guides/dependency-injection/)** - Use the container for object management
- **[Application Lifecycle](/guides/application-lifecycle/)** - Orchestrate complex applications with factories
- **[Advanced Features](/guides/advanced-features/)** - Explore Dynamic types and object references