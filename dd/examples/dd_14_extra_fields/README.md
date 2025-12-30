# dd_14_extra_fields - capturing unknown data fields

this example demonstrates the `+extra` tag for capturing unmatched data keys during binding and merging them back during unbinding. this enables forward compatibility and extension data patterns.

## key concepts demonstrated

### **extra field capture**
- **bind behavior**: unknown keys in input data are collected into a `map[string]any` field
- **unbind behavior**: extra field contents are merged back into the output map
- **type requirement**: the extra field must be of type `map[string]any`
- **single field**: only one `+extra` field is allowed per struct

### **nested behavior**
- **independent capture**: each nested struct captures its own extras independently
- **level isolation**: extras at parent level stay at parent, extras at child level stay at child
- **recursive support**: works with pointer structs, slices of structs, and map values

### **embedded struct behavior**
- **shared namespace**: embedded struct fields share the parent's key namespace
- **parent captures**: the parent's extra field captures keys not matched by any field (embedded or direct)

### **merge behavior**
- **additive merge**: when using `dd.Merge()`, new extras are added to existing extra map
- **preservation**: existing extra keys are preserved unless explicitly overwritten

## workflow demonstrated

1. **basic extra field capture**: bind data with unknown keys into a struct
2. **unbind with extras**: convert struct back to map with extras merged
3. **nested structs**: each level captures its own extras independently
4. **round-trip**: verify data preservation through bind/unbind cycle
5. **merge behavior**: demonstrate additive merging of extras
6. **embedded structs**: show how embedded fields share parent namespace
7. **slice of structs**: each element captures its own extras
8. **empty extras**: extra field remains nil when no unknown keys exist

## tag syntax

```go
type Config struct {
    // known fields
    Name    string `dd:"name"`
    Version string `dd:"version"`

    // extra field captures all unknown keys
    Extra map[string]any `dd:",+extra"`
}
```

## example scenarios

the example demonstrates:
- **application config**: capturing extension metadata like author, license
- **service config**: nested settings with independent extras at each level
- **document storage**: embedded base fields with document-level extras
- **item collections**: slice elements each capturing their own extras

## use cases

- **forward compatibility**: preserve fields from newer versions of data
- **extension data**: allow user-defined custom fields
- **configuration passthrough**: forward extra config to subsystems
- **api integration**: preserve unknown fields when round-tripping external data
- **schema evolution**: gracefully handle schema changes without data loss

## usage

```bash
go run main.go
```

## benefits

- **data preservation**: never lose unknown fields during binding
- **round-trip safety**: bind then unbind produces equivalent output
- **flexibility**: allow arbitrary extension data without schema changes
- **isolation**: nested structs manage their own extras independently
- **simplicity**: single tag enables the feature with no additional code
