# Configuration Defaults with Merge

This example demonstrates how to build robust configuration systems using `df.Merge()`. Unlike `df.Bind()` which overwrites the entire struct, `Merge()` intelligently overlays external data onto pre-initialized structs with sensible defaults.

## Key Concepts Demonstrated

### **Defaults Systems**
- **Pre-initialized structs**: Start with sensible default values  
- **Selective overrides**: External config only specifies what should change
- **Preserved defaults**: Unspecified fields keep their original values
- **Layered configuration**: Multiple sources can be merged progressively

### **Configuration Hierarchies**
- **Layer 1**: Application defaults (compiled into code)
- **Layer 2**: Environment-specific config (dev/staging/prod)
- **Layer 3**: User overrides (CLI flags, user preferences)
- **Final result**: Intelligent merge of all layers

### **Real-World Patterns**
- **12-Factor App compliance**: Environment-based configuration
- **Backward compatibility**: New fields with defaults don't break existing configs
- **Progressive enhancement**: Users can adopt new features gradually
- **Ops-friendly**: Minimal config files, maximum flexibility

## Workflow Demonstrated

1. **Initialize with defaults**: Create structs with sensible default values
2. **Apply partial config**: Use `df.Merge()` to overlay external configuration  
3. **Verify preservation**: Show which values were overridden vs preserved
4. **Configuration layering**: Demonstrate multiple merge operations

## Example Structure

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

// Intelligent merge
df.Merge(config, partialData)
```

## Running

```bash
go run main.go
```

The output clearly shows which values were updated from external config and which defaults were preserved, demonstrating the power of selective configuration merging.