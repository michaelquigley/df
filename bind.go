package df

import (
	"fmt"
	"reflect"
	"strings"
)

// Options configures binding behavior.
type Options struct {
	// DynamicBinders maps a discriminator string (found under the "type" key in the input map) to a function that
	// consumes the full map and returns a concrete value implementing the Dynamic interface.
	DynamicBinders map[string]func(map[string]any) (Dynamic, error)

	// FieldDynamicBinders allows specifying binder sets per field path. The key is the structured path of the field as
	// used internally by Bind, e.g.: "Root.Items" for a slice field, "Root.Nested.Field" for nested fields.
	// any array indices in the path are ignored for matching purposes.
	// when present for a field, this map takes precedence over DynamicBinders.
	FieldDynamicBinders map[string]map[string]func(map[string]any) (Dynamic, error)

	// Converters maps Go types to custom converters for type conversion.
	// the key is the reflect.Type of the target field, and the value is a Converter
	// that handles bidirectional conversion between raw data and the target type.
	Converters map[reflect.Type]Converter
}

// Bind populates the exported fields of target (a pointer to a struct) from the given data map. Keys are matched using
// either a struct tag `df:"name,required"` (where name overrides the key and the optional "required" flag enforces
// presence), `df:"-"` to skip a field, or, when no tag is provided, a best-effort snake_case conversion of the
// field name.
//
// Use Bind when you need to control how the prototype object is allocated. Use New when you just want to allocate a new
// object to bind off the heap.
//
// supported kinds:
// - primitives: string, bool, all int/uint sizes, float32/64, time.Duration
// - pointers to the above
// - structs and pointers to structs (recursively bound from map[string]any)
// - slices of the above (slice items are bound from []interface{})
//
// interface types and maps are not supported and will return an error if encountered,
// except for fields of type Dynamic which are resolved using Options.DynamicBinders.
//
// opts are optional; pass nil or omit to use defaults.
func Bind(target interface{}, data map[string]any, opts ...*Options) error {
	elem, err := validateTarget(target)
	if err != nil {
		return err
	}
	opt, err := getOptions(opts...)
	if err != nil {
		return err
	}
	return bindStruct(elem, data, elem.Type().Name(), opt, false)
}

// New creates and populates a new instance of type T from the given data map.
// Unlike Bind, which requires a pre-allocated target pointer, New automatically
// allocates the object and returns a pointer to the populated struct.
//
// Use Bind instead of New when you need to control where and how the target object is instantiated. New just allocates
// a fresh target off the heap.
//
// Example usage:
//
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//
//	data := map[string]any{"name": "John", "age": 30}
//	person, err := New[Person](data)
//	if err != nil {
//	    // handle error
//	}
//	// person is now *Person with Name="John" and Age=30
//
// supported kinds and field mapping rules are the same as Bind.
//
// opts are optional; pass nil or omit to use defaults.
func New[T any](data map[string]any, opts ...*Options) (*T, error) {
	// Create new instance of T
	target := new(T)

	// Use existing Bind function to populate it
	err := Bind(target, data, opts...)
	if err != nil {
		return nil, err
	}

	return target, nil
}

// Merge populates the exported fields of an existing target struct from the given data map, preserving
// any existing field values that are not present in the data. This allows binding partial data to
// pre-initialized structs with default values.
//
// uses the same field mapping rules as Bind: struct tags, snake_case conversion, etc.
//
// supported kinds are the same as Bind.
//
// opts are optional; pass nil or omit to use defaults.
func Merge(target interface{}, data map[string]any, opts ...*Options) error {
	elem, err := validateTarget(target)
	if err != nil {
		return err
	}
	opt, err := getOptions(opts...)
	if err != nil {
		return err
	}
	return bindStruct(elem, data, elem.Type().Name(), opt, true)
}

func bindStruct(structValue reflect.Value, data map[string]any, path string, opt *Options, preserveExisting bool) error {
	structType := structValue.Type()

	type deferredUnmarshal struct {
		fieldVal reflect.Value
		rawData  interface{}
		path     string
		name     string
	}
	var deferred []deferredUnmarshal

	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}

		fieldVal := structValue.Field(i)
		name, required, skip := parseDfTag(field)
		if skip {
			continue
		}
		if name == "" {
			name = toSnakeCase(field.Name)
		}

		raw, ok := data[name]
		if !ok {
			if required {
				return fmt.Errorf("%s.%s: required field missing", path, field.Name)
			}
			continue
		}

		// defer custom unmarshalers to run after all other fields are bound.
		if (fieldVal.CanAddr() && fieldVal.Addr().Type().Implements(unmarshalerInterfaceType)) || fieldVal.Type().Implements(unmarshalerInterfaceType) {
			deferred = append(deferred, deferredUnmarshal{
				fieldVal: fieldVal,
				rawData:  raw,
				path:     path + "." + field.Name,
				name:     name,
			})
			continue
		}

		if err := setField(fieldVal, raw, path+"."+field.Name, opt, preserveExisting); err != nil {
			return fmt.Errorf("binding field %s.%s from key %q: %w", path, field.Name, name, err)
		}
	}

	// run deferred unmarshalers now that all other fields are populated.
	for _, d := range deferred {
		if err := unmarshalFromMap(d.fieldVal, d.rawData, d.path); err != nil {
			return fmt.Errorf("binding field %s from key %q: %w", d.path, d.name, err)
		}
	}

	return nil
}

// unmarshalFromMap handles calling the UnmarshalDf method on a field.
func unmarshalFromMap(fieldVal reflect.Value, raw interface{}, path string) error {
	subMap, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("%s: expected object for unmarshaler, got %T", path, raw)
	}

	// handle pointer vs. value receiver for the unmarshaler
	if fieldVal.CanAddr() {
		ptr := fieldVal.Addr()
		if ptr.Type().Implements(unmarshalerInterfaceType) {
			// ensure pointer is allocated for pointer fields
			if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
				fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
			}
			return ptr.Interface().(Unmarshaler).UnmarshalDf(subMap)
		}
	}

	// must be a pointer type that implements the interface directly
	if fieldVal.Type().Implements(unmarshalerInterfaceType) {
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
		}
		return fieldVal.Interface().(Unmarshaler).UnmarshalDf(subMap)
	}

	return fmt.Errorf("%s: internal error: field does not implement unmarshaler", path) // should be unreachable
}

func setField(fieldVal reflect.Value, raw interface{}, path string, opt *Options, preserveExisting bool) error {
	fieldType := fieldVal.Type()

	// handle pointers by allocating as needed then setting the element
	if fieldType.Kind() == reflect.Ptr {
		elemType := fieldType.Elem()
		if elemType.Kind() == reflect.Struct {
			subMap, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("%s: expected object for struct pointer, got %T", path, raw)
			}
			// if preserveExisting and pointer is not nil, bind to existing struct
			if preserveExisting && !fieldVal.IsNil() {
				if err := bindStruct(fieldVal.Elem(), subMap, path, opt, preserveExisting); err != nil {
					return err
				}
			} else {
				// allocate new struct and bind into it
				newPtr := reflect.New(elemType)
				if err := bindStruct(newPtr.Elem(), subMap, path, opt, preserveExisting); err != nil {
					return err
				}
				fieldVal.Set(newPtr)
			}
			return nil
		}
		// pointer to primitive or slice
		newPtr := reflect.New(elemType)
		if err := setNonPtrValue(newPtr.Elem(), raw, path, opt, preserveExisting); err != nil {
			return err
		}
		fieldVal.Set(newPtr)
		return nil
	}

	return setNonPtrValue(fieldVal, raw, path, opt, preserveExisting)
}

func setNonPtrValue(fieldVal reflect.Value, raw interface{}, path string, opt *Options, preserveExisting bool) error {
	// check for custom converter first
	if converted, wasConverted, err := tryCustomConverter(fieldVal.Type(), raw, opt, true); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	} else if wasConverted {
		fieldVal.Set(reflect.ValueOf(converted))
		return nil
	}

	switch fieldVal.Kind() {
	case reflect.Struct:
		subMap, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object for struct, got %T", path, raw)
		}
		return bindStruct(fieldVal, subMap, path, opt, preserveExisting)

	case reflect.Slice:
		rawVal := reflect.ValueOf(raw)
		if rawVal.Kind() != reflect.Slice {
			return fmt.Errorf("%s: expected array for slice, got %T", path, raw)
		}
		elemType := fieldVal.Type().Elem()
		out := reflect.MakeSlice(fieldVal.Type(), 0, rawVal.Len())
		// handle slices of Dynamic interface specially
		if elemType.Kind() == reflect.Interface && elemType == dynamicInterfaceType {
			for idx := 0; idx < rawVal.Len(); idx++ {
				item := rawVal.Index(idx).Interface()
				itemPath := fmt.Sprintf("%s[%d]", path, idx)
				subMap, ok := item.(map[string]any)
				if !ok {
					return fmt.Errorf("%s: expected object for Dynamic element, got %T", itemPath, item)
				}
				dynVal, err := bindDynamic(subMap, itemPath, opt)
				if err != nil {
					return err
				}
				out = reflect.Append(out, reflect.ValueOf(dynVal))
			}
			fieldVal.Set(out)
			return nil
		}
		for idx := 0; idx < rawVal.Len(); idx++ {
			item := rawVal.Index(idx).Interface()
			itemPath := fmt.Sprintf("%s[%d]", path, idx)
			if elemType.Kind() == reflect.Ptr {
				elemPtr := reflect.New(elemType.Elem())
				if elemType.Elem().Kind() == reflect.Struct {
					subMap, ok := item.(map[string]any)
					if !ok {
						return fmt.Errorf("%s: expected object for struct slice element, got %T", itemPath, item)
					}
					if err := bindStruct(elemPtr.Elem(), subMap, itemPath, opt, preserveExisting); err != nil {
						return err
					}
					out = reflect.Append(out, elemPtr)
					continue
				}
				// pointer to primitive element
				if err := setNonPtrValue(elemPtr.Elem(), item, itemPath, opt, preserveExisting); err != nil {
					return err
				}
				out = reflect.Append(out, elemPtr)
				continue
			}

			// non-pointer element
			elemVal := reflect.New(elemType).Elem()
			if elemType.Kind() == reflect.Struct {
				subMap, ok := item.(map[string]any)
				if !ok {
					return fmt.Errorf("%s: expected object for struct slice element, got %T", itemPath, item)
				}
				if err := bindStruct(elemVal, subMap, itemPath, opt, preserveExisting); err != nil {
					return err
				}
				out = reflect.Append(out, elemVal)
				continue
			}
			if err := convertAndSet(elemVal, item, itemPath, opt); err != nil {
				return err
			}
			out = reflect.Append(out, elemVal)
		}
		fieldVal.Set(out)
		return nil

	case reflect.Interface:
		// support fields of type Dynamic via binder map
		if fieldVal.Type() == dynamicInterfaceType {
			subMap, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("%s: expected object for Dynamic, got %T", path, raw)
			}
			dynVal, err := bindDynamic(subMap, path, opt)
			if err != nil {
				return err
			}
			fieldVal.Set(reflect.ValueOf(dynVal))
			return nil
		}
		return fmt.Errorf("%s: interface fields are not supported", path)

	default:
		// check if this is a Pointer[T] type before falling back to convertAndSet
		if isPointerType(fieldVal.Type()) {
			subMap, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("%s: expected object for Pointer, got %T", path, raw)
			}
			return bindPointer(fieldVal, subMap, path)
		}
		return convertAndSet(fieldVal, raw, path, opt)
	}
}

// bindDynamic resolves a Dynamic implementation from a map using the Options registry.
func bindDynamic(m map[string]any, path string, opt *Options) (Dynamic, error) {
	if opt == nil {
		return nil, fmt.Errorf("%s: no options provided to resolve Dynamic field", path)
	}
	tVal, ok := m[TypeKey]
	if !ok {
		return nil, fmt.Errorf("%s: missing '%v' discriminator for Dynamic field", path, TypeKey)
	}
	typeStr, ok := tVal.(string)
	if !ok || strings.TrimSpace(typeStr) == "" {
		return nil, fmt.Errorf("%s: invalid '%v' discriminator for Dynamic field: %v", path, TypeKey, tVal)
	}
	// prefer field-specific binder set if provided
	var binder func(map[string]any) (Dynamic, error)
	if opt.FieldDynamicBinders != nil {
		if perField, ok := opt.FieldDynamicBinders[stripIndices(path)]; ok && perField != nil {
			binder = perField[typeStr]
		}
	}
	// fall back to global binders
	if binder == nil && opt.DynamicBinders != nil {
		binder = opt.DynamicBinders[typeStr]
	}
	if binder == nil {
		return nil, fmt.Errorf("%s: unknown Dynamic type %q", path, typeStr)
	}
	dynVal, err := binder(m)
	if err != nil {
		return nil, fmt.Errorf("%s: binding Dynamic type %q failed: %w", path, typeStr, err)
	}
	return dynVal, nil
}

// stripIndices removes any array index segments (e.g., "[0]") from a path like
// "Root.Items[0].Action", yielding "Root.Items.Action" for stable field matching.
func stripIndices(path string) string {
	if strings.IndexByte(path, '[') == -1 {
		return path
	}

	// Count brackets to estimate result size more accurately
	bracketCount := 0
	for _, r := range path {
		if r == '[' || r == ']' {
			bracketCount++
		}
	}

	// Estimate capacity: original length minus approximate bracket content
	// Assume average index is 2 chars (e.g., "[0]", "[12]")
	estimatedSize := len(path) - (bracketCount/2)*3 // bracketCount/2 pairs, ~3 chars each
	if estimatedSize < 0 {
		estimatedSize = len(path) / 2
	}

	var b strings.Builder
	b.Grow(estimatedSize)
	skip := 0
	for _, r := range path {
		switch r {
		case '[':
			skip++
			continue
		case ']':
			if skip > 0 {
				skip--
			}
			continue
		}
		if skip == 0 {
			b.WriteRune(r)
		}
	}
	return b.String()
}
