# basic bind/unbind example

this example demonstrates both the modern `df.New[T]()` and traditional `df.Bind()` functions along with `df.Unbind()` for converting between structured data and go structs.

## key concepts

- **New[T] API**: type-safe allocation using go generics
- **Bind API**: manual allocation control for advanced use cases
- **bidirectional data mapping**: convert from `map[string]any` to structs and back
- **struct tags**: use `df` tags for custom field mapping and validation
- **nested structures**: handle complex nested data with pointers to structs
- **round-trip compatibility**: data maintains integrity through bind/unbind cycles
- **error handling**: validate required fields and handle binding errors gracefully

## usage

```bash
go run main.go
```

## data structure

the example works with this user profile data:

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

## struct definitions

```go
type User struct {
    Name    string `df:"required"`       // required field
    Email   string                       // default field mapping
    Age     int                          // type conversion
    Active  bool                         // boolean handling
    Profile *Profile                     // nested struct (snake_case: "profile")
}

type Profile struct {
    Bio     string
    Website string
}
```

## workflow demonstrated

1. **allocate and bind**: use `df.New[T]()` for type-safe allocation and binding
2. **unbinding**: convert go structs back to `map[string]any`
3. **round-trip**: verify data integrity through the complete cycle
4. **error handling**: show validation behavior for missing required fields
5. **manual binding**: show `df.Bind()` for cases requiring manual allocation control

this example showcases both struct binding patterns, providing the foundation for data persistence, API marshaling, and configuration management patterns.