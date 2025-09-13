# custom marshaler/unmarshaler example

this example demonstrates the `Marshaler` and `Unmarshaler` interfaces for types that need complete control over their binding and unbinding process.

## key concepts demonstrated

### **custom marshaling/unmarshaling**
- **Marshaler interface**: types define their own unbinding logic via `MarshalDf()`
- **Unmarshaler interface**: types define their own binding logic via `UnmarshalDf()`
- **complete control**: full control over the entire binding/unbinding process
- **different from converters**: converters handle type conversion, marshalers handle the entire process

### **use cases**
- **legacy data formats**: handle data that doesn't map cleanly to struct fields
- **complex validation**: perform validation that requires access to the entire data map
- **custom serialization**: implement domain-specific serialization logic
- **computed fields**: fields that are calculated rather than stored directly

### **marshaler vs converter comparison**
- **converters**: handle type conversion between specific types (string → Email)
- **marshalers**: handle the entire binding/unbinding process for a type
- **converters**: work on individual field values
- **marshalers**: work on complete data maps for the type

## workflow demonstrated

1. **custom time handling**: demonstrate time parsing with multiple formats and validation
2. **complex validation**: show validation that requires multiple fields
3. **legacy format support**: handle data formats that don't match struct layout
4. **computed fields**: demonstrate fields calculated during marshaling/unmarshaling
5. **round-trip compatibility**: ensure data integrity through marshal/unmarshal cycles

## example types

the example implements marshalers for:
- **CustomTime**: handles multiple time formats and timezone normalization
- **ContactInfo**: validates and normalizes contact information
- **LegacyUser**: handles legacy user data format conversion

## usage

```bash
go run main.go
```

## benefits

- **complete control**: full authority over binding/unbinding process
- **validation integration**: embed validation directly in the binding process
- **legacy support**: seamlessly handle legacy or non-standard data formats
- **computed properties**: dynamically calculate fields during binding/unbinding
- **error context**: provide detailed error information specific to the type