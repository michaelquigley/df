# custom converters example

this example demonstrates how to use custom field converters in the `df` library to handle specialized data type conversions.

## overview

the `df` library supports custom converters through the `Converter` interface, allowing you to define bidirectional type conversion for your custom types. this is particularly useful when:

- you have domain-specific types (like `email`, `phone_number`, etc.)
- you need to parse complex data formats (like timestamps, temperatures with units)
- you want to add validation during the binding process
- you need to handle multiple input formats for the same logical type

## converter interface

```go
type Converter interface {
    // FromRaw converts a raw value (from the data map) to the target type
    FromRaw(raw interface{}) (interface{}, error)
    
    // ToRaw converts a typed value back to a raw value for serialization
    ToRaw(value interface{}) (interface{}, error)
}
```

## example types

this example implements converters for three custom types:

### 1. Email type with validation
- converts strings to validated `Email` type
- validates email format using regex
- provides meaningful error messages for invalid emails

### 2. Temperature type with unit handling
- supports both string format (`"23.5C"`) and object format (`{"value": 23.5, "unit": "C"}`)
- handles celsius and fahrenheit units
- validates unit values

### 3. timestamp conversion for time.Time
- parses various string timestamp formats (rfc3339, custom formats)
- handles unix timestamps (int64, float64)
- converts back to rfc3339 string format

## usage

```go
// setup converters
opts := &df.Options{
    Converters: map[reflect.Type]df.Converter{
        reflect.TypeOf(Email("")):     &EmailConverter{},
        reflect.TypeOf(Temperature{}): &TemperatureConverter{},
        reflect.TypeOf(time.Time{}):   &TimestampConverter{},
    },
}

// bind with custom converters
var user User
err := df.Bind(&user, data, opts)

// unbind with custom converters
data, err := df.Unbind(user, opts)
```

## running the example

```bash
cd examples/df_converters
go run main.go
```

## key features demonstrated

1. **type validation**: the email converter validates email format during binding
2. **multiple input formats**: temperature converter handles both string and object formats
3. **bidirectional conversion**: all converters support both binding and unbinding
4. **error handling**: converters provide descriptive error messages
5. **integration**: converters work seamlessly with pointers, slices, and nested structs

## benefits

- **type safety**: ensures data conforms to expected formats before binding
- **flexibility**: supports multiple input formats for the same logical type
- **reusability**: converters can be used across different struct definitions
- **maintainability**: centralizes conversion logic in dedicated converter types
- **validation**: built-in validation during the binding process