# dd - Dynamic Data

**Convert between Go structs and maps with ease**

The `dd` package provides bidirectional data binding between Go structs and `map[string]any`, enabling dynamic data handling for configuration, persistence, and API marshaling. Since maps are a foundational data structure, this facilitates seamless integration with any network protocol, object store, database, or file format that works with key-value data.

## Quick Start

```go
import "github.com/michaelquigley/df/dd"

// struct → map
user := User{Name: "John", Age: 30}
data, _ := dd.Unbind(user)
// data: map[string]any{"name": "John", "age": 30}

// map → struct  
userData := map[string]any{"name": "Alice", "age": 25}
user, _ := dd.New[User](userData)
// user: User{Name: "Alice", Age: 25}
```

## Key Features

- **Bidirectional Binding**: Seamlessly convert structs ↔ maps
- **Struct Tags**: Control field mapping with `dd` tags
- **Type Coercion**: Automatic type conversion (strings→numbers, etc.)
- **Typed Maps**: Full support for `map[K]V` with any comparable key type
- **File I/O**: Direct JSON/YAML binding with `BindFromJSON()`, `UnbindToYAML()`
- **Object References**: `Pointer[T]` type with cycle-safe linking
- **Dynamic Types**: Runtime type discrimination via `Dynamic` interface
- **Merge-Time Defaults**: Optional nested structs can provide defaults when `Merge()` allocates them
- **Validation**: Required fields and custom validation rules

## Core Functions

- **`dd.New[T](data)`** - Type-safe struct creation from map
- **`dd.Bind(target, data)`** - Bind data to existing struct
- **`dd.Unbind(struct)`** - Convert struct to map
- **`dd.Merge(target, data)`** - Overlay partial data onto an existing struct while preserving existing values

## Common Patterns

**Struct Tags for Control**
```go
type User struct {
    Name  string `dd:"+required"`           // required field
    Email string `dd:"email_address"`       // custom field name
    Token string `dd:"-"`                   // excluded from binding
    Age   int    `dd:",+omitempty"`         // omitted during Unbind when zero
}
```

**Merge-Time Defaults For Optional Nested Structs**
```go
type TLSConfig struct {
    ServerName string
    MinVersion string
}

func (c *TLSConfig) ApplyDefaults() {
    c.ServerName = "localhost"
    c.MinVersion = "1.3"
}

type Config struct {
    TLS *TLSConfig
}

cfg := &Config{}
dd.Merge(cfg, map[string]any{
    "tls": map[string]any{
        "server_name": "api.example.com",
    },
})

// cfg.TLS.ServerName == "api.example.com"
// cfg.TLS.MinVersion == "1.3"
```

`ApplyDefaults()` semantics:
- `Merge()` only
- runs only when `Merge()` allocates a fresh struct instance
- not called by `Bind()` or `New()`
- not called when `Merge()` is updating an existing non-nil pointer
- incoming data is bound after defaults are applied, so explicit values still win

**File Persistence**
```go
// Load config from JSON
config, _ := dd.BindFromJSON[AppConfig]("config.json")

// Save to YAML
dd.UnbindToYAML(config, "config.yaml")
```

**Dynamic Types**
```go
// Handle different object types at runtime
data := map[string]any{
    "type": "user",
    "name": "John",
}
obj, _ := dd.New[dd.Dynamic](data)  // Creates appropriate type
```

**Typed Maps**
```go
// Maps with typed keys and values
type ServerConfig struct {
    Servers map[int]Server  // int keys from JSON strings
    Cache   map[string]CacheConfig
}

// JSON: {"servers": {"1": {...}, "2": {...}}}
config, _ := dd.New[ServerConfig](data)
server := config.Servers[1]  // Direct typed access
```

## Examples

See [examples/](examples/) for progressive tutorials from basic binding to advanced object references and dynamic types.

---
*Part of the [df framework](../README.md) - dynamic foundation for Go applications*
