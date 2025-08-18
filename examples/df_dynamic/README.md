# dynamic types example

this example demonstrates polymorphic data binding using the `df.Dynamic` interface for runtime type discrimination and flexible data structures.

## key concepts demonstrated

### **polymorphic data binding**
- **df.Dynamic interface**: objects that can represent different types at runtime
- **type discrimination**: use type fields to determine concrete types during binding
- **flexible data structures**: single fields that can hold different struct types
- **runtime type resolution**: bind data to appropriate concrete types dynamically

### **type binder registration**
- **global binders**: register type mappings that apply to all dynamic fields
- **field-specific binders**: override global mappings for specific struct fields
- **type string mapping**: map discriminator strings to concrete Go types
- **extensible type system**: easily add new types without modifying existing code

### **real-world use cases**
- **configuration systems**: different config sections with different schemas
- **plugin architectures**: load different plugin types based on data
- **API responses**: handle different response types from external services
- **workflow engines**: process different action types in workflows

## dynamic interface

```go
type Dynamic interface {
    Type() string              // returns the discriminator string
    ToMap() map[string]any    // converts the struct to a map for unbinding
}
```

## workflow demonstrated

1. **type definition**: create concrete types implementing df.Dynamic
2. **binder registration**: map type strings to concrete Go types
3. **data binding**: bind polymorphic data using registered type mappings
4. **type assertions**: safely access concrete type functionality
5. **unbinding**: convert dynamic objects back to maps with type information

## example structure

the example demonstrates a notification system with different action types:
- **EmailAction**: send email notifications with customizable HTML format
- **SlackAction**: send slack messages with urgency flags
- **WebhookAction**: trigger HTTP webhooks with custom payloads

each action type has different fields and behavior but can be stored in the same `Action df.Dynamic` field.

## usage

```bash
go run main.go
```

## benefits

- **type safety**: compile-time safety with runtime flexibility
- **extensibility**: add new types without modifying existing code
- **clean data binding**: automatic type resolution during df.Bind()
- **serialization support**: round-trip compatibility with ToMap()
- **plugin-friendly**: ideal for extensible application architectures