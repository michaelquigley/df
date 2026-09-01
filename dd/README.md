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
- **File I/O**: Direct JSON/YAML file binding with `NewJSONFile[T]()`, `UnbindYAMLFile()`
- **Object References**: `Pointer[T]` type with cycle-safe linking
- **Dynamic Types**: Runtime type discrimination via `Dynamic` interface
- **Merge-Time Defaults**: Optional nested structs can provide defaults when `Merge()` allocates them
- **Validation**: Required fields and custom validation rules
- **Nullable Fields**: Per-field `+nullable` handling for explicit nulls
- **Strict Mode**: Opt-in exact acceptance for contract data — duplicate-key/unknown-field rejection, zero coercion, `+opaque` subtrees
- **Deterministic Output**: `UnbindJSON`/`UnbindYAML` produce byte-stable, sorted-key output
- **JSON Lines**: `UnbindJSONL`/`UnbindJSONLWriter` emit one compact, newline-terminated record per call for JSONL streams

## Core Functions

- **`dd.New[T](data)`** - Type-safe struct creation from map
- **`dd.Bind(target, data)`** - Bind data to existing struct
- **`dd.Unbind(struct)`** - Convert struct to map
- **`dd.Merge(target, data)`** - Overlay partial data onto an existing struct while preserving existing values

## Deterministic Output

`UnbindJSON`, `UnbindJSONL`, `UnbindYAML`, and their writer/file variants produce **deterministic** output: for a given input value, the serialized bytes are identical across runs, processes, and versions. All keys — struct field names and map keys alike — are emitted in **sorted order**, and slice/array element order is preserved as-is. Fields captured via `+extra` are interleaved in sorted order with the rest, not appended at the end. This makes `dd` output safe to commit to version control and diff without spurious churn.

The guarantee applies to the serialized forms. The raw `Unbind()` return is a Go `map[string]any` and is therefore unordered; ordering is realized only at serialization.

Two distinct keys of one map never collapse into one serialized key. Map keys are stringified on the way out, and a map whose keys share a spelling — an interface-keyed `map[any]string{1: "int", "1": "str"}` — has no lossless JSON or YAML form, so `Unbind` refuses it with a `KeyCollisionError` rather than letting map iteration order decide which entry survives. Typed maps (`map[int]V`, `map[string]V`) cannot collide in ordinary use.

```go
data, _ := dd.UnbindJSON(cfg)
// identical bytes every time, keys sorted — clean git diffs
```

## Strict Acceptance Mode

`dd`'s default posture is forgiving: unknown keys are ignored, duplicate JSON keys resolve last-wins inside the parser, and values coerce across types (`"5"` becomes an int, `5` becomes a string). Forgiving YAML retains `yaml.v3`'s existing duplicate-key rejection. That is right for config files and local records. **Strict mode** is the opposite posture, for data whose exact spelling is the contract — signed payloads, hash-pinned documents, normative wire formats:

```go
err := dd.BindJSON(&doc, data, dd.Strict())
```

With `dd.Strict()`:

- **Intake** (`BindJSON`, `BindYAML`, and their reader/file variants) rejects duplicate keys anywhere, trailing data after the document, YAML aliases and anchors, and YAML scalars outside the JSON value model (quote timestamps to bind them as strings). Numbers are preserved as `json.Number` — an authored YAML `5.00` reaches binding as `"5.00"`, never a float.
- **Binding** rejects input keys the target struct does not declare, and refuses type coercion entirely: a number arriving at a string field is an error, not a conversion. Integer fields require integer lexemes (no fractions, no exponents) and overflow is refused; native signed and unsigned Go values do not cross-bind. Map keys must have string as their underlying type; forgiving mode retains conversion into numeric and boolean map keys. `time.Time` accepts only its defined RFC3339 encoding; `time.Duration` only its duration string.
- A field tagged `+opaque` (a `map[string]any`) accepts any members and captures its raw subtree uninterpreted — syntactic intake rules still apply inside it, binding rules do not. A field tagged `+extra` still captures unknown keys by declared intent. A field tagged `+nullable` treats an explicit null as absent in both strict and forgiving mode; when combined with `+required`, null is rejected as required-missing.

The strict decoders are public for pipelines that need to inspect or normalize the tree between intake and binding:

```go
tree, err := dd.DecodeStrictYAML(data)   // duplicate-key-checked, lexeme-preserving
// ... normalize ...
err = dd.Bind(&doc, tree, dd.Strict())
```

Custom `Converters`, `Dynamic` binders, and `UnmarshalDd` implementations remain in effect under strict mode. Strict intake validates syntax before delegation, but the custom machinery owns its field and type acceptance. In particular, `dd.Strict()` does not make an `Unmarshaler` strict; do not use a permissive or legacy unmarshaler at an exact contract boundary unless it independently validates every accepted key and value. `Merge` does not support strict mode (partial overlay is the opposite posture by design). The forgiving default is unchanged and pinned by tests.

## Common Patterns

**Struct Tags for Control**
```go
type User struct {
    Name  string `dd:"+required"`           // required field
    Email string `dd:"email_address"`       // custom field name
    Token string `dd:"-"`                   // excluded from binding
    Age   int    `dd:",+omitempty"`         // omitted during Unbind when zero
    Bio   *string `dd:",+nullable"`          // explicit null binds as absent
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
// load config from JSON
config, _ := dd.NewJSONFile[AppConfig]("config.json")

// save to YAML
dd.UnbindYAMLFile(config, "config.yaml")
```

**JSON Lines**
```go
// one compact, newline-terminated record per call; append to any io.Writer
for _, ev := range events {
    dd.UnbindJSONLWriter(ev, w)
}

// reading back needs nothing new: one BindJSON per line
scanner := bufio.NewScanner(r)
for scanner.Scan() {
    var ev Event
    dd.BindJSON(&ev, scanner.Bytes())
}
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
