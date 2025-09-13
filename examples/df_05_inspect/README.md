# inspect example

this example demonstrates the `df.Inspect()` function, which provides human-readable output of bound configuration structures. the `Inspect` function is designed for debugging and validating configuration state.

## key features

- **secret filtering**: fields marked with `dd:",+secret"` are hidden by default
- **human-readable output**: clean, indented pseudo-data structure format  
- **configurable options**: custom indentation, depth limits, and secret visibility
- **type-aware display**: special handling for durations, pointers, slices, and nested structs

## usage

```bash
go run main.go
```

this will show three different inspection outputs:

1. **default**: secrets hidden, standard formatting
2. **with secrets**: all fields visible including sensitive data
3. **custom format**: custom indentation and options

## secret fields

fields can be marked as secret using the `+secret` flag in the `df` struct tag:

```go
type Config struct {
    PublicField string
    SecretField string `dd:"secret_field,+secret"`
}
```

secret fields are automatically hidden in the default `Inspect` output unless explicitly enabled with `ShowSecrets: true`.

## options

the `InspectOptions` struct provides configuration for output formatting:

- `MaxDepth`: limits recursion depth (default: 10)
- `Indent`: sets indentation string (default: "  ")  
- `ShowSecrets`: includes secret fields when true (default: false)