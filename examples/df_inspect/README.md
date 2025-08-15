# Inspect Example

This example demonstrates the `df.Inspect()` function, which provides human-readable output of bound configuration structures. The `Inspect` function is designed for debugging and validating configuration state.

## Key Features

- **Secret Filtering**: Fields marked with `df:",secret"` are hidden by default
- **Human-Readable Output**: Clean, indented pseudo-data structure format  
- **Configurable Options**: Custom indentation, depth limits, and secret visibility
- **Type-Aware Display**: Special handling for durations, pointers, slices, and nested structs

## Usage

```bash
go run main.go
```

This will show three different inspection outputs:

1. **Default**: Secrets hidden, standard formatting
2. **With Secrets**: All fields visible including sensitive data
3. **Custom Format**: Custom indentation and options

## Secret Fields

Fields can be marked as secret using the `secret` flag in the `df` struct tag:

```go
type Config struct {
    PublicField string `df:"public_field"`
    SecretField string `df:"secret_field,secret"`
}
```

Secret fields are automatically hidden in the default `Inspect` output unless explicitly enabled with `ShowSecrets: true`.

## Options

The `InspectOptions` struct provides configuration for output formatting:

- `MaxDepth`: Limits recursion depth (default: 10)
- `Indent`: Sets indentation string (default: "  ")  
- `ShowSecrets`: Includes secret fields when true (default: false)