package dd

import (
	"fmt"
	"io/fs"
	"reflect"
	"sort"
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
	// any array indices in the path are ignored for matching purposes.
	// when present for a field, this map takes precedence over DynamicBinders.
	FieldDynamicBinders map[string]map[string]func(map[string]any) (Dynamic, error)

	// Converters maps Go types to custom converters for type conversion.
	// the key is the reflect.Type of the target field, and the value is a Converter
	// that handles bidirectional conversion between raw data and the target type.
	Converters map[reflect.Type]Converter

	// File configures file I/O behavior for helpers such as UnbindJSONFile and UnbindYAMLFile.
	File *FileOptions

	// Strict enables strict acceptance mode: intake rejects duplicate keys
	// and non-JSON scalars and preserves numbers as json.Number; binding
	// rejects unknown input keys and refuses type coercion. see Strict() for
	// the full behavior. the default (false) is dd's forgiving posture, which
	// is unchanged.
	Strict bool
}

// FileOptions configures file output behavior for file-oriented helpers.
type FileOptions struct {
	// Mode sets the file mode used when writing output. If nil, file helpers preserve an
	// existing file's mode when possible and otherwise fall back to their default mode.
	Mode *fs.FileMode
}

// Bind populates the exported fields of target (a pointer to a struct) from the given data map. Keys are matched using
// either a struct tag `dd:"name,+required"` (where name overrides the key and the optional "+required" flag enforces
// presence), `dd:"-"` to skip a field, or, when no tag is provided, a best-effort snake_case conversion of the
// field name.
//
// Use Bind when you need to control how the prototype object is allocated. Use New when you just want to allocate a new
// object to bind off the heap.
//
// supported kinds:
//   - primitives: string, bool, all int/uint sizes, float32/64, time.Duration,
//     time.Time (from RFC3339 strings with optional fractional seconds)
//   - pointers to the above
//   - structs and pointers to structs (recursively bound from map[string]any)
//   - slices of the above (slice items are bound from []interface{})
//   - maps with comparable key types and any supported value type (map keys from JSON/YAML are coerced from strings)
//
// interface types are not supported and will return an error if encountered,
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
	return bindStruct(elem, data, elem.Type().Name(), opt, false, nil)
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
// pre-initialized structs with default values. If Merge has to allocate a fresh struct instance and
// that type implements Defaulter, ApplyDefaults is called before incoming data is bound.
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
	// merge is partial-overlay by design, the opposite posture of strict
	// acceptance — the combination has no coherent meaning.
	if opt != nil && opt.Strict {
		return &ValidationError{Message: "strict mode is not supported for merge operations"}
	}
	return bindStruct(elem, data, elem.Type().Name(), opt, true, nil)
}

func applyDefaultsIfSupported(v reflect.Value) {
	if !v.IsValid() {
		return
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		if v.Type().Implements(defaulterInterfaceType) {
			v.Interface().(Defaulter).ApplyDefaults()
		}
		return
	}
	if v.CanAddr() {
		ptr := v.Addr()
		if ptr.Type().Implements(defaulterInterfaceType) {
			ptr.Interface().(Defaulter).ApplyDefaults()
		}
	}
}

type bindingContext struct {
	consumedKeys   map[string]bool
	extraFieldPath []int
	hasExtra       bool
}

func newBindingContext(structType reflect.Type, path string) (*bindingContext, error) {
	ctx := &bindingContext{consumedKeys: make(map[string]bool)}
	if err := scanExtraField(structType, nil, path, ctx); err != nil {
		return nil, err
	}
	return ctx, nil
}

func scanExtraField(structType reflect.Type, prefix []int, path string, ctx *bindingContext) error {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		fieldPath := append(append([]int(nil), prefix...), i)
		if field.Anonymous {
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct {
				if err := scanExtraField(embeddedType, fieldPath, path, ctx); err != nil {
					return err
				}
			}
			continue
		}

		tag := parseDdTag(field)
		if tag.Skip || !tag.Extra {
			continue
		}
		if field.Type != reflect.TypeOf(map[string]any(nil)) {
			return &TypeMismatchError{
				Path:     path,
				Expected: "map[string]any for +extra field",
				Actual:   field.Type.String(),
			}
		}
		if ctx.hasExtra {
			return &MultipleExtraFieldsError{Path: path}
		}
		ctx.hasExtra = true
		ctx.extraFieldPath = fieldPath
	}
	return nil
}

func resolveExtraField(structValue reflect.Value, fieldPath []int, preserveExisting bool) (reflect.Value, error) {
	value := structValue
	for i, fieldIndex := range fieldPath {
		fieldVal := value.Field(fieldIndex)
		if i == len(fieldPath)-1 {
			return fieldVal, nil
		}
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				if preserveExisting {
					applyDefaultsIfSupported(fieldVal)
				}
			}
			fieldVal = fieldVal.Elem()
		}
		if fieldVal.Kind() != reflect.Struct {
			return reflect.Value{}, &ValidationError{Message: "invalid embedded path to +extra field"}
		}
		value = fieldVal
	}
	return reflect.Value{}, &ValidationError{Message: "missing +extra field path"}
}

func bindStruct(structValue reflect.Value, data map[string]any, path string, opt *Options, preserveExisting bool, ctx *bindingContext) error {
	structType := structValue.Type()

	type deferredUnmarshal struct {
		fieldVal reflect.Value
		rawData  interface{}
		path     string
		name     string
	}
	var deferred []deferredUnmarshal

	// embedded structs share one binding context with their flattened owner.
	// nested named structs start their own context.
	ownsContext := ctx == nil
	if ownsContext {
		var err error
		ctx, err = newBindingContext(structType, path)
		if err != nil {
			return err
		}
	}

	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}

		fieldVal := structValue.Field(i)

		// handle embedded structs by recursively binding their fields
		if field.Anonymous {
			if field.Type.Kind() == reflect.Ptr {
				// for pointer embedded structs, only allocate if there are fields for it in data
				if fieldVal.IsNil() {
					// check if any fields for this embedded struct exist in data
					hasEmbeddedFields := false
					embeddedType := field.Type.Elem()
					if embeddedType.Kind() == reflect.Struct {
						for j := 0; j < embeddedType.NumField(); j++ {
							embeddedField := embeddedType.Field(j)
							if embeddedField.PkgPath != "" { // unexported
								continue
							}
							embeddedTag := parseDdTag(embeddedField)
							if embeddedTag.Skip {
								continue
							}
							embeddedName := embeddedTag.Name
							if embeddedName == "" {
								embeddedName = toSnakeCase(embeddedField.Name)
							}
							if _, exists := data[embeddedName]; exists {
								hasEmbeddedFields = true
								break
							}
						}
					}

					if hasEmbeddedFields {
						// allocate new instance for pointer embedded struct
						fieldVal.Set(reflect.New(field.Type.Elem()))
						if preserveExisting {
							applyDefaultsIfSupported(fieldVal)
						}
					} else {
						// skip if no embedded fields in data
						continue
					}
				}

				if !fieldVal.IsNil() {
					embeddedVal := fieldVal.Elem()
					if embeddedVal.Kind() == reflect.Struct {
						if err := bindStruct(embeddedVal, data, path, opt, preserveExisting, ctx); err != nil {
							return err
						}
					}
				}
			} else {
				// value embedded struct
				embeddedVal := fieldVal
				if embeddedVal.Kind() == reflect.Struct {
					if err := bindStruct(embeddedVal, data, path, opt, preserveExisting, ctx); err != nil {
						return err
					}
				}
			}
			continue
		}

		tag := parseDdTag(field)
		if tag.Skip {
			continue
		}

		// handle +opaque field: the raw subtree is captured uninterpreted —
		// carried, never bound — in both modes.
		if tag.Opaque {
			if field.Type != reflect.TypeOf(map[string]any(nil)) {
				return &TypeMismatchError{
					Path:     path,
					Expected: "map[string]any for +opaque field",
					Actual:   field.Type.String(),
				}
			}
			name := tag.Name
			if name == "" {
				name = toSnakeCase(field.Name)
			}
			raw, ok := data[name]
			if !ok {
				if tag.Required {
					return &RequiredFieldError{Path: path, Field: field.Name}
				}
				continue
			}
			ctx.consumedKeys[name] = true
			subMap, ok := raw.(map[string]any)
			if !ok {
				return &TypeMismatchError{Path: path + "." + field.Name, Expected: "object for +opaque field", Actual: fmt.Sprintf("%T", raw)}
			}
			fieldVal.Set(reflect.ValueOf(subMap))
			continue
		}

		// handle +extra field for capturing unmatched keys
		if tag.Extra {
			continue
		}

		name := tag.Name
		if name == "" {
			name = toSnakeCase(field.Name)
		}

		raw, ok := data[name]
		if ok {
			ctx.consumedKeys[name] = true
		}
		if !ok {
			if tag.Required {
				return &RequiredFieldError{Path: path, Field: field.Name}
			}
			continue
		}

		// validate match constraint if specified
		if tag.HasMatch {
			actualStr := fmt.Sprintf("%v", raw)
			if actualStr != tag.MatchValue {
				return &ValueMismatchError{
					Path:     path,
					Field:    name,
					Expected: tag.MatchValue,
					Actual:   actualStr,
				}
			}
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
			return &BindingError{Path: path, Field: field.Name, Key: name, Cause: err}
		}
	}

	// run deferred unmarshalers now that all other fields are populated.
	for _, d := range deferred {
		if err := unmarshalFromMap(d.fieldVal, d.rawData, d.path, preserveExisting); err != nil {
			return &BindingError{Path: d.path, Key: d.name, Cause: err}
		}
	}

	// only the flattened owner handles unmatched keys after every parent and
	// embedded field has had a chance to consume its declared input.
	if ownsContext {
		var unknown []string
		for key := range data {
			if !ctx.consumedKeys[key] {
				unknown = append(unknown, key)
			}
		}
		sort.Strings(unknown)
		if len(unknown) > 0 {
			if !ctx.hasExtra {
				if opt != nil && opt.Strict {
					return &UnknownFieldError{Path: path, Key: unknown[0]}
				}
			} else {
				extraFieldVal, err := resolveExtraField(structValue, ctx.extraFieldPath, preserveExisting)
				if err != nil {
					return err
				}
				if preserveExisting && !extraFieldVal.IsNil() {
					existing := extraFieldVal.Interface().(map[string]any)
					for _, key := range unknown {
						existing[key] = data[key]
					}
				} else {
					extras := make(map[string]any, len(unknown))
					for _, key := range unknown {
						extras[key] = data[key]
					}
					extraFieldVal.Set(reflect.ValueOf(extras))
				}
			}
		}
	}

	return nil
}

// unmarshalFromMap handles calling the UnmarshalDd method on a field.
func unmarshalFromMap(fieldVal reflect.Value, raw interface{}, path string, preserveExisting bool) error {
	subMap, ok := raw.(map[string]any)
	if !ok {
		return &TypeMismatchError{Path: path, Expected: "object for unmarshaler", Actual: fmt.Sprintf("%T", raw)}
	}

	// handle pointer vs. value receiver for the unmarshaler
	if fieldVal.CanAddr() {
		ptr := fieldVal.Addr()
		if ptr.Type().Implements(unmarshalerInterfaceType) {
			// ensure pointer is allocated for pointer fields
			if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
				fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
			}
			return ptr.Interface().(Unmarshaler).UnmarshalDd(subMap)
		}
	}

	// must be a pointer type that implements the interface directly
	if fieldVal.Type().Implements(unmarshalerInterfaceType) {
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
			if preserveExisting {
				applyDefaultsIfSupported(fieldVal)
			}
		}
		return fieldVal.Interface().(Unmarshaler).UnmarshalDd(subMap)
	}

	return &ValidationError{Field: path, Message: "internal error: field does not implement unmarshaler"} // should be unreachable
}

func setField(fieldVal reflect.Value, raw interface{}, path string, opt *Options, preserveExisting bool) error {
	fieldType := fieldVal.Type()

	if isPointerType(fieldType) {
		subMap, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object for Pointer, got %T", path, raw)
		}
		return bindPointer(fieldVal, subMap, path, opt)
	}

	// handle pointers by allocating as needed then setting the element
	if fieldType.Kind() == reflect.Ptr {
		elemType := fieldType.Elem()

		if isPointerType(elemType) {
			subMap, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("%s: expected object for Pointer, got %T", path, raw)
			}
			if fieldVal.IsNil() {
				fieldVal.Set(reflect.New(elemType))
			}
			return bindPointer(fieldVal.Elem(), subMap, path, opt)
		}

		// special-case *time.Time before checking for struct pointer
		if elemType == reflect.TypeOf(time.Time{}) {
			newPtr := reflect.New(elemType)
			if err := setNonPtrValue(newPtr.Elem(), raw, path, opt, preserveExisting); err != nil {
				return err
			}
			fieldVal.Set(newPtr)
			return nil
		}

		if elemType.Kind() == reflect.Struct {
			subMap, ok := raw.(map[string]any)
			if !ok {
				return &TypeMismatchError{Path: path, Expected: "object for struct pointer", Actual: fmt.Sprintf("%T", raw)}
			}
			// if preserveExisting and pointer is not nil, bind to existing struct
			if preserveExisting && !fieldVal.IsNil() {
				if err := bindStruct(fieldVal.Elem(), subMap, path, opt, preserveExisting, nil); err != nil {
					return err
				}
			} else {
				// allocate new struct and bind into it
				newPtr := reflect.New(elemType)
				if preserveExisting {
					applyDefaultsIfSupported(newPtr)
				}
				if err := bindStruct(newPtr.Elem(), subMap, path, opt, preserveExisting, nil); err != nil {
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

	if isPointerType(fieldVal.Type()) {
		subMap, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object for Pointer, got %T", path, raw)
		}
		return bindPointer(fieldVal, subMap, path, opt)
	}

	// special-case time.Time before checking struct kind (since time.Time is a struct)
	if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
		if opt != nil && opt.Strict {
			return strictConvertAndSet(fieldVal, raw, path)
		}
		switch v := raw.(type) {
		case string:
			t, err := time.Parse(time.RFC3339Nano, v)
			if err != nil {
				return fmt.Errorf("%s: cannot parse time: %w", path, err)
			}
			fieldVal.Set(reflect.ValueOf(t))
			return nil
		case time.Time:
			fieldVal.Set(reflect.ValueOf(v))
			return nil
		default:
			return fmt.Errorf("%s: expected time (RFC3339 string with optional fractional seconds or time.Time), got %T", path, raw)
		}
	}

	switch fieldVal.Kind() {
	case reflect.Struct:
		subMap, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object for struct, got %T", path, raw)
		}
		return bindStruct(fieldVal, subMap, path, opt, preserveExisting, nil)

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
				if isPointerType(elemType.Elem()) {
					subMap, ok := item.(map[string]any)
					if !ok {
						return fmt.Errorf("%s: expected object for Pointer, got %T", itemPath, item)
					}
					if err := bindPointer(elemPtr.Elem(), subMap, itemPath, opt); err != nil {
						return err
					}
					out = reflect.Append(out, elemPtr)
					continue
				}
				if elemType.Elem().Kind() == reflect.Struct {
					subMap, ok := item.(map[string]any)
					if !ok {
						return fmt.Errorf("%s: expected object for struct slice element, got %T", itemPath, item)
					}
					if preserveExisting {
						applyDefaultsIfSupported(elemPtr)
					}
					if err := bindStruct(elemPtr.Elem(), subMap, itemPath, opt, preserveExisting, nil); err != nil {
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
			if isPointerType(elemType) {
				subMap, ok := item.(map[string]any)
				if !ok {
					return fmt.Errorf("%s: expected object for Pointer, got %T", itemPath, item)
				}
				if err := bindPointer(elemVal, subMap, itemPath, opt); err != nil {
					return err
				}
				out = reflect.Append(out, elemVal)
				continue
			}
			if elemType.Kind() == reflect.Struct {
				subMap, ok := item.(map[string]any)
				if !ok {
					return fmt.Errorf("%s: expected object for struct slice element, got %T", itemPath, item)
				}
				if preserveExisting {
					applyDefaultsIfSupported(elemVal)
				}
				if err := bindStruct(elemVal, subMap, itemPath, opt, preserveExisting, nil); err != nil {
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

	case reflect.Map:
		rawMap, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object for map field, got %T", path, raw)
		}

		keyType := fieldVal.Type().Key()
		elemType := fieldVal.Type().Elem()
		if opt != nil && opt.Strict && keyType.Kind() != reflect.String {
			return &TypeMismatchError{Path: path, Expected: "map with string keys", Actual: fieldVal.Type().String()}
		}

		// create new map
		newMap := reflect.MakeMap(fieldVal.Type())

		// populate map with converted keys and values
		for keyStr, value := range rawMap {
			itemPath := fmt.Sprintf("%s[%q]", path, keyStr)

			var keyVal reflect.Value
			if opt != nil && opt.Strict {
				keyVal = reflect.New(keyType).Elem()
				keyVal.SetString(keyStr)
			} else {
				// forgiving mode converts string keys to supported target key types
				var err error
				keyVal, err = stringToKey(keyStr, keyType)
				if err != nil {
					return fmt.Errorf("%s: %w", path, err)
				}
			}

			// handle different value types similar to slice element handling
			if elemType.Kind() == reflect.Interface && elemType == dynamicInterfaceType {
				// handle Dynamic interface values
				subMap, ok := value.(map[string]any)
				if !ok {
					return fmt.Errorf("%s: expected object for Dynamic element, got %T", itemPath, value)
				}
				dynVal, err := bindDynamic(subMap, itemPath, opt)
				if err != nil {
					return err
				}
				newMap.SetMapIndex(keyVal, reflect.ValueOf(dynVal))
				continue
			}

			if elemType.Kind() == reflect.Ptr {
				// pointer to value
				elemPtr := reflect.New(elemType.Elem())
				if isPointerType(elemType.Elem()) {
					subMap, ok := value.(map[string]any)
					if !ok {
						return fmt.Errorf("%s: expected object for Pointer, got %T", itemPath, value)
					}
					if err := bindPointer(elemPtr.Elem(), subMap, itemPath, opt); err != nil {
						return err
					}
					newMap.SetMapIndex(keyVal, elemPtr)
					continue
				}
				if elemType.Elem().Kind() == reflect.Struct {
					// pointer to struct
					subMap, ok := value.(map[string]any)
					if !ok {
						return fmt.Errorf("%s: expected object for struct map value, got %T", itemPath, value)
					}
					if preserveExisting {
						applyDefaultsIfSupported(elemPtr)
					}
					if err := bindStruct(elemPtr.Elem(), subMap, itemPath, opt, preserveExisting, nil); err != nil {
						return err
					}
					newMap.SetMapIndex(keyVal, elemPtr)
					continue
				}
				// pointer to primitive
				if err := setNonPtrValue(elemPtr.Elem(), value, itemPath, opt, preserveExisting); err != nil {
					return err
				}
				newMap.SetMapIndex(keyVal, elemPtr)
				continue
			}

			// non-pointer value
			elemVal := reflect.New(elemType).Elem()
			if isPointerType(elemType) {
				subMap, ok := value.(map[string]any)
				if !ok {
					return fmt.Errorf("%s: expected object for Pointer, got %T", itemPath, value)
				}
				if err := bindPointer(elemVal, subMap, itemPath, opt); err != nil {
					return err
				}
				newMap.SetMapIndex(keyVal, elemVal)
				continue
			}
			if elemType.Kind() == reflect.Struct {
				// struct value
				subMap, ok := value.(map[string]any)
				if !ok {
					return fmt.Errorf("%s: expected object for struct map value, got %T", itemPath, value)
				}
				if preserveExisting {
					applyDefaultsIfSupported(elemVal)
				}
				if err := bindStruct(elemVal, subMap, itemPath, opt, preserveExisting, nil); err != nil {
					return err
				}
				newMap.SetMapIndex(keyVal, elemVal)
				continue
			}
			if elemType.Kind() == reflect.Map {
				// nested map
				if err := setField(elemVal, value, itemPath, opt, preserveExisting); err != nil {
					return err
				}
				newMap.SetMapIndex(keyVal, elemVal)
				continue
			}
			if elemType.Kind() == reflect.Slice {
				// slice value
				if err := setNonPtrValue(elemVal, value, itemPath, opt, preserveExisting); err != nil {
					return err
				}
				newMap.SetMapIndex(keyVal, elemVal)
				continue
			}
			// primitive or interface value
			if elemType.Kind() == reflect.Interface {
				// interface{} or any type - store raw value
				newMap.SetMapIndex(keyVal, reflect.ValueOf(value))
				continue
			}
			// primitive value
			if err := convertAndSet(elemVal, value, itemPath, opt); err != nil {
				return err
			}
			newMap.SetMapIndex(keyVal, elemVal)
		}
		fieldVal.Set(newMap)
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
