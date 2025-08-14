package df

import (
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
