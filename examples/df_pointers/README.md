# Pointer References Example

This example demonstrates how to use `df.Pointer[T]` for object references that support cycles and type safety.

## Key Concepts

- **Identifiable Interface**: Objects implement `GetId() string` to participate in pointer references
- **Pointer[T] Type**: Generic type that holds references like `*df.Pointer[*User]`  
- **Two-Phase Process**: 
  1. `df.Bind()` - loads data and stores `$ref` strings
  2. `df.Link()` - resolves all references to actual objects
- **Type Namespacing**: Objects with same ID but different types don't clash (e.g., `User:1` vs `Document:1`)

## Usage

```bash
go run main.go
```

## JSON Structure

The example uses this data structure with `$ref` fields:

```json
{
  "users": [
    {"id": "user1", "name": "Alice Johnson", "age": 28},
    {"id": "user2", "name": "Bob Smith", "age": 35}
  ],
  "documents": [
    {
      "id": "doc1", 
      "title": "Go Programming Guide",
      "author": {"$ref": "user1"},
      "editor": {"$ref": "user2"}
    }
  ]
}
```

The `$ref` values are resolved to actual object pointers during the Link phase.