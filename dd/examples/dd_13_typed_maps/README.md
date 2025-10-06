# typed maps example

this example demonstrates df.dd's support for typed maps (`map[K]V`) with various key and value types, enabling type-safe access to map-based data structures.

## key features

- **multiple key types**: int, uint, string, bool, and their variants (int64, uint32, etc.)
- **flexible value types**: primitives, structs, pointers, slices, and nested maps
- **automatic key coercion**: JSON/YAML string keys are automatically converted to target key types
- **bidirectional conversion**: bind and unbind operations preserve data integrity
- **nested structures**: support for maps of maps, maps of slices, and complex combinations

## usage

```bash
go run main.go
```

## typed map patterns

### 1. maps with int keys

ideal for id-based lookups and indexed data:

```go
type ServerConfig struct {
    Servers map[int]ServerConfig `dd:"servers"`
}

// JSON: {"servers": {"1": {...}, "2": {...}}}
// keys "1", "2" → int 1, 2
```

### 2. maps with string keys

standard configuration and named collections:

```go
type Config struct {
    Cache map[string]CachePolicy `dd:"cache"`
}

// JSON: {"cache": {"users": {...}, "sessions": {...}}}
```

### 3. maps with pointer values

useful for nullable or shared references:

```go
type UserRegistry struct {
    Users map[int]*User `dd:"users"`
}

// pointer values allow nil entries and shared references
```

### 4. maps with slice values

group-based or collection-based data:

```go
type Groups struct {
    Members map[string][]string `dd:"members"`
}

// JSON: {"members": {"admins": ["alice", "bob"]}}
```

### 5. nested maps

hierarchical configuration and multi-level data:

```go
type EnvConfig struct {
    Envs map[string]map[string]string `dd:"envs"`
}

// JSON: {"envs": {"prod": {"db_host": "...", "api_url": "..."}}}
```

## key type coercion

since JSON and YAML always use string keys, df.dd automatically coerces them to the target key type:

| target type | JSON key | go key | example |
|-------------|----------|--------|---------|
| `map[int]T` | `"42"` | `42` | server IDs |
| `map[int64]T` | `"1001"` | `1001` | user IDs |
| `map[uint]T` | `"10"` | `10` | indices |
| `map[bool]T` | `"true"` | `true` | flags |
| `map[string]T` | `"key"` | `"key"` | no conversion |

## when to use typed maps vs slices

**use typed maps when:**
- you need direct key-based lookup (e.g., `servers[1]`)
- keys have semantic meaning (server IDs, user IDs)
- order doesn't matter
- you want to prevent duplicate keys

**use slices when:**
- order matters
- you need indexed access (e.g., `items[0]`)
- you need to iterate in insertion order
- keys are sequential integers starting from 0

## common use cases

1. **configuration management**: environment-specific settings (`map[string]EnvConfig`)
2. **user registries**: user ID to user mapping (`map[int]*User`)
3. **cache policies**: cache name to policy mapping (`map[string]CachePolicy`)
4. **feature flags**: boolean flags to descriptions (`map[bool]string`)
5. **database records**: record ID to record mapping (`map[uint64]Record`)
6. **routing tables**: route key to handler mapping (`map[string]Handler`)

## roundtrip behavior

when binding and unbinding:

1. **bind**: string keys from JSON/YAML → typed keys in go
2. **unbind**: typed keys in go → string keys in JSON/YAML
3. **roundtrip**: data integrity preserved through coercion

```go
// original: map[int]string{1: "one", 2: "two"}
// unbind:   {"1": "one", "2": "two"}
// rebind:   map[int]string{1: "one", 2: "two"}
```

## limitations

- map keys must be comparable types (no slices or maps as keys)
- key coercion from strings must be unambiguous
- float keys may have precision issues when used as strings
- complex/custom key types are not supported

## see also

- `dd_05_nested_structs`: nested struct handling
- `dd_03_type_coercion`: type conversion details
- `dd_04_file_io`: loading/saving with JSON/YAML
