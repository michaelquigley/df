package df

import (
	"reflect"
	"strings"
	"unicode"
)

// Dynamic fields can be used when the concrete type of a field is selected dynamically through the `type` data provided
// in the incoming `map` that will be passed to `Bind`.
type Dynamic interface {
	Type() string
	ToMap() map[string]any
}

func parseDFTag(sf reflect.StructField) (name string, required bool, skip bool) {
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
	var b strings.Builder
	b.Grow(len(in) + len(in)/2)
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
