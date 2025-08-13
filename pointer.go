package df

import (
	"fmt"
	"reflect"
)

// Pointer represents a reference to an object of type T that implements Identifiable.
// During binding, the reference is stored as a string. During linking, it's resolved to the actual object.
type Pointer[T Identifiable] struct {
	Ref      string `df:"$ref"`
	Resolved T      // internal resolved reference (exported for reflection)
}

// Resolve returns the resolved object, or the zero value of T if not yet resolved.
func (p *Pointer[T]) Resolve() T {
	return p.Resolved
}

// IsResolved returns true if the pointer has been resolved to an actual object.
func (p *Pointer[T]) IsResolved() bool {
	v := reflect.ValueOf(p.Resolved)
	if !v.IsValid() {
		return false
	}
	if v.Kind() == reflect.Ptr {
		return !v.IsNil()
	}
	return !v.IsZero()
}

// Link resolves all pointer references in the target object by building a registry of all
// Identifiable objects and then resolving Pointer fields to their target objects.
// objects are namespaced by their concrete type to prevent Id clashes between different types.
func Link(target interface{}) error {
	if target == nil {
		return fmt.Errorf("nil target provided")
	}
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to struct; got %T", target)
	}
	elem := value.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct; got %T", target)
	}

	// phase 1: collect all Identifiable objects with type-prefixed keys
	registry := make(map[string]reflect.Value)
	collectIdentifiableObjects(elem, registry)

	// phase 2: resolve all pointer references
	return resolvePointers(elem, registry)
}

// collectIdentifiableObjects recursively traverses the object tree and collects all
// objects that implement Identifiable, storing them with type-prefixed IDs.
func collectIdentifiableObjects(value reflect.Value, registry map[string]reflect.Value) {
	switch value.Kind() {
	case reflect.Struct:
		// check if this struct implements Identifiable
		if value.Addr().Type().Implements(identifiableInterfaceType) {
			identifiable := value.Addr().Interface().(Identifiable)
			typePrefix := value.Type().String()
			key := typePrefix + ":" + identifiable.GetId()
			registry[key] = value.Addr()
		}

		// recursively process struct fields
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if field.PkgPath != "" { // skip unexported fields
				continue
			}
			_, _, skip := parseDFTag(field)
			if skip {
				continue
			}
			collectIdentifiableObjects(value.Field(i), registry)
		}

	case reflect.Ptr:
		if !value.IsNil() {
			collectIdentifiableObjects(value.Elem(), registry)
		}

	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			collectIdentifiableObjects(value.Index(i), registry)
		}
	}
}

// resolvePointers recursively traverses the object tree and resolves all Pointer fields.
func resolvePointers(value reflect.Value, registry map[string]reflect.Value) error {
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if field.PkgPath != "" { // skip unexported fields
				continue
			}
			_, _, skip := parseDFTag(field)
			if skip {
				continue
			}

			fieldValue := value.Field(i)
			if err := resolvePointersInField(fieldValue, field.Type, registry); err != nil {
				return fmt.Errorf("resolving pointers in field %s: %w", field.Name, err)
			}
		}

	case reflect.Ptr:
		if !value.IsNil() {
			return resolvePointers(value.Elem(), registry)
		}

	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			if err := resolvePointers(value.Index(i), registry); err != nil {
				return fmt.Errorf("resolving pointers in slice[%d]: %w", i, err)
			}
		}
	}
	return nil
}

// resolvePointersInField handles pointer resolution for a specific field.
func resolvePointersInField(fieldValue reflect.Value, fieldType reflect.Type, registry map[string]reflect.Value) error {
	switch fieldValue.Kind() {
	case reflect.Ptr:
		if !fieldValue.IsNil() {
			// check if this is a Pointer[T] type
			if isPointerType(fieldType.Elem()) {
				return resolvePointerField(fieldValue.Elem(), fieldType.Elem(), registry)
			}
			// regular pointer, recurse into it
			return resolvePointersInField(fieldValue.Elem(), fieldType.Elem(), registry)
		}

	case reflect.Struct:
		// check if this is a Pointer[T] type
		if isPointerType(fieldType) {
			return resolvePointerField(fieldValue, fieldType, registry)
		}
		// regular struct, recurse into it
		return resolvePointers(fieldValue, registry)

	case reflect.Slice:
		for i := 0; i < fieldValue.Len(); i++ {
			elemType := fieldType.Elem()
			if err := resolvePointersInField(fieldValue.Index(i), elemType, registry); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
	}
	return nil
}

// isPointerType checks if the given type is a Pointer[T] generic type.
func isPointerType(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	// check if it has the structure of Pointer[T]: a "Ref" field and a "Resolved" field
	if t.NumField() >= 2 {
		field0 := t.Field(0)
		field1 := t.Field(1)
		return field0.Name == "Ref" && field0.Type.Kind() == reflect.String && field1.Name == "Resolved"
	}
	return false
}

// resolvePointerField resolves a single Pointer[T] field.
func resolvePointerField(pointerValue reflect.Value, pointerType reflect.Type, registry map[string]reflect.Value) error {
	refField := pointerValue.FieldByName("Ref")
	if !refField.IsValid() || refField.Kind() != reflect.String {
		return fmt.Errorf("invalid Pointer type: missing Ref field")
	}

	ref := refField.String()
	if ref == "" {
		return nil // empty reference, nothing to resolve
	}

	resolvedField := pointerValue.FieldByName("Resolved")
	if !resolvedField.IsValid() || !resolvedField.CanSet() {
		return fmt.Errorf("invalid Pointer type: missing or non-settable Resolved field")
	}

	// determine the target type from the resolved field's type
	targetType := resolvedField.Type()
	// handle both *T and T types
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}
	typePrefix := targetType.String()
	key := typePrefix + ":" + ref

	// look up the target object in the registry
	targetValue, exists := registry[key]
	if !exists {
		return fmt.Errorf("unresolved reference: %s (looking for %s)", ref, key)
	}

	// set the resolved field to the target object
	// if resolved field expects a pointer, use the registry value directly
	// if resolved field expects a value, dereference it
	if resolvedField.Type().Kind() == reflect.Ptr {
		resolvedField.Set(targetValue)
	} else {
		resolvedField.Set(targetValue.Elem())
	}
	return nil
}

// bindPointer binds data to a Pointer[T] field during the bind phase. only the $ref field is populated; resolution
// happens during the Link phase.
func bindPointer(pointerValue reflect.Value, data map[string]any, path string) error {
	// get the Ref field and set it from the $ref key in the data
	refField := pointerValue.FieldByName("Ref")
	if !refField.IsValid() || !refField.CanSet() || refField.Kind() != reflect.String {
		return fmt.Errorf("%s: invalid Pointer type: missing or non-settable Ref field", path)
	}

	refVal, ok := data["$ref"]
	if !ok {
		// empty reference is valid
		return nil
	}

	refStr, ok := refVal.(string)
	if !ok {
		return fmt.Errorf("%s: $ref must be a string, got %T", path, refVal)
	}

	refField.SetString(refStr)
	return nil
}

// pointerToMap converts a Pointer[T] struct to a map containing the $ref field.
func pointerToMap(pointerValue reflect.Value) (interface{}, bool, error) {
	refField := pointerValue.FieldByName("Ref")
	if !refField.IsValid() || refField.Kind() != reflect.String {
		return nil, false, fmt.Errorf("invalid Pointer type: missing Ref field")
	}

	ref := refField.String()
	if ref == "" {
		// empty reference - could omit entirely or include empty $ref
		return nil, false, nil
	}

	return map[string]any{"$ref": ref}, true, nil
}