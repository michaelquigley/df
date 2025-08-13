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
}

// Bind populates the exported fields of target (a pointer to a struct) from the given data map. Keys are matched using
// either a struct tag `df:"name,required"` (where name overrides the key and the optional "required" flag enforces
// presence), `df:"-"` to skip a field, or, when no tag is provided, a best-effort snake_case conversion of the
// field name.
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
	var opt *Options
	if len(opts) == 1 {
		opt = opts[0]
	} else if len(opts) > 1 {
		return fmt.Errorf("only one option allowed, got %d", len(opts))
	}
	return bindStruct(elem, data, elem.Type().Name(), opt)
}

func bindStruct(structValue reflect.Value, data map[string]any, path string, opt *Options) error {
	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		// skip unexported fields
		if field.PkgPath != "" { // unexported
			continue
		}

		fieldVal := structValue.Field(i)
		name, required, skip := parseDFTag(field)
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

		if err := setField(fieldVal, raw, path+"."+field.Name, opt); err != nil {
			return fmt.Errorf("binding field %s.%s from key %q: %w", path, field.Name, name, err)
		}
	}
	return nil
}

func setField(fieldVal reflect.Value, raw interface{}, path string, opt *Options) error {
	fieldType := fieldVal.Type()

	// handle pointers by allocating as needed then setting the element
	if fieldType.Kind() == reflect.Ptr {
		// if pointer to struct, allocate new struct and bind into it
		elemType := fieldType.Elem()
		newPtr := reflect.New(elemType)
		if elemType.Kind() == reflect.Struct {
			subMap, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("%s: expected object for struct pointer, got %T", path, raw)
			}
			if err := bindStruct(newPtr.Elem(), subMap, path, opt); err != nil {
				return err
			}
			fieldVal.Set(newPtr)
			return nil
		}
		// pointer to primitive or slice
		if err := setNonPtrValue(newPtr.Elem(), raw, path, opt); err != nil {
			return err
		}
		fieldVal.Set(newPtr)
		return nil
	}

	return setNonPtrValue(fieldVal, raw, path, opt)
}

func setNonPtrValue(fieldVal reflect.Value, raw interface{}, path string, opt *Options) error {
	switch fieldVal.Kind() {
	case reflect.Struct:
		subMap, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object for struct, got %T", path, raw)
		}
		return bindStruct(fieldVal, subMap, path, opt)

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
					if err := bindStruct(elemPtr.Elem(), subMap, itemPath, opt); err != nil {
						return err
					}
					out = reflect.Append(out, elemPtr)
					continue
				}
				// pointer to primitive element
				if err := setNonPtrValue(elemPtr.Elem(), item, itemPath, opt); err != nil {
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
				if err := bindStruct(elemVal, subMap, itemPath, opt); err != nil {
					return err
				}
				out = reflect.Append(out, elemVal)
				continue
			}
			if err := convertAndSet(elemVal, item, itemPath); err != nil {
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
		return convertAndSet(fieldVal, raw, path)
	}
}

// bindDynamic resolves a Dynamic implementation from a map using the Options registry.
func bindDynamic(m map[string]any, path string, opt *Options) (Dynamic, error) {
	if opt == nil {
		return nil, fmt.Errorf("%s: no options provided to resolve Dynamic field", path)
	}
	tVal, ok := m["type"]
	if !ok {
		return nil, fmt.Errorf("%s: missing 'type' discriminator for Dynamic field", path)
	}
	typeStr, ok := tVal.(string)
	if !ok || strings.TrimSpace(typeStr) == "" {
		return nil, fmt.Errorf("%s: invalid 'type' discriminator for Dynamic field: %v", path, tVal)
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







