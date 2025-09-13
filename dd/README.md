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

- **🔄 Bidirectional Binding**: Seamlessly convert structs ↔ maps
- **🏷️ Struct Tags**: Control field mapping with `df` tags
- **⚡ Type Coercion**: Automatic type conversion (strings→numbers, etc.)
- **📁 File I/O**: Direct JSON/YAML binding with `BindFromJSON()`, `UnbindToYAML()`
- **🔗 Object References**: `Pointer[T]` type with cycle-safe linking
- **🎭 Dynamic Types**: Runtime type discrimination via `Dynamic` interface
- **✅ Validation**: Required fields and custom validation rules

## Core Functions

- **`dd.New[T](data)`** - Type-safe struct creation from map
- **`dd.Bind(target, data)`** - Bind data to existing struct
- **`dd.Unbind(struct)`** - Convert struct to map
- **`dd.Merge(base, override)`** - Deep merge two data maps

## Common Patterns

**Struct Tags for Control**
```go
type User struct {
    Name  string `dd:"+required"`           // required field
    Email string `dd:"email_address"`       // custom field name
    Token string `dd:"-"`                   // excluded from binding
    Age   int    `dd:",default=18"`         // default value
}
```

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

## Examples

See [examples/](examples/) for progressive tutorials from basic binding to advanced object references and dynamic types.

---
*Part of the [df framework](../README.md) - dynamic foundation for Go applications*