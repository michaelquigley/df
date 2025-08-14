# df_defaults Example

This example demonstrates the `df.BindTo()` function, which allows binding partial data to existing structs while preserving default values.

## Key Features

- **Preserves existing values**: Fields not present in the source data retain their original values
- **Partial updates**: Only provided fields are updated
- **Nested struct support**: Works with nested structures
- **Same field mapping**: Uses identical struct tag and naming conventions as `df.Bind()`

## Use Case

Perfect for configuration scenarios where you have:
- Default configuration values
- Partial overrides from files, environment variables, or user input
- Need to merge multiple configuration sources

## Running

```bash
go run main.go
```

The example shows how `BindTo` preserves default values while updating only the fields present in the input data.