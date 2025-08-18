# field-specific dynamic binders example

this example demonstrates `FieldDynamicBinders` for per-field polymorphic type control, allowing different struct fields to use different sets of dynamic type mappings.

## key concepts demonstrated

### **field-specific polymorphic binding**
- **FieldDynamicBinders**: different type registries for different struct fields
- **field path targeting**: precise control over which fields use which binders
- **override behavior**: field-specific binders override global binders
- **path matching**: automatic handling of array indices in field paths

### **advanced dynamic typing**
- **heterogeneous polymorphism**: different fields support different type sets
- **context-aware binding**: type resolution based on field location
- **namespace isolation**: prevent type conflicts between different domains
- **extensible architecture**: easy addition of new types per field

### **real-world scenarios**
- **workflow systems**: steps and actions have different type sets
- **plugin architectures**: different extension points support different plugins
- **configuration systems**: different sections support different schema types
- **content management**: different content areas support different widget types

## workflow demonstrated

1. **field path definition**: understand how field paths are constructed
2. **binder registration**: register different type sets for different fields
3. **polymorphic binding**: bind data with field-specific type resolution
4. **type isolation**: demonstrate how types are isolated per field
5. **fallback behavior**: show global binder fallback when field binders aren't found

## example structure

the example demonstrates a workflow system with:
- **workflow steps**: process, decision, notification steps
- **step actions**: different action types per step type
- **trigger conditions**: different condition types for triggers
- **data transformations**: different transformer types for data processing

## usage

```bash
go run main.go
```

## benefits

- **type safety**: compile-time safety with field-specific runtime flexibility
- **namespace isolation**: prevent type name conflicts between different domains
- **extensible design**: easily add new types to specific fields without affecting others
- **clean architecture**: separate concerns by field rather than globally
- **precise control**: fine-grained control over polymorphic behavior