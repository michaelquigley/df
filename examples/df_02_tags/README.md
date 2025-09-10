# struct tag features example

this example demonstrates all supported struct tag features for customizing field behavior during data binding and unbinding operations.

## key concepts demonstrated

### **field naming control**
- **default behavior**: Go field names are converted to snake_case for binding
- **custom names**: override default field names using `df:"custom_name"`
- **name inheritance**: nested structs respect parent naming conventions

### **field validation**
- **required fields**: mark fields as mandatory using `df:",+required"`
- **validation timing**: required field validation occurs during binding
- **nested validation**: required flags apply to all nesting levels

### **security and privacy**
- **secret fields**: hide sensitive data using `df:",+secret"` 
- **inspect filtering**: secret fields are hidden by default in df.Inspect()
- **unbind behavior**: secret fields are included in unbinding (use with care)

### **field exclusion**
- **complete exclusion**: exclude fields entirely using `df:"-"`
- **binding exclusion**: excluded fields are not bound from data
- **unbinding exclusion**: excluded fields are not included in unbind output

### **tag combination**
- **multiple flags**: combine flags like `df:"name,+required,+secret"`
- **flag precedence**: required validation always applies regardless of other flags
- **inheritance**: nested struct fields inherit tag behavior appropriately

## workflow demonstrated

1. **basic field naming**: demonstrate custom field names vs default snake_case
2. **required field validation**: show validation errors for missing required fields
3. **secret field handling**: demonstrate secret field behavior with df.Inspect()
4. **field exclusion**: show complete exclusion with `df:"-"`
5. **complex combinations**: demonstrate all tag features together
6. **nested structures**: show tag behavior in hierarchical data

## tag syntax reference

```go
type Example struct {
    Field1 string `df:"custom_name"`              // custom field name
    Field2 string `df:",+required"`                // required field with default name
    Field3 string `df:",+secret"`                  // secret field (hidden in inspect)
    Field4 string `df:"custom,+required"`          // custom name + required
    Field5 string `df:"custom,+secret"`            // custom name + secret  
    Field6 string `df:"custom,+required,+secret"`   // custom name + required + secret
    Field7 string `df:",+required,+secret"`         // default name + required + secret
    Field8 string `df:"-"`                        // exclude from binding/unbinding
}
```

## example scenarios

the example demonstrates:
- **api configuration**: service settings with required fields and secrets
- **user profiles**: personal data with privacy controls
- **system settings**: hierarchical configuration with validation
- **feature flags**: boolean settings with custom naming
- **nested structures**: complex data with inherited tag behavior

## usage

```bash
go run main.go
```

## benefits

- **customization**: fine-grained control over field binding behavior
- **validation**: built-in required field validation during binding
- **security**: automatic handling of sensitive data in inspection
- **flexibility**: combine multiple tag features for complex requirements
- **maintainability**: clear, declarative field configuration