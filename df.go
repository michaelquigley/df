package df

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// Dynamic fields can be used when the concrete type of a field is selected dynamically through the `type` data provided
// in the incoming `map` that will be passed to `Bind`. A polymorphic field type.
type Dynamic interface {
	Type() string
	ToMap() map[string]any
}

// Identifiable objects can participate in pointer references by providing a unique Id.
type Identifiable interface {
	GetId() string
}

// Marshaler allows a type to define its own marshalling logic to a map[string]any.
type Marshaler interface {
	MarshalDf() (map[string]any, error)
}

// Unmarshaler allows a type to define its own unmarshalling logic from a map[string]any.
type Unmarshaler interface {
	UnmarshalDf(data map[string]any) error
}

// Converter defines a bidirectional type conversion interface for custom field types.
// it allows users to define how their custom types should be converted to/from the raw data.
type Converter interface {
	// FromRaw converts a raw value (from the data map) to the target type.
	// the input can be any type that appears in the data map (string, int, bool, etc.).
	FromRaw(raw interface{}) (interface{}, error)
	
	// ToRaw converts a typed value back to a raw value for serialization.
	// the output should be a type that can be marshaled (string, int, bool, etc.).
	ToRaw(value interface{}) (interface{}, error)
}

// parseDfTag parses the `df` struct tag on a field.
//
// tag format: df:"[name][,required]"
//
// special cases:
// - "-"          → skip the field entirely (skip=true)
// - missing/empty → no override (default name, required=false)
//
// rules:
// - tokens are comma-separated; surrounding whitespace is ignored.
// - if the first token is not "required", it is taken as the external field name.
// - the presence of a "required" token (any position) sets required=true.
// - unrecognized tokens are ignored.
func parseDfTag(sf reflect.StructField) (name string, required bool, skip bool) {
	tag := sf.Tag.Get("df")
	if tag == "-" {
		return "", false, true
	}
	if tag == "" {
		return "", false, false
	}
	parts := strings.Split(tag, ",")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if i == 0 && p != "required" { // first token as name unless it's literally "required"
			name = p
			continue
		}
		if p == "required" {
			required = true
		}
	}
	return name, required, false
}

func toSnakeCase(in string) string {
	if in == "" {
		return ""
	}

	// Count uppercase letters to estimate capacity more accurately
	upperCount := 0
	for _, r := range in {
		if unicode.IsUpper(r) {
			upperCount++
		}
	}

	// Allocate precise capacity: original length + underscores needed
	var b strings.Builder
	b.Grow(len(in) + upperCount - 1) // -1 because first upper doesn't get underscore

	for i, r := range in {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var dynamicInterfaceType = reflect.TypeOf((*Dynamic)(nil)).Elem()
var identifiableInterfaceType = reflect.TypeOf((*Identifiable)(nil)).Elem()
var marshalerInterfaceType = reflect.TypeOf((*Marshaler)(nil)).Elem()
var unmarshalerInterfaceType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
var converterInterfaceType = reflect.TypeOf((*Converter)(nil)).Elem()

// validateTarget validates that the target is a non-nil pointer to a struct.
// returns the struct element and any validation error.
func validateTarget(target interface{}) (reflect.Value, error) {
	if target == nil {
		return reflect.Value{}, fmt.Errorf("nil target provided")
	}
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return reflect.Value{}, fmt.Errorf("target must be a non-nil pointer to struct; got %T", target)
	}
	elem := value.Elem()
	if elem.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("target must be a pointer to struct; got %T", target)
	}
	return elem, nil
}

// getOptions extracts and validates options from variadic parameters.
// returns the options and any validation error.
func getOptions(opts ...*Options) (*Options, error) {
	if len(opts) == 0 {
		return nil, nil
	}
	if len(opts) == 1 {
		return opts[0], nil
	}
	return nil, fmt.Errorf("only one option allowed, got %d", len(opts))
}

// tryCustomConverter attempts to use a custom converter for the given field and raw value.
// returns (convertedValue, wasConverted, error).
func tryCustomConverter(fieldType reflect.Type, raw interface{}, opt *Options, forBinding bool) (interface{}, bool, error) {
	if opt == nil || opt.Converters == nil {
		return nil, false, nil
	}
	
	converter, ok := opt.Converters[fieldType]
	if !ok {
		return nil, false, nil
	}
	
	var result interface{}
	var err error
	
	if forBinding {
		result, err = converter.FromRaw(raw)
		if err != nil {
			return nil, true, fmt.Errorf("custom converter failed: %w", err)
		}
		// validate the converted type is assignable
		convertedValue := reflect.ValueOf(result)
		if !convertedValue.Type().AssignableTo(fieldType) {
			return nil, true, fmt.Errorf("custom converter returned incompatible type %T, expected %s", result, fieldType)
		}
	} else {
		result, err = converter.ToRaw(raw)
		if err != nil {
			return nil, true, fmt.Errorf("custom converter failed: %w", err)
		}
	}
	
	return result, true, nil
}
