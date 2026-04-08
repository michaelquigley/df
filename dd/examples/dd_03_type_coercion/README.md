# configuration defaults with merge

this example demonstrates how to build robust configuration systems using `dd.Merge()`. unlike `dd.Bind()` which overwrites the entire struct, `Merge()` intelligently overlays external data onto pre-initialized structs with sensible defaults.

## key concepts demonstrated

### **defaults systems**
- **pre-initialized structs**: start with sensible default values  
- **selective overrides**: external config only specifies what should change
- **preserved defaults**: unspecified fields keep their original values
- **layered configuration**: multiple sources can be merged progressively
- **optional nested defaults**: `ApplyDefaults()` can initialize optional pointer fields when `Merge()` allocates them

### **configuration hierarchies**
- **layer 1**: application defaults (compiled into code)
- **layer 2**: environment-specific config (dev/staging/prod)
- **layer 3**: user overrides (CLI flags, user preferences)
- **final result**: intelligent merge of all layers

### **real-world patterns**
- **12-factor app compliance**: environment-based configuration
- **backward compatibility**: new fields with defaults don't break existing configs
- **progressive enhancement**: users can adopt new features gradually
- **ops-friendly**: minimal config files, maximum flexibility

## workflow demonstrated

1. **initialize with defaults**: create structs with sensible default values
2. **apply partial config**: use `dd.Merge()` to overlay external configuration  
3. **verify preservation**: show which values were overridden vs preserved
4. **instantiate optional config**: show a nil optional pointer becoming populated with merge-time defaults

## example structure

```go
// Application defaults (compiled-in)
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
    TLS: nil, // remain optional until external data requests it
}

// Partial override (from config file/env/CLI)
partialData := map[string]any{
    "server": map[string]any{
        "host":  "api.example.com",
        "debug": true,
        // port and timeout not specified - will be preserved
    },
    "tls": map[string]any{
        "server_name": "api.example.com",
        // min_version comes from ApplyDefaults on TLSConfig
    },
}

// intelligent merge
dd.Merge(config, partialData)
```

## optional nested structs

the example adds a `TLS *TLSConfig` field that starts as `nil`. `TLSConfig` implements:

```go
func (c *TLSConfig) ApplyDefaults() {
    c.ServerName = "localhost"
    c.MinVersion = "1.3"
}
```

when merge input contains a `tls` object, `dd.Merge()` allocates `TLSConfig`, applies defaults, and then overlays external values. this keeps `TLS` optional when omitted, while still allowing internal defaults when it is present.

## running

```bash
go run main.go
```

the output clearly shows which values were updated from external config and which defaults were preserved, demonstrating the power of selective configuration merging.
