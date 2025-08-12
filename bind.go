package df

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Options configures binding behavior.
type Options struct {
	// DynamicBinders maps a discriminator string (found under the "type" key in the input map) to a function that
	// consumes the full map and returns a concrete value implementing the Dynamic interface.
	DynamicBinders map[string]func(map[string]any) (Dynamic, error)

	// FieldDynamicBinders allows specifying binder sets per field path. The key is the structured path of the field as
	// used internally by Bind, e.g.: "Root.Items" for a slice field, "Root.Nested.Field" for nested fields.
	// Any array indices in the path are ignored for matching purposes.
	// When present for a field, this map takes precedence over DynamicBinders.
	FieldDynamicBinders map[string]map[string]func(map[string]any) (Dynamic, error)
}

// Bind populates the exported fields of target (a pointer to a struct) from the given data map. Keys are matched using
// either a struct tag `df:"name,required"` (where name overrides the key and the optional "required" flag enforces
// presence), `df:"-"` to skip a field, or, when no tag is provided, a best-effort snake_case conversion of the
// field name.
//
// Supported kinds:
// - primitives: string, bool, all int/uint sizes, float32/64, time.Duration
// - pointers to the above
// - structs and pointers to structs (recursively bound from map[string]any)
// - slices of the above (slice items are bound from []interface{})
//
// Interface types and maps are not supported and will return an error if encountered,
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
			return err
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
		// Support fields of type Dynamic via binder map
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
	// Prefer field-specific binder set if provided
	var binder func(map[string]any) (Dynamic, error)
	if opt.FieldDynamicBinders != nil {
		if perField, ok := opt.FieldDynamicBinders[stripIndices(path)]; ok && perField != nil {
			binder = perField[typeStr]
		}
	}
	// Fall back to global binders
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
	var b strings.Builder
	b.Grow(len(path))
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

func convertAndSet(dst reflect.Value, raw interface{}, path string) error {
	dstKind := dst.Kind()
	// Special-case time.Duration (which is an int64 alias)
	if dst.Type() == reflect.TypeOf(time.Duration(0)) {
		switch v := raw.(type) {
		case string:
			d, err := time.ParseDuration(v)
			if err != nil {
				return fmt.Errorf("%s: invalid duration %q: %w", path, v, err)
			}
			dst.SetInt(int64(d))
			return nil
		case int, int32, int64:
			dst.SetInt(reflect.ValueOf(v).Int())
			return nil
		case float32, float64:
			dst.SetInt(int64(reflect.ValueOf(v).Float()))
			return nil
		default:
			return fmt.Errorf("%s: expected duration (string or number), got %T", path, raw)
		}
	}

	switch dstKind {
	case reflect.String:
		s, ok := raw.(string)
		if !ok {
			return fmt.Errorf("%s: expected string, got %T", path, raw)
		}
		dst.SetString(s)
		return nil

	case reflect.Bool:
		switch v := raw.(type) {
		case bool:
			dst.SetBool(v)
			return nil
		case string:
			b, err := strconv.ParseBool(strings.TrimSpace(v))
			if err != nil {
				return fmt.Errorf("%s: cannot parse bool %q", path, v)
			}
			dst.SetBool(b)
			return nil
		default:
			return fmt.Errorf("%s: expected bool, got %T", path, raw)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, ok := coerceToInt64(raw)
		if !ok {
			return fmt.Errorf("%s: expected integer, got %T", path, raw)
		}
		dst.SetInt(i64)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u64, ok := coerceToUint64(raw)
		if !ok {
			return fmt.Errorf("%s: expected unsigned integer, got %T", path, raw)
		}
		dst.SetUint(u64)
		return nil

	case reflect.Float32, reflect.Float64:
		f64, ok := coerceToFloat64(raw)
		if !ok {
			return fmt.Errorf("%s: expected float, got %T", path, raw)
		}
		dst.SetFloat(f64)
		return nil
	}

	return fmt.Errorf("%s: unsupported kind %s", path, dstKind)
}

func coerceToInt64(raw interface{}) (int64, bool) {
	switch v := raw.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return int64(reflect.ValueOf(v).Uint()), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		if v == "" {
			return 0, false
		}
		if i, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return int64(f), true
		}
	}
	return 0, false
}

func coerceToUint64(raw interface{}) (uint64, bool) {
	switch v := raw.(type) {
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	case uintptr:
		return uint64(v), true
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i < 0 {
			return 0, false
		}
		return uint64(i), true
	case float32:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case float64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case string:
		if v == "" {
			return 0, false
		}
		if u, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64); err == nil {
			return u, true
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && f >= 0 {
			return uint64(f), true
		}
	}
	return 0, false
}

func coerceToFloat64(raw interface{}) (float64, bool) {
	switch v := raw.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), true
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return float64(reflect.ValueOf(v).Uint()), true
	case string:
		if v == "" {
			return 0, false
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}
