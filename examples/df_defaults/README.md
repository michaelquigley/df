# configuration defaults with merge

this example demonstrates how to build robust configuration systems using `df.Merge()`. unlike `df.Bind()` which overwrites the entire struct, `Merge()` intelligently overlays external data onto pre-initialized structs with sensible defaults.

## key concepts demonstrated

### **defaults systems**
- **pre-initialized structs**: start with sensible default values  
- **selective overrides**: external config only specifies what should change
- **preserved defaults**: unspecified fields keep their original values
- **layered configuration**: multiple sources can be merged progressively

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
2. **apply partial config**: use `df.Merge()` to overlay external configuration  
3. **verify preservation**: show which values were overridden vs preserved
4. **configuration layering**: demonstrate multiple merge operations

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
}

// Partial override (from config file/env/CLI)
partialData := map[string]any{
    "server": map[string]any{
        "host":  "api.example.com",
        "debug": true,
        // port and timeout not specified - will be preserved
    },
}

// intelligent merge
df.Merge(config, partialData)
```

## running

```bash
go run main.go
```

the output clearly shows which values were updated from external config and which defaults were preserved, demonstrating the power of selective configuration merging.