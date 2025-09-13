package dd

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
	ToMap() (map[string]any, error)
}

// Identifiable objects can participate in pointer references by providing a unique Id.
type Identifiable interface {
	GetId() string
}

// Marshaler allows a type to define its own marshalling logic to a map[string]any.
type Marshaler interface {
	MarshalDd() (map[string]any, error)
}

// Unmarshaler allows a type to define its own unmarshalling logic from a map[string]any.
type Unmarshaler interface {
	UnmarshalDd(data map[string]any) error
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

// DdTag holds the parsed values from a `df` struct tag.
type DdTag struct {
	Name       string // external field name override, empty means use default
	Required   bool   // true if field is required during binding
	Secret     bool   // true if field contains sensitive data
	Skip       bool   // true if field should be skipped entirely
	MatchValue string // expected value that must match during binding, empty means no constraint
	HasMatch   bool   // true if a match constraint is specified
}

// parseDdTag parses the `df` struct tag on a field.
//
// tag format: df:"[name][,+required][,+secret][,+match=\"expected_value\"|+match=expected_value]"
//
// special cases:
// - "-"          → skip the field entirely (skip=true)
// - missing/empty → no override (default name, required=false, secret=false, no match constraint)
//
// rules:
// - tokens are comma-separated; surrounding whitespace is ignored.
// - if the first token is not "+required", "+secret", or "+match=...", it is taken as the external field name.
// - the presence of a "+required" token (any position) sets required=true.
// - the presence of a "+secret" token (any position) sets secret=true.
// - a "+match=\"value\"" or "+match=value" token sets a value constraint that must be satisfied during binding.
// - unrecognized tokens are ignored.
func parseDdTag(sf reflect.StructField) DdTag {
	tag := sf.Tag.Get("dd")
	if tag == "-" {
		return DdTag{Skip: true}
	}
	if tag == "" {
		return DdTag{}
	}

	var result DdTag
	parts := strings.Split(tag, ",")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// check for +match="value" or +match=value pattern
		if strings.HasPrefix(p, "+match=") {
			matchPart := strings.TrimPrefix(p, "+match=")
			if len(matchPart) >= 2 && matchPart[0] == '"' && matchPart[len(matchPart)-1] == '"' {
				// properly quoted value: remove quotes
				result.MatchValue = matchPart[1 : len(matchPart)-1]
				result.HasMatch = true
			} else if len(matchPart) > 0 && !strings.Contains(matchPart, "\"") {
				// unquoted value (no quotes at all): use as-is
				result.MatchValue = matchPart
				result.HasMatch = true
			}
			// malformed quoted values (incomplete quotes) are ignored
			continue
		}

		if i == 0 && p != "+required" && p != "+secret" && !strings.HasPrefix(p, "+match=") {
			// first token as name unless it's literally "+required", "+secret", or "+match=..."
			result.Name = p
			continue
		}
		if p == "+required" {
			result.Required = true
		}
		if p == "+secret" {
			result.Secret = true
		}
	}
	return result
}

func toSnakeCase(in string) string {
	if in == "" {
		return ""
	}

	runes := []rune(in)
	if len(runes) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(in) + len(in)/3) // estimate for underscores

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// Add underscore if:
			// 1. Not at start AND
			// 2. (Previous char is lowercase) OR
			//    (Next char is lowercase AND previous char is uppercase - end of acronym)
			if i > 0 {
				prevLower := unicode.IsLower(runes[i-1])
				prevUpper := unicode.IsUpper(runes[i-1])
				nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

				if prevLower || (prevUpper && nextLower) {
					b.WriteByte('_')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Magic string constants for special keys to avoid typos
const (
	TypeKey = "type" // discriminator key for Dynamic types
	RefKey  = "$ref" // reference key for Pointer types
)

var dynamicInterfaceType = reflect.TypeOf((*Dynamic)(nil)).Elem()
var identifiableInterfaceType = reflect.TypeOf((*Identifiable)(nil)).Elem()
var marshalerInterfaceType = reflect.TypeOf((*Marshaler)(nil)).Elem()
var unmarshalerInterfaceType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

// validateTarget validates that the target is a non-nil pointer to a struct.
// returns the struct element and any validation error.
func validateTarget(target interface{}) (reflect.Value, error) {
	if target == nil {
		return reflect.Value{}, &ValidationError{Message: "nil target provided"}
	}
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return reflect.Value{}, &TypeMismatchError{Expected: "non-nil pointer to struct", Actual: fmt.Sprintf("%T", target)}
	}
	elem := value.Elem()
	if elem.Kind() != reflect.Struct {
		return reflect.Value{}, &TypeMismatchError{Expected: "pointer to struct", Actual: fmt.Sprintf("%T", target)}
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
	return nil, &ValidationError{Message: fmt.Sprintf("only one option allowed, got %d", len(opts))}
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
			return nil, true, &ConversionError{Message: "custom converter failed", Cause: err}
		}
		// validate the converted type is assignable
		convertedValue := reflect.ValueOf(result)
		if !convertedValue.Type().AssignableTo(fieldType) {
			return nil, true, &TypeMismatchError{Expected: fieldType.String(), Actual: fmt.Sprintf("%T", result)}
		}
	} else {
		result, err = converter.ToRaw(raw)
		if err != nil {
			return nil, true, &ConversionError{Message: "custom converter failed", Cause: err}
		}
	}

	return result, true, nil
}
