package df

import (
	"fmt"
	"reflect"
	"time"
)

// Unbind converts a struct (or pointer to struct) into a map[string]any
// honoring the same `df` tags used by Bind:
// - `df:"name"` overrides the key name
// - `df:"-"` skips the field
// - when no tag is provided, the key defaults to snake_case of the field name
//
// pointers to values: if nil, the key is omitted; otherwise the pointed value is emitted.
// slices, structs, and nested pointers are handled recursively. time.Duration values
// are emitted as strings using Duration.String() (e.g., "30s"). Interface fields are
// not supported, except for fields of type `Dynamic` (and slices of `Dynamic`), which
// are converted via their ToMap() method. Map-typed fields are not supported.
//
// opts are optional; pass nil or omit to use defaults.
func Unbind(source interface{}, opts ...*Options) (map[string]any, error) {
	if source == nil {
		return nil, fmt.Errorf("nil source provided")
	}
	val := reflect.ValueOf(source)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("nil pointer provided")
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("source must be a struct or pointer to struct; got %T", source)
	}
	opt, err := getOptions(opts...)
	if err != nil {
		return nil, err
	}
	return structToMap(val, opt)
}

func structToMap(structVal reflect.Value, opt *Options) (map[string]any, error) {
	out := make(map[string]any)
	structType := structVal.Type()
	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		// skip unexported fields
		if field.PkgPath != "" { // unexported
			continue
		}

		tag := parseDfTag(field)
		if tag.Skip {
			continue
		}
		name := tag.Name
		if name == "" {
			name = toSnakeCase(field.Name)
		}

		fieldVal := structVal.Field(i)
		// omit nil pointer fields entirely
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			continue
		}

		v, ok, err := valueToInterface(fieldVal, opt)
		if err != nil {
			return nil, fmt.Errorf("unbinding field %s.%s to key %q: %w", structType.Name(), field.Name, name, err)
		}
		if !ok {
			// nothing to emit (e.g., nil pointer)
			continue
		}
		out[name] = v
	}
	return out, nil
}

// valueToInterface converts a reflected value into an interface suitable for maps.
// returns (value, present, error). present=false indicates the value should be omitted
// (e.g., nil pointer). For time.Duration, emits its String() representation.
func valueToInterface(v reflect.Value, opt *Options) (interface{}, bool, error) {
	// check for custom converter first
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil, false, nil
	}
	if converted, wasConverted, err := tryCustomConverter(v.Type(), v.Interface(), opt, false); err != nil {
		return nil, false, err
	} else if wasConverted {
		return converted, true, nil
	}

	// check for custom marshaler implementation
	if v.Type().Implements(marshalerInterfaceType) {
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return nil, false, nil
		}
		m, err := v.Interface().(Marshaler).MarshalDf()
		return m, true, err
	}
	if v.CanAddr() {
		ptr := v.Addr()
		if ptr.Type().Implements(marshalerInterfaceType) {
			m, err := ptr.Interface().(Marshaler).MarshalDf()
			return m, true, err
		}
	}

	// handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, false, nil
		}
		return valueToInterface(v.Elem(), opt)
	}

	// special-case time.Duration (alias of int64)
	if v.Type() == reflect.TypeOf(time.Duration(0)) {
		d := time.Duration(v.Int())
		return d.String(), true, nil
	}

	switch v.Kind() {
	case reflect.Struct:
		// check if this is a Pointer[T] type
		if isPointerType(v.Type()) {
			return pointerToMap(v)
		}

		// if the concrete struct implements Dynamic (directly or via pointer receiver),
		// prefer serializing via ToMap() to preserve the discriminator and schema.
		if v.Type().Implements(dynamicInterfaceType) {
			dyn := v.Interface().(Dynamic)
			return dynamicToMap(dyn), true, nil
		}
		if v.CanAddr() {
			ptr := v.Addr()
			if ptr.Type().Implements(dynamicInterfaceType) {
				dyn := ptr.Interface().(Dynamic)
				return dynamicToMap(dyn), true, nil
			}
		}
		m, err := structToMap(v, opt)
		if err != nil {
			return nil, false, err
		}
		return m, true, nil

	case reflect.Slice:
		length := v.Len()
		arr := make([]interface{}, 0, length)
		// special handling for slices of Dynamic (either interface type or concrete types implementing it)
		elemType := v.Type().Elem()
		isDynamicElem := false
		if elemType.Kind() == reflect.Interface {
			isDynamicElem = elemType == dynamicInterfaceType || elemType.Implements(dynamicInterfaceType)
		} else {
			isDynamicElem = elemType.Implements(dynamicInterfaceType)
		}
		if isDynamicElem {
			for i := 0; i < length; i++ {
				if v.Index(i).IsZero() {
					arr = append(arr, nil)
					continue
				}
				// recover the Dynamic interface from the original element value
				dynIfaceVal := v.Index(i).Interface()
				if dynIfaceVal == nil {
					arr = append(arr, nil)
					continue
				}
				dyn, ok := dynIfaceVal.(Dynamic)
				if !ok {
					return nil, false, fmt.Errorf("index %d: element does not implement Dynamic", i)
				}
				arr = append(arr, dynamicToMap(dyn))
			}
			return arr, true, nil
		}
		for i := 0; i < length; i++ {
			elem := v.Index(i)
			converted, present, err := valueToInterface(elem, opt)
			if err != nil {
				return nil, false, fmt.Errorf("index %d: %w", i, err)
			}
			if !present {
				// keep nils to preserve positional semantics
				arr = append(arr, nil)
				continue
			}
			arr = append(arr, converted)
		}
		return arr, true, nil

	case reflect.Interface:
		// omit nil interfaces
		if v.IsNil() {
			return nil, false, nil
		}
		// support Dynamic interface by delegating to ToMap(); handle both when the field type is Dynamic and when the
		// concrete value implements it
		if v.Type().Implements(dynamicInterfaceType) || reflect.TypeOf(v.Interface()).Implements(dynamicInterfaceType) {
			dyn := v.Interface().(Dynamic)
			return dynamicToMap(dyn), true, nil
		}
		return nil, false, fmt.Errorf("interface fields are not supported")

	case reflect.Map:
		return nil, false, fmt.Errorf("map fields are not supported")

	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return v.Interface(), true, nil
	}

	return nil, false, fmt.Errorf("unsupported kind %s", v.Kind())
}

// dynamicToMap converts a Dynamic value to a map and enforces that the discriminator key "type" is present and
// consistent with d.Type(). if ToMap() returns nil, an empty map is created.
func dynamicToMap(d Dynamic) map[string]any {
	m := d.ToMap()
	if m == nil {
		m = make(map[string]any)
	}
	m[TypeKey] = d.Type()
	return m
}
