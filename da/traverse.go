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

// addComponent extracts a component from a value, handling ptr, struct, and interface types.
// Returns the value to add and whether it's valid.
func addComponent(v reflect.Value) (reflect.Value, bool) {
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			return v, true
		}
	case reflect.Struct:
		if v.CanAddr() {
			return v.Addr(), true
		}
	case reflect.Interface:
		if !v.IsNil() {
			// unwrap interface to get underlying value
			return addComponent(v.Elem())
		}
	}
	return reflect.Value{}, false
}

func traverseRecursive(v reflect.Value, components *[]component) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice:
		// handle slice at top level - iterate through elements
		for i := 0; i < v.Len(); i++ {
			if val, ok := addComponent(v.Index(i)); ok {
				*components = append(*components, component{value: val, order: 0})
			}
		}
		return
	case reflect.Map:
		// handle map at top level - iterate through values
		iter := v.MapRange()
		for iter.Next() {
			if val, ok := addComponent(iter.Value()); ok {
				*components = append(*components, component{value: val, order: 0})
			}
		}
		return
	case reflect.Struct:
		// continue with struct traversal below
	default:
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
		case reflect.Interface:
			if val, ok := addComponent(field); ok {
				*components = append(*components, component{value: val, order: order})
			}
		case reflect.Struct:
			// recurse into embedded/nested structs
			traverseRecursive(field, components)
		case reflect.Slice:
			for j := 0; j < field.Len(); j++ {
				if val, ok := addComponent(field.Index(j)); ok {
					*components = append(*components, component{value: val, order: order})
				}
			}
		case reflect.Map:
			iter := field.MapRange()
			for iter.Next() {
				if val, ok := addComponent(iter.Value()); ok {
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
