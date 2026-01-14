package da

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// component represents a discovered component with its order for processing.
type component struct {
	value reflect.Value
	order int
}

// traverse finds all pointer fields in a struct recursively,
// sorted by `da:"order=N"` tag (lower first, default 0).
// Fields with `da:"-"` are skipped.
func traverse(v reflect.Value) []component {
	var components []component
	traverseRecursive(v, &components)
	sort.Slice(components, func(i, j int) bool {
		return components[i].order < components[j].order
	})
	return components
}

func traverseRecursive(v reflect.Value, components *[]component) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		// skip unexported fields
		if !structField.IsExported() {
			continue
		}

		// parse tag
		tag := structField.Tag.Get("da")
		if tag == "-" {
			continue
		}
		order := parseOrder(tag)

		// handle different field types
		switch field.Kind() {
		case reflect.Ptr:
			if !field.IsNil() {
				*components = append(*components, component{value: field, order: order})
			}
		case reflect.Struct:
			// recurse into embedded/nested structs
			traverseRecursive(field, components)
		case reflect.Slice:
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.Kind() == reflect.Ptr && !elem.IsNil() {
					*components = append(*components, component{value: elem, order: order})
				}
			}
		case reflect.Map:
			iter := field.MapRange()
			for iter.Next() {
				val := iter.Value()
				if val.Kind() == reflect.Ptr && !val.IsNil() {
					*components = append(*components, component{value: val, order: order})
				}
			}
		}
	}
}

func parseOrder(tag string) int {
	for _, part := range strings.Split(tag, ",") {
		if strings.HasPrefix(part, "order=") {
			if n, err := strconv.Atoi(strings.TrimPrefix(part, "order=")); err == nil {
				return n
			}
		}
	}
	return 0
}
