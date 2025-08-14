package df

import (
	"fmt"
	"reflect"
	"strings"
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

// LinkerOptions configures the behavior of a Linker instance.
type LinkerOptions struct {
	// EnableCaching enables registry caching for repeated linking operations
	EnableCaching bool
	// AllowPartialResolution allows linking to succeed even if some references can't be resolved
	AllowPartialResolution bool
}

// Linker encapsulates the linking process, providing enhanced state management and advanced features.
type Linker struct {
	options LinkerOptions
	cache   map[string]reflect.Value // cached registry for repeated operations
}

// NewLinker creates a new Linker with optional options.
// If no options are provided, default options are used.
func NewLinker(opts ...LinkerOptions) *Linker {
	var options LinkerOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	l := &Linker{
		options: options,
	}
	if options.EnableCaching {
		l.cache = make(map[string]reflect.Value)
	}
	return l
}

// ClearCache clears the internal registry cache if caching is enabled.
func (l *Linker) ClearCache() {
	if l.cache != nil {
		l.cache = make(map[string]reflect.Value)
	}
}

// Register performs phase 1 of linking: collecting all Identifiable objects.
// This can be used for multi-stage linking where you want to register objects from
// multiple sources before resolving references.
func (l *Linker) Register(targets ...interface{}) error {
	if len(targets) == 0 {
		return fmt.Errorf("no targets provided")
	}

	// ensure cache is available for collection
	if l.cache == nil {
		l.cache = make(map[string]reflect.Value)
	}

	for i, target := range targets {
		if target == nil {
			return fmt.Errorf("nil target provided at index %d", i)
		}
		value := reflect.ValueOf(target)
		if value.Kind() != reflect.Ptr || value.IsNil() {
			return fmt.Errorf("target at index %d must be a non-nil pointer to struct; got %T", i, target)
		}
		elem := value.Elem()
		if elem.Kind() != reflect.Struct {
			return fmt.Errorf("target at index %d must be a pointer to struct; got %T", i, target)
		}

		l.collectIdentifiableObjects(elem, l.cache)
	}
	return nil
}

// ResolveReferences performs phase 2 of linking: resolving all pointer references using
// the collected registry. This can be used after collecting from multiple sources.
func (l *Linker) ResolveReferences(target interface{}) error {
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

	if l.cache == nil {
		return fmt.Errorf("no registry available - call Register first")
	}

	return l.resolvePointers(elem, l.cache)
}

// Link resolves all pointer references in the target objects by building a registry of all
// Identifiable objects and then resolving Pointer fields to their target objects.
// objects are namespaced by their concrete type to prevent Id clashes between different types.
func (l *Linker) Link(targets ...interface{}) error {
	if len(targets) == 0 {
		return fmt.Errorf("no targets provided")
	}

	// phase 1: collect all Identifiable objects with type-prefixed keys
	var registry map[string]reflect.Value
	if l.options.EnableCaching && l.cache != nil {
		registry = l.cache
		// only collect if cache is empty or we want to refresh it
		if len(registry) == 0 {
			for i, target := range targets {
				if err := l.validateAndCollect(target, i, registry); err != nil {
					return err
				}
			}
		}
	} else {
		registry = make(map[string]reflect.Value)
		for i, target := range targets {
			if err := l.validateAndCollect(target, i, registry); err != nil {
				return err
			}
		}
	}

	// phase 2: resolve all pointer references in all targets
	for i, target := range targets {
		if target == nil {
			return fmt.Errorf("nil target provided at index %d", i)
		}
		value := reflect.ValueOf(target)
		if value.Kind() != reflect.Ptr || value.IsNil() {
			return fmt.Errorf("target at index %d must be a non-nil pointer to struct; got %T", i, target)
		}
		elem := value.Elem()
		if elem.Kind() != reflect.Struct {
			return fmt.Errorf("target at index %d must be a pointer to struct; got %T", i, target)
		}

		if err := l.resolvePointers(elem, registry); err != nil {
			return fmt.Errorf("resolving pointers in target %d: %w", i, err)
		}
	}
	return nil
}

// validateAndCollect validates a target and collects its identifiable objects.
func (l *Linker) validateAndCollect(target interface{}, index int, registry map[string]reflect.Value) error {
	elem, err := validateTarget(target)
	if err != nil {
		return fmt.Errorf("target at index %d: %w", index, err)
	}

	l.collectIdentifiableObjects(elem, registry)
	return nil
}

// Link resolves all pointer references in the target objects by building a registry of all
// Identifiable objects and then resolving Pointer fields to their target objects.
// objects are namespaced by their concrete type to prevent Id clashes between different types.
func Link(targets ...interface{}) error {
	linker := NewLinker()
	return linker.Link(targets...)
}

// collectIdentifiableObjects recursively traverses the object tree and collects all
// objects that implement Identifiable, storing them with type-prefixed IDs.
func (l *Linker) collectIdentifiableObjects(value reflect.Value, registry map[string]reflect.Value) {
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
			_, _, skip := parseDfTag(field)
			if skip {
				continue
			}
			l.collectIdentifiableObjects(value.Field(i), registry)
		}

	case reflect.Ptr:
		if !value.IsNil() {
			l.collectIdentifiableObjects(value.Elem(), registry)
		}

	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			l.collectIdentifiableObjects(value.Index(i), registry)
		}
	}
}

// resolvePointers recursively traverses the object tree and resolves all Pointer fields.
func (l *Linker) resolvePointers(value reflect.Value, registry map[string]reflect.Value) error {
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if field.PkgPath != "" { // skip unexported fields
				continue
			}
			_, _, skip := parseDfTag(field)
			if skip {
				continue
			}

			fieldValue := value.Field(i)
			if err := l.resolvePointersInField(fieldValue, field.Type, registry); err != nil {
				return fmt.Errorf("resolving pointers in field %s: %w", field.Name, err)
			}
		}

	case reflect.Ptr:
		if !value.IsNil() {
			return l.resolvePointers(value.Elem(), registry)
		}

	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			if err := l.resolvePointers(value.Index(i), registry); err != nil {
				return fmt.Errorf("resolving pointers in slice[%d]: %w", i, err)
			}
		}
	}
	return nil
}

// resolvePointersInField handles pointer resolution for a specific field.
func (l *Linker) resolvePointersInField(fieldValue reflect.Value, fieldType reflect.Type, registry map[string]reflect.Value) error {
	switch fieldValue.Kind() {
	case reflect.Ptr:
		if !fieldValue.IsNil() {
			// check if this is a Pointer[T] type
			if isPointerType(fieldType.Elem()) {
				return l.resolvePointerField(fieldValue.Elem(), fieldType.Elem(), registry)
			}
			// regular pointer, recurse into it
			return l.resolvePointersInField(fieldValue.Elem(), fieldType.Elem(), registry)
		}

	case reflect.Struct:
		// check if this is a Pointer[T] type
		if isPointerType(fieldType) {
			return l.resolvePointerField(fieldValue, fieldType, registry)
		}
		// regular struct, recurse into it
		return l.resolvePointers(fieldValue, registry)

	case reflect.Slice:
		for i := 0; i < fieldValue.Len(); i++ {
			elemType := fieldType.Elem()
			if err := l.resolvePointersInField(fieldValue.Index(i), elemType, registry); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
	}
	return nil
}

// isPointerType checks if the given type is a Pointer[T] generic type.
// performs more robust checking including package path and struct tags.
func isPointerType(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	// check if the type name starts with "Pointer[" (generic instantiation)
	if typeName := t.Name(); typeName != "" && !strings.HasPrefix(typeName, "Pointer[") {
		return false
	}

	// check if it has the exact structure of Pointer[T]: a "Ref" field with df:"$ref" tag and a "Resolved" field
	if t.NumField() != 2 {
		return false
	}

	field0 := t.Field(0)
	field1 := t.Field(1)

	// verify first field is "Ref" with correct type and tag
	if field0.Name != "Ref" || field0.Type.Kind() != reflect.String {
		return false
	}

	// verify the df struct tag matches our RefKey
	if dfTag := field0.Tag.Get("df"); dfTag != RefKey {
		return false
	}

	// verify second field is "Resolved"
	if field1.Name != "Resolved" {
		return false
	}

	return true
}

// resolvePointerField resolves a single Pointer[T] field.
func (l *Linker) resolvePointerField(pointerValue reflect.Value, pointerType reflect.Type, registry map[string]reflect.Value) error {
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
		if l.options.AllowPartialResolution {
			// skip this resolution but don't fail the entire process
			return nil
		}
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

	refVal, ok := data[RefKey]
	if !ok {
		// empty reference is valid
		return nil
	}

	refStr, ok := refVal.(string)
	if !ok {
		return fmt.Errorf("%s: '%s' must be a string, got '%T'", path, RefKey, refVal)
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

	return map[string]any{RefKey: ref}, true, nil
}
