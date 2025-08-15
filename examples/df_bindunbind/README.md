# Basic Bind/Unbind Example

This example demonstrates both the modern `df.New[T]()` and traditional `df.Bind()` functions along with `df.Unbind()` for converting between structured data and Go structs.

## Key Concepts

- **Modern New[T] API**: Type-safe allocation using Go generics
- **Traditional Bind API**: Manual allocation control for advanced use cases
- **Bidirectional Data Mapping**: Convert from `map[string]any` to structs and back
- **Struct Tags**: Use `df` tags for custom field mapping and validation
- **Nested Structures**: Handle complex nested data with pointers to structs
- **Round-trip Compatibility**: Data maintains integrity through bind/unbind cycles
- **Error Handling**: Validate required fields and handle binding errors gracefully

## Usage

```bash
go run main.go
```

## Data Structure

The example works with this user profile data:

```go
userData := map[string]any{
    "name":   "John Doe",
    "email":  "john@example.com", 
    "age":    30,
    "active": true,
    "profile": map[string]any{
        "bio":     "Software developer",
        "website": "https://johndoe.dev",
    },
}
```

## Struct Definitions

```go
type User struct {
    Name    string `df:"name,required"`  // Required field
    Email   string `df:"email"`          // Custom field mapping
    Age     int    `df:"age"`            // Type conversion
    Active  bool   `df:"active"`         // Boolean handling
    Profile *Profile                     // Nested struct (snake_case: "profile")
}

type Profile struct {
    Bio     string `df:"bio"`
    Website string `df:"website"`
}
```

## Workflow Demonstrated

1. **Modern Binding**: Use `df.New[T]()` for type-safe allocation and binding
2. **Unbinding**: Convert Go structs back to `map[string]any`
3. **Round-trip**: Verify data integrity through the complete cycle
4. **Error Handling**: Show validation behavior for missing required fields
5. **Traditional Binding**: Show `df.Bind()` for cases requiring manual allocation control

This example showcases both modern and traditional APIs, providing the foundation for data persistence, API marshaling, and configuration management patterns.